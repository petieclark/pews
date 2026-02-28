# CHANGES-62: Volunteer Blockout System Implementation

## Summary
Implemented volunteer blockout system allowing volunteers to mark unavailable dates (both one-time and recurring). This is the foundation for conflict detection (#67) and auto-scheduling (#66).

## Files Modified

### 1. Migration: `migrations/045_volunteer_blockouts.sql` (NEW)
Created new table with schema:
```sql
CREATE TABLE volunteer_blockouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    person_id UUID NOT NULL REFERENCES people(id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT,
    is_recurring BOOLEAN DEFAULT false,
    day_of_week INT,  -- 0=Sun..6=Sat
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

Indexes:
- `idx_vb_person` - person_id lookup
- `idx_vb_tenant` - tenant isolation
- `idx_vb_dates` - date range queries
- `idx_vb_recurring` - recurring blockout optimization

RLS policy: `vb_isolation_policy` ensures tenant data isolation.

### 2. Model: `internal/teams/model.go` (MODIFIED)
Added two new types:

```go
type VolunteerBlockout struct {
    ID          string    `json:"id"`
    TenantID    string    `json:"tenant_id"`
    PersonID    string    `json:"person_id"`
    StartDate   string    `json:"start_date"`
    EndDate     string    `json:"end_date"`
    Reason      *string   `json:"reason,omitempty"`
    IsRecurring bool      `json:"is_recurring"`
    DayOfWeek   *int      `json:"day_of_week,omitempty"` // 0=Sun..6=Sat
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type BlockoutMatch struct {
    Blockout    VolunteerBlockout `json:"blockout"`
    IsRecurring bool              `json:"is_recurring_match"` // true if day_of_week match, false if date range
}
```

### 3. Service: `internal/teams/service.go` (MODIFIED)
Added six new methods:

**Query Methods:**
- `GetPersonBlockouts(ctx, tenantID, personID)` - List all blockouts for a volunteer
- `IsVolunteerBlocked(ctx, tenantID, personID, checkDate)` - Check if blocked on specific date
  - Checks both one-time (date range) and recurring (day_of_week) blockouts
  - Returns BlockoutMatch with info about which type matched

**CRUD Methods:**
- `CreateBlockout()` - Create new blockout
- `UpdateBlockout()` - Update existing blockout  
- `DeleteBlockout()` - Delete blockout

**Conflict Detection:**
- `CheckAssignmentsForConflicts(ctx, tenantID, serviceDate, personIDs)` - Check multiple volunteers for conflicts on a given date
  - Returns map of person_id → matching blockout for those who are blocked

### 4. Handler: `internal/teams/handler.go` (MODIFIED)
Added four HTTP handlers:

**GET /api/teams/members/:personId/blockouts**
- Lists all blockouts for a volunteer
- Returns JSON: `{ "blockouts": [...] }`

**POST /api/teams/members/:personId/blockouts**
- Creates new blockout
- Request body: `createBlockoutRequest` with person_id, start_date, end_date, reason (optional), is_recurring, day_of_week (0-6)
- Validates recurring flag + day_of_week consistency
- Returns 201 Created

**PUT /api/teams/members/blockouts/:blockoutId**
- Updates existing blockout
- Same fields as create
- Returns updated blockout

**DELETE /api/teams/members/blockouts/:blockoutId**
- Deletes a blockout
- Returns 204 No Content on success, or error if not found

### 5. Router: `internal/router/router.go` (MODIFIED)
Registered four new routes under protected group (auth required):
```go
// Volunteer Blockouts
r.Get("/api/teams/members/{personId}/blockouts", teamsHandler.GetPersonBlockouts)
r.Post("/api/teams/members/{personId}/blockouts", teamsHandler.CreateBlockout)
r.Put("/api/teams/members/blockouts/{blockoutId}", teamsHandler.UpdateBlockout)
r.Delete("/api/teams/members/blockouts/{blockoutId}", teamsHandler.DeleteBlockout)
```

## Business Logic: isVolunteerBlocked

The core helper function implements the following logic:

1. **First**, query for non-recurring blockouts where `check_date >= start_date AND check_date <= end_date`
2. **If found**, return BlockoutMatch with `IsRecurring = false`
3. **Otherwise**, query for recurring blockouts where `EXTRACT(DOW FROM check_date) = day_of_week`
4. **If found**, return BlockoutMatch with `IsRecurring = true`
5. **If neither matches**, return nil (not blocked)

This allows volunteers to:
- Mark specific date ranges as unavailable (vacation, medical appointments, etc.)
- Mark recurring unavailability (e.g., every Sunday for church service, every Friday evenings)

## Integration Points

### For Scheduling Endpoints
When saving service assignments (`SaveServiceAssignments`), call:
```go
conflicts, err := service.CheckAssignmentsForConflicts(ctx, tenantID, serviceDate, personIDs)
if len(conflicts) > 0 {
    // Return 409 Conflict with list of blocked volunteers
    return http.StatusConflict
}
```

### For Volunteer Profile Page (Frontend)
The frontend should:
1. Fetch blockouts via `GET /api/teams/members/:personId/blockouts` on page load
2. Display as a list with start/end dates, reason, and recurring indicator
3. Provide form to create new blockout (date range picker + optional reason + recurring toggle)
4. Allow editing/deletion of existing blockouts
5. Show visual indication when admin overrides a blockout

## Testing Recommendations

1. **Create one-time blockout**: Volunteer marks March 15-20 unavailable → should conflict with any service on those dates
2. **Create recurring blockout**: Volunteer blocks every Sunday (day_of_week=0) → should conflict with all Sunday services
3. **Check date boundary**: Blockout Mar 15-20, check Mar 14 and Mar 21 → should NOT be blocked
4. **Conflict detection**: Try to assign blocked volunteer to service on blocked date → should return 409
5. **RLS isolation**: Verify tenant A cannot see tenant B's blockouts

## Next Steps (Not Implemented)

- Frontend UI for blockout management (date picker, recurring toggle)
- Integration with `SaveServiceAssignments` to return 409 on conflict
- Admin override capability (flag in DB or separate table)
- Email notification when blockout created that conflicts with existing assignment

## GitHub Issue Closure

To close issue #62:
```bash
gh issue close 62 --repo warpapaya/Pews -c "Implemented volunteer blockout system with API endpoints and conflict detection logic."
```

---

**Implementation Date**: 2026-02-28  
**Status**: Backend complete, pending frontend integration and DB migration execution  
**Tested**: Go compilation successful (pre-existing errors in other packages unrelated to this change)
