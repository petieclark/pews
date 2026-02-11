package streaming

import (
	"encoding/json"
	"fmt"
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

// parseFlexibleTime parses time from various formats including datetime-local input
func parseFlexibleTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}


// Stream handlers

type CreateStreamRequest struct {
	Title                 string  `json:"title"`
	Description           string  `json:"description,omitempty"`
	ServiceID             *string `json:"service_id,omitempty"`
	Status                string  `json:"status"`
	ScheduledStart        string  `json:"scheduled_start,omitempty"`
	StreamType            string  `json:"stream_type"`
	StreamURL             string  `json:"stream_url,omitempty"`
	StreamKey             string  `json:"stream_key,omitempty"`
	EmbedURL              string  `json:"embed_url"`
	ChatEnabled           bool    `json:"chat_enabled"`
	GivingEnabled         bool    `json:"giving_enabled"`
	ConnectionCardEnabled bool    `json:"connection_card_enabled"`
}

func (h *Handler) ListStreams(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	streams, total, err := h.service.ListStreams(r.Context(), claims.TenantID, status, page, limit)
	if err != nil {
		http.Error(w, "Failed to list streams: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"streams": streams,
		"total":   total,
		"page":    page,
		"limit":   limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetStream(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	streamID := chi.URLParam(r, "id")
	stream, err := h.service.GetStreamByID(r.Context(), claims.TenantID, streamID)
	if err != nil {
		http.Error(w, "Stream not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stream)
}

func (h *Handler) CreateStream(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	stream := &Stream{
		Title:                 req.Title,
		Description:           req.Description,
		ServiceID:             req.ServiceID,
		Status:                req.Status,
		StreamType:            req.StreamType,
		StreamURL:             req.StreamURL,
		StreamKey:             req.StreamKey,
		EmbedURL:              req.EmbedURL,
		ChatEnabled:           req.ChatEnabled,
		GivingEnabled:         req.GivingEnabled,
		ConnectionCardEnabled: req.ConnectionCardEnabled,
	}

	if req.ScheduledStart != "" {
		scheduledStart, err := parseFlexibleTime(req.ScheduledStart)
		if err != nil {
			http.Error(w, "Invalid scheduled_start format", http.StatusBadRequest)
			return
		}
		stream.ScheduledStart = &scheduledStart
	}

	if err := h.service.CreateStream(r.Context(), claims.TenantID, stream); err != nil {
		http.Error(w, "Failed to create stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(stream)
}

func (h *Handler) UpdateStream(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	streamID := chi.URLParam(r, "id")

	var req CreateStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	stream := &Stream{
		Title:                 req.Title,
		Description:           req.Description,
		ServiceID:             req.ServiceID,
		Status:                req.Status,
		StreamType:            req.StreamType,
		StreamURL:             req.StreamURL,
		StreamKey:             req.StreamKey,
		EmbedURL:              req.EmbedURL,
		ChatEnabled:           req.ChatEnabled,
		GivingEnabled:         req.GivingEnabled,
		ConnectionCardEnabled: req.ConnectionCardEnabled,
	}

	if req.ScheduledStart != "" {
		scheduledStart, err := parseFlexibleTime(req.ScheduledStart)
		if err != nil {
			http.Error(w, "Invalid scheduled_start format", http.StatusBadRequest)
			return
		}
		stream.ScheduledStart = &scheduledStart
	}

	if err := h.service.UpdateStream(r.Context(), claims.TenantID, streamID, stream); err != nil {
		http.Error(w, "Failed to update stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stream)
}

func (h *Handler) DeleteStream(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	streamID := chi.URLParam(r, "id")

	if err := h.service.DeleteStream(r.Context(), claims.TenantID, streamID); err != nil {
		http.Error(w, "Failed to delete stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GoLive(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	streamID := chi.URLParam(r, "id")

	if err := h.service.GoLive(r.Context(), claims.TenantID, streamID); err != nil {
		http.Error(w, "Failed to go live: "+err.Error(), http.StatusInternalServerError)
		return
	}

	stream, err := h.service.GetStreamByID(r.Context(), claims.TenantID, streamID)
	if err != nil {
		http.Error(w, "Failed to get stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stream)
}

func (h *Handler) EndStream(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	streamID := chi.URLParam(r, "id")

	if err := h.service.EndStream(r.Context(), claims.TenantID, streamID); err != nil {
		http.Error(w, "Failed to end stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	stream, err := h.service.GetStreamByID(r.Context(), claims.TenantID, streamID)
	if err != nil {
		http.Error(w, "Failed to get stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stream)
}

func (h *Handler) GetLiveStream(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stream, err := h.service.GetLiveStream(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get live stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if stream == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"stream": nil,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stream)
}

// Chat handlers

type SendChatMessageRequest struct {
	PersonID  *string `json:"person_id,omitempty"`
	GuestName string  `json:"guest_name,omitempty"`
	Message   string  `json:"message"`
}

func (h *Handler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")
	after := r.URL.Query().Get("after")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	messages, err := h.service.GetChatMessages(r.Context(), streamID, after, limit)
	if err != nil {
		http.Error(w, "Failed to get chat messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
	})
}

func (h *Handler) SendChatMessage(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")

	var req SendChatMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Try to get person ID from auth if available
	var personID *string
	if claims, ok := middleware.GetClaims(r.Context()); ok {
		personID = &claims.UserID
	} else if req.PersonID != nil {
		personID = req.PersonID
	}

	message, err := h.service.SendChatMessage(r.Context(), streamID, personID, req.GuestName, req.Message)
	if err != nil {
		http.Error(w, "Failed to send message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

func (h *Handler) PinChatMessage(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	messageID := chi.URLParam(r, "msgId")

	if err := h.service.PinChatMessage(r.Context(), claims.TenantID, messageID); err != nil {
		http.Error(w, "Failed to pin message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteChatMessage(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	messageID := chi.URLParam(r, "msgId")

	if err := h.service.DeleteChatMessage(r.Context(), claims.TenantID, messageID); err != nil {
		http.Error(w, "Failed to delete message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Viewer handlers

type JoinStreamRequest struct {
	PersonID  *string `json:"person_id,omitempty"`
	GuestName string  `json:"guest_name,omitempty"`
}

func (h *Handler) JoinStream(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")

	var req JoinStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Try to get person ID from auth if available
	var personID *string
	if claims, ok := middleware.GetClaims(r.Context()); ok {
		personID = &claims.UserID
	} else if req.PersonID != nil {
		personID = req.PersonID
	}

	viewer, err := h.service.JoinStream(r.Context(), streamID, personID, req.GuestName)
	if err != nil {
		http.Error(w, "Failed to join stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(viewer)
}

type LeaveStreamRequest struct {
	ViewerID string `json:"viewer_id"`
}

func (h *Handler) LeaveStream(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")

	var req LeaveStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.LeaveStream(r.Context(), streamID, req.ViewerID); err != nil {
		http.Error(w, "Failed to leave stream: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetViewers(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	streamID := chi.URLParam(r, "id")

	viewers, count, err := h.service.GetViewers(r.Context(), claims.TenantID, streamID)
	if err != nil {
		http.Error(w, "Failed to get viewers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"viewers": viewers,
		"count":   count,
	})
}

// Notes handlers

func (h *Handler) GetStreamNotes(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	streamID := chi.URLParam(r, "id")

	notes, err := h.service.GetStreamNotes(r.Context(), claims.TenantID, streamID, claims.UserID)
	if err != nil {
		http.Error(w, "Failed to get notes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if notes == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"notes": nil,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

type SaveStreamNotesRequest struct {
	Content string `json:"content"`
}

func (h *Handler) SaveStreamNotes(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	streamID := chi.URLParam(r, "id")

	var req SaveStreamNotesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	notes, err := h.service.SaveStreamNotes(r.Context(), claims.TenantID, streamID, claims.UserID, req.Content)
	if err != nil {
		http.Error(w, "Failed to save notes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

// Public handlers

func (h *Handler) GetWatchStream(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")

	stream, err := h.service.GetStreamByIDPublic(r.Context(), streamID)
	if err != nil {
		http.Error(w, "Stream not found: "+err.Error(), http.StatusNotFound)
		return
	}

	// Return only public-facing info
	response := map[string]interface{}{
		"id":                      stream.ID,
		"title":                   stream.Title,
		"description":             stream.Description,
		"status":                  stream.Status,
		"embed_url":               stream.EmbedURL,
		"chat_enabled":            stream.ChatEnabled,
		"giving_enabled":          stream.GivingEnabled,
		"connection_card_enabled": stream.ConnectionCardEnabled,
		"viewer_count":            stream.ViewerCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
