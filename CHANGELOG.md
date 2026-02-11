# Changelog

All notable changes to the Pews project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Major Features Added (Feb 10-11, 2026 Build Sprint)

#### New Modules

- **Calendar/Events Module** - Event management, recurring events, color-coding, location tracking
- **Prayer Requests Module** - Public/private prayer submissions, staff follow-up, status tracking, prayer followers
- **Reports/Analytics Dashboard** - Attendance trends, giving insights, membership growth, group participation metrics
- **Notifications System** - In-app notifications with read/unread status, deep links, user-specific delivery

#### Enhancements

- **Church Profile & Branding** - Custom church settings, logo, colors, public-facing information
- **Online Giving Portal** - Dedicated public giving page with Stripe Checkout
- **Recurring Giving** - Scheduled donations (weekly/monthly/yearly) with donor management
- **Volunteer Scheduling Backend** - Teams, positions, assignments, availability tracking
- **API Documentation** - Comprehensive RESTful API docs with examples
- **CSV/JSON Import** - Bulk data import (Phase 1 of PCO import)
- **HTML Email Templates** - Rich email support for Communication module
- **PDF Tax Statements** - Year-end giving statements

#### Infrastructure & DevOps

- **Production Hardening** - Traefik reverse proxy, Let's Encrypt SSL, Docker secrets, health checks
- **Automated API Testing** - Comprehensive test suite for all endpoints

#### Accessibility & UX

- **WCAG 2.1 AA Compliance** - Keyboard navigation, screen reader support, ARIA labels
- **SEO Optimizations** - Meta tags, Open Graph, structured data, sitemap
- **Error Handling** - Global error pages (404, 500), graceful error handling
- **Enhanced Public Pages** - Polished watch, connect, give pages
- **Demo Seed Data** - Comprehensive, realistic demo data script

### Bug Fixes

- Fixed NULL handling across all modules with COALESCE
- Fixed dark mode styling across all dashboard pages
- Fixed infinite spinner on Giving page
- Fixed stream date format and service type dropdown
- Fixed Svelte compilation errors
- Auto-create subscription on registration
- Fixed CORS origin configuration
- Fixed ambiguous status column in communication stats

### Improvements

- **Row-Level Security (RLS) Middleware** - Automatic tenant isolation
- **Mobile Responsive Design** - All dashboard pages optimized
- **Dashboard Enhancements** - Real stats, quick actions, improved navigation
- **Song Library** - Full CRUD, advanced search, usage tracking
- **Onboarding Flow** - Polished registration and welcome screens
- **Module Navigation** - Hide disabled modules dynamically

## [0.9.0] - 2026-02-10

### Added - Core Platform Launch

#### Core Features

- Multi-tenant architecture with Row-Level Security (RLS)
- JWT authentication with bcrypt password hashing
- Stripe integration for billing ($100/mo Pro plan)
- Module system registry

#### Modules

- **People Module** - Member database, households, tags, custom fields
- **Giving Module** - Stripe Connect, donations, recurring gifts, fund management, tax receipts
- **Groups Module** - Small groups, ministry teams, attendance tracking
- **Services Module** - Worship planning, song library, team scheduling
- **Check-Ins Module** - Child safety, kiosk mode, medical alerts, attendance
- **Communication Module** - Email/SMS campaigns, automated journeys, connection cards
- **Streaming Module** - Live stream embedding, chat, viewer tracking, public watch page

#### User Interface

- SvelteKit frontend with TailwindCSS
- Dark mode support (system-aware with manual toggle)
- Mobile-optimized responsive design
- Multi-page marketing site

#### Developer Experience

- 11 database migrations with RLS policies
- Docker Compose for local development
- Production Docker Compose with Traefik
- Modular backend architecture

## [0.1.0] - 2026-02-09

### Added

- Initial repository setup
- Project README and roadmap
- Landing page with branding
- Basic project structure

---

## Links

- [GitHub Repository](https://github.com/warpapaya/Pews)
- [Issues](https://github.com/warpapaya/Pews/issues)
- [Website](https://pews.app)
- [Demo](https://demo.pews.app)
