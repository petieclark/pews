package people

import (
	"encoding/json"
	"time"
)

type Person struct {
	ID               string          `json:"id"`
	TenantID         string          `json:"tenant_id"`
	FirstName        string          `json:"first_name"`
	LastName         string          `json:"last_name"`
	Email            string          `json:"email,omitempty"`
	Phone            string          `json:"phone,omitempty"`
	AddressLine1     string          `json:"address_line1,omitempty"`
	AddressLine2     string          `json:"address_line2,omitempty"`
	City             string          `json:"city,omitempty"`
	State            string          `json:"state,omitempty"`
	Zip              string          `json:"zip,omitempty"`
	Birthdate        *time.Time      `json:"birthdate,omitempty"`
	Gender           string          `json:"gender,omitempty"`
	MembershipStatus string          `json:"membership_status"`
	PhotoURL         string          `json:"photo_url,omitempty"`
	Notes            string          `json:"notes,omitempty"`
	CustomFields     json.RawMessage `json:"custom_fields,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	Household        *Household      `json:"household,omitempty"`
	Tags             []Tag           `json:"tags,omitempty"`
}

type Household struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	Name             string     `json:"name"`
	PrimaryContactID string     `json:"primary_contact_id,omitempty"`
	AddressLine1     string     `json:"address_line1,omitempty"`
	AddressLine2     string     `json:"address_line2,omitempty"`
	City             string     `json:"city,omitempty"`
	State            string     `json:"state,omitempty"`
	Zip              string     `json:"zip,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	Members          []Person   `json:"members,omitempty"`
}

type HouseholdMember struct {
	HouseholdID string `json:"household_id"`
	PersonID    string `json:"person_id"`
	Role        string `json:"role"`
}

type Tag struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	PersonCount int       `json:"person_count,omitempty"`
}
