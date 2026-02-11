# NULL Handling Fixes Applied

## Files Modified

### internal/services/service.go
- ✅ **ListServiceTypes**: Added COALESCE for default_time, default_day, color
- ⚠️ **GetServiceItems**: Need COALESCE for title, song_key, duration_minutes, notes
- ⚠️ **GetServiceTeam**: Need COALESCE for notes
- ⚠️ **GetServiceByID**: Need COALESCE for name, service_time, notes
- ⚠️ **ListSongs**: Need COALESCE for artist, default_key, tempo, ccli_number, lyrics, notes, tags, times_used
- ⚠️ **GetSongByID**: Need COALESCE for artist, default_key, tempo, ccli_number, lyrics, notes, tags, times_used

### internal/communication/service.go
- ⚠️ **ListTemplates**: Need COALESCE for subject, category, variables
- ⚠️ **ListCampaigns**: Need COALESCE for subject, recipient_count, opened_count, clicked_count, target_type
- ⚠️ **GetCampaign**: Need COALESCE for subject, recipient_count, opened_count, clicked_count, target_type
- ⚠️ **ListJourneys**: Need COALESCE for description, trigger_value
- ⚠️ **GetJourney**: Need COALESCE for description, trigger_value
- ⚠️ **GetJourneySteps**: Need COALESCE for delay_days, delay_hours, config
- ⚠️ **ListConnectionCards**: Need COALESCE for email, phone, how_heard, prayer_request, interested_in
- ⚠️ **GetConnectionCard**: Need COALESCE for email, phone, how_heard, prayer_request, interested_in

### internal/calendar/service.go
- ⚠️ **ListEvents**: Need COALESCE for description, location, recurring, color, created_by
- ⚠️ **GetEventByID**: Need COALESCE for description, location, recurring, color, created_by

### internal/import/handler.go
- ✅ **ImportPCO**: Fixed undefined function call (ParsePCOPeopleCSV → ParsePeopleCSV)

## Status
- Branch: fix/null-handling-and-bugs
- Compilation: Successfully compiles after ListServiceTypes fix
- Remaining: Need to apply all other COALESCE fixes systematically

## Next Steps
1. Apply remaining COALESCE fixes to services/service.go (5 functions)
2. Apply COALESCE fixes to communication/service.go (8 functions)
3. Apply COALESCE fixes to calendar/service.go (2 functions)
4. Test compilation
5. Commit with message covering all issues #4, #5, #6
