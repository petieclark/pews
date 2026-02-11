package rooms

import (
	"encoding/json"
	"net/http"
	"time"

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

// Rooms

type CreateRoomRequest struct {
	Name        string   `json:"name"`
	Capacity    *int     `json:"capacity"`
	Description string   `json:"description"`
	Color       string   `json:"color"`
	Amenities   []string `json:"amenities"`
}

type UpdateRoomRequest struct {
	Name        string   `json:"name"`
	Capacity    *int     `json:"capacity"`
	Description string   `json:"description"`
	Color       string   `json:"color"`
	Amenities   []string `json:"amenities"`
	IsActive    bool     `json:"is_active"`
}

func (h *Handler) ListRooms(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	activeOnly := r.URL.Query().Get("active") == "true"

	rooms, err := h.service.ListRooms(r.Context(), claims.TenantID, activeOnly)
	if err != nil {
		http.Error(w, "Failed to list rooms: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

func (h *Handler) GetRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	roomID := chi.URLParam(r, "id")

	room, err := h.service.GetRoom(r.Context(), claims.TenantID, roomID)
	if err != nil {
		http.Error(w, "Failed to get room: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	var req CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amenities == nil {
		req.Amenities = []string{}
	}

	room, err := h.service.CreateRoom(r.Context(), claims.TenantID, req.Name, req.Description, req.Color, req.Capacity, req.Amenities)
	if err != nil {
		http.Error(w, "Failed to create room: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(room)
}

func (h *Handler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	roomID := chi.URLParam(r, "id")

	var req UpdateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amenities == nil {
		req.Amenities = []string{}
	}

	room, err := h.service.UpdateRoom(r.Context(), claims.TenantID, roomID, req.Name, req.Description, req.Color, req.Capacity, req.Amenities, req.IsActive)
	if err != nil {
		http.Error(w, "Failed to update room: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

func (h *Handler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	roomID := chi.URLParam(r, "id")

	if err := h.service.DeleteRoom(r.Context(), claims.TenantID, roomID); err != nil {
		http.Error(w, "Failed to delete room: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Bookings

type CreateBookingRequest struct {
	RoomID    string  `json:"room_id"`
	EventName string  `json:"event_name"`
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Recurring *string `json:"recurring"`
	Status    string  `json:"status"`
	Notes     *string `json:"notes"`
}

type UpdateBookingRequest struct {
	EventName string  `json:"event_name"`
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Recurring *string `json:"recurring"`
	Status    string  `json:"status"`
	Notes     *string `json:"notes"`
}

func (h *Handler) ListRoomBookings(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	roomID := chi.URLParam(r, "id")

	var startTime, endTime *time.Time
	if startStr := r.URL.Query().Get("start"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = &t
		}
	}
	if endStr := r.URL.Query().Get("end"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = &t
		}
	}

	bookings, err := h.service.ListBookings(r.Context(), claims.TenantID, &roomID, startTime, endTime)
	if err != nil {
		http.Error(w, "Failed to list bookings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}

func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		http.Error(w, "Invalid start_time format", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		http.Error(w, "Invalid end_time format", http.StatusBadRequest)
		return
	}

	if endTime.Before(startTime) || endTime.Equal(startTime) {
		http.Error(w, "end_time must be after start_time", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		req.Status = "confirmed"
	}

	booking, err := h.service.CreateBooking(r.Context(), claims.TenantID, req.RoomID, req.EventName, &claims.UserID, startTime, endTime, req.Recurring, req.Status, req.Notes)
	if err != nil {
		http.Error(w, "Failed to create booking: "+err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(booking)
}

func (h *Handler) UpdateBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	bookingID := chi.URLParam(r, "bookingId")

	var req UpdateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		http.Error(w, "Invalid start_time format", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		http.Error(w, "Invalid end_time format", http.StatusBadRequest)
		return
	}

	if endTime.Before(startTime) || endTime.Equal(startTime) {
		http.Error(w, "end_time must be after start_time", http.StatusBadRequest)
		return
	}

	booking, err := h.service.UpdateBooking(r.Context(), claims.TenantID, bookingID, req.EventName, startTime, endTime, req.Recurring, req.Status, req.Notes)
	if err != nil {
		http.Error(w, "Failed to update booking: "+err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(booking)
}

func (h *Handler) DeleteBooking(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	bookingID := chi.URLParam(r, "bookingId")

	if err := h.service.DeleteBooking(r.Context(), claims.TenantID, bookingID); err != nil {
		http.Error(w, "Failed to delete booking: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		http.Error(w, "start and end query parameters required", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		http.Error(w, "Invalid start time format", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		http.Error(w, "Invalid end time format", http.StatusBadRequest)
		return
	}

	availability, err := h.service.GetRoomAvailability(r.Context(), claims.TenantID, startTime, endTime)
	if err != nil {
		http.Error(w, "Failed to check availability: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availability)
}
