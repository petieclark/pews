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

// TwilioClient sends SMS via the Twilio REST API
type TwilioClient struct {
	accountSID string
	authToken  string
	fromNumber string
}

// NewTwilioClient creates a Twilio client from environment variables.
// Returns nil if not configured.
func NewTwilioClient() *TwilioClient {
	sid := os.Getenv("TWILIO_ACCOUNT_SID")
	token := os.Getenv("TWILIO_AUTH_TOKEN")
	from := os.Getenv("TWILIO_FROM_NUMBER")

	if sid == "" || token == "" || from == "" {
		log.Println("[communication] WARNING: Twilio not configured (TWILIO_ACCOUNT_SID / TWILIO_AUTH_TOKEN / TWILIO_FROM_NUMBER). SMS sending disabled.")
		return nil
	}

	log.Printf("[communication] Twilio configured with account %s, from %s", sid[:8]+"...", from)
	return &TwilioClient{accountSID: sid, authToken: token, fromNumber: from}
}

// SendSMS sends an SMS via Twilio REST API.
func (t *TwilioClient) SendSMS(to, body string) error {
	if t == nil {
		return fmt.Errorf("twilio not configured")
	}

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.accountSID)

	data := url.Values{}
	data.Set("To", to)
	data.Set("From", t.fromNumber)
	data.Set("Body", body)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(t.accountSID, t.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("twilio request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twilio returned %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// IsConfigured returns true if the client is ready to send.
func (t *TwilioClient) IsConfigured() bool {
	return t != nil
}
