package ccli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/petieclark/pews/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetReport generates a CCLI usage report for a date range.
// GET /api/ccli/report?start=2026-01-01&end=2026-03-31
func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromClaims(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		http.Error(w, `{"error":"start and end query parameters are required"}`, http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		http.Error(w, `{"error":"invalid start date format, use YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		http.Error(w, `{"error":"invalid end date format, use YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	report, err := h.service.GenerateCCLIReport(r.Context(), tenantID, startDate, endDate)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// DownloadReport generates and downloads a CCLI report as CSV.
// GET /api/ccli/report/download?start=2026-01-01&end=2026-03-31
func (h *Handler) DownloadReport(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromClaims(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	if startStr == "" || endStr == "" {
		http.Error(w, `{"error":"start and end query parameters are required"}`, http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		http.Error(w, `{"error":"invalid start date format"}`, http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		http.Error(w, `{"error":"invalid end date format"}`, http.StatusBadRequest)
		return
	}

	report, err := h.service.GenerateCCLIReport(r.Context(), tenantID, startDate, endDate)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("ccli-report-%s-to-%s.csv", startStr, endStr)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// CCLI report header
	writer.Write([]string{"CCLI License Number", report.LicenseNumber})
	writer.Write([]string{"Reporting Period", report.Period})
	writer.Write([]string{""})

	// Column headers per CCLI reporting requirements
	writer.Write([]string{"Song Title", "CCLI Song Number", "Author/Composer", "Times Used", "Last Used"})

	for _, song := range report.Songs {
		writer.Write([]string{
			song.Title,
			song.CCLINumber,
			song.Artist,
			fmt.Sprintf("%d", song.TimesUsed),
			song.LastUsed,
		})
	}

	// Summary
	writer.Write([]string{""})
	writer.Write([]string{"Total Songs", fmt.Sprintf("%d", report.TotalSongs)})
	writer.Write([]string{"Total Uses", fmt.Sprintf("%d", report.TotalUses)})
}

// GetStats returns quick CCLI statistics.
// GET /api/ccli/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromClaims(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := h.service.GetStats(r.Context(), tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// SaveSettings saves CCLI license number and reporting preferences.
// POST /api/ccli/settings
func (h *Handler) SaveSettings(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromClaims(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		LicenseNumber   string `json:"license_number"`
		AutoReport      bool   `json:"auto_report"`
		ReportFrequency string `json:"report_frequency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	settings, err := h.service.SaveSettings(r.Context(), tenantID, req.LicenseNumber, req.AutoReport, req.ReportFrequency)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// GetSettings retrieves current CCLI settings.
// GET /api/ccli/settings
func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromClaims(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	settings, err := h.service.GetSettings(r.Context(), tenantID)
	if err != nil {
		// Return empty settings if not configured yet
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CCLISettings{ReportFrequency: "quarterly"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}
