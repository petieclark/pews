package module

import "time"

type Module struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Available   bool   `json:"available"`
}

type TenantModule struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenant_id"`
	ModuleName string     `json:"module_name"`
	Enabled    bool       `json:"enabled"`
	EnabledAt  *time.Time `json:"enabled_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
