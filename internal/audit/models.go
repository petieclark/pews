package audit

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID         string          `json:"id" db:"id"`
	TenantID   string          `json:"tenant_id" db:"tenant_id"`
	UserID     *string         `json:"user_id,omitempty" db:"user_id"`
	Timestamp  time.Time       `json:"timestamp" db:"timestamp"`
	Action     string          `json:"action" db:"action"`
	EntityType *string         `json:"entity_type,omitempty" db:"entity_type"`
	EntityID   *string         `json:"entity_id,omitempty" db:"entity_id"`
	IPAddress  *string         `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent  *string         `json:"user_agent,omitempty" db:"user_agent"`
	OldValue   json.RawMessage `json:"old_value,omitempty" db:"old_value"`
	NewValue   json.RawMessage `json:"new_value,omitempty" db:"new_value"`
	Metadata   json.RawMessage `json:"metadata,omitempty" db:"metadata"`
}

type FailedLoginAttempt struct {
	ID          string     `json:"id" db:"id"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	Email       string     `json:"email" db:"email"`
	IPAddress   *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent   *string    `json:"user_agent,omitempty" db:"user_agent"`
	AttemptedAt time.Time  `json:"attempted_at" db:"attempted_at"`
}

type UserSession struct {
	ID           string     `json:"id" db:"id"`
	TenantID     string     `json:"tenant_id" db:"tenant_id"`
	UserID       string     `json:"user_id" db:"user_id"`
	IPAddress    *string    `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string    `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	LastActivity time.Time  `json:"last_activity" db:"last_activity"`
	IsActive     bool       `json:"is_active" db:"is_active"`
}

type SecurityDashboard struct {
	ActiveSessionsCount    int                    `json:"active_sessions_count"`
	FailedLoginsLast24h    int                    `json:"failed_logins_last_24h"`
	UsersWithout2FA        int                    `json:"users_without_2fa"`
	RecentFailedLogins     []FailedLoginAttempt   `json:"recent_failed_logins"`
	UnusualActivities      []UnusualActivity      `json:"unusual_activities"`
	UserPasswordChanges    []UserPasswordChange   `json:"user_password_changes"`
}

type UnusualActivity struct {
	UserID    string    `json:"user_id" db:"user_id"`
	Email     string    `json:"email" db:"email"`
	Action    string    `json:"action" db:"action"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	Reason    string    `json:"reason"`
}

type UserPasswordChange struct {
	UserID            string     `json:"user_id" db:"user_id"`
	Email             string     `json:"email" db:"email"`
	PasswordChangedAt *time.Time `json:"password_changed_at,omitempty" db:"password_changed_at"`
	DaysSinceChange   *int       `json:"days_since_change,omitempty"`
}

// Action constants
const (
	ActionLogin            = "auth.login"
	ActionLoginFailed      = "auth.login_failed"
	ActionLogout           = "auth.logout"
	ActionRegister         = "auth.register"
	ActionPasswordChange   = "auth.password_change"
	ActionPasswordReset    = "auth.password_reset"
	
	ActionCreate           = "create"
	ActionUpdate           = "update"
	ActionDelete           = "delete"
	ActionExport           = "export"
	
	ActionSettingsChange   = "settings.change"
	ActionModuleEnable     = "module.enable"
	ActionModuleDisable    = "module.disable"
	
	ActionBillingChange    = "billing.change"
)

type LogEntry struct {
	Action     string
	EntityType *string
	EntityID   *string
	OldValue   interface{}
	NewValue   interface{}
	Metadata   map[string]interface{}
}
