# NULL Handling Audit for Pews

## Nullable Columns by Table (from migrations)

### people
- email, phone, address_line1, address_line2, city, state, zip, birthdate, gender, photo_url, notes

### households
- primary_contact_id, address_line1, address_line2, city, state, zip

### funds
- description

### donations
- person_id, payment_method, stripe_payment_intent_id, stripe_charge_id, recurring_frequency, stripe_subscription_id, memo

### giving_statements
- pdf_url

### groups
- description, meeting_day, meeting_time, meeting_location, max_members, photo_url

### service_types
- default_time, default_day

### services
- name, service_time, notes

### songs
- artist, default_key, tempo, ccli_number, lyrics, notes, tags, last_used

### service_items
- song_id, song_key, duration_minutes, notes, assigned_to

### service_teams
- notes

### checkin_stations
- location

### checkin_events
- service_id, station_id

### checkins
- station_id, checked_out_at, notes

### message_templates
- subject, category, variables

### campaigns
- template_id, subject, scheduled_at, sent_at, target_id

### campaign_recipients
- sent_at, opened_at, clicked_at

### journeys
- description, trigger_value

### journey_steps
- template_id

### journey_enrollments
- next_step_at, completed_at

### connection_cards
- last_name, email, phone, how_heard, prayer_request, interested_in, person_id

### streams
- description, service_id, scheduled_start, actual_start, actual_end, stream_url, stream_key, embed_url

### stream_chat
- person_id, guest_name

### stream_viewers
- person_id, guest_name, left_at, duration_seconds

### stream_notes
- person_id

### tenants
- stripe_account_id

## Files to Fix

### internal/giving/service.go
- [ ] ListFunds - description
- [ ] GetFund - description
- [ ] ListDonations - payment_method, stripe_payment_intent_id, stripe_charge_id, recurring_frequency, stripe_subscription_id, memo
- [ ] GetDonation - payment_method, stripe_payment_intent_id, stripe_charge_id, recurring_frequency, stripe_subscription_id, memo
- [ ] ListRecurringDonations - payment_method, stripe_subscription_id, recurring_frequency

### internal/giving/stripe.go
- [ ] GetConnectStatus - stripe_account_id

### internal/communication/service.go
- [ ] ListTemplates - subject, category, variables
- [ ] CreateTemplate - subject, category, variables
- [ ] ListCampaigns - template_id, subject, scheduled_at, sent_at, target_id
- [ ] GetCampaign - template_id, subject, scheduled_at, sent_at, target_id
- [ ] GetCampaignRecipients - sent_at, opened_at, clicked_at
- [ ] ListJourneys - description, trigger_value
- [ ] CreateJourney - description, trigger_value
- [ ] GetJourney - description, trigger_value
- [ ] GetJourneySteps - template_id
- [ ] GetJourneyEnrollments - next_step_at, completed_at
- [ ] SubmitConnectionCard - last_name, email, phone, how_heard, prayer_request, interested_in
- [ ] ListConnectionCards - last_name, email, phone, how_heard, prayer_request, interested_in, person_id
- [ ] GetConnectionCard - last_name, email, phone, how_heard, prayer_request, interested_in, person_id

### internal/streaming/service.go
- To be audited

### internal/checkins/service.go
- To be audited

### internal/groups/service.go
- To be audited

### internal/services/service.go
- To be audited

### internal/people/service.go
- To be audited
