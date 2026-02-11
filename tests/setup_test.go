package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petieclark/pews/internal/database"
)

var (
	testDB     *database.DB
	testDBURL  string
	jwtSecret  = "test-jwt-secret-key-for-testing-only"
	frontendURL = "http://localhost:5173"
)

// TestMain runs before all tests and sets up the test database
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Use test database URL
	testDBURL = os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "postgres://postgres:postgres@localhost:5432/pews_test?sslmode=disable"
	}

	// Connect to database
	var err error
	testDB, err = database.New(ctx, testDBURL)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := testDB.Migrate(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("Test database initialized successfully")

	// Run tests
	code := m.Run()

	// Cleanup
	cleanupDatabase(ctx)
	testDB.Close()

	os.Exit(code)
}

// cleanupDatabase truncates all tables between tests
func cleanupDatabase(ctx context.Context) {
	tables := []string{
		"sms_messages",
		"pledges",
		"connection_card_tags",
		"connection_cards",
		"journey_enrollments",
		"journey_steps",
		"journeys",
		"campaign_recipients",
		"campaigns",
		"templates",
		"chat_messages",
		"stream_viewers",
		"streams",
		"service_team",
		"service_items",
		"services",
		"service_types",
		"songs",
		"group_members",
		"groups",
		"household_members",
		"households",
		"person_tags",
		"tags",
		"recurring_donations",
		"donations",
		"funds",
		"checkin_pickups",
		"checkin_alerts",
		"checkins",
		"checkin_events",
		"checkin_stations",
		"people",
		"subscriptions",
		"enabled_modules",
		"modules",
		"users",
		"tenants",
	}

	for _, table := range tables {
		_, err := testDB.Pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			// Ignore errors for tables that might not exist
			log.Printf("Warning: failed to truncate table %s: %v", table, err)
		}
	}
}

// getTestPool returns the test database pool
func getTestPool() *pgxpool.Pool {
	return testDB.Pool
}

// setupTestTenant creates a test tenant and returns its ID
func setupTestTenant(ctx context.Context, name string) (string, error) {
	var tenantID string
	err := testDB.Pool.QueryRow(ctx, `
		INSERT INTO tenants (name, slug, settings)
		VALUES ($1, $2, '{}'::jsonb)
		RETURNING id
	`, name, slugify(name)).Scan(&tenantID)
	return tenantID, err
}

// slugify converts a name to a URL-friendly slug
func slugify(s string) string {
	// Simple slugification (in production, use a proper library)
	slug := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			slug += string(c)
		} else if c >= 'A' && c <= 'Z' {
			slug += string(c + 32) // Convert to lowercase
		} else if c == ' ' || c == '-' {
			slug += "-"
		}
	}
	return slug
}

// setupTestUser creates a test user for the given tenant
func setupTestUser(ctx context.Context, tenantID, email, password, role string) (string, error) {
	var userID string
	// In a real implementation, password should be hashed
	// For testing, we'll use a simple hash (in production tests, use the actual auth service)
	passwordHash := "$2a$10$dummy.hash.for.testing.purposes.only" // Placeholder
	err := testDB.Pool.QueryRow(ctx, `
		INSERT INTO users (tenant_id, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, tenantID, email, passwordHash, role).Scan(&userID)
	return userID, err
}
