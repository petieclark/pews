package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/auth"
	"github.com/petieclark/pews/internal/billing"
	"github.com/petieclark/pews/internal/checkins"
	"github.com/petieclark/pews/internal/communication"
	"github.com/petieclark/pews/internal/giving"
	"github.com/petieclark/pews/internal/groups"
	"github.com/petieclark/pews/internal/middleware"
	"github.com/petieclark/pews/internal/module"
	"github.com/petieclark/pews/internal/people"
	"github.com/petieclark/pews/internal/services"
	"github.com/petieclark/pews/internal/sms"
	"github.com/petieclark/pews/internal/streaming"
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
	groupsHandler *groups.Handler,
	servicesHandler *services.Handler,
	givingHandler *giving.Handler,
	streamingHandler *streaming.Handler,
	communicationHandler *communication.Handler,
	checkinsHandler *checkins.Handler,
	smsHandler *sms.Handler,
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

	// Public streaming routes (no auth required)
	r.Get("/api/streaming/watch/{id}", streamingHandler.GetWatchStream)
	r.Get("/api/streaming/{id}/chat", streamingHandler.GetChatMessages)
	r.Post("/api/streaming/{id}/chat", streamingHandler.SendChatMessage)
	r.Post("/api/streaming/{id}/join", streamingHandler.JoinStream)
	r.Post("/api/streaming/{id}/leave", streamingHandler.LeaveStream)

	// Public communication route - connection card submission (no auth required)
	r.Post("/api/communication/cards", communicationHandler.SubmitConnectionCard)

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

		// Groups
		r.Get("/api/groups", groupsHandler.ListGroups)
		r.Post("/api/groups", groupsHandler.CreateGroup)
		r.Get("/api/groups/{id}", groupsHandler.GetGroup)
		r.Put("/api/groups/{id}", groupsHandler.UpdateGroup)
		r.Delete("/api/groups/{id}", groupsHandler.DeleteGroup)
		r.Get("/api/groups/{id}/members", groupsHandler.GetGroupMembers)
		r.Post("/api/groups/{id}/members", groupsHandler.AddMemberToGroup)
		r.Put("/api/groups/{id}/members/{memberId}", groupsHandler.UpdateMemberRole)
		r.Delete("/api/groups/{id}/members/{memberId}", groupsHandler.RemoveMemberFromGroup)
		r.Get("/api/groups/person/{personId}", groupsHandler.GetPersonGroups)

		// Services - Service Types
		r.Get("/api/services/types", servicesHandler.ListServiceTypes)
		r.Post("/api/services/types", servicesHandler.CreateServiceType)
		r.Put("/api/services/types/{id}", servicesHandler.UpdateServiceType)

		// Services - Services
		r.Get("/api/services", servicesHandler.ListServices)
		r.Get("/api/services/upcoming", servicesHandler.GetUpcomingServices)
		r.Post("/api/services", servicesHandler.CreateService)
		r.Get("/api/services/{id}", servicesHandler.GetService)
		r.Put("/api/services/{id}", servicesHandler.UpdateService)
		r.Delete("/api/services/{id}", servicesHandler.DeleteService)

		// Services - Service Items
		r.Get("/api/services/{id}/items", servicesHandler.GetServiceItems)
		r.Post("/api/services/{id}/items", servicesHandler.AddServiceItem)
		r.Put("/api/services/{id}/items/{itemId}", servicesHandler.UpdateServiceItem)
		r.Delete("/api/services/{id}/items/{itemId}", servicesHandler.DeleteServiceItem)

		// Services - Service Teams
		r.Get("/api/services/{id}/team", servicesHandler.GetServiceTeam)
		r.Post("/api/services/{id}/team", servicesHandler.AddServiceTeamMember)
		r.Put("/api/services/{id}/team/{teamId}", servicesHandler.UpdateServiceTeamMember)
		r.Delete("/api/services/{id}/team/{teamId}", servicesHandler.DeleteServiceTeamMember)

		// Services - Songs
		r.Get("/api/services/songs", servicesHandler.ListSongs)
		r.Post("/api/services/songs", servicesHandler.CreateSong)
		r.Put("/api/services/songs/{id}", servicesHandler.UpdateSong)
		r.Delete("/api/services/songs/{id}", servicesHandler.DeleteSong)

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

		// Streaming - Streams
		r.Get("/api/streaming", streamingHandler.ListStreams)
		r.Post("/api/streaming", streamingHandler.CreateStream)
		r.Get("/api/streaming/live", streamingHandler.GetLiveStream)
		r.Get("/api/streaming/{id}", streamingHandler.GetStream)
		r.Put("/api/streaming/{id}", streamingHandler.UpdateStream)
		r.Delete("/api/streaming/{id}", streamingHandler.DeleteStream)
		r.Post("/api/streaming/{id}/go-live", streamingHandler.GoLive)
		r.Post("/api/streaming/{id}/end", streamingHandler.EndStream)

		// Streaming - Chat (admin actions)
		r.Put("/api/streaming/{id}/chat/{msgId}/pin", streamingHandler.PinChatMessage)
		r.Delete("/api/streaming/{id}/chat/{msgId}", streamingHandler.DeleteChatMessage)

		// Streaming - Viewers
		r.Get("/api/streaming/{id}/viewers", streamingHandler.GetViewers)

		// Streaming - Notes
		r.Get("/api/streaming/{id}/notes", streamingHandler.GetStreamNotes)
		r.Post("/api/streaming/{id}/notes", streamingHandler.SaveStreamNotes)

		// Communication - Templates
		r.Get("/api/communication/templates", communicationHandler.ListTemplates)
		r.Post("/api/communication/templates", communicationHandler.CreateTemplate)
		r.Put("/api/communication/templates/{id}", communicationHandler.UpdateTemplate)
		r.Delete("/api/communication/templates/{id}", communicationHandler.DeleteTemplate)

		// Communication - Campaigns
		r.Get("/api/communication/campaigns", communicationHandler.ListCampaigns)
		r.Post("/api/communication/campaigns", communicationHandler.CreateCampaign)
		r.Get("/api/communication/campaigns/{id}", communicationHandler.GetCampaign)
		r.Put("/api/communication/campaigns/{id}", communicationHandler.UpdateCampaign)
		r.Post("/api/communication/campaigns/{id}/send", communicationHandler.SendCampaign)
		r.Get("/api/communication/campaigns/{id}/recipients", communicationHandler.GetCampaignRecipients)

		// Communication - Journeys
		r.Get("/api/communication/journeys", communicationHandler.ListJourneys)
		r.Post("/api/communication/journeys", communicationHandler.CreateJourney)
		r.Get("/api/communication/journeys/{id}", communicationHandler.GetJourney)
		r.Put("/api/communication/journeys/{id}", communicationHandler.UpdateJourney)
		r.Delete("/api/communication/journeys/{id}", communicationHandler.DeleteJourney)
		r.Post("/api/communication/journeys/{id}/steps", communicationHandler.AddJourneyStep)
		r.Put("/api/communication/journeys/{id}/steps/{stepId}", communicationHandler.UpdateJourneyStep)
		r.Delete("/api/communication/journeys/{id}/steps/{stepId}", communicationHandler.DeleteJourneyStep)
		r.Post("/api/communication/journeys/{id}/enroll", communicationHandler.EnrollInJourney)
		r.Get("/api/communication/journeys/{id}/enrollments", communicationHandler.GetJourneyEnrollments)

		// Communication - Connection Cards (authenticated endpoints)
		r.Get("/api/communication/cards", communicationHandler.ListConnectionCards)
		r.Get("/api/communication/cards/{id}", communicationHandler.GetConnectionCard)
		r.Post("/api/communication/cards/{id}/process", communicationHandler.ProcessConnectionCard)

		// Communication - Stats
		r.Get("/api/communication/stats", communicationHandler.GetStats)

		// SMS - Sending
		r.Post("/api/sms/send", smsHandler.SendSMS)
		r.Post("/api/sms/bulk", smsHandler.SendBulkSMS)

		// SMS - History
		r.Get("/api/sms/history", smsHandler.GetHistory)

		// SMS - Templates
		r.Get("/api/sms/templates", smsHandler.ListTemplates)
		r.Post("/api/sms/templates", smsHandler.CreateTemplate)
		r.Put("/api/sms/templates/{id}", smsHandler.UpdateTemplate)
		r.Delete("/api/sms/templates/{id}", smsHandler.DeleteTemplate)

		// SMS - Settings
		r.Get("/api/sms/settings", smsHandler.GetSettings)
		r.Post("/api/sms/settings", smsHandler.SaveSettings)
		r.Post("/api/sms/settings/test", smsHandler.TestConnection)

		// Check-ins - Events
		r.Get("/api/checkins/events", checkinsHandler.ListEvents)
		r.Post("/api/checkins/events", checkinsHandler.CreateEvent)
		r.Get("/api/checkins/events/{id}", checkinsHandler.GetEvent)
		r.Put("/api/checkins/events/{id}", checkinsHandler.UpdateEvent)
		r.Post("/api/checkins/events/{id}/checkin", checkinsHandler.CheckIn)
		r.Post("/api/checkins/events/{id}/checkout", checkinsHandler.CheckOut)
		r.Get("/api/checkins/events/{id}/attendees", checkinsHandler.GetAttendees)

		// Check-ins - Stations
		r.Get("/api/checkins/stations", checkinsHandler.ListStations)
		r.Post("/api/checkins/stations", checkinsHandler.CreateStation)
		r.Put("/api/checkins/stations/{id}", checkinsHandler.UpdateStation)

		// Check-ins - Person
		r.Get("/api/checkins/person/{personId}/history", checkinsHandler.GetPersonHistory)
		r.Get("/api/checkins/person/{personId}/alerts", checkinsHandler.GetAlerts)
		r.Post("/api/checkins/person/{personId}/alerts", checkinsHandler.CreateAlert)
		r.Delete("/api/checkins/person/{personId}/alerts/{alertId}", checkinsHandler.DeleteAlert)
		r.Get("/api/checkins/person/{personId}/pickups", checkinsHandler.GetPickups)
		r.Post("/api/checkins/person/{personId}/pickups", checkinsHandler.CreatePickup)
		r.Delete("/api/checkins/person/{personId}/pickups/{pickupId}", checkinsHandler.DeletePickup)

		// Check-ins - Stats & Search
		r.Get("/api/checkins/stats", checkinsHandler.GetStats)
		r.Get("/api/checkins/search", checkinsHandler.SearchPeople)
	})

	return &Router{r}
}
