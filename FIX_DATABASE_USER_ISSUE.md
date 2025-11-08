# Fix Database Error: "database viskatera_user does not exist"

## Root Cause

Error ini terjadi karena:

1. **Healthcheck PostgreSQL** menggunakan `pg_isready` tanpa specify database name (`-d`)
   - Ketika tidak ada `-d`, PostgreSQL akan mencoba connect ke database default
   - Database default untuk user biasanya adalah database dengan nama yang sama dengan user
   - Jika database tersebut tidak ada, akan muncul error

2. **Volume PostgreSQL sudah ada** dengan data lama menggunakan user berbeda
   - Jika volume sudah ada dengan user `postgres`, dan sekarang menggunakan `DB_USER=viskatera_user`
   - User `viskatera_user` mungkin belum ada di database yang sudah ada

## Solution

### 1. Fix Healthcheck (Already Fixed)

Healthcheck di `docker-compose.prod.yml` sudah diperbaiki untuk menggunakan database name:

```yaml
healthcheck:
  test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-postgres} -d ${DB_NAME:-viskatera_db}"]
```

### 2. Fix Database User (Jika Volume Sudah Ada)

Jika volume PostgreSQL sudah ada dengan user berbeda, Anda perlu membuat user baru:

```bash
# Connect ke PostgreSQL container sebagai superuser postgres
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres

# Di dalam psql, jalankan:
CREATE USER viskatera_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE viskatera_db TO viskatera_user;
ALTER DATABASE viskatera_db OWNER TO viskatera_user;

# Connect ke database
\c viskatera_db

# Grant schema privileges
GRANT ALL ON SCHEMA public TO viskatera_user;
ALTER SCHEMA public OWNER TO viskatera_user;

# Exit
\q
```

### 3. Reset Database (HATI-HATI: Akan Menghapus Semua Data!)

Jika data tidak penting dan ingin mulai dari awal:

```bash
# Stop semua services
docker-compose -f docker-compose.prod.yml down

# Hapus volume PostgreSQL
docker volume rm viskatera-be-go_postgres_data

# Atau jika menggunakan nama volume yang berbeda:
docker-compose -f docker-compose.prod.yml down -v

# Start ulang (PostgreSQL akan membuat user dan database baru)
docker-compose -f docker-compose.prod.yml up -d
```

## Verifikasi

Setelah memperbaiki, verifikasi konfigurasi:

```bash
# Check user exists
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -c "\du"

# Check database exists
docker-compose -f docker-compose.prod.yml exec postgres psql -U postgres -c "\l"

# Test healthcheck
docker-compose -f docker-compose.prod.yml exec postgres pg_isready -U viskatera_user -d viskatera_db

# Test connection dari API container
docker-compose -f docker-compose.prod.yml exec postgres psql -U viskatera_user -d viskatera_db -c "SELECT version();"

# Check logs
docker-compose -f docker-compose.prod.yml logs -f postgres
```

## Konfigurasi .env yang Benar

Pastikan file `.env` production Anda memiliki konfigurasi seperti ini:

```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=viskatera_user
DB_PASSWORD=your_secure_password
DB_NAME=viskatera_db
```

**Penting:**
- `DB_USER` = nama user PostgreSQL (bukan nama database!)
- `DB_NAME` = nama database PostgreSQL
- `DB_PORT` = 5432 untuk koneksi internal container

## Troubleshooting

### Error: "role viskatera_user does not exist"
- User belum dibuat di database
- Jalankan perintah CREATE USER di atas

### Error: "database viskatera_db does not exist"
- Database belum dibuat
- PostgreSQL container seharusnya membuat database otomatis dari `POSTGRES_DB` di docker-compose
- Jika tidak, buat manual: `CREATE DATABASE viskatera_db;`

### Error: "password authentication failed"
- Password di `.env` tidak sesuai dengan password di database
- Update password: `ALTER USER viskatera_user WITH PASSWORD 'new_password';`
- Atau update `.env` dengan password yang benar

### Healthcheck selalu fail
- Pastikan healthcheck menggunakan `-d ${DB_NAME}` seperti yang sudah diperbaiki
- Check apakah user memiliki akses ke database
- Check apakah database sudah dibuat

## Prevention

Untuk menghindari masalah ini di masa depan:

1. **Selalu specify database name** saat menggunakan `pg_isready` atau `psql`
2. **Gunakan superuser `postgres`** untuk healthcheck jika memungkinkan (lebih reliable)
3. **Pastikan volume PostgreSQL dihapus** jika ingin menggunakan user baru
4. **Dokumentasikan** perubahan user database di changelog

