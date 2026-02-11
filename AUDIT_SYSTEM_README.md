# Audit Log & Security Dashboard

## Overview
This feature adds comprehensive audit logging and security monitoring to Pews.

## What Was Built

### 1. Database Schema (`internal/database/migrations/014_audit_logs.sql`)
- **`audit_logs` table**: Immutable log of all admin actions
  - Fields: timestamp, user_id, action, entity_type, entity_id, ip_address, user_agent, old_value (json), new_value (json)
  - RLS policies prevent UPDATE/DELETE even by admins
  - Indexes on tenant_id, user_id, timestamp, entity_type, action
  
- **`failed_login_attempts` table**: Tracks failed login attempts for security monitoring
  
- **`user_sessions` table**: Tracks active user sessions
  
- **User table enhancements**: Added `password_changed_at`, `last_login_at`, `last_login_ip` fields

### 2. Backend (`internal/audit/`)
- **Service** (`service.go`): Core audit logging functionality
  - `Log()` - Create immutable audit log entries
  - `GetLogs()` - Retrieve paginated audit logs with filters
  - `LogFailedLogin()` - Track failed login attempts
  - `CreateSession()`, `EndSession()` - Session management
  - `GetSecurityDashboard()` - Aggregate security metrics
  
- **Handler** (`handler.go`): HTTP endpoints
  - `GET /api/audit/logs` - Paginated audit log (with filters: user_id, action, entity_type)
  - `GET /api/audit/logs/user/:id` - Logs for specific user
  - `GET /api/audit/security` - Security dashboard data
  - `GET /api/audit/export` - Export logs as CSV
  
- **Models** (`models.go`): Data structures and action constants

### 3. Audit Middleware (`internal/middleware/audit.go`)
- Automatically logs ALL mutating requests (POST/PUT/DELETE)
- Captures: method, path, user, IP, user agent, request/response data
- Runs asynchronously to avoid blocking requests
- Smart entity extraction from URLs

### 4. Enhanced Auth Handler (`internal/auth/handler.go`)
- Logs successful logins with session creation
- Logs failed login attempts
- Tracks login IP and user agent
- Logs logout events

### 5. Frontend (`web/src/routes/dashboard/settings/audit/+page.svelte`)
- **Audit Log Tab**:
  - Searchable/filterable table of all audit events
  - Filters: User ID, Action Type, Entity Type
  - Pagination
  - CSV export button
  - View old/new values for each change
  
- **Security Dashboard Tab**:
  - Active sessions count
  - Failed login attempts (last 24h)
  - Users without 2FA count
  - Recent failed login attempts table
  - Unusual activity alerts (login from new IP)
  - User password status (days since last change, highlights >90 days)

## Action Types Logged
- `auth.login` - Successful login
- `auth.login_failed` - Failed login attempt
- `auth.logout` - User logout
- `create` - Entity creation (POST requests)
- `update` - Entity modification (PUT requests)
- `delete` - Entity deletion (DELETE requests)
- `settings.change` - Settings modifications
- `module.enable` - Module activation
- `module.disable` - Module deactivation

## Testing

### 1. Start the system:
```bash
docker compose up -d
```

### 2. Register/login as admin:
```bash
# Register
curl -X POST http://localhost:8190/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"tenant_name": "Test Church", "email": "admin@test.com", "password": "password123"}'

# Login  
curl -X POST http://localhost:8190/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"tenant_slug": "test-church", "email": "admin@test.com", "password": "password123"}'
```

### 3. Perform actions (creates audit logs):
```bash
TOKEN="your-token-here"

# Create a person
curl -X POST http://localhost:8190/api/people \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"first_name": "John", "last_name": "Doe", "email": "john@test.com"}'

# Update a person
curl -X PUT http://localhost:8190/api/people/{id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"first_name": "Jane", "last_name": "Doe"}'
```

### 4. View audit logs:
```bash
# Get audit logs
curl -X GET "http://localhost:8190/api/audit/logs?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"

# Get security dashboard
curl -X GET "http://localhost:8190/api/audit/security" \
  -H "Authorization: Bearer $TOKEN"

# Export as CSV
curl -X GET "http://localhost:8190/api/audit/export" \
  -H "Authorization: Bearer $TOKEN" -o audit_logs.csv
```

### 5. Frontend access:
Navigate to: `http://localhost:5273/dashboard/settings/audit`

## Security Features
1. **Immutable Logs**: Audit log entries cannot be edited or deleted, even by admins
2. **Comprehensive Tracking**: All admin actions are logged with full context
3. **IP & User Agent Logging**: Track where actions originated
4. **Failed Login Monitoring**: Detect potential security threats
5. **Session Management**: Track active user sessions
6. **Unusual Activity Detection**: Flag logins from new IPs
7. **Password Age Tracking**: Identify users with old passwords

## Database Policies
- SELECT: Tenant-scoped (users can only see their tenant's logs)
- INSERT: Allowed (system creates logs)
- UPDATE: Blocked (immutable)
- DELETE: Blocked (immutable)

## Performance Considerations
- Audit logging runs asynchronously to avoid blocking requests
- Indexes on frequently-queried columns (timestamp, user_id, action, entity_type)
- Pagination for large log sets

## Future Enhancements
- Email alerts for suspicious activity
- Audit log retention policies
- Advanced filtering (date ranges, IP ranges)
- Audit log archiving
- Compliance report generation (GDPR, etc.)
