# Pews Issue #70 — Song File Attachments Implementation Complete ✅

## Summary
Song attachment API endpoints are fully implemented and ready for frontend integration.

## Database Schema (Already Exists)
Migration `033_song_attachments.sql` creates the table:
```sql
CREATE TABLE IF NOT EXISTS song_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    song_id UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL DEFAULT 'application/pdf',
    file_data BYTEA NOT NULL,
    file_size INTEGER NOT NULL,
    uploaded_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_song_attachments_song ON song_attachments(song_id);
CREATE INDEX idx_song_attachments_tenant ON song_attachments(tenant_id);
```

## Files Modified/Created

### 1. `/internal/services/attachments.go` (UPDATED)
**Changes:**
- Updated `maxUploadSize` from 10MB to **20MB** (`MaxSongAttachmentSize`)
- Expanded allowed file types: **PDF, PNG, JPG, DOCX** (was PDF only)
- Improved error messages with clear file type requirements
- All handler methods already existed, now compliant with requirements

### 2. `/internal/services/service.go`
**Status:** Backend service methods already implemented:
- `CreateSongAttachment()` - inserts attachment record into DB
- `ListSongAttachments()` - returns list of attachments (no file data)
- `GetSongAttachment()` - returns single attachment WITH file data for download
- `DeleteSongAttachment()` - removes from DB

## API Endpoints Implemented

### POST `/api/services/songs/:id/attachments`
Upload file to song.

**Request:**
```
POST /api/services/songs/{songId}/attachments
Authorization: Bearer {token}
Content-Type: multipart/form-data

file: <binary>
```

**Constraints:**
- Max size: 20MB
- Allowed types: PDF, PNG, JPG, DOCX

**Response (201 Created):**
```json
{
    "id": "uuid",
    "tenant_id": "uuid",
    "song_id": "uuid",
    "filename": "hymn_sheet_music.pdf",
    "original_name": "Sunday Worship - Hymn 42.pdf",
    "content_type": "application/pdf",
    "file_size": 245678,
    "uploaded_by": "user-uuid",
    "created_at": "2026-02-27T19:30:00Z"
}
```

### GET `/api/services/songs/:id/attachments`
List attachments for a song (metadata only, no file data).

**Response (200 OK):**
```json
[
    {
        "id": "uuid",
        "song_id": "uuid",
        "filename": "hymn_sheet_music.pdf",
        "original_name": "Sunday Worship - Hymn 42.pdf",
        "content_type": "application/pdf",
        "file_size": 245678,
        "uploaded_by": "user-uuid",
        "created_at": "2026-02-27T19:30:00Z"
    }
]
```

### GET `/api/services/songs/attachments/:attachmentId`
Download attachment (returns raw file data).

**Response (200 OK):**
- Content-Type: `application/pdf` (or image/jpeg, etc.)
- Content-Disposition: inline; filename="hymn_sheet_music.pdf"
- Body: Raw binary file data for download

### DELETE `/api/services/songs/attachments/:attachmentId`
Delete attachment.

**Response (204 No Content)**

## Storage Pattern
Files are stored in PostgreSQL `file_data` BYTEA column. This is consistent with existing song_attachments migration pattern. File data is NOT served from filesystem.

## Security & Validation
- ✅ Authentication: All endpoints require JWT token via middleware
- ✅ Tenant isolation: Queries filter by tenant_id
- ✅ Song verification: Upload endpoint verifies song exists and belongs to tenant
- ✅ File type validation: Server-side check of content-type header
- ✅ Size limit: 20MB enforced before database insert
- ✅ Delete cascade: Removing a song automatically removes all attachments (FK constraint)

## Frontend Integration Notes

### Required Components (NOT YET BUILT - Separate task):

1. **Song Detail Page Upload Zone**
   - Drag-and-drop file upload area
   - File type validation UI feedback (show allowed types: PDF, PNG, JPG, DOCX)
   - Size limit indicator (max 20MB)
   - Progress indicator during upload

2. **Attachments List Display**
   - Table/list of uploaded files per song
   - Show filename, size (formatted), download button
   - Delete button with confirmation dialog
   - File type icons based on content_type

3. **Tokenized Plan View Integration (#69)**
   - On `/plan/:token` public view, show downloadable attachments for songs
   - Use GET endpoint directly without auth for public access (needs route config)
   - Consider signed URLs if files are sensitive

## Testing Commands

```bash
# Upload a PDF file
curl -X POST "http://localhost:8080/api/services/songs/{songId}/attachments" \
  -H "Authorization: Bearer {token}" \
  -F "file=@/path/to/chord-chart.pdf"

# List attachments for a song
curl -X GET "http://localhost:8080/api/services/songs/{songId}/attachments" \
  -H "Authorization: Bearer {token}"

# Download an attachment (returns raw file)
curl -X GET "http://localhost:8080/api/services/songs/attachments/{attachmentId}" \
  -H "Authorization: Bearer {token}" \
  -o downloaded-file.pdf

# Delete an attachment
curl -X DELETE "http://localhost:8080/api/services/songs/attachments/{attachmentId}" \
  -H "Authorization: Bearer {token}"
```

## Verification Checklist
- [x] Database migration exists (033_song_attachments.sql)
- [x] SongAttachment model defined in services/model.go
- [x] Backend service methods implemented (service.go)
- [x] Handler methods implemented and compliant with requirements (attachments.go)
  - [x] Max file size: 20MB ✅
  - [x] Allowed types: PDF, PNG, JPG, DOCX ✅
- [x] API routes defined in router (router.go lines 270-273)
- [x] Code compiles successfully (`go build ./internal/services/...`)

## Next Steps (Frontend Work)
1. Build upload UI component for song detail page
2. Implement file list display with download/delete controls
3. Add public access to attachment downloads on tokenized plan view (#69)
4. Test with actual PDF, PNG, JPG, DOCX files
5. Verify file size validation at client and server level

## Closing Issue
Once frontend integration is complete, close issue:
```bash
gh issue close 70 --repo warpapaya/Pews -c "Backend API fully implemented. Ready for frontend integration."
```
