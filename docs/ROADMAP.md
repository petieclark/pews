# Pews — Church Management, Simplified
*pews.app — The church platform that gets out of your way.*

## The Problem
PlanningCenter serves 90,000+ churches. Most of them use 30% of the features and pay for 100%. Every module is another bill. Every feature is another training session. Every integration is another Zapier recipe.

Churches don't need more software. They need less — that actually works.

## The Pews Philosophy

### Dead Simple
- Sign up. Import your people. Go.
- No module math. No config wizards. No "PlanningCenter admin" role needed.
- If your church secretary can use Facebook, she can use Pews.

### Only What You Need
- **Modular by design** — features you don't need don't exist in your instance
- Admin enables "Check-Ins" → we spin up the service, add the nav item, ready to go
- Disable it → gone. No clutter, no wasted resources, no confusion.
- Every church gets a clean, focused experience tailored to what they actually use.

### No Surprise Costs
- **One flat price. Period.**
- Email sending? Included.
- SMS? Included.
- Online giving? Stripe processing + 1% platform fee. That's it.
- No per-module pricing. No per-seat charges. No "contact sales for enterprise."

## Pricing Model

### Free Tier
- People database (unlimited members)
- Basic communication (email)
- 1 admin user

### Pews Pro — $100/month
- All modules available (enable what you need)
- Unlimited admin users
- SMS included (fair use)
- Email included
- Priority support
- Custom domain (yourchurch.pews.app)

### Online Giving
- Stripe processing: 2.15% + $0.30 (Stripe's rate)
- Pews platform fee: +1%
- Total: 3.15% + $0.30 per transaction
- ACH/bank transfer: 0% + $0.30 + 1% platform fee
- Tax-deductible receipts generated automatically

### Why This Works Financially
| Metric | Per Church | At 100 Churches | At 500 Churches | At 1000 Churches |
|--------|-----------|-----------------|-----------------|------------------|
| Monthly subscription | $100 | $10,000 | $50,000 | $100,000 |
| Avg annual giving | $200K | — | — | — |
| 1% platform fee/yr | $2,000 | $200,000 | $1,000,000 | $2,000,000 |
| SMS/email cost/mo | ~$10 | ~$1,000 | ~$5,000 | ~$10,000 |
| **Net monthly revenue** | ~$267 | ~$26,667 | ~$133,333 | ~$266,667 |

The giving fee is the real business. Subscriptions cover infrastructure. Giving fees scale linearly with church size — bigger churches pay more naturally without tiered pricing.

## Architecture — Modular Containers

### Core (Always Running)
```
┌─────────────────────────────────────┐
│  API Gateway / Auth / Router        │
│  People Database (PostgreSQL)       │
│  Web UI Shell                       │
│  Notification Service (email/SMS)   │
└─────────────────────────────────────┘
```

### Optional Modules (Spin Up On Enable)
```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Services    │  │  Groups      │  │  Check-Ins   │
│  (worship    │  │  (small      │  │  (child      │
│   planning)  │  │   groups)    │  │   safety)    │
└──────────────┘  └──────────────┘  └──────────────┘

┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Giving      │  │  Calendar    │  │  Church App  │
│  (Stripe     │  │  (room       │  │  (PWA for    │
│   Connect)   │  │   booking)   │  │   members)   │
└──────────────┘  └──────────────┘  └──────────────┘
```

### How Modular Deployment Works
1. Church admin goes to Settings → Modules
2. Toggles "Enable Check-Ins" → ON
3. Orchestrator spins up check-ins container for their tenant
4. Routes registered, nav item appears, feature is live
5. Toggle OFF → container stops, resources freed, nav item gone

**Resource savings at scale:**
- Full stack: ~8 containers per tenant
- Typical church (People + Giving + Groups): ~4 containers
- Minimal church (People only): ~2 containers
- At 500 tenants, this is 2,000 vs 4,000 containers — real infrastructure savings

### Tech Stack
- **Backend:** Go (fast, single binary per module, excellent container story)
- **Database:** PostgreSQL (multi-tenant, row-level security)
- **Auth:** Built-in (simple email/password) + optional SSO
- **API:** REST (simple, documented, webhook-capable)
- **Frontend:** SvelteKit (fast, lightweight, SSR)
- **Mobile:** PWA first, Flutter native later
- **Orchestration:** Docker Compose (self-hosted) / Kubernetes (managed platform)
- **Payments:** Stripe Connect (platform model with application fees)
- **Email:** Mailgun (subsidized)
- **SMS:** Twilio (subsidized)

### Multi-Tenancy Strategy
- **Managed (pews.app):** Shared PostgreSQL with row-level security, shared infrastructure
- **Self-hosted:** Single-tenant Docker Compose, full data ownership
- Both use identical container images — no code fork

## Modules

### People (Core — Always On)
The foundation. Every church needs this.
- Member database with households
- Tags, lists, custom fields
- Contact info, family relationships
- Communication (email + SMS)
- Search and filtering
- Import from PlanningCenter CSV
- Activity timeline per person

### Giving (High Priority)
Where the revenue model lives.
- Stripe Connect integration
- Online giving (one-time + recurring)
- Fund/designation management
- Automatic tax-deductible receipts
- Annual giving statements (one-click generate)
- Giving analytics (trends, retention)
- Text-to-give via Twilio
- ACH bank transfer support

### Services (High Priority)
Sunday morning coordination.
- Service templates (order of worship)
- Team scheduling with conflict detection
- Song database with lyrics, keys, notes
- Volunteer request/accept/decline
- Calendar view
- ProPresenter export

### Groups (High Priority)
Small group management.
- Group creation with leaders + members
- Meeting schedule
- Leader dashboard
- Member self-service sign-up
- Group finder (public, filterable)
- Group communication tools

### Check-Ins (Medium Priority)
Child safety and attendance.
- QR code / NFC check-in
- Child safety: parent receipt, authorized pickup
- Allergy/medical alerts
- Attendance tracking
- Volunteer ratio monitoring
- Kiosk mode (iPad at the door)

### Calendar (Medium Priority)
Facility and event management.
- Room booking with conflict detection
- Resource management
- Public calendar (iCal export)
- Approval workflows

### Church App (Later)
Member-facing PWA.
- Announcements + push notifications
- Sermon notes and audio
- Event registration
- Online giving
- Group finder
- Prayer request submission

## Competitive Positioning

| | PlanningCenter | Pews |
|---|---|---|
| **Pricing** | $300-500+/mo (PCO + Gloo combined) | $100/mo flat (all modules) |
| **Modules** | Pay for each | Enable what you need, included |
| **Giving fees** | 2.15% + $0.30 | 3.15% + $0.30 (covers platform) |
| **SMS** | $0.02/text extra | Included |
| **Email** | Included | Included |
| **Complexity** | High (90K church feature bloat) | Low (only see what you use) |
| **Self-hosted** | No | Yes (Docker Compose) |
| **Data ownership** | Theirs | Yours |
| **Setup time** | Hours/days | Minutes |

## Go-To-Market

### Phase 1: Destination Church (Dogfood)
- Build for your own church first
- Pastor and admin feedback loop
- Prove the People + Giving + Groups flow

### Phase 2: Local Churches (10 beta)
- Invite 10 churches in Southeast Georgia
- Free during beta
- Weekly feedback calls
- Iterate fast

### Phase 3: Launch (pews.app)
- Landing page with demo
- Self-service sign-up
- PlanningCenter migration tool
- Content marketing: "Why we left PlanningCenter"

### Phase 4: Growth
- Church conference presence
- Referral program (churches refer churches)
- YouTube tutorials
- Denominational partnerships

## Integration with Mutiny
- Shared infrastructure patterns (Docker Compose on same servers)
- Optional Mutiny integration for volunteer/team communication
- Not a dependency — Pews stands alone

## Development Timeline
- **Month 1-2:** People + Auth + Core architecture + modular container system
- **Month 3:** Giving (Stripe Connect integration)
- **Month 4:** Services + Groups
- **Month 5:** Check-Ins + Calendar
- **Month 6:** Church App (PWA) + PlanningCenter migration tool
- **Month 7:** Beta launch with 10 churches
- **Month 8:** Public launch on pews.app

## CTM Business Angle
- Pews becomes a CTM product line
- Migration consulting: $2-5K per church (PlanningCenter → Pews)
- Managed hosting revenue is recurring
- Church market is massive (300K+ churches in US alone)
- Low churn (churches don't switch platforms often)
- Word-of-mouth is the #1 acquisition channel in church world

---

*Pews: Sit down. We've got this.* 🪵
