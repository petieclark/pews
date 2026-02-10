package main

import (
	"context"
	"fmt"
	"log"

	"github.com/petieclark/pews/internal/auth"
	"github.com/petieclark/pews/internal/config"
	"github.com/petieclark/pews/internal/database"
	"github.com/petieclark/pews/internal/tenant"
)

func main() {
	ctx := context.Background()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize services
	tenantService := tenant.NewService(db.Pool)
	authService := auth.NewService(db.Pool, cfg.JWTSecret)

	// Create test tenant
	log.Println("Creating test tenant...")
	testTenant, err := tenantService.CreateTenant(ctx, "Test Church")
	if err != nil {
		log.Fatal("Failed to create tenant:", err)
	}
	log.Printf("Created tenant: %s (slug: %s, id: %s)\n", testTenant.Name, testTenant.Slug, testTenant.ID)

	// Create admin user
	log.Println("Creating admin user...")
	adminUser, err := authService.CreateUser(ctx, testTenant.ID, "admin@testchurch.com", "password123", "admin")
	if err != nil {
		log.Fatal("Failed to create admin user:", err)
	}
	log.Printf("Created admin user: %s (id: %s)\n", adminUser.Email, adminUser.ID)

	// Create member user
	log.Println("Creating member user...")
	memberUser, err := authService.CreateUser(ctx, testTenant.ID, "member@testchurch.com", "password123", "member")
	if err != nil {
		log.Fatal("Failed to create member user:", err)
	}
	log.Printf("Created member user: %s (id: %s)\n", memberUser.Email, memberUser.ID)

	fmt.Println("\n✅ Seed completed successfully!")
	fmt.Println("\nTest credentials:")
	fmt.Println("  Tenant slug: test-church")
	fmt.Println("  Admin: admin@testchurch.com / password123")
	fmt.Println("  Member: member@testchurch.com / password123")
}
