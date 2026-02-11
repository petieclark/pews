package prayer

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreatePrayerRequestPublic(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Tenant ID required", http.StatusBadRequest)
		return
	}

	var input CreatePrayerRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.RequestText == "" {
		http.Error(w, "Name and request text are required", http.StatusBadRequest)
		return
	}

	pr, err := h.service.CreatePrayerRequest(r.Context(), tenantID, input, nil)
	if err != nil {
		http.Error(w, "Failed to create prayer request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pr)
}

func (h *Handler) ListPublicPrayerRequests(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		http.Error(w, "Tenant ID required", http.StatusBadRequest)
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	requests, err := h.service.ListPublicPrayerRequests(r.Context(), tenantID, limit)
	if err != nil {
		http.Error(w, "Failed to list public prayer requests: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

func (h *Handler) CreatePrayerRequestAuth(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input CreatePrayerRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Name == "" || input.RequestText == "" {
		http.Error(w, "Name and request text are required", http.StatusBadRequest)
		return
	}

	var personID *string
	if pid := r.URL.Query().Get("person_id"); pid != "" {
		personID = &pid
	}

	pr, err := h.service.CreatePrayerRequest(r.Context(), claims.TenantID, input, personID)
	if err != nil {
		http.Error(w, "Failed to create prayer request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pr)
}

func (h *Handler) ListPrayerRequests(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var filter PrayerRequestFilter
	if err := decodeQueryParams(r, &filter); err != nil {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	if filter.Limit <= 0 {
		filter.Limit = 100
	}

	requests, err := h.service.ListPrayerRequests(r.Context(), claims.TenantID, filter, &claims.UserID)
	if err != nil {
		http.Error(w, "Failed to list prayer requests: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

func (h *Handler) GetPrayerRequest(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requestID := chi.URLParam(r, "id")
	if requestID == "" {
		http.Error(w, "Request ID required", http.StatusBadRequest)
		return
	}

	pr, err := h.service.GetPrayerRequest(r.Context(), claims.TenantID, requestID, &claims.UserID)
	if err != nil {
		http.Error(w, "Failed to get prayer request: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pr)
}

func (h *Handler) UpdatePrayerRequest(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requestID := chi.URLParam(r, "id")
	if requestID == "" {
		http.Error(w, "Request ID required", http.StatusBadRequest)
		return
	}

	var input UpdatePrayerRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Status != "" {
		validStatuses := map[string]bool{"pending": true, "praying": true, "answered": true, "archived": true}
		if !validStatuses[input.Status] {
			http.Error(w, "Invalid status. Must be one of: pending, praying, answered, archived", http.StatusBadRequest)
			return
		}
	}

	pr, err := h.service.UpdatePrayerRequest(r.Context(), claims.TenantID, requestID, input)
	if err != nil {
		http.Error(w, "Failed to update prayer request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pr)
}

func (h *Handler) FollowPrayerRequest(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requestID := chi.URLParam(r, "id")
	if requestID == "" {
		http.Error(w, "Request ID required", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodDelete || r.URL.Query().Get("unfollow") == "true" {
		err := h.service.UnfollowPrayerRequest(r.Context(), claims.TenantID, requestID, claims.UserID)
		if err != nil {
			http.Error(w, "Failed to unfollow prayer request: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		err := h.service.FollowPrayerRequest(r.Context(), claims.TenantID, requestID, claims.UserID)
		if err != nil {
			http.Error(w, "Failed to follow prayer request: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListFollowers(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	requestID := chi.URLParam(r, "id")
	if requestID == "" {
		http.Error(w, "Request ID required", http.StatusBadRequest)
		return
	}

	followers, err := h.service.ListFollowers(r.Context(), claims.TenantID, requestID)
	if err != nil {
		http.Error(w, "Failed to list followers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followers)
}

func (h *Handler) ImportFromConnectionCard(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	connectionCardID := chi.URLParam(r, "connectionCardId")
	if connectionCardID == "" {
		http.Error(w, "Connection card ID required", http.StatusBadRequest)
		return
	}

	pr, err := h.service.ImportFromConnectionCard(r.Context(), claims.TenantID, connectionCardID)
	if err != nil {
		http.Error(w, "Failed to import prayer request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pr)
}

func decodeQueryParams(r *http.Request, filter *PrayerRequestFilter) error {
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}

	if isPublicStr := r.URL.Query().Get("is_public"); isPublicStr != "" {
		isPublic := isPublicStr == "true"
		filter.IsPublic = &isPublic
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = o
		}
	}

	return nil
}
