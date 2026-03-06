package public

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

// NewHandler creates a new public handler with database and JWT secret
func NewHandler(db *pgxpool.Pool, jwtSecret string) *Handler {
	publicService := NewService(db, jwtSecret)
	return &Handler{service: publicService}
}

// RespondResponse handles volunteer accept/decline responses
type RespondResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Status    string `json:"status,omitempty"`
	Error     string `json:"error,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// Respond handles public response to volunteer assignment (no auth required)
func (h *Handler) Respond(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		h.sendError(w, "Invalid or missing token", http.StatusBadRequest)
		return
	}

	action := r.URL.Query().Get("action")
	if action != "accept" && action != "decline" {
		h.sendError(w, "Invalid action - must be 'accept' or 'decline'", http.StatusBadRequest)
		return
	}

	// Validate token and get assignment details
	assignmentData, err := h.service.ValidateAssignmentToken(token)
	if err != nil {
		h.sendError(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Process the response
	err = h.service.ProcessResponse(r.Context(), assignmentData.AssignmentID, assignmentData.PersonID, action)
	if err != nil {
		h.sendError(w, "Failed to process response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Render success page
	h.renderSuccessPage(w, action, assignmentData.ServiceName, assignmentData.ServiceDate)
}

func (h *Handler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)
	
	// Simple error page
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Error</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #f5f5f5; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
        .container { background-color: white; padding: 32px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); text-align: center; max-width: 400px; }
        h1 { color: #d9534f; margin-top: 0; }
        p { color: #333; line-height: 1.6; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Error</h1>
        <p>` + message + `</p>
    </div>
</body>
</html>`
	
	w.Write([]byte(html))
}

func (h *Handler) renderSuccessPage(w http.ResponseWriter, action, serviceName, serviceDate string) {
	w.Header().Set("Content-Type", "text/html")
	
	confirmationText := ""
	if action == "accept" {
		confirmationText = "Thank you for confirming! Your volunteer assignment has been recorded."
	} else {
		confirmationText = "We've received your response. Thank you for letting us know."
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + strings.Title(action) + ` - Volunteer Assignment</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #f5f5f5; display: flex; align-items: center; justify-content: center; min-height: 100vh; margin: 0; }
        .container { background-color: white; padding: 48px 32px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); text-align: center; max-width: 500px; }
        .icon { font-size: 64px; margin-bottom: 16px; }
        h1 { color: #0D1B2A; margin-top: 0; font-size: 28px; }
        p { color: #333; line-height: 1.6; font-size: 18px; }
        .details { background-color: #f8f9fa; padding: 20px; border-radius: 6px; margin: 24px 0; }
        .detail-row { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #e9ecef; }
        .detail-label { font-weight: 600; color: #555; }
        .detail-value { color: #333; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">` + getIcon(action) + `</div>
        <h1>` + strings.Title(action) + `!</h1>
        <p>` + confirmationText + `</p>
        
        <div class="details">
            <div class="detail-row">
                <span class="detail-label">Service:</span>
                <span class="detail-value">` + serviceName + `</span>
            </div>
            <div class="detail-row">
                <span class="detail-label">Date:</span>
                <span class="detail-value">` + serviceDate + `</span>
            </div>
        </div>
        
        <p style="font-size: 14px; color: #6c757d;">Thank you for serving! 🙏</p>
    </div>
    
    <script>
        // Auto-redirect to app after 3 seconds
        setTimeout(function() { window.location.href = '/'; }, 3000);
    </script>
</body>
</html>`

	w.Write([]byte(html))
}

// GetChurchInfo returns public church info
func (h *Handler) GetChurchInfo(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	info, err := h.service.GetChurchInfo(r.Context(), tenantID)
	if err != nil {
		http.Error(w, "Church not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// GetEvents returns public events
func (h *Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	events, err := h.service.GetPublicEvents(r.Context(), tenantID)
	if err != nil {
		http.Error(w, "Failed to get events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// GetPublicGroups returns public groups for the group finder
func (h *Handler) GetPublicGroups(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	groups, err := h.service.GetPublicGroups(r.Context(), tenantID)
	if err != nil {
		http.Error(w, "Failed to get groups", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// GroupSignup handles public group signup
func (h *Handler) GroupSignup(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.service.GroupSignup(r.Context(), groupID, req.Name, req.Email, req.Phone)
	if err != nil {
		http.Error(w, "Failed to process signup: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Signup submitted successfully"})
}

func getIcon(action string) string {
	if action == "accept" {
		return "✅"
	}
	return "👋"
}
