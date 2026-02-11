# Weekly Digest Feature - Implementation Summary

## Overview
Built a comprehensive weekly email digest system that summarizes church activity from the past week.

## What Was Built

### Backend (`internal/digest/`)

1. **Models** (`model.go`)
   - `DigestSettings`: Store tenant preferences (enabled, send_day, recipients)
   - `WeeklyDigest`: Main digest data structure containing all stats
   - `AttendanceStats`: This week vs last week with change percentage
   - `MemberStats`: New members this week + total active
   - `GivingStats`: This week + YTD giving with formatted display
   - `UpcomingService[]`: Services scheduled for next 7 days
   - `PrayerRequest[]`: New prayer requests from the week
   - `VolunteerSchedule[]`: Service team assignments for next week

2. **Service** (`service.go`)
   - `GetSettings()`: Retrieve or auto-create digest settings
   - `UpdateSettings()`: Update preferences
   - `GenerateWeeklyDigest()`: Compiles all statistics from multiple tables
   - `RenderDigestHTML()`: Renders the email template with data
   - Helper methods for each stat category (attendance, members, giving, etc.)

3. **Handler** (`handler.go`)
   - `GET /api/digest/settings`: Retrieve settings
   - `PUT /api/digest/settings`: Update settings (admin only)
   - `GET /api/digest/preview`: Full HTML preview (admin only)
   - `GET /api/digest/data`: JSON data preview (admin only)

4. **Email Template** (`templates/weekly.html`)
   - Clean, responsive HTML design
   - Pews branded gradient header
   - Stat cards with trend indicators (↑/↓ arrows)
   - Section-based layout with icons
   - Empty state handling
   - Mobile-responsive grid layout

### Database (`internal/database/migrations/012_digest.sql`)

1. **`digest_settings` table**
   - Stores per-tenant configuration
   - Unique constraint ensures one settings row per tenant
   - Default: enabled, send on Monday, empty recipients array

2. **`digest_history` table**
   - Tracks sent digests for audit/reporting
   - Stores week boundaries and recipient list
   - Indexed by tenant + date for fast queries

### Frontend (`web/src/routes/dashboard/settings/digest/+page.svelte`)

1. **Settings UI**
   - Enable/disable toggle
   - Day-of-week selector (defaults to Monday)
   - Email recipient management (add/remove)
   - Save settings button
   - Preview digest button (opens HTML in new tab)
   - Preview data card showing key metrics

2. **Features**
   - Form validation
   - Success/error messaging
   - Loading states
   - Responsive design
   - Links back to main settings page

## Integration Points

### Router (`internal/router/router.go`)
- Added digest import
- Added digestHandler parameter to New()
- Registered 4 new protected API routes

### Main (`cmd/pews/main.go`)
- Added digest import
- Initialize digestService with pgxpool
- Initialize digestHandler
- Pass digestHandler to router

## Testing

```bash
# Start services
cd ~/Projects/pews
docker compose up -d

# Services running on:
# - Backend: http://localhost:8190
# - Frontend: http://localhost:5273
# - Postgres: localhost:5432
```

### Test Endpoints
```bash
# Get settings (requires auth token)
curl http://localhost:8190/api/digest/settings \
  -H "Authorization: Bearer YOUR_TOKEN"

# Preview digest HTML (requires admin auth)
curl http://localhost:8190/api/digest/preview \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"

# Get digest data as JSON
curl http://localhost:8190/api/digest/data \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

### Frontend Access
1. Login to http://localhost:5273
2. Navigate to Settings
3. Go to /dashboard/settings/digest (or add a link in main settings)
4. Toggle settings
5. Click "Preview Digest" to see rendered HTML

## What's NOT Implemented
- **Email sending**: The digest generation works, but there's no scheduled job or SMTP integration to actually send the emails
- **Scheduled execution**: No cron/worker to trigger digest generation weekly
- **Email service integration**: Would need to add email service (SendGrid, AWS SES, etc.)
- **Link to main settings**: The digest settings page exists but isn't linked from the main settings page yet

## Next Steps (For Production)
1. Add link in main settings page to digest settings
2. Implement email service (SMTP or API-based)
3. Add background worker/cron to generate and send weekly
4. Add "Send Test Email" button to test before scheduling
5. Track delivery status in digest_history table
6. Add email templates for transactional messages (not just digest)
7. Implement unsubscribe functionality

## Stats Compiled
- **Attendance**: Unique check-ins this week vs last week with change %
- **Members**: New people added this week + total active count
- **Giving**: Total donations this week + YTD with formatted display
- **Services**: Upcoming services in next 7 days
- **Prayer**: New prayer requests from connection cards
- **Volunteers**: Service team assignments for next week

## Technical Details
- Uses Go's `embed` package for HTML template
- PostgreSQL arrays for recipients list
- Row-level security policies for multi-tenancy
- pgx/v5 for database access (not database/sql)
- Svelte for frontend with Tailwind CSS
- RESTful JSON API

## Branch
- Feature branch: `feat/weekly-digest`
- Commit: "feat: Add weekly digest email feature"
- Status: ✅ Complete (email sending not wired)
- Ready for: Code review, testing, email integration

## Files Changed
- `internal/database/migrations/012_digest.sql` (new)
- `internal/digest/handler.go` (new)
- `internal/digest/model.go` (new)
- `internal/digest/service.go` (new)
- `internal/digest/templates/weekly.html` (new)
- `web/src/routes/dashboard/settings/digest/+page.svelte` (new)
- `internal/router/router.go` (modified - added digest routes)
- `cmd/pews/main.go` (modified - added digest initialization)
- `go.mod` (updated dependencies)
- `go.sum` (updated checksums)
