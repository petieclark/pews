package communication

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/middleware"
	"github.com/petieclark/pews/internal/notification"
)

type Handler struct {
	service             *Service
	emailRenderer       *EmailRenderer
	notificationService *notification.NotificationService
}

func NewHandler(service *Service) *Handler {
	// Initialize email renderer
	renderer, err := NewEmailRenderer()
	if err != nil {
		// In production, handle this more gracefully
		panic(fmt.Sprintf("failed to initialize email renderer: %v", err))
	}

	return &Handler{
		service:             service,
		emailRenderer:       renderer,
		notificationService: notification.NewNotificationService(service.GetDB()),
	}
}

// ===== TEMPLATES =====

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channel := r.URL.Query().Get("channel")
	category := r.URL.Query().Get("category")

	templates, err := h.service.ListTemplates(r.Context(), claims.TenantID, channel, category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

type CreateTemplateRequest struct {
	Name      string `json:"name"`
	Subject   string `json:"subject,omitempty"`
	Body      string `json:"body"`
	Channel   string `json:"channel"`
	Category  string `json:"category,omitempty"`
	Variables string `json:"variables,omitempty"`
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	template := &MessageTemplate{
		Name:      req.Name,
		Subject:   req.Subject,
		Body:      req.Body,
		Channel:   req.Channel,
		Category:  req.Category,
		Variables: req.Variables,
	}

	created, err := h.service.CreateTemplate(r.Context(), claims.TenantID, template)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templateID := chi.URLParam(r, "id")

	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	template := &MessageTemplate{
		Name:      req.Name,
		Subject:   req.Subject,
		Body:      req.Body,
		Channel:   req.Channel,
		Category:  req.Category,
		Variables: req.Variables,
	}

	if err := h.service.UpdateTemplate(r.Context(), claims.TenantID, templateID, template); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templateID := chi.URLParam(r, "id")

	if err := h.service.DeleteTemplate(r.Context(), claims.TenantID, templateID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ===== CAMPAIGNS =====

func (h *Handler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	status := r.URL.Query().Get("status")

	campaigns, err := h.service.ListCampaigns(r.Context(), claims.TenantID, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaigns)
}

type CreateCampaignRequest struct {
	Name       string  `json:"name"`
	TemplateID *string `json:"template_id,omitempty"`
	Channel    string  `json:"channel"`
	Subject    string  `json:"subject,omitempty"`
	Body       string  `json:"body"`
	TargetType string  `json:"target_type"`
	TargetID   string  `json:"target_id,omitempty"`
}

func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	campaign := &Campaign{
		Name:       req.Name,
		TemplateID: req.TemplateID,
		Channel:    req.Channel,
		Subject:    req.Subject,
		Body:       req.Body,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
	}

	created, err := h.service.CreateCampaign(r.Context(), claims.TenantID, campaign)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "id")

	campaign, err := h.service.GetCampaign(r.Context(), claims.TenantID, campaignID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaign)
}

func (h *Handler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "id")

	var req CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	campaign := &Campaign{
		Name:       req.Name,
		Channel:    req.Channel,
		Subject:    req.Subject,
		Body:       req.Body,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
	}

	if err := h.service.UpdateCampaign(r.Context(), claims.TenantID, campaignID, campaign); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type SendCampaignRequest struct {
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

func (h *Handler) SendCampaign(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "id")

	var req SendCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.SendCampaign(r.Context(), claims.TenantID, campaignID, req.ScheduledAt); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetCampaignRecipients(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	campaignID := chi.URLParam(r, "id")

	recipients, err := h.service.GetCampaignRecipients(r.Context(), claims.TenantID, campaignID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recipients)
}

// ===== JOURNEYS =====

func (h *Handler) ListJourneys(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	journeys, err := h.service.ListJourneys(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(journeys)
}

type CreateJourneyRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	TriggerType  string `json:"trigger_type"`
	TriggerValue string `json:"trigger_value,omitempty"`
	IsActive     bool   `json:"is_active"`
}

func (h *Handler) CreateJourney(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateJourneyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	journey := &Journey{
		Name:         req.Name,
		Description:  req.Description,
		TriggerType:  req.TriggerType,
		TriggerValue: req.TriggerValue,
		IsActive:     req.IsActive,
	}

	created, err := h.service.CreateJourney(r.Context(), claims.TenantID, journey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) GetJourney(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	journeyID := chi.URLParam(r, "id")

	journey, err := h.service.GetJourney(r.Context(), claims.TenantID, journeyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(journey)
}

func (h *Handler) UpdateJourney(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	journeyID := chi.URLParam(r, "id")

	var req CreateJourneyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	journey := &Journey{
		Name:         req.Name,
		Description:  req.Description,
		TriggerType:  req.TriggerType,
		TriggerValue: req.TriggerValue,
		IsActive:     req.IsActive,
	}

	if err := h.service.UpdateJourney(r.Context(), claims.TenantID, journeyID, journey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteJourney(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	journeyID := chi.URLParam(r, "id")

	if err := h.service.DeleteJourney(r.Context(), claims.TenantID, journeyID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ===== JOURNEY STEPS =====

type CreateJourneyStepRequest struct {
	Position   int             `json:"position"`
	StepType   string          `json:"step_type"`
	DelayDays  int             `json:"delay_days"`
	DelayHours int             `json:"delay_hours"`
	TemplateID *string         `json:"template_id,omitempty"`
	Config     json.RawMessage `json:"config"`
}

func (h *Handler) AddJourneyStep(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	journeyID := chi.URLParam(r, "id")

	var req CreateJourneyStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	step := &JourneyStep{
		Position:   req.Position,
		StepType:   req.StepType,
		DelayDays:  req.DelayDays,
		DelayHours: req.DelayHours,
		TemplateID: req.TemplateID,
		Config:     req.Config,
	}

	created, err := h.service.AddJourneyStep(r.Context(), claims.TenantID, journeyID, step)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *Handler) UpdateJourneyStep(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stepID := chi.URLParam(r, "stepId")

	var req CreateJourneyStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	step := &JourneyStep{
		Position:   req.Position,
		StepType:   req.StepType,
		DelayDays:  req.DelayDays,
		DelayHours: req.DelayHours,
		TemplateID: req.TemplateID,
		Config:     req.Config,
	}

	if err := h.service.UpdateJourneyStep(r.Context(), claims.TenantID, stepID, step); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteJourneyStep(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stepID := chi.URLParam(r, "stepId")

	if err := h.service.DeleteJourneyStep(r.Context(), claims.TenantID, stepID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ===== JOURNEY ENROLLMENTS =====

type EnrollInJourneyRequest struct {
	PersonID string `json:"person_id"`
}

func (h *Handler) EnrollInJourney(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	journeyID := chi.URLParam(r, "id")

	var req EnrollInJourneyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	enrollment, err := h.service.EnrollInJourney(r.Context(), claims.TenantID, journeyID, req.PersonID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(enrollment)
}

func (h *Handler) GetJourneyEnrollments(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	journeyID := chi.URLParam(r, "id")

	enrollments, err := h.service.GetJourneyEnrollments(r.Context(), claims.TenantID, journeyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enrollments)
}

// ===== JOURNEY ACTIVATION =====

func (h *Handler) ToggleJourneyActive(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	journeyID := chi.URLParam(r, "id")

	journey, err := h.service.ToggleJourneyActive(r.Context(), claims.TenantID, journeyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(journey)
}

// ===== JOURNEY PROCESSING =====

func (h *Handler) ProcessJourneySteps(w http.ResponseWriter, r *http.Request) {
	processed, err := h.service.ProcessDueSteps(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"processed": processed,
		"status":    "ok",
	})
}

// ===== CONNECTION CARDS =====

type SubmitConnectionCardRequest struct {
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name,omitempty"`
	Email         string `json:"email,omitempty"`
	Phone         string `json:"phone,omitempty"`
	IsFirstVisit  bool   `json:"is_first_visit"`
	HowHeard      string `json:"how_heard,omitempty"`
	PrayerRequest string `json:"prayer_request,omitempty"`
	InterestedIn  string `json:"interested_in,omitempty"`
}

// PUBLIC endpoint - no auth required
func (h *Handler) SubmitConnectionCard(w http.ResponseWriter, r *http.Request) {
	// Extract tenant ID from request context (set by tenant extractor middleware)
	tenantID := middleware.GetTenantID(r.Context())
	if tenantID == "" {
		http.Error(w, "Tenant not found", http.StatusBadRequest)
		return
	}

	var req SubmitConnectionCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	card := &ConnectionCard{
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Email:         req.Email,
		Phone:         req.Phone,
		IsFirstVisit:  req.IsFirstVisit,
		HowHeard:      req.HowHeard,
		PrayerRequest: req.PrayerRequest,
		InterestedIn:  req.InterestedIn,
	}

	created, err := h.service.SubmitConnectionCard(r.Context(), tenantID, card)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send welcome email (async, best-effort)
	go h.service.GetSender().SendWelcomeEmail(context.Background(), tenantID, created)

	// Auto-enroll in connection_card journeys (best-effort, async)
	// First, create or find person for this card
	go func() {
		bgCtx := context.Background()
		// Try to find person by email, then enroll
		if created.Email != "" {
			var personID string
			err := h.service.GetDB().QueryRow(bgCtx,
				`SELECT id FROM people WHERE tenant_id = $1 AND email = $2 LIMIT 1`,
				tenantID, created.Email).Scan(&personID)
			if err == nil && personID != "" {
				h.service.AutoEnrollByTrigger(bgCtx, tenantID, "connection_card", personID)
			}
		}
	}()

	// Create notification for all admins (stubbed for now)
	notifTitle := "New Connection Card"
	notifMessage := fmt.Sprintf("%s %s submitted a connection card", created.FirstName, created.LastName)
	link := fmt.Sprintf("/communication/cards/%s", created.ID)
	_ = h.notificationService.CreateForAllAdmins(r.Context(), tenantID, notifTitle, notifMessage, notification.TypeInfo, &link)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// CreateForAllAdmins creates a notification for all admins (stubbed implementation)
func (h *Handler) CreateForAllAdmins(ctx context.Context, tenantID, title, message string, typ notification.NotificationType, link *string) error {
	// TODO: Implement admin notification creation
	// This is a stub to allow compilation - will be implemented separately
	return nil
}

func (h *Handler) ListConnectionCards(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cards, err := h.service.ListConnectionCards(r.Context(), claims.TenantID, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

func (h *Handler) GetConnectionCard(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cardID := chi.URLParam(r, "id")

	card, err := h.service.GetConnectionCard(r.Context(), claims.TenantID, cardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}

type ProcessConnectionCardRequest struct {
	PersonID string `json:"person_id"`
}

func (h *Handler) ProcessConnectionCard(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	cardID := chi.URLParam(r, "id")

	var req ProcessConnectionCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.ProcessConnectionCard(r.Context(), claims.TenantID, cardID, req.PersonID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ===== STATS =====

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := h.service.GetStats(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ===== EMAIL TEMPLATE PREVIEW =====

// PreviewEmailTemplate renders an email template with sample data for preview
func (h *Handler) PreviewEmailTemplate(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templateName := chi.URLParam(r, "name")
	data := GetSampleData(templateName)

	html, err := h.emailRenderer.RenderEmail(templateName, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// ListEmailTemplates returns a list of available email template names
func (h *Handler) ListEmailTemplates(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	templates := h.emailRenderer.GetTemplateNames()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
	})
}
