# Pews

**Church management, simplified.**

One flat price. Only the features you need. No surprises.

🌐 [pews.app](https://pews.app)

## Architecture

Pews is a modular church management platform. Each feature (People, Giving, Groups, Services, Check-Ins) runs as an independent module that can be enabled/disabled per tenant.

- **Backend:** Go
- **Frontend:** SvelteKit  
- **Database:** PostgreSQL (multi-tenant, row-level security)
- **Payments:** Stripe Connect

## Modules

| Module | Description | Status |
|--------|-------------|--------|
| Core | Auth, tenant management, module registry, billing | ✅ Complete |
| People | Member database, households, communication | 🔴 Not started |
| Giving | Online donations, recurring gifts, tax receipts | 🔴 Not started |
| Services | Worship planning, team scheduling | 🔴 Not started |
| Groups | Small group management | 🔴 Not started |
| Check-Ins | Child safety, attendance tracking | 🔴 Not started |

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for full setup instructions.

**Quick start:**
```bash
# Prerequisites: Go 1.22+, Node 20+, Docker

# Copy environment config
cp .env.example .env

# Start everything (PostgreSQL + backend + frontend)
make dev

# Seed test data
make seed

# Access: http://localhost:5173
# Login: test-church / admin@testchurch.com / password123
```

## License

Proprietary — © 2026 Clearline Technology Methods
