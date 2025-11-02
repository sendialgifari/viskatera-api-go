# Deployment Guide - Best Practices

## 📋 Table of Contents

- [Development Mode](#development-mode)
- [Production Mode](#production-mode)
- [Database Migration](#database-migration)
- [Environment Configuration](#environment-configuration)
- [Troubleshooting](#troubleshooting)

---

## 🛠️ Development Mode

### Quick Start

```bash
# Start everything (database + app)
./app.sh dev start
```

This will:
1. ✅ Check prerequisites (Go, Docker)
2. ✅ Create `.env` file if missing
3. ✅ Start PostgreSQL database (Docker)
4. ✅ Install dependencies
5. ✅ Start application with live reload (if Air/Reflex available)

### Development Workflow

#### Initial Setup
```bash
# 1. Setup environment
./app.sh dev start

# 2. Run fresh migration (if starting fresh)
./app.sh dev fresh-migrate

# 3. Seed sample data
./app.sh dev seed

# 4. Create admin user
./app.sh dev admin
```

#### Daily Development
```bash
# Start application
./app.sh dev start

# Or restart if already running
./app.sh dev restart

# Check status
./app.sh dev status
```

#### Working with Database
```bash
# Run regular migration (safe, preserves data)
./app.sh dev migrate

# Fresh migration (DANGER: deletes all data - dev only!)
./app.sh dev fresh-migrate

# Seed database
./app.sh dev seed
```

### Development Features

- **Live Reload**: Automatically restarts on code changes (Air/Reflex)
- **MailHog**: Email testing at http://localhost:8025
- **Database Admin**: Adminer at http://localhost:8081
- **Auto Database Setup**: Automatically starts and configures PostgreSQL
- **Hot Reload**: Code changes trigger automatic rebuild

### Development Environment Variables

```bash
# .env for development
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=viskatera_db
JWT_SECRET=dev-secret-key-change-in-production
PORT=8080

# SMTP (empty = use MailHog)
SMTP_HOST=
SMTP_PORT=1025
SMTP_FROM=noreply@viskatera.com
```

### Recommended Development Tools

1. **Air** (Recommended)
   ```bash
   go install github.com/air-verse/air@latest
   ```

2. **Reflex** (Alternative)
   ```bash
   go install github.com/cespare/reflex@latest
   ```

3. **MailHog** (For email testing)
   ```bash
   # Using Docker
   docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog
   ```

---

## 🚀 Production Mode

### Prerequisites

- ✅ Go 1.21+ installed
- ✅ PostgreSQL database (separate from app server)
- ✅ Environment variables configured
- ✅ SSL certificates (if using HTTPS)
- ✅ Process manager (systemd, PM2, or similar)

### Production Setup

#### 1. Environment Configuration

Create `.env` file with production values:

```bash
# Database (use production database, not Docker)
DB_HOST=your-production-db-host
DB_PORT=5432
DB_USER=viskatera_user
DB_PASSWORD=strong-password-here
DB_NAME=viskatera_db

# Security
JWT_SECRET=generate-strong-random-secret-here

# Server
PORT=8080
GIN_MODE=release

# SMTP (required for production)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your-email@example.com
SMTP_PASS=your-email-password
SMTP_FROM=noreply@yourdomain.com

# Application
APP_BASE_URL=https://api.yourdomain.com
```

**⚠️ Important**: 
- Never commit `.env` file to version control
- Use strong, randomly generated secrets
- Use environment-specific database credentials

#### 2. Build Application

```bash
# Build production binary
./app.sh prod build

# This creates optimized binary: viskatera-api
```

#### 3. Database Migration

```bash
# Run migration (safe, preserves data)
./app.sh prod migrate
```

**⚠️ Never run `fresh-migrate` in production!**

#### 4. Start Application

```bash
# Start application
./app.sh prod start

# Application runs as background process
# Logs saved to app.log
```

#### 5. Process Management

For production, use a proper process manager:

##### Using Systemd (Linux)

Create `/etc/systemd/system/viskatera-api.service`:

```ini
[Unit]
Description=Viskatera API Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/viskatera-api
ExecStart=/opt/viskatera-api/viskatera-api
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable viskatera-api
sudo systemctl start viskatera-api
sudo systemctl status viskatera-api
```

##### Using PM2 (Node.js Process Manager)

```bash
# Install PM2
npm install -g pm2

# Start application
pm2 start ./viskatera-api --name viskatera-api

# Save PM2 configuration
pm2 save

# Setup PM2 to start on boot
pm2 startup
```

### Production Best Practices

1. **Security**
   - ✅ Use strong JWT secrets (32+ characters, random)
   - ✅ Enable HTTPS (use reverse proxy like Nginx)
   - ✅ Set `GIN_MODE=release` for production
   - ✅ Use environment variables, never hardcode secrets
   - ✅ Regularly rotate secrets
   - ✅ Keep dependencies updated

2. **Database**
   - ✅ Use connection pooling
   - ✅ Regular backups
   - ✅ Monitor database performance
   - ✅ Use read replicas for scaling (if needed)

3. **Monitoring**
   - ✅ Set up logging (structured logs recommended)
   - ✅ Monitor application health (`/health` endpoint)
   - ✅ Set up alerts for errors
   - ✅ Monitor resource usage (CPU, memory, disk)

4. **Performance**
   - ✅ Use CDN for static assets
   - ✅ Enable gzip compression (via reverse proxy)
   - ✅ Use caching where appropriate
   - ✅ Optimize database queries

5. **Deployment**
   - ✅ Zero-downtime deployment (blue-green or rolling)
   - ✅ Database migrations run separately before deployment
   - ✅ Health checks before marking as ready
   - ✅ Rollback plan in place

### Production Checklist

Before going live:

- [ ] Environment variables configured
- [ ] Database migrated and tested
- [ ] SMTP configured and tested
- [ ] SSL certificates installed
- [ ] Reverse proxy configured (Nginx/Traefik)
- [ ] Process manager configured
- [ ] Monitoring and alerting setup
- [ ] Backup strategy in place
- [ ] Security audit completed
- [ ] Load testing performed
- [ ] Documentation updated

---

## 🗄️ Database Migration

### Regular Migration (Safe)

Runs automatically when models change. Preserves all data.

```bash
# Development
./app.sh dev migrate

# Production
./app.sh prod migrate
```

**When to use:**
- Adding new tables
- Adding new columns (with defaults)
- Adding indexes
- Safe schema changes

### Fresh Migration (DANGER - Dev Only!)

**⚠️ WARNING: This deletes ALL data in the database!**

```bash
# ONLY in development
./app.sh dev fresh-migrate
```

**When to use:**
- Starting fresh development
- Testing schema changes
- Resetting development database

**NEVER use in production!**

### Migration Best Practices

1. **Backup First**
   ```bash
   # Production database backup
   pg_dump -U postgres -d viskatera_db > backup_$(date +%Y%m%d_%H%M%S).sql
   ```

2. **Test Migrations**
   - Always test in development/staging first
   - Verify data integrity after migration

3. **Migration Order**
   - Run migrations in production during low-traffic periods
   - Have rollback plan ready

4. **Version Control**
   - Keep migrations in version control
   - Document breaking changes

---

## ⚙️ Environment Configuration

### Environment Variables Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_HOST` | Database host | `localhost` | Yes |
| `DB_PORT` | Database port | `5432` | Yes |
| `DB_USER` | Database user | - | Yes |
| `DB_PASSWORD` | Database password | - | Yes |
| `DB_NAME` | Database name | `viskatera_db` | Yes |
| `JWT_SECRET` | JWT signing secret | - | Yes |
| `PORT` | Server port | `8080` | No |
| `GIN_MODE` | Gin mode (debug/release) | `debug` | No |
| `SMTP_HOST` | SMTP server (empty = MailHog) | - | No* |
| `SMTP_PORT` | SMTP port | `1025` | No* |
| `SMTP_USER` | SMTP username | - | No* |
| `SMTP_PASS` | SMTP password | - | No* |
| `SMTP_FROM` | Email sender | `noreply@viskatera.com` | No |
| `APP_BASE_URL` | Application base URL | `http://localhost:8080` | No |

*Required for production, optional for development (uses MailHog)

### Environment-Specific Files

Create different `.env` files for different environments:

```bash
.env.development  # Local development
.env.staging      # Staging environment
.env.production   # Production environment
```

Then use:
```bash
cp .env.development .env  # For dev
cp .env.production .env    # For prod
```

---

## 🔧 Troubleshooting

### Application Won't Start

1. **Check logs**
   ```bash
   tail -f app.log  # Production
   # Or check console output in dev
   ```

2. **Check database connection**
   ```bash
   # Dev: Check Docker
   docker-compose ps postgres
   docker-compose logs postgres
   
   # Prod: Test connection
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME
   ```

3. **Check port availability**
   ```bash
   lsof -i :8080  # Check if port is in use
   ```

4. **Verify environment variables**
   ```bash
   ./app.sh status  # Shows configuration
   ```

### Database Connection Issues

1. **Dev Mode:**
   ```bash
   # Restart database
   docker-compose restart postgres
   
   # Check database logs
   docker-compose logs postgres
   ```

2. **Production:**
   - Verify database credentials
   - Check firewall rules
   - Verify database is accessible from app server
   - Check database logs

### Migration Issues

1. **Migration Fails**
   - Check database connection
   - Verify database user has proper permissions
   - Check migration logs

2. **Fresh Migration Issues**
   - Ensure you're in dev mode
   - Check for foreign key constraints
   - Verify all services are stopped

### Email Not Sending

1. **Development:**
   - Check MailHog is running: `docker ps | grep mailhog`
   - Access MailHog UI: http://localhost:8025
   - Verify SMTP_HOST is empty in `.env`

2. **Production:**
   - Verify SMTP credentials
   - Test SMTP connection
   - Check firewall rules
   - Check application logs

---

## 📚 Additional Resources

- [Go Best Practices](https://golang.org/doc/effective_go)
- [GORM Documentation](https://gorm.io/docs/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

---

## 🆘 Support

For issues or questions:
1. Check logs: `tail -f app.log`
2. Verify configuration: `./app.sh status`
3. Check documentation: `./app.sh help`
4. Review troubleshooting section above

