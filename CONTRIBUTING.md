# Contributing to Pews

Thank you for your interest in contributing to Pews! This guide will help you get started with development.

## Table of Contents

- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Architecture Overview](#architecture-overview)
- [Adding a New Module](#adding-a-new-module)
- [Testing](#testing)
- [Git Workflow](#git-workflow)
- [Pull Request Process](#pull-request-process)

## Development Setup

### Prerequisites

- **Go 1.22+**
- **Node.js 20+**
- **PostgreSQL 16+**
- **Docker & Docker Compose** (optional but recommended)
- **Git**

### Local Development Environment

#### Option 1: Docker Compose (Recommended)

\`\`\`bash
# Clone the repository
git clone git@github.com:warpapaya/Pews.git
cd Pews

# Start all services
docker compose up -d

# View logs
docker compose logs -f backend
docker compose logs -f frontend

# Stop services
docker compose down
\`\`\`

#### Option 2: Manual Setup

**1. Set up PostgreSQL:**

\`\`\`bash
# Create database
createdb pews

# Run migrations (they'll auto-run on backend startup, but you can run manually)
psql pews < internal/database/migrations/001_tenants.sql
psql pews < internal/database/migrations/002_users.sql
# ... etc
\`\`\`

**2. Configure environment:**

\`\`\`bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your values
nano .env
\`\`\`

Required environment variables:
- \`DATABASE_URL\` - PostgreSQL connection string
- \`JWT_SECRET\` - Secret key for JWT signing (use a strong random value)
- \`STRIPE_SECRET_KEY\` - Stripe API key (use test key for development: \`sk_test_...\`)
- \`STRIPE_WEBHOOK_SECRET\` - Stripe webhook secret
- \`STRIPE_PRICE_ID\` - Stripe price ID for subscriptions
- \`PORT\` - Backend port (default: 8190)
- \`FRONTEND_URL\` - Frontend URL (default: http://localhost:5173)

**3. Start the backend:**

\`\`\`bash
# Install Go dependencies
go mod download

# Run the backend
go run cmd/pews/main.go

# Backend will start on http://localhost:8190 (or PORT from .env)
\`\`\`

**4. Start the frontend:**

\`\`\`bash
cd web

# Install dependencies
npm install

# Start dev server
npm run dev

# Frontend will start on http://localhost:5173
\`\`\`

**5. Load demo data (optional):**

\`\`\`bash
# Seed database with realistic demo data
psql pews < scripts/seed-demo.sql
\`\`\`

### Testing the Setup

\`\`\`bash
# Health check
curl http://localhost:8190/health

# Register a test church
curl -X POST http://localhost:8190/api/auth/register \\
  -H "Content-Type: application/json" \\
  -d '{"tenant_name":"Test Church","email":"test@example.com","password":"password123"}'

# You should receive a JWT token in the response
\`\`\`

## Code Style

### Go Backend

We follow standard Go conventions with some specific guidelines.

See full Go style guide in the main CONTRIBUTING.md.

### SvelteKit Frontend

Use TypeScript, follow Svelte best practices, and use TailwindCSS for styling.

See full frontend style guide in the main CONTRIBUTING.md.

## Git Workflow

### Branch Naming

- \`feat/feature-name\` - New features
- \`fix/bug-description\` - Bug fixes  
- \`docs/what-changed\` - Documentation updates
- \`refactor/what-refactored\` - Code refactoring
- \`test/what-tested\` - Adding tests
- \`chore/what-changed\` - Maintenance tasks

### Commit Messages

Follow conventional commits:

\`\`\`
<type>: <description>

[optional body]
\`\`\`

**Types:** feat, fix, docs, style, refactor, test, chore

## Pull Request Process

1. Code follows style guidelines
2. All tests pass
3. New features include tests
4. Database migrations properly numbered with RLS policies
5. Documentation updated
6. No console.log() or debug statements
7. Dark mode styling for frontend changes

---

Happy coding! 🚀

For full details, see [CONTRIBUTING.md](./CONTRIBUTING.md) in the repository.
