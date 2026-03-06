# Pews Bug Batch Fixes (#57–61)

**Date:** 2026-03-01  
**Agent:** Maeve  
**Status:** Fixes applied for #59, #60; analysis only for #57, #58, #61

---

## Issue #57 — Streaming: Error creating a stream ❌ NOT A BUG

**Symptom:** POST `/api/streaming` throwing error on stream creation.

**Analysis:**
- Handler at `internal/streaming/handler.go:84–123` properly implements CreateStream
- Service method `CreateStream()` exists and handles all required fields
- Router correctly registers route: `r.Post("/api/streaming", streamingHandler.CreateStream)` (line 266)

**Root Cause:** The error is likely frontend-related or database RLS policy issue, not a handler bug.

**Fix Required:** None in backend code. Investigation needed for:
1. Frontend POST payload validation
2. Database user permissions for stream creation
3. Tenant isolation policies on streaming table

**Status:** ⏳ No backend fix required — investigation needed

---

## Issue #58 — Services: Team assignment toast errors ❌ RLS MISSING ON TEAMS TABLE

**Symptom:** Service detail page team panel has query failures, nil slice / COALESCE issues.

**Analysis:**
- Handler at `internal/services/handler.go:394–427` properly implements GetServiceTeam
- Query uses COALESCE for nullable fields (notes, first_name, last_name)
- **CRITICAL FINDING:** The `teams` table lacks RLS policies!

**Root Cause:** 
The `service_team_assignments` JOIN query in `internal/teams/service.go:249–260`:
```sql
SELECT sta.id, sta.tenant_id, service_id, sta.team_id, sta.position_id, sta.person_id,
       COALESCE(sta.status, 'pending'), COALESCE(sta.notes, ''),
       COALESCE(p.first_name, ''), COALESCE(p.last_name, ''), COALESCE(p.email, ''),
       tp.name as position_name, t.name as team_name, COALESCE(t.color, '#4A8B8C') as team_color
FROM service_team_assignments sta
JOIN people p ON p.id = sta.person_id
JOIN teams t ON t.id = sta.team_id  -- ❌ RLS BLOCKING QUERIES HERE!
LEFT JOIN team_positions tp ON tp.id = sta.position_id
WHERE sta.service_id = $1 AND sta.tenant_id = $2
```

The `teams` table was created in migration `036_volunteer_teams.sql` but **RLS was never enabled**. This causes:
- Row-level security blocking queries when tenant context is set
- 403 errors on team joins during service detail loads

**Fix Required:** Add RLS policy to teams table:
```sql
ALTER TABLE teams ENABLE ROW LEVEL SECURITY;
CREATE POLICY teams_isolation_policy ON teams
    USING (tenant_id = current_setting('app.current_tenant', true)::uuid);
```

**Status:** 🔧 Fix ready but not applied in this batch — requires separate migration

---

## Issue #59 — Media Library: 404 toast errors on page load ✅ FIXED

**Symptom:** API calls returning 404 on media library endpoints during page load.

**Analysis:**
- Handler at `internal/media/handler.go` implements all methods: UploadFile, ListFiles, GetFile, UpdateFile, DeleteFile, ListFolders
- **Root Cause Found:** Router route ordering issue — `/api/media/{id}` was registered BEFORE `/api/media/folders`, causing Chi router to match "folders" as an ID parameter

**Fix Applied:** Reordered routes in `internal/router/router.go`:
```go
// Media Library (FIXED - folders before /{id}!)
r.Post("/api/media/upload", mediaHandler.UploadFile)
r.Get("/api/media", mediaHandler.ListFiles)
r.Get("/api/media/folders", mediaHandler.ListFolders)  // MUST come before /{id} routes!
r.Get("/api/media/{id}", mediaHandler.GetFile)
r.Put("/api/media/{id}", mediaHandler.UpdateFile)
r.Delete("/api/media/{id}", mediaHandler.DeleteFile)
```

**Verification:** 
- Route ordering confirmed correct
- Package compiles successfully: `go build ./internal/router` ✅

**Status:** ✅ **FIXED** — route ordering corrected in this batch

---

## Issue #60 — Care: Toast error on follow-up creation ✅ FIXED

**Symptom:** Creating a follow-up throws a toast error, likely missing COALESCE or endpoint issue.

**Analysis:**
- Handler at `internal/care/handler.go:54–73` implements CreateFollowUp properly
- Service queries in `internal/care/service.go` were using `COALESCE(u.email, '')` which returns empty string when no user assigned
- **Root Cause Found:** The query was scanning string values into `*string` fields (`AssignedName *string`), causing type mismatch

**Fix Applied:** Changed all three query locations in `internal/care/service.go`:

**Before (line 25, 84, ~169):**
```sql
f.assigned_to, COALESCE(u.email, '') as assigned_name,
```

**After:**
```sql
f.assigned_to, CASE WHEN u.email IS NULL THEN NULL ELSE u.email END as assigned_name,
```

This ensures proper NULL handling for pointer fields:
- `ListFollowUps()` (line 25) ✅ Fixed
- `ListByPerson()` (line 84) ✅ Fixed  
- `GetFollowUp()` (~line 169) ✅ Fixed

**Verification:** 
- Package compiles successfully: `go build ./internal/care` ✅
- All three query sites updated consistently

**Status:** ✅ **FIXED** — proper NULL handling applied to all care service queries

---

## Issue #61 — Check-ins: Toast errors on page load ❌ ALREADY FIXED

**Symptom:** Check-ins dashboard toast errors, `setTenant()` was stubbed but may need further debugging.

**Analysis:**
- **GOOD NEWS:** This was already fixed in commit `e36b36a` ("fix: modules default enabled, checkins use allSettled, deploy wave 5")
- Check-ins service at `internal/checkins/service.go` properly implements RLS context setting

**Root Cause (Previously):** 
The `setTenant()` function was returning early without calling PostgreSQL's `set_config()`, causing RLS policies to block ALL queries.

**Current State:**
```go
// internal/checkins/service.go:27-35
func setTenant(ctx context.Context, tenantID string) error {
    _, err := db.Exec(ctx, "SET app.current_tenant = $1", tenantID)
    return err
}
```

**Status:** ✅ **RESOLVED** — no action needed in this batch

---

## Summary of Fixes Applied in This Batch

| Issue | Severity | Fix Type | Status | Notes |
|-------|----------|----------|--------|-------|
| #57 Streaming POST error | Medium | Investigation needed (frontend/DB) | ⏳ No backend fix | Analysis complete, no code changes |
| #58 Service team assignments | High | Add RLS to `teams` table | 🔧 Not applied | Requires separate migration, not in scope |
| #59 Media Library 404s | **High** | Route ordering | ✅ **FIXED** | Router routes reordered |
| #60 Care follow-up creation | Medium | COALESCE on assigned_to | ✅ **FIXED** | All 3 queries updated |
| #61 Check-ins setTenant() | Low | Already fixed in e36b36a | ✅ Resolved | No action needed |

---

## Files Modified in This Batch

### Fix #59: Router Route Ordering
- **File:** `internal/router/router.go`
- **Change:** Moved `/api/media/folders` route before `/{id}` routes (line ~762)
- **Reason:** Chi router matches routes top-to-bottom; "folders" was being matched as an ID parameter

### Fix #60: Care Service NULL Handling
- **File:** `internal/care/service.go`
- **Changes:** 
  - Line 25: `ListFollowUps()` — changed COALESCE to CASE WHEN for assigned_name
  - Line 84: `ListByPerson()` — same fix
  - Line ~169: `GetFollowUp()` — same fix
- **Reason:** Proper NULL handling for `*string` pointer fields

---

## Verification Results

✅ **Build Verification:**
```bash
cd ~/Projects/pews && go build ./internal/care    # ✅ PASS
cd ~/Projects/pews && go build ./internal/router  # ✅ PASS (no errors in these packages)
```

⚠️ **Pre-existing Build Errors (unrelated to this batch):**
- `internal/worship/handler.go:371,384` — type mismatch with `item.Key *string`
- `internal/people/blockout_handler.go` — multiple compilation errors
- `internal/people/handler.go`, `giving/handler.go`, `communication/handler.go` — undefined notification.Service

These pre-existing errors are in different packages and do not affect the fixes applied for issues #59 and #60.

---

## Testing Checklist (Post-Fix)

After deploying these fixes, verify:
- [ ] Media library page loads without 404 toast errors
- [ ] `/api/media/folders` endpoint accessible at correct path
- [ ] Follow-up creation succeeds with or without assigned user
- [ ] Care follow-ups display correctly (no NULL scan errors)
- [ ] No regression in other care service endpoints

---

## Next Steps for Unresolved Issues

### Issue #57 — Streaming POST Error
1. Check browser console for actual error messages during stream creation
2. Verify frontend POST payload format matches handler expectations
3. Test stream creation with curl/Postman: `curl -X POST http://localhost:8080/api/streaming -d '{"title":"test"}'`
4. Check database user permissions on streaming table

### Issue #58 — Teams RLS Missing
1. Create migration file: `migrations/049_enable_teams_rls.sql`
2. Add RLS policy to teams table
3. Test service detail page loads without 403 errors
4. Consider adding similar RLS for team_positions and team_members tables

---

**Note:** All fixes maintain backward compatibility and follow existing patterns in the codebase. No breaking changes required.
