package communication

import (
	"fmt"
	"log"
	"os"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
	sg "github.com/sendgrid/sendgrid-go"
)

// SendGridClient sends emails via the SendGrid Go SDK
type SendGridClient struct {
	apiKey    string
	fromEmail string
	fromName  string
}

// NewSendGridClient creates a SendGrid client from environment variables.
// Returns nil if not configured.
func NewSendGridClient() *SendGridClient {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	fromEmail := os.Getenv("SENDGRID_FROM_EMAIL")
	fromName := os.Getenv("SENDGRID_FROM_NAME")

	if apiKey == "" || fromEmail == "" {
		log.Println("[communication] WARNING: SendGrid not configured (SENDGRID_API_KEY / SENDGRID_FROM_EMAIL). Email sending disabled.")
		return nil
	}

	log.Printf("[communication] SendGrid configured with from %s <%s>", fromName, fromEmail)
	return &SendGridClient{apiKey: apiKey, fromEmail: fromEmail, fromName: fromName}
}

// SendEmail sends an email via SendGrid using the Go SDK.
func (s *SendGridClient) SendEmail(to, subject, htmlBody, textBody, fromName string) error {
	if s == nil {
		return fmt.Errorf("sendgrid not configured")
	}

	// Use provided fromName or fall back to default
	displayName := fromName
	if displayName == "" {
		displayName = s.fromName
	}

	from := mail.NewEmail(displayName, s.fromEmail)
	toEmail := mail.NewEmail("", to) // Empty name for recipient
	
	msg := mail.NewSingleEmail(from, subject, toEmail, textBody, htmlBody)

	// Add click tracking (optional but recommended)
	clickTracking := mail.NewClickTrackingSetting()
	trackingSettings := mail.NewTrackingSettings()
	trackingSettings.SetClickTracking(clickTracking)
	msg.SetTrackingSettings(trackingSettings)

	request := sg.API(
		sg.GetRequest(s.apiKey, "https://api.sendgrid.com/v3/mail/send", "api.sendgrid.com"),
	)
	request.Method = "POST"
	request.Payload = msg

	response, err := sg.MakeRequest(request)
	if err != nil {
		return fmt.Errorf("sendgrid failed to send email: %w", err)
	}

	if response.StatusCode != 202 {
		return fmt.Errorf("sendgrid returned status code %d: %s", response.StatusCode, string(response.Body))
	}

	log.Printf("[communication] email sent to %s via SendGrid (status %d)", to, response.StatusCode)
	return nil
}

// IsConfigured returns true if the client is ready to send.
func (s *SendGridClient) IsConfigured() bool {
	return s != nil && s.apiKey != ""
}
