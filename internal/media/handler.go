package media

import (
	"encoding/json"
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

type UploadResponse struct {
	File *MediaFile `json:"file"`
}

type UpdateFileRequest struct {
	Folder *string  `json:"folder,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (50MB max)
	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	folder := r.FormValue("folder")
	tagsStr := r.FormValue("tags")
	tags := []string{}
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		// Trim whitespace from tags
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	mediaFile, err := h.service.UploadFile(r.Context(), claims.TenantID, claims.UserID, file, header, folder, tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UploadResponse{File: mediaFile})
}

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	mediaTypeStr := r.URL.Query().Get("type")
	mediaType := MediaTypeAll
	if mediaTypeStr != "" {
		mediaType = MediaType(mediaTypeStr)
	}

	var folder *string
	if folderParam := r.URL.Query().Get("folder"); folderParam != "" {
		folder = &folderParam
	}

	tags := []string{}
	if tagsParam := r.URL.Query().Get("tags"); tagsParam != "" {
		tags = strings.Split(tagsParam, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	files, err := h.service.ListFiles(r.Context(), claims.TenantID, mediaType, folder, tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fileID := chi.URLParam(r, "id")
	file, err := h.service.GetFile(r.Context(), claims.TenantID, fileID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(file)
}

func (h *Handler) UpdateFile(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fileID := chi.URLParam(r, "id")

	var req UpdateFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	file, err := h.service.UpdateFile(r.Context(), claims.TenantID, fileID, req.Folder, req.Tags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(file)
}

func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fileID := chi.URLParam(r, "id")
	if err := h.service.DeleteFile(r.Context(), claims.TenantID, fileID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListFolders(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	folders, err := h.service.ListFolders(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if folders == nil {
		folders = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(folders)
}
