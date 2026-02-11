package backup

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	BackupDir       = "/backups"
	RetentionDays   = 30
	BackupExtension = ".sql.gz"
)

type Service struct {
	pool        *pgxpool.Pool
	backupDir   string
	databaseURL string
}

func NewService(pool *pgxpool.Pool, databaseURL string) *Service {
	return &Service{
		pool:        pool,
		backupDir:   BackupDir,
		databaseURL: databaseURL,
	}
}

// CreateBackup creates a tenant-scoped backup
func (s *Service) CreateBackup(ctx context.Context, tenantID, tenantSlug string) (*Backup, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s%s", tenantSlug, timestamp, BackupExtension)
	backupPath := filepath.Join(s.backupDir, filename)

	// Build SQL backup content
	var buffer bytes.Buffer
	
	// Write SQL header
	buffer.WriteString("-- Pews Backup\n")
	buffer.WriteString(fmt.Sprintf("-- Tenant: %s (ID: %s)\n", tenantSlug, tenantID))
	buffer.WriteString(fmt.Sprintf("-- Created: %s\n\n", time.Now().Format(time.RFC3339)))
	buffer.WriteString("BEGIN;\n\n")

	// Get list of tables that have tenant_id column
	tenantTables, err := s.getTenantTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant tables: %w", err)
	}

	// Export data for each tenant table
	for _, table := range tenantTables {
		buffer.WriteString(fmt.Sprintf("-- Table: %s\n", table))
		
		// Get column names
		columnQuery := fmt.Sprintf(`
			SELECT column_name 
			FROM information_schema.columns 
			WHERE table_name = '%s' AND table_schema = 'public'
			ORDER BY ordinal_position
		`, table)
		
		rows, err := s.pool.Query(ctx, columnQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", table, err)
		}
		
		var columns []string
		for rows.Next() {
			var col string
			if err := rows.Scan(&col); err != nil {
				rows.Close()
				return nil, fmt.Errorf("failed to scan column: %w", err)
			}
			columns = append(columns, col)
		}
		rows.Close()
		
		if len(columns) == 0 {
			continue
		}

		// Query tenant data
		dataQuery := fmt.Sprintf("SELECT * FROM %s WHERE tenant_id = $1", table)
		dataRows, err := s.pool.Query(ctx, dataQuery, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to query table %s: %w", table, err)
		}
		
		rowCount := 0
		for dataRows.Next() {
			values, err := dataRows.Values()
			if err != nil {
				dataRows.Close()
				return nil, fmt.Errorf("failed to get values: %w", err)
			}
			
			if rowCount == 0 {
				buffer.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES\n",
					table, strings.Join(columns, ", ")))
			} else {
				buffer.WriteString(",\n")
			}
			
			buffer.WriteString("  (")
			for i, val := range values {
				if i > 0 {
					buffer.WriteString(", ")
				}
				buffer.WriteString(formatSQLValue(val))
			}
			buffer.WriteString(")")
			rowCount++
		}
		dataRows.Close()
		
		if rowCount > 0 {
			buffer.WriteString(";\n\n")
		}
	}

	buffer.WriteString("COMMIT;\n")

	// Compress and write to file
	outFile, err := os.Create(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup file: %w", err)
	}
	defer outFile.Close()

	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	if _, err := gzipWriter.Write(buffer.Bytes()); err != nil {
		os.Remove(backupPath)
		return nil, fmt.Errorf("failed to write compressed backup: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		os.Remove(backupPath)
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	if err := outFile.Close(); err != nil {
		os.Remove(backupPath)
		return nil, fmt.Errorf("failed to close backup file: %w", err)
	}

	// Get file size
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat backup file: %w", err)
	}

	backup := &Backup{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		TenantSlug: tenantSlug,
		Filename:   filename,
		SizeBytes:  fileInfo.Size(),
		CreatedAt:  time.Now(),
	}

	return backup, nil
}

// ListBackups lists all backups for a specific tenant
func (s *Service) ListBackups(ctx context.Context, tenantSlug string) ([]Backup, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(s.backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []Backup
	prefix := tenantSlug + "_"

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backup := Backup{
			ID:         strings.TrimSuffix(entry.Name(), BackupExtension),
			TenantSlug: tenantSlug,
			Filename:   entry.Name(),
			SizeBytes:  info.Size(),
			CreatedAt:  info.ModTime(),
		}
		backups = append(backups, backup)
	}

	// Sort by creation time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// RestoreBackup restores a backup for a specific tenant
func (s *Service) RestoreBackup(ctx context.Context, tenantID, tenantSlug, filename string) error {
	// Security check: ensure filename belongs to tenant
	if !strings.HasPrefix(filename, tenantSlug+"_") {
		return fmt.Errorf("backup does not belong to tenant")
	}

	// Create a safety backup before restore
	safetySlug := tenantSlug + "_pre_restore"
	_, err := s.CreateBackup(ctx, tenantID, safetySlug)
	if err != nil {
		return fmt.Errorf("failed to create safety backup: %w", err)
	}

	backupPath := filepath.Join(s.backupDir, filename)
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup file not found: %w", err)
	}

	// Read and decompress backup file
	file, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	sqlData, err := io.ReadAll(gzipReader)
	if err != nil {
		return fmt.Errorf("failed to read backup data: %w", err)
	}

	// Start transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing tenant data
	tenantTables, err := s.getTenantTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant tables: %w", err)
	}

	for _, table := range tenantTables {
		query := fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1", table)
		_, err := tx.Exec(ctx, query, tenantID)
		if err != nil {
			return fmt.Errorf("failed to delete data from table %s: %w", table, err)
		}
	}

	// Execute backup SQL
	_, err = tx.Exec(ctx, string(sqlData))
	if err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit restore transaction: %w", err)
	}

	return nil
}

// DeleteBackup deletes a backup file
func (s *Service) DeleteBackup(ctx context.Context, tenantSlug, filename string) error {
	// Security check: ensure filename belongs to tenant
	if !strings.HasPrefix(filename, tenantSlug+"_") {
		return fmt.Errorf("backup does not belong to tenant")
	}

	backupPath := filepath.Join(s.backupDir, filename)
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	return nil
}

// DownloadBackup returns a reader for the backup file
func (s *Service) DownloadBackup(ctx context.Context, tenantSlug, filename string) (io.ReadCloser, error) {
	// Security check: ensure filename belongs to tenant
	if !strings.HasPrefix(filename, tenantSlug+"_") {
		return nil, fmt.Errorf("backup does not belong to tenant")
	}

	backupPath := filepath.Join(s.backupDir, filename)
	file, err := os.Open(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}

	return file, nil
}

// CleanupOldBackups removes backups older than the retention period
func (s *Service) CleanupOldBackups(ctx context.Context) error {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -RetentionDays)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip pre-restore safety backups (keep forever)
		if strings.Contains(entry.Name(), "_pre_restore_") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			backupPath := filepath.Join(s.backupDir, entry.Name())
			if err := os.Remove(backupPath); err != nil {
				fmt.Printf("Failed to delete old backup %s: %v\n", entry.Name(), err)
			}
		}
	}

	return nil
}

// getTenantTables returns a list of tables that have a tenant_id column
func (s *Service) getTenantTables(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT table_name 
		FROM information_schema.columns 
		WHERE column_name = 'tenant_id' 
		AND table_schema = 'public'
		ORDER BY table_name
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

// formatSQLValue formats a value for SQL insertion
func formatSQLValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}
	
	switch v := val.(type) {
	case string:
		// Escape single quotes
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case []byte:
		// Convert bytes to hex string
		return fmt.Sprintf("'\\x%x'", v)
	case time.Time:
		return fmt.Sprintf("'%s'", v.Format(time.RFC3339))
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	default:
		return fmt.Sprintf("%v", val)
	}
}
