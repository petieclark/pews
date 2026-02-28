package communication

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewSender initialises provider clients from env vars.
func NewSender(db *pgxpool.Pool) *Sender {
	devMode := os.Getenv("DEV_MODE") == "true"
	return &Sender{
		sendgrid: NewSendGridClient(),
		twilio:   NewTwilioClient(),
		db:       db,
		devMode:  devMode,
	}
}

// Sender orchestrates email/SMS delivery for campaigns and auto-responses.
type Sender struct {
	sendgrid *SendGridClient
	twilio   *TwilioClient
	db       *pgxpool.Pool
	devMode  bool // if true, logs instead of sending
}

// Message represents a unified message interface for email or SMS
type Message struct {
	ToEmail    string            // recipient email address (for email messages)
	ToPhone    string            // recipient phone number (for SMS)
	Subject    string            // email subject line (optional for SMS)
	Body       string            // message body (HTML for email, plain text for SMS)
	MergeData  RecipientInfo     // merge data for personalization
	ChurchName string            // church name for branding
}

// Send sends a unified message via the appropriate channel.
// It automatically detects whether to send email or SMS based on recipient type.
func (s *Sender) Send(ctx context.Context, msg Message) error {
	if s.devMode {
		log.Printf("[communication] [DEV MODE] would send %s to %s: %s", 
			map[bool]string{true: "email", false: "SMS"}[msg.ToPhone != ""],
			msg.ToEmail+msg.ToPhone, msg.Body[:min(50, len(msg.Body))]+"...")
		return nil // no actual send in dev mode
	}

	if msg.ToEmail != "" {
		// Send email via SendGrid
		subject := MergeTags(msg.Subject, msg.MergeData, msg.ChurchName)
		body := MergeTags(msg.Body, msg.MergeData, msg.ChurchName)
		
		if !s.sendgrid.IsConfigured() {
			return fmt.Errorf("SendGrid not configured for email delivery")
		}
		
		return s.sendgrid.SendEmail(msg.ToEmail, subject, body, "", msg.ChurchName)
	}

	if msg.ToPhone != "" {
		// Send SMS via Twilio
		body := MergeTags(msg.Body, msg.MergeData, msg.ChurchName)
		
		if !s.twilio.IsConfigured() {
			return fmt.Errorf("Twilio not configured for SMS delivery")
		}
		
		return s.twilio.SendSMS(msg.ToPhone, body)
	}

	return fmt.Errorf("no recipient specified (email or phone required)")
}

// GetSender returns the sender instance (for backward compatibility).
func (h *Handler) GetSender() *Sender {
	return h.service.GetSender()
}

// RecipientInfo holds the data needed to personalise and deliver a message.
type RecipientInfo struct {
	PersonID  string
	FirstName string
	LastName  string
	Email     string
	Phone     string
}

// MergeTags replaces merge tags in text with recipient data.
func MergeTags(text string, r RecipientInfo, churchName string) string {
	replacer := strings.NewReplacer(
		"{{first_name}}", r.FirstName,
		"{{last_name}}", r.LastName,
		"{{email}}", r.Email,
		"{{church_name}}", churchName,
		"{first_name}", r.FirstName,
		"{last_name}", r.LastName,
		"{email}", r.Email,
		"{church_name}", churchName,
	)
	return replacer.Replace(text)
}

// SendCampaign delivers a campaign to all its recipients.
// It resolves recipients, iterates them, sends via the appropriate channel,
// and tracks per-recipient status. It does NOT fail the whole campaign if
// individual sends fail.
func (s *Sender) SendCampaign(ctx context.Context, tenantID string, campaign *Campaign) error {
	// Get church name for merge tags
	churchName := s.getTenantName(ctx, tenantID)

	// Resolve recipients based on target type
	recipients, err := s.resolveRecipients(ctx, tenantID, campaign)
	if err != nil {
		return fmt.Errorf("failed to resolve recipients: %w", err)
	}

	if len(recipients) == 0 {
		log.Printf("[communication] campaign %s: no recipients found", campaign.ID)
		return nil
	}

	// Insert recipient rows as pending
	if err := s.insertRecipients(ctx, campaign.ID, recipients); err != nil {
		return fmt.Errorf("failed to insert recipients: %w", err)
	}

	// Update recipient count
	_, _ = s.db.Exec(ctx, "UPDATE campaigns SET recipient_count = $1 WHERE id = $2", len(recipients), campaign.ID)

	var sentCount, failCount int

	for _, r := range recipients {
		subject := MergeTags(campaign.Subject, r, churchName)
		body := MergeTags(campaign.Body, r, churchName)

		var sendErr error
		switch campaign.Channel {
		case "email":
			if !s.sendgrid.IsConfigured() {
				log.Printf("[communication] campaign %s: skipping email to %s — SendGrid not configured", campaign.ID, r.Email)
				continue
			}
			if r.Email == "" {
				s.updateRecipientStatus(ctx, campaign.ID, r.PersonID, "failed")
				failCount++
				continue
			}
			sendErr = s.sendgrid.SendEmail(r.Email, subject, body, "", churchName)

		case "sms":
			if !s.twilio.IsConfigured() {
				log.Printf("[communication] campaign %s: skipping SMS to %s — Twilio not configured", campaign.ID, r.PersonID)
				continue
			}
			if r.Phone == "" {
				s.updateRecipientStatus(ctx, campaign.ID, r.PersonID, "failed")
				failCount++
				continue
			}
			sendErr = s.twilio.SendSMS(r.Phone, body)

		default:
			log.Printf("[communication] campaign %s: unknown channel %q", campaign.ID, campaign.Channel)
			continue
		}

		if sendErr != nil {
			log.Printf("[communication] campaign %s: failed to send to %s: %v", campaign.ID, r.PersonID, sendErr)
			s.updateRecipientStatus(ctx, campaign.ID, r.PersonID, "failed")
			failCount++
		} else {
			s.updateRecipientStatus(ctx, campaign.ID, r.PersonID, "sent")
			sentCount++
		}

		// Simple rate limiting: 50ms between sends
		time.Sleep(50 * time.Millisecond)
	}

	log.Printf("[communication] campaign %s: sent=%d failed=%d total=%d", campaign.ID, sentCount, failCount, len(recipients))
	return nil
}

// SendWelcomeEmail sends a welcome email for a connection card submission.
func (s *Sender) SendWelcomeEmail(ctx context.Context, tenantID string, card *ConnectionCard) {
	if s.devMode {
		log.Printf("[communication] [DEV MODE] would send welcome email to %s", card.Email)
		return
	}

	if !s.sendgrid.IsConfigured() {
		log.Printf("[communication] skipping welcome email — SendGrid not configured")
		return
	}
	if card.Email == "" {
		return
	}

	churchName := s.getTenantName(ctx, tenantID)

	// Look for a welcome template for this tenant
	subject := fmt.Sprintf("Welcome to %s!", churchName)
	body := fmt.Sprintf(
		"<p>Hi %s,</p><p>Thank you for connecting with %s! We're so glad you visited.</p><p>We look forward to seeing you again soon.</p><p>— The %s Team</p>",
		card.FirstName, churchName, churchName,
	)

	// Try to load a custom welcome template
	var tmplSubject, tmplBody string
	err := s.db.QueryRow(ctx,
		`SELECT subject, body FROM message_templates WHERE tenant_id = $1 AND category = 'welcome' AND channel = 'email' LIMIT 1`,
		tenantID,
	).Scan(&tmplSubject, &tmplBody)
	if err == nil && tmplBody != "" {
		r := RecipientInfo{FirstName: card.FirstName, LastName: card.LastName, Email: card.Email, Phone: card.Phone}
		subject = MergeTags(tmplSubject, r, churchName)
		body = MergeTags(tmplBody, r, churchName)
	}

	if err := s.sendgrid.SendEmail(card.Email, subject, body, "", churchName); err != nil {
		log.Printf("[communication] failed to send welcome email to %s: %v", card.Email, err)
	} else {
		log.Printf("[communication] sent welcome email to %s", card.Email)
	}
}

// --- internal helpers ---

func (s *Sender) getTenantName(ctx context.Context, tenantID string) string {
	var name string
	_ = s.db.QueryRow(ctx, "SELECT name FROM tenants WHERE id = $1", tenantID).Scan(&name)
	if name == "" {
		name = "Our Church"
	}
	return name
}

func (s *Sender) resolveRecipients(ctx context.Context, tenantID string, campaign *Campaign) ([]RecipientInfo, error) {
	var query string
	var args []interface{}

	switch campaign.TargetType {
	case "all":
		query = `SELECT id, first_name, last_name, COALESCE(email,''), COALESCE(phone,'') FROM people WHERE tenant_id = $1`
		args = []interface{}{tenantID}
	case "tag":
		query = `SELECT p.id, p.first_name, p.last_name, COALESCE(p.email,''), COALESCE(p.phone,'')
		         FROM people p JOIN person_tags pt ON p.id = pt.person_id
		         WHERE p.tenant_id = $1 AND pt.tag_id = $2`
		args = []interface{}{tenantID, campaign.TargetID}
	case "group":
		query = `SELECT p.id, p.first_name, p.last_name, COALESCE(p.email,''), COALESCE(p.phone,'')
		         FROM people p JOIN group_members gm ON p.id = gm.person_id
		         WHERE p.tenant_id = $1 AND gm.group_id = $2`
		args = []interface{}{tenantID, campaign.TargetID}
	default:
		// manual — recipients should already be in campaign_recipients
		return s.getExistingRecipientInfo(ctx, tenantID, campaign.ID)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []RecipientInfo
	for rows.Next() {
		var r RecipientInfo
		if err := rows.Scan(&r.PersonID, &r.FirstName, &r.LastName, &r.Email, &r.Phone); err != nil {
			return nil, err
		}
		recipients = append(recipients, r)
	}
	return recipients, nil
}

func (s *Sender) getExistingRecipientInfo(ctx context.Context, tenantID, campaignID string) ([]RecipientInfo, error) {
	query := `SELECT p.id, p.first_name, p.last_name, COALESCE(p.email,''), COALESCE(p.phone,'')
	          FROM campaign_recipients cr JOIN people p ON cr.person_id = p.id
	          WHERE cr.campaign_id = $1`
	rows, err := s.db.Query(ctx, query, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recipients []RecipientInfo
	for rows.Next() {
		var r RecipientInfo
		if err := rows.Scan(&r.PersonID, &r.FirstName, &r.LastName, &r.Email, &r.Phone); err != nil {
			return nil, err
		}
		recipients = append(recipients, r)
	}
	return recipients, nil
}

func (s *Sender) insertRecipients(ctx context.Context, campaignID string, recipients []RecipientInfo) error {
	for _, r := range recipients {
		_, err := s.db.Exec(ctx,
			`INSERT INTO campaign_recipients (id, campaign_id, person_id, status)
			 VALUES (gen_random_uuid(), $1, $2, 'pending')
			 ON CONFLICT (campaign_id, person_id) DO NOTHING`,
			campaignID, r.PersonID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Sender) updateRecipientStatus(ctx context.Context, campaignID, personID, status string) {
	sentAt := time.Now()
	var sentAtPtr *time.Time
	if status == "sent" {
		sentAtPtr = &sentAt
	}
	_, _ = s.db.Exec(ctx,
		`UPDATE campaign_recipients SET status = $1, sent_at = $2 WHERE campaign_id = $3 AND person_id = $4`,
		status, sentAtPtr, campaignID, personID,
	)
}
