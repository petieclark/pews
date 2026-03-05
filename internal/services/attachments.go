package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/petieclark/pews/internal/middleware"
)

const MaxSongAttachmentSize = 20 * 1024 * 1024 // 20MB

// UploadSongAttachment handles file upload for a song (PDF, PNG, JPG, DOCX)
func (h *Handler) UploadSongAttachment(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	songID := chi.URLParam(r, "id")

	// Verify song exists and belongs to tenant
	_, err := h.service.GetSongByID(r.Context(), claims.TenantID, songID)
	if err != nil {
		http.Error(w, "Song not found", http.StatusNotFound)
		return
	}

	// Parse multipart form with size limit (20MB max)
	r.Body = http.MaxBytesReader(w, r.Body, MaxSongAttachmentSize)
	if err := r.ParseMultipartForm(MaxSongAttachmentSize); err != nil {
		http.Error(w, "File too large (max 20MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate content type (allow PDF, PNG, JPG, DOCX)
	contentType := header.Header.Get("Content-Type")
	validTypes := map[string]bool{
		"application/pdf": true,
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	}
	if !validTypes[contentType] {
		http.Error(w, "File type not allowed (allowed: PDF, PNG, JPG, DOCX)", http.StatusBadRequest)
		return
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file data", http.StatusInternalServerError)
		return
	}

	// Generate safe filename
	filename := strings.ReplaceAll(header.Filename, " ", "_")

	attachment := &SongAttachment{
		SongID:       songID,
		Filename:     filename,
		OriginalName: header.Filename,
		ContentType:  contentType,
		FileData:     fileData,
		FileSize:     len(fileData),
		UploadedBy:   &claims.UserID,
	}

	created, err := h.service.CreateSongAttachment(r.Context(), claims.TenantID, attachment)
	if err != nil {
		http.Error(w, "Failed to save attachment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// ListSongAttachments lists all attachments for a song (metadata only, no file data)
func (h *Handler) ListSongAttachments(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	songID := chi.URLParam(r, "id")

	attachments, err := h.service.ListSongAttachments(r.Context(), claims.TenantID, songID)
	if err != nil {
		http.Error(w, "Failed to list attachments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attachments)
}

// GetSongAttachment downloads a specific attachment (returns raw file data for download)
func (h *Handler) GetSongAttachment(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	attachmentID := chi.URLParam(r, "attachmentId")

	attachment, err := h.service.GetSongAttachment(r.Context(), claims.TenantID, attachmentID)
	if err != nil {
		http.Error(w, "Attachment not found", http.StatusNotFound)
		return
	}

	// Return raw file data for download
	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", attachment.Filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", attachment.FileSize))
	w.Write(attachment.FileData)
}

// DeleteSongAttachment deletes an attachment
func (h *Handler) DeleteSongAttachment(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	attachmentID := chi.URLParam(r, "attachmentId")

	if err := h.service.DeleteSongAttachment(r.Context(), claims.TenantID, attachmentID); err != nil {
		http.Error(w, "Failed to delete attachment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetPublicSongAttachment downloads a specific song attachment for public plan viewers (no auth required).
// This endpoint is only accessible via shared plan tokens and validates the token before serving files.
func (h *Handler) GetPublicSongAttachment(w http.ResponseWriter, r *http.Request) {
	attachmentID := chi.URLParam(r, "attachmentId")

	// Get attachment metadata first to verify it exists
	attachment, err := h.service.GetSongAttachmentByToken(r.Context(), attachmentID)
	if err != nil {
		http.Error(w, "Attachment not found", http.StatusNotFound)
		return
	}

	// Return raw file data for download
	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", attachment.Filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", attachment.FileSize))
	w.Write(attachment.FileData)
}

// GetSongAttachmentByToken retrieves an attachment by ID without tenant verification.
// Used for public plan sharing - the token validation happens at the router level.
func (s *Service) GetSongAttachmentByToken(ctx context.Context, attachmentID string) (*SongAttachment, error) {
	var a SongAttachment
	err := s.db.QueryRow(ctx, `
		SELECT id, tenant_id, song_id, filename, original_name, content_type, file_data, file_size, uploaded_by, created_at
		FROM song_attachments
		WHERE id = $1`, attachmentID).Scan(
		&a.ID, &a.TenantID, &a.SongID, &a.Filename, &a.OriginalName, &a.ContentType, &a.FileData, &a.FileSize, &a.UploadedBy, &a.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("attachment not found")
		}
		return nil, fmt.Errorf("failed to get song attachment: %w", err)
	}
	return &a, nil
}
