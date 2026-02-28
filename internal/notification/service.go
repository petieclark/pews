package notification

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NotificationService handles sending volunteer assignment notifications
type NotificationService struct {
	db           *pgxpool.Pool
	sendGridURL  string
	fromEmail    string
	fromName     string
	pewsBaseURL  string
	tokenSecret  string
}

// AssignmentNotificationData holds data for the email template
type AssignmentNotificationData struct {
	PersonFirstName string
	PersonLastName  string
	ServiceName     string
	ServiceDate     string
	ServiceTime     string
	TeamName        string
	PositionName    string
	AceptURL        string
	DeclineURL      string
	ChurchName      string
	CurrentYear     int
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *pgxpool.Pool) *NotificationService {
	return &NotificationService{
		db:          db,
		sendGridURL: os.Getenv("SENDGRID_API_KEY"),
		fromEmail:   os.Getenv("SENDGRID_FROM_EMAIL"),
		fromName:    os.Getenv("SENDGRID_FROM_NAME"),
		pewsBaseURL: getBaseURL(),
		tokenSecret: os.Getenv("JWT_SECRET"),
	}
}

// getBaseURL returns the base URL for the application
func getBaseURL() string {
	url := os.Getenv("PEWS_BASE_URL")
	if url == "" {
		url = "https://app.pews.local" // default for development
	}
	return url
}

// SendAssignmentNotification sends a volunteer assignment notification email
func (ns *NotificationService) SendAssignmentNotification(ctx context.Context, assignmentID, personID string) error {
	// Get person and assignment details from database
	personName, serviceDetails, err := ns.getAssignmentDetails(ctx, assignmentID, personID)
	if err != nil {
		return fmt.Errorf("failed to get assignment details: %w", err)
	}

	// Generate tokenized URLs for accept/decline
	acceptURL := fmt.Sprintf("%s/respond/%s?action=accept", ns.pewsBaseURL, assignmentID)
	declineURL := fmt.Sprintf("%s/respond/%s?action=decline", ns.pewsBaseURL, assignmentID)

	// Prepare email template data
	data := AssignmentNotificationData{
		PersonFirstName: personName.FirstName,
		PersonLastName:  personName.LastName,
		ServiceName:     serviceDetails.ServiceName,
		ServiceDate:     serviceDetails.ServiceDate,
		ServiceTime:     serviceDetails.ServiceTime,
		TeamName:        serviceDetails.TeamName,
		PositionName:    serviceDetails.PositionName,
		AceptURL:        acceptURL,
		DeclineURL:      declineURL,
		ChurchName:      ns.fromName,
		CurrentYear:     time.Now().Year(),
	}

	// Render HTML email template
	htmlContent, err := ns.renderTemplate(data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Send via SendGrid (or log in dev mode)
	err = ns.sendEmail(personName.Email, "Volunteer Assignment - "+serviceDetails.ServiceName, htmlContent)
	if err != nil {
		return fmt.Errorf("failed to send notification email: %w", err)
	}

	// Mark as notified in database
	if err := ns.markNotified(ctx, assignmentID); err != nil {
		return fmt.Errorf("failed to mark assignment as notified: %w", err)
	}

	return nil
}

type PersonName struct {
	FirstName string
	LastName  string
	Email     string
}

type ServiceDetails struct {
	ServiceName   string
	ServiceDate   string
	ServiceTime   string
	TeamName      string
	PositionName  string
}

func (ns *NotificationService) getAssignmentDetails(ctx context.Context, assignmentID, personID string) (*PersonName, *ServiceDetails, error) {
	person := &PersonName{}
	service := &ServiceDetails{}

	err := ns.db.QueryRow(ctx, `
		SELECT p.first_name, p.last_name, COALESCE(p.email, ''),
		       s.name as service_name, TO_CHAR(s.service_date, 'Month DD, YYYY'),
		       COALESCE(s.service_time, ''),
		       t.name as team_name,
		       COALESCE(tp.name, 'Volunteer') as position_name
		FROM service_team_assignments sta
		JOIN people p ON p.id = sta.person_id
		JOIN services s ON s.id = sta.service_id
		JOIN teams t ON t.id = sta.team_id
		LEFT JOIN team_positions tp ON tp.id = sta.position_id
		WHERE sta.id = $1 AND sta.person_id = $2`,
		assignmentID, personID).Scan(
		&person.FirstName, &person.LastName, &person.Email,
		&service.ServiceName, &service.ServiceDate, &service.ServiceTime,
		&service.TeamName, &service.PositionName,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to query assignment details: %w", err)
	}

	return person, service, nil
}

func (ns *NotificationService) renderTemplate(data AssignmentNotificationData) (string, error) {
	tmplPath := "internal/email/templates/volunteer-assignment.html"
	
	// Read template file
	templateBytes, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to read email template: %w", err)
	}

	tmpl, err := template.New("volunteer-assignment").Parse(string(templateBytes))
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}

	var rendered string
	buf := &strings.Builder{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", fmt.Errorf("failed to execute email template: %w", err)
	}
	rendered = buf.String()

	return rendered, nil
}

func (ns *NotificationService) sendEmail(toEmail, subject, htmlContent string) error {
	devMode := os.Getenv("DEV_MODE") == "true"

	if devMode {
		fmt.Printf("[notification] [DEV MODE] Would send email to %s: %s\n", toEmail, subject)
		return nil
	}

	// TODO: Implement SendGrid client integration here
	// For now, return a placeholder error that can be implemented later
	return fmt.Errorf("SendGrid integration not yet implemented - implement using internal/communication/sendgrid.go")
}

func (ns *NotificationService) markNotified(ctx context.Context, assignmentID string) error {
	now := time.Now()
	_, err := ns.db.Exec(ctx, `
		UPDATE service_team_assignments 
		SET notified_at = $1, notification_sent = true 
		WHERE id = $2`,
		now, assignmentID,
	)

	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}
