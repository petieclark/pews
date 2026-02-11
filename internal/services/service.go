package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// ServiceType operations

func (s *Service) ListServiceTypes(ctx context.Context, tenantID string) ([]ServiceType, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, tenant_id, name, COALESCE(default_time, ''), COALESCE(default_day, ''), 
		       COALESCE(color, '#4A8B8C'), is_active, created_at, updated_at
		FROM service_types
		ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to list service types: %w", err)
	}
	defer rows.Close()

	types := []ServiceType{}
	for rows.Next() {
		var st ServiceType
		err := rows.Scan(&st.ID, &st.TenantID, &st.Name, &st.DefaultTime, &st.DefaultDay,
			&st.Color, &st.IsActive, &st.CreatedAt, &st.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service type: %w", err)
		}
		types = append(types, st)
	}

	return types, nil
}

func (s *Service) CreateServiceType(ctx context.Context, tenantID string, st *ServiceType) (*ServiceType, error) {
	st.ID = uuid.New().String()
	st.TenantID = tenantID

	err := s.db.QueryRow(ctx, `
		INSERT INTO service_types (id, tenant_id, name, default_time, default_day, color, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`,
		st.ID, st.TenantID, st.Name, st.DefaultTime, st.DefaultDay, st.Color, st.IsActive,
	).Scan(&st.CreatedAt, &st.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create service type: %w", err)
	}

	return st, nil
}

func (s *Service) UpdateServiceType(ctx context.Context, tenantID, typeID string, st *ServiceType) (*ServiceType, error) {
	err := s.db.QueryRow(ctx, `
		UPDATE service_types SET 
			name = $1, default_time = $2, default_day = $3, color = $4, is_active = $5
		WHERE id = $6
		RETURNING created_at, updated_at`,
		st.Name, st.DefaultTime, st.DefaultDay, st.Color, st.IsActive, typeID,
	).Scan(&st.CreatedAt, &st.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("service type not found")
		}
		return nil, fmt.Errorf("failed to update service type: %w", err)
	}

	st.ID = typeID
	st.TenantID = tenantID

	return st, nil
}

// Services operations

func (s *Service) ListServices(ctx context.Context, tenantID, fromDate, toDate, typeID, status string, page, limit int) ([]ChurchService, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	if fromDate != "" {
		whereClause += fmt.Sprintf(" AND service_date >= $%d", argPos)
		args = append(args, fromDate)
		argPos++
	}

	if toDate != "" {
		whereClause += fmt.Sprintf(" AND service_date <= $%d", argPos)
		args = append(args, toDate)
		argPos++
	}

	if typeID != "" {
		whereClause += fmt.Sprintf(" AND service_type_id = $%d", argPos)
		args = append(args, typeID)
		argPos++
	}

	if status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM services " + whereClause
	err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count services: %w", err)
	}

	// Get services
	query := fmt.Sprintf(`
		SELECT s.id, s.tenant_id, s.service_type_id, COALESCE(s.name, ''), s.service_date, 
		       COALESCE(s.service_time, ''), COALESCE(s.notes, ''), COALESCE(s.status, 'draft'), s.created_at, s.updated_at,
		       COALESCE(st.name, ''), COALESCE(st.color, '#4A8B8C')
		FROM services s
		JOIN service_types st ON st.id = s.service_type_id
		%s
		ORDER BY s.service_date DESC, s.service_time
		LIMIT $%d OFFSET $%d`, whereClause, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list services: %w", err)
	}
	defer rows.Close()

	svcs := []ChurchService{}
	for rows.Next() {
		var svc ChurchService
		var stName, stColor string
		err := rows.Scan(&svc.ID, &svc.TenantID, &svc.ServiceTypeID, &svc.Name, &svc.ServiceDate,
			&svc.ServiceTime, &svc.Notes, &svc.Status, &svc.CreatedAt, &svc.UpdatedAt,
			&stName, &stColor)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan service: %w", err)
		}

		svc.ServiceType = &ServiceType{Name: stName, Color: stColor}

		// Count team members
		var teamCount int
		s.db.QueryRow(ctx, "SELECT COUNT(*) FROM service_teams WHERE service_id = $1", svc.ID).Scan(&teamCount)
		
		svcs = append(svcs, svc)
	}

	return svcs, total, nil
}

func (s *Service) GetUpcomingServices(ctx context.Context, tenantID string, limit int) ([]ChurchService, error) {
	if limit < 1 || limit > 20 {
		limit = 4
	}

	today := time.Now().Format("2006-01-02")

	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.tenant_id, s.service_type_id, COALESCE(s.name, ''), s.service_date, 
		       COALESCE(s.service_time, ''), COALESCE(s.notes, ''), COALESCE(s.status, 'draft'), s.created_at, s.updated_at,
		       COALESCE(st.name, ''), COALESCE(st.color, '#4A8B8C')
		FROM services s
		JOIN service_types st ON st.id = s.service_type_id
		WHERE s.service_date >= $1
		ORDER BY s.service_date ASC, s.service_time
		LIMIT $2`, today, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming services: %w", err)
	}
	defer rows.Close()

	svcs := []ChurchService{}
	for rows.Next() {
		var svc ChurchService
		var stName, stColor string
		err := rows.Scan(&svc.ID, &svc.TenantID, &svc.ServiceTypeID, &svc.Name, &svc.ServiceDate,
			&svc.ServiceTime, &svc.Notes, &svc.Status, &svc.CreatedAt, &svc.UpdatedAt,
			&stName, &stColor)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}

		svc.ServiceType = &ServiceType{Name: stName, Color: stColor}
		svcs = append(svcs, svc)
	}

	return svcs, nil
}

func (s *Service) GetServiceByID(ctx context.Context, tenantID, serviceID string) (*ChurchService, error) {
	var svc ChurchService
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, service_type_id, COALESCE(name, ''), service_date, COALESCE(service_time, ''), COALESCE(notes, ''), COALESCE(status, 'draft'), created_at, updated_at
		FROM services WHERE id = $1`, serviceID).Scan(
		&svc.ID, &svc.TenantID, &svc.ServiceTypeID, &svc.Name, &svc.ServiceDate,
		&svc.ServiceTime, &svc.Notes, &svc.Status, &svc.CreatedAt, &svc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("service not found")
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	// Load service type
	var st ServiceType
	err2 := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, name, COALESCE(default_time, ''), COALESCE(default_day, ''), COALESCE(color, '#4A8B8C'), is_active, created_at, updated_at
		FROM service_types WHERE id = $1`, svc.ServiceTypeID).Scan(
		&st.ID, &st.TenantID, &st.Name, &st.DefaultTime, &st.DefaultDay,
		&st.Color, &st.IsActive, &st.CreatedAt, &st.UpdatedAt,
	)
	if err2 == nil {
		svc.ServiceType = &st
	}

	// Load items
	items, err2 := s.GetServiceItems(ctx, tenantID, serviceID)
	if err2 == nil {
		svc.Items = items
	}

	// Load team
	team, err2 := s.GetServiceTeam(ctx, tenantID, serviceID)
	if err2 == nil {
		svc.Team = team
	}

	return &svc, nil
}

func (s *Service) CreateService(ctx context.Context, tenantID string, svc *ChurchService) (*ChurchService, error) {
	svc.ID = uuid.New().String()
	svc.TenantID = tenantID

	err := s.db.QueryRow(ctx, `
		INSERT INTO services (id, tenant_id, service_type_id, name, service_date, service_time, notes, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`,
		svc.ID, svc.TenantID, svc.ServiceTypeID, svc.Name, svc.ServiceDate, svc.ServiceTime, svc.Notes, svc.Status,
	).Scan(&svc.CreatedAt, &svc.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return svc, nil
}

func (s *Service) UpdateService(ctx context.Context, tenantID, serviceID string, svc *ChurchService) (*ChurchService, error) {
	err := s.db.QueryRow(ctx, `
		UPDATE services SET 
			service_type_id = $1, name = $2, service_date = $3, service_time = $4, notes = $5, status = $6
		WHERE id = $7
		RETURNING created_at, updated_at`,
		svc.ServiceTypeID, svc.Name, svc.ServiceDate, svc.ServiceTime, svc.Notes, svc.Status, serviceID,
	).Scan(&svc.CreatedAt, &svc.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("service not found")
		}
		return nil, fmt.Errorf("failed to update service: %w", err)
	}

	svc.ID = serviceID
	svc.TenantID = tenantID

	return svc, nil
}

func (s *Service) DeleteService(ctx context.Context, tenantID, serviceID string) error {
	result, err := s.db.Exec(ctx, "DELETE FROM services WHERE id = $1", serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("service not found")
	}

	return nil
}

// Service Items operations

func (s *Service) GetServiceItems(ctx context.Context, tenantID, serviceID string) ([]ServiceItem, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, service_id, item_type, title, song_id, COALESCE(song_key, ''), position, duration_minutes, COALESCE(notes, ''), COALESCE(assigned_to, '')
		FROM service_items
		WHERE service_id = $1
		ORDER BY position`, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get service items: %w", err)
	}
	defer rows.Close()

	items := []ServiceItem{}
	for rows.Next() {
		var item ServiceItem
		err := rows.Scan(&item.ID, &item.ServiceID, &item.ItemType, &item.Title, &item.SongID,
			&item.SongKey, &item.Position, &item.DurationMinutes, &item.Notes, &item.AssignedTo)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service item: %w", err)
		}

		// Load song if present
		if item.SongID != nil {
			song, err := s.GetSongByID(ctx, tenantID, *item.SongID)
			if err == nil {
				item.Song = song
			}
		}

		items = append(items, item)
	}

	return items, nil
}

func (s *Service) AddServiceItem(ctx context.Context, tenantID string, item *ServiceItem) (*ServiceItem, error) {
	item.ID = uuid.New().String()

	_, err := s.db.Exec(ctx, `
		INSERT INTO service_items (id, service_id, item_type, title, song_id, song_key, position, duration_minutes, notes, assigned_to)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		item.ID, item.ServiceID, item.ItemType, item.Title, item.SongID, item.SongKey, item.Position, item.DurationMinutes, item.Notes, item.AssignedTo,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to add service item: %w", err)
	}

	// Update song usage if song_id is present
	if item.SongID != nil {
		s.db.Exec(ctx, `
			UPDATE songs SET times_used = times_used + 1, last_used = CURRENT_DATE 
			WHERE id = $1`, *item.SongID)
	}

	return item, nil
}

func (s *Service) UpdateServiceItem(ctx context.Context, tenantID, itemID string, item *ServiceItem) (*ServiceItem, error) {
	result, err := s.db.Exec(ctx, `
		UPDATE service_items SET 
			item_type = $1, title = $2, song_id = $3, song_key = $4, position = $5, duration_minutes = $6, notes = $7, assigned_to = $8
		WHERE id = $9`,
		item.ItemType, item.Title, item.SongID, item.SongKey, item.Position, item.DurationMinutes, item.Notes, item.AssignedTo, itemID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update service item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("service item not found")
	}

	item.ID = itemID

	return item, nil
}

func (s *Service) DeleteServiceItem(ctx context.Context, tenantID, itemID string) error {
	result, err := s.db.Exec(ctx, "DELETE FROM service_items WHERE id = $1", itemID)
	if err != nil {
		return fmt.Errorf("failed to delete service item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("service item not found")
	}

	return nil
}

// Service Team operations

func (s *Service) GetServiceTeam(ctx context.Context, tenantID, serviceID string) ([]ServiceTeam, error) {
	rows, err := s.db.Query(ctx, `
		SELECT st.id, st.service_id, st.person_id, st.role, st.status, COALESCE(st.notes, ''),
		       p.first_name, p.last_name
		FROM service_teams st
		JOIN people p ON p.id = st.person_id
		WHERE st.service_id = $1
		ORDER BY st.role`, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get service team: %w", err)
	}
	defer rows.Close()

	team := []ServiceTeam{}
	for rows.Next() {
		var member ServiceTeam
		err := rows.Scan(&member.ID, &member.ServiceID, &member.PersonID, &member.Role, &member.Status, &member.Notes,
			&member.PersonFirstName, &member.PersonLastName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service team member: %w", err)
		}
		team = append(team, member)
	}

	return team, nil
}

func (s *Service) AddServiceTeamMember(ctx context.Context, tenantID string, member *ServiceTeam) (*ServiceTeam, error) {
	member.ID = uuid.New().String()

	_, err := s.db.Exec(ctx, `
		INSERT INTO service_teams (id, service_id, person_id, role, status, notes)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		member.ID, member.ServiceID, member.PersonID, member.Role, member.Status, member.Notes,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to add service team member: %w", err)
	}

	return member, nil
}

func (s *Service) UpdateServiceTeamMember(ctx context.Context, tenantID, teamID string, member *ServiceTeam) (*ServiceTeam, error) {
	result, err := s.db.Exec(ctx, `
		UPDATE service_teams SET role = $1, status = $2, notes = $3
		WHERE id = $4`,
		member.Role, member.Status, member.Notes, teamID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update service team member: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("service team member not found")
	}

	member.ID = teamID

	return member, nil
}

func (s *Service) DeleteServiceTeamMember(ctx context.Context, tenantID, teamID string) error {
	result, err := s.db.Exec(ctx, "DELETE FROM service_teams WHERE id = $1", teamID)
	if err != nil {
		return fmt.Errorf("failed to delete service team member: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("service team member not found")
	}

	return nil
}

// Songs operations

func (s *Service) ListSongs(ctx context.Context, tenantID, query string, page, limit int) ([]Song, int, error) {
	return s.ListSongsFiltered(ctx, tenantID, SongListParams{
		Query: query,
		Page:  page,
		Limit: limit,
	})
}

func (s *Service) ListSongsFiltered(ctx context.Context, tenantID string, params SongListParams) ([]Song, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 20
	}
	offset := (params.Page - 1) * params.Limit

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	if params.Query != "" {
		whereClause += fmt.Sprintf(" AND (title ILIKE $%d OR artist ILIKE $%d OR tags ILIKE $%d OR ccli_number ILIKE $%d)", argPos, argPos, argPos, argPos)
		args = append(args, "%"+params.Query+"%")
		argPos++
	}

	if params.Key != "" {
		whereClause += fmt.Sprintf(" AND default_key = $%d", argPos)
		args = append(args, params.Key)
		argPos++
	}

	if params.Tag != "" {
		whereClause += fmt.Sprintf(" AND tags ILIKE $%d", argPos)
		args = append(args, "%"+params.Tag+"%")
		argPos++
	}

	if params.HasLyrics == "yes" {
		whereClause += " AND lyrics IS NOT NULL AND lyrics != ''"
	} else if params.HasLyrics == "no" {
		whereClause += " AND (lyrics IS NULL OR lyrics = '')"
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM songs " + whereClause
	err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count songs: %w", err)
	}

	// Determine sort order
	orderBy := "ORDER BY title ASC"
	switch params.Sort {
	case "title_desc":
		orderBy = "ORDER BY title DESC"
	case "last_used":
		orderBy = "ORDER BY last_used DESC NULLS LAST"
	case "times_used":
		orderBy = "ORDER BY times_used DESC"
	case "recently_added":
		orderBy = "ORDER BY created_at DESC"
	case "artist":
		orderBy = "ORDER BY artist ASC, title ASC"
	}

	// Get songs
	sqlQuery := fmt.Sprintf(`
		SELECT id, tenant_id, title, COALESCE(artist, ''), COALESCE(default_key, ''), COALESCE(tempo, 0), COALESCE(ccli_number, ''), COALESCE(lyrics, ''), COALESCE(notes, ''), COALESCE(tags, ''), last_used, times_used, created_at, updated_at
		FROM songs
		%s
		%s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argPos, argPos+1)

	args = append(args, params.Limit, offset)

	rows, err := s.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list songs: %w", err)
	}
	defer rows.Close()

	songs := []Song{}
	for rows.Next() {
		var song Song
		err := rows.Scan(&song.ID, &song.TenantID, &song.Title, &song.Artist, &song.DefaultKey,
			&song.Tempo, &song.CCLINumber, &song.Lyrics, &song.Notes, &song.Tags,
			&song.LastUsed, &song.TimesUsed, &song.CreatedAt, &song.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan song: %w", err)
		}
		songs = append(songs, song)
	}

	return songs, total, nil
}

func (s *Service) GetSongStats(ctx context.Context, tenantID string) (*SongStats, error) {
	stats := &SongStats{}

	// Total songs, with lyrics
	err := s.db.QueryRow(ctx, `
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE lyrics IS NOT NULL AND lyrics != '')
		FROM songs`).Scan(&stats.TotalSongs, &stats.WithLyrics)
	if err != nil {
		return nil, fmt.Errorf("failed to get song counts: %w", err)
	}

	// With attachments
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT song_id) FROM song_attachments`).Scan(&stats.WithAttachments)
	if err != nil {
		stats.WithAttachments = 0 // table might not exist
	}

	// Most used (top 5)
	rows, err := s.db.Query(ctx, `
		SELECT id, tenant_id, title, COALESCE(artist, ''), COALESCE(default_key, ''), COALESCE(tempo, 0), COALESCE(ccli_number, ''), '', '', COALESCE(tags, ''), last_used, times_used, created_at, updated_at
		FROM songs
		WHERE times_used > 0
		ORDER BY times_used DESC
		LIMIT 5`)
	if err != nil {
		return nil, fmt.Errorf("failed to get most used songs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var song Song
		err := rows.Scan(&song.ID, &song.TenantID, &song.Title, &song.Artist, &song.DefaultKey,
			&song.Tempo, &song.CCLINumber, &song.Lyrics, &song.Notes, &song.Tags,
			&song.LastUsed, &song.TimesUsed, &song.CreatedAt, &song.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan most used song: %w", err)
		}
		stats.MostUsed = append(stats.MostUsed, song)
	}

	// Recently added (last 5)
	rows2, err := s.db.Query(ctx, `
		SELECT id, tenant_id, title, COALESCE(artist, ''), COALESCE(default_key, ''), COALESCE(tempo, 0), COALESCE(ccli_number, ''), '', '', COALESCE(tags, ''), last_used, times_used, created_at, updated_at
		FROM songs
		ORDER BY created_at DESC
		LIMIT 5`)
	if err != nil {
		return nil, fmt.Errorf("failed to get recently added songs: %w", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var song Song
		err := rows2.Scan(&song.ID, &song.TenantID, &song.Title, &song.Artist, &song.DefaultKey,
			&song.Tempo, &song.CCLINumber, &song.Lyrics, &song.Notes, &song.Tags,
			&song.LastUsed, &song.TimesUsed, &song.CreatedAt, &song.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recently added song: %w", err)
		}
		stats.RecentlyAdded = append(stats.RecentlyAdded, song)
	}

	// All unique keys
	rows3, err := s.db.Query(ctx, `
		SELECT DISTINCT default_key FROM songs 
		WHERE default_key IS NOT NULL AND default_key != '' 
		ORDER BY default_key`)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var key string
			if err := rows3.Scan(&key); err == nil {
				stats.AllKeys = append(stats.AllKeys, key)
			}
		}
	}

	// All unique tags (split comma-separated)
	rows4, err := s.db.Query(ctx, `
		SELECT DISTINCT TRIM(unnest(string_to_array(tags, ','))) as tag 
		FROM songs 
		WHERE tags IS NOT NULL AND tags != ''
		ORDER BY tag`)
	if err == nil {
		defer rows4.Close()
		for rows4.Next() {
			var tag string
			if err := rows4.Scan(&tag); err == nil && tag != "" {
				stats.AllTags = append(stats.AllTags, tag)
			}
		}
	}

	return stats, nil
}

func (s *Service) GetSongByID(ctx context.Context, tenantID, songID string) (*Song, error) {
	var song Song
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, title, COALESCE(artist, ''), COALESCE(default_key, ''), COALESCE(tempo, 0), COALESCE(ccli_number, ''), COALESCE(lyrics, ''), COALESCE(notes, ''), COALESCE(tags, ''), last_used, times_used, created_at, updated_at
		FROM songs WHERE id = $1`, songID).Scan(
		&song.ID, &song.TenantID, &song.Title, &song.Artist, &song.DefaultKey,
		&song.Tempo, &song.CCLINumber, &song.Lyrics, &song.Notes, &song.Tags,
		&song.LastUsed, &song.TimesUsed, &song.CreatedAt, &song.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("song not found")
		}
		return nil, fmt.Errorf("failed to get song: %w", err)
	}

	return &song, nil
}

func (s *Service) CreateSong(ctx context.Context, tenantID string, song *Song) (*Song, error) {
	song.ID = uuid.New().String()
	song.TenantID = tenantID

	err := s.db.QueryRow(ctx, `
		INSERT INTO songs (id, tenant_id, title, artist, default_key, tempo, ccli_number, lyrics, notes, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at`,
		song.ID, song.TenantID, song.Title, song.Artist, song.DefaultKey, song.Tempo, song.CCLINumber, song.Lyrics, song.Notes, song.Tags,
	).Scan(&song.CreatedAt, &song.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create song: %w", err)
	}

	return song, nil
}

func (s *Service) UpdateSong(ctx context.Context, tenantID, songID string, song *Song) (*Song, error) {
	err := s.db.QueryRow(ctx, `
		UPDATE songs SET 
			title = $1, artist = $2, default_key = $3, tempo = $4, ccli_number = $5, lyrics = $6, notes = $7, tags = $8
		WHERE id = $9
		RETURNING created_at, updated_at`,
		song.Title, song.Artist, song.DefaultKey, song.Tempo, song.CCLINumber, song.Lyrics, song.Notes, song.Tags, songID,
	).Scan(&song.CreatedAt, &song.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("song not found")
		}
		return nil, fmt.Errorf("failed to update song: %w", err)
	}

	song.ID = songID
	song.TenantID = tenantID

	return song, nil
}

func (s *Service) DeleteSong(ctx context.Context, tenantID, songID string) error {
	result, err := s.db.Exec(ctx, "DELETE FROM songs WHERE id = $1", songID)
	if err != nil {
		return fmt.Errorf("failed to delete song: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("song not found")
	}

	return nil
}

func (s *Service) GetSongUsage(ctx context.Context, tenantID, songID string) ([]SongUsage, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT si.service_id, s.name, s.service_date, s.service_time, si.song_key, si.position
		FROM service_items si
		JOIN services s ON s.id = si.service_id
		WHERE si.song_id = $1
		ORDER BY s.service_date DESC, si.position`, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to get song usage: %w", err)
	}
	defer rows.Close()

	usage := []SongUsage{}
	for rows.Next() {
		var u SongUsage
		err := rows.Scan(&u.ServiceID, &u.ServiceName, &u.ServiceDate, &u.ServiceTime, &u.SongKey, &u.Position)
		if err != nil {
			return nil, fmt.Errorf("failed to scan song usage: %w", err)
		}
		usage = append(usage, u)
	}

	return usage, nil
}

// Song Attachment operations

func (s *Service) CreateSongAttachment(ctx context.Context, tenantID string, attachment *SongAttachment) (*SongAttachment, error) {
	attachment.ID = uuid.New().String()
	attachment.TenantID = tenantID

	err := s.db.QueryRow(ctx, `
		INSERT INTO song_attachments (id, tenant_id, song_id, filename, original_name, content_type, file_data, file_size, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at`,
		attachment.ID, attachment.TenantID, attachment.SongID, attachment.Filename, attachment.OriginalName,
		attachment.ContentType, attachment.FileData, attachment.FileSize, attachment.UploadedBy,
	).Scan(&attachment.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create song attachment: %w", err)
	}

	// Clear FileData from response (already stored)
	attachment.FileData = nil

	return attachment, nil
}

func (s *Service) ListSongAttachments(ctx context.Context, tenantID, songID string) ([]SongAttachment, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, tenant_id, song_id, filename, original_name, content_type, file_size, uploaded_by, created_at
		FROM song_attachments
		WHERE tenant_id = $1 AND song_id = $2
		ORDER BY created_at DESC`, tenantID, songID)
	if err != nil {
		return nil, fmt.Errorf("failed to list song attachments: %w", err)
	}
	defer rows.Close()

	attachments := []SongAttachment{}
	for rows.Next() {
		var a SongAttachment
		err := rows.Scan(&a.ID, &a.TenantID, &a.SongID, &a.Filename, &a.OriginalName, &a.ContentType, &a.FileSize, &a.UploadedBy, &a.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan song attachment: %w", err)
		}
		attachments = append(attachments, a)
	}

	return attachments, nil
}

func (s *Service) GetSongAttachment(ctx context.Context, tenantID, attachmentID string) (*SongAttachment, error) {
	var a SongAttachment
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, song_id, filename, original_name, content_type, file_data, file_size, uploaded_by, created_at
		FROM song_attachments
		WHERE tenant_id = $1 AND id = $2`, tenantID, attachmentID).Scan(
		&a.ID, &a.TenantID, &a.SongID, &a.Filename, &a.OriginalName, &a.ContentType, &a.FileData, &a.FileSize, &a.UploadedBy, &a.CreatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("attachment not found")
		}
		return nil, fmt.Errorf("failed to get song attachment: %w", err)
	}

	return &a, nil
}

func (s *Service) DeleteSongAttachment(ctx context.Context, tenantID, attachmentID string) error {
	result, err := s.db.Exec(ctx, "DELETE FROM song_attachments WHERE tenant_id = $1 AND id = $2", tenantID, attachmentID)
	if err != nil {
		return fmt.Errorf("failed to delete song attachment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("attachment not found")
	}

	return nil
}

// Service Template operations

func (s *Service) ListTemplates(ctx context.Context, tenantID string) ([]ServiceTemplate, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, tenant_id, name, COALESCE(description, ''), template_data, created_at, updated_at
		FROM service_templates
		WHERE tenant_id = $1
		ORDER BY name`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	defer rows.Close()

	templates := []ServiceTemplate{}
	for rows.Next() {
		var t ServiceTemplate
		err := rows.Scan(&t.ID, &t.TenantID, &t.Name, &t.Description, &t.TemplateData, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (s *Service) GetTemplate(ctx context.Context, tenantID, templateID string) (*ServiceTemplate, error) {
	var t ServiceTemplate
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, name, COALESCE(description, ''), template_data, created_at, updated_at
		FROM service_templates
		WHERE id = $1 AND tenant_id = $2`, templateID, tenantID).Scan(
		&t.ID, &t.TenantID, &t.Name, &t.Description, &t.TemplateData, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return &t, nil
}

func (s *Service) CreateTemplate(ctx context.Context, tenantID string, t *ServiceTemplate) (*ServiceTemplate, error) {
	t.ID = uuid.New().String()
	t.TenantID = tenantID

	if t.TemplateData == nil {
		t.TemplateData = json.RawMessage(`{}`)
	}

	err := s.db.QueryRow(ctx, `
		INSERT INTO service_templates (id, tenant_id, name, description, template_data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`,
		t.ID, t.TenantID, t.Name, t.Description, t.TemplateData,
	).Scan(&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}
	return t, nil
}

func (s *Service) UpdateTemplate(ctx context.Context, tenantID, templateID string, t *ServiceTemplate) (*ServiceTemplate, error) {
	err := s.db.QueryRow(ctx, `
		UPDATE service_templates SET name = $1, description = $2, template_data = $3, updated_at = NOW()
		WHERE id = $4 AND tenant_id = $5
		RETURNING created_at, updated_at`,
		t.Name, t.Description, t.TemplateData, templateID, tenantID,
	).Scan(&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("template not found")
		}
		return nil, fmt.Errorf("failed to update template: %w", err)
	}
	t.ID = templateID
	t.TenantID = tenantID
	return t, nil
}

func (s *Service) DeleteTemplate(ctx context.Context, tenantID, templateID string) error {
	result, err := s.db.Exec(ctx, "DELETE FROM service_templates WHERE id = $1 AND tenant_id = $2", templateID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("template not found")
	}
	return nil
}

// SaveAsTemplate creates a template from an existing service
func (s *Service) SaveAsTemplate(ctx context.Context, tenantID, serviceID, name, description string) (*ServiceTemplate, error) {
	svc, err := s.GetServiceByID(ctx, tenantID, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	type TemplateItem struct {
		ItemType        string  `json:"item_type"`
		Title           string  `json:"title"`
		SongID          *string `json:"song_id,omitempty"`
		SongKey         string  `json:"song_key,omitempty"`
		Position        int     `json:"position"`
		DurationMinutes *int    `json:"duration_minutes,omitempty"`
		Notes           string  `json:"notes,omitempty"`
		AssignedTo      string  `json:"assigned_to,omitempty"`
	}

	type TemplateTeamRole struct {
		Role string `json:"role"`
	}

	type TemplateData struct {
		Items []TemplateItem     `json:"items"`
		Roles []TemplateTeamRole `json:"roles"`
	}

	td := TemplateData{}
	for _, item := range svc.Items {
		td.Items = append(td.Items, TemplateItem{
			ItemType:        item.ItemType,
			Title:           item.Title,
			SongID:          item.SongID,
			SongKey:         item.SongKey,
			Position:        item.Position,
			DurationMinutes: item.DurationMinutes,
			Notes:           item.Notes,
			AssignedTo:      item.AssignedTo,
		})
	}

	// Extract unique roles from team
	rolesSeen := map[string]bool{}
	for _, member := range svc.Team {
		if !rolesSeen[member.Role] {
			td.Roles = append(td.Roles, TemplateTeamRole{Role: member.Role})
			rolesSeen[member.Role] = true
		}
	}

	templateDataJSON, err := json.Marshal(td)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal template data: %w", err)
	}

	t := &ServiceTemplate{
		Name:         name,
		Description:  description,
		TemplateData: templateDataJSON,
	}

	return s.CreateTemplate(ctx, tenantID, t)
}

// CopyService duplicates a service with its items (no team)
func (s *Service) CopyService(ctx context.Context, tenantID, serviceID string, newDate time.Time) (*ChurchService, error) {
	original, err := s.GetServiceByID(ctx, tenantID, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original service: %w", err)
	}

	newSvc := &ChurchService{
		ServiceTypeID: original.ServiceTypeID,
		Name:          original.Name,
		ServiceDate:   newDate,
		ServiceTime:   original.ServiceTime,
		Notes:         original.Notes,
		Status:        "draft",
	}

	created, err := s.CreateService(ctx, tenantID, newSvc)
	if err != nil {
		return nil, fmt.Errorf("failed to create copy: %w", err)
	}

	// Copy items
	for _, item := range original.Items {
		newItem := &ServiceItem{
			ServiceID:       created.ID,
			ItemType:        item.ItemType,
			Title:           item.Title,
			SongID:          item.SongID,
			SongKey:         item.SongKey,
			Position:        item.Position,
			DurationMinutes: item.DurationMinutes,
			Notes:           item.Notes,
			AssignedTo:      item.AssignedTo,
		}
		if _, err := s.AddServiceItem(ctx, tenantID, newItem); err != nil {
			return nil, fmt.Errorf("failed to copy item: %w", err)
		}
	}

	return s.GetServiceByID(ctx, tenantID, created.ID)
}

// ReorderItems batch-updates item positions
func (s *Service) ReorderItems(ctx context.Context, tenantID, serviceID string, itemIDs []string) error {
	for i, id := range itemIDs {
		_, err := s.db.Exec(ctx, `UPDATE service_items SET position = $1 WHERE id = $2 AND service_id = $3`, i+1, id, serviceID)
		if err != nil {
			return fmt.Errorf("failed to reorder item %s: %w", id, err)
		}
	}
	return nil
}

// DeleteServiceType soft-deletes a service type
func (s *Service) DeleteServiceType(ctx context.Context, tenantID, typeID string) error {
	result, err := s.db.Exec(ctx, "UPDATE service_types SET is_active = false WHERE id = $1 AND tenant_id = $2", typeID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete service type: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("service type not found")
	}
	return nil
}
