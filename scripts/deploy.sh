#!/usr/bin/env bash
# Pews Production Deployment Script
# Builds images, transfers to remote, and performs zero-downtime deployment

set -euo pipefail

# ============================================
# Configuration
# ============================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Load environment variables
if [[ -f "$PROJECT_ROOT/.env.production" ]]; then
    set -a
    source "$PROJECT_ROOT/.env.production"
    set +a
fi

# Default values
DEPLOY_TARGET="${DEPLOY_TARGET:-CTMProd@ctmprod}"
DEPLOY_PATH="${DEPLOY_PATH:-/home/CTMProd/pews-demo}"
COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-pews}"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

check_requirements() {
    log_step "Checking requirements..."
    
    local MISSING_DEPS=()
    
    if ! command -v docker &> /dev/null; then
        MISSING_DEPS+=("docker")
    fi
    
    if ! command -v ssh &> /dev/null; then
        MISSING_DEPS+=("ssh")
    fi
    
    if ! command -v rsync &> /dev/null; then
        log_warn "rsync not found, will use scp for file transfer (slower)"
    fi
    
    if [[ ${#MISSING_DEPS[@]} -gt 0 ]]; then
        log_error "Missing required dependencies: ${MISSING_DEPS[*]}"
        exit 1
    fi
    
    # Test SSH connection
    if ! ssh -q -o BatchMode=yes -o ConnectTimeout=5 "$DEPLOY_TARGET" exit; then
        log_error "Cannot connect to $DEPLOY_TARGET via SSH"
        log_info "Make sure SSH keys are configured and the host is reachable"
        exit 1
    fi
    
    log_info "All requirements satisfied"
}

build_images() {
    log_step "Building Docker images..."
    
    cd "$PROJECT_ROOT"
    
    # Build backend
    log_info "Building backend image..."
    if docker build -t pews-backend:latest -f Dockerfile .; then
        log_info "Backend image built successfully"
    else
        log_error "Failed to build backend image"
        exit 1
    fi
    
    # Build frontend
    log_info "Building frontend image..."
    if docker build -t pews-frontend:latest \
        --build-arg VITE_API_URL="${FRONTEND_URL:-https://demo.pews.app}" \
        -f web/Dockerfile ./web; then
        log_info "Frontend image built successfully"
    else
        log_error "Failed to build frontend image"
        exit 1
    fi
    
    log_info "All images built successfully"
}

transfer_images() {
    log_step "Transferring images to remote server..."
    
    # Create temporary directory for image tarballs
    local TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT
    
    # Save images to tar files
    log_info "Saving backend image..."
    docker save pews-backend:latest | gzip > "$TEMP_DIR/pews-backend.tar.gz"
    
    log_info "Saving frontend image..."
    docker save pews-frontend:latest | gzip > "$TEMP_DIR/pews-frontend.tar.gz"
    
    # Transfer to remote
    log_info "Transferring images to $DEPLOY_TARGET..."
    
    if command -v rsync &> /dev/null; then
        rsync -avz --progress "$TEMP_DIR/"*.tar.gz "$DEPLOY_TARGET:$DEPLOY_PATH/"
    else
        scp "$TEMP_DIR/"*.tar.gz "$DEPLOY_TARGET:$DEPLOY_PATH/"
    fi
    
    # Load images on remote
    log_info "Loading images on remote server..."
    ssh "$DEPLOY_TARGET" << EOF
        cd "$DEPLOY_PATH"
        docker load < pews-backend.tar.gz
        docker load < pews-frontend.tar.gz
        rm -f pews-backend.tar.gz pews-frontend.tar.gz
        echo "Images loaded successfully"
EOF
    
    log_info "Images transferred successfully"
}

sync_config_files() {
    log_step "Syncing configuration files..."
    
    # Sync docker-compose.prod.yml
    log_info "Syncing docker-compose.prod.yml..."
    scp "$PROJECT_ROOT/docker-compose.prod.yml" "$DEPLOY_TARGET:$DEPLOY_PATH/docker-compose.yml"
    
    # Sync .env.production if it exists (but not secrets!)
    if [[ -f "$PROJECT_ROOT/.env.production" ]]; then
        log_info "Syncing .env.production..."
        scp "$PROJECT_ROOT/.env.production" "$DEPLOY_TARGET:$DEPLOY_PATH/.env"
    fi
    
    log_info "Configuration files synced"
}

create_secrets() {
    log_step "Ensuring Docker secrets are configured..."
    
    ssh "$DEPLOY_TARGET" << 'EOF'
        cd /home/CTMProd/pews-demo
        
        # Create secrets directory if it doesn't exist
        mkdir -p secrets
        
        # Check if secret files exist
        if [[ ! -f secrets/db_password.txt ]]; then
            echo "WARNING: secrets/db_password.txt not found on remote"
            echo "Please create this file manually with the database password"
        fi
        
        if [[ ! -f secrets/jwt_secret.txt ]]; then
            echo "WARNING: secrets/jwt_secret.txt not found on remote"
            echo "Please create this file manually with the JWT secret"
        fi
        
        if [[ ! -f secrets/stripe_secret_key.txt ]]; then
            echo "WARNING: secrets/stripe_secret_key.txt not found on remote"
            echo "Please create this file manually with the Stripe secret key"
        fi
        
        # Set proper permissions
        chmod 600 secrets/*.txt 2>/dev/null || true
        
        echo "Secret files check complete"
EOF
}

deploy_with_zero_downtime() {
    log_step "Deploying with zero-downtime..."
    
    ssh "$DEPLOY_TARGET" << EOF
        set -e
        cd "$DEPLOY_PATH"
        
        echo "Current container status:"
        docker compose ps
        
        echo ""
        echo "Starting deployment..."
        
        # Pull/update postgres first (if needed)
        echo "Ensuring database is running..."
        docker compose up -d postgres
        sleep 5
        
        # Deploy backend with rolling update
        echo "Deploying backend..."
        docker compose up -d --no-deps --scale backend=2 backend
        sleep 10
        
        # Health check backend
        echo "Checking backend health..."
        BACKEND_CONTAINER=\$(docker compose ps -q backend | head -n 1)
        if docker exec \$BACKEND_CONTAINER wget -q -O- http://localhost:8080/api/health > /dev/null 2>&1; then
            echo "Backend health check passed"
        else
            echo "WARNING: Backend health check failed, but continuing..."
        fi
        
        # Scale back to 1 backend instance
        docker compose up -d --no-deps --scale backend=1 backend
        sleep 5
        
        # Deploy frontend with rolling update
        echo "Deploying frontend..."
        docker compose up -d --no-deps --scale frontend=2 frontend
        sleep 10
        
        # Health check frontend
        echo "Checking frontend health..."
        FRONTEND_CONTAINER=\$(docker compose ps -q frontend | head -n 1)
        if docker exec \$FRONTEND_CONTAINER wget -q -O- http://localhost:3000 > /dev/null 2>&1; then
            echo "Frontend health check passed"
        else
            echo "WARNING: Frontend health check failed, but continuing..."
        fi
        
        # Scale back to 1 frontend instance
        docker compose up -d --no-deps --scale frontend=1 frontend
        
        echo ""
        echo "Deployment complete! Current status:"
        docker compose ps
        
        # Cleanup old images
        echo ""
        echo "Cleaning up old images..."
        docker image prune -f
EOF
    
    log_info "Zero-downtime deployment completed"
}

verify_deployment() {
    log_step "Verifying deployment..."
    
    local ERRORS=0
    
    # Check container status
    log_info "Checking container health..."
    if ssh "$DEPLOY_TARGET" "cd $DEPLOY_PATH && docker compose ps --format json" | grep -q '"Health":"healthy"'; then
        log_info "✓ Containers are healthy"
    else
        log_warn "⚠ Some containers may not be healthy"
        ERRORS=$((ERRORS + 1))
    fi
    
    # Check if services are reachable (if we have a domain)
    if [[ -n "${FRONTEND_URL:-}" ]]; then
        log_info "Checking application endpoint..."
        if curl -ksf "${FRONTEND_URL}/api/health" > /dev/null 2>&1; then
            log_info "✓ Application is responding"
        else
            log_warn "⚠ Application endpoint check failed"
            ERRORS=$((ERRORS + 1))
        fi
    fi
    
    if [[ $ERRORS -eq 0 ]]; then
        log_info "✓ All verification checks passed"
        return 0
    else
        log_warn "⚠ Verification completed with $ERRORS warning(s)"
        return 1
    fi
}

show_logs() {
    log_step "Showing recent logs..."
    ssh "$DEPLOY_TARGET" << EOF
        cd "$DEPLOY_PATH"
        echo "=== Backend logs (last 20 lines) ==="
        docker compose logs --tail=20 backend
        echo ""
        echo "=== Frontend logs (last 20 lines) ==="
        docker compose logs --tail=20 frontend
EOF
}

rollback() {
    log_error "Deployment failed! Rolling back..."
    ssh "$DEPLOY_TARGET" << EOF
        cd "$DEPLOY_PATH"
        docker compose down
        docker compose up -d
        echo "Rollback complete"
EOF
}

show_usage() {
    cat << EOF
Pews Production Deployment Script

Usage: $0 [OPTIONS]

Options:
    -h, --help              Show this help message
    -t, --target HOST       SSH target (default: $DEPLOY_TARGET)
    -p, --path PATH         Remote deployment path (default: $DEPLOY_PATH)
    --skip-build            Skip building images (use existing local images)
    --skip-transfer         Skip transferring images (use images already on remote)
    --config-only           Only sync configuration files, don't restart services
    --logs                  Show logs after deployment
    --no-verify             Skip post-deployment verification

Examples:
    $0                      # Full deployment with default settings
    $0 --target user@host   # Deploy to custom target
    $0 --skip-build         # Deploy using existing local images
    $0 --config-only        # Update config without restarting
    $0 --logs               # Deploy and show logs

EOF
}

# ============================================
# Main Script
# ============================================
main() {
    local SKIP_BUILD=false
    local SKIP_TRANSFER=false
    local CONFIG_ONLY=false
    local SHOW_LOGS=false
    local SKIP_VERIFY=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -t|--target)
                DEPLOY_TARGET="$2"
                shift 2
                ;;
            -p|--path)
                DEPLOY_PATH="$2"
                shift 2
                ;;
            --skip-build)
                SKIP_BUILD=true
                shift
                ;;
            --skip-transfer)
                SKIP_TRANSFER=true
                shift
                ;;
            --config-only)
                CONFIG_ONLY=true
                shift
                ;;
            --logs)
                SHOW_LOGS=true
                shift
                ;;
            --no-verify)
                SKIP_VERIFY=true
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    log_info "=== Pews Production Deployment ==="
    log_info "Started at: $(date)"
    log_info "Target: $DEPLOY_TARGET"
    log_info "Path: $DEPLOY_PATH"
    echo ""
    
    check_requirements
    
    if [[ "$CONFIG_ONLY" = true ]]; then
        sync_config_files
        log_info "=== Configuration Update Complete ==="
        exit 0
    fi
    
    # Build and transfer images
    if [[ "$SKIP_BUILD" = false ]]; then
        build_images
    else
        log_info "Skipping build (using existing images)"
    fi
    
    if [[ "$SKIP_TRANSFER" = false ]]; then
        transfer_images
    else
        log_info "Skipping transfer (using images on remote)"
    fi
    
    # Sync configuration
    sync_config_files
    create_secrets
    
    # Deploy
    if deploy_with_zero_downtime; then
        log_info "✓ Deployment successful"
    else
        rollback
        exit 1
    fi
    
    # Verify
    if [[ "$SKIP_VERIFY" = false ]]; then
        verify_deployment || log_warn "Verification had warnings"
    fi
    
    # Show logs if requested
    if [[ "$SHOW_LOGS" = true ]]; then
        show_logs
    fi
    
    log_info "=== Deployment Complete ==="
    log_info "Finished at: $(date)"
    log_info "Application URL: ${FRONTEND_URL:-N/A}"
}

main "$@"
