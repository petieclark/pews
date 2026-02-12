package communication

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Sender orchestrates email/SMS delivery for campaigns and auto-responses.
type Sender struct {
	mailgun *MailgunClient
	twilio  *TwilioClient
	db      *pgxpool.Pool
}

// NewSender initialises provider clients from env vars.
func NewSender(db *pgxpool.Pool) *Sender {
	return &Sender{
		mailgun: NewMailgunClient(),
		twilio:  NewTwilioClient(),
		db:      db,
	}
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
			if !s.mailgun.IsConfigured() {
				log.Printf("[communication] campaign %s: skipping email to %s — Mailgun not configured", campaign.ID, r.Email)
				continue
			}
			if r.Email == "" {
				s.updateRecipientStatus(ctx, campaign.ID, r.PersonID, "failed")
				failCount++
				continue
			}
			sendErr = s.mailgun.SendEmail(r.Email, subject, body, "", churchName)

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
	if !s.mailgun.IsConfigured() {
		log.Printf("[communication] skipping welcome email — Mailgun not configured")
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

	if err := s.mailgun.SendEmail(card.Email, subject, body, "", churchName); err != nil {
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
