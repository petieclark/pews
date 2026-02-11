package sms

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TwilioClient handles Twilio API interactions
type TwilioClient struct {
	accountSID string
	authToken  string
	fromNumber string
	httpClient *http.Client
}

// NewTwilioClient creates a new Twilio client
func NewTwilioClient(accountSID, authToken, fromNumber string) *TwilioClient {
	return &TwilioClient{
		accountSID: accountSID,
		authToken:  authToken,
		fromNumber: fromNumber,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TwilioMessageResponse represents the Twilio API response
type TwilioMessageResponse struct {
	SID          string `json:"sid"`
	Status       string `json:"status"`
	ErrorCode    int    `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// SendSMS sends an SMS via Twilio REST API
func (c *TwilioClient) SendSMS(to, body string) (*TwilioMessageResponse, error) {
	// Construct API URL
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", c.accountSID)

	// Prepare form data
	data := url.Values{}
	data.Set("To", to)
	data.Set("From", c.fromNumber)
	data.Set("Body", body)

	// Create request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.accountSID, c.authToken)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var twilioResp TwilioMessageResponse
	if err := json.Unmarshal(bodyBytes, &twilioResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		return &twilioResp, fmt.Errorf("twilio error %d: %s", twilioResp.ErrorCode, twilioResp.ErrorMessage)
	}

	return &twilioResp, nil
}

// TestConnection tests the Twilio credentials by sending a validation request
func (c *TwilioClient) TestConnection() error {
	// Test by fetching account info
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s.json", c.accountSID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.SetBasicAuth(c.accountSID, c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Twilio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("twilio authentication failed (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// RateLimiter implements simple rate limiting (1 msg/second)
type RateLimiter struct {
	ticker *time.Ticker
	tokens chan struct{}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(perSecond int) *RateLimiter {
	rl := &RateLimiter{
		ticker: time.NewTicker(time.Second / time.Duration(perSecond)),
		tokens: make(chan struct{}, perSecond),
	}

	// Fill initial tokens
	for i := 0; i < perSecond; i++ {
		rl.tokens <- struct{}{}
	}

	// Refill tokens
	go func() {
		for range rl.ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
			}
		}
	}()

	return rl
}

// Wait blocks until a token is available
func (rl *RateLimiter) Wait() {
	<-rl.tokens
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	rl.ticker.Stop()
}
