package qr

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/petieclark/pews/internal/middleware"
	qrcode "github.com/skip2/go-qrcode"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GenerateCheckinQR generates a QR code for a check-in station
func (h *Handler) GenerateCheckinQR(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("station")
	if stationID == "" {
		http.Error(w, "station parameter is required", http.StatusBadRequest)
		return
	}

	size := h.getQRSize(r)
	
	// Get tenant subdomain from context
	tenantSubdomain := middleware.GetTenantID(r.Context())
	if tenantSubdomain == "" {
		http.Error(w, "tenant not found", http.StatusBadRequest)
		return
	}

	url := h.service.BuildCheckinURL(tenantSubdomain, stationID)
	png, err := qrcode.Encode(url, qrcode.Medium, size)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"checkin-%s.png\"", stationID))
	w.Write(png)
}

// GenerateConnectQR generates a QR code for connection cards
func (h *Handler) GenerateConnectQR(w http.ResponseWriter, r *http.Request) {
	size := h.getQRSize(r)
	
	tenantSubdomain := middleware.GetTenantID(r.Context())
	if tenantSubdomain == "" {
		http.Error(w, "tenant not found", http.StatusBadRequest)
		return
	}

	url := h.service.BuildConnectURL(tenantSubdomain)
	png, err := qrcode.Encode(url, qrcode.Medium, size)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "inline; filename=\"connect.png\"")
	w.Write(png)
}

// GenerateGiveQR generates a QR code for giving
func (h *Handler) GenerateGiveQR(w http.ResponseWriter, r *http.Request) {
	size := h.getQRSize(r)
	
	tenantSubdomain := middleware.GetTenantID(r.Context())
	if tenantSubdomain == "" {
		http.Error(w, "tenant not found", http.StatusBadRequest)
		return
	}

	url := h.service.BuildGiveURL(tenantSubdomain)
	png, err := qrcode.Encode(url, qrcode.Medium, size)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "inline; filename=\"give.png\"")
	w.Write(png)
}

// GeneratePrayerQR generates a QR code for prayer submission
func (h *Handler) GeneratePrayerQR(w http.ResponseWriter, r *http.Request) {
	size := h.getQRSize(r)
	
	tenantSubdomain := middleware.GetTenantID(r.Context())
	if tenantSubdomain == "" {
		http.Error(w, "tenant not found", http.StatusBadRequest)
		return
	}

	url := h.service.BuildPrayerURL(tenantSubdomain)
	png, err := qrcode.Encode(url, qrcode.Medium, size)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "inline; filename=\"prayer.png\"")
	w.Write(png)
}

// GenerateCustomQR generates a QR code for a custom URL (authenticated)
func (h *Handler) GenerateCustomQR(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	customURL := r.URL.Query().Get("url")
	if customURL == "" {
		http.Error(w, "url parameter is required", http.StatusBadRequest)
		return
	}

	size := h.getQRSize(r)
	
	png, err := qrcode.Encode(customURL, qrcode.Medium, size)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "inline; filename=\"custom.png\"")
	w.Write(png)
}

// getQRSize extracts size from query parameter or returns default
func (h *Handler) getQRSize(r *http.Request) int {
	sizeStr := r.URL.Query().Get("size")
	if sizeStr == "" {
		return 300 // default size
	}
	
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 100 || size > 1000 {
		return 300 // default if invalid
	}
	
	return size
}
