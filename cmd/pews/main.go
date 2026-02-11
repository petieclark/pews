package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/petieclark/pews/internal/activity"
	"github.com/petieclark/pews/internal/audit"
	"github.com/petieclark/pews/internal/auth"
	"github.com/petieclark/pews/internal/billing"
	"github.com/petieclark/pews/internal/calendar"
	"github.com/petieclark/pews/internal/care"
	"github.com/petieclark/pews/internal/checkins"
	"github.com/petieclark/pews/internal/communication"
	"github.com/petieclark/pews/internal/config"
	"github.com/petieclark/pews/internal/database"
	"github.com/petieclark/pews/internal/drip"
	"github.com/petieclark/pews/internal/engagement"
	"github.com/petieclark/pews/internal/giving"
	"github.com/petieclark/pews/internal/groups"
	"github.com/petieclark/pews/internal/i18n"
	importpkg "github.com/petieclark/pews/internal/import"
	"github.com/petieclark/pews/internal/module"
	"github.com/petieclark/pews/internal/notification"
	"github.com/petieclark/pews/internal/people"
	"github.com/petieclark/pews/internal/prayer"
	"github.com/petieclark/pews/internal/qr"
	"github.com/petieclark/pews/internal/reports"
	"github.com/petieclark/pews/internal/router"
	"github.com/petieclark/pews/internal/search"
	"github.com/petieclark/pews/internal/services"
	"github.com/petieclark/pews/internal/sermons"
	"github.com/petieclark/pews/internal/sms"
	"github.com/petieclark/pews/internal/streaming"
	"github.com/petieclark/pews/internal/teams"
	"github.com/petieclark/pews/internal/tenant"
	"github.com/petieclark/pews/internal/website"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Connect to database
	ctx := context.Background()
	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	// Run migrations
	if err := db.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed")

	// Initialize services
	activityService := activity.NewService(db.Pool)
	authService := auth.NewService(db.Pool, cfg.JWTSecret)
	tenantService := tenant.NewService(db.Pool)
	moduleService := module.NewService(db.Pool)
	billingService := billing.NewService(db.Pool, cfg.StripeSecretKey, cfg.StripePriceID, cfg.FrontendURL)
	peopleService := people.NewService(db.Pool)
	groupsService := groups.NewService(db.Pool)
	servicesService := services.NewService(db.Pool)
	sermonsService := sermons.NewService(db.Pool)
	givingService := giving.NewService(db.Pool)
	givingStripeService := giving.NewStripeService(db.Pool, cfg.StripeSecretKey, cfg.FrontendURL)
	streamingService := streaming.NewService(db.Pool)
	communicationService := communication.NewService(db.Pool)
	dripService := drip.NewService(db.Pool)
	checkinsService := checkins.NewService(db.Pool)
	reportsService := reports.NewService(db.Pool)
	calendarService := calendar.NewService(db.Pool)
	careService := care.NewService(db.Pool)
	prayerService := prayer.NewService(db.Pool)
	searchService := search.NewService(db.Pool)
	notificationService := notification.NewService(db.Pool)
	websiteService := website.NewService(db.Pool)
	qrService := qr.NewService(cfg.FrontendURL)
	smsService := sms.NewService(db.Pool)
	i18nService := i18n.NewService()
	auditService := audit.NewService(db.DB)

	// Initialize handlers
	authHandler := auth.NewHandler(authService, tenantService, billingService, auditService)
	tenantHandler := tenant.NewHandler(tenantService)
	moduleHandler := module.NewHandler(moduleService)
	billingHandler := billing.NewHandler(billingService)
	peopleHandler := people.NewHandler(peopleService, activityService)
	groupsHandler := groups.NewHandler(groupsService)
	servicesHandler := services.NewHandler(servicesService)
	sermonsHandler := sermons.NewHandler(sermonsService)
	givingHandler := giving.NewHandler(givingService, givingStripeService, activityService)
	streamingHandler := streaming.NewHandler(streamingService)
	communicationHandler := communication.NewHandler(communicationService)
	dripHandler := drip.NewHandler(dripService)
	checkinsHandler := checkins.NewHandler(checkinsService)
	reportsHandler := reports.NewHandler(reportsService)
	calendarHandler := calendar.NewHandler(calendarService)
	careHandler := care.NewHandler(careService)
	prayerHandler := prayer.NewHandler(prayerService)
	searchHandler := search.NewHandler(searchService)
	notificationHandler := notification.NewHandler(notificationService)
	websiteHandler := website.NewHandler(websiteService)
	qrHandler := qr.NewHandler(qrService)
	
	// Engagement (Phase 6)
	engagementService := engagement.NewService(db.Pool)
	engagementHandler := engagement.NewHandler(engagementService)
	
	// SMS (Phase 7)
	smsHandler := sms.NewHandler(smsService)
	
	// i18n (Phase 8)
	i18nHandler := i18n.NewHandler(i18nService)
	
	// Import
	importService := importpkg.NewService(db.Pool)
	importHandler := importpkg.NewHandler(importService)

	// Teams
	teamsService := teams.NewService(db.Pool)
	teamsHandler := teams.NewHandler(teamsService)

	// Setup router
	r := router.New(
		authHandler,
		authService,
		tenantHandler,
		moduleHandler,
		billingHandler,
		peopleHandler,
		groupsHandler,
		servicesHandler,
		sermonsHandler,
		givingHandler,
		streamingHandler,
		communicationHandler,
		dripHandler,
		checkinsHandler,
		reportsHandler,
		calendarHandler,
		prayerHandler,
		searchHandler,
		notificationHandler,
		websiteHandler,
		qrHandler,
		engagementHandler,
		smsHandler,
		i18nHandler,
		importHandler,
		teamsHandler,
		careHandler,
		cfg.StripeWebhookSecret,
		cfg.StripeWebhookSecret, // Use same webhook secret for giving
		cfg.FrontendURL,
	)

	// Start server
	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited")
	return nil
}
