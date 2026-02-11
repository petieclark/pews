package services

import (
	"context"
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, tenant_id, name, default_time, default_day, color, is_active, created_at, updated_at
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	st.ID = uuid.New().String()
	st.TenantID = tenantID

	err = s.db.QueryRow(ctx, `
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	err = s.db.QueryRow(ctx, `
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

	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to set tenant context: %w", err)
	}

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
	err = s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count services: %w", err)
	}

	// Get services
	query := fmt.Sprintf(`
		SELECT s.id, s.tenant_id, s.service_type_id, s.name, s.service_date, 
		       s.service_time, s.notes, s.status, s.created_at, s.updated_at,
		       st.name, st.color
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

	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	today := time.Now().Format("2006-01-02")

	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.tenant_id, s.service_type_id, s.name, s.service_date, 
		       s.service_time, s.notes, s.status, s.created_at, s.updated_at,
		       st.name, st.color
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	var svc ChurchService
	err = s.db.QueryRow(ctx, `
		SELECT id, tenant_id, service_type_id, name, service_date, service_time, notes, status, created_at, updated_at
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
	err = s.db.QueryRow(ctx, `
		SELECT id, tenant_id, name, default_time, default_day, color, is_active, created_at, updated_at
		FROM service_types WHERE id = $1`, svc.ServiceTypeID).Scan(
		&st.ID, &st.TenantID, &st.Name, &st.DefaultTime, &st.DefaultDay,
		&st.Color, &st.IsActive, &st.CreatedAt, &st.UpdatedAt,
	)
	if err == nil {
		svc.ServiceType = &st
	}

	// Load items
	items, err := s.GetServiceItems(ctx, tenantID, serviceID)
	if err == nil {
		svc.Items = items
	}

	// Load team
	team, err := s.GetServiceTeam(ctx, tenantID, serviceID)
	if err == nil {
		svc.Team = team
	}

	return &svc, nil
}

func (s *Service) CreateService(ctx context.Context, tenantID string, svc *ChurchService) (*ChurchService, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	svc.ID = uuid.New().String()
	svc.TenantID = tenantID

	err = s.db.QueryRow(ctx, `
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	err = s.db.QueryRow(ctx, `
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, service_id, item_type, title, song_id, song_key, position, duration_minutes, notes, assigned_to
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	item.ID = uuid.New().String()

	_, err = s.db.Exec(ctx, `
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT st.id, st.service_id, st.person_id, st.role, st.status, st.notes,
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	member.ID = uuid.New().String()

	_, err = s.db.Exec(ctx, `
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

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
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to set tenant context: %w", err)
	}

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	if query != "" {
		whereClause += fmt.Sprintf(" AND (title ILIKE $%d OR artist ILIKE $%d OR tags ILIKE $%d)", argPos, argPos, argPos)
		args = append(args, "%"+query+"%")
		argPos++
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM songs " + whereClause
	err = s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count songs: %w", err)
	}

	// Get songs
	sqlQuery := fmt.Sprintf(`
		SELECT id, tenant_id, title, artist, default_key, tempo, ccli_number, lyrics, notes, tags, last_used, times_used, created_at, updated_at
		FROM songs
		%s
		ORDER BY title
		LIMIT $%d OFFSET $%d`, whereClause, argPos, argPos+1)

	args = append(args, limit, offset)

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

func (s *Service) GetSongByID(ctx context.Context, tenantID, songID string) (*Song, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	var song Song
	err = s.db.QueryRow(ctx, `
		SELECT id, tenant_id, title, artist, default_key, tempo, ccli_number, lyrics, notes, tags, last_used, times_used, created_at, updated_at
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	song.ID = uuid.New().String()
	song.TenantID = tenantID

	err = s.db.QueryRow(ctx, `
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	err = s.db.QueryRow(ctx, `
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
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

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
