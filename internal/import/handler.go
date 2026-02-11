package importpkg

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/petieclark/pews/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ImportPeople handles POST /api/import/people
func (h *Handler) ImportPeople(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if dry_run query param is set
	dryRun := r.URL.Query().Get("dry_run") == "true"

	var req ImportPeopleRequest
	contentType := r.Header.Get("Content-Type")

	// Handle CSV content type
	if strings.Contains(contentType, "text/csv") {
		people, err := ParsePeopleCSV(r.Body)
		if err != nil {
			http.Error(w, "Failed to parse CSV: "+err.Error(), http.StatusBadRequest)
			return
		}
		req.People = people
		req.DryRun = dryRun
	} else {
		// Handle JSON content type
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Query param overrides request body
		if dryRun {
			req.DryRun = true
		}
	}

	result, err := h.service.ImportPeople(r.Context(), claims.TenantID, req.People, req.DryRun)
	if err != nil {
		http.Error(w, "Failed to import people: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ImportGroups handles POST /api/import/groups
func (h *Handler) ImportGroups(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dryRun := r.URL.Query().Get("dry_run") == "true"

	var req ImportGroupsRequest
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "text/csv") {
		groups, err := ParseGroupsCSV(r.Body)
		if err != nil {
			http.Error(w, "Failed to parse CSV: "+err.Error(), http.StatusBadRequest)
			return
		}
		req.Groups = groups
		req.DryRun = dryRun
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		if dryRun {
			req.DryRun = true
		}
	}

	result, err := h.service.ImportGroups(r.Context(), claims.TenantID, req.Groups, req.DryRun)
	if err != nil {
		http.Error(w, "Failed to import groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ImportSongs handles POST /api/import/songs
func (h *Handler) ImportSongs(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dryRun := r.URL.Query().Get("dry_run") == "true"

	var req ImportSongsRequest
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "text/csv") {
		songs, err := ParseSongsCSV(r.Body)
		if err != nil {
			http.Error(w, "Failed to parse CSV: "+err.Error(), http.StatusBadRequest)
			return
		}
		req.Songs = songs
		req.DryRun = dryRun
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		if dryRun {
			req.DryRun = true
		}
	}

	result, err := h.service.ImportSongs(r.Context(), claims.TenantID, req.Songs, req.DryRun)
	if err != nil {
		http.Error(w, "Failed to import songs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ImportGiving handles POST /api/import/giving
func (h *Handler) ImportGiving(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	dryRun := r.URL.Query().Get("dry_run") == "true"

	var req ImportGivingRequest
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "text/csv") {
		donations, err := ParseGivingCSV(r.Body)
		if err != nil {
			http.Error(w, "Failed to parse CSV: "+err.Error(), http.StatusBadRequest)
			return
		}
		req.Donations = donations
		req.DryRun = dryRun
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		if dryRun {
			req.DryRun = true
		}
	}

	result, err := h.service.ImportGiving(r.Context(), claims.TenantID, req.Donations, req.DryRun)
	if err != nil {
		http.Error(w, "Failed to import donations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ImportPCOPeople handles POST /api/import/pco/people
// Supports flexible PCO column naming and stores unmapped columns in custom_fields
func (h *Handler) ImportPCOPeople(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(50 << 20) // 50 MB max
	if err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Parse PCO CSV with flexible column mapping
	people, err := ParsePCOPeopleCSV(file)
	if err != nil {
		http.Error(w, "Failed to parse PCO CSV: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check for update mode
	updateMode := r.FormValue("update_mode") // "skip" or "update"
	if updateMode == "" {
		updateMode = "skip"
	}

	result, err := h.service.ImportPCOPeople(r.Context(), claims.TenantID, people, updateMode)
	if err != nil {
		http.Error(w, "Failed to import people: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ImportPCOSongs handles POST /api/import/pco/songs
// Supports flexible PCO column naming and stores unmapped columns in notes
func (h *Handler) ImportPCOSongs(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(50 << 20) // 50 MB max
	if err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Parse PCO CSV with flexible column mapping
	songs, err := ParsePCOSongsCSV(file)
	if err != nil {
		http.Error(w, "Failed to parse PCO CSV: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Check for update mode
	updateMode := r.FormValue("update_mode") // "skip" or "update"
	if updateMode == "" {
		updateMode = "skip"
	}

	result, err := h.service.ImportPCOSongs(r.Context(), claims.TenantID, songs, updateMode)
	if err != nil {
		http.Error(w, "Failed to import songs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetImportStatus handles GET /api/import/status
func (h *Handler) GetImportStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	status, err := h.service.GetImportHistory(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get import status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
