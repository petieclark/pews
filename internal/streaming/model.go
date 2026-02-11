package streaming

import (
	"time"
)

type Stream struct {
	ID                    string     `json:"id"`
	TenantID              string     `json:"tenant_id"`
	Title                 string     `json:"title"`
	Description           string     `json:"description,omitempty"`
	ServiceID             *string    `json:"service_id,omitempty"`
	Status                string     `json:"status"` // scheduled, live, ended, archived
	ScheduledStart        *time.Time `json:"scheduled_start,omitempty"`
	ActualStart           *time.Time `json:"actual_start,omitempty"`
	ActualEnd             *time.Time `json:"actual_end,omitempty"`
	StreamType            string     `json:"stream_type"` // youtube, facebook, vimeo, rtmp_custom
	StreamURL             string     `json:"stream_url,omitempty"`
	StreamKey             string     `json:"stream_key,omitempty"`
	EmbedURL              string     `json:"embed_url"`
	ChatEnabled           bool       `json:"chat_enabled"`
	GivingEnabled         bool       `json:"giving_enabled"`
	ConnectionCardEnabled bool       `json:"connection_card_enabled"`
	ViewerCount           int        `json:"viewer_count"`
	PeakViewers           int        `json:"peak_viewers"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type ChatMessage struct {
	ID         string     `json:"id"`
	StreamID   string     `json:"stream_id"`
	PersonID   *string    `json:"person_id,omitempty"`
	GuestName  string     `json:"guest_name,omitempty"`
	Message    string     `json:"message"`
	IsPinned   bool       `json:"is_pinned"`
	IsDeleted  bool       `json:"is_deleted"`
	CreatedAt  time.Time  `json:"created_at"`
}

type StreamViewer struct {
	ID              string     `json:"id"`
	StreamID        string     `json:"stream_id"`
	PersonID        *string    `json:"person_id,omitempty"`
	GuestName       string     `json:"guest_name,omitempty"`
	JoinedAt        time.Time  `json:"joined_at"`
	LeftAt          *time.Time `json:"left_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`
}

type StreamNote struct {
	ID        string    `json:"id"`
	StreamID  string    `json:"stream_id"`
	PersonID  *string   `json:"person_id,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
