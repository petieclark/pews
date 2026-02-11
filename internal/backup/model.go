package backup

import "time"

type Backup struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	TenantSlug string    `json:"tenant_slug"`
	Filename   string    `json:"filename"`
	SizeBytes  int64     `json:"size_bytes"`
	CreatedAt  time.Time `json:"created_at"`
}

type BackupListResponse struct {
	Backups []Backup `json:"backups"`
	Total   int      `json:"total"`
}
