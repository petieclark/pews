# Pews Backup & Restore System

## Overview
Automated database backup and one-click restore system for church data safety, implemented in the `feat/backup-restore` branch.

## Features Implemented

### Backend (`internal/backup/`)

#### Service Layer (`service.go`)
- **CreateBackup()** - Creates tenant-scoped SQL backups
  - Queries all tables with `tenant_id` column
  - Exports data as SQL INSERT statements
  - Compresses to `.sql.gz` format
  - Filename format: `{tenant_slug}_{timestamp}.sql.gz`
  
- **ListBackups()** - Lists all backups for a tenant
  - Filters by tenant slug
  - Sorts by creation time (newest first)
  - Returns metadata (size, date, filename)

- **RestoreBackup()** - Restores from backup
  - Creates automatic safety backup before restore
  - Deletes existing tenant data
  - Executes backup SQL in transaction
  - Rollback on failure

- **DeleteBackup()** - Removes backup files
  - Security check: filename must match tenant slug

- **DownloadBackup()** - Provides file download stream
  - Returns `io.ReadCloser` for HTTP response

- **CleanupOldBackups()** - Retention policy
  - Auto-deletes backups older than 30 days
  - Keeps pre-restore safety backups forever

#### API Endpoints (`handler.go`)
- `POST /api/admin/backup` - Trigger manual backup
- `GET /api/admin/backups` - List available backups
- `POST /api/admin/restore/{filename}` - Restore from backup
- `DELETE /api/admin/backups/{filename}` - Delete backup
- `GET /api/admin/backups/{filename}/download` - Download backup file

### Frontend (`/dashboard/settings/backups`)

#### Features
- **Create Backup** button with loading state
- **Backups table** showing:
  - Creation date/time
  - Filename (with tenant prefix)
  - File size (formatted in KB/MB)
  - Action buttons (Download, Restore, Delete)
- **Restore dialog** with:
  - Big warning about data replacement
  - List of what will happen
  - Confirmation input (must type "RESTORE")
  - Cancel/Confirm buttons
- **Safety notice** section explaining:
  - Automatic safety backups
  - 30-day retention
  - Download for permanent storage

#### UX Considerations
- All operations show loading states
- Success/error messages displayed prominently
- Restore requires explicit "RESTORE" confirmation
- Delete requires browser confirm dialog

### Docker Integration

Added backup volume to `docker-compose.yml`:
```yaml
volumes:
  - backup_data:/backups
```

Backups persist across container restarts.

## Security

1. **Tenant Isolation**
   - All operations scoped to authenticated tenant
   - Filename validation prevents cross-tenant access
   - Backups only contain tenant's own data

2. **Restore Safety**
   - Automatic pre-restore backup created
   - Confirmation dialog with explicit text input
   - Transaction-based restore (atomic)

3. **Authentication**
   - All endpoints require valid JWT token
   - Tenant ID extracted from authenticated context

## File Structure

```
internal/backup/
├── model.go      # Backup struct and response types
├── service.go    # Core backup logic
└── handler.go    # HTTP handlers

web/src/routes/dashboard/settings/
├── +page.svelte          # Main settings (with "Manage Backups" link)
└── backups/+page.svelte  # Backup management UI
```

## Testing Checklist

- [ ] Start services: `docker compose up -d`
- [ ] Navigate to `/dashboard/settings/backups`
- [ ] Click "Create Backup Now"
- [ ] Verify backup appears in list
- [ ] Download backup and inspect contents
- [ ] Test restore:
  - Make a change to data
  - Restore backup
  - Verify change is reverted
- [ ] Test delete backup
- [ ] Verify safety backup is created before restore

## Technical Notes

### Backup Format
- Plain SQL with INSERT statements
- Gzip compressed (`.sql.gz`)
- Header includes tenant info and timestamp
- BEGIN/COMMIT transaction wrappers

### Database Requirements
- Tables must have `tenant_id` column to be backed up
- Uses `information_schema` to discover tenant tables
- Compatible with PostgreSQL

### Future Enhancements (Not Implemented)
- Scheduled auto-backups (daily at 2 AM)
- Email notifications on backup completion
- Backup encryption
- Remote storage (S3, GCS)
- Backup verification/integrity checks
- Incremental backups
- Backup size limits

## Deployment Notes

- Ensure `/backups` directory has write permissions
- Consider backup volume size limits
- Monitor disk usage (30-day retention × backup size)
- Pre-restore backups kept forever (manual cleanup may be needed)

## Commit Details

**Branch:** `feat/backup-restore`  
**Commit:** `25b09b2` (feat: Add automated backup and restore system)

**Files Changed:**
- `internal/backup/handler.go` (new)
- `internal/backup/model.go` (new)
- `internal/backup/service.go` (new)
- `web/src/routes/dashboard/settings/backups/+page.svelte` (new)
- `cmd/pews/main.go` (modified)
- `internal/router/router.go` (modified)
- `docker-compose.yml` (modified)
- `web/src/routes/dashboard/settings/+page.svelte` (modified)

**Lines Changed:** +998 insertions, -1 deletion
