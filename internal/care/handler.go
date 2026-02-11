package care

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

func (h *Handler) ListFollowUps(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	status := r.URL.Query().Get("status")
	typ := r.URL.Query().Get("type")
	priority := r.URL.Query().Get("priority")

	items, err := h.service.ListFollowUps(r.Context(), claims.TenantID, status, typ, priority)
	if err != nil {
		http.Error(w, "Failed to list follow-ups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"follow_ups": items})
}

func (h *Handler) GetFollowUp(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	item, err := h.service.GetFollowUp(r.Context(), claims.TenantID, id)
	if err != nil {
		http.Error(w, "Not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (h *Handler) CreateFollowUp(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input CreateFollowUpInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.PersonID == "" || input.Title == "" {
		http.Error(w, "person_id and title are required", http.StatusBadRequest)
		return
	}

	item, err := h.service.CreateFollowUp(r.Context(), claims.TenantID, claims.UserID, input)
	if err != nil {
		http.Error(w, "Failed to create: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h *Handler) UpdateFollowUp(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	var input UpdateFollowUpInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	item, err := h.service.UpdateFollowUp(r.Context(), claims.TenantID, id, input)
	if err != nil {
		http.Error(w, "Failed to update: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (h *Handler) DeleteFollowUp(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.service.DeleteFollowUp(r.Context(), claims.TenantID, id); err != nil {
		http.Error(w, "Failed to delete: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AddNote(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")

	var body struct {
		Note string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Note == "" {
		http.Error(w, "note is required", http.StatusBadRequest)
		return
	}

	note, err := h.service.AddNote(r.Context(), claims.TenantID, id, claims.UserID, body.Note)
	if err != nil {
		http.Error(w, "Failed to add note: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

func (h *Handler) ListNotes(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	notes, err := h.service.ListNotes(r.Context(), claims.TenantID, id)
	if err != nil {
		http.Error(w, "Failed to list notes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := h.service.GetStats(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
