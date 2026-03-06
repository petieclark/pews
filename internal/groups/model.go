package groups

import (
	"time"

	"github.com/petieclark/pews/internal/people"
)

type Group struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description,omitempty"`
	GroupType       string    `json:"group_type"`
	MeetingDay      string    `json:"meeting_day,omitempty"`
	MeetingTime     string    `json:"meeting_time,omitempty"`
	MeetingLocation string    `json:"meeting_location,omitempty"`
	IsPublic        bool      `json:"is_public"`
	MaxMembers      *int      `json:"max_members,omitempty"`
	IsActive        bool      `json:"is_active"`
	PhotoURL        string    `json:"photo_url,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	MemberCount     int       `json:"member_count,omitempty"`
	Members         []Member  `json:"members,omitempty"`
}

type Member struct {
	ID       string         `json:"id"`
	GroupID  string         `json:"group_id"`
	PersonID string         `json:"person_id"`
	Role     string         `json:"role"`
	JoinedAt time.Time      `json:"joined_at"`
	Person   *people.Person `json:"person,omitempty"`
}

type JoinRequest struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	GroupID   string    `json:"group_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	Message   string    `json:"message,omitempty"`
	Status    string    `json:"status"` // pending, approved, declined
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	GroupName string    `json:"group_name,omitempty"`
}
