# Pews API Endpoints Reference

Complete list of all API endpoints organized by module.

## Legend
- 🔓 Public (no authentication required)
- 🔒 Authenticated
- 👑 Admin only

---

## Authentication

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | 🔓 | Register new tenant + admin user |
| POST | `/auth/login` | 🔓 | Authenticate user, get JWT token |
| POST | `/auth/logout` | 🔒 | Logout (client-side token deletion) |

---

## Tenant Management

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/tenant` | 🔒 | Get current tenant details |
| PUT | `/tenant` | 👑 | Update tenant name/domain |

---

## Modules

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/tenant/modules` | 🔒 | List all available modules with enabled status |
| POST | `/tenant/modules/{name}/enable` | 👑 | Enable a module |
| POST | `/tenant/modules/{name}/disable` | 👑 | Disable a module |

---

## Billing

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/billing/subscription` | 🔒 | Get current subscription details |
| POST | `/billing/checkout` | 👑 | Create Stripe checkout session |
| POST | `/billing/portal` | 👑 | Create Stripe customer portal session |
| POST | `/billing/webhook` | 🔓 | Stripe webhook handler |

---

## People

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/people` | 🔒 | List people (supports search, pagination) |
| POST | `/people` | 🔒 | Create new person |
| GET | `/people/{id}` | 🔒 | Get person details |
| PUT | `/people/{id}` | 🔒 | Update person |
| DELETE | `/people/{id}` | 🔒 | Delete person |
| POST | `/people/{id}/tags` | 🔒 | Add tag to person |
| DELETE | `/people/{id}/tags/{tagId}` | 🔒 | Remove tag from person |

### Tags

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/tags` | 🔒 | List all tags |
| POST | `/tags` | 🔒 | Create tag |

### Households

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/households` | 🔒 | List households |
| POST | `/households` | 🔒 | Create household |
| PUT | `/households/{id}` | 🔒 | Update household |
| POST | `/households/{id}/members` | 🔒 | Add member to household |
| DELETE | `/households/{id}/members/{personId}` | 🔒 | Remove member from household |

---

## Groups

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/groups` | 🔒 | List groups (filter by type, active status) |
| POST | `/groups` | 🔒 | Create group |
| GET | `/groups/{id}` | 🔒 | Get group details |
| PUT | `/groups/{id}` | 🔒 | Update group |
| DELETE | `/groups/{id}` | 🔒 | Delete group |
| GET | `/groups/{id}/members` | 🔒 | Get group members |
| POST | `/groups/{id}/members` | 🔒 | Add member to group |
| PUT | `/groups/{id}/members/{memberId}` | 🔒 | Update member role |
| DELETE | `/groups/{id}/members/{memberId}` | 🔒 | Remove member from group |
| GET | `/groups/person/{personId}` | 🔒 | Get all groups for a person |

---

## Services

### Service Types

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/services/types` | 🔒 | List service types |
| POST | `/services/types` | 🔒 | Create service type |
| PUT | `/services/types/{id}` | 🔒 | Update service type |

### Services

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/services` | 🔒 | List services (filter by date range, type, status) |
| POST | `/services` | 🔒 | Create service |
| GET | `/services/upcoming` | 🔒 | Get upcoming services |
| GET | `/services/{id}` | 🔒 | Get service details |
| PUT | `/services/{id}` | 🔒 | Update service |
| DELETE | `/services/{id}` | 🔒 | Delete service |

### Service Items

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/services/{id}/items` | 🔒 | Get service items (order of service) |
| POST | `/services/{id}/items` | 🔒 | Add item to service |
| PUT | `/services/{id}/items/{itemId}` | 🔒 | Update service item |
| DELETE | `/services/{id}/items/{itemId}` | 🔒 | Delete service item |

### Service Team

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/services/{id}/team` | 🔒 | Get service team members |
| POST | `/services/{id}/team` | 🔒 | Add team member to service |
| PUT | `/services/{id}/team/{teamId}` | 🔒 | Update team member |
| DELETE | `/services/{id}/team/{teamId}` | 🔒 | Remove team member |

### Songs

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/services/songs` | 🔒 | List songs (search, pagination) |
| POST | `/services/songs` | 🔒 | Create song |
| PUT | `/services/songs/{id}` | 🔒 | Update song |
| DELETE | `/services/songs/{id}` | 🔒 | Delete song |

---

## Giving

### Funds

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/giving/funds` | 🔒 | List giving funds |
| POST | `/giving/funds` | 👑 | Create fund |
| PUT | `/giving/funds/{id}` | 👑 | Update fund |

### Donations

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/giving/donations` | 🔒 | List donations (filter by person, fund, date range) |
| POST | `/giving/donations` | 👑 | Create donation (manual entry) |
| GET | `/giving/donations/{id}` | 🔒 | Get donation details |
| GET | `/giving/stats` | 🔒 | Get giving statistics |
| GET | `/giving/person/{personId}` | 🔒 | Get person's giving history |
| GET | `/giving/recurring` | 🔒 | List recurring donations |
| POST | `/giving/statements/{year}` | 👑 | Generate annual giving statement |

### Stripe Connect

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/giving/connect/onboard` | 👑 | Create Stripe Connect onboarding link |
| GET | `/giving/connect/status` | 🔒 | Get Stripe Connect account status |
| GET | `/giving/connect/return` | 🔒 | Handle return from Stripe onboarding |
| GET | `/giving/connect/refresh` | 👑 | Refresh onboarding link |
| POST | `/giving/checkout` | 🔒 | Create donation checkout session |
| POST | `/giving/webhook` | 🔓 | Stripe giving webhook handler |

---

## Streaming

### Streams

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/streaming` | 🔒 | List streams (filter by status) |
| POST | `/streaming` | 🔒 | Create stream |
| GET | `/streaming/live` | 🔒 | Get current live stream |
| GET | `/streaming/{id}` | 🔒 | Get stream details |
| PUT | `/streaming/{id}` | 🔒 | Update stream |
| DELETE | `/streaming/{id}` | 🔒 | Delete stream |
| POST | `/streaming/{id}/go-live` | 🔒 | Set stream to live status |
| POST | `/streaming/{id}/end` | 🔒 | End stream |
| GET | `/streaming/watch/{id}` | 🔓 | Public watch page data |

### Chat

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/streaming/{id}/chat` | 🔓 | Get chat messages (public) |
| POST | `/streaming/{id}/chat` | 🔓 | Send chat message (public or authenticated) |
| PUT | `/streaming/{id}/chat/{msgId}/pin` | 🔒 | Pin chat message (admin) |
| DELETE | `/streaming/{id}/chat/{msgId}` | 🔒 | Delete chat message (admin) |

### Viewers

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/streaming/{id}/join` | 🔓 | Join stream as viewer (public) |
| POST | `/streaming/{id}/leave` | 🔓 | Leave stream |
| GET | `/streaming/{id}/viewers` | 🔒 | Get viewer list (admin) |

### Notes

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/streaming/{id}/notes` | 🔒 | Get user's stream notes |
| POST | `/streaming/{id}/notes` | 🔒 | Save stream notes |

---

## Communication

### Templates

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/communication/templates` | 🔒 | List message templates (filter by channel, category) |
| POST | `/communication/templates` | 🔒 | Create template |
| PUT | `/communication/templates/{id}` | 🔒 | Update template |
| DELETE | `/communication/templates/{id}` | 🔒 | Delete template |

### Campaigns

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/communication/campaigns` | 🔒 | List campaigns (filter by status) |
| POST | `/communication/campaigns` | 🔒 | Create campaign |
| GET | `/communication/campaigns/{id}` | 🔒 | Get campaign details |
| PUT | `/communication/campaigns/{id}` | 🔒 | Update campaign |
| POST | `/communication/campaigns/{id}/send` | 🔒 | Send/schedule campaign |
| GET | `/communication/campaigns/{id}/recipients` | 🔒 | Get campaign recipients |

### Journeys

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/communication/journeys` | 🔒 | List automated journeys |
| POST | `/communication/journeys` | 🔒 | Create journey |
| GET | `/communication/journeys/{id}` | 🔒 | Get journey details |
| PUT | `/communication/journeys/{id}` | 🔒 | Update journey |
| DELETE | `/communication/journeys/{id}` | 🔒 | Delete journey |
| POST | `/communication/journeys/{id}/steps` | 🔒 | Add journey step |
| PUT | `/communication/journeys/{id}/steps/{stepId}` | 🔒 | Update journey step |
| DELETE | `/communication/journeys/{id}/steps/{stepId}` | 🔒 | Delete journey step |
| POST | `/communication/journeys/{id}/enroll` | 🔒 | Enroll person in journey |
| GET | `/communication/journeys/{id}/enrollments` | 🔒 | Get journey enrollments |

### Connection Cards

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/communication/cards` | 🔓 | Submit connection card (public) |
| GET | `/communication/cards` | 🔒 | List connection cards |
| GET | `/communication/cards/{id}` | 🔒 | Get connection card details |
| POST | `/communication/cards/{id}/process` | 🔒 | Process card (link to person) |

### Stats

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/communication/stats` | 🔒 | Get communication statistics |

---

## Check-ins

### Stations

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/checkins/stations` | 🔒 | List check-in stations |
| POST | `/checkins/stations` | 🔒 | Create station |
| PUT | `/checkins/stations/{id}` | 🔒 | Update station |

### Events

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/checkins/events` | 🔒 | List check-in events |
| POST | `/checkins/events` | 🔒 | Create event |
| GET | `/checkins/events/{id}` | 🔒 | Get event details |
| PUT | `/checkins/events/{id}` | 🔒 | Update event |
| POST | `/checkins/events/{id}/checkin` | 🔒 | Check in person to event |
| POST | `/checkins/events/{id}/checkout` | 🔒 | Check out person from event |
| GET | `/checkins/events/{id}/attendees` | 🔒 | Get event attendees |

### Person History

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/checkins/person/{personId}/history` | 🔒 | Get person's check-in history |

### Medical Alerts

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/checkins/person/{personId}/alerts` | 🔒 | Get person's medical alerts |
| POST | `/checkins/person/{personId}/alerts` | 🔒 | Create medical alert |
| DELETE | `/checkins/person/{personId}/alerts/{alertId}` | 🔒 | Delete medical alert |

### Authorized Pickups

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/checkins/person/{personId}/pickups` | 🔒 | Get authorized pickups for child |
| POST | `/checkins/person/{personId}/pickups` | 🔒 | Add authorized pickup |
| DELETE | `/checkins/person/{personId}/pickups/{pickupId}` | 🔒 | Remove authorized pickup |

### Stats & Search

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/checkins/stats` | 🔒 | Get check-in statistics |
| GET | `/checkins/search` | 🔒 | Search people for check-in (query param: `q`) |

---

## Summary

- **Total Endpoints:** 130+
- **Public Endpoints:** 11
- **Authenticated Endpoints:** 110+
- **Admin-Only Endpoints:** 9
- **Modules:** 11
