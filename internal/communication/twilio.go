package communication

import (
	"fmt"
	"log"
	"os"

	twilio "github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

// TwilioClient sends SMS via the Twilio Go SDK
type TwilioClient struct {
	accountSID   string
	authToken    string
	fromNumber   string
	client       *twilio.RestClient
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

	log.Printf("[communication] Twilio configured with account %s..., from %s", sid[:8], from)
	
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username:   sid, // Account SID as username
		Password:   token, // Auth token as password
		AccountSid: sid,
	})
	
	return &TwilioClient{
		accountSID: sid,
		authToken:  token,
		fromNumber: from,
		client:     client,
	}
}

// SendSMS sends an SMS via Twilio using the Go SDK.
func (t *TwilioClient) SendSMS(to, body string) error {
	if t == nil {
		return fmt.Errorf("twilio not configured")
	}

	params := &openapi.CreateMessageParams{
		To:         &to,
		From:       &t.fromNumber,
		MediaUrl:   (*[]string)(nil), // No media attachments for plain SMS
		StatusCallback: (*string)(nil), // Optional callback URL
	}

	message, err := t.client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("twilio failed to send SMS: %w", err)
	}

	log.Printf("[communication] SMS sent to %s via Twilio (SID: %s)", to, *message.Sid)
	return nil
}

// IsConfigured returns true if the client is ready to send.
func (t *TwilioClient) IsConfigured() bool {
	return t != nil && t.client != nil
}
