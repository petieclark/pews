# Pews

**Church management, simplified.**

One platform. One price. Everything your church needs.

🌐 [pews.app](https://pews.app) · 🎮 [Demo](https://demo.pews.app) · 📋 [Roadmap](https://github.com/warpapaya/Pews/milestones) · 📚 [API Docs](https://github.com/warpapaya/Pews/tree/main/docs)

---

## What is Pews?

Pews replaces the patchwork of expensive church software — PlanningCenter ($99-299/mo), Gloo ($100-200/mo), Church Online Platform, separate giving tools — with a single integrated platform for **$100/month flat**.

Built by a former worship pastor who got tired of juggling five apps to run a Sunday morning.

## ✨ Features

### Core Modules

| Module | Description | Status |
|--------|-------------|--------|
| **Core Platform** | JWT authentication, multi-tenant architecture, module registry, Stripe billing | ✅ Production |
| **People** | Member database, households, tags, custom fields, advanced search | ✅ Production |
| **Giving** | Stripe Connect donations, recurring gifts, tax receipts, fund tracking, goals | ✅ Production |
| **Groups** | Small groups, ministry teams, member management, attendance | ✅ Production |
| **Services** | Worship planning, team scheduling, song library with search | ✅ Production |
| **Check-Ins** | Child safety, kiosk mode, medical alerts, authorized pickups, attendance tracking | ✅ Production |
| **Communication** | Email/SMS campaigns, automated journeys, connection cards, HTML templates | ✅ Production |
| **Streaming** | Live stream embed (YouTube/FB/Vimeo/RTMP), chat, giving overlay, viewer tracking | ✅ Production |
| **Calendar** | Event management, recurring events, color-coding, public calendar view | ✅ Production |
| **Prayer** | Prayer request management, public/private requests, staff follow-up, status tracking | ✅ Production |
| **Reports** | Analytics dashboard with attendance trends, giving insights, membership growth, group participation | ✅ Production |
| **Notifications** | In-app notification system for important updates and reminders | ✅ Production |

### User Experience

| Feature | Description | Status |
|---------|-------------|--------|
| **Dark Mode** | System-aware theme with manual toggle | ✅ |
| **Mobile Responsive** | Fully optimized for mobile, tablet, and desktop | ✅ |
| **Accessibility** | WCAG 2.1 AA compliant with keyboard navigation and screen reader support | ✅ |
| **SEO Optimized** | Meta tags, structured data, sitemap, performance optimized | ✅ |
| **Public Pages** | Branded watch page, connection card form, online giving portal | ✅ |

### Coming Soon

| Feature | Target | Issue |
|---------|--------|-------|
| Volunteer Scheduling | v1.0 | [#22](https://github.com/warpapaya/Pews/issues/22) |
| Song Library PDF uploads | v1.0 | [#7](https://github.com/warpapaya/Pews/issues/7) |
| PCO Import (API-first) | v1.0 | [#8](https://github.com/warpapaya/Pews/issues/8) |
| CCLI / SongSelect integration | v1.0 | [#18](https://github.com/warpapaya/Pews/issues/18) |
| ProPresenter integration | v1.0 | [#21](https://github.com/warpapaya/Pews/issues/21) |
| Email/SMS delivery (SendGrid/Twilio) | v1.0 | [#13](https://github.com/warpapaya/Pews/issues/13) |
| iOS App | v1.0 | [#15](https://github.com/warpapaya/Pews/issues/15) |
| Android App | v1.1 | [#16](https://github.com/warpapaya/Pews/issues/16) |
| MultiTracks integration | v1.1 | [#19](https://github.com/warpapaya/Pews/issues/19) |
| Accounting integrations (QuickBooks) | v1.2 | TBD |

See all issues: [github.com/warpapaya/Pews/issues](https://github.com/warpapaya/Pews/issues)

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────┐
│              SvelteKit Frontend                      │
│  (TailwindCSS, Dark Mode, Mobile Responsive)        │
│  • Dashboard (all modules)                          │
│  • Public pages (watch, give, connect)              │
│  • Real-time updates                                │
├─────────────────────────────────────────────────────┤
│                 Go Backend                          │
│  (Chi router, JWT auth, RLS middleware)             │
│  • Modular architecture (pluggable modules)         │
│  • RESTful API with comprehensive documentation     │
│  • Row-Level Security automatic tenant isolation    │
│  • Stripe webhook handling                          │
├─────────────────────────────────────────────────────┤
│            PostgreSQL 16                            │
│  (Multi-tenant with Row-Level Security)             │
│  • Automatic tenant isolation via RLS policies      │
│  • Migrations with updated_at triggers              │
│  • Optimized indexes for performance                │
├─────────────────────────────────────────────────────┤
│              Stripe Connect                         │
│  • Platform billing ($100/mo subscriptions)         │
│  • Church giving accounts (1% platform fee)         │
│  • Tax receipt generation                           │
├─────────────────────────────────────────────────────┤
│            Production Stack                         │
│  • Docker Compose orchestration                     │
│  • Traefik reverse proxy with Let's Encrypt         │
│  • Docker secrets for credential management         │
│  • Automated health checks                          │
└─────────────────────────────────────────────────────┘
```

**Tech Stack:**
- **Backend:** Go 1.22+ with Chi router
- **Frontend:** SvelteKit with TailwindCSS
- **Database:** PostgreSQL 16 (multi-tenant via RLS)
- **Payments:** Stripe Connect (1% platform fee on giving)
- **Auth:** JWT with bcrypt password hashing
- **Deployment:** Docker Compose with Traefik

**Key Design Patterns:**
- **Row-Level Security (RLS):** Automatic tenant isolation at the database level
- **Module System:** Pluggable architecture for easy feature additions
- **Multi-tenancy:** Complete data isolation between churches
- **API-first:** RESTful endpoints with comprehensive documentation

See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) for detailed technical documentation.

## 🚀 Quick Start

### Prerequisites

- **Docker & Docker Compose** (recommended) OR
- **Go 1.22+** + **Node.js 20+** + **PostgreSQL 16+**

### Option 1: Docker Compose (Recommended)

```bash
# Clone the repository
git clone git@github.com:warpapaya/Pews.git
cd Pews

# Start all services (PostgreSQL + backend + frontend)
docker compose up -d

# Backend will be available at: http://localhost:8190
# Frontend will be available at: http://localhost:5273
```

### Option 2: Local Development

```bash
# Clone the repository
git clone git@github.com:warpapaya/Pews.git
cd Pews

# Set up PostgreSQL
createdb pews
psql pews < internal/database/migrations/*.sql

# Set up environment variables
cp .env.example .env
# Edit .env with your values

# Start the backend
go run cmd/pews/main.go

# In another terminal, start the frontend
cd web
npm install
npm run dev
```

### Register Your First Church

```bash
curl -X POST http://localhost:8190/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_name": "My Church",
    "email": "admin@mychurch.com",
    "password": "securepassword123"
  }'
```

You'll receive a JWT token. Use it in the `Authorization: Bearer <token>` header for authenticated requests.

### Run Demo Seed Data (Optional)

```bash
# Seed the database with realistic demo data
psql pews < scripts/seed-demo.sql

# Demo credentials:
# Email: demo@pews.app
# Password: demo1234
```

## 📁 Project Structure

```
pews/
├── cmd/pews/                    # Application entrypoint
│   └── main.go                  # Server initialization, module registration
├── internal/                    # Internal packages (not importable)
│   ├── auth/                    # JWT authentication, user management
│   ├── billing/                 # Stripe billing & subscriptions
│   ├── calendar/                # Event/calendar management
│   ├── checkins/                # Check-in system, child safety
│   ├── communication/           # Email/SMS campaigns, journeys, connection cards
│   ├── config/                  # Configuration management
│   ├── database/                # DB connection, migrations, seed data
│   │   └── migrations/          # All database migrations (001-012)
│   ├── giving/                  # Donations, funds, Stripe Connect, tax receipts
│   ├── groups/                  # Small groups, ministry teams
│   ├── middleware/              # HTTP middleware (RLS, auth, CORS)
│   ├── module/                  # Module system registry
│   ├── notifications/           # In-app notification system
│   ├── people/                  # Member database, households, tags
│   ├── prayer/                  # Prayer request management
│   ├── reports/                 # Analytics dashboard
│   ├── router/                  # HTTP routing (Chi)
│   ├── services/                # Worship planning, songs, scheduling
│   ├── streaming/               # Live streaming, chat, viewer tracking
│   └── tenant/                  # Multi-tenant management
├── web/                         # SvelteKit frontend
│   ├── src/
│   │   ├── lib/                 # Shared components & utilities
│   │   │   └── components/      # Reusable UI components
│   │   ├── routes/              # SvelteKit file-based routing
│   │   │   ├── dashboard/       # Authenticated admin pages
│   │   │   │   ├── calendar/    # Calendar module
│   │   │   │   ├── checkins/    # Check-ins module
│   │   │   │   ├── communication/ # Communication module
│   │   │   │   ├── giving/      # Giving module
│   │   │   │   ├── groups/      # Groups module
│   │   │   │   ├── notifications/ # Notifications
│   │   │   │   ├── people/      # People module
│   │   │   │   ├── prayer/      # Prayer requests
│   │   │   │   ├── reports/     # Analytics dashboard
│   │   │   │   ├── services/    # Services module
│   │   │   │   ├── settings/    # Tenant settings, billing
│   │   │   │   └── streaming/   # Streaming module
│   │   │   ├── connect/         # Public connection card form
│   │   │   ├── give/            # Public online giving portal
│   │   │   ├── login/           # Login page
│   │   │   ├── register/        # Registration page
│   │   │   └── watch/           # Public streaming viewer
│   │   ├── app.css              # Global styles, dark mode
│   │   └── app.html             # HTML template
│   ├── static/                  # Static assets
│   ├── package.json
│   └── svelte.config.js
├── landing/                     # Marketing website (pews.app)
│   ├── index.html               # Home page
│   ├── about.html               # About page
│   ├── pricing.html             # Pricing page
│   └── screenshots/             # Product screenshots
├── docs/                        # Documentation
│   ├── ARCHITECTURE.md          # Technical architecture deep dive
│   ├── API.md                   # API documentation
│   └── *.md                     # Various guides and reports
├── scripts/                     # Utility scripts
│   ├── seed-demo.sql            # Demo data seeding script
│   └── test-*.sh                # API testing scripts
├── docker-compose.yml           # Development environment
├── docker-compose.prod.yml      # Production environment with Traefik
├── Dockerfile                   # Backend image
├── .env.example                 # Environment variables template
├── CHANGELOG.md                 # Version history and changes
├── CONTRIBUTING.md              # Developer contribution guide
└── README.md                    # This file
```

## 🗄️ Database Migrations

Migrations run automatically on backend startup. Located in `internal/database/migrations/`:

| # | Migration | Description |
|---|-----------|-------------|
| 001 | tenants | Multi-tenant core table with domain and status |
| 002 | users | User accounts with JWT authentication |
| 003 | modules | Module registry for feature toggles |
| 004 | billing | Stripe billing, subscriptions, payment methods |
| 005 | people | Member database, households, tags, custom fields |
| 006 | giving | Donations, funds, recurring gifts, Stripe Connect accounts |
| 007 | groups | Small groups, ministry teams, membership |
| 008 | services | Worship services, songs, team positions, assignments |
| 009 | checkins | Events, stations, check-in logs, child safety |
| 010 | communication | Campaigns, journeys, connection cards, templates |
| 011 | streaming | Live streams, chat messages, viewer sessions |
| 012 | calendar/events | Events, recurring events, calendar management |
| 012 | prayer_requests | Prayer requests, followers, status tracking |
| 012 | notifications | In-app notifications for users |

**Note:** Migrations 012 are feature-branch specific and will be renumbered when merged.

All tables use Row-Level Security (RLS) for automatic tenant isolation.

## 🎮 Demo

Live demo: **[demo.pews.app](https://demo.pews.app)**

**Demo Credentials:**
- **Church slug:** `demo-church`
- **Email:** `demo@pews.app`
- **Password:** `demo1234`

**Test Giving:**
Stripe is in test mode. Use card `4242 4242 4242 4242` with any future expiration and any 3-digit CVC.

## 💰 Pricing

| Plan | Price | Includes |
|------|-------|----------|
| **Free** | $0/mo | People database, basic communication, 1 admin user |
| **Pro** | **$100/mo** | **Everything** — all modules, unlimited admins, unlimited members |

**Giving Module:**
- Stripe processing fees (2.9% + $0.30)
- 1% platform fee
- No monthly charge for the Giving module itself

**Why $100 flat?**
Most churches pay $200-500/month across multiple tools. Pews gives you everything in one place for one predictable price.

## 🧪 Testing

### Run Backend Tests

```bash
go test ./...
```

### Run API Tests

```bash
# Automated test suite
bash scripts/test-api.sh

# Test specific module
bash scripts/test-people.sh
bash scripts/test-giving.sh
```

### Run Frontend Tests

```bash
cd web
npm test
```

See `docs/testing/` for detailed testing guides.

## 📸 Screenshots

_Coming soon: Product screenshots showcasing all modules_

- Dashboard overview
- People management
- Giving portal
- Service planning
- Check-in kiosk
- Communication campaigns
- Live streaming
- Calendar view
- Prayer requests
- Analytics reports

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for:
- Development setup
- Code style guidelines
- How to add a new module
- Testing requirements
- Pull request process
- Branch naming conventions

## 📄 License

Proprietary — © 2026 Clearline Technology Methods

Pews is currently a private project. For partnership or licensing inquiries, contact: **hello@pews.app**

## 🆘 Support

- **Documentation:** [docs/](./docs/)
- **Issues:** [GitHub Issues](https://github.com/warpapaya/Pews/issues)
- **Email:** support@pews.app
- **Website:** [pews.app](https://pews.app)

## 🗺️ Roadmap

See our [GitHub Milestones](https://github.com/warpapaya/Pews/milestones) for upcoming features and releases.

**Current Focus (v1.0):**
- ✅ Core platform with multi-tenancy
- ✅ All primary modules (People, Giving, Groups, Services, Check-Ins, Communication, Streaming)
- ✅ Calendar, Prayer, Reports, Notifications
- 🚧 Volunteer scheduling
- 🚧 CCLI/SongSelect integration
- 🚧 ProPresenter integration
- 🚧 PCO data import
- 🚧 iOS app

**Future (v1.1+):**
- Mobile apps (iOS/Android)
- Advanced integrations (MultiTracks, accounting)
- White-label options for denominations
- Advanced reporting and analytics

---

Built with ❤️ by [Clearline Technology Methods](https://clearlinetechmethods.com)
