package care

import "time"

type FollowUp struct {
	ID           string     `json:"id"`
	TenantID     string     `json:"tenant_id"`
	PersonID     string     `json:"person_id"`
	PersonName   string     `json:"person_name"`
	AssignedTo   *string    `json:"assigned_to"`
	AssignedName *string    `json:"assigned_name,omitempty"`
	Title        string     `json:"title"`
	Type         string     `json:"type"`
	Priority     string     `json:"priority"`
	Status       string     `json:"status"`
	DueDate      *string    `json:"due_date"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Notes        []Note     `json:"notes,omitempty"`
}

type Note struct {
	ID         string    `json:"id"`
	FollowUpID string    `json:"follow_up_id"`
	AuthorID   *string   `json:"author_id"`
	AuthorName string    `json:"author_name"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateFollowUpInput struct {
	PersonID   string  `json:"person_id"`
	AssignedTo *string `json:"assigned_to"`
	Title      string  `json:"title"`
	Type       string  `json:"type"`
	Priority   string  `json:"priority"`
	DueDate    *string `json:"due_date"`
	Notes      string  `json:"notes"`
}

type UpdateFollowUpInput struct {
	Title      *string `json:"title"`
	Type       *string `json:"type"`
	Priority   *string `json:"priority"`
	Status     *string `json:"status"`
	AssignedTo *string `json:"assigned_to"`
	DueDate    *string `json:"due_date"`
}

type Stats struct {
	NewCount            int `json:"new_count"`
	InProgressCount     int `json:"in_progress_count"`
	WaitingCount        int `json:"waiting_count"`
	CompletedCount      int `json:"completed_count"`
	OverdueCount        int `json:"overdue_count"`
	DueTodayCount       int `json:"due_today_count"`
	DueThisWeekCount    int `json:"due_this_week_count"`
}
