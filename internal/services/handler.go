package services

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

	serviceDate, err := time.Parse("2006-01-02", req.ServiceDate)
	if err != nil {
		http.Error(w, "Invalid service_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	svc := &ChurchService{
		ServiceTypeID: req.ServiceTypeID,
		Name:          req.Name,
		ServiceDate:   serviceDate,
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

	serviceDate, err := time.Parse("2006-01-02", req.ServiceDate)
	if err != nil {
		http.Error(w, "Invalid service_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	svc := &ChurchService{
		ServiceTypeID: req.ServiceTypeID,
		Name:          req.Name,
		ServiceDate:   serviceDate,
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
	Title         string `json:"title"`
	Artist        string `json:"artist,omitempty"`
	DefaultKey    string `json:"default_key,omitempty"`
	Tempo         int    `json:"tempo,omitempty"`
	CCLINumber    string `json:"ccli_number,omitempty"`
	Lyrics        string `json:"lyrics,omitempty"`
	Notes         string `json:"notes,omitempty"`
	Tags          string `json:"tags,omitempty"`
	YoutubeURL    string `json:"youtube_url,omitempty"`
	SpotifyURL    string `json:"spotify_url,omitempty"`
	AppleMusicURL string `json:"apple_music_url,omitempty"`
	RehearsalURL  string `json:"rehearsal_url,omitempty"`
}

func (h *Handler) ListSongs(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	params := SongListParams{
		Query:     r.URL.Query().Get("q"),
		Key:       r.URL.Query().Get("key"),
		Tag:       r.URL.Query().Get("tag"),
		HasLyrics: r.URL.Query().Get("has_lyrics"),
		Sort:      r.URL.Query().Get("sort"),
	}
	params.Page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	params.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))

	songs, total, err := h.service.ListSongsFiltered(r.Context(), claims.TenantID, params)
	if err != nil {
		http.Error(w, "Failed to list songs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"songs": songs,
		"total": total,
		"page":  params.Page,
		"limit": params.Limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetSongStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := h.service.GetSongStats(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get song stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
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
		Title:         req.Title,
		Artist:        req.Artist,
		DefaultKey:    req.DefaultKey,
		Tempo:         req.Tempo,
		CCLINumber:    req.CCLINumber,
		Lyrics:        req.Lyrics,
		Notes:         req.Notes,
		Tags:          req.Tags,
		YoutubeURL:    req.YoutubeURL,
		SpotifyURL:    req.SpotifyURL,
		AppleMusicURL: req.AppleMusicURL,
		RehearsalURL:  req.RehearsalURL,
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
		Title:         req.Title,
		Artist:        req.Artist,
		DefaultKey:    req.DefaultKey,
		Tempo:         req.Tempo,
		CCLINumber:    req.CCLINumber,
		Lyrics:        req.Lyrics,
		Notes:         req.Notes,
		Tags:          req.Tags,
		YoutubeURL:    req.YoutubeURL,
		SpotifyURL:    req.SpotifyURL,
		AppleMusicURL: req.AppleMusicURL,
		RehearsalURL:  req.RehearsalURL,
	}

	updatedSong, err := h.service.UpdateSong(r.Context(), claims.TenantID, songID, song)
	if err != nil {
		http.Error(w, "Failed to update song: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSong)
}

func (h *Handler) GetSong(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	songID := chi.URLParam(r, "id")
	song, err := h.service.GetSongByID(r.Context(), claims.TenantID, songID)
	if err != nil {
		http.Error(w, "Song not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(song)
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

// Template handlers

type CreateTemplateRequest struct {
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	TemplateData json.RawMessage `json:"template_data,omitempty"`
}

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templates, err := h.service.ListTemplates(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to list templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templateID := chi.URLParam(r, "id")
	t, err := h.service.GetTemplate(r.Context(), claims.TenantID, templateID)
	if err != nil {
		http.Error(w, "Template not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	t := &ServiceTemplate{
		Name:         req.Name,
		Description:  req.Description,
		TemplateData: req.TemplateData,
	}

	created, err := h.service.CreateTemplate(r.Context(), claims.TenantID, t)
	if err != nil {
		http.Error(w, "Failed to create template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templateID := chi.URLParam(r, "id")

	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	t := &ServiceTemplate{
		Name:         req.Name,
		Description:  req.Description,
		TemplateData: req.TemplateData,
	}

	updated, err := h.service.UpdateTemplate(r.Context(), claims.TenantID, templateID, t)
	if err != nil {
		http.Error(w, "Failed to update template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templateID := chi.URLParam(r, "id")
	if err := h.service.DeleteTemplate(r.Context(), claims.TenantID, templateID); err != nil {
		http.Error(w, "Failed to delete template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type SaveAsTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func (h *Handler) SaveAsTemplate(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceID := chi.URLParam(r, "id")

	var req SaveAsTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	t, err := h.service.SaveAsTemplate(r.Context(), claims.TenantID, serviceID, req.Name, req.Description)
	if err != nil {
		http.Error(w, "Failed to save as template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

type CopyServiceRequest struct {
	ServiceDate string `json:"service_date"`
}

func (h *Handler) CopyService(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceID := chi.URLParam(r, "id")

	var req CopyServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newDate, err := time.Parse("2006-01-02", req.ServiceDate)
	if err != nil {
		http.Error(w, "Invalid service_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	copied, err := h.service.CopyService(r.Context(), claims.TenantID, serviceID, newDate)
	if err != nil {
		http.Error(w, "Failed to copy service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(copied)
}

type ReorderItemsRequest struct {
	ItemIDs []string `json:"item_ids"`
}

func (h *Handler) ReorderItems(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	serviceID := chi.URLParam(r, "id")

	var req ReorderItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.ReorderItems(r.Context(), claims.TenantID, serviceID, req.ItemIDs); err != nil {
		http.Error(w, "Failed to reorder items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteServiceType(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	typeID := chi.URLParam(r, "id")
	if err := h.service.DeleteServiceType(r.Context(), claims.TenantID, typeID); err != nil {
		http.Error(w, "Failed to delete service type: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetSongUsage(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	songID := chi.URLParam(r, "id")
	usage, err := h.service.GetSongUsage(r.Context(), claims.TenantID, songID)
	if err != nil {
		http.Error(w, "Failed to get song usage: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}
// Volunteer Teams handlers

type CreateVolunteerTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	IsActive    bool   `json:"is_active"`
}

func (h *Handler) ListVolunteerTeams(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teams, err := h.service.ListVolunteerTeams(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to list volunteer teams: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

func (h *Handler) GetVolunteerTeam(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID := chi.URLParam(r, "id")

	team, err := h.service.GetVolunteerTeamByID(r.Context(), claims.TenantID, teamID)
	if err != nil {
		http.Error(w, "Failed to get volunteer team: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func (h *Handler) CreateVolunteerTeam(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateVolunteerTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	team := &VolunteerTeam{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		IsActive:    req.IsActive,
	}

	if team.Color == "" {
		team.Color = "#4A8B8C"
	}

	createdTeam, err := h.service.CreateVolunteerTeam(r.Context(), claims.TenantID, team)
	if err != nil {
		http.Error(w, "Failed to create volunteer team: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTeam)
}

func (h *Handler) UpdateVolunteerTeam(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID := chi.URLParam(r, "id")

	var req CreateVolunteerTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	team := &VolunteerTeam{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		IsActive:    req.IsActive,
	}

	updatedTeam, err := h.service.UpdateVolunteerTeam(r.Context(), claims.TenantID, teamID, team)
	if err != nil {
		http.Error(w, "Failed to update volunteer team: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTeam)
}

func (h *Handler) DeleteVolunteerTeam(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID := chi.URLParam(r, "id")

	if err := h.service.DeleteVolunteerTeam(r.Context(), claims.TenantID, teamID); err != nil {
		http.Error(w, "Failed to delete volunteer team: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Team Members handlers

type AddTeamMemberRequest struct {
	PersonID string `json:"person_id"`
	Role     string `json:"role,omitempty"`
	IsActive bool   `json:"is_active"`
}

func (h *Handler) GetTeamMembers(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID := chi.URLParam(r, "id")

	members, err := h.service.GetTeamMembers(r.Context(), claims.TenantID, teamID)
	if err != nil {
		http.Error(w, "Failed to get team members: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

func (h *Handler) GetPersonTeams(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	personID := chi.URLParam(r, "id")

	teams, err := h.service.GetPersonTeams(r.Context(), claims.TenantID, personID)
	if err != nil {
		http.Error(w, "Failed to get person teams: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

func (h *Handler) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID := chi.URLParam(r, "id")

	var req AddTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	member := &TeamMember{
		TeamID:   teamID,
		PersonID: req.PersonID,
		Role:     req.Role,
		IsActive: req.IsActive,
	}

	addedMember, err := h.service.AddTeamMember(r.Context(), claims.TenantID, member)
	if err != nil {
		http.Error(w, "Failed to add team member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(addedMember)
}

func (h *Handler) UpdateTeamMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	memberID := chi.URLParam(r, "id")

	var req AddTeamMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	member := &TeamMember{
		Role:     req.Role,
		IsActive: req.IsActive,
	}

	updatedMember, err := h.service.UpdateTeamMember(r.Context(), claims.TenantID, memberID, member)
	if err != nil {
		http.Error(w, "Failed to update team member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedMember)
}

func (h *Handler) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	memberID := chi.URLParam(r, "id")

	if err := h.service.RemoveTeamMember(r.Context(), claims.TenantID, memberID); err != nil {
		http.Error(w, "Failed to remove team member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Volunteer Availability handlers

type AddAvailabilityRequest struct {
	PersonID  string `json:"person_id"`
	TeamID    string `json:"team_id,omitempty"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Reason    string `json:"reason,omitempty"`
}

func (h *Handler) GetPersonAvailability(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	personID := chi.URLParam(r, "id")

	availability, err := h.service.GetPersonAvailability(r.Context(), claims.TenantID, personID)
	if err != nil {
		http.Error(w, "Failed to get person availability: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availability)
}

func (h *Handler) AddAvailability(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req AddAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		http.Error(w, "Invalid end date format", http.StatusBadRequest)
		return
	}

	avail := &VolunteerAvailability{
		PersonID:  req.PersonID,
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    req.Reason,
	}

	if req.TeamID != "" {
		avail.TeamID = &req.TeamID
	}

	addedAvail, err := h.service.AddAvailability(r.Context(), claims.TenantID, avail)
	if err != nil {
		http.Error(w, "Failed to add availability: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(addedAvail)
}

func (h *Handler) UpdateAvailability(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	availID := chi.URLParam(r, "id")

	var req AddAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		http.Error(w, "Invalid end date format", http.StatusBadRequest)
		return
	}

	avail := &VolunteerAvailability{
		StartDate: startDate,
		EndDate:   endDate,
		Reason:    req.Reason,
	}

	if req.TeamID != "" {
		avail.TeamID = &req.TeamID
	}

	updatedAvail, err := h.service.UpdateAvailability(r.Context(), claims.TenantID, availID, avail)
	if err != nil {
		http.Error(w, "Failed to update availability: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedAvail)
}

func (h *Handler) DeleteAvailability(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	availID := chi.URLParam(r, "id")

	if err := h.service.DeleteAvailability(r.Context(), claims.TenantID, availID); err != nil {
		http.Error(w, "Failed to delete availability: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Scheduling helpers

func (h *Handler) GetSchedulingConflicts(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "date query parameter required", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	conflicts, err := h.service.GetSchedulingConflicts(r.Context(), claims.TenantID, date)
	if err != nil {
		http.Error(w, "Failed to get scheduling conflicts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conflicts)
}

func (h *Handler) GetAvailableVolunteers(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamID := chi.URLParam(r, "id")
	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		http.Error(w, "date query parameter required", http.StatusBadRequest)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	volunteers, err := h.service.GetAvailableVolunteers(r.Context(), claims.TenantID, teamID, date)
	if err != nil {
		http.Error(w, "Failed to get available volunteers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(volunteers)
}

type UpdateServiceTeamStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) UpdateServiceTeamStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	teamMemberID := chi.URLParam(r, "id")

	var req UpdateServiceTeamStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateServiceTeamStatus(r.Context(), claims.TenantID, teamMemberID, req.Status); err != nil {
		http.Error(w, "Failed to update service team status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
