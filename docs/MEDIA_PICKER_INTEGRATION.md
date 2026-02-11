# MediaPicker Integration Guide

## Overview
The MediaPicker component provides a reusable way to select files from the media library or upload new ones. Perfect for adding images to events, audio to sermons, or documents to announcements.

## Import

```svelte
<script>
  import MediaPicker from '$lib/components/MediaPicker.svelte';
</script>
```

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `type` | string | `'all'` | Filter files by type: `'all'`, `'image'`, `'document'`, `'audio'` |
| `value` | object \| null | `null` | Currently selected file (bind this) |
| `label` | string | `'Select Media'` | Button label text |
| `allowUpload` | boolean | `true` | Show upload button in picker |

## Events

| Event | Payload | Description |
|-------|---------|-------------|
| `select` | File object or null | Fires when a file is selected or cleared |

## File Object Structure

When a file is selected, the `value` will be a MediaFile object:

```typescript
{
  id: string,
  tenant_id: string,
  filename: string,          // UUID-based filename
  original_name: string,     // User's original filename
  content_type: string,      // MIME type
  size_bytes: number,
  url: string,              // e.g., "/uploads/{tenant_id}/{filename}"
  folder: string,
  uploaded_by: string | null,
  tags: string[],
  created_at: string,
  updated_at: string
}
```

## Usage Examples

### 1. Simple Image Picker

```svelte
<script>
  let eventImage = null;
</script>

<div class="form-group">
  <label>Event Image</label>
  <MediaPicker type="image" bind:value={eventImage} label="Choose Image" />
</div>

{#if eventImage}
  <p>Selected: {eventImage.original_name}</p>
  <img src={`${import.meta.env.VITE_API_URL}${eventImage.url}`} alt="Preview" />
{/if}
```

### 2. Audio File for Sermon

```svelte
<script>
  let sermonAudio = null;
  
  function saveSermon() {
    const sermon = {
      title: 'Sunday Sermon',
      speaker: 'Pastor John',
      audio_url: sermonAudio?.url || null
    };
    
    // Save sermon...
  }
</script>

<form on:submit|preventDefault={saveSermon}>
  <div class="form-group">
    <label for="title">Title</label>
    <input id="title" type="text" />
  </div>
  
  <div class="form-group">
    <label>Sermon Audio</label>
    <MediaPicker 
      type="audio" 
      bind:value={sermonAudio}
      label="Select Audio File"
    />
  </div>
  
  <button type="submit">Save Sermon</button>
</form>
```

### 3. Document Picker with Event Handler

```svelte
<script>
  let policyDocument = null;
  
  function handleFileSelect(event) {
    const file = event.detail;
    console.log('File selected:', file);
    
    if (file) {
      // Do something when file is selected
      alert(`Selected: ${file.original_name}`);
    }
  }
</script>

<MediaPicker 
  type="document" 
  bind:value={policyDocument}
  label="Select Policy PDF"
  on:select={handleFileSelect}
/>
```

### 4. Multiple Pickers in One Form

```svelte
<script>
  let eventData = {
    title: '',
    description: '',
    image: null,
    flyer: null,
    audioGuide: null
  };
  
  async function saveEvent() {
    const payload = {
      title: eventData.title,
      description: eventData.description,
      image_url: eventData.image?.url,
      flyer_url: eventData.flyer?.url,
      audio_guide_url: eventData.audioGuide?.url
    };
    
    await api('/api/events', {
      method: 'POST',
      body: JSON.stringify(payload)
    });
  }
</script>

<form on:submit|preventDefault={saveEvent}>
  <div class="form-group">
    <label>Event Image</label>
    <MediaPicker type="image" bind:value={eventData.image} />
  </div>
  
  <div class="form-group">
    <label>Event Flyer (PDF)</label>
    <MediaPicker type="document" bind:value={eventData.flyer} />
  </div>
  
  <div class="form-group">
    <label>Audio Guide</label>
    <MediaPicker type="audio" bind:value={eventData.audioGuide} />
  </div>
  
  <button type="submit">Create Event</button>
</form>
```

### 5. Read-Only Picker (No Upload)

```svelte
<MediaPicker 
  type="image" 
  bind:value={selectedImage}
  label="Choose from Library"
  allowUpload={false}
/>
```

### 6. Using the URL

Once a file is selected, use the `url` property to reference it:

```svelte
<script>
  let backgroundImage = null;
</script>

<MediaPicker type="image" bind:value={backgroundImage} />

{#if backgroundImage}
  <!-- Display image -->
  <img src={`${import.meta.env.VITE_API_URL}${backgroundImage.url}`} alt="Background" />
  
  <!-- Use in CSS -->
  <div style="background-image: url('{import.meta.env.VITE_API_URL}{backgroundImage.url}')">
    Content here
  </div>
  
  <!-- Link to download -->
  <a href={`${import.meta.env.VITE_API_URL}${backgroundImage.url}`} download>
    Download {backgroundImage.original_name}
  </a>
{/if}
```

### 7. Resetting/Clearing Selection

The clear button is built-in, but you can also programmatically clear:

```svelte
<script>
  let selectedFile = null;
  
  function clearFile() {
    selectedFile = null;
  }
</script>

<MediaPicker bind:value={selectedFile} />

{#if selectedFile}
  <button on:click={clearFile}>Clear Selection</button>
{/if}
```

## Integration with Existing Modules

### Adding to Services Module

**File:** `web/src/routes/dashboard/services/+page.svelte`

```svelte
<script>
  import MediaPicker from '$lib/components/MediaPicker.svelte';
  
  let newService = {
    // ... existing fields
    sermon_audio: null,
    sermon_notes_pdf: null
  };
</script>

<!-- In your create/edit modal: -->
<div class="form-group">
  <label>Sermon Audio</label>
  <MediaPicker 
    type="audio" 
    bind:value={newService.sermon_audio}
    label="Upload or Select Audio"
  />
</div>

<div class="form-group">
  <label>Sermon Notes (PDF)</label>
  <MediaPicker 
    type="document" 
    bind:value={newService.sermon_notes_pdf}
    label="Upload or Select PDF"
  />
</div>
```

When saving:
```javascript
const serviceData = {
  // ... other fields
  sermon_audio_url: newService.sermon_audio?.url || null,
  sermon_notes_url: newService.sermon_notes_pdf?.url || null
};
```

### Adding to Events/Groups Module

```svelte
<script>
  import MediaPicker from '$lib/components/MediaPicker.svelte';
  
  let group = {
    name: '',
    description: '',
    banner_image: null,
    meeting_location_map: null
  };
</script>

<div class="form-group">
  <label>Group Banner</label>
  <MediaPicker type="image" bind:value={group.banner_image} />
</div>

<div class="form-group">
  <label>Meeting Location Map</label>
  <MediaPicker type="image" bind:value={group.meeting_location_map} />
</div>
```

## Database Schema Updates

If integrating with existing modules, you may want to add URL columns:

```sql
-- Example: Add sermon audio to services table
ALTER TABLE services ADD COLUMN sermon_audio_url TEXT;
ALTER TABLE services ADD COLUMN sermon_notes_url TEXT;

-- Example: Add banner to groups table
ALTER TABLE groups ADD COLUMN banner_url TEXT;
```

Then reference the media files by their URL rather than storing file IDs.

## Best Practices

1. **Always use `${import.meta.env.VITE_API_URL}${file.url}` when displaying files**
2. **Store the URL in your entity, not the file ID** (simpler, more flexible)
3. **Use specific type filters** (`type="image"` instead of `type="all"`)
4. **Provide clear labels** that describe what type of file is expected
5. **Handle null values** - not all entities need media attachments

## Troubleshooting

### Image not showing
- Make sure you're prefixing with `${import.meta.env.VITE_API_URL}`
- Check browser console for 404 errors
- Verify the file exists in `/uploads/{tenant_id}/`

### Upload failing
- Check file size (max 50MB)
- Verify file type is allowed
- Check browser console for error messages

### Files not appearing in picker
- Ensure you're authenticated
- Check the type filter matches uploaded files
- Try filtering by the correct folder

## See Also

- [Media Library User Guide](./MEDIA_LIBRARY.md)
- [API Documentation](./API.md#media-endpoints)
