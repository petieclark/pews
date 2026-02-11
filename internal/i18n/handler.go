package i18n

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetTranslations handles GET /api/i18n/:locale
func (h *Handler) GetTranslations(w http.ResponseWriter, r *http.Request) {
	locale := chi.URLParam(r, "locale")
	
	if locale == "" {
		locale = "en" // Default to English
	}

	translations, err := h.service.GetTranslations(locale)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	json.NewEncoder(w).Encode(translations)
}

// GetSupportedLocales handles GET /api/i18n/locales
func (h *Handler) GetSupportedLocales(w http.ResponseWriter, r *http.Request) {
	locales := h.service.GetSupportedLocales()
	
	response := map[string]interface{}{
		"locales": locales,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
