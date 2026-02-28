package teams

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationService handles volunteer assignment token generation and email sending
type NotificationService struct {
	db              *pgxpool.Pool
	tokenSecret     string
	expiryHours     int
	sendgridEnabled bool
	fromEmail       string
	fromName        string
}

// AssignmentNotification holds data for generating response tokens
type AssignmentNotification struct {
	AssignmentID string
	PersonID     string
	Token        string
	ResponseURL  string
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *pgxpool.Pool, jwtSecret string) *NotificationService {
	expiryHours := 168 // 7 days
	
	// Check SendGrid configuration
	sendgridEnabled := false
	fromEmail := ""
	fromName := "Pews Church Management"
	
	if apiKey := getEnv("SENDGRID_API_KEY"); apiKey != "" && len(apiKey) > 10 {
		sendgridEnabled = true
	}
	if fromEmail = getEnv("SENDGRID_FROM_EMAIL"); fromEmail == "" {
		fromEmail = "noreply@localhost"
	}
	if fromName = getEnv("SENDGRID_FROM_NAME"); fromName == "" {
		fromName = "Pews Church Management"
	}

	return &NotificationService{
		db:              db,
		tokenSecret:     jwtSecret,
		expiryHours:     expiryHours,
		sendgridEnabled: sendgridEnabled,
		fromEmail:       fromEmail,
		fromName:        fromName,
	}
}

// GenerateToken creates a signed JWT-like token for volunteer assignment response
func (ns *NotificationService) GenerateToken(assignmentID, personID string) (string, error) {
	tokenData := map[string]interface{}{
		"assignment_id": assignmentID,
		"person_id":     personID,
		"expires_at":    time.Now().Add(time.Duration(ns.expiryHours) * time.Hour),
	}

	jsonBytes, err := json.Marshal(tokenData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token data: %w", err)
	}

	// Create HMAC signature
	h := hmac.New(sha256.New, []byte(ns.tokenSecret))
	h.Write(jsonBytes)
	signature := h.Sum(nil)

	// Base64 encode both payload and signature
	payload := base64.URLEncoding.EncodeToString(jsonBytes)
	sig := base64.URLEncoding.EncodeToString(signature)

	return fmt.Sprintf("%s.%s", payload, sig), nil
}

// SendAssignmentNotification sends email notification to volunteer for new assignment
func (ns *NotificationService) SendAssignmentNotification(ctx context.Context, assignmentID, serviceID string) error {
	// Fetch assignment details with person and service info
	var personFirstName, personLastName, personEmail, serviceName, serviceDate, teamName string
	
	err := ns.db.QueryRow(ctx, `
		SELECT p.first_name, p.last_name, COALESCE(p.email, ''), s.name, 
		       TO_CHAR(s.service_date, 'Month DD, YYYY'), t.name
		FROM service_team_assignments sta
		JOIN people p ON p.id = sta.person_id
		JOIN services s ON s.id = sta.service_id
		JOIN teams t ON t.id = sta.team_id
		WHERE sta.id = $1 AND sta.service_id = $2`,
		assignmentID, serviceID).Scan(&personFirstName, &personLastName, &personEmail, &serviceName, &serviceDate, &teamName)

	if err != nil {
		return fmt.Errorf("failed to fetch assignment details: %w", err)
	}

	if personEmail == "" {
		return fmt.Errorf("no email address for volunteer %s %s", personFirstName, personLastName)
	}

	// Generate response token
	token, err := ns.GenerateToken(assignmentID, personEmail)
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	responseURL := fmt.Sprintf("https://outbound.clearlinelims.com/respond/%s?accept=true", token)
	declineURL := fmt.Sprintf("https://outbound.clearlinelims.com/respond/%s?decline=true", token)

	// Check dev mode
	devMode := getEnv("DEV_MODE") == "true"
	
	if devMode {
		fmt.Printf("[NOTIFICATION] [DEV MODE] Would send email to %s (%s): %s assigned to %s on %s\n", 
			personFirstName, personEmail, teamName, serviceName, serviceDate)
		fmt.Printf("[NOTIFICATION] [DEV MODE] Response URL: %s\n", responseURL)
		return nil // Skip actual send in dev mode
	}

	if !ns.sendgridEnabled {
		fmt.Printf("[NOTIFICATION] SendGrid not configured - skipping email to %s\n", personEmail)
		return fmt.Errorf("SendGrid not configured")
	}

	// Render email body (simple HTML template)
	emailBody := ns.renderAssignmentEmail(personFirstName, serviceName, serviceDate, teamName, responseURL, declineURL)

	// Send via SendGrid
	err = ns.sendViaSendGrid(personEmail, fmt.Sprintf("Volunteer Assignment: %s", serviceName), emailBody)
	if err != nil {
		return fmt.Errorf("failed to send notification email: %w", err)
	}

	return nil
}

// renderAssignmentEmail creates HTML email body for assignment notification
func (ns *NotificationService) renderAssignmentEmail(firstName, serviceName, serviceDate, teamName, acceptURL, declineURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Volunteer Assignment</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #f5f5f5; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header { background-color: #0D1B2A; color: white; padding: 32px 24px; text-align: center; }
        .header h1 { margin: 0; font-size: 24px; }
        .content { padding: 32px 24px; }
        .greeting { font-size: 18px; color: #333; margin-bottom: 20px; }
        .assignment-details { background-color: #f8f9fa; padding: 20px; border-radius: 6px; margin: 24px 0; }
        .detail-row { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid #e9ecef; }
        .detail-label { font-weight: 600; color: #555; }
        .detail-value { color: #333; }
        .detail-row:last-child { border-bottom: none; }
        .cta-buttons { text-align: center; margin: 32px 0; }
        .btn { display: inline-block; padding: 14px 32px; margin: 8px; border-radius: 6px; text-decoration: none; font-weight: 600; font-size: 16px; }
        .btn-accept { background-color: #28a745; color: white; }
        .btn-decline { background-color: #dc354f; color: white; }
        .footer { background-color: #f8f9fa; padding: 24px; text-align: center; font-size: 14px; color: #6c757d; border-top: 1px solid #e9ecef; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🙏 Volunteer Assignment</h1>
        </div>
        
        <div class="content">
            <p class="greeting">Hi %s,</p>
            
            <p>You've been assigned to serve on the team for an upcoming service. Please confirm your availability using the buttons below.</p>
            
            <div class="assignment-details">
                <div class="detail-row">
                    <span class="detail-label">Service:</span>
                    <span class="detail-value">%s</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Date:</span>
                    <span class="detail-value">%s</span>
                </div>
                <div class="detail-row">
                    <span class="detail-label">Team:</span>
                    <span class="detail-value">%s</span>
                </div>
            </div>
            
            <div class="cta-buttons">
                <a href="%s" class="btn btn-accept">✅ Confirm Attendance</a>
                <a href="%s" class="btn btn-decline">👋 Can't Make It</a>
            </div>
            
            <p style="font-size: 14px; color: #6c757d;">If the buttons don't work, you can copy and paste this link into your browser:</p>
            <p style="font-size: 12px; word-break: break-all; color: #0d6efd;">%s</p>
        </div>
        
        <div class="footer">
            Thank you for serving! Your contribution makes a difference. 🙏<br/>
            If you have questions, please contact your team leader.
        </div>
    </div>
</body>
</html>`, firstName, serviceName, serviceDate, teamName, acceptURL, declineURL, acceptURL)
}

// sendViaSendGrid sends email via SendGrid API v3
func (ns *NotificationService) sendViaSendGrid(toEmail, subject, htmlBody string) error {
	apiKey := getEnv("SENDGRID_API_KEY")
	if apiKey == "" || len(apiKey) < 10 {
		return fmt.Errorf("SendGrid API key not configured")
	}

	// Build SendGrid request payload
	message := map[string]interface{}{
		"personalizations": []map[string]interface{}{{
			"to": []map[string]string{{"email": toEmail}},
		}},
		"from": map[string]string{
			"email": ns.fromEmail,
			"name":  ns.fromName,
		},
		"subject": subject,
		"content": []map[string]interface{}{{
			"type":  "text/html",
			"value": htmlBody,
		}},
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal SendGrid message: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("SendGrid API returned status %d", resp.StatusCode)
	}

	return nil
}

// UpdateNotificationSent marks an assignment as having notification sent
func (ns *NotificationService) UpdateNotificationSent(ctx context.Context, assignmentID string) error {
	now := time.Now()
	_, err := ns.db.Exec(ctx, `
		UPDATE service_team_assignments 
		SET notification_sent = true, notified_at = $2, updated_at = NOW()
		WHERE id = $1`,
		assignmentID, now)
	return err
}

// getEnv reads an environment variable (wrapper around os.Getenv)
func getEnv(key string) string {
	return os.Getenv(key)
}
