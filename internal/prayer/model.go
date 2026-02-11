package prayer

import (
	"time"
)

type PrayerRequest struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	PersonID         *string    `json:"person_id"`
	PersonName       *string    `json:"person_name,omitempty"`
	Name             string     `json:"name"`
	Email            *string    `json:"email"`
	RequestText      string     `json:"request_text"`
	IsPublic         bool       `json:"is_public"`
	Status           string     `json:"status"` // pending, praying, answered, archived
	ConnectionCardID *string    `json:"connection_card_id,omitempty"`
	Notes            *string    `json:"notes"`
	SubmittedAt      time.Time  `json:"submitted_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	FollowerCount    *int       `json:"follower_count,omitempty"`
	IsFollowing      *bool      `json:"is_following,omitempty"`
}

type PrayerFollower struct {
	ID               string    `json:"id"`
	PrayerRequestID  string    `json:"prayer_request_id"`
	UserID           string    `json:"user_id"`
	UserName         *string   `json:"user_name,omitempty"`
	FollowedAt       time.Time `json:"followed_at"`
}

type CreatePrayerRequestInput struct {
	Name        string  `json:"name" binding:"required"`
	Email       *string `json:"email"`
	RequestText string  `json:"request_text" binding:"required"`
	IsPublic    bool    `json:"is_public"`
}

type UpdatePrayerRequestInput struct {
	Status string  `json:"status"`
	Notes  *string `json:"notes"`
}

type PrayerRequestFilter struct {
	Status   *string `form:"status"`
	IsPublic *bool   `form:"is_public"`
	Limit    int     `form:"limit"`
	Offset   int     `form:"offset"`
}
