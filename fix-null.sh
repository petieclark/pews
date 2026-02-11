#!/bin/bash
# Script to apply all COALESCE NULL fixes

set -e

echo "Applying NULL handling fixes..."

# Groups service
echo "Fixing internal/groups/service.go..."
sed -i '' 's/SELECT g.id, g.tenant_id, g.name, g.description, g.group_type,/SELECT g.id, g.tenant_id, g.name, COALESCE(g.description, '\''\'\''), g.group_type,/g' internal/groups/service.go
sed -i '' 's/g.meeting_day, g.meeting_time, g.meeting_location,/COALESCE(g.meeting_day, '\''\'\''), COALESCE(g.meeting_time, '\''\'\''), COALESCE(g.meeting_location, '\''\'\''),/g' internal/groups/service.go
sed -i '' 's/, g.photo_url,/, COALESCE(g.photo_url, '\''\'\''),/g' internal/groups/service.go
sed -i '' 's/p.first_name, p.last_name, p.email, p.phone,/p.first_name, p.last_name, COALESCE(p.email, '\''\'\''), COALESCE(p.phone, '\''\'\''),/g' internal/groups/service.go
sed -i '' 's/p.address_line1, p.address_line2, p.city, p.state, p.zip,/COALESCE(p.address_line1, '\''\'\''), COALESCE(p.address_line2, '\''\'\''), COALESCE(p.city, '\''\'\''), COALESCE(p.state, '\''\'\''), COALESCE(p.zip, '\''\'\''),/g' internal/groups/service.go
sed -i '' 's/p.birthdate, p.gender, p.membership_status, p.photo_url, p.notes,/p.birthdate, COALESCE(p.gender, '\''\'\''), p.membership_status, COALESCE(p.photo_url, '\''\'\''), COALESCE(p.notes, '\''\'\''),/g' internal/groups/service.go

# Services service
echo "Fixing internal/services/service.go..."
sed -i '' 's/name, default_time, default_day, color/name, COALESCE(default_time, '\''\'\''), COALESCE(default_day, '\''\'\''), color/g' internal/services/service.go
sed -i '' 's/s.service_type_id, s.name, s.service_date,/s.service_type_id, COALESCE(s.name, '\''\'\''), s.service_date,/g' internal/services/service.go
sed -i '' 's/s.service_time, s.notes, s.status/COALESCE(s.service_time, '\''\'\''), COALESCE(s.notes, '\''\'\''), s.status/g' internal/services/service.go
sed -i '' 's/item_type, title, song_id, song_key, position, duration_minutes, notes, assigned_to/item_type, title, song_id, COALESCE(song_key, '\''\'\''), position, COALESCE(duration_minutes, 0), COALESCE(notes, '\''\'\''), COALESCE(assigned_to, '\''\'\'')/ g' internal/services/service.go
sed -i '' 's/st.role, st.status, st.notes,/st.role, st.status, COALESCE(st.notes, '\''\'\''),/g' internal/services/service.go
sed -i '' 's/title, artist, default_key, tempo/title, COALESCE(artist, '\''\'\''), COALESCE(default_key, '\''\'\''), COALESCE(tempo, 0)/g' internal/services/service.go
sed -i '' 's/ccli_number, lyrics, notes, tags/COALESCE(ccli_number, '\''\'\''), COALESCE(lyrics, '\''\'\''), COALESCE(notes, '\''\'\''), COALESCE(tags, '\''\'\'')/ g' internal/services/service.go

# Checkins service
echo "Fixing internal/checkins/service.go..."
sed -i '' 's/name, location, is_active/name, COALESCE(location, '\''\'\''), is_active/g' internal/checkins/service.go
sed -i '' 's/c.checked_in_at, c.checked_out_at, c.notes, c.created_at/c.checked_in_at, c.checked_out_at, COALESCE(c.notes, '\''\'\''), c.created_at/g' internal/checkins/service.go
sed -i '' 's/GROUP BY e.id ORDER/GROUP BY e.id, e.service_id, e.station_id ORDER/g' internal/checkins/service.go

# Communication service  
echo "Fixing internal/communication/service.go..."
sed -i '' 's/name, subject, body, channel/name, COALESCE(subject, '\''\'\''), body, channel/g' internal/communication/service.go
sed -i '' 's/channel, category, variables/channel, COALESCE(category, '\''\'\''), COALESCE(variables, '\''\'\'')/ g' internal/communication/service.go
sed -i '' 's/channel, subject, body, status,/channel, COALESCE(subject, '\''\'\''), body, status,/g' internal/communication/service.go
sed -i '' 's/target_type, target_id, created_at/target_type, COALESCE(target_id, '\''\'\''), created_at/g' internal/communication/service.go
sed -i '' 's/j.name, j.description, j.trigger_type, j.trigger_value/j.name, COALESCE(j.description, '\''\'\''), j.trigger_type, COALESCE(j.trigger_value, '\''\'\'')/ g' internal/communication/service.go
sed -i '' 's/first_name, last_name, email, phone, is_first_visit/first_name, COALESCE(last_name, '\''\'\''), COALESCE(email, '\''\'\''), COALESCE(phone, '\''\'\''), is_first_visit/g' internal/communication/service.go
sed -i '' 's/is_first_visit, how_heard, prayer_request, interested_in/is_first_visit, COALESCE(how_heard, '\''\'\''), COALESCE(prayer_request, '\''\'\''), COALESCE(interested_in, '\''\'\'')/ g' internal/communication/service.go

# Streaming service
echo "Fixing internal/streaming/service.go..."
sed -i '' 's/title, description, service_id/title, COALESCE(description, '\''\'\''), service_id/g' internal/streaming/service.go
sed -i '' 's/stream_url, stream_key, embed_url/COALESCE(stream_url, '\''\'\''), COALESCE(stream_key, '\''\'\''), COALESCE(embed_url, '\''\'\'')/ g' internal/streaming/service.go
sed -i '' 's/person_id, guest_name, message/person_id, COALESCE(guest_name, '\''\'\''), message/g' internal/streaming/service.go
sed -i '' 's/joined_at, left_at, duration_seconds/joined_at, left_at, COALESCE(duration_seconds, 0)/g' internal/streaming/service.go

echo "All fixes applied successfully!"
