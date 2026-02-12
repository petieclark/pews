package module

import (
	"encoding/json"
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

type ModuleResponse struct {
	Module
	Enabled   bool `json:"enabled"`
	EnabledAt *string `json:"enabled_at,omitempty"`
}

func (h *Handler) ListModules(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tenantModules, err := h.service.GetTenantModules(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get modules: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Merge with available modules
	moduleMap := make(map[string]TenantModule)
	for _, tm := range tenantModules {
		moduleMap[tm.ModuleName] = tm
	}

	response := []ModuleResponse{}
	for _, m := range AvailableModules {
		resp := ModuleResponse{
			Module: m,
		}
		if tm, ok := moduleMap[m.Name]; ok {
			resp.Enabled = tm.Enabled
			if tm.EnabledAt != nil {
				enabledAt := tm.EnabledAt.Format("2006-01-02T15:04:05Z")
				resp.EnabledAt = &enabledAt
			}
		}
		response = append(response, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) EnableModule(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	moduleName := chi.URLParam(r, "name")
	if moduleName == "" {
		http.Error(w, "Module name required", http.StatusBadRequest)
		return
	}

	err := h.service.EnableModule(r.Context(), claims.TenantID, moduleName)
	if err != nil {
		http.Error(w, "Failed to enable module: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Module enabled successfully"})
}

func (h *Handler) DisableModule(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	moduleName := chi.URLParam(r, "name")
	if moduleName == "" {
		http.Error(w, "Module name required", http.StatusBadRequest)
		return
	}

	err := h.service.DisableModule(r.Context(), claims.TenantID, moduleName)
	if err != nil {
		http.Error(w, "Failed to disable module: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Module disabled successfully"})
}
