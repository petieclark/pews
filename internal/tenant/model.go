package tenant

import "time"

type Tenant struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Domain        string    `json:"domain,omitempty"`
	Plan          string    `json:"plan"`
	AddressLine1  string    `json:"address_line1,omitempty"`
	AddressLine2  string    `json:"address_line2,omitempty"`
	City          string    `json:"city,omitempty"`
	State         string    `json:"state,omitempty"`
	Zip           string    `json:"zip,omitempty"`
	Phone         string    `json:"phone,omitempty"`
	Website       string    `json:"website,omitempty"`
	Email         string    `json:"email,omitempty"`
	EIN           string    `json:"ein,omitempty"`
	Logo          string    `json:"logo,omitempty"`
	About         string    `json:"about,omitempty"`
	DefaultLocale string    `json:"default_locale"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UpdateProfileRequest struct {
	Name         string `json:"name"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	Zip          string `json:"zip"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
	Email        string `json:"email"`
	EIN          string `json:"ein"`
	About        string `json:"about"`
}
