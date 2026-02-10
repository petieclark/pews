# Pews Development Guide

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 16 (via Docker)

## Quick Start

1. **Clone the repository:**
   ```bash
   git clone <repo-url>
   cd pews
   ```

2. **Create `.env` file:**
   ```bash
   cp .env.example .env
   ```
   
   Edit `.env` and set your values (especially `JWT_SECRET` and Stripe keys for production).

3. **Start development environment:**
   ```bash
   make dev
   ```

   This will start:
   - PostgreSQL on `:5432`
   - Go backend on `:8080`
   - SvelteKit frontend on `:5173`

4. **Access the application:**
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080/api/health

## Development Workflow

### Backend Development

**Run migrations:**
```bash
make migrate
```

**Seed test data:**
```bash
make seed
```

This creates a test tenant with:
- Tenant slug: `test-church`
- Admin: `admin@testchurch.com` / `password123`
- Member: `member@testchurch.com` / `password123`

**Build backend:**
```bash
make build
```

**Run tests:**
```bash
make test
```

### Frontend Development

The frontend runs in dev mode with hot reload via Docker Compose.

**Install dependencies manually:**
```bash
cd web
npm install
```

**Run frontend standalone:**
```bash
cd web
npm run dev
```

## Project Structure

```
pews/
├── cmd/pews/              # Main entry point
├── internal/              # Internal packages
│   ├── auth/              # Authentication & JWT
│   ├── billing/           # Stripe integration
│   ├── config/            # Configuration loading
│   ├── database/          # Database connection & migrations
│   │   └── migrations/    # SQL migration files
│   ├── middleware/        # HTTP middleware
│   ├── module/            # Module registry
│   ├── router/            # HTTP router setup
│   └── tenant/            # Tenant management
├── web/                   # SvelteKit frontend
│   └── src/
│       ├── lib/           # Shared utilities
│       └── routes/        # Pages
├── scripts/               # Helper scripts
├── docker-compose.yml     # Docker setup
├── Dockerfile             # Backend container
└── Makefile               # Development commands
```

## API Endpoints

### Public
- `POST /api/auth/register` - Register new tenant + admin user
- `POST /api/auth/login` - Login
- `POST /api/billing/webhook` - Stripe webhooks
- `GET /api/health` - Health check

### Protected (requires JWT)
- `POST /api/auth/logout` - Logout (client-side token deletion)
- `GET /api/tenant` - Get current tenant
- `PUT /api/tenant` - Update tenant (admin only)
- `GET /api/tenant/modules` - List modules
- `POST /api/tenant/modules/:name/enable` - Enable module (admin only)
- `POST /api/tenant/modules/:name/disable` - Disable module (admin only)
- `GET /api/billing/subscription` - Get subscription
- `POST /api/billing/checkout` - Create Stripe checkout session (admin only)
- `POST /api/billing/portal` - Create Stripe portal session (admin only)

## Database Migrations

Migrations are automatically embedded in the binary and run on startup.

**Migration files:**
- `001_tenants.sql` - Tenants table + RLS
- `002_users.sql` - Users table + RLS
- `003_modules.sql` - Tenant modules table + RLS
- `004_billing.sql` - Subscriptions table + RLS

All tables use Row-Level Security (RLS) for tenant isolation.

## Configuration

Environment variables (`.env`):

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://pews:pews@localhost:5432/pews?sslmode=disable` |
| `JWT_SECRET` | Secret for signing JWTs | `your-secret-key-here` |
| `STRIPE_SECRET_KEY` | Stripe API key | `sk_test_xxx` |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook signing secret | `whsec_xxx` |
| `STRIPE_PRICE_ID` | Stripe price ID for Pro plan | `price_xxx` |
| `PORT` | Backend port | `8080` |
| `FRONTEND_URL` | Frontend URL for CORS/redirects | `http://localhost:5173` |

## Stripe Integration

### Setup

1. Create a Stripe account
2. Get your API keys from https://dashboard.stripe.com/apikeys
3. Create a product and price (recurring, $29/month)
4. Copy the price ID to `.env`
5. Setup webhook endpoint: `https://your-domain.com/api/billing/webhook`
6. Copy webhook signing secret to `.env`

### Webhook Events

The app handles:
- `checkout.session.completed` - Subscription created
- `invoice.paid` - Subscription renewed
- `invoice.payment_failed` - Payment failed
- `customer.subscription.deleted` - Subscription canceled

## Available Modules

- **People** - Member management
- **Giving** - Donation tracking
- **Services** - Worship planning
- **Groups** - Small groups
- **Check-ins** - Children's ministry

Modules are hardcoded in `internal/module/registry.go`. Each tenant can enable/disable modules.

## Testing

**Login with test credentials:**
1. Start dev environment: `make dev`
2. Run seed: `make seed`
3. Go to http://localhost:5173/login
4. Use: `test-church` / `admin@testchurch.com` / `password123`

## Troubleshooting

**Database connection refused:**
- Ensure PostgreSQL is running: `docker-compose ps`
- Check DATABASE_URL in `.env`

**Frontend can't reach API:**
- Ensure backend is running on `:8080`
- Check VITE_API_URL in frontend

**Migrations fail:**
- Drop database and restart: `make clean && make dev`

## Production Deployment

1. Set strong `JWT_SECRET`
2. Use production Stripe keys
3. Set proper `FRONTEND_URL`
4. Enable SSL/TLS
5. Use managed PostgreSQL (RDS, Cloud SQL, etc.)
6. Setup proper CORS origins
7. Configure webhook endpoints in Stripe dashboard

## Next Steps

- Add actual module implementations (people, giving, etc.)
- Add email verification
- Add password reset
- Add invite system
- Add audit logging
- Add API rate limiting
- Add comprehensive tests
