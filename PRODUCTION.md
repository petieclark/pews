# Pews Production Deployment Guide

This guide covers deploying Pews to production with Docker Compose and Traefik.

## Overview

The production setup includes:
- **Hardened Docker Compose configuration** with health checks, resource limits, and restart policies
- **Docker secrets** for sensitive credentials (no plaintext passwords in env vars)
- **Traefik integration** using labels for automatic service discovery (no hardcoded IPs)
- **Automated backup script** with retention management
- **Zero-downtime deployment script** with health checks and rollback

## Prerequisites

### Local Machine
- Docker and Docker Compose installed
- SSH access to production server
- Git repository cloned

### Production Server
- Docker and Docker Compose installed
- Traefik running with Docker provider enabled
- `web_proxy` network created
- SSH access configured

## Initial Setup

### 1. Configure Environment

Copy and configure production environment:
```bash
cp .env.production.example .env.production
vim .env.production
```

**Required variables:**
```bash
DOMAIN=demo.pews.app
FRONTEND_URL=https://demo.pews.app
DEPLOY_TARGET=CTMProd@ctmprod
DEPLOY_PATH=/home/CTMProd/pews-demo
```

### 2. Set Up Secrets

**On production server:**
```bash
ssh CTMProd@ctmprod
cd /home/CTMProd/pews-demo
mkdir -p secrets

# Database password
echo "your-secure-db-password-min-32-chars" > secrets/db_password.txt

# JWT secret (generate strong random value)
openssl rand -hex 64 > secrets/jwt_secret.txt

# Stripe secret key
echo "sk_live_your_stripe_key" > secrets/stripe_secret_key.txt

# Secure the files
chmod 600 secrets/*.txt
ls -la secrets/
```

### 3. Create Required Directories

**On production server:**
```bash
cd /home/CTMProd/pews-demo
mkdir -p backups logs
```

### 4. Verify Traefik Configuration

Ensure Traefik is configured for Docker provider:
```bash
# Check Traefik config
cat /home/CTMProd/homelabCloud/traefik/traefik.yml

# Should include:
# providers:
#   docker:
#     endpoint: "unix:///var/run/docker.sock"
#     exposedByDefault: false
#     network: web_proxy

# Verify web_proxy network exists
docker network ls | grep web_proxy
```

## Deployment

### First-Time Deployment

1. **Build and deploy:**
```bash
./scripts/deploy.sh
```

This will:
- Build Docker images locally
- Transfer images to production server
- Sync configuration files
- Deploy with zero-downtime strategy
- Verify deployment health

2. **Verify deployment:**
```bash
ssh CTMProd@ctmprod "cd /home/CTMProd/pews-demo && docker compose ps"
curl https://demo.pews.app/api/health
```

3. **Check Traefik routing:**
```bash
ssh CTMProd@ctmprod "docker logs traefik | grep pews"
```

### Subsequent Deployments

For updates:
```bash
# Full deployment
./scripts/deploy.sh

# Or with logs
./scripts/deploy.sh --logs

# Or skip rebuild if only config changed
./scripts/deploy.sh --config-only
```

## Traefik Integration Details

### Why Labels Instead of Static Config?

**Before (hardcoded IPs):**
```yaml
# /home/CTMProd/homelabCloud/traefik/config/dynamic/pews-demo.yml
services:
  pews-demo-backend:
    loadBalancer:
      servers:
        - url: "http://172.18.0.56:8080"  # FRAGILE!
```

**Problems:**
- IPs change when containers restart
- Manual IP management required
- No automatic health checks
- Brittle configuration

**After (dynamic labels):**
```yaml
# docker-compose.prod.yml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.pews-backend.rule=Host(`demo.pews.app`) && PathPrefix(`/api`)"
  - "traefik.http.services.pews-backend.loadbalancer.server.port=8080"
```

**Benefits:**
- Automatic service discovery
- Dynamic IP resolution
- Built-in health checks
- Self-healing configuration

### Traefik Labels Explained

The production docker-compose includes comprehensive Traefik labels:

```yaml
backend:
  labels:
    - "traefik.enable=true"  # Enable Traefik for this container
    
    # Router configuration
    - "traefik.http.routers.pews-backend.rule=Host(`demo.pews.app`) && PathPrefix(`/api`)"
    - "traefik.http.routers.pews-backend.entrypoints=websecure"  # HTTPS
    - "traefik.http.routers.pews-backend.tls=true"
    - "traefik.http.routers.pews-backend.tls.certresolver=myresolver"  # Let's Encrypt
    - "traefik.http.routers.pews-backend.priority=10"  # Higher priority for API routes
    
    # Service configuration
    - "traefik.http.services.pews-backend.loadbalancer.server.port=8080"
    - "traefik.http.services.pews-backend.loadbalancer.healthcheck.path=/api/health"
    - "traefik.http.services.pews-backend.loadbalancer.healthcheck.interval=30s"
    
    # Network (important!)
    - "traefik.docker.network=web_proxy"
```

### Removing Old Static Config

Once deployed with labels, the static config is obsolete:

```bash
ssh CTMProd@ctmprod
cd /home/CTMProd/homelabCloud/traefik/config/dynamic

# Backup the old config
mv pews-demo.yml pews-demo.yml.backup

# Restart Traefik to remove old routes
cd /home/CTMProd/homelabCloud/traefik
docker compose restart traefik

# Verify new dynamic routes are working
curl -k https://demo.pews.app/api/health
```

## Backup and Restore

### Automated Backups

Set up daily backups with cron:

```bash
ssh CTMProd@ctmprod
crontab -e

# Add daily backup at 2 AM
0 2 * * * cd /home/CTMProd/pews-demo && ./scripts/backup.sh >> logs/backup.log 2>&1
```

### Manual Backup

```bash
# On production server
cd /home/CTMProd/pews-demo
./scripts/backup.sh

# List backups
./scripts/backup.sh --list
```

### Restore from Backup

```bash
# On production server
cd /home/CTMProd/pews-demo

# Stop the application
docker compose down backend

# Restore database
gunzip -c backups/pews_backup_YYYYMMDD_HHMMSS.sql.gz | \
  docker compose exec -T postgres psql -U pews pews

# Restart
docker compose up -d
```

## Monitoring

### Container Health

```bash
# Check status
docker compose ps

# View logs
docker compose logs -f backend
docker compose logs -f frontend

# Resource usage
docker stats
```

### Application Health

```bash
# Health endpoints
curl https://demo.pews.app/api/health

# Test functionality
curl https://demo.pews.app/

# Check response times
curl -w "@curl-format.txt" -o /dev/null -s https://demo.pews.app/
```

### Traefik Dashboard

Access Traefik dashboard (if enabled):
```bash
https://traefik.petieclark.com/dashboard/
```

Look for:
- Pews routers active
- Services showing healthy
- No backend errors

## Troubleshooting

### Containers Won't Start

```bash
# Check logs
docker compose logs postgres
docker compose logs backend
docker compose logs frontend

# Common issues:
# 1. Missing secrets
ls -la secrets/

# 2. Database connection
docker compose exec backend env | grep DATABASE

# 3. Port conflicts
docker ps | grep -E "(8080|3000|5432)"
```

### Health Checks Failing

```bash
# Test health checks manually
docker compose exec backend wget -O- http://localhost:8080/api/health
docker compose exec frontend wget -O- http://localhost:3000

# Check container logs
docker compose logs backend | grep -i error
```

### Traefik Not Routing

```bash
# Verify labels are applied
docker inspect pews-backend-1 | grep -A 30 Labels

# Check Traefik can see the container
docker logs traefik | grep pews

# Verify network connectivity
docker network inspect web_proxy | grep pews

# Test from Traefik container
docker exec traefik wget -O- http://pews-backend-1:8080/api/health
```

### Database Issues

```bash
# Connect to database
docker compose exec postgres psql -U pews pews

# Check connections
SELECT * FROM pg_stat_activity;

# Check database size
SELECT pg_size_pretty(pg_database_size('pews'));

# Vacuum and analyze
VACUUM ANALYZE;
```

## Security Checklist

- [ ] All secrets stored in Docker secrets (not env vars)
- [ ] Secret files have 600 permissions
- [ ] Strong passwords (min 32 chars)
- [ ] JWT secret is random (min 64 chars)
- [ ] HTTPS enabled with valid certificates
- [ ] Database not exposed to public internet
- [ ] Regular backups scheduled
- [ ] Backup restore tested
- [ ] Resource limits configured
- [ ] Logs rotated (max-size/max-file set)
- [ ] `.env.production` not committed to git
- [ ] Stripe production keys (not test keys)

## Performance Tuning

### Resource Limits

Current limits in docker-compose.prod.yml:
```yaml
postgres:
  deploy:
    resources:
      limits:
        cpus: '1.0'
        memory: 512M

backend:
  deploy:
    resources:
      limits:
        cpus: '2.0'
        memory: 1G

frontend:
  deploy:
    resources:
      limits:
        cpus: '1.0'
        memory: 512M
```

Adjust based on actual usage:
```bash
# Monitor resource usage
docker stats

# If containers are hitting limits, increase in docker-compose.prod.yml
# Then redeploy:
./scripts/deploy.sh --config-only
```

### Database Optimization

```sql
-- Connect to database
docker compose exec postgres psql -U pews pews

-- Check slow queries
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Add indexes as needed
CREATE INDEX idx_users_email ON users(email);

-- Vacuum regularly
VACUUM ANALYZE;
```

## Scaling

### Horizontal Scaling

To run multiple instances:

1. Update docker-compose.prod.yml:
```yaml
backend:
  deploy:
    replicas: 3  # Add this
```

2. Deploy:
```bash
./scripts/deploy.sh
```

Traefik will automatically load balance across replicas.

### Vertical Scaling

Increase resource limits:
```yaml
backend:
  deploy:
    resources:
      limits:
        cpus: '4.0'
        memory: 2G
```

## Maintenance Windows

For major updates requiring downtime:

```bash
# 1. Notify users
# 2. Create backup
./scripts/backup.sh

# 3. Stop services
docker compose down

# 4. Update configuration/database schema
# 5. Deploy new version
./scripts/deploy.sh

# 6. Verify deployment
curl https://demo.pews.app/api/health

# 7. Restore if needed (from backup)
```

## Rollback Procedure

If deployment fails:

```bash
# Automatic rollback (handled by deploy.sh)
# Manual rollback if needed:

ssh CTMProd@ctmprod
cd /home/CTMProd/pews-demo

# Use previous images
docker images | grep pews
docker tag pews-backend:<previous-tag> pews-backend:latest
docker tag pews-frontend:<previous-tag> pews-frontend:latest

# Restart
docker compose down
docker compose up -d

# Verify
docker compose ps
curl https://demo.pews.app/api/health
```

## Support

For issues or questions:
1. Check logs: `docker compose logs`
2. Review this guide
3. Check Traefik documentation: https://doc.traefik.io/traefik/
4. Contact repository maintainer

## Change Log

- **2026-02-11**: Initial production hardening
  - Added Docker secrets support
  - Implemented Traefik labels (replaced hardcoded IPs)
  - Added health checks and resource limits
  - Created backup and deployment scripts
  - Added comprehensive documentation
