#!/usr/bin/env bash
# Pews Database Backup Script
# Creates timestamped PostgreSQL dumps and maintains retention policy

set -euo pipefail

# ============================================
# Configuration
# ============================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Load environment variables if .env.production exists
if [[ -f "$PROJECT_ROOT/.env.production" ]]; then
    set -a
    source "$PROJECT_ROOT/.env.production"
    set +a
fi

# Default values
BACKUP_DIR="${BACKUP_DIR:-$PROJECT_ROOT/backups}"
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-7}"
COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-pews}"
DB_CONTAINER="${COMPOSE_PROJECT_NAME}-postgres-1"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="pews_backup_${TIMESTAMP}.sql.gz"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ============================================
# Functions
# ============================================
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_container() {
    if ! docker ps --format '{{.Names}}' | grep -q "^${DB_CONTAINER}$"; then
        log_error "PostgreSQL container '$DB_CONTAINER' is not running"
        log_info "Available containers:"
        docker ps --format 'table {{.Names}}\t{{.Status}}'
        exit 1
    fi
}

create_backup_dir() {
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log_info "Creating backup directory: $BACKUP_DIR"
        mkdir -p "$BACKUP_DIR"
    fi
}

perform_backup() {
    log_info "Starting database backup..."
    log_info "Container: $DB_CONTAINER"
    log_info "Backup file: $BACKUP_FILE"
    
    # Perform the backup using pg_dumpall for complete backup
    if docker exec "$DB_CONTAINER" pg_dump -U pews pews | gzip > "$BACKUP_DIR/$BACKUP_FILE"; then
        BACKUP_SIZE=$(du -h "$BACKUP_DIR/$BACKUP_FILE" | cut -f1)
        log_info "Backup completed successfully: $BACKUP_SIZE"
        log_info "Location: $BACKUP_DIR/$BACKUP_FILE"
        return 0
    else
        log_error "Backup failed!"
        return 1
    fi
}

cleanup_old_backups() {
    log_info "Cleaning up backups older than $BACKUP_RETENTION_DAYS days..."
    
    # Find and delete old backups
    DELETED_COUNT=0
    while IFS= read -r -d '' backup; do
        log_info "Deleting old backup: $(basename "$backup")"
        rm -f "$backup"
        ((DELETED_COUNT++))
    done < <(find "$BACKUP_DIR" -name "pews_backup_*.sql.gz" -type f -mtime "+$BACKUP_RETENTION_DAYS" -print0)
    
    if [[ $DELETED_COUNT -gt 0 ]]; then
        log_info "Deleted $DELETED_COUNT old backup(s)"
    else
        log_info "No old backups to delete"
    fi
}

list_backups() {
    log_info "Available backups in $BACKUP_DIR:"
    if [[ -d "$BACKUP_DIR" ]]; then
        find "$BACKUP_DIR" -name "pews_backup_*.sql.gz" -type f -printf "%T@ %Tc %s %p\n" | sort -rn | awk '{
            # Convert bytes to human readable
            size = $10
            for (i=11; i<=NF; i++) size = size " " $i
            
            # Extract just filename
            n = split($11, arr, "/")
            filename = arr[n]
            
            # Format: Date | Size | Filename
            printf "  %s %s %s | %7.2f MB | %s\n", $2, $3, $4, $10/1024/1024, filename
        }' | head -n 20
    else
        log_warn "Backup directory does not exist: $BACKUP_DIR"
    fi
}

show_usage() {
    cat << EOF
Pews Database Backup Script

Usage: $0 [OPTIONS]

Options:
    -h, --help              Show this help message
    -l, --list              List available backups
    -r, --retention DAYS    Set retention period (default: $BACKUP_RETENTION_DAYS)
    -d, --dir PATH          Set backup directory (default: $BACKUP_DIR)
    -n, --no-cleanup        Skip cleanup of old backups

Examples:
    $0                      # Create backup with default settings
    $0 --list               # List all available backups
    $0 --retention 14       # Create backup and keep 14 days of backups
    $0 --no-cleanup         # Create backup without deleting old ones

Cron Example (daily at 2 AM):
    0 2 * * * cd $PROJECT_ROOT && $0 >> $PROJECT_ROOT/logs/backup.log 2>&1

EOF
}

# ============================================
# Main Script
# ============================================
main() {
    local SKIP_CLEANUP=false
    local LIST_ONLY=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -l|--list)
                LIST_ONLY=true
                shift
                ;;
            -r|--retention)
                BACKUP_RETENTION_DAYS="$2"
                shift 2
                ;;
            -d|--dir)
                BACKUP_DIR="$2"
                shift 2
                ;;
            -n|--no-cleanup)
                SKIP_CLEANUP=true
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # If list only, show backups and exit
    if [[ "$LIST_ONLY" = true ]]; then
        list_backups
        exit 0
    fi
    
    # Run backup process
    log_info "=== Pews Database Backup ==="
    log_info "Started at: $(date)"
    
    check_container
    create_backup_dir
    
    if perform_backup; then
        if [[ "$SKIP_CLEANUP" = false ]]; then
            cleanup_old_backups
        fi
        
        log_info "=== Backup Process Completed ==="
        log_info "Finished at: $(date)"
        exit 0
    else
        log_error "=== Backup Process Failed ==="
        exit 1
    fi
}

main "$@"
