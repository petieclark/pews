package tenant

import "time"

type Tenant struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Domain        string    `json:"domain,omitempty"`
	Plan          string    `json:"plan"`
	DefaultLocale string    `json:"default_locale"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
