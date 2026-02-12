package media

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const MaxFileSize = 50 * 1024 * 1024 // 50MB

type Service struct {
	db          *pgxpool.Pool
	uploadsPath string
}

func NewService(db *pgxpool.Pool, uploadsPath string) *Service {
	// Ensure uploads directory exists
	os.MkdirAll(uploadsPath, 0755)
	return &Service{
		db:          db,
		uploadsPath: uploadsPath,
	}
}

func (s *Service) UploadFile(ctx context.Context, tenantID, userID string, file multipart.File, header *multipart.FileHeader, folder string, tags []string) (*MediaFile, error) {
	// Validate file size
	if header.Size > MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum of 50MB")
	}

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if !IsAllowedContentType(contentType) {
		return nil, fmt.Errorf("file type not allowed: %s", contentType)
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	
	// Create tenant folder if it doesn't exist
	tenantPath := filepath.Join(s.uploadsPath, tenantID)
	os.MkdirAll(tenantPath, 0755)

	// Save file to disk
	filePath := filepath.Join(tenantPath, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create database record
	url := fmt.Sprintf("/uploads/%s/%s", tenantID, filename)
	
	if tags == nil {
		tags = []string{}
	}
	tagsJSON, _ := json.Marshal(tags)

	var mediaFile MediaFile
	query := `
		INSERT INTO media_files (tenant_id, filename, original_name, content_type, size_bytes, url, folder, uploaded_by, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, tenant_id, filename, original_name, content_type, size_bytes, url, folder, uploaded_by, tags, created_at, updated_at
	`
	err = s.db.QueryRow(ctx, query, tenantID, filename, header.Filename, contentType, header.Size, url, folder, userID, tagsJSON).Scan(
		&mediaFile.ID,
		&mediaFile.TenantID,
		&mediaFile.Filename,
		&mediaFile.OriginalName,
		&mediaFile.ContentType,
		&mediaFile.SizeBytes,
		&mediaFile.URL,
		&mediaFile.Folder,
		&mediaFile.UploadedBy,
		&mediaFile.Tags,
		&mediaFile.CreatedAt,
		&mediaFile.UpdatedAt,
	)
	if err != nil {
		// Clean up file if database insert fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	return &mediaFile, nil
}

func (s *Service) ListFiles(ctx context.Context, tenantID string, mediaType MediaType, folder *string, tags []string) ([]MediaFile, error) {
	query := `
		SELECT id, tenant_id, filename, original_name, content_type, size_bytes, url, folder, uploaded_by, tags, created_at, updated_at
		FROM media_files
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argCount := 1

	// Filter by media type (content type pattern)
	if mediaType != MediaTypeAll {
		argCount++
		switch mediaType {
		case MediaTypeImage:
			query += fmt.Sprintf(" AND content_type LIKE $%d", argCount)
			args = append(args, "image/%")
		case MediaTypeDocument:
			query += fmt.Sprintf(" AND content_type IN ($%d, $%d)", argCount, argCount+1)
			args = append(args, "application/pdf", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
			argCount++
		case MediaTypeAudio:
			query += fmt.Sprintf(" AND content_type LIKE $%d", argCount)
			args = append(args, "audio/%")
		}
	}

	// Filter by folder
	if folder != nil {
		argCount++
		query += fmt.Sprintf(" AND folder = $%d", argCount)
		args = append(args, *folder)
	}

	// Filter by tags (at least one tag matches)
	if len(tags) > 0 {
		argCount++
		tagsJSON, _ := json.Marshal(tags)
		query += fmt.Sprintf(" AND tags ?| $%d", argCount)
		args = append(args, tagsJSON)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := []MediaFile{}
	for rows.Next() {
		var file MediaFile
		err := rows.Scan(
			&file.ID,
			&file.TenantID,
			&file.Filename,
			&file.OriginalName,
			&file.ContentType,
			&file.SizeBytes,
			&file.URL,
			&file.Folder,
			&file.UploadedBy,
			&file.Tags,
			&file.CreatedAt,
			&file.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

func (s *Service) GetFile(ctx context.Context, tenantID, fileID string) (*MediaFile, error) {
	var file MediaFile
	query := `
		SELECT id, tenant_id, filename, original_name, content_type, size_bytes, url, folder, uploaded_by, tags, created_at, updated_at
		FROM media_files
		WHERE id = $1 AND tenant_id = $2
	`
	err := s.db.QueryRow(ctx, query, fileID, tenantID).Scan(
		&file.ID,
		&file.TenantID,
		&file.Filename,
		&file.OriginalName,
		&file.ContentType,
		&file.SizeBytes,
		&file.URL,
		&file.Folder,
		&file.UploadedBy,
		&file.Tags,
		&file.CreatedAt,
		&file.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	return &file, nil
}

func (s *Service) UpdateFile(ctx context.Context, tenantID, fileID string, folder *string, tags []string) (*MediaFile, error) {
	// Build update query dynamically
	updates := []string{}
	args := []interface{}{fileID, tenantID}
	argCount := 2

	if folder != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("folder = $%d", argCount))
		args = append(args, *folder)
	}

	if tags != nil {
		argCount++
		tagsJSON, _ := json.Marshal(tags)
		updates = append(updates, fmt.Sprintf("tags = $%d", argCount))
		args = append(args, tagsJSON)
	}

	if len(updates) == 0 {
		return s.GetFile(ctx, tenantID, fileID)
	}

	updates = append(updates, fmt.Sprintf("updated_at = $%d", argCount+1))
	args = append(args, time.Now())

	query := fmt.Sprintf(`
		UPDATE media_files
		SET %s
		WHERE id = $1 AND tenant_id = $2
		RETURNING id, tenant_id, filename, original_name, content_type, size_bytes, url, folder, uploaded_by, tags, created_at, updated_at
	`, string(updates[0]))

	for i := 1; i < len(updates); i++ {
		query = fmt.Sprintf("%s, %s", query, updates[i])
	}

	var file MediaFile
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&file.ID,
		&file.TenantID,
		&file.Filename,
		&file.OriginalName,
		&file.ContentType,
		&file.SizeBytes,
		&file.URL,
		&file.Folder,
		&file.UploadedBy,
		&file.Tags,
		&file.CreatedAt,
		&file.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &file, nil
}

func (s *Service) DeleteFile(ctx context.Context, tenantID, fileID string) error {
	// Get file info first to delete from disk
	file, err := s.GetFile(ctx, tenantID, fileID)
	if err != nil {
		return err
	}

	// Delete from database
	_, err = s.db.Exec(ctx, "DELETE FROM media_files WHERE id = $1 AND tenant_id = $2", fileID, tenantID)
	if err != nil {
		return err
	}

	// Delete from disk
	filePath := filepath.Join(s.uploadsPath, tenantID, file.Filename)
	os.Remove(filePath) // Ignore error - file might already be deleted

	return nil
}

func (s *Service) ListFolders(ctx context.Context, tenantID string) ([]string, error) {
	query := `
		SELECT DISTINCT folder
		FROM media_files
		WHERE tenant_id = $1 AND folder != ''
		ORDER BY folder
	`
	rows, err := s.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	folders := []string{}
	for rows.Next() {
		var folder string
		if err := rows.Scan(&folder); err != nil {
			return nil, err
		}
		folders = append(folders, folder)
	}

	return folders, nil
}
