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
		SELECT st.id, st.service_id, st.person_id, st.team_id, st.role, st.status, st.notes,
		       st.notified_at, st.responded_at, st.notification_sent,
		       p.first_name, p.last_name, p.email,
		       vt.name
		FROM service_teams st
		JOIN people p ON p.id = st.person_id
		LEFT JOIN volunteer_teams vt ON vt.id = st.team_id
		WHERE st.service_id = $1
		ORDER BY vt.name, st.role`, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get service team: %w", err)
	}
	defer rows.Close()

	team := []ServiceTeam{}
	for rows.Next() {
		var member ServiceTeam
		err := rows.Scan(&member.ID, &member.ServiceID, &member.PersonID, &member.TeamID, &member.Role, 
			&member.Status, &member.Notes, &member.NotifiedAt, &member.RespondedAt, &member.NotificationSent,
			&member.PersonFirstName, &member.PersonLastName, &member.PersonEmail, &member.TeamName)
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

// Volunteer Teams operations

func (s *Service) ListVolunteerTeams(ctx context.Context, tenantID string) ([]VolunteerTeam, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT vt.id, vt.tenant_id, vt.name, vt.description, vt.color, vt.is_active, 
		       vt.created_at, vt.updated_at,
		       COUNT(tm.id) as member_count
		FROM volunteer_teams vt
		LEFT JOIN team_members tm ON tm.team_id = vt.id AND tm.is_active = TRUE
		GROUP BY vt.id
		ORDER BY vt.name`)
	if err != nil {
		return nil, fmt.Errorf("failed to list volunteer teams: %w", err)
	}
	defer rows.Close()

	teams := []VolunteerTeam{}
	for rows.Next() {
		var team VolunteerTeam
		err := rows.Scan(&team.ID, &team.TenantID, &team.Name, &team.Description, &team.Color,
			&team.IsActive, &team.CreatedAt, &team.UpdatedAt, &team.MemberCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan volunteer team: %w", err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (s *Service) GetVolunteerTeamByID(ctx context.Context, tenantID, teamID string) (*VolunteerTeam, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	var team VolunteerTeam
	err = s.db.QueryRow(ctx, `
		SELECT id, tenant_id, name, description, color, is_active, created_at, updated_at
		FROM volunteer_teams WHERE id = $1`, teamID).Scan(
		&team.ID, &team.TenantID, &team.Name, &team.Description, &team.Color,
		&team.IsActive, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("volunteer team not found")
		}
		return nil, fmt.Errorf("failed to get volunteer team: %w", err)
	}

	return &team, nil
}

func (s *Service) CreateVolunteerTeam(ctx context.Context, tenantID string, team *VolunteerTeam) (*VolunteerTeam, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	team.ID = uuid.New().String()
	team.TenantID = tenantID

	err = s.db.QueryRow(ctx, `
		INSERT INTO volunteer_teams (id, tenant_id, name, description, color, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`,
		team.ID, team.TenantID, team.Name, team.Description, team.Color, team.IsActive,
	).Scan(&team.CreatedAt, &team.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create volunteer team: %w", err)
	}

	return team, nil
}

func (s *Service) UpdateVolunteerTeam(ctx context.Context, tenantID, teamID string, team *VolunteerTeam) (*VolunteerTeam, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	err = s.db.QueryRow(ctx, `
		UPDATE volunteer_teams SET 
			name = $1, description = $2, color = $3, is_active = $4
		WHERE id = $5
		RETURNING created_at, updated_at`,
		team.Name, team.Description, team.Color, team.IsActive, teamID,
	).Scan(&team.CreatedAt, &team.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("volunteer team not found")
		}
		return nil, fmt.Errorf("failed to update volunteer team: %w", err)
	}

	team.ID = teamID
	team.TenantID = tenantID

	return team, nil
}

func (s *Service) DeleteVolunteerTeam(ctx context.Context, tenantID, teamID string) error {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	result, err := s.db.Exec(ctx, "DELETE FROM volunteer_teams WHERE id = $1", teamID)
	if err != nil {
		return fmt.Errorf("failed to delete volunteer team: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("volunteer team not found")
	}

	return nil
}

// Team Members operations

func (s *Service) GetTeamMembers(ctx context.Context, tenantID, teamID string) ([]TeamMember, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT tm.id, tm.team_id, tm.person_id, tm.role, tm.is_active, tm.added_at,
		       p.first_name, p.last_name, p.email,
		       vt.name, vt.color
		FROM team_members tm
		JOIN people p ON p.id = tm.person_id
		JOIN volunteer_teams vt ON vt.id = tm.team_id
		WHERE tm.team_id = $1
		ORDER BY tm.is_active DESC, p.first_name, p.last_name`, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	defer rows.Close()

	members := []TeamMember{}
	for rows.Next() {
		var member TeamMember
		err := rows.Scan(&member.ID, &member.TeamID, &member.PersonID, &member.Role,
			&member.IsActive, &member.AddedAt, &member.PersonFirstName, &member.PersonLastName,
			&member.PersonEmail, &member.TeamName, &member.TeamColor)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

func (s *Service) GetPersonTeams(ctx context.Context, tenantID, personID string) ([]TeamMember, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT tm.id, tm.team_id, tm.person_id, tm.role, tm.is_active, tm.added_at,
		       p.first_name, p.last_name, p.email,
		       vt.name, vt.color
		FROM team_members tm
		JOIN people p ON p.id = tm.person_id
		JOIN volunteer_teams vt ON vt.id = tm.team_id
		WHERE tm.person_id = $1
		ORDER BY tm.is_active DESC, vt.name`, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get person teams: %w", err)
	}
	defer rows.Close()

	members := []TeamMember{}
	for rows.Next() {
		var member TeamMember
		err := rows.Scan(&member.ID, &member.TeamID, &member.PersonID, &member.Role,
			&member.IsActive, &member.AddedAt, &member.PersonFirstName, &member.PersonLastName,
			&member.PersonEmail, &member.TeamName, &member.TeamColor)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

func (s *Service) AddTeamMember(ctx context.Context, tenantID string, member *TeamMember) (*TeamMember, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	member.ID = uuid.New().String()

	err = s.db.QueryRow(ctx, `
		INSERT INTO team_members (id, team_id, person_id, role, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING added_at`,
		member.ID, member.TeamID, member.PersonID, member.Role, member.IsActive,
	).Scan(&member.AddedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to add team member: %w", err)
	}

	return member, nil
}

func (s *Service) UpdateTeamMember(ctx context.Context, tenantID, memberID string, member *TeamMember) (*TeamMember, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	err = s.db.QueryRow(ctx, `
		UPDATE team_members SET role = $1, is_active = $2
		WHERE id = $3
		RETURNING added_at`,
		member.Role, member.IsActive, memberID,
	).Scan(&member.AddedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("team member not found")
		}
		return nil, fmt.Errorf("failed to update team member: %w", err)
	}

	member.ID = memberID

	return member, nil
}

func (s *Service) RemoveTeamMember(ctx context.Context, tenantID, memberID string) error {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	result, err := s.db.Exec(ctx, "DELETE FROM team_members WHERE id = $1", memberID)
	if err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("team member not found")
	}

	return nil
}

// Volunteer Availability operations

func (s *Service) GetPersonAvailability(ctx context.Context, tenantID, personID string) ([]VolunteerAvailability, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT va.id, va.person_id, va.team_id, va.start_date, va.end_date, va.reason,
		       va.created_at, va.updated_at,
		       p.first_name, p.last_name
		FROM volunteer_availability va
		JOIN people p ON p.id = va.person_id
		WHERE va.person_id = $1
		ORDER BY va.start_date DESC`, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get person availability: %w", err)
	}
	defer rows.Close()

	avail := []VolunteerAvailability{}
	for rows.Next() {
		var a VolunteerAvailability
		err := rows.Scan(&a.ID, &a.PersonID, &a.TeamID, &a.StartDate, &a.EndDate, &a.Reason,
			&a.CreatedAt, &a.UpdatedAt, &a.PersonFirstName, &a.PersonLastName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan availability: %w", err)
		}
		avail = append(avail, a)
	}

	return avail, nil
}

func (s *Service) AddAvailability(ctx context.Context, tenantID string, avail *VolunteerAvailability) (*VolunteerAvailability, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	avail.ID = uuid.New().String()

	err = s.db.QueryRow(ctx, `
		INSERT INTO volunteer_availability (id, person_id, team_id, start_date, end_date, reason)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`,
		avail.ID, avail.PersonID, avail.TeamID, avail.StartDate, avail.EndDate, avail.Reason,
	).Scan(&avail.CreatedAt, &avail.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to add availability: %w", err)
	}

	return avail, nil
}

func (s *Service) UpdateAvailability(ctx context.Context, tenantID, availID string, avail *VolunteerAvailability) (*VolunteerAvailability, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	err = s.db.QueryRow(ctx, `
		UPDATE volunteer_availability SET 
			start_date = $1, end_date = $2, reason = $3, team_id = $4
		WHERE id = $5
		RETURNING created_at, updated_at`,
		avail.StartDate, avail.EndDate, avail.Reason, avail.TeamID, availID,
	).Scan(&avail.CreatedAt, &avail.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("availability not found")
		}
		return nil, fmt.Errorf("failed to update availability: %w", err)
	}

	avail.ID = availID

	return avail, nil
}

func (s *Service) DeleteAvailability(ctx context.Context, tenantID, availID string) error {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	result, err := s.db.Exec(ctx, "DELETE FROM volunteer_availability WHERE id = $1", availID)
	if err != nil {
		return fmt.Errorf("failed to delete availability: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("availability not found")
	}

	return nil
}

func (s *Service) IsPersonAvailable(ctx context.Context, tenantID, personID string, date time.Time) (bool, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return false, fmt.Errorf("failed to set tenant context: %w", err)
	}

	var available bool
	err = s.db.QueryRow(ctx, "SELECT is_person_available($1, $2)", personID, date).Scan(&available)
	if err != nil {
		return false, fmt.Errorf("failed to check availability: %w", err)
	}

	return available, nil
}

// Scheduling helpers

func (s *Service) GetSchedulingConflicts(ctx context.Context, tenantID string, serviceDate time.Time) ([]SchedulingConflict, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT person_id, first_name, last_name, service_count
		FROM get_scheduling_conflicts($1, $2)`, serviceDate, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduling conflicts: %w", err)
	}
	defer rows.Close()

	conflicts := []SchedulingConflict{}
	for rows.Next() {
		var c SchedulingConflict
		err := rows.Scan(&c.PersonID, &c.FirstName, &c.LastName, &c.ServiceCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conflict: %w", err)
		}
		conflicts = append(conflicts, c)
	}

	return conflicts, nil
}

func (s *Service) GetAvailableVolunteers(ctx context.Context, tenantID, teamID string, date time.Time) ([]TeamMember, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT tm.id, tm.team_id, tm.person_id, tm.role, tm.is_active, tm.added_at,
		       p.first_name, p.last_name, p.email,
		       vt.name, vt.color
		FROM team_members tm
		JOIN people p ON p.id = tm.person_id
		JOIN volunteer_teams vt ON vt.id = tm.team_id
		WHERE tm.team_id = $1 
		AND tm.is_active = TRUE
		AND is_person_available(tm.person_id, $2)
		AND NOT EXISTS (
			SELECT 1 FROM service_teams st
			JOIN services s ON s.id = st.service_id
			WHERE st.person_id = tm.person_id
			AND s.service_date = $2
		)
		ORDER BY p.first_name, p.last_name`, teamID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get available volunteers: %w", err)
	}
	defer rows.Close()

	members := []TeamMember{}
	for rows.Next() {
		var member TeamMember
		err := rows.Scan(&member.ID, &member.TeamID, &member.PersonID, &member.Role,
			&member.IsActive, &member.AddedAt, &member.PersonFirstName, &member.PersonLastName,
			&member.PersonEmail, &member.TeamName, &member.TeamColor)
		if err != nil {
			return nil, fmt.Errorf("failed to scan available volunteer: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

func (s *Service) UpdateServiceTeamStatus(ctx context.Context, tenantID, teamMemberID, status string) error {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	result, err := s.db.Exec(ctx, `
		UPDATE service_teams SET status = $1, responded_at = NOW()
		WHERE id = $2`,
		status, teamMemberID,
	)

	if err != nil {
		return fmt.Errorf("failed to update service team status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("service team member not found")
	}

	return nil
}
