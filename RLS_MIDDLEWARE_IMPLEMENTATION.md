# RLS Middleware Implementation Summary

**Branch:** `feat/rls-middleware`  
**Issue:** #11 - Proper Tenant Context Injection  
**Status:** ✅ Complete - Ready for Testing

## What Was Implemented

### 1. Chi Middleware (`internal/middleware/tenant_rls.go`)
- Extracts `tenant_id` from JWT claims (set by auth middleware)
- Stores tenant ID in request context
- Applied to all protected routes in the router

### 2. Database Connection Pool Hook (`internal/database/postgres.go`)
- Implemented `BeforeAcquire` hook using pgx
- Automatically sets `app.current_tenant_id` session variable when acquiring a connection
- Reads tenant ID from request context
- Transparent to application code - no manual calls needed

### 3. Helper for Public Endpoints (`internal/database/tenant_helper.go`)
- `SetTenantContext()` function for endpoints without auth
- Used by public streaming endpoints (chat, join/leave stream)
- Public endpoints fetch stream first to get tenant_id, then call helper

### 4. Service File Cleanup
Removed manual `set_config` calls from:
- `internal/people/service.go` (16 calls removed)
- `internal/groups/service.go` (10 calls removed)
- `internal/services/service.go` (21 calls removed)
- `internal/checkins/service.go` (1 call removed)
- `internal/communication/service.go` (1 call removed)
- `internal/streaming/service.go` (partial - public endpoints now use helper)

### 5. RLS Re-enablement Migration (`012_ensure_rls_enabled.sql`)
- Ensures RLS is enabled on all 50+ tables
- Idempotent - safe to run multiple times
- Covers all modules: people, groups, services, checkins, communication, streaming, giving

### 6. Router Integration
- Added `middleware.TenantRLS` to protected route group
- Applied after auth middleware (so claims are available)
- Applied before all handlers (so tenant context is set)

## How It Works

### Authenticated Requests
```
1. Request → Auth Middleware (extracts JWT, stores claims in context)
2. Request → TenantRLS Middleware (extracts tenant_id from claims, stores in context)
3. Handler acquires DB connection → BeforeAcquire hook sets session variable
4. Query runs → RLS policies enforce automatically
```

### Public Requests (Streaming Chat, etc.)
```
1. Handler receives request (no auth)
2. Handler fetches stream to get tenant_id
3. Handler calls database.SetTenantContext(ctx, pool, tenantID)
4. Subsequent queries have tenant context set
```

## Code Changes Summary

**Files Created:**
- `internal/middleware/tenant_rls.go` (44 lines)
- `internal/database/tenant_helper.go` (20 lines)
- `internal/database/migrations/012_ensure_rls_enabled.sql` (58 lines)

**Files Modified:**
- `internal/database/postgres.go` (added BeforeAcquire hook)
- `internal/router/router.go` (added TenantRLS middleware)
- 6 service files (removed manual set_config calls)

**Net Change:** +149 insertions, -238 deletions (simplified!)

## Testing Instructions

### Build & Start
```bash
cd ~/Projects/pews
docker build -t pews-backend:latest .
docker compose up -d
```

### Test Tenant Isolation
1. **Register Two Churches:**
   ```bash
   # Church A
   curl -X POST http://localhost:8080/api/auth/register \
     -H "Content-Type: application/json" \
     -d '{"church_name":"Church A","admin_email":"admin@church-a.com","admin_password":"password123"}'
   
   # Church B
   curl -X POST http://localhost:8080/api/auth/register \
     -H "Content-Type: application/json" \
     -d '{"church_name":"Church B","admin_email":"admin@church-b.com","admin_password":"password123"}'
   ```

2. **Login as Church A:**
   ```bash
   curl -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@church-a.com","password":"password123"}'
   
   # Save the token from response
   export TOKEN_A="<token>"
   ```

3. **Create Data in Church A:**
   ```bash
   # Create a person
   curl -X POST http://localhost:8080/api/people \
     -H "Authorization: Bearer $TOKEN_A" \
     -H "Content-Type: application/json" \
     -d '{"first_name":"John","last_name":"Doe","email":"john@church-a.com"}'
   
   # Create a group
   curl -X POST http://localhost:8080/api/groups \
     -H "Authorization: Bearer $TOKEN_A" \
     -H "Content-Type: application/json" \
     -d '{"name":"Youth Group","description":"Weekly youth meetings"}'
   ```

4. **Login as Church B:**
   ```bash
   curl -X POST http://localhost:8080/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@church-b.com","password":"password123"}'
   
   export TOKEN_B="<token>"
   ```

5. **Verify Isolation:**
   ```bash
   # Church B should see 0 people (not Church A's data)
   curl http://localhost:8080/api/people \
     -H "Authorization: Bearer $TOKEN_B"
   
   # Should return: {"people":[],"total":0,"page":1,"limit":50}
   
   # Church B should see 0 groups
   curl http://localhost:8080/api/groups \
     -H "Authorization: Bearer $TOKEN_B"
   ```

6. **Verify Church A Still Has Data:**
   ```bash
   curl http://localhost:8080/api/people \
     -H "Authorization: Bearer $TOKEN_A"
   
   # Should return John Doe
   ```

### Test All Modules
Repeat the above pattern for:
- Groups (`/api/groups`)
- Services (`/api/services`)
- Giving (`/api/giving/funds`, `/api/giving/donations`)
- Check-ins (`/api/checkins/events`)
- Communication (`/api/communication/templates`)
- Streaming (`/api/streaming`)

All should respect tenant isolation automatically.

### Test Public Streaming Endpoints
```bash
# Create a stream as Church A
curl -X POST http://localhost:8080/api/streaming \
  -H "Authorization: Bearer $TOKEN_A" \
  -H "Content-Type: application/json" \
  -d '{"title":"Sunday Service","stream_url":"rtmp://example.com/live","stream_key":"abc123"}'

# Get the stream ID from response
export STREAM_ID="<id>"

# Public access should work (no auth required)
curl http://localhost:8080/api/streaming/watch/$STREAM_ID

# Send chat message (public endpoint)
curl -X POST http://localhost:8080/api/streaming/$STREAM_ID/chat \
  -H "Content-Type: application/json" \
  -d '{"guest_name":"Visitor","message":"Hello!"}'

# Verify RLS allows chat for this stream
curl http://localhost:8080/api/streaming/$STREAM_ID/chat
```

## Expected Behavior

### ✅ Success Indicators
- All authenticated endpoints return data only for the logged-in tenant
- Cross-tenant data is completely invisible (not just unauthorized)
- Public streaming endpoints work without auth
- No "permission denied" errors (RLS returns empty results, not errors)
- All modules work normally

### ❌ Failure Indicators
- "no rows" errors (means RLS is blocking but tenant context wasn't set)
- "permission denied" errors (means RLS policy is too restrictive)
- Cross-tenant data leakage (Church A sees Church B's data)
- Public endpoints don't work

## Rollback Plan

If issues occur in production:

```bash
# Temporarily disable RLS on all tables
psql $DATABASE_URL << 'EOF'
-- Disable RLS (ONLY IN EMERGENCY)
ALTER TABLE people DISABLE ROW LEVEL SECURITY;
ALTER TABLE groups DISABLE ROW LEVEL SECURITY;
-- ... etc for all tables
EOF
```

Then revert the commit:
```bash
git revert 57ea953
git push origin main
```

## Next Steps

1. ✅ Merge to main after testing confirms isolation works
2. Deploy to staging/production
3. Monitor logs for "failed to set tenant context" warnings
4. Consider adding metrics/observability for RLS enforcement
5. Update developer documentation with new architecture

## Notes

- The middleware approach is cleaner and more maintainable than per-query set_config
- BeforeAcquire hook ensures tenant context is ALWAYS set (no way to forget)
- Public endpoints remain a special case but use a clearly marked helper
- Migration 012 is safe to run - it's idempotent and only enables RLS

## Files to Review

Key files for code review:
- `internal/middleware/tenant_rls.go` - Main middleware logic
- `internal/database/postgres.go` - BeforeAcquire hook
- `internal/router/router.go` - Middleware application
- `internal/database/migrations/012_ensure_rls_enabled.sql` - RLS enablement
