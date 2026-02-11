# Pews Production Scripts

This directory contains operational scripts for managing the Pews application in production.

## Available Scripts

### backup.sh
Database backup script with automatic retention management.

**Features:**
- Creates timestamped PostgreSQL dumps
- Compresses backups with gzip
- Automatically cleans up old backups (default: 7 days)
- Can be scheduled via cron

**Usage:**
```bash
# Create a backup with default settings
./scripts/backup.sh

# List available backups
./scripts/backup.sh --list

# Create backup with custom retention
./scripts/backup.sh --retention 14

# Create backup without cleanup
./scripts/backup.sh --no-cleanup

# Custom backup directory
./scripts/backup.sh --dir /path/to/backups
```

**Cron Example:**
```cron
# Daily backup at 2 AM
0 2 * * * cd /home/CTMProd/pews-demo && ./scripts/backup.sh >> logs/backup.log 2>&1
```

**Restore Example:**
```bash
# Restore from backup
gunzip -c backups/pews_backup_YYYYMMDD_HHMMSS.sql.gz | \
  docker exec -i pews-postgres-1 psql -U pews pews
```

### deploy.sh
Zero-downtime deployment script for production.

**Features:**
- Builds Docker images locally
- Transfers images to remote server
- Rolling deployment (scales up, then down)
- Health checks after deployment
- Automatic rollback on failure
- Syncs configuration files

**Usage:**
```bash
# Full deployment
./scripts/deploy.sh

# Deploy to custom target
./scripts/deploy.sh --target user@host --path /custom/path

# Skip building (use existing images)
./scripts/deploy.sh --skip-build

# Only update configuration
./scripts/deploy.sh --config-only

# Deploy and show logs
./scripts/deploy.sh --logs

# Skip verification checks
./scripts/deploy.sh --no-verify
```

**Prerequisites:**
1. SSH access to target server configured
2. Docker installed locally and on remote
3. `.env.production` file configured
4. Secret files created on remote server:
   - `secrets/db_password.txt`
   - `secrets/jwt_secret.txt`
   - `secrets/stripe_secret_key.txt`

## Setting Up Secrets

On the remote server, create the secrets directory and files:

```bash
cd /home/CTMProd/pews-demo
mkdir -p secrets

# Database password
echo "your-strong-db-password" > secrets/db_password.txt

# JWT secret (generate with: openssl rand -hex 64)
openssl rand -hex 64 > secrets/jwt_secret.txt

# Stripe secret key
echo "sk_live_your_stripe_key" > secrets/stripe_secret_key.txt

# Set proper permissions
chmod 600 secrets/*.txt
```

## Environment Configuration

Copy `.env.production.example` to `.env.production` and configure:

```bash
cp .env.production.example .env.production
vim .env.production  # Edit with your values
```

**Critical variables:**
- `DOMAIN`: Your application domain
- `FRONTEND_URL`: Full URL with protocol
- `DEPLOY_TARGET`: SSH target for deployment
- `DEPLOY_PATH`: Remote deployment directory
- Stripe keys and secrets

## Traefik Integration

The production docker-compose file uses Traefik labels instead of hardcoded IPs.

**Benefits:**
- Automatic service discovery
- No manual IP management
- Dynamic configuration
- Built-in health checks

**Required:**
- Traefik must be running with Docker provider enabled
- Containers must be on `web_proxy` network
- Traefik must have access to Docker socket

**Migration from static config:**
Once deployed with labels, you can remove the static Traefik configuration file:
```bash
# Old: /home/CTMProd/homelabCloud/traefik/config/dynamic/pews-demo.yml
# This file is no longer needed - labels handle routing
```

## Monitoring

Check container status:
```bash
docker compose ps
docker compose logs -f backend
docker compose logs -f frontend
```

Check health endpoints:
```bash
curl https://demo.pews.app/api/health
```

View resource usage:
```bash
docker stats
```

## Troubleshooting

**Containers not starting:**
```bash
docker compose logs backend
docker compose logs frontend
# Check for missing secrets or environment variables
```

**Health checks failing:**
```bash
# Check if services are listening
docker compose exec backend wget -O- http://localhost:8080/api/health
docker compose exec frontend wget -O- http://localhost:3000
```

**Traefik not routing:**
```bash
# Verify containers are on web_proxy network
docker network inspect web_proxy

# Check Traefik logs
docker logs traefik | grep pews

# Verify labels are applied
docker inspect pews-backend-1 | grep -A 20 Labels
```

**Rollback to previous version:**
```bash
# Use previously saved images
docker images | grep pews
docker tag pews-backend:previous pews-backend:latest
docker compose up -d
```

## Best Practices

1. **Always test locally first:**
   ```bash
   docker compose -f docker-compose.prod.yml config
   docker compose -f docker-compose.prod.yml up -d
   ```

2. **Backup before deploying:**
   ```bash
   ./scripts/backup.sh
   ```

3. **Monitor logs during deployment:**
   ```bash
   ./scripts/deploy.sh --logs
   ```

4. **Keep secrets secure:**
   - Never commit secret files
   - Use Docker secrets, not environment variables
   - Rotate secrets regularly

5. **Schedule regular backups:**
   ```cron
   0 2 * * * cd /home/CTMProd/pews-demo && ./scripts/backup.sh
   ```

6. **Test restore procedure:**
   - Regularly verify backups can be restored
   - Document restore process
   - Practice in staging environment
