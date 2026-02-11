# Import API Documentation

Phase 1 of PCO Import (Issue #8) - Bulk data import endpoints.

## Overview

The Import API provides endpoints for bulk importing data from JSON or CSV sources. All endpoints support:
- **JSON format** (`Content-Type: application/json`)
- **CSV format** (`Content-Type: text/csv`)
- **Dry run mode** (preview without writing) via `?dry_run=true` query parameter
- **Authentication required** (admin users only)

## Endpoints

### 1. Import People

**Endpoint:** `POST /api/import/people`

**JSON Example:**
```json
{
  "people": [
    {
      "first_name": "John",
      "last_name": "Doe",
      "email": "john@example.com",
      "phone": "555-1234",
      "address_line1": "123 Main St",
      "city": "Springfield",
      "state": "IL",
      "zip": "62701",
      "gender": "male",
      "membership_status": "active"
    }
  ],
  "dry_run": false
}
```

**CSV Headers:**
```
first_name,last_name,email,phone,address_line1,city,state,zip,gender,membership_status
```

**Features:**
- Duplicate detection by email address
- Auto-generates missing IDs
- Sets default `membership_status` to "active" if not provided

---

### 2. Import Groups

**Endpoint:** `POST /api/import/groups`

**JSON Example:**
```json
{
  "groups": [
    {
      "name": "Monday Bible Study",
      "description": "Weekly Bible study",
      "type": "small_group",
      "meeting_day": "Monday",
      "meeting_time": "19:00",
      "meeting_location": "Room 101",
      "is_public": true,
      "max_members": 20,
      "members": ["john@example.com", "jane@example.com"]
    }
  ],
  "dry_run": false
}
```

**CSV Headers:**
```
name,description,type,meeting_day,meeting_time,meeting_location,is_public,max_members,members
```

**Notes:**
- `members` field is comma-separated emails in CSV: `"john@example.com,jane@example.com"`
- Members are linked by email (person must exist first)
- Default `type` is "small_group" if not provided

---

### 3. Import Songs

**Endpoint:** `POST /api/import/songs`

**JSON Example:**
```json
{
  "songs": [
    {
      "title": "Amazing Grace",
      "artist": "John Newton",
      "key": "G",
      "tempo": 75,
      "ccli_number": "4669344",
      "lyrics": "Amazing grace how sweet the sound...",
      "tags": "hymn,classic"
    }
  ],
  "dry_run": false
}
```

**CSV Headers:**
```
title,artist,key,tempo,ccli_number,lyrics,tags
```

**Features:**
- Duplicate detection by title + artist
- CCLI number for copyright tracking
- Tags for organization (comma-separated in CSV)

---

### 4. Import Giving (Donations)

**Endpoint:** `POST /api/import/giving`

**JSON Example:**
```json
{
  "donations": [
    {
      "donor_email": "john@example.com",
      "fund_name": "General Fund",
      "amount_cents": 10000,
      "currency": "USD",
      "payment_method": "check",
      "memo": "Monthly tithe",
      "donated_at": "2026-01-15T10:00:00Z"
    }
  ],
  "dry_run": false
}
```

**CSV Headers:**
```
donor_email,fund_name,amount_cents,currency,payment_method,memo,donated_at
```

**Features:**
- Matches donors by email (person must exist)
- Creates fund if it doesn't exist
- Default `currency` is "USD"
- `donated_at` is ISO 8601 format (defaults to current time if not provided)
- Amount is in cents (10000 = $100.00)

---

## Response Format

All endpoints return:

```json
{
  "created": 5,
  "skipped": 2,
  "errors": [
    "Row 3: missing required field (email)",
    "Row 7: duplicate email found"
  ]
}
```

**Fields:**
- `created` - Number of records successfully created
- `skipped` - Number of records skipped (duplicates)
- `errors` - Array of error messages (empty if no errors)

---

## Testing

### Using cURL (JSON)

```bash
# Get auth token
TOKEN=$(curl -s -X POST http://localhost:8190/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Password123!","tenant_name":"My Church"}' \
  | jq -r '.token')

# Import people (dry run)
curl -X POST "http://localhost:8190/api/import/people?dry_run=true" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d @test_import_people.json

# Import people (actual)
curl -X POST "http://localhost:8190/api/import/people" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d @test_import_people.json
```

### Using cURL (CSV)

```bash
curl -X POST "http://localhost:8190/api/import/people" \
  -H "Content-Type: text/csv" \
  -H "Authorization: Bearer $TOKEN" \
  --data-binary @test_import_people.csv
```

---

## Implementation Details

### Package Structure
```
internal/import/
├── handler.go    # HTTP handlers
├── service.go    # Business logic
├── model.go      # Request/response types
└── csv.go        # CSV parsing utilities
```

### CSV Header Normalization

The CSV parser normalizes headers by:
- Converting to lowercase
- Replacing spaces and hyphens with underscores
- Trimming whitespace

Examples:
- `"First Name"` → `"first_name"`
- `"Email Address"` → `"email_address"`
- `"CCLI-Number"` → `"ccli_number"`

### Error Handling

- Invalid CSV format returns HTTP 400
- Missing required fields are reported in errors array
- Database errors are returned with row numbers for debugging
- Authentication failures return HTTP 401

---

## Next Steps (Phase 2)

- PCO API integration (pull data from Planning Center Online)
- Scheduled imports
- Import history/logging
- Rollback functionality
- Mapping UI for custom CSV headers
