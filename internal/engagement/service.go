package engagement

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

// CalculateScore computes engagement score for a person based on:
// - Attendance frequency (weight: 30%)
// - Giving consistency (weight: 20%)
// - Group participation (weight: 20%)
// - Volunteering (weight: 15%)
// - Connection (completed profile, in directory) (weight: 15%)
func (s *Service) CalculateScore(ctx context.Context, tenantID, personID string) (*EngagementScore, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Calculate attendance score (0-100, weight 30%)
	attendanceScore, err := s.calculateAttendanceScore(ctx, tenantID, personID)
	if err != nil {
		return nil, err
	}

	// Calculate giving score (0-100, weight 20%)
	givingScore, err := s.calculateGivingScore(ctx, tenantID, personID)
	if err != nil {
		return nil, err
	}

	// Calculate group participation score (0-100, weight 20%)
	groupScore, err := s.calculateGroupScore(ctx, tenantID, personID)
	if err != nil {
		return nil, err
	}

	// Calculate volunteer score (0-100, weight 15%)
	volunteerScore, err := s.calculateVolunteerScore(ctx, tenantID, personID)
	if err != nil {
		return nil, err
	}

	// Calculate connection score (0-100, weight 15%)
	connectionScore, err := s.calculateConnectionScore(ctx, tenantID, personID)
	if err != nil {
		return nil, err
	}

	// Weighted total score
	totalScore := int(
		float64(attendanceScore)*0.30 +
			float64(givingScore)*0.20 +
			float64(groupScore)*0.20 +
			float64(volunteerScore)*0.15 +
			float64(connectionScore)*0.15,
	)

	// Upsert engagement score
	score := &EngagementScore{
		TenantID:        tenantID,
		PersonID:        personID,
		Score:           totalScore,
		AttendanceScore: attendanceScore,
		GivingScore:     givingScore,
		GroupScore:      groupScore,
		VolunteerScore:  volunteerScore,
		ConnectionScore: connectionScore,
		CalculatedAt:    time.Now(),
	}

	err = s.db.QueryRow(ctx, `
		INSERT INTO engagement_scores (
			id, tenant_id, person_id, score,
			attendance_score, giving_score, group_score,
			volunteer_score, connection_score, calculated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (tenant_id, person_id) 
		DO UPDATE SET 
			score = $4,
			attendance_score = $5,
			giving_score = $6,
			group_score = $7,
			volunteer_score = $8,
			connection_score = $9,
			calculated_at = $10
		RETURNING id, calculated_at`,
		uuid.New().String(), tenantID, personID, totalScore,
		attendanceScore, givingScore, groupScore,
		volunteerScore, connectionScore, time.Now(),
	).Scan(&score.ID, &score.CalculatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to save engagement score: %w", err)
	}

	return score, nil
}

func (s *Service) calculateAttendanceScore(ctx context.Context, tenantID, personID string) (int, error) {
	// Check attendance in last 12 weeks
	// 100 points = attended 10+ times (83%+)
	// 75 points = attended 8-9 times (67-75%)
	// 50 points = attended 5-7 times (42-58%)
	// 25 points = attended 2-4 times (17-33%)
	// 0 points = attended 0-1 times (0-8%)

	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT ce.event_date)
		FROM checkins c
		JOIN checkin_events ce ON ce.id = c.event_id
		WHERE c.person_id = $1 
		  AND c.tenant_id = $2
		  AND ce.event_date >= CURRENT_DATE - INTERVAL '12 weeks'`,
		personID, tenantID,
	).Scan(&count)

	if err != nil && err != pgx.ErrNoRows {
		return 0, fmt.Errorf("failed to calculate attendance score: %w", err)
	}

	switch {
	case count >= 10:
		return 100, nil
	case count >= 8:
		return 75, nil
	case count >= 5:
		return 50, nil
	case count >= 2:
		return 25, nil
	default:
		return 0, nil
	}
}

func (s *Service) calculateGivingScore(ctx context.Context, tenantID, personID string) (int, error) {
	// Check giving consistency in last 6 months
	// 100 points = gave 5+ months (83%+)
	// 75 points = gave 4 months (67%)
	// 50 points = gave 2-3 months (33-50%)
	// 25 points = gave 1 month (17%)
	// 0 points = no giving

	var monthsWithGiving int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT DATE_TRUNC('month', donated_at))
		FROM donations
		WHERE person_id = $1 
		  AND tenant_id = $2
		  AND status = 'completed'
		  AND donated_at >= CURRENT_DATE - INTERVAL '6 months'`,
		personID, tenantID,
	).Scan(&monthsWithGiving)

	if err != nil && err != pgx.ErrNoRows {
		return 0, fmt.Errorf("failed to calculate giving score: %w", err)
	}

	switch {
	case monthsWithGiving >= 5:
		return 100, nil
	case monthsWithGiving == 4:
		return 75, nil
	case monthsWithGiving >= 2:
		return 50, nil
	case monthsWithGiving == 1:
		return 25, nil
	default:
		return 0, nil
	}
}

func (s *Service) calculateGroupScore(ctx context.Context, tenantID, personID string) (int, error) {
	// Check active group participation
	// 100 points = member of 2+ active groups
	// 75 points = member of 1 active group + joined recently (within 3 months)
	// 50 points = member of 1 active group + joined 3-12 months ago
	// 25 points = member of 1 active group + joined 12+ months ago
	// 0 points = not in any active groups

	var groupCount int
	var recentJoin bool

	err := s.db.QueryRow(ctx, `
		SELECT 
			COUNT(*) as group_count,
			BOOL_OR(gm.joined_at >= CURRENT_DATE - INTERVAL '3 months') as recent_join
		FROM group_members gm
		JOIN groups g ON g.id = gm.group_id
		WHERE gm.person_id = $1 
		  AND g.tenant_id = $2
		  AND g.is_active = TRUE`,
		personID, tenantID,
	).Scan(&groupCount, &recentJoin)

	if err != nil && err != pgx.ErrNoRows {
		return 0, fmt.Errorf("failed to calculate group score: %w", err)
	}

	if groupCount >= 2 {
		return 100, nil
	} else if groupCount == 1 {
		if recentJoin {
			return 75, nil
		}
		return 50, nil
	}

	return 0, nil
}

func (s *Service) calculateVolunteerScore(ctx context.Context, tenantID, personID string) (int, error) {
	// Check volunteering activity (service team participation)
	// 100 points = served 8+ times in last 12 weeks
	// 75 points = served 5-7 times
	// 50 points = served 3-4 times
	// 25 points = served 1-2 times
	// 0 points = no serving

	var serveCount int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT s.id)
		FROM service_team_members stm
		JOIN services s ON s.id = stm.service_id
		WHERE stm.person_id = $1 
		  AND s.tenant_id = $2
		  AND s.service_date >= CURRENT_DATE - INTERVAL '12 weeks'`,
		personID, tenantID,
	).Scan(&serveCount)

	if err != nil && err != pgx.ErrNoRows {
		return 0, fmt.Errorf("failed to calculate volunteer score: %w", err)
	}

	switch {
	case serveCount >= 8:
		return 100, nil
	case serveCount >= 5:
		return 75, nil
	case serveCount >= 3:
		return 50, nil
	case serveCount >= 1:
		return 25, nil
	default:
		return 0, nil
	}
}

func (s *Service) calculateConnectionScore(ctx context.Context, tenantID, personID string) (int, error) {
	// Check connection level
	// 100 points = profile completed + in directory + has household
	// 75 points = profile completed + in directory
	// 50 points = profile completed OR in directory
	// 25 points = has email or phone
	// 0 points = minimal info

	var profileCompleted, inDirectory bool
	var hasEmail, hasPhone, hasHousehold bool

	err := s.db.QueryRow(ctx, `
		SELECT 
			COALESCE(p.profile_completed, FALSE) as profile_completed,
			COALESCE(p.in_directory, FALSE) as in_directory,
			p.email IS NOT NULL AND p.email != '' as has_email,
			p.phone IS NOT NULL AND p.phone != '' as has_phone,
			EXISTS(SELECT 1 FROM household_members hm WHERE hm.person_id = p.id) as has_household
		FROM people p
		WHERE p.id = $1 AND p.tenant_id = $2`,
		personID, tenantID,
	).Scan(&profileCompleted, &inDirectory, &hasEmail, &hasPhone, &hasHousehold)

	if err != nil {
		return 0, fmt.Errorf("failed to calculate connection score: %w", err)
	}

	if profileCompleted && inDirectory && hasHousehold {
		return 100, nil
	} else if profileCompleted && inDirectory {
		return 75, nil
	} else if profileCompleted || inDirectory {
		return 50, nil
	} else if hasEmail || hasPhone {
		return 25, nil
	}

	return 0, nil
}

// GetAllScores returns all engagement scores for a tenant
func (s *Service) GetAllScores(ctx context.Context, tenantID string) ([]EngagementScore, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT 
			es.id, es.tenant_id, es.person_id, es.score,
			es.attendance_score, es.giving_score, es.group_score,
			es.volunteer_score, es.connection_score, es.calculated_at,
			COALESCE(p.first_name, '') || ' ' || COALESCE(p.last_name, '') as person_name,
			p.email as person_email
		FROM engagement_scores es
		JOIN people p ON p.id = es.person_id
		WHERE es.tenant_id = $1
		ORDER BY es.score DESC`, tenantID)

	if err != nil {
		return nil, fmt.Errorf("failed to get all scores: %w", err)
	}
	defer rows.Close()

	scores := []EngagementScore{}
	for rows.Next() {
		var score EngagementScore
		err := rows.Scan(
			&score.ID, &score.TenantID, &score.PersonID, &score.Score,
			&score.AttendanceScore, &score.GivingScore, &score.GroupScore,
			&score.VolunteerScore, &score.ConnectionScore, &score.CalculatedAt,
			&score.PersonName, &score.PersonEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan score: %w", err)
		}
		scores = append(scores, score)
	}

	return scores, nil
}

// GetPersonScore returns engagement score for a specific person
func (s *Service) GetPersonScore(ctx context.Context, tenantID, personID string) (*EngagementScore, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	var score EngagementScore
	err = s.db.QueryRow(ctx, `
		SELECT 
			es.id, es.tenant_id, es.person_id, es.score,
			es.attendance_score, es.giving_score, es.group_score,
			es.volunteer_score, es.connection_score, es.calculated_at,
			COALESCE(p.first_name, '') || ' ' || COALESCE(p.last_name, '') as person_name,
			p.email as person_email
		FROM engagement_scores es
		JOIN people p ON p.id = es.person_id
		WHERE es.person_id = $1 AND es.tenant_id = $2`,
		personID, tenantID,
	).Scan(
		&score.ID, &score.TenantID, &score.PersonID, &score.Score,
		&score.AttendanceScore, &score.GivingScore, &score.GroupScore,
		&score.VolunteerScore, &score.ConnectionScore, &score.CalculatedAt,
		&score.PersonName, &score.PersonEmail,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("engagement score not found")
		}
		return nil, fmt.Errorf("failed to get person score: %w", err)
	}

	return &score, nil
}

// GetAtRiskPeople returns people whose engagement dropped >20% in last 30 days
func (s *Service) GetAtRiskPeople(ctx context.Context, tenantID string) ([]AtRiskPerson, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	rows, err := s.db.Query(ctx, `
		WITH current_scores AS (
			SELECT person_id, score, calculated_at
			FROM engagement_scores
			WHERE tenant_id = $1
		),
		previous_scores AS (
			SELECT DISTINCT ON (person_id) 
				person_id, score
			FROM engagement_score_history
			WHERE tenant_id = $1
			  AND recorded_at <= CURRENT_TIMESTAMP - INTERVAL '30 days'
			ORDER BY person_id, recorded_at DESC
		)
		SELECT 
			cs.person_id,
			COALESCE(p.first_name, '') || ' ' || COALESCE(p.last_name, '') as person_name,
			p.email,
			cs.score as current_score,
			COALESCE(ps.score, cs.score) as previous_score,
			cs.score - COALESCE(ps.score, cs.score) as score_change,
			CASE 
				WHEN COALESCE(ps.score, 0) = 0 THEN 0
				ELSE ((cs.score - COALESCE(ps.score, cs.score))::FLOAT / ps.score) * 100
			END as percent_change,
			cs.calculated_at
		FROM current_scores cs
		LEFT JOIN previous_scores ps ON ps.person_id = cs.person_id
		JOIN people p ON p.id = cs.person_id
		WHERE CASE 
			WHEN COALESCE(ps.score, 0) = 0 THEN FALSE
			ELSE ((cs.score - COALESCE(ps.score, cs.score))::FLOAT / ps.score) * 100 <= -20
		END
		ORDER BY percent_change ASC`, tenantID)

	if err != nil {
		return nil, fmt.Errorf("failed to get at-risk people: %w", err)
	}
	defer rows.Close()

	atRisk := []AtRiskPerson{}
	for rows.Next() {
		var person AtRiskPerson
		err := rows.Scan(
			&person.PersonID, &person.PersonName, &person.PersonEmail,
			&person.CurrentScore, &person.PreviousScore, &person.ScoreChange,
			&person.PercentChange, &person.LastCalculated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan at-risk person: %w", err)
		}
		atRisk = append(atRisk, person)
	}

	return atRisk, nil
}

// GetDashboardKPIs returns key metrics for the admin dashboard
func (s *Service) GetDashboardKPIs(ctx context.Context, tenantID string) (*DashboardKPIs, error) {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	kpis := &DashboardKPIs{}

	// Total active members
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*) 
		FROM people 
		WHERE tenant_id = $1 AND membership_status = 'active'`,
		tenantID,
	).Scan(&kpis.TotalActiveMembers)
	if err != nil {
		return nil, fmt.Errorf("failed to get active members: %w", err)
	}

	// Average attendance (last 4 weeks)
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(attendee_count), 0)
		FROM (
			SELECT COUNT(DISTINCT c.person_id) as attendee_count
			FROM checkins c
			JOIN checkin_events ce ON ce.id = c.event_id
			WHERE c.tenant_id = $1
			  AND ce.event_date >= CURRENT_DATE - INTERVAL '4 weeks'
			GROUP BY ce.event_date
		) weekly_attendance`,
		tenantID,
	).Scan(&kpis.AverageAttendance)
	if err != nil {
		return nil, fmt.Errorf("failed to get average attendance: %w", err)
	}

	// Giving this month vs last month
	err = s.db.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(CASE WHEN DATE_TRUNC('month', donated_at) = DATE_TRUNC('month', CURRENT_DATE) THEN amount_cents ELSE 0 END), 0) as this_month,
			COALESCE(SUM(CASE WHEN DATE_TRUNC('month', donated_at) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month') THEN amount_cents ELSE 0 END), 0) as last_month
		FROM donations
		WHERE tenant_id = $1 AND status = 'completed'
		  AND donated_at >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')`,
		tenantID,
	).Scan(&kpis.GivingThisMonth, &kpis.GivingLastMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get giving stats: %w", err)
	}

	if kpis.GivingLastMonth > 0 {
		kpis.GivingPercentChange = float64(kpis.GivingThisMonth-kpis.GivingLastMonth) / float64(kpis.GivingLastMonth) * 100
	}

	// New visitors this month
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT c.person_id)
		FROM checkins c
		JOIN checkin_events ce ON ce.id = c.event_id
		WHERE c.tenant_id = $1
		  AND c.first_time = TRUE
		  AND ce.event_date >= DATE_TRUNC('month', CURRENT_DATE)`,
		tenantID,
	).Scan(&kpis.NewVisitorsThisMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get new visitors: %w", err)
	}

	// Engagement distribution
	err = s.db.QueryRow(ctx, `
		SELECT 
			COUNT(CASE WHEN score >= 75 THEN 1 END) as high,
			COUNT(CASE WHEN score >= 50 AND score < 75 THEN 1 END) as medium,
			COUNT(CASE WHEN score >= 25 AND score < 50 THEN 1 END) as low,
			COUNT(CASE WHEN score < 25 THEN 1 END) as inactive
		FROM engagement_scores
		WHERE tenant_id = $1`,
		tenantID,
	).Scan(&kpis.EngagementDistribution.High, &kpis.EngagementDistribution.Medium,
		&kpis.EngagementDistribution.Low, &kpis.EngagementDistribution.Inactive)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get engagement distribution: %w", err)
	}

	// Attendance trend (last 8 weeks)
	rows, err := s.db.Query(ctx, `
		SELECT COUNT(DISTINCT c.person_id) as attendees
		FROM checkins c
		JOIN checkin_events ce ON ce.id = c.event_id
		WHERE c.tenant_id = $1
		  AND ce.event_date >= CURRENT_DATE - INTERVAL '8 weeks'
		GROUP BY DATE_TRUNC('week', ce.event_date)
		ORDER BY DATE_TRUNC('week', ce.event_date) ASC`,
		tenantID,
	)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get attendance trend: %w", err)
	}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var count int
			if err := rows.Scan(&count); err != nil {
				return nil, fmt.Errorf("failed to scan attendance trend: %w", err)
			}
			kpis.AttendanceTrend = append(kpis.AttendanceTrend, count)
		}
	}

	return kpis, nil
}

// RecalculateAllScores recalculates engagement scores for all active people in tenant
func (s *Service) RecalculateAllScores(ctx context.Context, tenantID string) error {
	// Set tenant context
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}

	// Get all active people
	rows, err := s.db.Query(ctx, `
		SELECT id FROM people 
		WHERE tenant_id = $1 AND membership_status = 'active'`,
		tenantID,
	)
	if err != nil {
		return fmt.Errorf("failed to query people: %w", err)
	}
	defer rows.Close()

	personIDs := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan person ID: %w", err)
		}
		personIDs = append(personIDs, id)
	}

	// Calculate score for each person
	for _, personID := range personIDs {
		_, err := s.CalculateScore(ctx, tenantID, personID)
		if err != nil {
			// Log error but continue
			fmt.Printf("Error calculating score for person %s: %v\n", personID, err)
		}
	}

	return nil
}

// GetDashboardActivity returns recent activity across all modules
func (s *Service) GetDashboardActivity(ctx context.Context, tenantID string) ([]DashboardActivity, error) {
	_, err := s.db.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, TRUE)", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to set tenant context: %w", err)
	}

	activities := []DashboardActivity{}

	// Recent activity log entries
	rows, err := s.db.Query(ctx, `
		SELECT al.id, al.action, al.entity_type, al.entity_id, al.details, al.created_at,
		       COALESCE(u.email, 'System') as user_email
		FROM activity_log al
		LEFT JOIN users u ON al.user_id = u.id
		ORDER BY al.created_at DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id, action, entityType string
			var entityID *string
			var details []byte
			var createdAt time.Time
			var userEmail string
			if err := rows.Scan(&id, &action, &entityType, &entityID, &details, &createdAt, &userEmail); err != nil {
				continue
			}

			activity := DashboardActivity{
				ID:        id,
				Type:      entityType,
				Timestamp: createdAt,
			}

			// Parse details for display name
			var detailMap map[string]interface{}
			if len(details) > 0 {
				json.Unmarshal(details, &detailMap)
			}
			name := ""
			if detailMap != nil {
				if n, ok := detailMap["name"].(string); ok {
					name = n
				}
			}

			switch action {
			case "person.created":
				activity.Icon = "user-plus"
				activity.Title = "New person added"
				activity.Description = name + " was added by " + userEmail
				if entityID != nil {
					activity.Link = "/dashboard/people/" + *entityID
				}
			case "person.updated":
				activity.Icon = "user-edit"
				activity.Title = "Person updated"
				activity.Description = name + " was updated"
				if entityID != nil {
					activity.Link = "/dashboard/people/" + *entityID
				}
			case "person.deleted":
				activity.Icon = "user-minus"
				activity.Title = "Person removed"
				activity.Description = name + " was removed"
			case "donation.created":
				activity.Icon = "dollar"
				activity.Title = "New donation"
				if amt, ok := detailMap["amount"]; ok {
					activity.Description = fmt.Sprintf("$%.2f received", amt)
				} else {
					activity.Description = "Donation received"
				}
				activity.Link = "/dashboard/giving"
			case "checkin.created":
				activity.Icon = "check-circle"
				activity.Title = "Check-in"
				activity.Description = name + " checked in"
			default:
				activity.Icon = "activity"
				activity.Title = strings.ReplaceAll(action, ".", " ")
				activity.Description = entityType + " activity"
			}

			activities = append(activities, activity)
		}
	}

	return activities, nil
}
