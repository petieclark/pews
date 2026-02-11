package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/petieclark/pews/internal/tenant"
)

type Handler struct {
	authService    *Service
	tenantService  *tenant.Service
	billingService BillingSubscriptionCreator
}

// BillingSubscriptionCreator allows auth handler to create subscriptions on registration
type BillingSubscriptionCreator interface {
	EnsureSubscription(ctx context.Context, tenantID string) error
}

func NewHandler(authService *Service, tenantService *tenant.Service, billingCreator BillingSubscriptionCreator) *Handler {
	return &Handler{
		authService:    authService,
		tenantService:  tenantService,
		billingService: billingCreator,
	}
}

type RegisterRequest struct {
	TenantName string `json:"tenant_name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type LoginRequest struct {
	TenantSlug string `json:"tenant_slug"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type AuthResponse struct {
	Token    string `json:"token"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TenantName == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create tenant
	tenant, err := h.tenantService.CreateTenant(r.Context(), req.TenantName)
	if err != nil {
		http.Error(w, "Failed to create tenant: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create free subscription for tenant
	if h.billingService != nil {
		if err := h.billingService.EnsureSubscription(r.Context(), tenant.ID); err != nil {
			log.Printf("Warning: failed to create subscription for tenant %s: %v", tenant.ID, err)
		}
	}

	// Create admin user
	user, err := h.authService.CreateUser(r.Context(), tenant.ID, req.Email, req.Password, "admin")
	if err != nil {
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(user.ID, user.TenantID, user.Email, user.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token:    token,
		TenantID: tenant.ID,
		Email:    user.Email,
		Role:     user.Role,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TenantSlug == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Get tenant by slug
	tenant, err := h.tenantService.GetTenantBySlug(r.Context(), req.TenantSlug)
	if err != nil {
		http.Error(w, "Invalid tenant", http.StatusUnauthorized)
		return
	}

	// Get user by email
	user, err := h.authService.GetUserByEmail(r.Context(), tenant.ID, req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := h.authService.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(user.ID, user.TenantID, user.Email, user.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		Token:    token,
		TenantID: tenant.ID,
		Email:    user.Email,
		Role:     user.Role,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// In a JWT-based system, logout is client-side (delete token)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}
