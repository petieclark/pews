package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/middleware"
)

const (
	maxUploadSize = 10 << 20 // 10 MB
)

// UploadSongAttachment handles PDF upload for a song
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

	// Parse multipart form with size limit
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large (max 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if contentType != "application/pdf" {
		http.Error(w, "Only PDF files are allowed", http.StatusBadRequest)
		return
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file data", http.StatusInternalServerError)
		return
	}

	// Generate filename
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

// ListSongAttachments lists all attachments for a song
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

// GetSongAttachment downloads a specific attachment
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
