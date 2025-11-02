# Deployment Guide - Viskatera API Go

Panduan lengkap untuk deploy Viskatera API ke VPS Ubuntu 20.04 dengan Docker dan HTTPS (Let's Encrypt).

## 📋 Prerequisites

- VPS Ubuntu 20.04 dengan minimal 2GB RAM dan 20GB disk space
- Domain `api.ahmadcorp.com` sudah diarahkan ke IP VPS (A record)
- Akses root atau sudo ke VPS
- Docker dan Docker Compose sudah terinstall

## 🔧 Step 1: Persiapan Server

### 1.1 Update System

```bash
sudo apt update
sudo apt upgrade -y
```

### 1.2 Install Dependencies

```bash
# Install basic tools
sudo apt install -y curl wget git ufw fail2ban

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installations
docker --version
docker-compose --version
```

### 1.3 Setup Firewall

```bash
# Enable UFW
sudo ufw enable

# Allow SSH (jangan skip ini!)
sudo ufw allow 22/tcp

# Allow HTTP dan HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Check status
sudo ufw status
```

### 1.4 Setup Fail2Ban (Optional tapi recommended)

```bash
# Fail2ban sudah terinstall, cukup enable
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

## 📦 Step 2: Setup Project di Server

### 2.1 Clone atau Upload Project

```bash
# Pilihan 1: Clone dari repository
cd /opt
sudo git clone <your-repo-url> viskatera-api-go
cd viskatera-api-go

# Pilihan 2: Upload via SCP dari local machine
# scp -r /path/to/viskatera-api-go user@your-server:/opt/
```

### 2.2 Buat Directory untuk Data

```bash
sudo mkdir -p /opt/viskatera-api-go/backups
sudo mkdir -p /opt/viskatera-api-go/nginx/ssl
sudo chown -R $USER:$USER /opt/viskatera-api-go
```

### 2.3 Setup Environment Variables

```bash
# Copy env example
cp env.example .env

# Edit .env file dengan production values
nano .env
```

**Isi .env dengan nilai production:**

```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=viskatera_user
DB_PASSWORD=<GENERATE_STRONG_PASSWORD>
DB_NAME=viskatera_db

# Application
ENVIRONMENT=production
GIN_MODE=release
PORT=8080

# Security - GENERATE STRONG SECRET!
JWT_SECRET=<GENERATE_64_CHAR_RANDOM_STRING>

# Xendit (Production Keys)
XENDIT_SECRET_KEY=<your-production-secret-key>
XENDIT_PUBLIC_KEY=<your-production-public-key>
XENDIT_API_URL=https://api.xendit.co

# Google OAuth
GOOGLE_CLIENT_ID=<your-production-client-id>
GOOGLE_CLIENT_SECRET=<your-production-client-secret>
GOOGLE_REDIRECT_URL=https://api.ahmadcorp.com/api/v1/auth/google/callback

# URLs
APP_BASE_URL=https://api.ahmadcorp.com

# SMTP (Production)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=noreply@ahmadcorp.com
SMTP_PASS=<your-smtp-password>
SMTP_FROM=noreply@ahmadcorp.com

# File Upload
UPLOAD_DIR=/app/uploads
MAX_UPLOAD_SIZE=10485760
```

**Generate JWT_SECRET yang kuat:**
```bash
# Generate random 64 character string
openssl rand -base64 64 | tr -d '\n' | cut -c1-64
```

**Generate DB_PASSWORD yang kuat:**
```bash
openssl rand -base64 32
```

### 2.4 Set Permissions untuk .env

```bash
chmod 600 .env
```

## 🔒 Step 3: Setup SSL Certificate (Let's Encrypt)

### 3.1 Initial SSL Certificate Request

**PENTING:** Pastikan domain sudah mengarah ke IP VPS sebelum lanjut!

```bash
# Stop nginx sementara untuk initial certificate
docker-compose -f docker-compose.prod.yml down

# Request initial certificate
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

# Jika sukses, certificate akan ada di nginx/ssl/live/api.ahmadcorp.com/
```

**Alternative: Manual Certificate Setup (Jika otomatis gagal)**

```bash
# Install certbot di host
sudo apt install -y certbot

# Request certificate
sudo certbot certonly --standalone -d api.ahmadcorp.com

# Copy certificates ke directory project
sudo mkdir -p nginx/ssl/live/api.ahmadcorp.com
sudo cp /etc/letsencrypt/live/api.ahmadcorp.com/fullchain.pem nginx/ssl/live/api.ahmadcorp.com/
sudo cp /etc/letsencrypt/live/api.ahmadcorp.com/privkey.pem nginx/ssl/live/api.ahmadcorp.com/
sudo cp /etc/letsencrypt/live/api.ahmadcorp.com/chain.pem nginx/ssl/live/api.ahmadcorp.com/
sudo chown -R $USER:$USER nginx/ssl
```

### 3.2 Update nginx config untuk initial setup

Sebelum certificate tersedia, update nginx config untuk allow HTTP sementara:

Edit `nginx/conf.d/api.ahmadcorp.com.conf` dan comment SSL lines untuk initial setup.

## 🚀 Step 4: Build dan Deploy

### 4.1 Build Docker Images

```bash
# Build images
docker-compose -f docker-compose.prod.yml build --no-cache
```

### 4.2 Start Services

```bash
# Start semua services
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# Check logs
docker-compose -f docker-compose.prod.yml logs -f
```

### 4.3 Verify Deployment

```bash
# Check health endpoint
curl http://localhost:8080/health

# Check via nginx (setelah SSL setup)
curl https://api.ahmadcorp.com/health

# Check all containers running
docker ps
```

## 🔄 Step 5: Setup Auto-Renewal SSL Certificate

### 5.1 Create Renewal Script

```bash
cat > /opt/viskatera-api-go/renew-ssl.sh << 'EOF'
#!/bin/bash
cd /opt/viskatera-api-go

# Renew certificate
docker-compose -f docker-compose.prod.yml exec certbot certbot renew

# Reload nginx
docker-compose -f docker-compose.prod.yml exec nginx nginx -s reload
EOF

chmod +x /opt/viskatera-api-go/renew-ssl.sh
```

### 5.2 Setup Cron Job untuk Auto-Renewal

```bash
# Edit crontab
crontab -e

# Tambahkan baris ini (cek setiap 12 jam, renew jika perlu)
0 */12 * * * /opt/viskatera-api-go/renew-ssl.sh >> /var/log/ssl-renewal.log 2>&1
```

**Alternative: Gunakan Certbot container yang sudah ada di docker-compose.prod.yml**

Container certbot di docker-compose sudah otomatis renew setiap 12 jam.

## 📊 Step 6: Monitoring dan Maintenance

### 6.1 Check Logs

```bash
# API logs
docker-compose -f docker-compose.prod.yml logs -f api

# Nginx logs
docker-compose -f docker-compose.prod.yml logs -f nginx

# All logs
docker-compose -f docker-compose.prod.yml logs -f
```

### 6.2 Database Backup

**Manual Backup:**
```bash
docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U viskatera_user viskatera_db > backups/backup_$(date +%Y%m%d_%H%M%S).sql
```

**Automated Daily Backup Script:**
```bash
cat > /opt/viskatera-api-go/backup-db.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/viskatera-api-go/backups"
DATE=$(date +%Y%m%d_%H%M%S)
FILENAME="backup_${DATE}.sql"

cd /opt/viskatera-api-go

docker-compose -f docker-compose.prod.yml exec -T postgres pg_dump -U viskatera_user viskatera_db > "${BACKUP_DIR}/${FILENAME}"

# Compress backup
gzip "${BACKUP_DIR}/${FILENAME}"

# Keep only last 7 days of backups
find "${BACKUP_DIR}" -name "backup_*.sql.gz" -mtime +7 -delete

echo "Backup completed: ${FILENAME}.gz"
EOF

chmod +x /opt/viskatera-api-go/backup-db.sh

# Add to crontab (daily at 2 AM)
crontab -e
# Add: 0 2 * * * /opt/viskatera-api-go/backup-db.sh >> /var/log/db-backup.log 2>&1
```

### 6.3 Restore Database

```bash
# Decompress if compressed
gunzip backups/backup_YYYYMMDD_HHMMSS.sql.gz

# Restore
cat backups/backup_YYYYMMDD_HHMMSS.sql | docker-compose -f docker-compose.prod.yml exec -T postgres psql -U viskatera_user -d viskatera_db
```

## 🔄 Step 7: Update Application

```bash
cd /opt/viskatera-api-go

# Pull latest code (jika dari git)
git pull

# Rebuild images
docker-compose -f docker-compose.prod.yml build --no-cache api

# Restart dengan zero downtime (rolling update)
docker-compose -f docker-compose.prod.yml up -d --no-deps api

# Atau restart semua services
docker-compose -f docker-compose.prod.yml restart
```

## 🛠️ Troubleshooting

### Issue: Certificate tidak ter-renew

```bash
# Test renewal manual
docker-compose -f docker-compose.prod.yml exec certbot certbot renew --dry-run

# Check certificate expiry
openssl x509 -in nginx/ssl/live/api.ahmadcorp.com/fullchain.pem -noout -dates
```

### Issue: Nginx tidak start

```bash
# Test nginx config
docker-compose -f docker-compose.prod.yml exec nginx nginx -t

# Check nginx logs
docker-compose -f docker-compose.prod.yml logs nginx
```

### Issue: API tidak bisa connect ke database

```bash
# Check database logs
docker-compose -f docker-compose.prod.yml logs postgres

# Test database connection
docker-compose -f docker-compose.prod.yml exec api ping postgres

# Check database is ready
docker-compose -f docker-compose.prod.yml exec postgres pg_isready -U viskatera_user
```

### Issue: Container restart loop

```bash
# Check container logs
docker-compose -f docker-compose.prod.yml logs --tail=100 api

# Check resource usage
docker stats

# Check disk space
df -h
```

## 📈 Performance Optimization

### 1. Increase Nginx Worker Connections

Edit `nginx/nginx.conf`:
```nginx
events {
    worker_connections 2048;  # Increase dari 1024
}
```

Restart nginx:
```bash
docker-compose -f docker-compose.prod.yml restart nginx
```

### 2. Database Connection Pool

Sudah dikonfigurasi di `config/database.go`. Sesuaikan jika perlu:
- MaxIdleConns: 10
- MaxOpenConns: 100

### 3. Enable HTTP/2

Sudah diaktifkan di nginx config dengan `http2`.

## 🔐 Security Hardening

### 1. Update .env Permissions
```bash
chmod 600 .env
```

### 2. Regular Security Updates
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Update Docker images
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d
```

### 3. Monitor Logs untuk Suspicious Activity
```bash
# Check nginx access logs untuk suspicious requests
tail -f /var/log/nginx/api.ahmadcorp.com.access.log | grep -E "(401|403|404|500)"
```

## 📝 Environment Variables Checklist

Pastikan semua environment variables berikut sudah di-set dengan benar:

- [ ] `DB_PASSWORD` - Strong password
- [ ] `JWT_SECRET` - 64 character random string
- [ ] `XENDIT_SECRET_KEY` - Production key dari Xendit
- [ ] `XENDIT_PUBLIC_KEY` - Production key dari Xendit
- [ ] `GOOGLE_CLIENT_ID` - Production client ID
- [ ] `GOOGLE_CLIENT_SECRET` - Production client secret
- [ ] `SMTP_PASS` - SMTP password untuk email

## ✅ Post-Deployment Checklist

- [ ] SSL certificate valid dan auto-renewal setup
- [ ] Health check endpoint accessible: `https://api.ahmadcorp.com/health`
- [ ] API endpoints working: `https://api.ahmadcorp.com/api/v1/...`
- [ ] Database backup script setup dan tested
- [ ] Logs accessible dan monitored
- [ ] Firewall configured (UFW enabled)
- [ ] Domain DNS pointing ke VPS IP
- [ ] All environment variables set correctly
- [ ] Swagger docs accessible: `https://api.ahmadcorp.com/swagger/index.html`

## 🆘 Support & Maintenance

### Useful Commands

```bash
# View running containers
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f [service_name]

# Restart service
docker-compose -f docker-compose.prod.yml restart [service_name]

# Stop all services
docker-compose -f docker-compose.prod.yml down

# Start all services
docker-compose -f docker-compose.prod.yml up -d

# Execute command in container
docker-compose -f docker-compose.prod.yml exec [service_name] [command]

# View resource usage
docker stats
```

---

**Selamat! API Anda sudah siap untuk production! 🎉**

Jika ada pertanyaan atau issue, check logs terlebih dahulu dan refer ke PRODUCTION_CHECKLIST.md untuk best practices.

