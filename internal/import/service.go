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
