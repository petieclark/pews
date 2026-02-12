package sermons

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type CreateSermonRequest struct {
	ServiceID            *string `json:"service_id,omitempty"`
	Title                string  `json:"title"`
	Speaker              string  `json:"speaker"`
	SermonDate           string  `json:"sermon_date"`
	ScriptureReference   string  `json:"scripture_reference,omitempty"`
	NotesText            string  `json:"notes_text,omitempty"`
	AudioURL             string  `json:"audio_url,omitempty"`
	VideoURL             string  `json:"video_url,omitempty"`
	SeriesName           string  `json:"series_name,omitempty"`
	AudioDurationSeconds *int    `json:"audio_duration_seconds,omitempty"`
	Published            bool    `json:"published"`
}

func (h *Handler) ListSermons(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filters := SermonFilters{
		Query:    r.URL.Query().Get("q"),
		Series:   r.URL.Query().Get("series"),
		Speaker:  r.URL.Query().Get("speaker"),
		DateFrom: r.URL.Query().Get("date_from"),
		DateTo:   r.URL.Query().Get("date_to"),
	}

	if pub := r.URL.Query().Get("published"); pub != "" {
		published := pub == "true"
		filters.Published = &published
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filters.Limit = l
		}
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filters.Offset = o
		}
	}

	sermons, err := h.service.ListSermons(r.Context(), claims.TenantID, filters)
	if err != nil {
		http.Error(w, "Failed to list sermons: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sermons)
}

func (h *Handler) GetSermon(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sermonID := chi.URLParam(r, "id")
	sermon, err := h.service.GetSermon(r.Context(), claims.TenantID, sermonID)
	if err != nil {
		http.Error(w, "Sermon not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sermon)
}

func (h *Handler) CreateSermon(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateSermonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Speaker == "" || req.SermonDate == "" {
		http.Error(w, "Title, speaker, and sermon_date are required", http.StatusBadRequest)
		return
	}

	sermonDate, err := time.Parse("2006-01-02", req.SermonDate)
	if err != nil {
		http.Error(w, "Invalid date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	sermon := &Sermon{
		ServiceID:            req.ServiceID,
		Title:                req.Title,
		Speaker:              req.Speaker,
		SermonDate:           sermonDate,
		ScriptureReference:   req.ScriptureReference,
		NotesText:            req.NotesText,
		AudioURL:             req.AudioURL,
		VideoURL:             req.VideoURL,
		SeriesName:           req.SeriesName,
		AudioDurationSeconds: req.AudioDurationSeconds,
		Published:            req.Published,
	}

	createdSermon, err := h.service.CreateSermon(r.Context(), claims.TenantID, sermon)
	if err != nil {
		http.Error(w, "Failed to create sermon: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdSermon)
}

func (h *Handler) UpdateSermon(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sermonID := chi.URLParam(r, "id")

	var req CreateSermonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	sermonDate, err := time.Parse("2006-01-02", req.SermonDate)
	if err != nil {
		http.Error(w, "Invalid date format (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	sermon := &Sermon{
		ID:                   sermonID,
		ServiceID:            req.ServiceID,
		Title:                req.Title,
		Speaker:              req.Speaker,
		SermonDate:           sermonDate,
		ScriptureReference:   req.ScriptureReference,
		NotesText:            req.NotesText,
		AudioURL:             req.AudioURL,
		VideoURL:             req.VideoURL,
		SeriesName:           req.SeriesName,
		AudioDurationSeconds: req.AudioDurationSeconds,
		Published:            req.Published,
	}

	if err := h.service.UpdateSermon(r.Context(), claims.TenantID, sermon); err != nil {
		http.Error(w, "Failed to update sermon: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sermon)
}

func (h *Handler) DeleteSermon(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sermonID := chi.URLParam(r, "id")
	if err := h.service.DeleteSermon(r.Context(), claims.TenantID, sermonID); err != nil {
		http.Error(w, "Failed to delete sermon: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PublishSermon(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sermonID := chi.URLParam(r, "id")

	var req struct {
		Published bool `json:"published"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to toggling to published
		req.Published = true
	}

	if err := h.service.SetPublished(r.Context(), claims.TenantID, sermonID, req.Published); err != nil {
		http.Error(w, "Failed to update publish status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"published": req.Published})
}

func (h *Handler) GetPublicSermons(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		if ctx := r.Context().Value("tenant_id"); ctx != nil {
			tenantID = ctx.(string)
		}
	}

	if tenantID == "" {
		http.Error(w, "Tenant ID required", http.StatusBadRequest)
		return
	}

	filters := SermonFilters{
		Series:   r.URL.Query().Get("series"),
		Speaker:  r.URL.Query().Get("speaker"),
		DateFrom: r.URL.Query().Get("date_from"),
		DateTo:   r.URL.Query().Get("date_to"),
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filters.Limit = l
		}
	} else {
		filters.Limit = 50
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filters.Offset = o
		}
	}

	sermons, err := h.service.GetPublicSermons(r.Context(), tenantID, filters)
	if err != nil {
		http.Error(w, "Failed to retrieve sermons: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sermons)
}

func (h *Handler) GetPodcastFeed(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		if ctx := r.Context().Value("tenant_id"); ctx != nil {
			tenantID = ctx.(string)
		}
	}

	if tenantID == "" {
		http.Error(w, "Tenant ID required", http.StatusBadRequest)
		return
	}

	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	baseURL := scheme + "://" + r.Host

	feedXML, err := h.service.GeneratePodcastFeed(r.Context(), tenantID, baseURL)
	if err != nil {
		http.Error(w, "Failed to generate podcast feed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Write(feedXML)
}
