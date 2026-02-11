package media

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type MediaFile struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	URL          string    `json:"url"`
	Folder       string    `json:"folder"`
	UploadedBy   *string   `json:"uploaded_by,omitempty"`
	Tags         TagList   `json:"tags"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TagList []string

func (t TagList) Value() (driver.Value, error) {
	if t == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(t)
}

func (t *TagList) Scan(value interface{}) error {
	if value == nil {
		*t = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		*t = []string{}
		return nil
	}

	var tags []string
	if err := json.Unmarshal(bytes, &tags); err != nil {
		*t = []string{}
		return err
	}
	*t = tags
	return nil
}

type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeDocument MediaType = "document"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeAll      MediaType = "all"
)

var allowedContentTypes = map[string]bool{
	// Images
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,

	// Documents
	"application/pdf": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,

	// Audio
	"audio/mpeg": true,
	"audio/mp3":  true,
	"audio/wav":  true,
}

func IsAllowedContentType(contentType string) bool {
	return allowedContentTypes[contentType]
}

func GetMediaType(contentType string) MediaType {
	switch {
	case contentType == "image/jpeg" || contentType == "image/jpg" || 
		contentType == "image/png" || contentType == "image/gif" || 
		contentType == "image/webp":
		return MediaTypeImage
	case contentType == "application/pdf" || 
		contentType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return MediaTypeDocument
	case contentType == "audio/mpeg" || contentType == "audio/mp3" || 
		contentType == "audio/wav":
		return MediaTypeAudio
	default:
		return MediaTypeAll
	}
}
