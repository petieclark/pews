# Pews

**Church management, simplified.**

One platform. One price. Everything your church needs.

🌐 [pews.app](https://pews.app) · 🎮 [Demo](https://demo.pews.app) · 📋 [Roadmap](https://github.com/warpapaya/Pews/milestones)

---

## What is Pews?

Pews replaces the patchwork of expensive church software — PlanningCenter ($99-299/mo), Gloo ($100-200/mo), Church Online Platform, separate giving tools — with a single integrated platform for **$100/month flat**.

Built by a former worship pastor who got tired of juggling five apps to run a Sunday morning.

## Features

| Module | Description | Status |
|--------|-------------|--------|
| **Core** | Auth, multi-tenant, module registry, Stripe billing | ✅ |
| **People** | Member database, households, tags, search | ✅ |
| **Giving** | Stripe Connect donations, recurring gifts, tax receipts, fund tracking | ✅ |
| **Groups** | Small groups, ministry teams, member management | ✅ |
| **Services** | Worship planning, team scheduling, song library | ✅ |
| **Check-Ins** | Child safety, kiosk mode, medical alerts, authorized pickups, attendance | ✅ |
| **Communication** | Email/SMS campaigns, automated journeys, connection cards, templates | ✅ |
| **Streaming** | Live stream embed (YouTube/FB/Vimeo/RTMP), chat, giving overlay, viewer tracking | ✅ |
| **Dark Mode** | System-aware theme with manual toggle | ✅ |

### Coming Soon

| Feature | Target | Issue |
|---------|--------|-------|
| Song Library PDF uploads | v1.0 | [#7](https://github.com/warpapaya/Pews/issues/7) |
| PCO Import (API-first) | v1.0 | [#8](https://github.com/warpapaya/Pews/issues/8) |
| CCLI / SongSelect integration | v1.0 | [#18](https://github.com/warpapaya/Pews/issues/18) |
| ProPresenter integration | v1.0 | [#21](https://github.com/warpapaya/Pews/issues/21) |
| Public REST API + docs | v1.0 | [#17](https://github.com/warpapaya/Pews/issues/17) |
| Email/SMS delivery (SendGrid/Twilio) | v1.0 | [#13](https://github.com/warpapaya/Pews/issues/13) |
| iOS App | v1.0 | [#15](https://github.com/warpapaya/Pews/issues/15) |
| Android App | v1.1 | [#16](https://github.com/warpapaya/Pews/issues/16) |
| MultiTracks integration | v1.1 | [#19](https://github.com/warpapaya/Pews/issues/19) |
| Calendar / event management | v1.1 | [#9](https://github.com/warpapaya/Pews/issues/9) |

See all issues: [github.com/warpapaya/Pews/issues](https://github.com/warpapaya/Pews/issues)

## Architecture

```
┌─────────────────────────────────────────┐
│              SvelteKit Frontend          │
│         (TailwindCSS, Dark Mode)        │
├─────────────────────────────────────────┤
│               Go Backend                │
│  (Chi router, JWT auth, module system)  │
├─────────────────────────────────────────┤
│          PostgreSQL 16                  │
│  (Multi-tenant, Row-Level Security)     │
├─────────────────────────────────────────┤
│            Stripe Connect               │
│  (Billing + church giving accounts)     │
└─────────────────────────────────────────┘
```

- **Backend:** Go 1.22 with Chi router
- **Frontend:** SvelteKit with TailwindCSS
- **Database:** PostgreSQL 16 (multi-tenant via RLS)
- **Payments:** Stripe Connect (1% platform fee on giving)
- **Auth:** JWT with bcrypt password hashing
- **Deployment:** Docker Compose

## Development

### Prerequisites

- Docker & Docker Compose
- Go 1.22+ (or use Docker for builds)
- Node.js 20+

### Quick Start

```bash
# Clone
git clone git@github.com:warpapaya/Pews.git
cd Pews

# Start all services (PostgreSQL + backend + frontend)
docker compose up -d

# Backend: http://localhost:8190
# Frontend: http://localhost:5273

# Register a new church
curl -X POST http://localhost:8190/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"tenant_name":"My Church","email":"admin@mychurch.com","password":"password123"}'
```

### Production Deployment

```bash
# Build images
docker build --platform linux/amd64 -t pews-backend:latest .
docker build --platform linux/amd64 -t pews-frontend:latest ./web

# Deploy with production compose
docker compose -f docker-compose.prod.yml up -d
```

### Project Structure

```
├── cmd/pews/           # Application entrypoint
├── internal/
│   ├── auth/           # Authentication (JWT, bcrypt)
│   ├── billing/        # Stripe billing & subscriptions
│   ├── checkins/       # Check-in events, stations, child safety
│   ├── communication/  # Campaigns, journeys, connection cards
│   ├── database/       # DB connection, migrations
│   ├── giving/         # Donations, funds, Stripe Connect
│   ├── groups/         # Small groups, ministry teams
│   ├── router/         # HTTP routing (chi)
│   ├── services/       # Worship services, songs, scheduling
│   ├── streaming/      # Live streaming, chat, viewer tracking
│   └── tenant/         # Multi-tenant management
├── web/                # SvelteKit frontend
│   └── src/routes/
│       ├── dashboard/  # Admin pages (all modules)
│       └── watch/      # Public streaming viewer page
├── landing/            # Marketing site (pews.app)
├── docker-compose.yml  # Development
└── docker-compose.prod.yml  # Production
```

### Database Migrations

Migrations run automatically on backend startup. Located in `internal/database/migrations/`:

| Migration | Module |
|-----------|--------|
| 001 | Tenants |
| 002 | Users |
| 003 | Modules |
| 004 | Billing/Subscriptions |
| 005 | People |
| 006 | Giving |
| 007 | Groups |
| 008 | Services |
| 009 | Check-Ins |
| 010 | Communication |
| 011 | Streaming |

## Demo

Live demo: **[demo.pews.app](https://demo.pews.app)**

- **Church slug:** `demo-church`
- **Email:** `demo@pews.app`
- **Password:** `demo1234`

Stripe is in test mode — use card `4242 4242 4242 4242` for testing.

## Pricing

| Plan | Price | Includes |
|------|-------|----------|
| Free | $0/mo | People database, basic email, 1 admin |
| Pro | $100/mo | Everything — all modules, unlimited admins |

**Giving:** Stripe processing fees + 1% platform fee. No monthly charge for the Giving module.

## Contributing

Pews is currently a private project by [Clearline Technology Methods](https://clearlinetechmethods.com). If you're interested in contributing or partnering, reach out at hello@pews.app.

## License

Proprietary — © 2026 Clearline Technology Methods
