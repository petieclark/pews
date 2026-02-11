package engagement

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

// GetAllScores returns all engagement scores (admin)
// GET /api/engagement/scores
func (h *Handler) GetAllScores(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenant_id").(string); if !ok { http.Error(w, "unauthorized", http.StatusUnauthorized); return }

	scores, err := h.service.GetAllScores(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

// GetPersonScore returns individual score breakdown
// GET /api/engagement/scores/:personID
func (h *Handler) GetPersonScore(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenant_id").(string); if !ok { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
	personID := chi.URLParam(r, "personID")

	score, err := h.service.GetPersonScore(r.Context(), tenantID, personID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(score)
}

// CalculatePersonScore calculates/updates engagement score for a person
// POST /api/engagement/scores/:personID/calculate
func (h *Handler) CalculatePersonScore(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenant_id").(string); if !ok { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
	personID := chi.URLParam(r, "personID")

	score, err := h.service.CalculateScore(r.Context(), tenantID, personID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(score)
}

// GetAtRiskPeople returns people with declining engagement
// GET /api/engagement/at-risk
func (h *Handler) GetAtRiskPeople(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenant_id").(string); if !ok { http.Error(w, "unauthorized", http.StatusUnauthorized); return }

	atRisk, err := h.service.GetAtRiskPeople(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(atRisk)
}

// GetDashboardKPIs returns key dashboard metrics
// GET /api/dashboard/kpis
func (h *Handler) GetDashboardKPIs(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenant_id").(string); if !ok { http.Error(w, "unauthorized", http.StatusUnauthorized); return }

	kpis, err := h.service.GetDashboardKPIs(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(kpis)
}

// RecalculateAllScores triggers recalculation for all people
// POST /api/engagement/recalculate
func (h *Handler) RecalculateAllScores(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenant_id").(string); if !ok { http.Error(w, "unauthorized", http.StatusUnauthorized); return }

	err := h.service.RecalculateAllScores(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"message": "Engagement scores recalculated for all active members",
	})
}
