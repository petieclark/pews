# Church Profile & Branding Feature

**Branch:** `feat/church-profile`
**Status:** Backend complete, Frontend complete, Migration applied

## Summary

Added comprehensive church profile and branding customization to allow churches to personalize their Pews instance with logo, contact information, address, and about text.

## What Was Built

### 1. Database Migration
**File:** `internal/database/migrations/012_church_profile.sql`

Added columns to `tenants` table:
- `phone` (VARCHAR(50))
- `website` (VARCHAR(255))
- `email` (VARCHAR(255))  
- `logo` (TEXT) - stores base64 data URL
- `about` (TEXT)

Note: `address_line1`, `address_line2`, `city`, `state`, `zip`, and `ein` already existed from previous migrations.

### 2. Backend Changes

**Model (`internal/tenant/model.go`):**
- Extended `Tenant` struct with all profile fields
- Added `UpdateProfileRequest` struct for profile updates

**Service (`internal/tenant/service.go`):**
- Updated `GetTenantByID` and `GetTenantBySlug` to return all profile fields
- Added `UpdateProfile()` method for profile updates
- Added `UpdateLogo()` method for logo upload

**Handler (`internal/tenant/handler.go`):**
- Added `GetProfile()` endpoint - GET /api/tenant/profile
- Added `UpdateProfile()` endpoint - PUT /api/tenant/profile  
- Added `UploadLogo()` endpoint - POST /api/tenant/profile/logo (multipart form, max 5MB)

**Router (`internal/router/router.go`):**
- Registered new profile endpoints

### 3. Frontend Changes

**Church Profile Page (`web/src/routes/dashboard/settings/profile/+page.svelte`):**
Complete settings page with sections for:
- Logo upload with preview
- Basic information (name, email, phone, website)
- Address (line1, line2, city, state, zip)
- Tax information (EIN)
- About/description

**Settings Page (`web/src/routes/dashboard/settings/+page.svelte`):**
- Added "Quick Links" section with link to Church Profile

**Dashboard Layout (`web/src/routes/dashboard/+layout.svelte`):**
- Fetches tenant profile on load
- Displays church logo in nav (if set)
- Displays church name instead of generic "Pews"

## Logo Storage

Logos are stored as base64-encoded data URLs directly in the database (`logo` TEXT column). Format: `data:image/png;base64,iVBORw0KGg...`

**Pros:**
- Simple implementation, no file storage needed
- Works in Docker without volume mounts
- Easy backup (just database)

**Cons:**
- Database size increases with images
- Not ideal for production at scale

**Future Enhancement:** Consider moving to S3/Cloud Storage with URL reference in database.

## Testing

1. Start services: `cd ~/Projects/pews && docker compose up -d`
2. Login to dashboard
3. Navigate to Settings → Church Profile
4. Fill out form fields and upload logo
5. Save and verify logo appears in nav

## API Endpoints

### GET /api/tenant/profile
Returns complete tenant profile including all branding fields.

**Response:**
```json
{
  "id": "uuid",
  "name": "First Baptist Church",
  "slug": "first-baptist-church",
  "logo": "data:image/png;base64,...",
  "email": "contact@church.org",
  "phone": "(555) 123-4567",
  "website": "https://church.org",
  "address_line1": "123 Main St",
  "city": "Springfield",
  "state": "IL",
  "zip": "62701",
  "ein": "12-3456789",
  "about": "We are a welcoming community...",
  ...
}
```

### PUT /api/tenant/profile
Updates church profile (admin only).

**Request Body:**
```json
{
  "name": "First Baptist Church",
  "address_line1": "123 Main St",
  "address_line2": "",
  "city": "Springfield",
  "state": "IL",
  "zip": "62701",
  "phone": "(555) 123-4567",
  "website": "https://church.org",
  "email": "contact@church.org",
  "ein": "12-3456789",
  "about": "We are a welcoming community..."
}
```

### POST /api/tenant/profile/logo
Uploads church logo (admin only).

**Request:** multipart/form-data with `logo` file field
**File Requirements:** 
- Must be an image (image/*)
- Max size: 5MB

**Response:**
```json
{
  "message": "Logo uploaded successfully",
  "logo": "data:image/png;base64,..."
}
```

## Migration Applied

Migration `012_church_profile.sql` was successfully applied to the database. All new columns were added with `IF NOT EXISTS` to handle existing columns gracefully.

## Files Modified/Created

**Backend:**
- `internal/database/migrations/012_church_profile.sql` (new)
- `internal/tenant/model.go` (modified)
- `internal/tenant/service.go` (modified)
- `internal/tenant/handler.go` (modified)
- `internal/router/router.go` (modified)

**Frontend:**
- `web/src/routes/dashboard/settings/profile/+page.svelte` (new)
- `web/src/routes/dashboard/settings/+page.svelte` (modified)
- `web/src/routes/dashboard/+layout.svelte` (modified)

## Next Steps

1. **Commit all changes** to feat/church-profile branch
2. **Test thoroughly** - verify all CRUD operations work
3. **Consider public pages** - use logo/about on public-facing pages (watch, give, connect)
4. **Email templates** - embed logo in automated emails
5. **Consider file storage migration** for production scalability

## Notes

- Admin role required for all profile updates
- Logo stored as base64 in database (simple but not scalable)
- State field is VARCHAR(2) (intended for 2-letter codes like "IL", "CA")
- All profile fields are optional except church name
- Church slug is read-only (auto-generated from name)
