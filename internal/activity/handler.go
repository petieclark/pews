package activity

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListActivity handles GET /api/activity
func (h *Handler) ListActivity(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value("tenant_id").(string); if !ok { http.Error(w, "unauthorized", http.StatusUnauthorized); return }

	// Parse query parameters
	query := r.URL.Query()
	entityType := query.Get("entity_type")
	userID := query.Get("user_id")
	
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	var startDate, endDate *time.Time
	if startDateStr := query.Get("start_date"); startDateStr != "" {
		if t, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &t
		}
	}
	if endDateStr := query.Get("end_date"); endDateStr != "" {
		if t, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &t
		}
	}

	params := ListActivityParams{
		TenantID:   tenantID,
		EntityType: entityType,
		UserID:     userID,
		StartDate:  startDate,
		EndDate:    endDate,
		Page:       page,
		Limit:      limit,
	}

	logs, total, err := h.service.ListActivity(r.Context(), params)
	if err != nil {
		http.Error(w, "Failed to fetch activity logs", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
