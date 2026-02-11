package giving

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/activity"
	"github.com/petieclark/pews/internal/middleware"
	"github.com/petieclark/pews/internal/notification"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

type Handler struct {
	service             *Service
	stripeService       *StripeService
	activityService     *activity.Service
	notificationService *notification.Service
}

func NewHandler(service *Service, stripeService *StripeService, activityService *activity.Service) *Handler {
	// Create notification service with same db pool
	return &Handler{
		service:             service,
		stripeService:       stripeService,
		activityService:     activityService,
		notificationService: notification.NewService(service.GetDB()),
	}
}

// Funds

type CreateFundRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
}

type UpdateFundRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"is_default"`
	IsActive    bool   `json:"is_active"`
}

func (h *Handler) ListFunds(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	funds, err := h.service.ListFunds(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to list funds: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(funds)
}

func (h *Handler) CreateFund(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	var req CreateFundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fund, err := h.service.CreateFund(r.Context(), claims.TenantID, req.Name, req.Description, req.IsDefault)
	if err != nil {
		http.Error(w, "Failed to create fund: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fund)
}

func (h *Handler) UpdateFund(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	fundID := chi.URLParam(r, "id")
	if fundID == "" {
		http.Error(w, "Fund ID is required", http.StatusBadRequest)
		return
	}

	var req UpdateFundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fund, err := h.service.UpdateFund(r.Context(), claims.TenantID, fundID, req.Name, req.Description, req.IsDefault, req.IsActive)
	if err != nil {
		http.Error(w, "Failed to update fund: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fund)
}

// Donations

type CreateDonationRequest struct {
	PersonID      *string `json:"person_id"`
	FundID        string  `json:"fund_id"`
	AmountCents   int     `json:"amount_cents"`
	PaymentMethod string  `json:"payment_method"`
	Memo          string  `json:"memo"`
	DonatedAt     string  `json:"donated_at"`
}

type ListDonationsResponse struct {
	Donations []Donation `json:"donations"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	PerPage   int        `json:"per_page"`
}

func (h *Handler) ListDonations(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	personID := r.URL.Query().Get("person_id")
	fundID := r.URL.Query().Get("fund_id")
	fromDate := r.URL.Query().Get("from")
	toDate := r.URL.Query().Get("to")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	donations, total, err := h.service.ListDonations(r.Context(), claims.TenantID, personID, fundID, fromDate, toDate, perPage, offset)
	if err != nil {
		http.Error(w, "Failed to list donations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := ListDonationsResponse{
		Donations: donations,
		Total:     total,
		Page:      page,
		PerPage:   perPage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetDonation(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	donationID := chi.URLParam(r, "id")
	if donationID == "" {
		http.Error(w, "Donation ID is required", http.StatusBadRequest)
		return
	}

	donation, err := h.service.GetDonation(r.Context(), claims.TenantID, donationID)
	if err != nil {
		http.Error(w, "Donation not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(donation)
}

func (h *Handler) CreateDonation(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	var req CreateDonationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse donated_at
	var donatedAt time.Time
	if req.DonatedAt != "" {
		var err error
		donatedAt, err = time.Parse(time.RFC3339, req.DonatedAt)
		if err != nil {
			http.Error(w, "Invalid donated_at format", http.StatusBadRequest)
			return
		}
	} else {
		donatedAt = time.Now()
	}

	donation, err := h.service.CreateDonation(r.Context(), claims.TenantID, req.PersonID, req.FundID, req.AmountCents, req.PaymentMethod, req.Memo, donatedAt)
	if err != nil {
		http.Error(w, "Failed to create donation: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create notification for all admins
	notifTitle := "New Donation Received"
	notifMessage := fmt.Sprintf("A donation of $%.2f was received for %s", float64(donation.AmountCents)/100.0, donation.FundName)
	link := fmt.Sprintf("/giving/donations/%s", donation.ID)
	_ = h.notificationService.CreateForAllAdmins(r.Context(), claims.TenantID, notifTitle, notifMessage, notification.TypeSuccess, &link)

	// Log activity
	ipAddress := r.RemoteAddr
	details := map[string]interface{}{
		"amount":         float64(donation.AmountCents) / 100.0,
		"fund":           donation.FundName,
		"payment_method": req.PaymentMethod,
	}
	h.activityService.LogActivity(r.Context(), claims.TenantID, "donation.recorded", "giving", &claims.UserID, &donation.ID, &ipAddress, details)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(donation)
}

// Stats

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := h.service.GetGivingStats(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Person giving history

func (h *Handler) GetPersonGivingHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	personID := chi.URLParam(r, "personId")
	if personID == "" {
		http.Error(w, "Person ID is required", http.StatusBadRequest)
		return
	}

	donations, total, err := h.service.GetPersonGivingHistory(r.Context(), claims.TenantID, personID)
	if err != nil {
		http.Error(w, "Failed to get giving history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"donations":   donations,
		"total_cents": total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Statements

type GenerateStatementRequest struct {
	PersonID string `json:"person_id"`
}

func (h *Handler) GenerateStatement(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	yearStr := chi.URLParam(r, "year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	var req GenerateStatementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	statement, err := h.service.GenerateGivingStatement(r.Context(), claims.TenantID, req.PersonID, year)
	if err != nil {
		http.Error(w, "Failed to generate statement: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(statement)
}

// GenerateStatementPDF generates a PDF tax statement for a specific person and year
func (h *Handler) GenerateStatementPDF(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	yearStr := chi.URLParam(r, "year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	personID := chi.URLParam(r, "personId")
	if personID == "" {
		http.Error(w, "Person ID is required", http.StatusBadRequest)
		return
	}

	// Get tenant info
	tenantInfo, err := h.service.GetTenantInfoForStatement(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get tenant info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get person info
	personInfo, err := h.service.GetPersonInfoForStatement(r.Context(), claims.TenantID, personID)
	if err != nil {
		http.Error(w, "Failed to get person info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get donations for the year
	donations, err := h.service.GetDonationsForStatement(r.Context(), claims.TenantID, personID, year)
	if err != nil {
		http.Error(w, "Failed to get donations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if len(donations) == 0 {
		http.Error(w, "No donations found for this person in the specified year", http.StatusNotFound)
		return
	}

	// Generate PDF
	pdfBytes, err := GenerateTaxStatementPDF(tenantInfo, personInfo, year, donations)
	if err != nil {
		http.Error(w, "Failed to generate PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"tax-statement-%d-%s.pdf\"", year, personID[:8]))
	w.Write(pdfBytes)
}

// GenerateBatchStatementsPDF generates PDF tax statements for all donors in a year
func (h *Handler) GenerateBatchStatementsPDF(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	yearStr := chi.URLParam(r, "year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}

	// Get all donors who gave in this year
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

	rows, err := h.service.GetDB().Query(r.Context(),
		`SELECT DISTINCT person_id FROM donations 
		 WHERE tenant_id = $1 AND status = 'completed' 
		   AND donated_at >= $2 AND donated_at < $3
		   AND person_id IS NOT NULL
		 ORDER BY person_id`,
		claims.TenantID, yearStart, yearEnd,
	)
	if err != nil {
		http.Error(w, "Failed to get donors: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var donorIDs []string
	for rows.Next() {
		var personID string
		if err := rows.Scan(&personID); err != nil {
			http.Error(w, "Failed to scan donor: "+err.Error(), http.StatusInternalServerError)
			return
		}
		donorIDs = append(donorIDs, personID)
	}

	if len(donorIDs) == 0 {
		http.Error(w, "No donations found for the specified year", http.StatusNotFound)
		return
	}

	// Return list of donors who can have statements generated
	// Frontend can then generate them one by one or batch download
	resp := map[string]interface{}{
		"year":       year,
		"donor_count": len(donorIDs),
		"message":    fmt.Sprintf("Found %d donors with contributions in %d. Use individual statement endpoint to generate PDFs.", len(donorIDs), year),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Recurring donations

func (h *Handler) ListRecurringDonations(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	donations, err := h.service.ListRecurringDonations(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to list recurring donations: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(donations)
}

// Stripe Connect

func (h *Handler) CreateConnectOnboard(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	// Get tenant info from database
	tenantName, err := h.stripeService.GetTenantName(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get tenant info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Use logged-in user's email from claims
	url, err := h.stripeService.CreateConnectOnboardingLink(r.Context(), claims.TenantID, tenantName, claims.Email)
	if err != nil {
		http.Error(w, "Failed to create onboarding link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"url": url}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetConnectStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	status, err := h.stripeService.GetConnectStatus(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get connect status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *Handler) HandleConnectReturn(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Refresh the Connect status to update onboarding completion
	status, err := h.stripeService.GetConnectStatus(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to get connect status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return status to frontend
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *Handler) HandleConnectRefresh(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.Role != "admin" {
		http.Error(w, "Forbidden: admin role required", http.StatusForbidden)
		return
	}

	// Generate new onboarding link
	url, err := h.stripeService.RefreshOnboardingLink(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to refresh onboarding link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"url": url}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Checkout

type CreateCheckoutRequest struct {
	PersonID    string `json:"person_id"`
	FundID      string `json:"fund_id"`
	AmountCents int    `json:"amount_cents"`
}

func (h *Handler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateCheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	url, err := h.stripeService.CreateCheckoutSession(r.Context(), claims.TenantID, req.PersonID, req.FundID, req.AmountCents)
	if err != nil {
		http.Error(w, "Failed to create checkout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"url": url}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Webhook

func (h *Handler) HandleWebhook(webhookSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), webhookSecret)
		if err != nil {
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}

		switch event.Type {
		case "payment_intent.succeeded":
			var pi stripe.PaymentIntent
			if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
				http.Error(w, "Failed to parse event", http.StatusBadRequest)
				return
			}

			if err := h.stripeService.HandlePaymentIntentSucceeded(r.Context(), &pi); err != nil {
				http.Error(w, "Failed to handle payment: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
