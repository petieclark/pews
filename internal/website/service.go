package website

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// GetConfig retrieves the website configuration for a tenant
func (s *Service) GetConfig(ctx context.Context, tenantID string) (*Config, error) {
	var settingsJSON []byte
	err := s.db.QueryRow(ctx,
		`SELECT COALESCE(settings->>'website', '{}') FROM tenants WHERE id = $1`,
		tenantID,
	).Scan(&settingsJSON)

	if err != nil {
		return nil, fmt.Errorf("failed to get website config: %w", err)
	}

	config := DefaultConfig()
	if len(settingsJSON) > 0 && string(settingsJSON) != "{}" {
		if err := json.Unmarshal(settingsJSON, config); err != nil {
			return nil, fmt.Errorf("failed to parse website config: %w", err)
		}
	}

	return config, nil
}

// UpdateConfig updates the website configuration for a tenant
func (s *Service) UpdateConfig(ctx context.Context, tenantID string, config *Config) (*Config, error) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	_, err = s.db.Exec(ctx,
		`UPDATE tenants SET settings = COALESCE(settings, '{}'::jsonb) || jsonb_build_object('website', $1::jsonb) WHERE id = $2`,
		configJSON, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update website config: %w", err)
	}

	return config, nil
}

// GetTenantInfo retrieves basic tenant information for website display
func (s *Service) GetTenantInfo(ctx context.Context, tenantSlug string) (string, string, error) {
	var tenantID, tenantName string
	err := s.db.QueryRow(ctx,
		`SELECT id, name FROM tenants WHERE slug = $1`,
		tenantSlug,
	).Scan(&tenantID, &tenantName)

	if err != nil {
		return "", "", fmt.Errorf("tenant not found: %w", err)
	}

	return tenantID, tenantName, nil
}

// GetUpcomingEvents retrieves upcoming events for the website
func (s *Service) GetUpcomingEvents(ctx context.Context, tenantID string, limit int) ([]Event, error) {
	// Check if checkins module is enabled and events exist
	rows, err := s.db.Query(ctx,
		`SELECT id, name, COALESCE(description, ''), start_time, end_time, COALESCE(location, '')
		 FROM checkin_events
		 WHERE tenant_id = $1 AND start_time >= NOW()
		 ORDER BY start_time ASC
		 LIMIT $2`,
		tenantID, limit,
	)
	if err != nil {
		return []Event{}, nil // Return empty array if table doesn't exist or other error
	}
	defer rows.Close()

	events := []Event{}
	for rows.Next() {
		var event Event
		var startTime, endTime time.Time
		if err := rows.Scan(&event.ID, &event.Name, &event.Description, &startTime, &endTime, &event.Location); err != nil {
			continue
		}
		event.StartTime = startTime.Format("2006-01-02T15:04:05Z07:00")
		event.EndTime = endTime.Format("2006-01-02T15:04:05Z07:00")
		events = append(events, event)
	}

	return events, nil
}

// GetLatestSermons retrieves recent sermons for the website
func (s *Service) GetLatestSermons(ctx context.Context, tenantID string, limit int) ([]Sermon, error) {
	// Get recent streams with sermon notes
	rows, err := s.db.Query(ctx,
		`SELECT id, title, COALESCE(scheduled_at, created_at), COALESCE(notes->>'speaker', '')
		 FROM streams
		 WHERE tenant_id = $1 AND notes IS NOT NULL AND notes->>'speaker' IS NOT NULL
		 ORDER BY COALESCE(scheduled_at, created_at) DESC
		 LIMIT $2`,
		tenantID, limit,
	)
	if err != nil {
		return []Sermon{}, nil // Return empty array if table doesn't exist or other error
	}
	defer rows.Close()

	sermons := []Sermon{}
	for rows.Next() {
		var sermon Sermon
		var date time.Time
		if err := rows.Scan(&sermon.ID, &sermon.Title, &date, &sermon.Speaker); err != nil {
			continue
		}
		sermon.Date = date.Format("January 2, 2006")
		sermon.StreamID = sermon.ID
		sermons = append(sermons, sermon)
	}

	return sermons, nil
}
