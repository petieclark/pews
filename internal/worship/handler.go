package worship

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListPlans returns all service plans
func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	plans, err := h.service.ListPlans(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to list plans: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}

// GetPlan returns a specific service plan with items
func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	planID := chi.URLParam(r, "id")
	if planID == "" {
		http.Error(w, "Plan ID is required", http.StatusBadRequest)
		return
	}

	plan, err := h.service.GetPlan(r.Context(), claims.TenantID, planID)
	if err != nil {
		http.Error(w, "Plan not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

// CreatePlanRequest represents the request to create a service plan
type CreatePlanRequest struct {
	ServiceID string `json:"service_id"`
	Notes     string `json:"notes"`
}

// CreatePlan creates a new service plan
func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ServiceID == "" {
		http.Error(w, "service_id is required", http.StatusBadRequest)
		return
	}

	plan, err := h.service.CreatePlan(r.Context(), claims.TenantID, req.ServiceID, claims.UserID, req.Notes)
	if err != nil {
		http.Error(w, "Failed to create plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(plan)
}

// UpdatePlanRequest represents the request to update a service plan
type UpdatePlanRequest struct {
	Notes string `json:"notes"`
}

// UpdatePlan updates a service plan
func (h *Handler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	planID := chi.URLParam(r, "id")
	if planID == "" {
		http.Error(w, "Plan ID is required", http.StatusBadRequest)
		return
	}

	var req UpdatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	plan, err := h.service.UpdatePlan(r.Context(), claims.TenantID, planID, req.Notes)
	if err != nil {
		http.Error(w, "Failed to update plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

// PublishPlan publishes a service plan
func (h *Handler) PublishPlan(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	planID := chi.URLParam(r, "id")
	if planID == "" {
		http.Error(w, "Plan ID is required", http.StatusBadRequest)
		return
	}

	plan, err := h.service.PublishPlan(r.Context(), claims.TenantID, planID)
	if err != nil {
		http.Error(w, "Failed to publish plan: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

// AddItemRequest represents the request to add an item to a service plan
type AddItemRequest struct {
	ItemOrder       int     `json:"item_order"`
	ItemType        string  `json:"item_type"`
	Title           string  `json:"title"`
	DurationMinutes *int    `json:"duration_minutes,omitempty"`
	Notes           string  `json:"notes,omitempty"`
	SongID          *string `json:"song_id,omitempty"`
	AssignedTo      *string `json:"assigned_to,omitempty"`
	Key             *string `json:"key,omitempty"`  // Song key for transposition (e.g., "G", "Cm")
}

// AddItem adds an item to a service plan
func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	planID := chi.URLParam(r, "id")
	if planID == "" {
		http.Error(w, "Plan ID is required", http.StatusBadRequest)
		return
	}

	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ItemType == "" || req.Title == "" {
		http.Error(w, "item_type and title are required", http.StatusBadRequest)
		return
	}

	item, err := h.service.AddItem(r.Context(), claims.TenantID, planID, req.ItemOrder, req.ItemType, req.Title, req.DurationMinutes, req.Notes, req.SongID, req.AssignedTo, req.Key)
	if err != nil {
		http.Error(w, "Failed to add item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// UpdateItemRequest represents the request to update a plan item
type UpdateItemRequest struct {
	ItemOrder       int     `json:"item_order"`
	ItemType        string  `json:"item_type"`
	Title           string  `json:"title"`
	DurationMinutes *int    `json:"duration_minutes,omitempty"`
	Notes           string  `json:"notes,omitempty"`
	SongID          *string `json:"song_id,omitempty"`
	AssignedTo      *string `json:"assigned_to,omitempty"`
	Key             *string `json:"key,omitempty"`  // Song key for transposition (e.g., "G", "Cm")
}

// UpdateItem updates a plan item
func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	planID := chi.URLParam(r, "id")
	itemID := chi.URLParam(r, "itemId")
	if planID == "" || itemID == "" {
		http.Error(w, "Plan ID and Item ID are required", http.StatusBadRequest)
		return
	}

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ItemType == "" || req.Title == "" {
		http.Error(w, "item_type and title are required", http.StatusBadRequest)
		return
	}

	item, err := h.service.UpdateItem(r.Context(), claims.TenantID, planID, itemID, req.ItemOrder, req.ItemType, req.Title, req.DurationMinutes, req.Notes, req.SongID, req.AssignedTo, req.Key)
	if err != nil {
		http.Error(w, "Failed to update item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// DeleteItem deletes a plan item
func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	planID := chi.URLParam(r, "id")
	itemID := chi.URLParam(r, "itemId")
	if planID == "" || itemID == "" {
		http.Error(w, "Plan ID and Item ID are required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteItem(r.Context(), claims.TenantID, planID, itemID); err != nil {
		http.Error(w, "Failed to delete item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ExportPlan exports a service plan as HTML for printing
func (h *Handler) ExportPlan(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	planID := chi.URLParam(r, "id")
	if planID == "" {
		http.Error(w, "Plan ID is required", http.StatusBadRequest)
		return
	}

	plan, err := h.service.GetPlan(r.Context(), claims.TenantID, planID)
	if err != nil {
		http.Error(w, "Plan not found", http.StatusNotFound)
		return
	}

	// Generate HTML
	html := generatePlanHTML(plan)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// GetSharedPlan returns a service plan by token (no auth required for public access)
func (h *Handler) GetSharedPlan(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	plan, err := h.service.GetPlanByToken(r.Context(), token)
	if err != nil {
		http.Error(w, "Plan not found or invalid token", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plan)
}

// generatePlanHTML generates a printable HTML version of the service plan with CCLI attributions
func generatePlanHTML(plan *ServicePlan) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Service Plan</title>
    <style>
        body { font-family: Arial, sans-serif; padding: 20px; }
        h1 { border-bottom: 2px solid #333; padding-bottom: 10px; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f2f2f2; }
        .notes { font-style: italic; color: #666; }
        .attribution { font-size: 0.85em; color: #555; margin-top: 4px; line-height: 1.4; }
        .key-badge { display: inline-block; padding: 2px 8px; background-color: #e0f2fe; color: #0369a1; border-radius: 4px; font-weight: bold; font-size: 0.9em; margin-left: 8px; }
        @media print {
            body { margin: 0; }
            button { display: none; }
        }
    </style>
</head>
<body>
    <button onclick="window.print()">Print</button>
    <h1>Service Plan</h1>
    <p><strong>Status:</strong> ` + plan.Status + `</p>
    <p><strong>Notes:</strong> ` + plan.Notes + `</p>
    <table>
        <thead>
            <tr>
                <th>#</th>
                <th>Type</th>
                <th>Title & Key</th>
                <th>Duration</th>
                <th>Assigned To</th>
                <th>Notes</th>
            </tr>
        </thead>
        <tbody>`

	totalDuration := 0
	for _, item := range plan.Items {
		durationStr := ""
		if item.DurationMinutes != nil {
			durationStr = fmt.Sprintf("%d min", *item.DurationMinutes)
			totalDuration += *item.DurationMinutes
		}

		// Build title with key badge if present
		titleCell := item.Title
		if item.Key != "" {
			titleCell += fmt.Sprintf(` <span class="key-badge">%s</span>`, item.Key)
		}

		// Add CCLI attribution line for songs
		attributionHTML := ""
		if item.ItemType == "song" && item.SongTitle != "" {
			var attribParts []string
			// Song title (in case different from plan item title)
			if item.SongTitle != item.Title {
				attribParts = append(attribParts, fmt.Sprintf("Song: %s", item.SongTitle))
			}
			// Key display in attribution if not shown as badge
			if item.Key != "" {
				attribParts = append(attribParts, fmt.Sprintf("Key: %s", item.Key))
			}
			if len(attribParts) > 0 {
				attributionHTML = fmt.Sprintf(`<div class="attribution">%s</div>`, strings.Join(attribParts, " • "))
			}
		}

		html += fmt.Sprintf(`
            <tr>
                <td>%d</td>
                <td>%s</td>
                <td>
                    %s
                    %s
                </td>
                <td>%s</td>
                <td>%s</td>
                <td class="notes">%s</td>
            </tr>`,
			item.ItemOrder,
			item.ItemType,
			titleCell,
			attributionHTML,
			durationStr,
			item.AssignedToName,
			item.Notes,
		)
	}

	html += fmt.Sprintf(`
        </tbody>
        <tfoot>
            <tr>
                <th colspan="3">Total Duration</th>
                <th>%d min</th>
                <th colspan="2"></th>
            </tr>
        </tfoot>
    </table>
</body>
</html>`, totalDuration)

	return html
}
