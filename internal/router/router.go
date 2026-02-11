package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/auth"
	"github.com/petieclark/pews/internal/billing"
	"github.com/petieclark/pews/internal/giving"
	"github.com/petieclark/pews/internal/middleware"
	"github.com/petieclark/pews/internal/module"
	"github.com/petieclark/pews/internal/people"
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
	peopleHandler *people.Handler,
	givingHandler *giving.Handler,
	webhookSecret string,
	givingWebhookSecret string,
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
	r.Post("/api/giving/webhook", givingHandler.HandleWebhook(givingWebhookSecret))

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

		// People
		r.Get("/api/people", peopleHandler.ListPeople)
		r.Post("/api/people", peopleHandler.CreatePerson)
		r.Get("/api/people/{id}", peopleHandler.GetPerson)
		r.Put("/api/people/{id}", peopleHandler.UpdatePerson)
		r.Delete("/api/people/{id}", peopleHandler.DeletePerson)
		r.Post("/api/people/{id}/tags", peopleHandler.AddTagToPerson)
		r.Delete("/api/people/{id}/tags/{tagId}", peopleHandler.RemoveTagFromPerson)

		// Tags
		r.Get("/api/tags", peopleHandler.ListTags)
		r.Post("/api/tags", peopleHandler.CreateTag)

		// Households
		r.Get("/api/households", peopleHandler.ListHouseholds)
		r.Post("/api/households", peopleHandler.CreateHousehold)
		r.Put("/api/households/{id}", peopleHandler.UpdateHousehold)
		r.Post("/api/households/{id}/members", peopleHandler.AddMemberToHousehold)
		r.Delete("/api/households/{id}/members/{personId}", peopleHandler.RemoveMemberFromHousehold)

		// Giving - Funds
		r.Get("/api/giving/funds", givingHandler.ListFunds)
		r.Post("/api/giving/funds", givingHandler.CreateFund)
		r.Put("/api/giving/funds/{id}", givingHandler.UpdateFund)

		// Giving - Donations
		r.Get("/api/giving/donations", givingHandler.ListDonations)
		r.Post("/api/giving/donations", givingHandler.CreateDonation)
		r.Get("/api/giving/donations/{id}", givingHandler.GetDonation)

		// Giving - Stats & Reports
		r.Get("/api/giving/stats", givingHandler.GetStats)
		r.Get("/api/giving/person/{personId}", givingHandler.GetPersonGivingHistory)
		r.Get("/api/giving/recurring", givingHandler.ListRecurringDonations)

		// Giving - Statements
		r.Post("/api/giving/statements/{year}", givingHandler.GenerateStatement)

		// Giving - Stripe Connect
		r.Post("/api/giving/connect/onboard", givingHandler.CreateConnectOnboard)
		r.Get("/api/giving/connect/status", givingHandler.GetConnectStatus)
		r.Get("/api/giving/connect/return", givingHandler.HandleConnectReturn)
		r.Get("/api/giving/connect/refresh", givingHandler.HandleConnectRefresh)
		r.Post("/api/giving/checkout", givingHandler.CreateCheckout)
	})

	return &Router{r}
}
