# Drip Campaigns

Automated communication sequences triggered by specific events (new member, connection card, first visit).

## Overview

Drip campaigns automatically send a series of timed communications when triggered by specific events. Unlike general journeys, drip campaigns are specifically designed for onboarding new visitors and members.

## Features

- **Event-based triggers**: Automatically start when someone becomes a new member, submits a connection card, or visits for the first time
- **Timed steps**: Schedule communications at specific intervals (day 0, day 3, day 7, etc.)
- **Multiple action types**: 
  - Email (with subject line)
  - SMS
  - Follow-up tasks (reminders for staff)
- **Drag-and-drop step ordering**: Easily reorder steps in the campaign
- **Enrollment tracking**: See who's currently enrolled and their progress
- **Active/inactive toggle**: Pause campaigns without deleting them

## Database Schema

### `drip_campaigns`
- `id`: UUID
- `tenant_id`: UUID (references tenants)
- `name`: VARCHAR(255)
- `trigger_event`: VARCHAR(50) - `new_member`, `connection_card`, or `first_visit`
- `is_active`: BOOLEAN
- `created_at`, `updated_at`: TIMESTAMP

### `drip_steps`
- `id`: UUID
- `campaign_id`: UUID (references drip_campaigns)
- `step_order`: INTEGER - Order of execution
- `delay_days`: INTEGER - Days to wait before executing
- `action_type`: VARCHAR(50) - `email`, `sms`, or `follow_up`
- `subject`: VARCHAR(500) - For email actions
- `body`: TEXT - Message content
- `template_id`: UUID (optional reference to message_templates)
- `created_at`, `updated_at`: TIMESTAMP

### `drip_enrollments`
- `id`: UUID
- `campaign_id`: UUID (references drip_campaigns)
- `person_id`: UUID (references people)
- `status`: VARCHAR(50) - `active`, `completed`, `paused`, `cancelled`
- `enrolled_at`: TIMESTAMP
- `completed_at`: TIMESTAMP (nullable)

### `drip_step_executions`
- `id`: UUID
- `enrollment_id`: UUID (references drip_enrollments)
- `step_id`: UUID (references drip_steps)
- `status`: VARCHAR(50) - `pending`, `sent`, `failed`
- `scheduled_at`: TIMESTAMP - When the step should be executed
- `executed_at`: TIMESTAMP (nullable)
- `error_message`: TEXT (nullable)

## API Endpoints

### Campaigns

- `GET /api/drip/campaigns` - List all campaigns
- `POST /api/drip/campaigns` - Create new campaign
  ```json
  {
    "name": "New Visitor Welcome",
    "trigger_event": "first_visit",
    "is_active": true
  }
  ```
- `GET /api/drip/campaigns/:id` - Get campaign details with steps
- `PUT /api/drip/campaigns/:id` - Update campaign
- `DELETE /api/drip/campaigns/:id` - Delete campaign

### Steps

- `GET /api/drip/campaigns/:id/steps` - List campaign steps
- `POST /api/drip/campaigns/:id/steps` - Add step
  ```json
  {
    "step_order": 1,
    "delay_days": 0,
    "action_type": "email",
    "subject": "Welcome!",
    "body": "Hi {{first_name}}, welcome to our church!"
  }
  ```
- `PUT /api/drip/campaigns/:campaignId/steps/:stepId` - Update step
- `DELETE /api/drip/campaigns/:campaignId/steps/:stepId` - Delete step

### Enrollments

- `POST /api/drip/campaigns/:id/enroll/:personId` - Manually enroll someone
- `GET /api/drip/campaigns/:id/enrollments` - List enrollments

### Processing

- `POST /api/drip/process` - Process pending steps (called by cron job)

## How It Works

1. **Trigger**: When a qualifying event occurs (e.g., someone submits a connection card), they are enrolled in the matching campaign
2. **Scheduling**: All steps are immediately scheduled based on the enrollment date + delay_days
3. **Processing**: A cron job regularly checks for pending steps that are due and executes them
4. **Execution**: Depending on the action_type:
   - `email`: Sends email via configured SMTP/service
   - `sms`: Sends SMS via Twilio/service
   - `follow_up`: Creates a task/reminder for staff
5. **Completion**: Once all steps are executed, the enrollment is marked as completed

## Default Campaigns

Three pre-built templates are provided in `scripts/seed-default-campaigns.sql`:

1. **New Visitor Welcome** (trigger: `first_visit`)
   - Day 0: Welcome email
   - Day 3: Follow-up call reminder
   - Day 7: Small groups invitation
   - Day 14: Serving opportunities

2. **New Member Onboarding** (trigger: `new_member`)
   - Day 0: Welcome to the family
   - Day 7: New members class info
   - Day 30: Check-in

3. **Connection Card Follow-up** (trigger: `connection_card`)
   - Day 0: Thank you email
   - Day 1: Review and follow-up reminder
   - Day 7: Stay connected

## Frontend Routes

- `/dashboard/communication/drip` - Campaign list
- `/dashboard/communication/drip/:id` - Campaign editor with steps and enrollments
- `/dashboard/communication/drip/new` - Create new campaign

## Testing

1. Start the development environment:
   ```bash
   docker compose up -d
   ```

2. Access the frontend at `http://localhost:5173`

3. Navigate to Communication > Drip Campaigns

4. Create a campaign:
   - Set name and trigger event
   - Add steps with delays and messages
   - Activate the campaign

5. Enroll a test person manually

6. Verify steps are scheduled:
   ```sql
   SELECT * FROM drip_step_executions WHERE enrollment_id = 'YOUR_ENROLLMENT_ID';
   ```

7. Manually trigger processing:
   ```bash
   curl -X POST http://localhost:8190/api/drip/process \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

## Future Enhancements

- [ ] Automatic triggers when connection cards are submitted
- [ ] Automatic triggers on first check-in
- [ ] A/B testing for step content
- [ ] Analytics dashboard (open rates, conversion rates)
- [ ] Conditional branching (if opened, send X; if not, send Y)
- [ ] Integration with actual email/SMS services
- [ ] Template variable preview
- [ ] Duplicate campaign feature
- [ ] Export/import campaign templates
- [ ] Timezone-aware scheduling

## Notes

- Campaigns are tenant-isolated with Row Level Security
- All tables have proper indexes for performance
- The system uses optimistic scheduling (all steps scheduled upfront)
- Failed steps are logged with error messages for debugging
- Enrollments are unique per (campaign, person) to prevent duplicates
