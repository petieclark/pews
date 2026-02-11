package drip

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ===== CAMPAIGNS =====

func (h *Handler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaigns, err := h.service.ListCampaigns(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaigns)
}

func (h *Handler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "id")

	campaign, err := h.service.GetCampaign(r.Context(), claims.TenantID, campaignID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaign)
}

func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate trigger_event
	validTriggers := map[string]bool{
		"new_member":       true,
		"connection_card":  true,
		"first_visit":      true,
	}
	if !validTriggers[req.TriggerEvent] {
		http.Error(w, "Invalid trigger_event. Must be one of: new_member, connection_card, first_visit", http.StatusBadRequest)
		return
	}

	campaign, err := h.service.CreateCampaign(r.Context(), claims.TenantID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(campaign)
}

func (h *Handler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "id")

	var req UpdateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	campaign, err := h.service.UpdateCampaign(r.Context(), claims.TenantID, campaignID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaign)
}

func (h *Handler) DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "id")

	err := h.service.DeleteCampaign(r.Context(), claims.TenantID, campaignID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ===== STEPS =====

func (h *Handler) ListSteps(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "id")

	steps, err := h.service.ListSteps(r.Context(), claims.TenantID, campaignID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(steps)
}

func (h *Handler) CreateStep(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "id")

	var req CreateStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate action_type
	validActions := map[string]bool{
		"email":     true,
		"sms":       true,
		"follow_up": true,
	}
	if !validActions[req.ActionType] {
		http.Error(w, "Invalid action_type. Must be one of: email, sms, follow_up", http.StatusBadRequest)
		return
	}

	step, err := h.service.CreateStep(r.Context(), claims.TenantID, campaignID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(step)
}

func (h *Handler) UpdateStep(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "campaignId")
	stepID := chi.URLParam(r, "stepId")

	var req UpdateStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	step, err := h.service.UpdateStep(r.Context(), claims.TenantID, campaignID, stepID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(step)
}

func (h *Handler) DeleteStep(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "campaignId")
	stepID := chi.URLParam(r, "stepId")

	err := h.service.DeleteStep(r.Context(), claims.TenantID, campaignID, stepID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ===== ENROLLMENTS =====

func (h *Handler) EnrollPerson(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "id")
	personID := chi.URLParam(r, "personId")

	enrollment, err := h.service.EnrollPerson(r.Context(), claims.TenantID, campaignID, personID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(enrollment)
}

func (h *Handler) ListEnrollments(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "id")

	enrollments, err := h.service.ListEnrollments(r.Context(), claims.TenantID, campaignID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enrollments)
}

// ===== PROCESSING =====

func (h *Handler) ProcessPendingSteps(w http.ResponseWriter, r *http.Request) {
	// This endpoint can be called by a cron job to process pending steps
	err := h.service.ProcessPendingSteps(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pending steps processed"))
}
