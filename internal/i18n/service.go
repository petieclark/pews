package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed locales/*.json
var localesFS embed.FS

var (
	translations map[string]map[string]string
	once         sync.Once
	loadErr      error
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

// GetTranslations returns all translations for a given locale
func (s *Service) GetTranslations(locale string) (map[string]string, error) {
	// Load translations on first access
	once.Do(func() {
		loadErr = loadTranslations()
	})

	if loadErr != nil {
		return nil, loadErr
	}

	trans, ok := translations[locale]
	if !ok {
		// Fallback to English if locale not found
		trans, ok = translations["en"]
		if !ok {
			return nil, fmt.Errorf("locale not found: %s", locale)
		}
	}

	return trans, nil
}

// GetSupportedLocales returns list of available locales
func (s *Service) GetSupportedLocales() []string {
	once.Do(func() {
		loadErr = loadTranslations()
	})

	locales := make([]string, 0, len(translations))
	for locale := range translations {
		locales = append(locales, locale)
	}
	return locales
}

// loadTranslations loads all translation files from embedded filesystem
func loadTranslations() error {
	translations = make(map[string]map[string]string)
	
	locales := []string{"en", "es", "pt", "ko"}
	
	for _, locale := range locales {
		filename := fmt.Sprintf("locales/%s.json", locale)
		data, err := localesFS.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", filename, err)
		}

		var trans map[string]string
		if err := json.Unmarshal(data, &trans); err != nil {
			return fmt.Errorf("failed to parse %s: %w", filename, err)
		}

		translations[locale] = trans
	}

	return nil
}
