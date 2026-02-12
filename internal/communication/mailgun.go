package communication

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// MailgunClient sends emails via the Mailgun HTTP API
type MailgunClient struct {
	apiKey string
	domain string
}

// NewMailgunClient creates a Mailgun client from environment variables.
// Returns nil if not configured.
func NewMailgunClient() *MailgunClient {
	apiKey := os.Getenv("MAILGUN_API_KEY")
	domain := os.Getenv("MAILGUN_DOMAIN")

	if apiKey == "" || domain == "" {
		log.Println("[communication] WARNING: Mailgun not configured (MAILGUN_API_KEY / MAILGUN_DOMAIN). Email sending disabled.")
		return nil
	}

	log.Printf("[communication] Mailgun configured for domain %s", domain)
	return &MailgunClient{apiKey: apiKey, domain: domain}
}

// SendEmail sends an email via Mailgun HTTP API.
func (m *MailgunClient) SendEmail(to, subject, htmlBody, textBody, fromName string) error {
	if m == nil {
		return fmt.Errorf("mailgun not configured")
	}

	fromAddr := fmt.Sprintf("%s <noreply@%s>", fromName, m.domain)

	data := url.Values{}
	data.Set("from", fromAddr)
	data.Set("to", to)
	data.Set("subject", subject)
	if htmlBody != "" {
		data.Set("html", htmlBody)
	}
	if textBody != "" {
		data.Set("text", textBody)
	}

	endpoint := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", m.domain)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", m.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("mailgun request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailgun returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// IsConfigured returns true if the client is ready to send.
func (m *MailgunClient) IsConfigured() bool {
	return m != nil
}
