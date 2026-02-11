package sermons

import "time"

type Sermon struct {
	ID                   string     `json:"id"`
	TenantID             string     `json:"tenant_id"`
	ServiceID            *string    `json:"service_id,omitempty"`
	Title                string     `json:"title"`
	Speaker              string     `json:"speaker"`
	SermonDate           time.Time  `json:"sermon_date"`
	ScriptureReference   string     `json:"scripture_reference,omitempty"`
	NotesText            string     `json:"notes_text,omitempty"`
	AudioURL             string     `json:"audio_url,omitempty"`
	VideoURL             string     `json:"video_url,omitempty"`
	SeriesName           string     `json:"series_name,omitempty"`
	AudioDurationSeconds *int       `json:"audio_duration_seconds,omitempty"`
	Published            bool       `json:"published"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type SermonFilters struct {
	Series      string
	Speaker     string
	DateFrom    string
	DateTo      string
	Published   *bool
	Limit       int
	Offset      int
}
