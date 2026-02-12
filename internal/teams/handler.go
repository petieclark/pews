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

type updateMemberStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) UpdateMemberStatus(w http.ResponseWriter, r *http.Request) {
	memberID := chi.URLParam(r, "memberId")
	var req updateMemberStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.UpdateMemberStatus(r.Context(), memberID, req.Status); err != nil {
		http.Error(w, "Failed to update member status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type updatePositionRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

func (h *Handler) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	positionID := chi.URLParam(r, "positionId")
	var req updatePositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	pos, err := h.service.UpdatePosition(r.Context(), id, positionID, req.Name, req.SortOrder)
	if err != nil {
		http.Error(w, "Failed to update position", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pos)
}

// ---- Service Team Assignments ----

func (h *Handler) GetServiceAssignments(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	serviceID := chi.URLParam(r, "id")
	assignments, err := h.service.GetServiceAssignments(r.Context(), claims.TenantID, serviceID)
	if err != nil {
		http.Error(w, "Failed to get assignments: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if assignments == nil {
		assignments = []ServiceTeamAssignment{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"assignments": assignments})
}

type saveAssignmentsRequest struct {
	Assignments []struct {
		TeamID     string  `json:"team_id"`
		PositionID *string `json:"position_id,omitempty"`
		PersonID   string  `json:"person_id"`
		Status     string  `json:"status"`
		Notes      string  `json:"notes"`
	} `json:"assignments"`
}

func (h *Handler) SaveServiceAssignments(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	serviceID := chi.URLParam(r, "id")

	var req saveAssignmentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var assignments []ServiceTeamAssignment
	for _, a := range req.Assignments {
		status := a.Status
		if status == "" {
			status = "pending"
		}
		assignments = append(assignments, ServiceTeamAssignment{
			TeamID:     a.TeamID,
			PositionID: a.PositionID,
			PersonID:   a.PersonID,
			Status:     status,
			Notes:      a.Notes,
		})
	}

	if err := h.service.SaveServiceAssignments(r.Context(), claims.TenantID, serviceID, assignments); err != nil {
		http.Error(w, "Failed to save assignments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated assignments
	h.GetServiceAssignments(w, r)
}

func (h *Handler) CopyServiceAssignments(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	targetID := chi.URLParam(r, "id")
	sourceID := chi.URLParam(r, "sourceId")

	assignments, err := h.service.CopyServiceAssignments(r.Context(), claims.TenantID, targetID, sourceID)
	if err != nil {
		http.Error(w, "Failed to copy assignments: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if assignments == nil {
		assignments = []ServiceTeamAssignment{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"assignments": assignments})
}

func (h *Handler) UpdateAssignmentStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	assignmentID := chi.URLParam(r, "assignmentId")

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateAssignmentStatus(r.Context(), claims.TenantID, assignmentID, req.Status); err != nil {
		http.Error(w, "Failed to update assignment status", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetPersonSchedule(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	personID := chi.URLParam(r, "personId")

	assignments, err := h.service.GetPersonSchedule(r.Context(), claims.TenantID, personID)
	if err != nil {
		http.Error(w, "Failed to get schedule: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if assignments == nil {
		assignments = []ServiceTeamAssignment{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"assignments": assignments})
}
