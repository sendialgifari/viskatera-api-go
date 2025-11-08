# Production Checklist & Best Practices

## ✅ Implemented Improvements

### 1. Database Connection Pooling
- ✅ Configured connection pool dengan MaxIdleConns: 10, MaxOpenConns: 100
- ✅ Connection lifetime di-set ke 1 jam untuk reuse koneksi yang efisien
- ✅ Log level disesuaikan dengan environment (Error di production, Info di development)

### 2. Application Performance
- ✅ HTTP server timeouts (Read: 15s, Write: 15s, Idle: 60s)
- ✅ Graceful shutdown dengan timeout 30 detik
- ✅ Database connection ditutup dengan benar saat shutdown
- ✅ Gin mode di-set ke Release mode di production untuk performa optimal

### 3. Last Login Tracking
- ✅ Field `LastLoginAt` ditambahkan ke model User
- ✅ Update LastLoginAt di semua endpoint login (password, OTP, Google OAuth)

### 4. Security Best Practices
- ✅ JWT authentication implemented
- ✅ Password hashing dengan bcrypt
- ✅ Role-based access control (RBAC)
- ✅ Soft delete untuk data integrity

## 📋 Pre-Production Checklist

### Security
- [ ] **Environment Variables**
  - [ ] Semua secrets (JWT_SECRET, DB_PASSWORD, XENDIT_SECRET_KEY) harus menggunakan strong, unique values
  - [ ] Jangan commit .env file ke repository
  - [ ] Gunakan secrets management (HashiCorp Vault, AWS Secrets Manager, atau env files di server)

- [ ] **Database Security**
  - [ ] Enable SSL/TLS untuk database connection (ubah sslmode=disable menjadi sslmode=require)
  - [ ] Gunakan database user dengan privileges terbatas (bukan superuser)
  - [ ] Backup database rutin dijadwalkan
  - [ ] Database password harus kuat (minimal 16 karakter, kombinasi alphanumeric + special chars)

- [ ] **API Security**
  - [ ] Rate limiting diimplementasikan (recommend: github.com/ulule/limiter/v3)
  - [ ] CORS di-set dengan origin spesifik (bukan "*")
  - [ ] Input validation sudah lengkap di semua endpoint
  - [ ] Sanitize semua user inputs
  - [ ] Implementasi CSRF protection untuk stateful requests

- [ ] **HTTPS/SSL**
  - [ ] Certificate SSL/TLS valid dan auto-renewal setup (Let's Encrypt)
  - [ ] HTTP redirect ke HTTPS
  - [ ] Security headers (HSTS, X-Frame-Options, X-Content-Type-Options, etc.)

### Monitoring & Logging
- [ ] **Application Logging**
  - [ ] Structured logging dengan JSON format (recommend: github.com/sirupsen/logrus atau zap)
  - [ ] Log levels dikonfigurasi dengan benar (INFO, WARN, ERROR)
  - [ ] Log rotation untuk mencegah disk space issues
  - [ ] Sensitive data tidak ada di logs (passwords, tokens, etc.)

- [ ] **Monitoring**
  - [ ] Health check endpoint (`/health`) di-monitor
  - [ ] Database connection health monitoring
  - [ ] Application metrics (response time, request rate, error rate)
  - [ ] Uptime monitoring (UptimeRobot, Pingdom, atau self-hosted)
  - [ ] Error tracking (Sentry, Rollbar, atau similar)

- [ ] **Alerts**
  - [ ] Alert untuk database connection failures
  - [ ] Alert untuk high error rate
  - [ ] Alert untuk high response time
  - [ ] Alert untuk disk space usage
  - [ ] Alert untuk memory/CPU usage

### Performance
- [ ] **Database Optimization**
  - [ ] Indexes sudah optimal (cek slow queries)
  - [ ] Query optimization (gunakan EXPLAIN untuk analyze queries)
  - [ ] Database connection pool sudah diset sesuai dengan beban aplikasi
  - [ ] Consider read replicas jika traffic tinggi

- [ ] **Application Optimization**
  - [ ] Enable Gzip compression (nginx atau middleware)
  - [ ] Static file caching headers
  - [ ] API response caching untuk data yang tidak sering berubah
  - [ ] Pagination untuk list endpoints

- [ ] **Infrastructure**
  - [ ] Load balancer jika multiple instances
  - [ ] CDN untuk static assets jika diperlukan
  - [ ] Auto-scaling setup jika menggunakan cloud

### Backup & Recovery
- [ ] **Database Backup**
  - [ ] Automated daily backups
  - [ ] Backup retention policy (minimal 7 hari, ideal 30 hari)
  - [ ] Backup testing (restore test dilakukan secara rutin)
  - [ ] Offsite backup storage

- [ ] **Application Backup**
  - [ ] Uploaded files backup (avatars, visa documents)
  - [ ] Configuration backup
  - [ ] Disaster recovery plan documented

### Documentation
- [ ] API documentation updated dan accessible
- [ ] Deployment procedures documented
- [ ] Incident response plan documented
- [ ] Runbook untuk common issues

## 🚀 Recommended Additional Improvements

### High Priority
1. **Rate Limiting** - Prevent abuse dan DDoS
   ```go
   // Add to go.mod
   github.com/ulule/limiter/v3 v3.11.2
   ```

2. **Structured Logging** - Better monitoring dan debugging
   ```go
   // Add to go.mod
   github.com/sirupsen/logrus v1.9.3
   // atau
   go.uber.org/zap v1.26.0
   ```

3. **Database SSL** - Secure database connection
   ```go
   // Update config/database.go
   sslmode=require atau sslmode=verify-full
   ```

4. **CORS Configuration** - Specific origins, bukan wildcard
   ```go
   // Update routes/routes.go
   // Replace "*" with specific allowed origins
   ```

### Medium Priority
1. **Request ID Middleware** - Track requests across logs
2. **Metrics Endpoint** - Prometheus metrics endpoint
3. **Graceful Shutdown Timeout** - Sudah diimplementasikan ✅
4. **Request/Response Logging Middleware** - Untuk debugging

### Low Priority
1. **API Versioning** - Untuk future compatibility
2. **GraphQL Support** - Jika diperlukan
3. **WebSocket Support** - Untuk real-time features

## 📊 Monitoring Stack Recommendations

### Option 1: Open Source (Self-hosted)
- **Prometheus** - Metrics collection
- **Grafana** - Visualization & dashboards
- **Loki** - Log aggregation
- **Alertmanager** - Alerting

### Option 2: Cloud Services
- **Datadog** - All-in-one monitoring
- **New Relic** - APM + monitoring
- **Sentry** - Error tracking
- **LogRocket** - Session replay + logs

### Option 3: Minimal Setup
- **UptimeRobot** - Free uptime monitoring
- **Sentry** - Free tier error tracking
- **Application logs** - File-based dengan log rotation

## 🔒 Security Headers Recommendation

Tambahkan middleware untuk security headers:
```go
r.Use(func(c *gin.Context) {
    c.Header("X-Content-Type-Options", "nosniff")
    c.Header("X-Frame-Options", "DENY")
    c.Header("X-XSS-Protection", "1; mode=block")
    c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    c.Header("Content-Security-Policy", "default-src 'self'")
})
```

## 🎯 Performance Targets

- **Response Time**: P95 < 500ms, P99 < 1s
- **Availability**: 99.9% uptime (max 8.76 hours downtime/year)
- **Error Rate**: < 0.1% of total requests
- **Database Connection**: < 80% of max connections under normal load

## 📝 Environment Variables for Production

```env
# Application
ENVIRONMENT=production
GIN_MODE=release
PORT=8080

# Database (with SSL)
DB_HOST=your-db-host
DB_PORT=5433
DB_USER=viskatera_user
DB_PASSWORD=<strong-password>
DB_NAME=viskatera_db
DB_SSLMODE=require

# Security
JWT_SECRET=<generate-strong-secret-64-chars>

# External Services
XENDIT_SECRET_KEY=<production-key>
XENDIT_PUBLIC_KEY=<production-key>
GOOGLE_CLIENT_ID=<production-client-id>
GOOGLE_CLIENT_SECRET=<production-client-secret>

# URLs
APP_BASE_URL=https://api.ahmadcorp.com
GOOGLE_REDIRECT_URL=https://api.ahmadcorp.com/api/v1/auth/google/callback

# SMTP (Production)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=noreply@ahmadcorp.com
SMTP_PASS=<smtp-password>
SMTP_FROM=noreply@ahmadcorp.com
```

