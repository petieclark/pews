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
| Core | Auth, tenant management, module registry, billing | 🔴 Not started |
| People | Member database, households, communication | 🔴 Not started |
| Giving | Online donations, recurring gifts, tax receipts | 🔴 Not started |
| Services | Worship planning, team scheduling | 🔴 Not started |
| Groups | Small group management | 🔴 Not started |
| Check-Ins | Child safety, attendance tracking | 🔴 Not started |

## Development

```bash
# Prerequisites: Go 1.22+, Node 20+, PostgreSQL 16+, Docker

# Start local dev environment
docker compose up -d postgres
go run cmd/pews/main.go

# Frontend
cd web && npm run dev
```

## License

Proprietary — © 2026 Clearline Technology Methods
