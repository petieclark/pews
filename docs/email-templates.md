# Email Template System

## Overview

The email template system provides HTML email rendering for the Communication module. Templates are built using Go's `html/template` package with email-friendly HTML (table-based layouts with inline CSS).

## Templates

All templates are located in `internal/communication/templates/`:

### Base Template (`base.html`)
- Shared layout with navy header and teal accents
- Responsive design (600px max width)
- Includes footer with church info, social links, and CAN-SPAM compliant unsubscribe link
- All other templates extend this base

### Available Email Templates

#### 1. Welcome Email (`welcome.html`)
**Purpose:** New member welcome message

**Required Variables:**
- `FirstName` - Recipient's first name
- `ServiceTimes` - Service schedule text
- `CTAText` - Call-to-action button text
- `CTALink` - Call-to-action button URL
- `PastorName` - Pastor's full name
- `PastorTitle` - Pastor's title

**Optional Variables:**
- `ChurchName`, `ChurchAddress`, `ChurchPhone`, etc.

#### 2. Event Reminder (`event-reminder.html`)
**Purpose:** Upcoming service or event reminder

**Required Variables:**
- `FirstName`
- `EventName`
- `EventDate`
- `EventTime`
- `EventLocation`

**Optional Variables:**
- `EventDescription`
- `RequiresRSVP` (bool)
- `RSVPLink`
- `AddToCalendarLink`

#### 3. Volunteer Schedule (`volunteer-schedule.html`)
**Purpose:** "You're scheduled to serve" notification

**Required Variables:**
- `FirstName`
- `Role` - Volunteer role/position
- `ServiceDate`
- `ServiceTime`
- `Location`
- `ConfirmLink`
- `CantMakeItLink`

**Optional Variables:**
- `ArrivalTime`
- `TeamLeader`
- `TeamLeaderPhone`
- `SpecialInstructions`

#### 4. Giving Receipt (`giving-receipt.html`)
**Purpose:** Donation receipt/confirmation

**Required Variables:**
- `FirstName`
- `Date` - Transaction date
- `Amount` - Donation amount (formatted)
- `PaymentMethod`
- `Fund` - Fund designation

**Optional Variables:**
- `TransactionID`
- `TaxID` - Church Tax ID
- `IsRecurring` (bool)
- `RecurringFrequency`
- `NextDate`
- `ImpactMessage`
- `ManageGivingLink`

#### 5. Newsletter (`newsletter.html`)
**Purpose:** General church newsletter

**Required Variables:**
- `FirstName`
- `Title` - Newsletter title

**Optional Variables:**
- `HeaderImageURL`
- `Subtitle`
- `Introduction`
- `Articles` - Array of article objects:
  ```go
  {
    "Title": "Article Title",
    "Content": "Article body text",
    "ImageURL": "optional",
    "ButtonText": "optional CTA",
    "ButtonLink": "optional CTA URL",
    "IsLast": bool
  }
  ```
- `Events` - Array of event objects:
  ```go
  {
    "Name": "Event Name",
    "Date": "Event Date",
    "Time": "Event Time",
    "Location": "optional"
  }
  ```
- `PrayerRequests`
- `Verse` - Scripture text
- `VerseReference`
- `ClosingMessage`
- `SenderName`
- `SenderTitle`

## API Endpoints

### List Available Templates
```
GET /api/communication/email-templates
```
Returns array of available template names.

**Response:**
```json
{
  "templates": ["welcome", "event-reminder", "volunteer-schedule", "giving-receipt", "newsletter"]
}
```

### Preview Template
```
GET /api/communication/email-templates/{name}/preview
```
Renders the specified template with sample data for preview in browser.

**Parameters:**
- `name` - Template name (welcome, event-reminder, volunteer-schedule, giving-receipt, newsletter)

**Response:** HTML email (Content-Type: text/html)

## Usage in Code

### Initialize Email Renderer
```go
renderer, err := communication.NewEmailRenderer()
if err != nil {
    log.Fatal(err)
}
```

### Render Template
```go
data := map[string]interface{}{
    "ChurchName": "Grace Community Church",
    "FirstName": "John",
    "ServiceTimes": "Sundays at 9 AM and 11 AM",
    // ... other required fields
}

html, err := renderer.RenderEmail("welcome", data)
if err != nil {
    log.Fatal(err)
}

// html is ready to send via email provider (SendGrid, Mailgun, etc.)
```

### Validate Data
```go
err := communication.ValidateTemplateData("welcome", data)
if err != nil {
    // Missing required fields
    log.Printf("Validation error: %v", err)
}
```

### Get Sample Data
```go
sampleData := communication.GetSampleData("welcome")
// Returns map with all sample values filled in
```

## Email Design Principles

### HTML Structure
- **Table-based layout** - Modern CSS doesn't work reliably in email clients
- **Inline CSS** - Email clients strip `<style>` tags
- **Max width 600px** - Standard email width
- **No external CSS or JavaScript**

### Brand Colors
- **Navy header:** `#1e3a5f`
- **Teal accents:** `#2dd4bf`
- **Gray text:** `#374151`
- **Light backgrounds:** `#f0fdfa`, `#eff6ff`, `#f9fafb`

### CAN-SPAM Compliance
All templates include:
- Physical mailing address in footer
- Unsubscribe link
- "Why you're receiving this" text

## Testing

### Build & Run
```bash
# Build Docker image
docker build -t pews-backend:latest .

# Run container
docker run -p 8080:8080 pews-backend:latest

# Preview templates in browser
open http://localhost:8080/api/communication/email-templates/welcome/preview
```

### Test with HTML Email Checker
Visit https://www.htmlemailcheck.com/ and paste rendered HTML to test compatibility across email clients.

### Manual Testing
1. Send test emails to yourself on different clients (Gmail, Outlook, Apple Mail)
2. Check mobile rendering (most people read emails on mobile)
3. Test all links and buttons
4. Verify unsubscribe link is visible

## Future Integration

### Email Sending (Not Yet Implemented)
When ready to wire up email sending, the rendered HTML can be sent via:

**SendGrid:**
```go
from := mail.NewEmail("Grace Church", "hello@gracechurch.com")
subject := "Welcome to Grace Church"
to := mail.NewEmail("John Doe", "john@example.com")
htmlContent := renderedHTML
message := mail.NewSingleEmail(from, subject, to, "", htmlContent)
client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
response, err := client.Send(message)
```

**Mailgun:**
```go
mg := mailgun.NewMailgun("your-domain.com", "api-key")
m := mg.NewMessage(
    "hello@gracechurch.com",
    "Welcome to Grace Church",
    "", // plain text (optional)
    "john@example.com",
)
m.SetHtml(renderedHTML)
ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
defer cancel()
resp, id, err := mg.Send(ctx, m)
```

## Notes

- Templates are embedded at compile time using `//go:embed`
- No external template files needed at runtime
- All templates are pre-parsed at initialization
- Sample data includes realistic examples for all template types
- Preview endpoint requires authentication but not tenant context

## Branch Status

✅ **Committed to:** `feat/email-templates`  
❌ **Not merged to main** (as requested)

## Next Steps

1. Wire up actual email sending (SendGrid/Mailgun integration)
2. Add template variable substitution from database models (Person, Event, etc.)
3. Create email preference management for unsubscribe handling
4. Add email tracking pixels for open/click tracking
5. Build admin UI for previewing templates before sending campaigns
