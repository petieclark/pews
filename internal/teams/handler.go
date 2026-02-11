package teams

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

func (h *Handler) ListTeams(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teams, err := h.service.ListTeams(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to list teams: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if teams == nil {
		teams = []Team{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"teams": teams})
}

type createTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req createTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	team, err := h.service.CreateTeam(r.Context(), claims.TenantID, req.Name, req.Description, req.Color)
	if err != nil {
		http.Error(w, "Failed to create team: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	team, err := h.service.GetTeam(r.Context(), claims.TenantID, id)
	if err != nil {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

type updateTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	IsActive    bool   `json:"is_active"`
}

func (h *Handler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	var req updateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	team, err := h.service.UpdateTeam(r.Context(), claims.TenantID, id, req.Name, req.Description, req.Color, req.IsActive)
	if err != nil {
		http.Error(w, "Failed to update team: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func (h *Handler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.service.DeleteTeam(r.Context(), claims.TenantID, id); err != nil {
		http.Error(w, "Failed to delete team: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type addPositionRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

func (h *Handler) AddPosition(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req addPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	pos, err := h.service.AddPosition(r.Context(), id, req.Name, req.SortOrder)
	if err != nil {
		http.Error(w, "Failed to add position: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pos)
}

func (h *Handler) DeletePosition(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	positionID := chi.URLParam(r, "positionId")

	if err := h.service.DeletePosition(r.Context(), id, positionID); err != nil {
		http.Error(w, "Failed to delete position: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type addMemberRequest struct {
	PersonID   string  `json:"person_id"`
	PositionID *string `json:"position_id,omitempty"`
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.PersonID == "" {
		http.Error(w, "person_id is required", http.StatusBadRequest)
		return
	}

	member, err := h.service.AddMember(r.Context(), id, req.PersonID, req.PositionID)
	if err != nil {
		http.Error(w, "Failed to add member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(member)
}

type updateMemberRequest struct {
	PositionID *string `json:"position_id"`
}

func (h *Handler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "memberId")
	var req updateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateMember(r.Context(), memberID, req.PositionID); err != nil {
		http.Error(w, "Failed to update member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	memberID := chi.URLParam(r, "memberId")

	if err := h.service.DeleteMember(r.Context(), id, memberID); err != nil {
		http.Error(w, "Failed to remove member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
