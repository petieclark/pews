package search

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

type PersonResult struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

type SongResult struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist,omitempty"`
}

type GroupResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ServiceResult struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Notes string `json:"notes,omitempty"`
	Date  string `json:"date"`
}

type SearchResults struct {
	People   []PersonResult  `json:"people"`
	Songs    []SongResult    `json:"songs"`
	Groups   []GroupResult   `json:"groups"`
	Services []ServiceResult `json:"services"`
}

func (s *Service) Search(ctx context.Context, tenantID string, query string) (*SearchResults, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	results := &SearchResults{
		People:   []PersonResult{},
		Songs:    []SongResult{},
		Groups:   []GroupResult{},
		Services: []ServiceResult{},
	}

	searchPattern := "%" + query + "%"

	// Search People (name, email)
	peopleRows, err := s.db.Query(ctx, `
		SELECT id, COALESCE(first_name, '') || ' ' || COALESCE(last_name, '') as name, COALESCE(email, '')
		FROM people
		WHERE first_name ILIKE $1 
		   OR last_name ILIKE $1 
		   OR email ILIKE $1
		ORDER BY first_name, last_name
		LIMIT 5
	`, searchPattern)

	if err == nil {
		defer peopleRows.Close()
		for peopleRows.Next() {
			var p PersonResult
			if err := peopleRows.Scan(&p.ID, &p.Name, &p.Email); err == nil {
				results.People = append(results.People, p)
			}
		}
	}

	// Search Songs (title, artist, CCLI)
	songRows, err := s.db.Query(ctx, `
		SELECT id, title, COALESCE(artist, '')
		FROM songs
		WHERE title ILIKE $1 
		   OR artist ILIKE $1
		   OR ccli_number ILIKE $1
		ORDER BY title
		LIMIT 5
	`, searchPattern)

	if err == nil {
		defer songRows.Close()
		for songRows.Next() {
			var song SongResult
			if err := songRows.Scan(&song.ID, &song.Title, &song.Artist); err == nil {
				results.Songs = append(results.Songs, song)
			}
		}
	}

	// Search Groups (name)
	groupRows, err := s.db.Query(ctx, `
		SELECT id, name
		FROM groups
		WHERE name ILIKE $1
		ORDER BY name
		LIMIT 5
	`, searchPattern)

	if err == nil {
		defer groupRows.Close()
		for groupRows.Next() {
			var g GroupResult
			if err := groupRows.Scan(&g.ID, &g.Name); err == nil {
				results.Groups = append(results.Groups, g)
			}
		}
	}

	// Search Services (notes)
	serviceRows, err := s.db.Query(ctx, `
		SELECT s.id, 
		       COALESCE(s.name, st.name, '') as name,
		       COALESCE(s.notes, ''),
		       TO_CHAR(s.service_date, 'YYYY-MM-DD') as date
		FROM church_services s
		LEFT JOIN service_types st ON st.id = s.service_type_id
		WHERE s.notes ILIKE $1
		   OR s.name ILIKE $1
		   OR st.name ILIKE $1
		ORDER BY s.service_date DESC
		LIMIT 5
	`, searchPattern)

	if err == nil {
		defer serviceRows.Close()
		for serviceRows.Next() {
			var svc ServiceResult
			if err := serviceRows.Scan(&svc.ID, &svc.Name, &svc.Notes, &svc.Date); err == nil {
				results.Services = append(results.Services, svc)
			}
		}
	}

	return results, nil
}
