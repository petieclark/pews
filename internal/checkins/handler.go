package checkins

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

// ========== Stations ==========

func (h *Handler) ListStations(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	stations, err := h.service.ListStations(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to list stations: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stations)
}

type CreateStationRequest struct {
	Name     string `json:"name"`
	Location string `json:"location,omitempty"`
	IsActive bool   `json:"is_active"`
}

func (h *Handler) CreateStation(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var req CreateStationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	station := &Station{Name: req.Name, Location: req.Location, IsActive: req.IsActive}
	created, err := h.service.CreateStation(r.Context(), claims.TenantID, station)
	if err != nil {
		http.Error(w, "Failed to create station: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) UpdateStation(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	stationID := chi.URLParam(r, "id")
	var req CreateStationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	station := &Station{Name: req.Name, Location: req.Location, IsActive: req.IsActive}
	updated, err := h.service.UpdateStation(r.Context(), claims.TenantID, stationID, station)
	if err != nil {
		http.Error(w, "Failed to update station: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// ========== Events ==========

func (h *Handler) ListEvents(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	events, err := h.service.ListEvents(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to list events: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

type CreateEventRequest struct {
	Name      string  `json:"name"`
	EventDate string  `json:"event_date"`
	ServiceID *string `json:"service_id,omitempty"`
	StationID *string `json:"station_id,omitempty"`
	IsActive  bool    `json:"is_active"`
}

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
	event := &Event{Name: req.Name, EventDate: req.EventDate, ServiceID: req.ServiceID, StationID: req.StationID, IsActive: req.IsActive}
	created, err := h.service.CreateEvent(r.Context(), claims.TenantID, event)
	if err != nil {
		http.Error(w, "Failed to create event: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

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
	event := &Event{Name: req.Name, EventDate: req.EventDate, ServiceID: req.ServiceID, StationID: req.StationID, IsActive: req.IsActive}
	updated, err := h.service.UpdateEvent(r.Context(), claims.TenantID, eventID, event)
	if err != nil {
		http.Error(w, "Failed to update event: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	eventID := chi.URLParam(r, "id")
	event, err := h.service.GetEvent(r.Context(), claims.TenantID, eventID)
	if err != nil {
		http.Error(w, "Event not found: "+err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

// ========== Check-in / Check-out ==========

type CheckInRequest struct {
	PersonID  string  `json:"person_id"`
	StationID *string `json:"station_id,omitempty"`
	Notes     string  `json:"notes,omitempty"`
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	eventID := chi.URLParam(r, "id")
	var req CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	result, err := h.service.CheckIn(r.Context(), claims.TenantID, eventID, req.PersonID, req.StationID, req.Notes)
	if err != nil {
		http.Error(w, "Failed to check in: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

type CheckOutRequest struct {
	PersonID string `json:"person_id"`
}

func (h *Handler) CheckOut(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	eventID := chi.URLParam(r, "id")
	var req CheckOutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.service.CheckOut(r.Context(), claims.TenantID, eventID, req.PersonID); err != nil {
		http.Error(w, "Failed to check out: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetAttendees(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	eventID := chi.URLParam(r, "id")
	attendees, err := h.service.GetAttendees(r.Context(), claims.TenantID, eventID)
	if err != nil {
		http.Error(w, "Failed to get attendees: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attendees)
}

// ========== Person History ==========

func (h *Handler) GetPersonHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	personID := chi.URLParam(r, "personId")
	history, err := h.service.GetPersonHistory(r.Context(), claims.TenantID, personID)
	if err != nil {
		http.Error(w, "Failed to get history: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// ========== Medical Alerts ==========

func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	personID := chi.URLParam(r, "personId")
	alerts, err := h.service.GetAlerts(r.Context(), claims.TenantID, personID)
	if err != nil {
		http.Error(w, "Failed to get alerts: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

type CreateAlertRequest struct {
	AlertType   string `json:"alert_type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

func (h *Handler) CreateAlert(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	personID := chi.URLParam(r, "personId")
	var req CreateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	alert := &MedicalAlert{PersonID: personID, AlertType: req.AlertType, Severity: req.Severity, Description: req.Description}
	created, err := h.service.CreateAlert(r.Context(), claims.TenantID, alert)
	if err != nil {
		http.Error(w, "Failed to create alert: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) DeleteAlert(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	alertID := chi.URLParam(r, "alertId")
	if err := h.service.DeleteAlert(r.Context(), claims.TenantID, alertID); err != nil {
		http.Error(w, "Failed to delete alert: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ========== Authorized Pickups ==========

func (h *Handler) GetPickups(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	personID := chi.URLParam(r, "personId")
	pickups, err := h.service.GetPickups(r.Context(), claims.TenantID, personID)
	if err != nil {
		http.Error(w, "Failed to get pickups: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pickups)
}

type CreatePickupRequest struct {
	PickupPersonID string `json:"pickup_person_id"`
	Relationship   string `json:"relationship"`
	IsActive       bool   `json:"is_active"`
}

func (h *Handler) CreatePickup(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	personID := chi.URLParam(r, "personId")
	var req CreatePickupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	pickup := &AuthorizedPickup{ChildID: personID, PickupPersonID: req.PickupPersonID, Relationship: req.Relationship, IsActive: req.IsActive}
	created, err := h.service.CreatePickup(r.Context(), claims.TenantID, pickup)
	if err != nil {
		http.Error(w, "Failed to create pickup: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) DeletePickup(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	pickupID := chi.URLParam(r, "pickupId")
	if err := h.service.DeletePickup(r.Context(), claims.TenantID, pickupID); err != nil {
		http.Error(w, "Failed to delete pickup: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ========== Stats ==========

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	stats, err := h.service.GetTodayStats(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get stats: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ========== Search ==========

func (h *Handler) SearchPeople(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]PersonSearchResult{})
		return
	}
	results, err := h.service.SearchPeople(r.Context(), claims.TenantID, query)
	if err != nil {
		http.Error(w, "Failed to search: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
