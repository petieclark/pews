package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL              string
	JWTSecret                string
	StripeSecretKey          string
	StripePublishableKey     string
	StripeWebhookSecret      string
	StripeGivingWebhookSecret string
	StripePriceID            string
	StripeTestSecretKey      string
	StripeTestPublishableKey string
	Port                     string
	FrontendURL              string
	SMSEncryptionKey         string
}

func Load() (*Config, error) {
	// Load .env file if it exists (dev environment)
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:              getEnv("DATABASE_URL", ""),
		JWTSecret:                getEnv("JWT_SECRET", ""),
		StripeSecretKey:          getEnv("STRIPE_SECRET_KEY", ""),
		StripePublishableKey:     getEnv("STRIPE_PUBLISHABLE_KEY", ""),
		StripeWebhookSecret:      getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeGivingWebhookSecret: getEnv("STRIPE_GIVING_WEBHOOK_SECRET", getEnv("STRIPE_WEBHOOK_SECRET", "")),
		StripePriceID:            getEnv("STRIPE_PRICE_ID", ""),
		StripeTestSecretKey:      getEnv("STRIPE_TEST_SECRET_KEY", ""),
		StripeTestPublishableKey: getEnv("STRIPE_TEST_PUBLISHABLE_KEY", ""),
		Port:                     getEnv("PORT", "8080"),
		FrontendURL:              getEnv("FRONTEND_URL", "http://localhost:5173"),
		SMSEncryptionKey:         getEnv("SMS_ENCRYPTION_KEY", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
