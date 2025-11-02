# Getting Started with Viskatera API

## 🚀 Quick Start (Recommended)

### One-Command Setup
```bash
./app.sh dev start
```

This will:
- ✅ Check prerequisites (Go, Docker)
- ✅ Create `.env` file if missing
- ✅ Start PostgreSQL database with Docker
- ✅ Install Go dependencies
- ✅ Start the API server with live reload

### Alternative: Using Makefile
```bash
make dev    # Start in development mode
make start  # Same as above
```

---

## 🛠️ Manual Setup

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 12 or higher
- Docker (optional, for easy database setup)

### 1. Install Dependencies
```bash
go mod tidy
```

### 2. Setup Database

#### Using Docker (Recommended)
```bash
# Start PostgreSQL
docker-compose up -d postgres

# Wait for database to be ready
sleep 10
```

#### Using Local PostgreSQL
```sql
-- Create database
CREATE DATABASE viskatera_db;

-- Create user (optional)
CREATE USER viskatera_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE viskatera_db TO viskatera_user;
```

### 3. Configure Environment
```bash
# Copy environment file
cp env.example .env

# Edit .env file with your database credentials
nano .env
```

### 4. Seed Database
```bash
go run scripts/seed_data.go
```

### 5. Start Application
```bash
go run main.go
```

---

## 🧪 Testing the API

### Using the Test Script
```bash
./test_api.sh
```

### Using curl
```bash
# Health check
curl http://localhost:8080/health

# Register user
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"Test User"}'

# Login
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Get visas
curl http://localhost:8080/api/v1/visas
```

### Using Postman
1. Import the API collection
2. Set base URL to `http://localhost:8080`
3. Run the collection

---

## 📊 Access Points

- **API Server**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **Database Admin**: http://localhost:8081 (if using Docker)

---

## 🔧 Troubleshooting

### Database Connection Issues
```bash
# Check if PostgreSQL is running
docker-compose ps

# Check database logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres
```

### Port Already in Use
```bash
# Change port in .env file
PORT=8081
```

### JWT Token Issues
1. Check JWT_SECRET in `.env` file
2. Ensure token is included in Authorization header
3. Format: `Authorization: Bearer <token>`

### Go Module Issues
```bash
# Clean module cache
go clean -modcache

# Reinstall dependencies
go mod tidy
```

---

## 📁 Project Structure

```
viskatera-api-go/
├── config/          # Database configuration
├── controllers/     # API controllers
├── middleware/      # Authentication middleware
├── models/          # Database models
├── routes/          # API routes
├── scripts/         # Utility scripts
├── utils/           # Helper functions
├── main.go          # Application entry point
├── go.mod           # Go modules
├── docker-compose.yml # Database setup
├── Makefile         # Build commands
└── README.md        # Documentation
```

---

## 🎯 Available Commands

### Using app.sh (Recommended)
```bash
./app.sh help          # Show help message
./app.sh dev start     # Start in development mode
./app.sh dev stop      # Stop application
./app.sh dev restart   # Restart application
./app.sh dev migrate   # Run database migration
./app.sh dev fresh-migrate  # Fresh migration (dev only, deletes all data)
./app.sh dev seed      # Seed database with sample data
./app.sh dev admin     # Create admin user
./app.sh dev status    # Show application status

./app.sh prod start    # Start in production mode
./app.sh prod build    # Build production binary
./app.sh prod migrate  # Run migration in production
```

### Using Makefile (Aliases)
```bash
make help          # Show available commands
make dev           # Start in development mode
make start         # Start application (dev mode)
make stop          # Stop application
make restart       # Restart application
make migrate       # Run database migration
make fresh-migrate # Fresh migration (dev only)
make seed          # Seed database
make admin         # Create admin user
make build         # Build production binary
make status        # Show application status
make install       # Install dependencies
make test          # Test API endpoints
make clean         # Clean build artifacts
```

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment guide.

---

## 📚 Documentation

- [API Endpoints](API_ENDPOINTS.md) - Complete API documentation
- [Quick Start](QUICK_START.md) - Quick start guide
- [README](README.md) - Project overview

---

## 🆘 Need Help?

1. Check the logs: `docker-compose logs postgres`
2. Verify database connection: `docker-compose exec postgres pg_isready -U postgres`
3. Check environment variables in `.env` file
4. Ensure all prerequisites are installed

---

## 🎉 Success!

If everything is working correctly, you should see:
- API server running on http://localhost:8080
- Health check returning `{"status":"ok","message":"Viskatera API is running"}`
- Database accessible at http://localhost:8081 (if using Docker)
- Test script completing successfully
