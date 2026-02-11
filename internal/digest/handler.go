package digest

import (
	"encoding/json"
	"net/http"

	"github.com/petieclark/pews/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetSettings retrieves digest settings
func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	settings, err := h.service.GetSettings(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateSettings updates digest settings
type UpdateSettingsRequest struct {
	Enabled    bool     `json:"enabled"`
	SendDay    string   `json:"send_day"`
	Recipients []string `json:"recipients"`
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate send_day
	validDays := map[string]bool{
		"monday": true, "tuesday": true, "wednesday": true,
		"thursday": true, "friday": true, "saturday": true, "sunday": true,
	}
	if !validDays[req.SendDay] {
		http.Error(w, "Invalid send_day", http.StatusBadRequest)
		return
	}

	settings, err := h.service.UpdateSettings(r.Context(), claims.TenantID, req.Enabled, req.SendDay, req.Recipients)
	if err != nil {
		http.Error(w, "Failed to update settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// PreviewDigest generates and returns the digest HTML
func (h *Handler) PreviewDigest(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	digest, err := h.service.GenerateWeeklyDigest(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to generate digest: "+err.Error(), http.StatusInternalServerError)
		return
	}

	html, err := h.service.RenderDigestHTML(digest)
	if err != nil {
		http.Error(w, "Failed to render digest: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// GetDigestData returns the digest data as JSON (for frontend preview)
func (h *Handler) GetDigestData(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	digest, err := h.service.GenerateWeeklyDigest(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to generate digest: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(digest)
}
