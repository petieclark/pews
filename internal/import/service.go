package importpkg

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// ImportPeople bulk imports people
func (s *Service) ImportPeople(ctx context.Context, tenantID string, people []PersonImport, dryRun bool) (*ImportResult, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	result := &ImportResult{
		Errors: []string{},
	}

	for i, person := range people {
		// Skip if missing required fields
		if person.FirstName == "" || person.LastName == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: missing required fields (first_name, last_name)", i+1))
			continue
		}

		// Check for duplicate by email
		if person.Email != "" {
			var exists bool
			err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM people WHERE email = $1)", person.Email).Scan(&exists)
			if err == nil && exists {
				result.Skipped++
				continue
			}
		}

		if dryRun {
			result.Created++
			continue
		}

		// Insert person
		id := uuid.New().String()
		membershipStatus := person.MembershipStatus
		if membershipStatus == "" {
			membershipStatus = "active"
		}

		_, err := s.db.Exec(ctx, `
			INSERT INTO people (
				id, tenant_id, first_name, last_name, email, phone,
				address_line1, address_line2, city, state, zip,
				gender, membership_status, photo_url, notes, custom_fields,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, NOW(), NOW()
			)`,
			id, tenantID, person.FirstName, person.LastName, person.Email, person.Phone,
			person.AddressLine1, person.AddressLine2, person.City, person.State, person.Zip,
			person.Gender, membershipStatus, person.PhotoURL, person.Notes, person.CustomFields,
		)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+1, err))
			continue
		}

		result.Created++
	}

	return result, nil
}

// ImportGroups bulk imports groups
func (s *Service) ImportGroups(ctx context.Context, tenantID string, groups []GroupImport, dryRun bool) (*ImportResult, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	result := &ImportResult{
		Errors: []string{},
	}

	for i, group := range groups {
		// Skip if missing required fields
		if group.Name == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: missing required field (name)", i+1))
			continue
		}

		// Check for duplicate by name
		var exists bool
		err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM groups WHERE name = $1)", group.Name).Scan(&exists)
		if err == nil && exists {
			result.Skipped++
			continue
		}

		if dryRun {
			result.Created++
			continue
		}

		// Insert group
		groupID := uuid.New().String()
		groupType := group.Type
		if groupType == "" {
			groupType = "small_group"
		}

		_, err = s.db.Exec(ctx, `
			INSERT INTO groups (
				id, tenant_id, name, description, group_type,
				meeting_day, meeting_time, meeting_location,
				is_public, max_members, is_active,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, TRUE, NOW(), NOW()
			)`,
			groupID, tenantID, group.Name, group.Description, groupType,
			group.MeetingDay, group.MeetingTime, group.MeetingLocation,
			group.IsPublic, group.MaxMembers,
		)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+1, err))
			continue
		}

		// Add members if provided
		if len(group.Members) > 0 {
			for _, email := range group.Members {
				// Find person by email
				var personID string
				err := s.db.QueryRow(ctx, "SELECT id FROM people WHERE email = $1", email).Scan(&personID)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Row %d: member email %s not found", i+1, email))
					continue
				}

				// Add to group
				memberID := uuid.New().String()
				_, err = s.db.Exec(ctx, `
					INSERT INTO group_members (id, group_id, person_id, role, joined_at)
					VALUES ($1, $2, $3, 'member', NOW())
				`, memberID, groupID, personID)

				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Row %d: failed to add member %s: %v", i+1, email, err))
				}
			}
		}

		result.Created++
	}

	return result, nil
}

// ImportSongs bulk imports songs
func (s *Service) ImportSongs(ctx context.Context, tenantID string, songs []SongImport, dryRun bool) (*ImportResult, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	result := &ImportResult{
		Errors: []string{},
	}

	for i, song := range songs {
		// Skip if missing required fields
		if song.Title == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: missing required field (title)", i+1))
			continue
		}

		// Check for duplicate by title and artist
		var exists bool
		err := s.db.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM songs WHERE title = $1 AND artist = $2)",
			song.Title, song.Artist).Scan(&exists)
		if err == nil && exists {
			result.Skipped++
			continue
		}

		if dryRun {
			result.Created++
			continue
		}

		// Insert song
		id := uuid.New().String()
		_, err = s.db.Exec(ctx, `
			INSERT INTO songs (
				id, tenant_id, title, artist, default_key, tempo,
				ccli_number, lyrics, notes, tags, times_used,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 0, NOW(), NOW()
			)`,
			id, tenantID, song.Title, song.Artist, song.Key, song.Tempo,
			song.CCLINumber, song.Lyrics, song.Notes, song.Tags,
		)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+1, err))
			continue
		}

		result.Created++
	}

	return result, nil
}

// ImportGiving bulk imports donations
func (s *Service) ImportGiving(ctx context.Context, tenantID string, donations []DonationImport, dryRun bool) (*ImportResult, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	result := &ImportResult{
		Errors: []string{},
	}

	for i, donation := range donations {
		// Skip if missing required fields
		if donation.DonorEmail == "" || donation.FundName == "" || donation.AmountCents <= 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: missing required fields (donor_email, fund_name, amount_cents)", i+1))
			continue
		}

		// Find donor by email
		var personID string
		err := s.db.QueryRow(ctx, "SELECT id FROM people WHERE email = $1", donation.DonorEmail).Scan(&personID)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: donor email %s not found", i+1, donation.DonorEmail))
			continue
		}

		// Find or create fund by name
		var fundID string
		err = s.db.QueryRow(ctx, "SELECT id FROM funds WHERE name = $1", donation.FundName).Scan(&fundID)
		if err != nil {
			// Fund doesn't exist, create it
			fundID = uuid.New().String()
			_, err = s.db.Exec(ctx, `
				INSERT INTO funds (id, tenant_id, name, description, is_default, is_active, created_at, updated_at)
				VALUES ($1, $2, $3, '', FALSE, TRUE, NOW(), NOW())
			`, fundID, tenantID, donation.FundName)
			
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Row %d: failed to create fund: %v", i+1, err))
				continue
			}
		}

		if dryRun {
			result.Created++
			continue
		}

		// Parse donated_at timestamp
		var donatedAt time.Time
		if donation.DonatedAt != "" {
			donatedAt, err = time.Parse(time.RFC3339, donation.DonatedAt)
			if err != nil {
				donatedAt = time.Now()
			}
		} else {
			donatedAt = time.Now()
		}

		// Insert donation
		id := uuid.New().String()
		currency := donation.Currency
		if currency == "" {
			currency = "USD"
		}

		_, err = s.db.Exec(ctx, `
			INSERT INTO donations (
				id, tenant_id, person_id, fund_id, amount_cents, currency,
				payment_method, status, is_recurring, memo, donated_at,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, 'completed', FALSE, $8, $9, NOW(), NOW()
			)`,
			id, tenantID, personID, fundID, donation.AmountCents, currency,
			donation.PaymentMethod, donation.Memo, donatedAt,
		)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+1, err))
			continue
		}

		result.Created++
	}

	return result, nil
}

// ImportPCOPeople bulk imports people from PCO with support for custom fields and updates
func (s *Service) ImportPCOPeople(ctx context.Context, tenantID string, people []PersonImport, updateMode string) (*PCOImportResult, error) {
	result := &PCOImportResult{
		Imported: 0,
		Updated:  0,
		Skipped:  0,
		Errors:   []string{},
	}

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set tenant context INSIDE the transaction
	_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	for i, person := range people {
		// Skip if missing required fields
		if person.FirstName == "" || person.LastName == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: missing required fields (first_name, last_name)", i+1))
			continue
		}

		savepointName := fmt.Sprintf("sp_%d", i)
		_, _ = tx.Exec(ctx, "SAVEPOINT "+savepointName)

		// Check for duplicate by email
		var existingID string
		if person.Email != "" {
			err := tx.QueryRow(ctx, "SELECT id FROM people WHERE email = $1", person.Email).Scan(&existingID)
			if err == nil {
				// Person exists
				if updateMode == "skip" {
					result.Skipped++
					_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT "+savepointName)
					continue
				} else if updateMode == "update" {
					_, err := tx.Exec(ctx, `
						UPDATE people SET
							first_name = COALESCE(NULLIF($1, ''), first_name),
							last_name = COALESCE(NULLIF($2, ''), last_name),
							phone = COALESCE(NULLIF($3, ''), phone),
							address_line1 = COALESCE(NULLIF($4, ''), address_line1),
							address_line2 = COALESCE(NULLIF($5, ''), address_line2),
							city = COALESCE(NULLIF($6, ''), city),
							state = COALESCE(NULLIF($7, ''), state),
							zip = COALESCE(NULLIF($8, ''), zip),
							birthdate = CASE WHEN $9 != '' THEN $9::date ELSE birthdate END,
							gender = COALESCE(NULLIF($10, ''), gender),
							membership_status = COALESCE(NULLIF($11, ''), membership_status),
							photo_url = COALESCE(NULLIF($12, ''), photo_url),
							notes = COALESCE(NULLIF($13, ''), notes),
							custom_fields = CASE
								WHEN $14::jsonb IS NOT NULL
								THEN COALESCE(custom_fields, '{}'::jsonb) || $14::jsonb
								ELSE custom_fields
							END,
							updated_at = NOW()
						WHERE id = $15
					`,
						person.FirstName, person.LastName, person.Phone,
						person.AddressLine1, person.AddressLine2, person.City, person.State, person.Zip,
						person.Birthdate, person.Gender, person.MembershipStatus, person.PhotoURL, person.Notes,
						person.CustomFields, existingID,
					)

					if err != nil {
						result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+1, err))
						_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+savepointName)
						continue
					}

					result.Updated++
					_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT "+savepointName)
					continue
				}
			} else {
				// No match — rollback savepoint from failed SELECT
				_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+savepointName)
				_, _ = tx.Exec(ctx, "SAVEPOINT "+savepointName)
			}
		}

		// Insert new person
		id := uuid.New().String()
		membershipStatus := person.MembershipStatus
		if membershipStatus == "" {
			membershipStatus = "active"
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO people (
				id, tenant_id, first_name, last_name, email, phone,
				address_line1, address_line2, city, state, zip,
				birthdate, gender, membership_status, photo_url, notes, custom_fields,
				in_directory, profile_completed,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
				NULLIF($12, '')::date, $13, $14, $15, $16, $17,
				FALSE, FALSE,
				NOW(), NOW()
			)`,
			id, tenantID, person.FirstName, person.LastName, person.Email, person.Phone,
			person.AddressLine1, person.AddressLine2, person.City, person.State, person.Zip,
			person.Birthdate, person.Gender, membershipStatus, person.PhotoURL, person.Notes, person.CustomFields,
		)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+1, err))
			_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+savepointName)
			continue
		}

		result.Imported++
		_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT "+savepointName)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction (%d imported, %d errors): %w", result.Imported, len(result.Errors), err)
	}

	// Log import to history
	s.logImport(ctx, tenantID, "pco_people", result.Imported, result.Updated, result.Skipped, len(result.Errors))

	return result, nil
}

// ImportPCOSongs bulk imports songs from PCO with support for updates
func (s *Service) ImportPCOSongs(ctx context.Context, tenantID string, songs []SongImport, updateMode string) (*PCOImportResult, error) {
	result := &PCOImportResult{
		Imported: 0,
		Updated:  0,
		Skipped:  0,
		Errors:   []string{},
	}

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set tenant context INSIDE the transaction
	_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	for i, song := range songs {
		// Skip if missing required fields
		if song.Title == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: missing required field (title)", i+1))
			continue
		}

		// Use savepoint so a failed row doesn't poison the entire transaction
		savepointName := fmt.Sprintf("sp_%d", i)
		_, _ = tx.Exec(ctx, "SAVEPOINT "+savepointName)

		// Check for duplicate by title+artist or CCLI number
		var existingID string
		var err error

		if song.CCLINumber != "" {
			err = tx.QueryRow(ctx, "SELECT id FROM songs WHERE ccli_number = $1 AND ccli_number != ''", song.CCLINumber).Scan(&existingID)
		} else if song.Artist != "" {
			err = tx.QueryRow(ctx, "SELECT id FROM songs WHERE title = $1 AND artist = $2", song.Title, song.Artist).Scan(&existingID)
		} else {
			err = tx.QueryRow(ctx, "SELECT id FROM songs WHERE title = $1", song.Title).Scan(&existingID)
		}

		if err == nil {
			// Song exists
			if updateMode == "skip" {
				result.Skipped++
				_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT "+savepointName)
				continue
			} else if updateMode == "update" {
				_, err := tx.Exec(ctx, `
					UPDATE songs SET
						artist = COALESCE(NULLIF($1, ''), artist),
						default_key = COALESCE(NULLIF($2, ''), default_key),
						tempo = CASE WHEN $3 > 0 THEN $3 ELSE tempo END,
						ccli_number = COALESCE(NULLIF($4, ''), ccli_number),
						lyrics = COALESCE(NULLIF($5, ''), lyrics),
						notes = COALESCE(NULLIF($6, ''), notes),
						tags = COALESCE(NULLIF($7, ''), tags),
						updated_at = NOW()
					WHERE id = $8
				`,
					song.Artist, song.Key, song.Tempo, song.CCLINumber, song.Lyrics, song.Notes, song.Tags,
					existingID,
				)

				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+1, err))
					_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+savepointName)
					continue
				}

				result.Updated++
				_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT "+savepointName)
				continue
			}
		} else {
			// No match found — rollback the failed SELECT savepoint (pgx marks it failed on no rows in some cases)
			_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+savepointName)
			_, _ = tx.Exec(ctx, "SAVEPOINT "+savepointName)
		}

		// Parse last_used date if present (PCO format: "January 12, 2022")
		var lastUsed *time.Time
		if song.LastUsed != "" {
			for _, layout := range []string{
				"January 2, 2006",
				"Jan 2, 2006",
				"2006-01-02",
				"01/02/2006",
				"1/2/2006",
			} {
				if t, e := time.Parse(layout, song.LastUsed); e == nil {
					lastUsed = &t
					break
				}
			}
		}

		// Insert new song
		id := uuid.New().String()
		_, err = tx.Exec(ctx, `
			INSERT INTO songs (
				id, tenant_id, title, artist, default_key, tempo,
				ccli_number, lyrics, notes, tags, times_used, last_used,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 0, $11, NOW(), NOW()
			)`,
			id, tenantID, song.Title, song.Artist, song.Key, song.Tempo,
			song.CCLINumber, song.Lyrics, song.Notes, song.Tags, lastUsed,
		)

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %v", i+1, err))
			_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+savepointName)
			continue
		}

		result.Imported++
		_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT "+savepointName)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction (%d imported, %d errors): %w", result.Imported, len(result.Errors), err)
	}

	// Log import to history
	s.logImport(ctx, tenantID, "pco_songs", result.Imported, result.Updated, result.Skipped, len(result.Errors))

	return result, nil
}

// logImport logs an import operation to the import_history table
func (s *Service) logImport(ctx context.Context, tenantID, importType string, imported, updated, skipped, errorCount int) {
	id := uuid.New().String()
	_, _ = s.db.Exec(ctx, `
		INSERT INTO import_history (
			id, tenant_id, import_type, imported_count, updated_count, skipped_count, error_count, imported_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`, id, tenantID, importType, imported, updated, skipped, errorCount)
}

// GetImportHistory returns recent import history for a tenant
func (s *Service) GetImportHistory(ctx context.Context, tenantID string) ([]ImportHistoryRecord, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, import_type, imported_count, updated_count, skipped_count, error_count, imported_at
		FROM import_history
		ORDER BY imported_at DESC
		LIMIT 50
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query import history: %w", err)
	}
	defer rows.Close()

	var history []ImportHistoryRecord
	for rows.Next() {
		var record ImportHistoryRecord
		err := rows.Scan(
			&record.ID, &record.ImportType, &record.ImportedCount, &record.UpdatedCount,
			&record.SkippedCount, &record.ErrorCount, &record.ImportedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan import history: %w", err)
		}
		history = append(history, record)
	}

	return history, nil
}
