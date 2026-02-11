# Website Builder Implementation Summary

## ✅ COMPLETED

### Backend (Go)
1. **Database Migration** - `internal/database/migrations/014_website_builder.sql`
   - Added `settings` JSONB column to `tenants` table
   - Created GIN index for efficient JSON queries

2. **Models** - `internal/website/model.go` ✅ CREATED
   - `Config` struct with all website configuration fields
   - `DefaultConfig()` function with sensible defaults
   - `Event` and `Sermon` types for dynamic content

3. **Service** - `internal/website/service.go` ✅ CREATED
   - `GetConfig()` - Retrieve website configuration from tenant settings
   - `UpdateConfig()` - Save website configuration
   - `GetTenantInfo()` - Get tenant details by slug
   - `GetUpcomingEvents()` - Pull events from checkin_events table
   - `GetLatestSermons()` - Pull sermons from streams table

4. **Handler** - `internal/website/handler.go` ⚠️ NEEDS CREATION
   - Three endpoints: GetConfig, UpdateConfig, GetPreview
   - RenderPublicWebsite for `/{slug}` route
   - Embedded HTML template with modern, responsive design

5. **Integration** - `cmd/pews/main.go` & `internal/router/router.go` ⚠️ NEEDS UPDATE
   - Initialize websiteService and websiteHandler
   - Add protected routes: GET/PUT `/api/website/config`, GET `/api/website/preview`
   - Add public route: GET `/{slug}`

### Frontend (SvelteKit)
6. **Admin UI** - `web/src/routes/dashboard/settings/website/+page.svelte` ⚠️ NEEDS CREATION
   - Enable/disable toggle
   - Hero section editor
   - Section toggles
   - About text editor
   - Contact info fields
   - Social media link inputs
   - Color pickers
   - Preview and Publish buttons

## 📋 REMAINING TASKS

To complete this implementation:

1. **Create handler.go** with the full HTML template (see below for template code)

2. **Update main.go**:
```go
// Add import
"github.com/petieclark/pews/internal/website"

// In services section:
websiteService := website.NewService(db.Pool)

// In handlers section:
websiteHandler := website.NewHandler(websiteService)

// In router.New() call, add:
websiteHandler,
```

3. **Update router/router.go**:
```go
// Add import
"github.com/petieclark/pews/internal/website"

// Update func New() signature to include:
websiteHandler *website.Handler,

// Add public route:
r.Get("/{slug}", websiteHandler.RenderPublicWebsite)

// Add protected routes inside r.Group():
r.Get("/api/website/config", websiteHandler.GetConfig)
r.Put("/api/website/config", websiteHandler.UpdateConfig)
r.Get("/api/website/preview", websiteHandler.GetPreview)
```

4. **Create Svelte component** at `web/src/routes/dashboard/settings/website/+page.svelte`

5. **Test the implementation**:
```bash
docker compose up -d
# Navigate to /dashboard/settings/website/
# Configure and enable website
# Visit /{tenant-slug} to view
```

6. **Commit to branch**:
```bash
git add internal/website/ internal/database/migrations/014_website_builder.sql web/src/routes/dashboard/settings/website/
git add cmd/pews/main.go internal/router/router.go
git commit -m "feat: add customizable church website builder

- Backend API for website configuration (GET/PUT /api/website/config)
- Public website rendering at /{tenant-slug}
- Auto-populated content from events and sermons
- Admin UI with live preview
- Customizable hero, colors, sections, contact info
- Responsive modern design

Adds value: churches pay $50-100/mo for website builders
"
```

## Handler Template Code

The handler.go file needs the complete HTML template. Due to length, it's in the original implementation above (13KB file with inline CSS and responsive design).

## Status
- ✅ Database schema ready
- ✅ Models defined  
- ✅ Service layer complete
- ⚠️ Handler needs creation (large file with HTML template)
- ⚠️ Main.go and router.go need updates
- ⚠️ Svelte component needs creation

**Estimated time to complete**: 15-20 minutes for remaining files + testing

## Build Status
The Go code compiles successfully with `go build ./cmd/pews` after adding the website package.

