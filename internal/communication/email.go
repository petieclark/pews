package communication

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
)

//go:embed templates/*.html
var templatesFS embed.FS

// EmailRenderer handles rendering HTML email templates
type EmailRenderer struct {
	templates map[string]*template.Template
}

// NewEmailRenderer creates a new email renderer with all templates loaded
func NewEmailRenderer() (*EmailRenderer, error) {
	renderer := &EmailRenderer{
		templates: make(map[string]*template.Template),
	}

	// Load base template
	baseContent, err := templatesFS.ReadFile("templates/base.html")
	if err != nil {
		return nil, fmt.Errorf("failed to read base template: %w", err)
	}

	// Load and parse each email template
	templateNames := []string{
		"welcome",
		"event-reminder",
		"volunteer-schedule",
		"giving-receipt",
		"newsletter",
	}

	for _, name := range templateNames {
		tmplContent, err := templatesFS.ReadFile(fmt.Sprintf("templates/%s.html", name))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s template: %w", name, err)
		}

		// Combine base and specific template
		combined := string(baseContent) + "\n" + string(tmplContent)
		
		tmpl, err := template.New(name).Parse(combined)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s template: %w", name, err)
		}

		renderer.templates[name] = tmpl
	}

	return renderer, nil
}

// RenderEmail renders an email template with the provided data
func (r *EmailRenderer) RenderEmail(templateName string, data map[string]interface{}) (string, error) {
	tmpl, exists := r.templates[templateName]
	if !exists {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// GetTemplateNames returns a list of all available template names
func (r *EmailRenderer) GetTemplateNames() []string {
	names := make([]string, 0, len(r.templates))
	for name := range r.templates {
		names = append(names, name)
	}
	return names
}

// GetSampleData returns sample data for previewing a template
func GetSampleData(templateName string) map[string]interface{} {
	// Common church data used across templates
	baseData := map[string]interface{}{
		"ChurchName":     "Grace Community Church",
		"ChurchAddress":  "123 Main Street, Anytown, CA 12345",
		"ChurchPhone":    "(555) 123-4567",
		"WebsiteURL":     "https://gracecommunity.church",
		"FacebookURL":    "https://facebook.com/gracecommunity",
		"InstagramURL":   "https://instagram.com/gracecommunity",
		"UnsubscribeURL": "https://gracecommunity.church/unsubscribe?token=sample123",
		"PreferencesURL": "https://gracecommunity.church/preferences?token=sample123",
	}

	// Template-specific sample data
	switch templateName {
	case "welcome":
		baseData["FirstName"] = "John"
		baseData["ServiceTimes"] = "Sundays at 9:00 AM and 11:00 AM"
		baseData["CTAText"] = "Learn More About Us"
		baseData["CTALink"] = "https://gracecommunity.church/about"
		baseData["PastorName"] = "Rev. Sarah Johnson"
		baseData["PastorTitle"] = "Lead Pastor"

	case "event-reminder":
		baseData["FirstName"] = "Sarah"
		baseData["EventName"] = "Church Picnic"
		baseData["EventDate"] = "Saturday, June 15th, 2024"
		baseData["EventTime"] = "2:00 PM"
		baseData["EventLocation"] = "Central Park Pavilion"
		baseData["EventDescription"] = "Join us for an afternoon of food, games, and fellowship! Bring your family and friends for a great time together."
		baseData["RequiresRSVP"] = true
		baseData["RSVPLink"] = "https://gracecommunity.church/events/picnic/rsvp"
		baseData["AddToCalendarLink"] = "https://gracecommunity.church/events/picnic/calendar"

	case "volunteer-schedule":
		baseData["FirstName"] = "Michael"
		baseData["Role"] = "Worship Team - Audio Tech"
		baseData["ServiceDate"] = "Sunday, June 16th, 2024"
		baseData["ServiceTime"] = "11:00 AM Service"
		baseData["ArrivalTime"] = "10:15 AM"
		baseData["Location"] = "Main Auditorium Sound Booth"
		baseData["TeamLeader"] = "David Martinez"
		baseData["TeamLeaderPhone"] = "(555) 987-6543"
		baseData["SpecialInstructions"] = "We'll be testing new microphones during soundcheck, so please plan for an extra 15 minutes."
		baseData["ConfirmLink"] = "https://gracecommunity.church/volunteers/confirm/abc123"
		baseData["CantMakeItLink"] = "https://gracecommunity.church/volunteers/decline/abc123"

	case "giving-receipt":
		baseData["FirstName"] = "Jennifer"
		baseData["Date"] = "June 10, 2024"
		baseData["Amount"] = "250.00"
		baseData["PaymentMethod"] = "Credit Card ending in 4242"
		baseData["Fund"] = "General Fund"
		baseData["TransactionID"] = "ch_3N2k1nF8j2k3n4l5m"
		baseData["TaxID"] = "12-3456789"
		baseData["IsRecurring"] = true
		baseData["RecurringFrequency"] = "monthly"
		baseData["NextDate"] = "July 10, 2024"
		baseData["ImpactMessage"] = "Your monthly gift is helping us expand our children's ministry and provide resources for 50+ kids every Sunday!"
		baseData["ManageGivingLink"] = "https://gracecommunity.church/giving/manage"

	case "newsletter":
		baseData["FirstName"] = "Alex"
		baseData["Title"] = "This Week at Grace Community"
		baseData["Subtitle"] = "June 14, 2024"
		baseData["Introduction"] = "We hope you had a great week! Here's what's happening at Grace Community this week."
		
		// Articles array
		baseData["Articles"] = []map[string]interface{}{
			{
				"Title":      "New Summer Series Starts This Sunday",
				"Content":    "We're kicking off a brand new teaching series called 'Foundations' where we'll explore the core beliefs that shape our faith. Join us this Sunday as Pastor Sarah begins with 'Why We Believe.'",
				"ButtonText":  "Watch Online",
				"ButtonLink":  "https://gracecommunity.church/watch",
				"IsLast":     false,
			},
			{
				"ImageURL":   "https://gracecommunity.church/images/vbs-2024.jpg",
				"Title":      "VBS Registration is Open!",
				"Content":    "Vacation Bible School is coming July 15-19! Kids ages 4-12 are invited for a week of games, crafts, worship, and learning about God's love. Space is limited, so register today!",
				"ButtonText":  "Register Now",
				"ButtonLink":  "https://gracecommunity.church/vbs",
				"IsLast":     false,
			},
			{
				"Title":      "Serve: Volunteer Appreciation Dinner",
				"Content":    "All volunteers are invited to a special appreciation dinner on June 22nd at 6:00 PM. We want to celebrate YOU and say thank you for all you do to make our church amazing!",
				"ButtonText":  "RSVP Here",
				"ButtonLink":  "https://gracecommunity.church/volunteer-dinner",
				"IsLast":     true,
			},
		}
		
		// Upcoming events
		baseData["Events"] = []map[string]interface{}{
			{
				"Name":     "Men's Breakfast",
				"Date":     "Saturday, June 15",
				"Time":     "8:00 AM",
				"Location": "Fellowship Hall",
			},
			{
				"Name":     "Youth Group",
				"Date":     "Wednesday, June 19",
				"Time":     "7:00 PM",
				"Location": "Youth Center",
			},
		}
		
		baseData["PrayerRequests"] = "Please pray for the Johnson family as they welcome their new baby, for healing for Tom Henderson who is recovering from surgery, and for our mission team traveling to Guatemala next month."
		baseData["Verse"] = "For God so loved the world that he gave his one and only Son, that whoever believes in him shall not perish but have eternal life."
		baseData["VerseReference"] = "John 3:16"
		baseData["ClosingMessage"] = "We're so glad you're part of our church family. Have a blessed week!"
		baseData["SenderName"] = "Pastor Sarah Johnson"
		baseData["SenderTitle"] = "Lead Pastor"
	}

	return baseData
}

// ValidateTemplateData checks if required fields are present in the data
func ValidateTemplateData(templateName string, data map[string]interface{}) error {
	requiredFields := getRequiredFields(templateName)
	
	var missing []string
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			missing = append(missing, field)
		}
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}
	
	return nil
}

// getRequiredFields returns the required data fields for each template
func getRequiredFields(templateName string) []string {
	commonFields := []string{"ChurchName", "ChurchAddress", "UnsubscribeURL"}
	
	switch templateName {
	case "welcome":
		return append(commonFields, "FirstName", "ServiceTimes", "CTAText", "CTALink")
	case "event-reminder":
		return append(commonFields, "FirstName", "EventName", "EventDate", "EventTime", "EventLocation")
	case "volunteer-schedule":
		return append(commonFields, "FirstName", "Role", "ServiceDate", "ServiceTime", "Location", "ConfirmLink", "CantMakeItLink")
	case "giving-receipt":
		return append(commonFields, "FirstName", "Date", "Amount", "PaymentMethod", "Fund")
	case "newsletter":
		return append(commonFields, "FirstName", "Title")
	default:
		return commonFields
	}
}
