# Docker Deployment Quick Start

## 🚀 Quick Deployment

### Prerequisites
- Docker & Docker Compose installed
- Domain `api.ahmadcorp.com` pointing to server IP
- Environment variables configured in `.env`

### Steps

1. **Setup Environment**
   ```bash
   cp env.example .env
   nano .env  # Edit with production values
   ```

2. **Build & Deploy**
   ```bash
   ./deploy.sh build
   ./deploy.sh up
   ```

3. **Setup SSL (First Time)**
   ```bash
   # Option 1: Using certbot container
   docker run -it --rm \
     -v "$(pwd)/nginx/ssl:/etc/letsencrypt" \
     -v "$(pwd)/nginx/certbot:/var/www/certbot" \
     certbot/certbot certonly \
     --webroot \
     --webroot-path=/var/www/certbot \
     --email admin@ahmadcorp.com \
     --agree-tos \
     --no-eff-email \
     -d api.ahmadcorp.com

   # Option 2: Using host certbot (recommended for initial setup)
   sudo apt install certbot
   sudo certbot certonly --standalone -d api.ahmadcorp.com
   
   # Copy certificates
   sudo cp /etc/letsencrypt/live/api.ahmadcorp.com/* nginx/ssl/live/api.ahmadcorp.com/
   sudo chown -R $USER:$USER nginx/ssl
   ```

4. **Restart Services**
   ```bash
   ./deploy.sh restart
   ```

## 📝 Common Commands

```bash
# Start services
./deploy.sh up

# Stop services
./deploy.sh down

# View logs
./deploy.sh logs
./deploy.sh logs api      # Specific service

# Backup database
./deploy.sh backup

# Restore database
./deploy.sh restore backups/backup_20240101_120000.sql.gz

# Health check
./deploy.sh health

# Rebuild after code changes
./deploy.sh build
./deploy.sh up
```

## 🔧 Manual Docker Compose Commands

```bash
# Start
docker-compose -f docker-compose.prod.yml up -d

# Stop
docker-compose -f docker-compose.prod.yml down

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Restart specific service
docker-compose -f docker-compose.prod.yml restart api

# Execute command in container
docker-compose -f docker-compose.prod.yml exec api sh
docker-compose -f docker-compose.prod.yml exec postgres psql -U viskatera_user -d viskatera_db
```

## 📊 Service URLs

- API: `https://api.ahmadcorp.com`
- Health Check: `https://api.ahmadcorp.com/health`
- Swagger Docs: `https://api.ahmadcorp.com/swagger/index.html`
- API Base: `https://api.ahmadcorp.com/api/v1`

## 🔍 Troubleshooting

### Check Service Status
```bash
docker-compose -f docker-compose.prod.yml ps
```

### Check Logs
```bash
# All services
docker-compose -f docker-compose.prod.yml logs

# Specific service
docker-compose -f docker-compose.prod.yml logs api
docker-compose -f docker-compose.prod.yml logs nginx
```

### View Resource Usage
```bash
docker stats
```

### Database Connection Issue
```bash
# Check database is running
docker-compose -f docker-compose.prod.yml exec postgres pg_isready

# Test connection from API container
docker-compose -f docker-compose.prod.yml exec api ping postgres
```

### SSL Certificate Issues
```bash
# Check certificate expiry
openssl x509 -in nginx/ssl/live/api.ahmadcorp.com/fullchain.pem -noout -dates

# Test renewal
docker-compose -f docker-compose.prod.yml exec certbot certbot renew --dry-run
```

## 🔄 Update Application

```bash
# Pull latest code (if using git)
git pull

# Rebuild and restart
./deploy.sh build
./deploy.sh restart
```

## 📦 Backup & Restore

### Automated Backup
Setup cron job:
```bash
# Daily backup at 2 AM
0 2 * * * cd /opt/viskatera-api-go && ./deploy.sh backup
```

### Manual Backup
```bash
./deploy.sh backup
```

### Restore
```bash
./deploy.sh restore backups/backup_20240101_120000.sql.gz
```

## 🔒 Security Notes

1. **Never commit .env file** - Contains sensitive data
2. **Use strong passwords** - Generate with `openssl rand -base64 32`
3. **Regular updates** - Update Docker images and system packages
4. **Monitor logs** - Check for suspicious activity
5. **Firewall** - Ensure UFW is enabled and configured

## 📚 More Information

- Full deployment guide: `DEPLOYMENT_GUIDE.md`
- Production checklist: `PRODUCTION_CHECKLIST.md`
- API documentation: `https://api.ahmadcorp.com/swagger/index.html`

