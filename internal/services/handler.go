package services

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

// ServiceType handlers

type CreateServiceTypeRequest struct {
	Name        string `json:"name"`
	DefaultTime string `json:"default_time,omitempty"`
	DefaultDay  string `json:"default_day,omitempty"`
	Color       string `json:"color,omitempty"`
	IsActive    bool   `json:"is_active"`
}

func (h *Handler) ListServiceTypes(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	types, err := h.service.ListServiceTypes(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to list service types: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types)
}

func (h *Handler) CreateServiceType(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateServiceTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	st := &ServiceType{
		Name:        req.Name,
		DefaultTime: req.DefaultTime,
		DefaultDay:  req.DefaultDay,
		Color:       req.Color,
		IsActive:    req.IsActive,
	}

	if st.Color == "" {
		st.Color = "#4A8B8C"
	}

	createdType, err := h.service.CreateServiceType(r.Context(), claims.TenantID, st)
	if err != nil {
		http.Error(w, "Failed to create service type: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdType)
}

func (h *Handler) UpdateServiceType(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	typeID := chi.URLParam(r, "id")

	var req CreateServiceTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	st := &ServiceType{
		Name:        req.Name,
		DefaultTime: req.DefaultTime,
		DefaultDay:  req.DefaultDay,
		Color:       req.Color,
		IsActive:    req.IsActive,
	}

	updatedType, err := h.service.UpdateServiceType(r.Context(), claims.TenantID, typeID, st)
	if err != nil {
		http.Error(w, "Failed to update service type: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedType)
}

// Services handlers

type CreateServiceRequest struct {
	ServiceTypeID string `json:"service_type_id"`
	Name          string `json:"name,omitempty"`
	ServiceDate   string `json:"service_date"`
	ServiceTime   string `json:"service_time,omitempty"`
	Notes         string `json:"notes,omitempty"`
	Status        string `json:"status,omitempty"`
}

func (h *Handler) ListServices(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fromDate := r.URL.Query().Get("from")
	toDate := r.URL.Query().Get("to")
	typeID := r.URL.Query().Get("type_id")
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	services, total, err := h.service.ListServices(r.Context(), claims.TenantID, fromDate, toDate, typeID, status, page, limit)
	if err != nil {
		http.Error(w, "Failed to list services: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"services": services,
		"total":    total,
		"page":     page,
		"limit":    limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetUpcomingServices(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 4
	}

	services, err := h.service.GetUpcomingServices(r.Context(), claims.TenantID, limit)
	if err != nil {
		http.Error(w, "Failed to get upcoming services: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func (h *Handler) GetService(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceID := chi.URLParam(r, "id")
	service, err := h.service.GetServiceByID(r.Context(), claims.TenantID, serviceID)
	if err != nil {
		http.Error(w, "Service not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(service)
}

func (h *Handler) CreateService(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	svc := &Service{
		ServiceTypeID: req.ServiceTypeID,
		Name:          req.Name,
		ServiceDate:   req.ServiceDate,
		ServiceTime:   req.ServiceTime,
		Notes:         req.Notes,
		Status:        req.Status,
	}

	if svc.Status == "" {
		svc.Status = "planning"
	}

	createdService, err := h.service.CreateService(r.Context(), claims.TenantID, svc)
	if err != nil {
		http.Error(w, "Failed to create service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdService)
}

func (h *Handler) UpdateService(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceID := chi.URLParam(r, "id")

	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	svc := &Service{
		ServiceTypeID: req.ServiceTypeID,
		Name:          req.Name,
		ServiceDate:   req.ServiceDate,
		ServiceTime:   req.ServiceTime,
		Notes:         req.Notes,
		Status:        req.Status,
	}

	updatedService, err := h.service.UpdateService(r.Context(), claims.TenantID, serviceID, svc)
	if err != nil {
		http.Error(w, "Failed to update service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedService)
}

func (h *Handler) DeleteService(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceID := chi.URLParam(r, "id")

	if err := h.service.DeleteService(r.Context(), claims.TenantID, serviceID); err != nil {
		http.Error(w, "Failed to delete service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Service Items handlers

type CreateServiceItemRequest struct {
	ItemType        string  `json:"item_type"`
	Title           string  `json:"title"`
	SongID          *string `json:"song_id,omitempty"`
	SongKey         string  `json:"song_key,omitempty"`
	Position        int     `json:"position"`
	DurationMinutes *int    `json:"duration_minutes,omitempty"`
	Notes           string  `json:"notes,omitempty"`
	AssignedTo      string  `json:"assigned_to,omitempty"`
}

func (h *Handler) GetServiceItems(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceID := chi.URLParam(r, "id")
	items, err := h.service.GetServiceItems(r.Context(), claims.TenantID, serviceID)
	if err != nil {
		http.Error(w, "Failed to get service items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *Handler) AddServiceItem(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceID := chi.URLParam(r, "id")

	var req CreateServiceItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	item := &ServiceItem{
		ServiceID:       serviceID,
		ItemType:        req.ItemType,
		Title:           req.Title,
		SongID:          req.SongID,
		SongKey:         req.SongKey,
		Position:        req.Position,
		DurationMinutes: req.DurationMinutes,
		Notes:           req.Notes,
		AssignedTo:      req.AssignedTo,
	}

	createdItem, err := h.service.AddServiceItem(r.Context(), claims.TenantID, item)
	if err != nil {
		http.Error(w, "Failed to add service item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdItem)
}

func (h *Handler) UpdateServiceItem(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	itemID := chi.URLParam(r, "itemId")

	var req CreateServiceItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	item := &ServiceItem{
		ItemType:        req.ItemType,
		Title:           req.Title,
		SongID:          req.SongID,
		SongKey:         req.SongKey,
		Position:        req.Position,
		DurationMinutes: req.DurationMinutes,
		Notes:           req.Notes,
		AssignedTo:      req.AssignedTo,
	}

	updatedItem, err := h.service.UpdateServiceItem(r.Context(), claims.TenantID, itemID, item)
	if err != nil {
		http.Error(w, "Failed to update service item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedItem)
}

func (h *Handler) DeleteServiceItem(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	itemID := chi.URLParam(r, "itemId")

	if err := h.service.DeleteServiceItem(r.Context(), claims.TenantID, itemID); err != nil {
		http.Error(w, "Failed to delete service item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Service Team handlers

type CreateServiceTeamRequest struct {
	PersonID string `json:"person_id"`
	Role     string `json:"role"`
	Status   string `json:"status,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

func (h *Handler) GetServiceTeam(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceID := chi.URLParam(r, "id")
	team, err := h.service.GetServiceTeam(r.Context(), claims.TenantID, serviceID)
	if err != nil {
		http.Error(w, "Failed to get service team: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func (h *Handler) AddServiceTeamMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceID := chi.URLParam(r, "id")

	var req CreateServiceTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	member := &ServiceTeam{
		ServiceID: serviceID,
		PersonID:  req.PersonID,
		Role:      req.Role,
		Status:    req.Status,
		Notes:     req.Notes,
	}

	if member.Status == "" {
		member.Status = "pending"
	}

	createdMember, err := h.service.AddServiceTeamMember(r.Context(), claims.TenantID, member)
	if err != nil {
		http.Error(w, "Failed to add service team member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdMember)
}

func (h *Handler) UpdateServiceTeamMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID := chi.URLParam(r, "teamId")

	var req CreateServiceTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	member := &ServiceTeam{
		Role:   req.Role,
		Status: req.Status,
		Notes:  req.Notes,
	}

	updatedMember, err := h.service.UpdateServiceTeamMember(r.Context(), claims.TenantID, teamID, member)
	if err != nil {
		http.Error(w, "Failed to update service team member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedMember)
}

func (h *Handler) DeleteServiceTeamMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID := chi.URLParam(r, "teamId")

	if err := h.service.DeleteServiceTeamMember(r.Context(), claims.TenantID, teamID); err != nil {
		http.Error(w, "Failed to delete service team member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Songs handlers

type CreateSongRequest struct {
	Title      string `json:"title"`
	Artist     string `json:"artist,omitempty"`
	DefaultKey string `json:"default_key,omitempty"`
	Tempo      int    `json:"tempo,omitempty"`
	CCLINumber string `json:"ccli_number,omitempty"`
	Lyrics     string `json:"lyrics,omitempty"`
	Notes      string `json:"notes,omitempty"`
	Tags       string `json:"tags,omitempty"`
}

func (h *Handler) ListSongs(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	songs, total, err := h.service.ListSongs(r.Context(), claims.TenantID, query, page, limit)
	if err != nil {
		http.Error(w, "Failed to list songs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"songs": songs,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) CreateSong(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	song := &Song{
		Title:      req.Title,
		Artist:     req.Artist,
		DefaultKey: req.DefaultKey,
		Tempo:      req.Tempo,
		CCLINumber: req.CCLINumber,
		Lyrics:     req.Lyrics,
		Notes:      req.Notes,
		Tags:       req.Tags,
	}

	createdSong, err := h.service.CreateSong(r.Context(), claims.TenantID, song)
	if err != nil {
		http.Error(w, "Failed to create song: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdSong)
}

func (h *Handler) UpdateSong(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	songID := chi.URLParam(r, "id")

	var req CreateSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	song := &Song{
		Title:      req.Title,
		Artist:     req.Artist,
		DefaultKey: req.DefaultKey,
		Tempo:      req.Tempo,
		CCLINumber: req.CCLINumber,
		Lyrics:     req.Lyrics,
		Notes:      req.Notes,
		Tags:       req.Tags,
	}

	updatedSong, err := h.service.UpdateSong(r.Context(), claims.TenantID, songID, song)
	if err != nil {
		http.Error(w, "Failed to update song: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSong)
}

func (h *Handler) DeleteSong(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	songID := chi.URLParam(r, "id")

	if err := h.service.DeleteSong(r.Context(), claims.TenantID, songID); err != nil {
		http.Error(w, "Failed to delete song: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
