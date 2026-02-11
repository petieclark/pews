# NULL Handling Audit - COALESCE Fixes Applied

## Summary
Fixed NULL handling across all Pews service files by adding COALESCE() to nullable columns that are scanned into Go strings.

## Pattern
Go's pgx cannot scan SQL NULL into `string` types. Solution: Use `COALESCE(column, '')` for strings, `COALESCE(column, 0)` for integers.

## Files Fixed

### ✅ internal/people/service.go
- `ListPeople()`: email, phone, address fields, photo_url, notes, gender
- `GetPersonByID()`: Same nullable columns
- `GetPersonHousehold()`: household address fields
- `ListHouseholds()`: household address fields

### ✅ internal/giving/service.go
- `ListFunds()`: description
- `GetFund()`: description
- `ListDonations()`: payment_method, stripe_payment_intent_id, stripe_charge_id, recurring_frequency, stripe_subscription_id, memo
- `GetDonation()`: Same donation nullable fields
- `ListRecurringDonations()`: payment_method, stripe_subscription_id, recurring_frequency

### ✅ internal/groups/service.go
- `ListGroups()`: description, meeting_day, meeting_time, meeting_location, photo_url
- `GetGroupByID()`: Same group nullable columns
- `GetGroupMembers()`: All person nullable columns (email, phone, address, photo_url, notes, gender)
- `GetPersonGroups()`: description, meeting fields, photo_url

### ⏳ internal/services/service.go (NEEDS FIXING)
**ListServiceTypes():**
- default_time, default_day

**ListServices(), GetUpcomingServices():**
- name, service_time, notes

**GetServiceByID():**
- name, service_time, notes
- service_types: default_time, default_day

**GetServiceItems():**
- song_key, duration_minutes (use 0), notes, assigned_to

**GetServiceTeam():**
- notes

**ListSongs(), GetSongByID():**
- artist, default_key, tempo (use 0), ccli_number, lyrics, notes, tags

### ⏳ internal/checkins/service.go (NEEDS FIXING)
**ListStations():**
- location

**ListEvents(), GetEvent():**
- service_id, station_id (UUIDs - keep nullable)
- Fix GROUP BY to include service_id, station_id

**GetAttendees(), GetPersonHistory():**
- notes (already has COALESCE for person email)

**SearchPeople():**
- Also add COALESCE in WHERE clause for email, phone

### ⏳ internal/communication/service.go (NEEDS FIXING)
**ListTemplates():**
- subject, category, variables

**ListCampaigns(), GetCampaign():**
- subject, target_id

**ListJourneys(), GetJourney():**
- description, trigger_value

**ListConnectionCards(), GetConnectionCard():**
- last_name, email, phone, how_heard, prayer_request, interested_in

### ⏳ internal/streaming/service.go (NEEDS FIXING)
**All stream SELECT queries (ListStreams, GetStreamByID, GetStreamByIDPublic, GetLiveStream):**
- description, stream_url, stream_key, embed_url

**GetChatMessages():**
- guest_name

**GetViewers():**
- guest_name, duration_seconds (use 0)

## Testing Checklist
- [x] `docker build` - Compilation successful
- [ ] Start local: `docker compose up -d`
- [ ] Test endpoints:
  - [ ] GET /api/people
  - [ ] GET /api/giving/dashboard
  - [ ] GET /api/groups
  - [ ] GET /api/services
  - [ ] GET /api/checkins/dashboard
  - [ ] GET /api/communication/dashboard
  - [ ] GET /api/streaming

## Next Steps
1. Apply remaining fixes to services, checkins, communication, streaming
2. Test compilation
3. Test all API endpoints
4. Commit to fix/null-audit branch
5. **DO NOT MERGE TO MAIN** - Needs testing first
