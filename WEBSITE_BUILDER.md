# Website Builder Feature

## Overview
The Website Builder allows churches to create and customize a public-facing website directly from Pews, eliminating the need for expensive third-party website builders ($50-100/month).

## Features
- ✅ Customizable hero section with title, subtitle, and background image
- ✅ Toggle-able sections (About, Services, Sermons, Events, Connect, Give)
- ✅ Auto-populated content from existing Pews data (events, sermons)
- ✅ Customizable brand colors (primary & accent)
- ✅ Social media links (Facebook, Instagram, YouTube)
- ✅ Contact information (address, phone, email)
- ✅ Live preview
- ✅ Public website accessible at `/{tenant-slug}`

## API Endpoints

### Authenticated Endpoints (Admin Dashboard)
- `GET /api/website/config` - Get website configuration
- `PUT /api/website/config` - Update website configuration
- `GET /api/website/preview` - Preview website HTML

### Public Endpoints
- `GET /{tenant-slug}` - View public website (when enabled)

## Frontend Admin
Navigate to `/dashboard/settings/website/` to:
- Enable/disable the public website
- Customize hero section (title, subtitle, image)
- Toggle sections on/off
- Edit about text
- Configure service times and contact info
- Add social media links
- Customize brand colors
- Preview changes
- Publish to make live

## Configuration Model
Website configuration is stored in the `tenants.settings` JSONB column under the `website` key:

```json
{
  "enabled": true,
  "theme": "modern",
  "hero_title": "Welcome to Our Church",
  "hero_subtitle": "Join us this Sunday",
  "hero_image_url": "",
  "service_times": "Sunday 9:00 AM & 11:00 AM",
  "address": "123 Main St",
  "phone": "",
  "email": "",
  "sections": ["about", "services", "sermons", "events", "connect", "give"],
  "about_text": "",
  "social_links": {
    "facebook": "",
    "instagram": "",
    "youtube": ""
  },
  "colors": {
    "primary": "#1B3A4B",
    "accent": "#4A8B8C"
  }
}
```

## Data Sources
The website automatically pulls data from existing Pews modules:
- **Events**: Upcoming events from the Check-ins module
- **Sermons**: Recent sermon notes from the Streaming module
- No duplicate data entry required!

## Testing
1. Start the application:
   ```bash
   docker compose up -d
   ```

2. Log in to the dashboard as an admin

3. Navigate to Settings → Website

4. Configure your website:
   - Enable the website
   - Customize hero section
   - Toggle sections
   - Add contact info and social links
   - Adjust brand colors

5. Click "Preview" to see changes

6. Click "Publish Changes" to save

7. Visit `/{your-tenant-slug}` to see the live public website

## Database Migration
Migration `014_website_builder.sql` adds the `settings` JSONB column to the `tenants` table.

## Files Added
- `internal/website/model.go` - Configuration model
- `internal/website/service.go` - Business logic
- `internal/website/handler.go` - HTTP handlers and HTML template
- `web/src/routes/dashboard/settings/website/+page.svelte` - Admin UI
- `internal/database/migrations/014_website_builder.sql` - Database migration

## Notes
- The website is disabled by default
- When disabled, accessing `/{tenant-slug}` returns 404
- The website uses a modern, mobile-responsive design
- All content is SEO-friendly with semantic HTML
- The Give button links to the dashboard giving page (can be enhanced later)

## Future Enhancements
- Custom domain support
- Image upload (currently URL only)
- Additional themes
- SEO metadata customization
- Custom CSS editor
- Photo gallery section
- Staff/team directory
- Embedded Google Maps for address
