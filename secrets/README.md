# Secrets Directory

This directory contains sensitive credentials for production deployment using Docker secrets.

## Required Files

Create the following files with their respective secrets:

### db_password.txt
PostgreSQL database password
```bash
echo "your-strong-database-password" > db_password.txt
chmod 600 db_password.txt
```

### jwt_secret.txt
JWT signing secret (minimum 64 characters recommended)
```bash
# Generate with OpenSSL
openssl rand -hex 64 > jwt_secret.txt
chmod 600 jwt_secret.txt
```

### stripe_secret_key.txt
Stripe API secret key
```bash
echo "sk_live_your_stripe_secret_key" > stripe_secret_key.txt
chmod 600 stripe_secret_key.txt
```

## Security Notes

⚠️ **NEVER commit these files to version control**

- All files in this directory are ignored by git
- Set file permissions to 600 (read/write for owner only)
- Use different secrets for development/staging/production
- Rotate secrets regularly
- Store backups securely (password manager, encrypted vault)

## Production Setup

On production server:
```bash
cd /home/CTMProd/pews-demo
mkdir -p secrets
cd secrets

# Create each secret file
echo "production-db-password" > db_password.txt
openssl rand -hex 64 > jwt_secret.txt
echo "sk_live_xxx" > stripe_secret_key.txt

# Set proper permissions
chmod 600 *.txt
```

## Docker Secrets

These files are mounted into containers as `/run/secrets/[filename]` and read by the application at runtime. This is more secure than environment variables because:

1. Secrets are not visible in `docker inspect`
2. Secrets are not stored in image layers
3. Secrets are only available to authorized containers
4. Secrets can be rotated without rebuilding images

## Verification

Check if secrets are properly mounted:
```bash
docker compose ps
docker compose exec backend ls -la /run/secrets/
```
