package reports

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/petieclark/pews/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func getTenantUUID(r *http.Request) (uuid.UUID, *middleware.Claims, error) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		return uuid.UUID{}, nil, http.ErrNoCookie // will be caught
	}
	id, err := uuid.Parse(claims.TenantID)
	return id, claims, err
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) GetAttendanceReport(w http.ResponseWriter, r *http.Request) {
	tenantID, _, err := getTenantUUID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rangeStr := r.URL.Query().Get("range")
	report, err := h.service.GetAttendanceReport(r.Context(), tenantID, rangeStr)
	if err != nil {
		http.Error(w, "Failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, report)
}

func (h *Handler) GetGivingReport(w http.ResponseWriter, r *http.Request) {
	tenantID, _, err := getTenantUUID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rangeStr := r.URL.Query().Get("range")
	report, err := h.service.GetGivingReport(r.Context(), tenantID, rangeStr)
	if err != nil {
		http.Error(w, "Failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, report)
}

func (h *Handler) GetGrowthReport(w http.ResponseWriter, r *http.Request) {
	tenantID, _, err := getTenantUUID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rangeStr := r.URL.Query().Get("range")
	report, err := h.service.GetGrowthReport(r.Context(), tenantID, rangeStr)
	if err != nil {
		http.Error(w, "Failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, report)
}

func (h *Handler) GetSongsReport(w http.ResponseWriter, r *http.Request) {
	tenantID, _, err := getTenantUUID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	report, err := h.service.GetSongsReport(r.Context(), tenantID)
	if err != nil {
		http.Error(w, "Failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, report)
}

func (h *Handler) GetEngagementReport(w http.ResponseWriter, r *http.Request) {
	tenantID, _, err := getTenantUUID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	report, err := h.service.GetEngagementReport(r.Context(), tenantID)
	if err != nil {
		http.Error(w, "Failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, report)
}

// Legacy endpoints kept for backward compatibility
func (h *Handler) GetMembershipReport(w http.ResponseWriter, r *http.Request) {
	h.GetGrowthReport(w, r)
}

func (h *Handler) GetGroupParticipationReport(w http.ResponseWriter, r *http.Request) {
	tenantID, _, err := getTenantUUID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Keep legacy group endpoint working
	report, err := h.service.GetGrowthReport(r.Context(), tenantID, "12m")
	if err != nil {
		http.Error(w, "Failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, report)
}
