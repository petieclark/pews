package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// AssignmentToken represents a signed token for volunteer assignment responses
type AssignmentToken struct {
	AssignmentID string    `json:"assignment_id"`
	PersonID     string    `json:"person_id"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// TokenService handles signing and validating assignment response tokens
type TokenService struct {
	secretKey []byte
	expiry    time.Duration
}

// NewTokenService creates a new token service with the given secret and expiry
func NewTokenService(secretKey string, expiryHours int) *TokenService {
	return &TokenService{
		secretKey: []byte(secretKey),
		expiry:    time.Duration(expiryHours) * time.Hour,
	}
}

// Generate creates a signed token for an assignment response link
func (ts *TokenService) Generate(asignmentID, personID string) (string, error) {
	tokenData := AssignmentToken{
		AssignmentID: asignmentID,
		PersonID:     personID,
		ExpiresAt:    time.Now().Add(ts.expiry),
	}

	jsonBytes, err := json.Marshal(tokenData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token data: %w", err)
	}

	// Create HMAC signature
	h := hmac.New(sha256.New, ts.secretKey)
	h.Write(jsonBytes)
	signature := h.Sum(nil)

	// Base64 encode both payload and signature
	payload := base64.URLEncoding.EncodeToString(jsonBytes)
	sig := base64.URLEncoding.EncodeToString(signature)

	return fmt.Sprintf("%s.%s", payload, sig), nil
}

// Validate parses and validates a token, returning the claims if valid
func (ts *TokenService) Validate(tokenString string) (*AssignmentToken, error) {
	parts := splitToken(tokenString)
	if len(parts) != 2 {
		return nil, errors.New("invalid token format")
	}

	payload, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	signature, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify signature
	h := hmac.New(sha256.New, ts.secretKey)
	h.Write(payload)
	expectedSig := h.Sum(nil)

	if !hmac.Equal(signature, expectedSig) {
		return nil, errors.New("invalid token signature")
	}

	// Parse payload
	var claims AssignmentToken
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	// Check expiry
	if time.Now().After(claims.ExpiresAt) {
		return nil, errors.New("token has expired")
	}

	return &claims, nil
}

func splitToken(token string) []string {
	parts := make([]string, 0, 2)
	current := ""
	inQuotes := false
	
	for _, c := range token {
		if c == '"' {
			inQuotes = !inQuotes
			current += string(c)
		} else if c == '.' && !inQuotes {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}
