package groups

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
	return &Handler{service: service}
}

// Group handlers

type CreateGroupRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	GroupType       string `json:"group_type"`
	MeetingDay      string `json:"meeting_day,omitempty"`
	MeetingTime     string `json:"meeting_time,omitempty"`
	MeetingLocation string `json:"meeting_location,omitempty"`
	IsPublic        bool   `json:"is_public"`
	MaxMembers      *int   `json:"max_members,omitempty"`
	IsActive        bool   `json:"is_active"`
	PhotoURL        string `json:"photo_url,omitempty"`
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupType := r.URL.Query().Get("type")
	activeStr := r.URL.Query().Get("active")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	var active *bool
	if activeStr != "" {
		val := activeStr == "true"
		active = &val
	}

	groups, total, err := h.service.ListGroups(r.Context(), claims.TenantID, groupType, active, page, limit)
	if err != nil {
		http.Error(w, "Failed to list groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"groups": groups,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := chi.URLParam(r, "id")
	group, err := h.service.GetGroupByID(r.Context(), claims.TenantID, groupID)
	if err != nil {
		http.Error(w, "Group not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	group := &Group{
		Name:            req.Name,
		Description:     req.Description,
		GroupType:       req.GroupType,
		MeetingDay:      req.MeetingDay,
		MeetingTime:     req.MeetingTime,
		MeetingLocation: req.MeetingLocation,
		IsPublic:        req.IsPublic,
		MaxMembers:      req.MaxMembers,
		IsActive:        req.IsActive,
		PhotoURL:        req.PhotoURL,
	}

	if group.GroupType == "" {
		group.GroupType = "small_group"
	}

	createdGroup, err := h.service.CreateGroup(r.Context(), claims.TenantID, group)
	if err != nil {
		http.Error(w, "Failed to create group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdGroup)
}

func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := chi.URLParam(r, "id")

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	group := &Group{
		Name:            req.Name,
		Description:     req.Description,
		GroupType:       req.GroupType,
		MeetingDay:      req.MeetingDay,
		MeetingTime:     req.MeetingTime,
		MeetingLocation: req.MeetingLocation,
		IsPublic:        req.IsPublic,
		MaxMembers:      req.MaxMembers,
		IsActive:        req.IsActive,
		PhotoURL:        req.PhotoURL,
	}

	updatedGroup, err := h.service.UpdateGroup(r.Context(), claims.TenantID, groupID, group)
	if err != nil {
		http.Error(w, "Failed to update group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedGroup)
}

func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := chi.URLParam(r, "id")

	if err := h.service.DeleteGroup(r.Context(), claims.TenantID, groupID); err != nil {
		http.Error(w, "Failed to delete group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Member handlers

func (h *Handler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := chi.URLParam(r, "id")
	members, err := h.service.GetGroupMembers(r.Context(), claims.TenantID, groupID)
	if err != nil {
		http.Error(w, "Failed to get group members: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

type AddMemberRequest struct {
	PersonID string `json:"person_id"`
	Role     string `json:"role"`
}

func (h *Handler) AddMemberToGroup(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := chi.URLParam(r, "id")

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	member, err := h.service.AddMemberToGroup(r.Context(), claims.TenantID, groupID, req.PersonID, req.Role)
	if err != nil {
		http.Error(w, "Failed to add member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(member)
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role"`
}

func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	memberID := chi.URLParam(r, "memberId")

	var req UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateMemberRole(r.Context(), claims.TenantID, memberID, req.Role); err != nil {
		http.Error(w, "Failed to update member role: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveMemberFromGroup(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	memberID := chi.URLParam(r, "memberId")

	if err := h.service.RemoveMemberFromGroup(r.Context(), claims.TenantID, memberID); err != nil {
		http.Error(w, "Failed to remove member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetPersonGroups(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	personID := chi.URLParam(r, "personId")
	groups, err := h.service.GetPersonGroups(r.Context(), claims.TenantID, personID)
	if err != nil {
		http.Error(w, "Failed to get person groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}
