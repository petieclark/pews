package qr

import (
	"fmt"
)

type Service struct {
	baseURL string
}

func NewService(baseURL string) *Service {
	return &Service{
		baseURL: baseURL,
	}
}

// BuildCheckinURL builds the full URL for check-in with station ID
func (s *Service) BuildCheckinURL(tenantSubdomain, stationID string) string {
	return fmt.Sprintf("%s/checkin?station=%s", s.getTenantURL(tenantSubdomain), stationID)
}

// BuildConnectURL builds the full URL for connection cards
func (s *Service) BuildConnectURL(tenantSubdomain string) string {
	return fmt.Sprintf("%s/connect", s.getTenantURL(tenantSubdomain))
}

// BuildGiveURL builds the full URL for giving
func (s *Service) BuildGiveURL(tenantSubdomain string) string {
	return fmt.Sprintf("%s/give", s.getTenantURL(tenantSubdomain))
}

// BuildPrayerURL builds the full URL for prayer submission
func (s *Service) BuildPrayerURL(tenantSubdomain string) string {
	return fmt.Sprintf("%s/prayer/submit", s.getTenantURL(tenantSubdomain))
}

// getTenantURL returns the full tenant URL with subdomain
func (s *Service) getTenantURL(tenantSubdomain string) string {
	// Handle base URL that might already be a full domain (e.g., https://demo.pews.app)
	// or just a base domain (e.g., pews.app)
	if tenantSubdomain == "" {
		return s.baseURL
	}
	
	// If baseURL already contains subdomain placeholder or is the actual domain
	// For example: https://pews.app -> https://demo.pews.app
	return fmt.Sprintf("https://%s.pews.app", tenantSubdomain)
}
