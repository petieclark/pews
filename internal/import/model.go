package importpkg

import (
	"encoding/json"
)

// ImportPeopleRequest is the request body for bulk person import
type ImportPeopleRequest struct {
	People []PersonImport `json:"people"`
	DryRun bool           `json:"dry_run"`
}

// PersonImport represents a person to be imported
type PersonImport struct {
	FirstName        string          `json:"first_name"`
	LastName         string          `json:"last_name"`
	Email            string          `json:"email,omitempty"`
	Phone            string          `json:"phone,omitempty"`
	AddressLine1     string          `json:"address_line1,omitempty"`
	AddressLine2     string          `json:"address_line2,omitempty"`
	City             string          `json:"city,omitempty"`
	State            string          `json:"state,omitempty"`
	Zip              string          `json:"zip,omitempty"`
	Birthdate        string          `json:"birthdate,omitempty"`
	Gender           string          `json:"gender,omitempty"`
	MembershipStatus string          `json:"membership_status,omitempty"`
	PhotoURL         string          `json:"photo_url,omitempty"`
	Notes            string          `json:"notes,omitempty"`
	CustomFields     json.RawMessage `json:"custom_fields,omitempty"`
}

// ImportGroupsRequest is the request body for bulk group import
type ImportGroupsRequest struct {
	Groups []GroupImport `json:"groups"`
	DryRun bool          `json:"dry_run"`
}

// GroupImport represents a group to be imported
type GroupImport struct {
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	Type            string   `json:"type,omitempty"`
	MeetingDay      string   `json:"meeting_day,omitempty"`
	MeetingTime     string   `json:"meeting_time,omitempty"`
	MeetingLocation string   `json:"meeting_location,omitempty"`
	IsPublic        bool     `json:"is_public"`
	MaxMembers      *int     `json:"max_members,omitempty"`
	Members         []string `json:"members,omitempty"` // Array of email addresses
}

// ImportSongsRequest is the request body for bulk song import
type ImportSongsRequest struct {
	Songs  []SongImport `json:"songs"`
	DryRun bool         `json:"dry_run"`
}

// SongImport represents a song to be imported
type SongImport struct {
	Title      string `json:"title"`
	Artist     string `json:"artist,omitempty"`
	Key        string `json:"key,omitempty"`
	Tempo      int    `json:"tempo,omitempty"`
	CCLINumber string `json:"ccli_number,omitempty"`
	Lyrics     string `json:"lyrics,omitempty"`
	Notes      string `json:"notes,omitempty"`
	Tags       string `json:"tags,omitempty"`
}

// ImportGivingRequest is the request body for bulk donation import
type ImportGivingRequest struct {
	Donations []DonationImport `json:"donations"`
	DryRun    bool             `json:"dry_run"`
}

// DonationImport represents a donation to be imported
type DonationImport struct {
	DonorEmail      string  `json:"donor_email"`
	FundName        string  `json:"fund_name"`
	AmountCents     int     `json:"amount_cents"`
	Currency        string  `json:"currency,omitempty"`
	PaymentMethod   string  `json:"payment_method,omitempty"`
	Memo            string  `json:"memo,omitempty"`
	DonatedAt       string  `json:"donated_at,omitempty"` // ISO 8601 format
}

// ImportResult is the response for all import operations
type ImportResult struct {
	Created int      `json:"created"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}
