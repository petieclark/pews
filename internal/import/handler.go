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
