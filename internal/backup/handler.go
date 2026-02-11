package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type CreateBackupRequest struct {
	// Empty for now - all info comes from context
}

type RestoreBackupRequest struct {
	Confirmation string `json:"confirmation"`
}

// CreateBackup triggers a manual backup
func (h *Handler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get tenant slug from database
	var tenantSlug string
	err := h.service.pool.QueryRow(r.Context(), 
		"SELECT slug FROM tenants WHERE id = $1", claims.TenantID).Scan(&tenantSlug)
	if err != nil {
		http.Error(w, "Failed to get tenant info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	backup, err := h.service.CreateBackup(r.Context(), claims.TenantID, tenantSlug)
	if err != nil {
		http.Error(w, "Failed to create backup: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(backup)
}

// ListBackups lists all available backups for the tenant
func (h *Handler) ListBackups(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get tenant slug
	var tenantSlug string
	err := h.service.pool.QueryRow(r.Context(), 
		"SELECT slug FROM tenants WHERE id = $1", claims.TenantID).Scan(&tenantSlug)
	if err != nil {
		http.Error(w, "Failed to get tenant info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	backups, err := h.service.ListBackups(r.Context(), tenantSlug)
	if err != nil {
		http.Error(w, "Failed to list backups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := BackupListResponse{
		Backups: backups,
		Total:   len(backups),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RestoreBackup restores a backup
func (h *Handler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filename := chi.URLParam(r, "filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	var req RestoreBackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify confirmation
	if req.Confirmation != "RESTORE" {
		http.Error(w, "Confirmation must be 'RESTORE'", http.StatusBadRequest)
		return
	}

	// Get tenant slug
	var tenantSlug string
	err := h.service.pool.QueryRow(r.Context(), 
		"SELECT slug FROM tenants WHERE id = $1", claims.TenantID).Scan(&tenantSlug)
	if err != nil {
		http.Error(w, "Failed to get tenant info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.service.RestoreBackup(r.Context(), claims.TenantID, tenantSlug, filename); err != nil {
		http.Error(w, "Failed to restore backup: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Backup restored successfully",
	})
}

// DeleteBackup deletes a backup
func (h *Handler) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filename := chi.URLParam(r, "filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Get tenant slug
	var tenantSlug string
	err := h.service.pool.QueryRow(r.Context(), 
		"SELECT slug FROM tenants WHERE id = $1", claims.TenantID).Scan(&tenantSlug)
	if err != nil {
		http.Error(w, "Failed to get tenant info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.service.DeleteBackup(r.Context(), tenantSlug, filename); err != nil {
		http.Error(w, "Failed to delete backup: "+err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DownloadBackup downloads a backup file
func (h *Handler) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filename := chi.URLParam(r, "filename")
	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Get tenant slug
	var tenantSlug string
	err := h.service.pool.QueryRow(r.Context(), 
		"SELECT slug FROM tenants WHERE id = $1", claims.TenantID).Scan(&tenantSlug)
	if err != nil {
		http.Error(w, "Failed to get tenant info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	reader, err := h.service.DownloadBackup(r.Context(), tenantSlug, filename)
	if err != nil {
		http.Error(w, "Failed to download backup: "+err.Error(), http.StatusNotFound)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	
	// Copy file content to response
	if _, err := io.Copy(w, reader); err != nil {
		return
	}
}
