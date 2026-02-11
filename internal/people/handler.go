package people

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/petieclark/pews/internal/activity"
	"github.com/petieclark/pews/internal/middleware"
	"github.com/petieclark/pews/internal/notification"
)

type Handler struct {
	service             *Service
	activityService     *activity.Service
	notificationService *notification.Service
}

func NewHandler(service *Service, activityService *activity.Service) *Handler {
	return &Handler{
		service:             service,
		activityService:     activityService,
		notificationService: notification.NewService(service.GetDB()),
	}
}

// People handlers

type CreatePersonRequest struct {
	FirstName        string          `json:"first_name"`
	LastName         string          `json:"last_name"`
	Email            string          `json:"email,omitempty"`
	Phone            string          `json:"phone,omitempty"`
	AddressLine1     string          `json:"address_line1,omitempty"`
	AddressLine2     string          `json:"address_line2,omitempty"`
	City             string          `json:"city,omitempty"`
	State            string          `json:"state,omitempty"`
	Zip              string          `json:"zip,omitempty"`
	Birthdate        string          `json:"birthdate,omitempty"`
	Gender           string          `json:"gender,omitempty"`
	MembershipStatus string          `json:"membership_status,omitempty"`
	PhotoURL         string          `json:"photo_url,omitempty"`
	Notes            string          `json:"notes,omitempty"`
	CustomFields     json.RawMessage `json:"custom_fields,omitempty"`
}

func (h *Handler) ListPeople(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	people, total, err := h.service.ListPeople(r.Context(), claims.TenantID, query, page, limit)
	if err != nil {
		http.Error(w, "Failed to list people: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"people": people,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetPerson(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	personID := chi.URLParam(r, "id")
	person, err := h.service.GetPersonByID(r.Context(), claims.TenantID, personID)
	if err != nil {
		http.Error(w, "Person not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(person)
}

func (h *Handler) CreatePerson(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreatePersonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	person := &Person{
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Email:            req.Email,
		Phone:            req.Phone,
		AddressLine1:     req.AddressLine1,
		AddressLine2:     req.AddressLine2,
		City:             req.City,
		State:            req.State,
		Zip:              req.Zip,
		Gender:           req.Gender,
		MembershipStatus: req.MembershipStatus,
		PhotoURL:         req.PhotoURL,
		Notes:            req.Notes,
		CustomFields:     req.CustomFields,
	}

	if person.MembershipStatus == "" {
		person.MembershipStatus = "active"
	}

	createdPerson, err := h.service.CreatePerson(r.Context(), claims.TenantID, person)
	if err != nil {
		http.Error(w, "Failed to create person: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create notification for all admins
	notifTitle := "New Member Registered"
	notifMessage := fmt.Sprintf("%s %s has been added to the directory", createdPerson.FirstName, createdPerson.LastName)
	link := fmt.Sprintf("/people/%s", createdPerson.ID)
	_ = h.notificationService.CreateForAllAdmins(r.Context(), claims.TenantID, notifTitle, notifMessage, notification.TypeInfo, &link)

	// Log activity
	ipAddress := r.RemoteAddr
	details := map[string]interface{}{
		"name":  createdPerson.FirstName + " " + createdPerson.LastName,
		"email": createdPerson.Email,
	}
	_ = h.activityService.LogActivity(r.Context(), claims.TenantID, "person.created", "people", &claims.UserID, &createdPerson.ID, &ipAddress, details)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPerson)
}

func (h *Handler) UpdatePerson(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	personID := chi.URLParam(r, "id")

	var req CreatePersonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	person := &Person{
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Email:            req.Email,
		Phone:            req.Phone,
		AddressLine1:     req.AddressLine1,
		AddressLine2:     req.AddressLine2,
		City:             req.City,
		State:            req.State,
		Zip:              req.Zip,
		Gender:           req.Gender,
		MembershipStatus: req.MembershipStatus,
		PhotoURL:         req.PhotoURL,
		Notes:            req.Notes,
		CustomFields:     req.CustomFields,
	}

	updatedPerson, err := h.service.UpdatePerson(r.Context(), claims.TenantID, personID, person)
	if err != nil {
		http.Error(w, "Failed to update person: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log activity
	ipAddress := r.RemoteAddr
	details := map[string]interface{}{
		"name":  updatedPerson.FirstName + " " + updatedPerson.LastName,
		"email": updatedPerson.Email,
	}
	_ = h.activityService.LogActivity(r.Context(), claims.TenantID, "person.updated", "people", &claims.UserID, &updatedPerson.ID, &ipAddress, details)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPerson)
}

func (h *Handler) DeletePerson(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	personID := chi.URLParam(r, "id")

	// Get person details before deletion for logging
	person, _ := h.service.GetPersonByID(r.Context(), claims.TenantID, personID)

	if err := h.service.DeletePerson(r.Context(), claims.TenantID, personID); err != nil {
		http.Error(w, "Failed to delete person: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log activity
	if person != nil {
		ipAddress := r.RemoteAddr
		details := map[string]interface{}{
			"name": person.FirstName + " " + person.LastName,
		}
		_ = h.activityService.LogActivity(r.Context(), claims.TenantID, "person.deleted", "people", &claims.UserID, &personID, &ipAddress, details)
	}

	w.WriteHeader(http.StatusNoContent)
}

// Tag handlers

type AddTagRequest struct {
	TagID string `json:"tag_id"`
}

func (h *Handler) AddTagToPerson(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	personID := chi.URLParam(r, "id")

	var req AddTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.AddTagToPerson(r.Context(), claims.TenantID, personID, req.TagID); err != nil {
		http.Error(w, "Failed to add tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveTagFromPerson(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	personID := chi.URLParam(r, "id")
	tagID := chi.URLParam(r, "tagId")

	if err := h.service.RemoveTagFromPerson(r.Context(), claims.TenantID, personID, tagID); err != nil {
		http.Error(w, "Failed to remove tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type CreateTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tags, err := h.service.ListTags(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to list tags: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tags)
}

func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tag := &Tag{
		Name:  req.Name,
		Color: req.Color,
	}

	if tag.Color == "" {
		tag.Color = "#4A8B8C"
	}

	createdTag, err := h.service.CreateTag(r.Context(), claims.TenantID, tag)
	if err != nil {
		http.Error(w, "Failed to create tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdTag)
}

// Household handlers

type CreateHouseholdRequest struct {
	Name             string `json:"name"`
	PrimaryContactID string `json:"primary_contact_id,omitempty"`
	AddressLine1     string `json:"address_line1,omitempty"`
	AddressLine2     string `json:"address_line2,omitempty"`
	City             string `json:"city,omitempty"`
	State            string `json:"state,omitempty"`
	Zip              string `json:"zip,omitempty"`
}

func (h *Handler) ListHouseholds(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	households, err := h.service.ListHouseholds(r.Context(), claims.TenantID)
	if err != nil {
		http.Error(w, "Failed to list households: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(households)
}

func (h *Handler) CreateHousehold(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateHouseholdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	household := &Household{
		Name:             req.Name,
		PrimaryContactID: req.PrimaryContactID,
		AddressLine1:     req.AddressLine1,
		AddressLine2:     req.AddressLine2,
		City:             req.City,
		State:            req.State,
		Zip:              req.Zip,
	}

	createdHousehold, err := h.service.CreateHousehold(r.Context(), claims.TenantID, household)
	if err != nil {
		http.Error(w, "Failed to create household: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdHousehold)
}

func (h *Handler) UpdateHousehold(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	householdID := chi.URLParam(r, "id")

	var req CreateHouseholdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	household := &Household{
		Name:             req.Name,
		PrimaryContactID: req.PrimaryContactID,
		AddressLine1:     req.AddressLine1,
		AddressLine2:     req.AddressLine2,
		City:             req.City,
		State:            req.State,
		Zip:              req.Zip,
	}

	updatedHousehold, err := h.service.UpdateHousehold(r.Context(), claims.TenantID, householdID, household)
	if err != nil {
		http.Error(w, "Failed to update household: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedHousehold)
}

type AddMemberRequest struct {
	PersonID string `json:"person_id"`
	Role     string `json:"role"`
}

func (h *Handler) AddMemberToHousehold(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	householdID := chi.URLParam(r, "id")

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	if err := h.service.AddMemberToHousehold(r.Context(), claims.TenantID, householdID, req.PersonID, req.Role); err != nil {
		http.Error(w, "Failed to add member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveMemberFromHousehold(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	householdID := chi.URLParam(r, "id")
	personID := chi.URLParam(r, "personId")

	if err := h.service.RemoveMemberFromHousehold(r.Context(), claims.TenantID, householdID, personID); err != nil {
		http.Error(w, "Failed to remove member: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
