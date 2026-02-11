# Worship Planning / Service Builder Feature

## Overview
Complete worship service planning tool for worship teams to plan service order, assign songs, schedule band members, and share notes.

## What Was Built

### 1. Database Migration (014_worship_planning.sql)
- **service_plans** table
  - Links to services and tracks who created the plan
  - Status field (draft/published)
  - Overall notes for the plan
  - Full RLS tenant isolation

- **service_plan_items** table
  - Individual items in the plan (songs, scripture, prayer, etc.)
  - Order tracking for drag-and-drop reordering
  - Duration tracking for each item
  - Assignment to team members (users)
  - Per-item notes
  - Optional link to song library
  - Full RLS tenant isolation

### 2. Backend API (`internal/worship/`)
Complete REST API with these endpoints:

- `GET /api/worship/plans` — List all plans for tenant
- `POST /api/worship/plans` — Create new plan for a service
- `GET /api/worship/plans/:id` — Get plan detail with all items
- `PUT /api/worship/plans/:id` — Update plan notes
- `POST /api/worship/plans/:id/publish` — Publish plan (makes visible to team)
- `GET /api/worship/plans/:id/export` — Export as printable HTML
- `POST /api/worship/plans/:id/items` — Add item to plan
- `PUT /api/worship/plans/:id/items/:itemId` — Update item (including reorder)
- `DELETE /api/worship/plans/:id/items/:itemId` — Remove item from plan

**Architecture:**
- `model.go` — Data structures
- `service.go` — Business logic and database operations
- `handler.go` — HTTP handlers with auth middleware

### 3. Frontend Dashboard (`/dashboard/worship`)

#### Main List View (`+page.svelte`)
- List all service plans
- Status badges (draft/published)
- Service selector for creating new plans
- Quick view of item counts
- Click to edit functionality

#### Plan Builder (`[id]/+page.svelte`)
**Features:**
- Service selector dropdown
- Drag-and-drop item reordering
- Add items with multiple types:
  - Song (with song library integration)
  - Scripture Reading
  - Prayer
  - Announcement
  - Video
  - Other
- Duration tracker showing total service time
- Assign team members to each item
- Per-item notes (e.g., "key of G", "start soft")
- Overall plan notes
- Publish/share workflow
- Export to printable HTML

**UX Details:**
- Real-time duration calculation
- Drag handles on each item for reordering
- Modal dialogs for adding items
- Song library dropdown for quick selection
- User/people dropdown for assignments
- Status indicators

### 4. Integration Points
- Integrated with existing `services` table
- Uses existing `songs` table for song library
- Uses existing `users` table for assignments
- Follows Pews authentication & authorization patterns
- Tenant isolation via RLS policies

## Testing

### 1. Start Services
```bash
docker compose up -d
```

Backend: http://localhost:8190
Frontend: http://localhost:5273

### 2. Create Test Data
1. Log in to the dashboard
2. Navigate to Services → Create a service
3. Navigate to Worship Planning
4. Click "New Service Plan"
5. Select the service you created
6. Add items (songs, prayers, etc.)
7. Drag items to reorder
8. Assign team members
9. Add notes to items
10. Publish the plan
11. Export/print the plan

### 3. API Testing
```bash
# List plans (requires auth token)
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8190/api/worship/plans

# Get specific plan
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8190/api/worship/plans/:id

# Export plan (printable HTML)
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8190/api/worship/plans/:id/export
```

## Technical Decisions

1. **Separate from service_items**: Created dedicated `service_plans` and `service_plan_items` tables separate from the existing `service_items` table to allow for a distinct planning workflow that doesn't interfere with actual service execution data.

2. **Drag-and-drop**: Used native HTML5 drag-and-drop API with `item_order` field for persistent ordering.

3. **Duration tracking**: Calculated client-side in real-time, stored per-item for flexibility.

4. **Assignment**: Links to `users` table rather than free-text, enabling better team management.

5. **Export**: Server-side HTML generation for print-friendly output, avoiding client-side PDF dependencies.

6. **Status workflow**: Simple draft → published flow to control visibility.

## Migration Note
The migration file is numbered `014_worship_planning.sql`. When merging to main, this may need renumbering if other feature branches add migrations first.

## Branch
This feature is committed to: `feat/worship-planning`

**Do NOT merge to main yet** — awaiting code review and testing.

## Files Changed
- `cmd/pews/main.go` — Added worship service/handler initialization
- `internal/router/router.go` — Registered worship API routes
- `internal/database/migrations/014_worship_planning.sql` — Database schema
- `internal/worship/handler.go` — HTTP handlers (378 lines)
- `internal/worship/model.go` — Data models (32 lines)
- `internal/worship/service.go` — Business logic (301 lines)
- `web/src/routes/dashboard/worship/+page.svelte` — List view (237 lines)
- `web/src/routes/dashboard/worship/[id]/+page.svelte` — Plan builder (471 lines)

**Total:** 8 files changed, 1,515 insertions

## Next Steps
1. Code review
2. User acceptance testing with worship team
3. Add print CSS optimization
4. Consider PDF export library integration
5. Add email notification when plan is published
6. Consider mobile app view for band members
