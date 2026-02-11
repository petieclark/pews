# SMS Messaging Feature

## Overview
The SMS messaging feature allows churches to send text messages to their members using their own Twilio account.

## Architecture

### Backend (`internal/sms/`)
- **model.go** - Data models (Message, Template, Settings)
- **twilio.go** - Twilio API client with rate limiting (1 msg/sec)
- **service.go** - Business logic, encryption, database operations
- **handler.go** - HTTP handlers for REST API

### Database Migration
- **014_sms_messaging.sql** - Creates tables:
  - `sms_messages` - Message history with delivery status
  - `sms_templates` - Reusable message templates
  - `tenant_sms_settings` - Encrypted Twilio credentials per tenant

### Frontend (`web/src/routes/dashboard/communication/sms/`)
- **+page.svelte** - Compose and send SMS (single or bulk)
- **templates/+page.svelte** - Manage message templates
- **settings/+page.svelte** - Configure Twilio credentials

## API Endpoints

### Sending
- `POST /api/sms/send` - Send SMS to one person
- `POST /api/sms/bulk` - Send SMS to multiple recipients

### History
- `GET /api/sms/history` - Retrieve message history

### Templates
- `GET /api/sms/templates` - List templates
- `POST /api/sms/templates` - Create template
- `PUT /api/sms/templates/:id` - Update template
- `DELETE /api/sms/templates/:id` - Delete template

### Settings
- `GET /api/sms/settings` - Get Twilio configuration
- `POST /api/sms/settings` - Save Twilio credentials
- `POST /api/sms/settings/test` - Test Twilio connection

## Environment Variables

Add to `.env`:
```bash
SMS_ENCRYPTION_KEY="your-32-character-encryption-key-here"
```

## Twilio Setup (for Churches)

1. Sign up at https://www.twilio.com
2. Get Account SID and Auth Token from Console
3. Purchase a phone number (or use trial number)
4. Configure in Pews Settings → SMS → Settings

## Security

- Twilio Auth Tokens are encrypted using AES-256-GCM
- Per-tenant credentials stored separately
- Rate limiting prevents abuse (1 msg/sec)
- Phone numbers validated before sending

## Features

### Single Message
- Send to one phone number
- Use templates with merge fields
- Character count with SMS segment calculation

### Bulk Message
- Send to multiple recipients:
  - **All people** - Everyone with a phone number
  - **Group** - All members of a selected group
  - **Person IDs** - Specific individuals
- Merge fields auto-populate: `{first_name}`, `{last_name}`

### Templates
- Create reusable message templates
- Support for merge fields: `{first_name}`, `{last_name}`, `{church_name}`
- Easy template selection when composing

### Message History
- View all sent messages
- Delivery status tracking:
  - Queued
  - Sent
  - Delivered
  - Failed
- Error messages for failed sends

## Testing

### 1. Start the backend
```bash
cd ~/Projects/pews
docker compose up -d
```

### 2. Configure Twilio credentials
- Navigate to **Dashboard → Communication → SMS → Settings**
- Enter your Twilio Account SID, Auth Token, and From Number
- Click **Test Connection** to verify

### 3. Send a test message
- Go to **Dashboard → Communication → SMS**
- Enter a phone number (your own for testing)
- Type a message
- Click **Send SMS**

### 4. Create a template
- Go to **Templates**
- Create a template with merge fields
- Test it by selecting when composing a message

### 5. Bulk send (optional)
- Switch to **Bulk Message** mode
- Select "All People" or a specific group
- Send to multiple recipients

## Rate Limiting
- **1 message per second** to comply with Twilio best practices
- Bulk sends process sequentially with rate limiting
- Large groups will take time proportional to recipient count

## SMS Segments
- Standard SMS: **160 characters per segment**
- Unicode (emojis, special chars): **70 characters per segment**
- Twilio charges per segment
- Character counter shows segment count

## Merge Fields
Supported merge fields:
- `{first_name}` - Person's first name
- `{last_name}` - Person's last name
- `{church_name}` - Tenant's church name (if configured)
- `{email}` - Person's email address

## Cost Considerations
- Churches bring their own Twilio account
- Twilio pricing: ~$0.0075 per SMS segment (US)
- Monitor usage in Twilio Console
- Set up usage alerts in Twilio to prevent overages

## Troubleshooting

### "SMS is not enabled for this tenant"
→ Configure Twilio credentials in Settings first

### "Failed to decrypt auth token"
→ SMS_ENCRYPTION_KEY changed or corrupted. Re-save credentials

### "Connection test failed"
→ Verify Account SID and Auth Token are correct

### "Invalid phone number"
→ Phone numbers must be in E.164 format: +1234567890

### Messages stuck in "queued" status
→ Check Twilio Console for delivery logs and errors

## Future Enhancements
- Webhook to receive delivery status updates from Twilio
- Scheduled bulk sends
- SMS campaigns with tracking
- Two-way messaging (receive replies)
- Template variables from person custom fields
- Export message history to CSV
