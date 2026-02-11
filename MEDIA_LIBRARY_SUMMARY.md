# Pews Media Library - Implementation Summary

## ✅ Completed Features

### 1. Database Migration (`013_media_files.sql`)
- `media_files` table with:
  - id, tenant_id, filename, original_name
  - content_type, size_bytes, url
  - folder, uploaded_by, tags (JSONB)
  - created_at, updated_at timestamps
- Indexes on:
  - tenant_id, folder, content_type
  - tags (GIN index for JSONB)
  - created_at (DESC for sorting)
- Automatic updated_at trigger

### 2. Backend Package (`internal/media/`)

**Model (model.go):**
- MediaFile struct with all fields
- TagList custom type for JSONB handling
- MediaType enum (image, document, audio, all)
- Content type validation:
  - Images: jpg, png, gif, webp
  - Documents: pdf, docx
  - Audio: mp3, wav
- Max file size: 50MB

**Service (service.go):**
- `UploadFile()` - Upload with validation, disk storage, DB record
- `ListFiles()` - Filter by type, folder, tags
- `GetFile()` - Get single file by ID
- `UpdateFile()` - Update folder and tags
- `DeleteFile()` - Remove from disk and DB
- `ListFolders()` - Get distinct folders
- File organization:
  - Stored in `/uploads/{tenant_id}/{uuid}.ext`
  - URL format: `/uploads/{tenant_id}/{filename}`

**Handler (handler.go):**
- `POST /api/media/upload` - Multipart form upload
- `GET /api/media` - List files (query params: type, folder, tags)
- `GET /api/media/folders` - List folders
- `GET /api/media/{id}` - Get file metadata
- `PUT /api/media/{id}` - Update folder/tags
- `DELETE /api/media/{id}` - Delete file

### 3. Router Integration (`internal/router/router.go`)
- Added media routes to protected group (auth required)
- Static file server at `/uploads/*` for file serving

### 4. Configuration Updates
- Added `UPLOADS_PATH` env var (default: `./uploads`)
- Docker compose volume `uploads:/uploads`
- Config struct includes UploadsPath field

### 5. Frontend - Media Library Page (`/dashboard/media/`)

**Features:**
- Grid view with file cards showing:
  - Image thumbnails (actual image preview)
  - Icons for documents/audio
  - File name, size, date
  - Folder badge
  - Tags
- Filter by:
  - Media type (all, images, documents, audio)
  - Folder
  - Search files
- Upload modal:
  - File selection or drag-and-drop
  - Folder assignment (create or select)
  - Tag input (comma-separated)
  - Multi-file upload support
- File management:
  - Bulk selection with checkboxes
  - Bulk delete
  - Individual delete
  - Download
  - Preview modal (images, PDFs, audio player)
- Folder creation:
  - Create new folders on-the-fly
  - Folder dropdown filter

**UI/UX:**
- Drag-and-drop zone with visual feedback
- Empty state with upload prompt
- Responsive grid layout
- Modal dialogs for upload/preview
- File type icons
- Formatted file sizes
- Date formatting

### 6. Reusable Component - MediaPicker

**Location:** `web/src/lib/components/MediaPicker.svelte`

**Props:**
- `type` - Filter by media type ('image', 'document', 'audio', 'all')
- `value` - Currently selected file object
- `label` - Button text (default: "Select Media")
- `allowUpload` - Enable upload from picker (default: true)

**Events:**
- `on:select` - Fires when file is selected/cleared

**Features:**
- Compact file selector button
- Shows selected file with preview
- Clear/remove selection
- Modal file picker with:
  - Search files
  - Filter by folder
  - Upload new files
- Auto-selects uploaded file

**Usage Example:**
```svelte
<script>
  import MediaPicker from '$lib/components/MediaPicker.svelte';
  let selectedImage = null;
</script>

<MediaPicker 
  type="image" 
  bind:value={selectedImage}
  label="Select Sermon Image"
  on:select={(e) => console.log('Selected:', e.detail)}
/>
```

## 🔧 Integration Points

**For other modules (e.g., Services, Streaming):**

Add to your form:
```svelte
<script>
  import MediaPicker from '$lib/components/MediaPicker.svelte';
  
  let service = {
    // ... other fields
    imageFile: null,
    audioFile: null
  };
</script>

<form>
  <!-- ... other fields ... -->
  
  <div class="form-group">
    <label>Sermon Image</label>
    <MediaPicker type="image" bind:value={service.imageFile} />
  </div>
  
  <div class="form-group">
    <label>Sermon Audio</label>
    <MediaPicker type="audio" bind:value={service.audioFile} />
  </div>
</form>
```

Then use `service.imageFile.url` and `service.audioFile.url` when saving.

## 🧪 Testing Steps

1. **Start services:**
   ```bash
   cd ~/Projects/pews
   docker compose up -d
   ```

2. **Access media library:**
   - Navigate to `http://localhost:5273`
   - Login to dashboard
   - Click "Media" in navigation

3. **Test upload:**
   - Click "Upload Files"
   - Select multiple images/documents/audio
   - Assign to folder (e.g., "Sermons")
   - Add tags (e.g., "worship, music")
   - Upload
   - Verify files appear in grid

4. **Test filtering:**
   - Change media type dropdown
   - Select different folders
   - Use select all / bulk delete

5. **Test preview:**
   - Click on an image → see full preview
   - Click on a PDF → see PDF viewer
   - Click on audio → see audio player

6. **Test MediaPicker component:**
   - Go to another module that uses it
   - Click "Select Media"
   - Search/filter files
   - Select a file
   - Verify it shows in the form

7. **Test file serving:**
   - Right-click a file → "Open in new tab"
   - Verify file loads correctly at `/uploads/{tenant_id}/{filename}`

## 📝 API Examples

**Upload a file:**
```bash
curl -X POST http://localhost:8190/api/media/upload \
  -H "Authorization: Bearer {token}" \
  -F "file=@/path/to/image.jpg" \
  -F "folder=Sermons" \
  -F "tags=worship,sunday"
```

**List all images:**
```bash
curl http://localhost:8190/api/media?type=image \
  -H "Authorization: Bearer {token}"
```

**Get file by ID:**
```bash
curl http://localhost:8190/api/media/{file-id} \
  -H "Authorization: Bearer {token}"
```

**Delete file:**
```bash
curl -X DELETE http://localhost:8190/api/media/{file-id} \
  -H "Authorization: Bearer {token}"
```

## 🚀 Deployment Notes

**Environment Variables:**
- `UPLOADS_PATH=/uploads` (set in docker-compose.yml)

**Docker Volume:**
- `uploads:/uploads` - Persists uploaded files across container restarts

**File Storage:**
- Files stored in tenant-specific folders: `/uploads/{tenant_id}/`
- Unique filenames prevent collisions (UUID + extension)
- Database tracks metadata, filesystem stores actual files

**Cleanup:**
- When deleting a file, both DB record and disk file are removed
- If DB delete fails, disk file is left intact (manual cleanup may be needed)

## ✅ Branch Status

All changes committed to `feat/media-library` branch.

**DO NOT MERGE TO MAIN** (as requested).

## 🔮 Future Enhancements (Not Implemented)

- **Image processing:**
  - Auto-generate thumbnails
  - Image resizing/compression
  - Multiple resolution variants
  
- **Advanced search:**
  - Full-text search in filenames
  - Filter by date range
  - Filter by uploader
  
- **Permissions:**
  - Role-based access (admin vs staff vs volunteer)
  - Private vs public files
  
- **Cloud storage:**
  - S3/GCS integration
  - CDN support
  
- **Analytics:**
  - Track downloads
  - View counts
  - Storage usage reports
  
- **Bulk operations:**
  - Move files between folders
  - Copy files
  - Bulk tagging
  
- **File versioning:**
  - Keep history of uploads
  - Restore previous versions
