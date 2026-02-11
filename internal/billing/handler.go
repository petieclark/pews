package billing

import (
	"encoding/json"
	"net/http"

	"github.com/petieclark/pews/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sub, err := h.service.GetSubscription(r.Context(), claims.TenantID)
	if err != nil {
		// If subscription doesn't exist, create a free one
		if err := h.service.EnsureSubscription(r.Context(), claims.TenantID); err != nil {
			http.Error(w, "Failed to create subscription: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Try to get again
		sub, err = h.service.GetSubscription(r.Context(), claims.TenantID)
		if err != nil {
			http.Error(w, "Failed to get subscription: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

func (h *Handler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	url, err := h.service.CreateCheckoutSession(r.Context(), claims.TenantID, claims.Email)
	if err != nil {
		http.Error(w, "Failed to create checkout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}

func (h *Handler) CreatePortal(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	url, err := h.service.CreatePortalSession(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to create portal session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}
