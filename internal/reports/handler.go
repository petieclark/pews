package reports

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

// GetAttendanceReport returns weekly attendance trends
func (h *Handler) GetAttendanceReport(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	report, err := h.service.GetAttendanceReport(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to generate attendance report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetGivingReport returns monthly giving trends
func (h *Handler) GetGivingReport(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	report, err := h.service.GetGivingReport(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to generate giving report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetMembershipReport returns membership growth trends
func (h *Handler) GetMembershipReport(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	report, err := h.service.GetMembershipReport(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to generate membership report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetGroupParticipationReport returns group participation stats
func (h *Handler) GetGroupParticipationReport(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	report, err := h.service.GetGroupParticipationReport(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to generate group participation report: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}
