# Pews API Documentation

This directory contains the OpenAPI 3.0 specification for the Pews Church Management API.

## Files

- **openapi.yaml** - Complete OpenAPI 3.0 specification
- **api/index.html** - Swagger UI documentation viewer

## Viewing the Documentation

### Local Development

1. Start the Pews server:
   ```bash
   go run cmd/server/main.go
   ```

2. Open `docs/api/index.html` in your browser

3. The Swagger UI will load the OpenAPI specification and provide an interactive API explorer

### API Overview

The Pews API is organized into the following modules:

#### Core
- **Auth** - Registration, login, logout, JWT authentication
- **Tenant** - Organization management and configuration
- **Modules** - Enable/disable platform features

#### People Management
- **People** - Individual contact management with custom fields
- **Tags** - Categorization and filtering
- **Households** - Family grouping and address management

#### Ministry
- **Groups** - Small groups, classes, ministry teams
- **Services** - Service planning with items, team scheduling
- **Songs** - Worship song library with CCLI tracking

#### Engagement
- **Giving** - Donations, funds, Stripe integration, tax statements
- **Streaming** - Live streaming with chat, viewer tracking, notes
- **Communication** - Templates, campaigns, journeys, connection cards

#### Operations
- **Check-ins** - Event attendance, medical alerts, authorized pickups
- **Billing** - Subscription management via Stripe

## Authentication

All protected endpoints require a Bearer token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

Tokens are obtained via:
- `POST /api/auth/register` - Create new tenant + admin user
- `POST /api/auth/login` - Authenticate existing user

## Base URL

- **Development:** `http://localhost:8080/api`
- **Production:** `https://api.pews.app/api`

## Common Patterns

### Pagination

List endpoints support pagination via query parameters:
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)

### Filtering

Many endpoints support filtering:
- `q` - Search query (people, songs, etc.)
- Date ranges: `from`, `to`
- Status filters: `status`, `active`
- Type filters: `type`, `type_id`

### Response Format

Standard list responses:
```json
{
  "items": [...],
  "total": 150,
  "page": 1,
  "limit": 20
}
```

### Error Responses

Errors follow a consistent format:
```json
{
  "error": "Error Type",
  "message": "Detailed error message"
}
```

Common status codes:
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `500` - Internal Server Error

## Public Endpoints

These endpoints do not require authentication:

- `GET /api/health`
- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/billing/webhook`
- `POST /api/giving/webhook`
- `GET /api/streaming/watch/{id}`
- `GET /api/streaming/{id}/chat`
- `POST /api/streaming/{id}/chat`
- `POST /api/streaming/{id}/join`
- `POST /api/streaming/{id}/leave`
- `POST /api/communication/cards`

## Admin-Only Endpoints

These endpoints require the `admin` role:

- Tenant updates
- Module management
- Billing operations
- Fund management
- Donation creation (manual entry)
- Giving statement generation
- Stripe Connect onboarding

## Multi-Tenant Architecture

All API requests are scoped to a tenant. The tenant is identified via:

1. **Subdomain** - `tenant-slug.pews.app`
2. **Custom domain** - `church.org`
3. **Login** - Tenant slug provided in login request

## Rate Limiting

(To be implemented)

## Versioning

Current version: **v1.0.0**

API versioning will follow semantic versioning. Breaking changes will be introduced in new major versions with migration guides.

## Support

For API support:
- GitHub Issues: https://github.com/warpapaya/Pews/issues
- Documentation: https://github.com/warpapaya/Pews/docs

## License

See LICENSE file in repository root.
