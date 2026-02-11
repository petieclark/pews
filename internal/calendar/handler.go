package calendar

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

type CreateEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
	StartTime   string `json:"start_time"` // ISO 8601 format
	EndTime     string `json:"end_time"`   // ISO 8601 format
	AllDay      bool   `json:"all_day"`
	Recurring   string `json:"recurring"` // none, weekly, monthly
	Color       string `json:"color,omitempty"`
}

// ListEvents returns all events, optionally filtered by date range
func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fromDate := r.URL.Query().Get("from")
	toDate := r.URL.Query().Get("to")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	events, total, err := h.service.ListEvents(r.Context(), claims.TenantID, fromDate, toDate, page, limit)
	if err != nil {
		http.Error(w, "Failed to list events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"events": events,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetEvent returns a single event by ID
func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	eventID := chi.URLParam(r, "id")
	event, err := h.service.GetEventByID(r.Context(), claims.TenantID, eventID)
	if err != nil {
		http.Error(w, "Event not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

// CreateEvent creates a new event
func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse times
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		http.Error(w, "Invalid start_time format (expected RFC3339/ISO 8601)", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		http.Error(w, "Invalid end_time format (expected RFC3339/ISO 8601)", http.StatusBadRequest)
		return
	}

	// Validate recurring value
	if req.Recurring != "none" && req.Recurring != "weekly" && req.Recurring != "monthly" {
		req.Recurring = "none"
	}

	event := &Event{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartTime:   startTime,
		EndTime:     endTime,
		AllDay:      req.AllDay,
		Recurring:   req.Recurring,
		Color:       req.Color,
	}

	if event.Color == "" {
		event.Color = "#4A8B8C"
	}

	createdEvent, err := h.service.CreateEvent(r.Context(), claims.TenantID, claims.UserID, event)
	if err != nil {
		http.Error(w, "Failed to create event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdEvent)
}

// UpdateEvent updates an existing event
func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	eventID := chi.URLParam(r, "id")

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse times
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		http.Error(w, "Invalid start_time format (expected RFC3339/ISO 8601)", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		http.Error(w, "Invalid end_time format (expected RFC3339/ISO 8601)", http.StatusBadRequest)
		return
	}

	// Validate recurring value
	if req.Recurring != "none" && req.Recurring != "weekly" && req.Recurring != "monthly" {
		req.Recurring = "none"
	}

	event := &Event{
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		StartTime:   startTime,
		EndTime:     endTime,
		AllDay:      req.AllDay,
		Recurring:   req.Recurring,
		Color:       req.Color,
	}

	if event.Color == "" {
		event.Color = "#4A8B8C"
	}

	updatedEvent, err := h.service.UpdateEvent(r.Context(), claims.TenantID, eventID, event)
	if err != nil {
		http.Error(w, "Failed to update event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedEvent)
}

// DeleteEvent deletes an event
func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	eventID := chi.URLParam(r, "id")

	if err := h.service.DeleteEvent(r.Context(), claims.TenantID, eventID); err != nil {
		http.Error(w, "Failed to delete event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUpcomingEvents returns the next N upcoming events
func (h *Handler) GetUpcomingEvents(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}

	events, err := h.service.GetUpcomingEvents(r.Context(), claims.TenantID, limit)
	if err != nil {
		http.Error(w, "Failed to get upcoming events: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// ExportICal exports events as iCalendar format
func (h *Handler) ExportICal(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ical, err := h.service.GenerateICal(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to generate iCal: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=calendar.ics")
	w.Write([]byte(ical))
}
