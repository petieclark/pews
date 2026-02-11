# Pews Architecture Documentation

## System Overview

Pews is a multi-tenant SaaS platform for church management built with Go, SvelteKit, and PostgreSQL.

### High-Level Architecture

```
Client (Browser/Mobile)
        ↓
   HTTPS (TLS)
        ↓
Traefik Reverse Proxy (Let's Encrypt SSL)
        ↓
    ┌───┴───┐
    │       │
SvelteKit   Go Backend (Chi Router)
Frontend    ├── Middleware (CORS, Auth, RLS)
            ├── Module System (pluggable)
            └── RESTful API
        ↓
PostgreSQL 16 (Row-Level Security)
        ↓
External Services (Stripe Connect)
```

## Technology Stack

- **Backend:** Go 1.22+ with Chi router
- **Frontend:** SvelteKit + TailwindCSS + Vite
- **Database:** PostgreSQL 16 with RLS
- **Payments:** Stripe Connect
- **Deployment:** Docker Compose + Traefik

## Multi-Tenancy Design

**Shared database, shared schema with Row-Level Security (RLS)**

### Why RLS?

✅ Cost-effective (single database)  
✅ Database-level security enforcement  
✅ No manual tenant filtering in queries  
✅ Easy maintenance and migrations

### Implementation

1. **Tenant Table (Master)**
\`\`\`sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY,
    name VARCHAR(255),
    slug VARCHAR(100) UNIQUE,  -- subdomain
    domain VARCHAR(255) UNIQUE, -- custom domain
    status VARCHAR(20),         -- active, suspended
    created_at TIMESTAMP
);
\`\`\`

2. **Foreign Key on All Tables**
\`\`\`sql
CREATE TABLE people (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id),
    first_name VARCHAR(100),
    -- ...
);
\`\`\`

3. **RLS Policies**
\`\`\`sql
ALTER TABLE people ENABLE ROW LEVEL SECURITY;

CREATE POLICY people_isolation_policy ON people
    USING (tenant_id = current_setting('app.current_tenant_id', TRUE)::UUID);
\`\`\`

4. **Set Context Per Request**
\`\`\`go
// Middleware extracts tenant_id from JWT
_, err := db.ExecContext(ctx, "SET app.current_tenant_id = $1", tenantID)
\`\`\`

**Result:** All queries automatically filtered by tenant!

## Database Schema Overview

### Core Tables

- **tenants** - Multi-tenant master table
- **users** - User accounts with JWT auth
- **modules** - Feature toggles per tenant
- **billing** - Stripe subscriptions

### Module Tables

- **people** - Members, households, tags
- **giving** - Donations, funds, recurring gifts, Stripe accounts
- **groups** - Small groups, ministry teams
- **services** - Worship planning, songs, assignments
- **checkins** - Events, stations, check-in logs
- **communication** - Campaigns, journeys, connection cards
- **streaming** - Live streams, chat, viewer sessions
- **events** - Calendar, recurring events
- **prayer_requests** - Prayer management, followers
- **notifications** - In-app notifications

All tables have:
- `tenant_id` foreign key
- RLS policies enabled
- Indexes on `tenant_id` and common queries
- `updated_at` triggers

## Authentication & Authorization

### JWT Authentication Flow

1. User logs in → credentials validated (bcrypt)
2. Backend generates JWT with payload:
   \`\`\`json
   {
     "user_id": "uuid",
     "tenant_id": "uuid",
     "email": "user@example.com",
     "role": "admin",
     "exp": timestamp
   }
   \`\`\`
3. Client stores JWT (localStorage)
4. Client sends JWT in \`Authorization: Bearer <token>\` header
5. AuthMiddleware validates JWT
6. Extract user_id, tenant_id → add to request context
7. RLSMiddleware sets tenant context in PostgreSQL
8. Request proceeds to handler

### Password Security

- Bcrypt hashing with default cost (10)
- Salted hashes stored in database
- No plaintext passwords ever stored

### Roles

- **Admin** - Full access
- **Staff** - Module-specific access
- **Volunteer** - Limited access (future)

## Module System

Modules are self-contained features.

### Structure

\`\`\`
internal/modulename/
├── models.go       # Data structures
├── service.go      # Business logic (DB operations)
├── handler.go      # HTTP handlers (request/response)
└── routes.go       # Route registration
\`\`\`

### Pattern

**Service Layer:**
\`\`\`go
type Service struct { db *sql.DB }

func (s *Service) Create(ctx context.Context, tenantID uuid.UUID, req *CreateRequest) (*Model, error)
func (s *Service) List(ctx context.Context, tenantID uuid.UUID) ([]*Model, error)
func (s *Service) Get(ctx context.Context, tenantID, id uuid.UUID) (*Model, error)
func (s *Service) Update(ctx context.Context, tenantID, id uuid.UUID, req *UpdateRequest) (*Model, error)
func (s *Service) Delete(ctx context.Context, tenantID, id uuid.UUID) error
\`\`\`

**Handler Layer:**
\`\`\`go
type Handler struct { service *Service }

func (h *Handler) Create(w http.ResponseWriter, r *http.Request)
func (h *Handler) List(w http.ResponseWriter, r *http.Request)
// ...
\`\`\`

**Routes:**
\`\`\`go
func RegisterRoutes(r chi.Router, handler *Handler) {
    r.Route("/resource", func(r chi.Router) {
        r.Get("/", handler.List)
        r.Post("/", handler.Create)
        r.Get("/{id}", handler.Get)
        r.Put("/{id}", handler.Update)
        r.Delete("/{id}", handler.Delete)
    })
}
\`\`\`

## API Design Patterns

### RESTful Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | \`/api/resource\` | List all |
| POST | \`/api/resource\` | Create new |
| GET | \`/api/resource/:id\` | Get single |
| PUT | \`/api/resource/:id\` | Update |
| DELETE | \`/api/resource/:id\` | Delete |

### Status Codes

- 200 OK - Success (GET, PUT)
- 201 Created - Resource created (POST)
- 204 No Content - Deleted (DELETE)
- 400 Bad Request - Invalid input
- 401 Unauthorized - Missing/invalid JWT
- 403 Forbidden - Insufficient permissions
- 404 Not Found - Resource doesn't exist
- 500 Internal Server Error

### Response Format

**Success:**
\`\`\`json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "field": "value",
  "created_at": "2026-02-11T00:00:00Z"
}
\`\`\`

**Error:**
\`\`\`json
{
  "error": "Descriptive message"
}
\`\`\`

## Frontend Architecture

### SvelteKit File-Based Routing

\`\`\`
routes/
├── +layout.svelte              # Root layout (auth check)
├── +page.svelte                # Home (/)
├── login/+page.svelte          # Login
├── dashboard/
│   ├── +layout.svelte          # Dashboard layout (sidebar)
│   ├── +page.svelte            # Dashboard home
│   ├── people/
│   │   ├── +page.svelte        # List
│   │   ├── +page.js            # Load function
│   │   └── [id]/+page.svelte   # Detail
│   └── giving/...
└── watch/[id]/+page.svelte     # Public watch page
\`\`\`

### Data Loading

\`\`\`javascript
// +page.js
export async function load({ fetch }) {
  const token = localStorage.getItem('token');
  const res = await fetch(\`\${API_URL}/api/people\`, {
    headers: { 'Authorization': \`Bearer \${token}\` }
  });
  return { people: await res.json() };
}
\`\`\`

### Dark Mode

\`\`\`svelte
<div class="bg-white dark:bg-gray-800 text-gray-900 dark:text-white">
  Content
</div>
\`\`\`

Toggle with \`document.documentElement.classList.add('dark')\`

## Deployment Architecture

### Production Stack

\`\`\`yaml
services:
  traefik:
    - Let's Encrypt SSL (automatic)
    - Reverse proxy
    - Load balancing
  
  frontend:
    - SvelteKit (SSR)
    - Port: 5173
  
  backend:
    - Go binary
    - Port: 8190
  
  db:
    - PostgreSQL 16
    - Port: 5432 (internal)
\`\`\`

### Docker Secrets

Production uses Docker secrets for:
- Database credentials
- JWT secret
- Stripe API keys

## Security

### Checklist

- [x] HTTPS only (Traefik + Let's Encrypt)
- [x] JWT authentication
- [x] Bcrypt password hashing
- [x] Parameterized SQL queries (prevent injection)
- [x] Row-Level Security (tenant isolation)
- [x] CORS configuration
- [x] Docker secrets management
- [x] Input validation
- [x] XSS prevention (Svelte auto-escaping)

### Best Practices

1. **Never commit secrets** - Use .env.example templates
2. **RLS on every table** - Automatic tenant isolation
3. **Prepared statements** - SQL injection prevention
4. **Bcrypt for passwords** - Never store plaintext
5. **JWT expiration** - Tokens expire after 24 hours
6. **HTTPS only** - Enforced by Traefik

---

For more details, see:
- [README.md](../README.md)
- [CONTRIBUTING.md](../CONTRIBUTING.md)
- [GitHub Issues](https://github.com/warpapaya/Pews/issues)

Contact: **dev@pews.app**
