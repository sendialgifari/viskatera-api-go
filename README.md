# Viskatera API Go

Backend API untuk layanan visa dengan PostgreSQL dan JWT authentication.

## Fitur

- **Public APIs** (tanpa login):
  - List visa
  - Detail visa
  - Register user
  - Login user (email/password or OTP)
  - Request OTP for login
  - Verify OTP for login
  - Forgot password (email reset)
  - Reset password with token
  - Google OAuth login

- **Protected APIs** (perlu login):
  - Purchase visa
  - List purchase history
  - Update purchase status

## Setup

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Setup Database

Pastikan PostgreSQL sudah running, lalu buat database:

```sql
CREATE DATABASE viskatera_db;
```

### 3. Environment Variables

Copy file `env.example` menjadi `.env` dan sesuaikan konfigurasi:

```bash
cp env.example .env
```

Edit file `.env`:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=viskatera_db
JWT_SECRET=your-super-secret-jwt-key-here
PORT=8080

# MailHog (Development) - Email will be sent to MailHog at localhost:1025
# MailHog web UI: http://localhost:8025
SMTP_HOST=
SMTP_PORT=1025
SMTP_FROM=noreply@viskatera.com

# For production, configure real SMTP settings:
# SMTP_HOST=smtp.example.com
# SMTP_PORT=587
# SMTP_USER=your-email@example.com
# SMTP_PASS=your-password
```

### 4. Run Application

#### Development Mode (Recommended)
```bash
./app.sh dev start
```

Atau menggunakan Makefile:
```bash
make dev
```

#### Production Mode
```bash
./app.sh prod build
./app.sh prod start
```

Server akan berjalan di `http://localhost:8080`

Lihat [DEPLOYMENT.md](DEPLOYMENT.md) untuk panduan lengkap deployment.

## API Endpoints

### Public Endpoints

#### Register User
```
POST /api/v1/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

#### Login (Email/Password)
```
POST /api/v1/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

#### Login with OTP
Request OTP:
```
POST /api/v1/auth/request-otp
Content-Type: application/json

{
  "email": "user@example.com"
}
```

Verify OTP:
```
POST /api/v1/auth/verify-otp
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "123456"
}
```

#### Forgot Password
```
POST /api/v1/auth/forgot-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

#### Reset Password
```
POST /api/v1/auth/reset-password
Content-Type: application/json

{
  "token": "reset_token_from_email",
  "new_password": "newpassword123"
}
```

#### Get All Visas
```
GET /api/v1/visas
GET /api/v1/visas?country=Japan
GET /api/v1/visas?type=Tourist
```

#### Get Visa by ID
```
GET /api/v1/visas/{id}
```

### Protected Endpoints (Require JWT Token)

#### Purchase Visa
```
POST /api/v1/purchases
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "visa_id": 1,
  "visa_option_id": 1  // optional
}
```

#### Get User Purchases
```
GET /api/v1/purchases
Authorization: Bearer {jwt_token}
```

#### Get Purchase by ID
```
GET /api/v1/purchases/{id}
Authorization: Bearer {jwt_token}
```

#### Update Purchase Status
```
PUT /api/v1/purchases/{id}/status
Authorization: Bearer {jwt_token}
Content-Type: application/json

{
  "status": "completed"  // pending, completed, cancelled
}
```

### Admin Endpoints (for testing)

#### Create Visa
```
POST /api/v1/admin/visas
Content-Type: application/json

{
  "country": "Japan",
  "type": "Tourist",
  "description": "Tourist visa for Japan",
  "price": 500000,
  "duration": 30
}
```

#### Update Visa
```
PUT /api/v1/admin/visas/{id}
Content-Type: application/json

{
  "country": "Japan",
  "type": "Tourist",
  "description": "Updated description",
  "price": 600000,
  "duration": 30,
  "is_active": true
}
```

#### Delete Visa
```
DELETE /api/v1/admin/visas/{id}
```

## Testing dengan Postman/curl

### 1. Register User
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 3. Create Visa (Admin)
```bash
curl -X POST http://localhost:8080/api/v1/admin/visas \
  -H "Content-Type: application/json" \
  -d '{
    "country": "Japan",
    "type": "Tourist",
    "description": "Tourist visa for Japan",
    "price": 500000,
    "duration": 30
  }'
```

### 4. Get Visas
```bash
curl -X GET http://localhost:8080/api/v1/visas
```

### 5. Purchase Visa (with JWT token)
```bash
curl -X POST http://localhost:8080/api/v1/purchases \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "visa_id": 1
  }'
```

## Database Schema

- **users**: User data
- **visas**: Visa information
- **visa_options**: Additional visa options
- **visa_purchases**: Purchase records

## Available Commands

### Using app.sh (Recommended)
```bash
./app.sh help          # Show help
./app.sh dev start     # Start in development mode
./app.sh dev stop      # Stop application
./app.sh dev restart   # Restart application
./app.sh dev migrate   # Run database migration
./app.sh dev fresh-migrate  # Fresh migration (dev only)
./app.sh dev seed      # Seed database
./app.sh prod start    # Start in production mode
```

### Using Makefile
```bash
make help    # Show available commands
make dev     # Start in development mode
make start   # Start application
make stop    # Stop application
make migrate # Run migration
```

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment guide.

## Health Check

```
GET /health
```

Response:
```json
{
  "status": "ok",
  "message": "Viskatera API is running"
}
```
