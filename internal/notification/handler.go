package notification

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
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

// ListNotifications handles GET /api/notifications
func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	unreadStr := r.URL.Query().Get("unread")

	params := ListNotificationsParams{
		Limit:  20,
		Offset: 0,
		Unread: false,
	}

	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	if offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			params.Offset = offset
		}
	}

	if unreadStr == "true" {
		params.Unread = true
	}

	list, err := h.service.List(r.Context(), claims.TenantID, claims.UserID, params)
	if err != nil {
		http.Error(w, "Failed to list notifications: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// MarkAsRead handles PUT /api/notifications/:id/read
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	notificationID := chi.URLParam(r, "id")
	if notificationID == "" {
		http.Error(w, "Notification ID required", http.StatusBadRequest)
		return
	}

	err := h.service.MarkAsRead(r.Context(), claims.TenantID, claims.UserID, notificationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Notification not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to mark notification as read: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MarkAllAsRead handles PUT /api/notifications/read-all
func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.service.MarkAllAsRead(r.Context(), claims.TenantID, claims.UserID)
	if err != nil {
		http.Error(w, "Failed to mark all as read: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUnreadCount handles GET /api/notifications/unread-count
func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	count, err := h.service.GetUnreadCount(r.Context(), claims.TenantID, claims.UserID)
	if err != nil {
		http.Error(w, "Failed to get unread count: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UnreadCountResponse{Count: count})
}
