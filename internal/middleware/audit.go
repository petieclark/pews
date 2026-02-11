package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/petieclark/pews/internal/audit"
)

type AuditMiddleware struct {
	auditService *audit.Service
}

func NewAuditMiddleware(auditService *audit.Service) *AuditMiddleware {
	return &AuditMiddleware{auditService: auditService}
}

// responseCapture wraps http.ResponseWriter to capture response body
type responseCapture struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body.Write(b)
	return rc.ResponseWriter.Write(b)
}

func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.status = statusCode
	rc.ResponseWriter.WriteHeader(statusCode)
}

// AuditLog middleware logs all mutating requests (POST/PUT/DELETE)
func (am *AuditMiddleware) AuditLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only log mutating requests
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
			next.ServeHTTP(w, r)
			return
		}

		// Skip webhook and public routes
		if strings.Contains(r.URL.Path, "/webhook") || 
		   strings.Contains(r.URL.Path, "/api/auth/register") ||
		   strings.Contains(r.URL.Path, "/api/auth/login") {
			next.ServeHTTP(w, r)
			return
		}

		// Extract request context
		tenantID, _ := r.Context().Value("tenant_id").(string)
		userID, _ := r.Context().Value("user_id").(string)
		userEmail, _ := r.Context().Value("email").(string)

		// Get IP address
		ipAddress := r.Header.Get("X-Forwarded-For")
		if ipAddress == "" {
			ipAddress = r.Header.Get("X-Real-IP")
		}
		if ipAddress == "" {
			ipAddress = r.RemoteAddr
		}

		// Get user agent
		userAgent := r.Header.Get("User-Agent")

		// Read and restore request body
		var requestBody []byte
		if r.Body != nil {
			requestBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Capture response
		responseBuffer := &bytes.Buffer{}
		captureWriter := &responseCapture{
			ResponseWriter: w,
			body:           responseBuffer,
			status:         http.StatusOK,
		}

		// Process request
		next.ServeHTTP(captureWriter, r)

		// Only log successful requests (2xx status codes)
		if captureWriter.status < 200 || captureWriter.status >= 300 {
			return
		}

		// Determine action and entity from URL and method
		action := am.determineAction(r.Method, r.URL.Path)
		entityType, entityID := am.extractEntity(r.URL.Path, r.Method)

		// Parse request/response for old/new values
		var oldValue, newValue interface{}
		
		if r.Method == http.MethodPost {
			// For POST, new value is the request body
			json.Unmarshal(requestBody, &newValue)
		} else if r.Method == http.MethodPut {
			// For PUT, new value is request, old value could be in response (if we fetch before update)
			json.Unmarshal(requestBody, &newValue)
		} else if r.Method == http.MethodDelete {
			// For DELETE, we might not have old value unless we fetched it
			// The ID is in the URL
		}

		// Create metadata
		metadata := map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"user_email": userEmail,
		}

		// Log the action
		var userIDPtr *string
		if userID != "" {
			userIDPtr = &userID
		}

		var ipPtr, uaPtr *string
		if ipAddress != "" {
			ipPtr = &ipAddress
		}
		if userAgent != "" {
			uaPtr = &userAgent
		}

		entry := audit.LogEntry{
			Action:     action,
			EntityType: entityType,
			EntityID:   entityID,
			OldValue:   oldValue,
			NewValue:   newValue,
			Metadata:   metadata,
		}

		// Log asynchronously to not block the request
		go am.auditService.Log(r.Context(), &tenantID, userIDPtr, ipPtr, uaPtr, entry)
	})
}

// determineAction maps HTTP method and path to action constant
func (am *AuditMiddleware) determineAction(method, path string) string {
	switch method {
	case http.MethodPost:
		if strings.Contains(path, "/enable") {
			return audit.ActionModuleEnable
		}
		if strings.Contains(path, "/disable") {
			return audit.ActionModuleDisable
		}
		return audit.ActionCreate
	case http.MethodPut:
		if strings.Contains(path, "/tenant") && !strings.Contains(path, "/modules") {
			return audit.ActionSettingsChange
		}
		return audit.ActionUpdate
	case http.MethodDelete:
		return audit.ActionDelete
	default:
		return "unknown"
	}
}

// extractEntity attempts to extract entity type and ID from path
func (am *AuditMiddleware) extractEntity(path, method string) (*string, *string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	
	// Pattern: /api/{entity_type}/... or /api/{entity_type}/{id}
	if len(parts) < 3 {
		return nil, nil
	}

	entityType := parts[2] // e.g., "people", "groups", "giving"
	
	// For paths like /api/people/{id} or /api/groups/{id}/members
	var entityID *string
	if len(parts) >= 4 && parts[3] != "" && !strings.Contains(parts[3], "{") {
		// Check if it's a UUID-like ID (not a subresource like "members")
		if !isSubResource(parts[3]) {
			entityID = &parts[3]
		}
	}

	return &entityType, entityID
}

// isSubResource checks if a path segment is a subresource name rather than an ID
func isSubResource(segment string) bool {
	subresources := []string{"members", "tags", "items", "team", "steps", "funds", "donations", "modules", "checkups"}
	for _, sr := range subresources {
		if segment == sr {
			return true
		}
	}
	return false
}
