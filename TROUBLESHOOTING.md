# Troubleshooting Guide

## Masalah Umum dan Solusi

### 1. Nginx Error: SSL Certificate Tidak Ditemukan

**Error:**
```
nginx: [emerg] cannot load certificate "/etc/letsencrypt/live/api.ahmadcorp.com/fullchain.pem"
nginx exited with code 1
```

**Solusi:**
Konfigurasi nginx sudah diperbaiki untuk bisa berjalan tanpa SSL certificate. Saat ini nginx menggunakan konfigurasi HTTP-only. 

Untuk mengaktifkan HTTPS setelah SSL certificate tersedia:
1. Pastikan SSL certificate sudah ada di `/etc/letsencrypt/live/api.ahmadcorp.com/`
2. Uncomment HTTPS server block di `nginx/conf.d/api.ahmadcorp.com.conf`
3. Comment bagian proxy_pass di HTTP server block dan uncomment redirect ke HTTPS
4. Restart nginx: `docker-compose -f docker-compose.prod.yml restart nginx`

### 2. Nginx Warning: Deprecated http2 Directive

**Warning:**
```
the "listen ... http2" directive is deprecated, use the "http2" directive instead
```

**Solusi:**
Sudah diperbaiki. Menggunakan `http2 on;` directive terpisah sesuai dengan nginx versi terbaru.

### 3. Database Error: "database does not exist"

**Error:**
```
FATAL: database "viskatera_api_go" does not exist
```

**Penjelasan:**
Error ini kemungkinan berasal dari koneksi luar (host machine, tool, atau script lain) yang menggunakan nama database yang salah. 

Aplikasi Go di container sudah berhasil connect dan migrate dengan database yang benar (`viskatera_db`). Error ini biasanya tidak critical dan tidak mempengaruhi operasi aplikasi.

**Solusi:**
- Pastikan file `.env` menggunakan `DB_NAME=viskatera_db`
- Jika ada koneksi dari host machine atau tool lain, pastikan menggunakan nama database yang sama: `viskatera_db`
- Check apakah ada script atau tool yang mencoba connect dengan nama database yang salah

**Verifikasi:**
```bash
# Check database yang ada di container
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -c "\l"

# Check koneksi aplikasi
docker-compose -f docker-compose.prod.yml logs api | grep -i database
```

### 4. Nginx Tidak Start

**Kemungkinan Penyebab:**
1. SSL certificate belum tersedia (sudah diperbaiki)
2. Konfigurasi nginx syntax error
3. Port sudah digunakan

**Solusi:**
```bash
# Test nginx configuration
docker-compose -f docker-compose.prod.yml exec nginx nginx -t

# Check logs
docker-compose -f docker-compose.prod.yml logs nginx

# Restart nginx
docker-compose -f docker-compose.prod.yml restart nginx
```

### 5. Database Connection Error

**Error:**
```
Failed to connect to database
```

**Solusi:**
1. Pastikan postgres container sudah running: `docker-compose -f docker-compose.prod.yml ps`
2. Check environment variables: `docker-compose -f docker-compose.prod.yml exec api env | grep DB_`
3. Pastikan postgres sudah healthy: `docker-compose -f docker-compose.prod.yml ps postgres`
4. Check database logs: `docker-compose -f docker-compose.prod.yml logs postgres`

### 6. Port Already in Use

**Error:**
```
Bind for 0.0.0.0:8080 failed: port is already allocated
```

**Solusi:**
```bash
# Check process yang menggunakan port
sudo lsof -i :8080
# atau
sudo netstat -tulpn | grep 8080

# Stop process atau ubah port di docker-compose.prod.yml
```

## Verifikasi Deployment

### Check Semua Services
```bash
docker-compose -f docker-compose.prod.yml ps
```

### Check Logs
```bash
# All services
docker-compose -f docker-compose.prod.yml logs -f

# Specific service
docker-compose -f docker-compose.prod.yml logs -f api
docker-compose -f docker-compose.prod.yml logs -f nginx
docker-compose -f docker-compose.prod.yml logs -f postgres
```

### Test API
```bash
# Health check
curl http://localhost:8080/health

# Via nginx (setelah setup)
curl http://localhost:8001/health
```

### Check Database
```bash
# Connect to database
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -d viskatera_db

# List tables
\dt

# Exit
\q
```

## Reset Deployment

Jika perlu reset semua:
```bash
# Stop semua services
docker-compose -f docker-compose.prod.yml down

# Hapus volumes (HATI-HATI: akan menghapus data database!)
docker-compose -f docker-compose.prod.yml down -v

# Start ulang
docker-compose -f docker-compose.prod.yml up -d
```

