# Scheduling Needs Dashboard - Issue #65 Implementation

**Status:** ✅ COMPLETE  
**Task ID:** 15  
**Assignee:** maeve  
**Date:** 2026-02-28  

---

## Overview

Implemented a scheduling gaps dashboard for Pews that allows administrators to view all unfilled volunteer positions across upcoming services and quickly assign team members without leaving the page.

---

## Features Delivered

### Backend API Endpoints

1. **GET `/api/scheduling/needs?start=&end=`**
   - Returns all unfilled service team positions
   - Default: shows next 4 weeks from today
   - Date range filterable via query params
   - Groups results by date in response
   - Includes urgency flag for services within 7 days

2. **GET `/api/people/search?q=`**
   - Searches people by name or email (ILIKE)
   - Returns availability status for scheduling conflicts check
   - Used by quick-assign feature

3. **PUT `/api/services/service-teams/{id}/assign`**
   - Assigns a person to an unfilled service team role
   - Body: `{ "person_id": "<uuid>" }`
   - Updates `service_teams.person_id`, sets status to 'confirmed'

### Frontend Dashboard Page

**Route:** `/dashboard/scheduling/needs`

#### Key Features

1. **Grouped View by Date → Team → Position**
   - Services grouped chronologically (ascending)
   - Within each date, positions grouped by volunteer team
   - Color-coded team indicators from `volunteer_teams.color`

2. **Urgency Highlighting**
   - ⚠️ Critical: ≤3 days until service (red border/badge)
   - 🚨 Urgent: 4-7 days (orange border/badge)
   - ℹ️ Soon: 8-14 days (no badge but visible)
   - Badge appears on date headers for services within 7 days

3. **Quick-Assign Inline Search**
   - "Assign" button per unfilled position
   - Opens modal with search input
   - Real-time person search as you type
   - Shows availability status: green badge = available, orange = busy
   - Click to assign instantly
   - Auto-refreshes dashboard after successful assignment

4. **Date Range Filter**
   - Start date and end date pickers at top of page
   - "Update" button applies filter
   - Default: next 4 weeks from today

5. **Empty States**
   - No scheduling needs found → friendly message with icon
   - Search no results → "No matches found"
   - Initial load before typing → prompt to start searching

---

## Files Modified/Created

### Frontend (Svelte)

1. **`web/src/routes/dashboard/scheduling/+layout.svelte`** (NEW)
   - Layout wrapper for scheduling sub-routes
   - Simple slot-based layout

2. **`web/src/routes/dashboard/scheduling/needs/+page.svelte`** (NEW, 11KB)
   - Main dashboard page component
   - State management: `schedulingData`, `loading`, search state
   - Functions: `loadSchedulingNeeds()`, `searchPeople()`, `assignPerson()`
   - Grouping logic for positions by team
   - Urgency calculation and color coding
   - Modal overlay for quick-assign flow

3. **`web/src/routes/dashboard/+layout.svelte`** (MODIFIED)
   - Added "Scheduling Needs" link to Ministry navigation section
   - Route: `/dashboard/scheduling/needs`
   - Icon: `alert-circle`
   - Module dependency: `services`

### Backend (Go)

No new files created. All required API handlers already exist in existing codebase:

- **`internal/services/service.go`**
  - `GetSchedulingNeeds()` → returns grouped unfilled positions
  - `SearchPeople()` → person search with availability check
  - `AssignPersonToRole()` → assigns person to service team role
  - Types: `UnfilledPosition`, `SchedulingNeedsResponse`, `SearchPersonResult`

- **`internal/services/handler.go`**
  - Handler wrappers for all three endpoints already implemented

- **`internal/router/router.go`** (Line 306-308)
  - Routes already registered:
    ```go
    r.Get("/api/scheduling/needs", servicesHandler.GetSchedulingNeeds)
    r.Get("/api/people/search", servicesHandler.SearchPeople)
    r.Put("/api/services/service-teams/{id}/assign", servicesHandler.AssignPersonToRole)
    ```

---

## Database Schema Used

The implementation relies on existing tables:

```sql
-- Services and their types
services (id, tenant_id, service_type_id, name, service_date, service_time, notes, status)
service_types (id, tenant_id, name, default_time, default_day, color, is_active)

-- Service team assignments (volunteer scheduling)
service_teams (id, service_id, person_id, role, status, notes)

-- Volunteer teams (role categories like "Ushers", "Greeters")
volunteer_teams (id, tenant_id, name, description, color, is_active)

-- People directory
people (id, tenant_id, first_name, last_name, email, ...)

-- Availability checking helper function
is_person_available(person_id, check_date) → boolean
```

---

## User Flow

### Viewing Scheduling Needs

1. Admin navigates to `/dashboard/scheduling/needs`
2. Dashboard loads unfilled positions for next 4 weeks (or custom date range)
3. Positions grouped by service date, then by volunteer team
4. Urgent positions highlighted with color-coded borders and badges

### Assigning a Volunteer

1. Click "Assign" button on an unfilled position card
2. Modal opens with search input
3. Type person's name or email
4. Results appear showing full name, email, and availability badge
5. Click result to assign instantly
6. Success toast appears
7. Dashboard refreshes automatically to show updated status

---

## Styling Details

- **Urgency colors:**
  - Critical (≤3 days): `border-red-500`, `bg-red-500/20`, red badge
  - Urgent (4-7 days): `border-orange-500`, `bg-orange-500/20`, orange badge
  
- **Team colors:** Applied as left border accent on position cards, matching `volunteer_teams.color`

- **Modal overlay:** Fixed z-index 50, backdrop blur, centered card layout

---

## Testing Notes

### Manual Test Cases

1. ✅ Load page with unfilled positions → displays grouped by date
2. ✅ Click "Assign" → modal opens
3. ✅ Type search query → results filter in real-time
4. ✅ Assign person → success toast, dashboard refreshes
5. ✅ Date range filter → applies and updates list
6. ✅ Urgency highlighting → red for ≤3 days, orange for 4-7 days
7. ✅ Empty state → shows when no unfilled positions exist

### API Test Cases

1. `GET /api/scheduling/needs` without params → returns next 4 weeks
2. `GET /api/scheduling/needs?start=2026-03-01&end=2026-03-31` → returns specified range
3. `GET /api/people/search?q=john` → returns matching people with availability
4. `PUT /api/services/service-teams/{id}/assign { person_id: "..." }` → assigns and updates status

---

## Known Limitations & Future Enhancements

### Current Limitations

1. **No conflict detection on assign** - Does not check if person is already assigned to another team for same service date
2. **No availability validation** - Assigns even if person has blockout/absence configured
3. **No email notification** - Person doesn't get notified of new assignment
4. **Limited search scope** - Only searches name/email, not phone or other fields

### Recommended Future Enhancements

1. Add conflict detection before assign (check `service_teams` for same date)
2. Integrate with volunteer availability system (blockouts/absences)
3. Send email/SMS notification on assignment via existing notification module
4. Expand search to include phone, tags, skills matching role requirements
5. Add "bulk assign" feature from team roster view
6. Export scheduling needs report (CSV/PDF)

---

## GitHub Issue Closure

This implementation fulfills all requirements from **Issue #65**:

> Admin view of all open volunteer positions across upcoming services

✅ Route: `/dashboard/scheduling/needs`  
✅ Shows next 4 weeks by default, date range filterable  
✅ Groups by service date → team → position  
✅ Quick-assign inline with search and assign without leaving page  
✅ Urgency highlighting for services in next 7 days  

**Closing command:**
```bash
gh issue close 65 --repo warpapaya/Pews -c "Implemented scheduling needs dashboard with quick-assign feature. Backend API endpoints complete, frontend dashboard functional. See projects/pews/unfilled-needs-dashboard.md for full documentation."
```

---

## Related Issues

- **Issue #74** (CCLI metadata) - separate task
- **Issue #75** (song key transposition) - separate task  
- Backend infrastructure for scheduling already in place from prior work
