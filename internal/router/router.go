package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/auth"
	"github.com/petieclark/pews/internal/billing"
	"github.com/petieclark/pews/internal/middleware"
	"github.com/petieclark/pews/internal/module"
	"github.com/petieclark/pews/internal/tenant"
)

type Router struct {
	chi.Router
}

func New(
	authHandler *auth.Handler,
	authService *auth.Service,
	tenantHandler *tenant.Handler,
	moduleHandler *module.Handler,
	billingHandler *billing.Handler,
	webhookSecret string,
	frontendURL string,
) *Router {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logging)
	r.Use(middleware.CORS(frontendURL))
	r.Use(middleware.TenantExtractor)

	// Public routes
	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/billing/webhook", billingHandler.HandleWebhook(webhookSecret))

	// Health check
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authService.Middleware)

		// Auth
		r.Post("/api/auth/logout", authHandler.Logout)

		// Tenant
		r.Get("/api/tenant", tenantHandler.GetTenant)
		r.Put("/api/tenant", tenantHandler.UpdateTenant)

		// Modules
		r.Get("/api/tenant/modules", moduleHandler.ListModules)
		r.Post("/api/tenant/modules/{name}/enable", moduleHandler.EnableModule)
		r.Post("/api/tenant/modules/{name}/disable", moduleHandler.DisableModule)

		// Billing
		r.Get("/api/billing/subscription", billingHandler.GetSubscription)
		r.Post("/api/billing/checkout", billingHandler.CreateCheckout)
		r.Post("/api/billing/portal", billingHandler.CreatePortal)
	})

	return &Router{r}
}
