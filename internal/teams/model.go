package teams

import "time"

type Team struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Color       string     `json:"color"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	MemberCount int        `json:"member_count,omitempty"`
	Positions   []Position `json:"positions,omitempty"`
	Members     []Member   `json:"members,omitempty"`
}

type Position struct {
	ID        string    `json:"id"`
	TeamID    string    `json:"team_id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type Member struct {
	ID         string    `json:"id"`
	TeamID     string    `json:"team_id"`
	PersonID   string    `json:"person_id"`
	PositionID *string   `json:"position_id,omitempty"`
	Status     string    `json:"status"`
	JoinedAt   time.Time `json:"joined_at"`
	// Joined fields
	FirstName    string  `json:"first_name,omitempty"`
	LastName     string  `json:"last_name,omitempty"`
	Email        string  `json:"email,omitempty"`
	PositionName *string `json:"position_name,omitempty"`
}
