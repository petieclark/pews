package audit

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type GetLogsResponse struct {
	Logs       []AuditLog `json:"logs"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// GetLogs handles GET /api/audit/logs
func (h *Handler) GetLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	// Parse query parameters
	page := 1
	pageSize := 50
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	offset := (page - 1) * pageSize

	// Optional filters
	var userID, action, entityType *string
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		userID = &uid
	}
	if act := r.URL.Query().Get("action"); act != "" {
		action = &act
	}
	if et := r.URL.Query().Get("entity_type"); et != "" {
		entityType = &et
	}

	logs, total, err := h.service.GetLogs(r.Context(), tenantID, userID, action, entityType, pageSize, offset)
	if err != nil {
		http.Error(w, "Failed to fetch audit logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := (total + pageSize - 1) / pageSize

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetLogsResponse{
		Logs:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetUserLogs handles GET /api/audit/logs/user/:id
func (h *Handler) GetUserLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)
	userID := chi.URLParam(r, "id")

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Parse pagination
	page := 1
	pageSize := 50
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	offset := (page - 1) * pageSize

	logs, total, err := h.service.GetLogsByUser(r.Context(), tenantID, userID, pageSize, offset)
	if err != nil {
		http.Error(w, "Failed to fetch user logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := (total + pageSize - 1) / pageSize

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetLogsResponse{
		Logs:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetSecurityDashboard handles GET /api/audit/security
func (h *Handler) GetSecurityDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	dashboard, err := h.service.GetSecurityDashboard(r.Context(), tenantID)
	if err != nil {
		http.Error(w, "Failed to fetch security dashboard: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// ExportLogs handles GET /api/audit/export with CSV output
func (h *Handler) ExportLogs(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value("tenant_id").(string)

	// Fetch all logs (with reasonable limit)
	logs, _, err := h.service.GetLogs(r.Context(), tenantID, nil, nil, nil, 10000, 0)
	if err != nil {
		http.Error(w, "Failed to export logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=audit_logs_%s.csv", time.Now().Format("2006-01-02")))

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write header row
	csvWriter.Write([]string{
		"Timestamp", "User ID", "Action", "Entity Type", "Entity ID", "IP Address", "User Agent", "Old Value", "New Value",
	})

	// Write data rows
	for _, log := range logs {
		userID := ""
		if log.UserID != nil {
			userID = *log.UserID
		}
		entityType := ""
		if log.EntityType != nil {
			entityType = *log.EntityType
		}
		entityID := ""
		if log.EntityID != nil {
			entityID = *log.EntityID
		}
		ipAddress := ""
		if log.IPAddress != nil {
			ipAddress = *log.IPAddress
		}
		userAgent := ""
		if log.UserAgent != nil {
			userAgent = *log.UserAgent
		}

		csvWriter.Write([]string{
			log.Timestamp.Format(time.RFC3339),
			userID,
			log.Action,
			entityType,
			entityID,
			ipAddress,
			userAgent,
			string(log.OldValue),
			string(log.NewValue),
		})
	}
}
