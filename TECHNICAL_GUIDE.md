# Viskatera API - Technical Development Guide

Panduan lengkap dan mendalam untuk membangun aplikasi Viskatera API dari awal hingga production-ready.

---

## Table of Contents

1. [Overview & Architecture](#overview--architecture)
2. [Project Setup](#project-setup)
3. [Database Design](#database-design)
4. [Authentication & Authorization](#authentication--authorization)
5. [API Development](#api-development)
6. [Performance Optimization](#performance-optimization)
7. [Security Implementation](#security-implementation)
8. [Testing & Deployment](#testing--deployment)

---

## Overview & Architecture

### Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: Gin (github.com/gin-gonic/gin)
- **ORM**: GORM (gorm.io/gorm)
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Message Queue**: RabbitMQ 3.12 (github.com/rabbitmq/amqp091-go)
- **Authentication**: JWT (JSON Web Tokens)
- **Payment Gateway**: Xendit
- **PDF Generation**: gofpdf (github.com/jung-kurt/gofpdf)
- **Documentation**: Swagger/OpenAPI

### Application Architecture

```
┌─────────────────┐
│   Client App    │
└────────┬────────┘
         │ HTTPS
         ▼
┌─────────────────┐
│  Nginx (VPS)    │ ← SSL Termination, Rate Limiting
└────────┬────────┘
         │ HTTP
         ▼
┌─────────────────┐
│  Go API (8080)  │ ← Business Logic
└────────┬────────┘
         │
    ┌────┴────┐
    ▼        ▼
┌────────┐ ┌────────┐
│PostgreSQL│ │ Redis  │
│  (5432) │ │ (6379) │
└────────┘ └────────┘
```

### Project Structure

```
viskatera-api-go/
├── main.go                 # Application entry point
├── config/                 # Configuration modules
│   ├── database.go        # Database connection & pooling
│   └── cache.go           # Redis cache utilities
├── models/                 # Database models
│   ├── user.go            # User model & DTOs
│   ├── visa.go            # Visa, VisaOption, VisaPurchase
│   ├── payment.go         # Payment model
│   ├── auth.go            # OTP, PasswordReset models
│   ├── activity.go        # ActivityLog model
│   └── response.go        # API response models
├── controllers/            # Request handlers
│   ├── auth.go            # Authentication endpoints
│   ├── otp.go             # OTP login endpoints
│   ├── visa.go            # Visa management
│   ├── purchase.go        # Purchase management
│   ├── payment.go         # Payment processing
│   ├── upload.go          # File upload handlers
│   ├── export.go          # Excel/PDF export
│   └── activity.go        # Activity logging
├── middleware/             # HTTP middleware
│   ├── auth.go            # JWT authentication
│   └── admin.go           # Admin authorization
├── routes/                 # Route definitions
│   └── routes.go          # API routing
├── utils/                  # Utility functions
│   ├── jwt.go             # JWT token generation
│   ├── password.go        # Password hashing
│   ├── email.go           # Email sending with attachment
│   ├── pdf.go             # PDF invoice generation
│   ├── image.go           # Image processing
│   ├── token.go           # Token generation
│   └── activity_logger.go # Activity logging
├── workers/                # Background workers
│   ├── worker.go          # Worker initialization
│   └── email_worker.go    # Email and PDF worker
├── scripts/                # Utility scripts
│   ├── migrate.go         # Database migration
│   ├── seed_data.go       # Seed sample data
│   └── create_admin.go    # Create admin user
├── docs/                   # Swagger documentation
├── uploads/                # Uploaded files
├── docker-compose.yml      # Development Docker setup
├── docker-compose.prod.yml # Production Docker setup
├── Dockerfile              # Production Docker image
└── nginx.conf              # Nginx configuration
```

---

## Project Setup

### Step 1: Initialize Go Project

```bash
# Create project directory
mkdir viskatera-api-go
cd viskatera-api-go

# Initialize Go module
go mod init viskatera-api-go

# Create directory structure
mkdir -p config controllers middleware models routes utils scripts docs uploads/{avatars,visas}
```

### Step 2: Install Dependencies

```bash
# Core dependencies
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/golang-jwt/jwt/v5
go get github.com/redis/go-redis/v9
go get github.com/joho/godotenv
go get golang.org/x/crypto

# Documentation
go get github.com/swaggo/swag/cmd/swag
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files

# Utilities
go get github.com/nfnt/resize
go get github.com/jung-kurt/gofpdf
go get github.com/xuri/excelize/v2

# Message Queue
go get github.com/rabbitmq/amqp091-go
```

### Step 3: Create Environment Configuration

Create `env.example`:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=viskatera_db
DB_SSLMODE=disable

# Redis Cache Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
CACHE_ENABLED=true
CACHE_TTL=300

# Application Configuration
ENVIRONMENT=development
GIN_MODE=debug
PORT=8080

# Security
JWT_SECRET=your-super-secret-jwt-key-here

# External Services
XENDIT_SECRET_KEY=xnd_secret_development_xxxxxxxxxxxxx
XENDIT_PUBLIC_KEY=xnd_public_development_xxxxxxxxxxxxx
XENDIT_API_URL=https://api.xendit.co

# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/google/callback

# Application URLs
APP_BASE_URL=http://localhost:8080

# SMTP Configuration
SMTP_HOST=
SMTP_PORT=1025
SMTP_USER=
SMTP_PASS=
SMTP_FROM=noreply@viskatera.com

# File Upload
UPLOAD_DIR=./uploads
MAX_UPLOAD_SIZE=10485760
```

### Step 4: Setup Database Connection

Create `config/database.go`:

```go
package config

import (
    "fmt"
    "log"
    "os"
    "time"
    "viskatera-api-go/models"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
    var err error
    
    sslMode := os.Getenv("DB_SSLMODE")
    if sslMode == "" {
        sslMode = "disable"
    }
    
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"),
        os.Getenv("DB_PORT"),
        sslMode,
    )
    
    logLevel := logger.Info
    if os.Getenv("GIN_MODE") == "release" {
        logLevel = logger.Error
    }
    
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logLevel),
        NowFunc: func() time.Time {
            return time.Now().Local()
        },
    })
    
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // Connection pool configuration
    sqlDB, err := DB.DB()
    if err != nil {
        log.Fatal("Failed to get database instance:", err)
    }
    
    sqlDB.SetMaxIdleConns(25)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    sqlDB.SetConnMaxIdleTime(10 * time.Minute)
    
    log.Println("Database connected successfully!")
}

func MigrateDB() {
    err := DB.AutoMigrate(
        &models.User{},
        &models.Visa{},
        &models.VisaOption{},
        &models.VisaPurchase{},
        &models.PasswordResetToken{},
        &models.OTP{},
        &models.Payment{},
        &models.ActivityLog{},
    )
    
    if err != nil {
        log.Fatal("Failed to migrate database:", err)
    }
    
    log.Println("Database migration completed!")
}

func CloseDB() error {
    if DB != nil {
        sqlDB, err := DB.DB()
        if err != nil {
            return err
        }
        return sqlDB.Close()
    }
    return nil
}
```

### Step 5: Create Main Application File

Create `main.go`:

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    "viskatera-api-go/config"
    "viskatera-api-go/routes"
    
    _ "viskatera-api-go/docs"
    
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using system environment variables")
    }
    
    // Set Gin mode
    if os.Getenv("GIN_MODE") == "" {
        env := os.Getenv("ENVIRONMENT")
        if env == "production" || env == "prod" {
            gin.SetMode(gin.ReleaseMode)
        }
    } else {
        gin.SetMode(os.Getenv("GIN_MODE"))
    }
    
    // Connect to database
    config.ConnectDB()
    config.MigrateDB()
    
    // Connect to Redis cache
    config.ConnectRedis()
    
    // Setup routes
    r := routes.SetupRoutes()
    
    // Get port
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    // Create HTTP server
    srv := &http.Server{
        Addr:         ":" + port,
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
    
    // Start server
    go func() {
        log.Printf("Server starting on port %s (mode: %s)", port, gin.Mode())
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    // Close connections
    config.CloseDB()
    config.CloseRedis()
    
    log.Println("Server exited gracefully")
}
```

---

## Database Design

### Entity Relationship Diagram

```
┌──────────┐      ┌──────────────┐      ┌──────────┐
│   User   │──────│VisaPurchase  │──────│  Visa    │
└──────────┘      └──────────────┘      └──────────┘
     │                    │                    │
     │                    │                    │
     │                    │              ┌─────────────┐
     │                    │              │ VisaOption │
     │                    │              └─────────────┘
     │                    │
     │                    │
┌──────────┐      ┌──────────────┐
│ Payment  │──────│VisaPurchase  │
└──────────┘      └──────────────┘
     │
     │
┌──────────────┐
│ ActivityLog │
└──────────────┘
```

### Model Definitions

#### User Model

```go
type User struct {
    ID          uint           `json:"id" gorm:"primaryKey"`
    Email       string         `json:"email" gorm:"unique;not null;index:idx_user_email_active"`
    Password    string         `json:"-" gorm:"not null"`
    Name        string         `json:"name" gorm:"not null"`
    AvatarURL   string         `json:"avatar_url"`
    GoogleID    string         `json:"google_id" gorm:"index:idx_user_google,unique"`
    Role        UserRole       `json:"role" gorm:"type:varchar(20);default:'customer'"`
    IsActive    bool           `json:"is_active" gorm:"default:true"`
    LastLoginAt *time.Time     `json:"last_login_at"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}
```

**Key Features:**
- Soft delete dengan `gorm.DeletedAt`
- Indexes untuk performa query
- Role-based access control (customer/admin)
- Google OAuth support

#### Visa Model

```go
type Visa struct {
    ID              uint           `json:"id" gorm:"primaryKey"`
    Country         string         `json:"country" gorm:"not null;index:idx_visa_country_active"`
    Type            string         `json:"type" gorm:"not null;index:idx_visa_type_active"`
    Description     string         `json:"description"`
    Price           float64        `json:"price" gorm:"not null;index:idx_visa_price"`
    Duration        int            `json:"duration" gorm:"not null"`
    VisaDocumentURL string         `json:"visa_document_url"`
    IsActive        bool           `json:"is_active" gorm:"default:true"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
    DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}
```

#### VisaPurchase Model

```go
type VisaPurchase struct {
    ID           uint           `json:"id" gorm:"primaryKey"`
    UserID       uint           `json:"user_id" gorm:"not null"`
    User         User           `json:"user" gorm:"foreignKey:UserID"`
    VisaID       uint           `json:"visa_id" gorm:"not null"`
    Visa         Visa           `json:"visa" gorm:"foreignKey:VisaID"`
    VisaOptionID *uint          `json:"visa_option_id"`
    VisaOption   *VisaOption    `json:"visa_option" gorm:"foreignKey:VisaOptionID"`
    TotalPrice   float64        `json:"total_price" gorm:"not null"`
    Status       string         `json:"status" gorm:"default:'pending'"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}
```

### Database Indexes Strategy

Indexes dibuat untuk optimasi query:

1. **Composite Indexes**: Untuk query dengan multiple WHERE conditions
   - `idx_user_email_active`: Email + IsActive
   - `idx_visa_country_active`: Country + IsActive
   - `idx_purchase_user_status`: UserID + Status

2. **Foreign Key Indexes**: Untuk JOIN operations
   - Semua foreign keys otomatis di-index oleh GORM

3. **Date Indexes**: Untuk time-based queries
   - `idx_user_created`: CreatedAt
   - `idx_purchase_status_created`: Status + CreatedAt

---

## Authentication & Authorization

### JWT Implementation

#### Token Generation (`utils/jwt.go`)

```go
package utils

import (
    "time"
    "os"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID uint   `json:"user_id"`
    Email  string `json:"email"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(userID uint, email, role string) (string, error) {
    expiry := 24 * time.Hour
    if exp := os.Getenv("JWT_EXPIRY"); exp != "" {
        if d, err := time.ParseDuration(exp); err == nil {
            expiry = d
        }
    }
    
    claims := &Claims{
        UserID: userID,
        Email:  email,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "viskatera-api",
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("JWT_SECRET")), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, jwt.ErrSignatureInvalid
}
```

#### Authentication Middleware (`middleware/auth.go`)

```go
package middleware

import (
    "net/http"
    "strings"
    "viskatera-api-go/models"
    "viskatera-api-go/utils"
    
    "github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, models.ErrorResponse("Authorization header required", nil))
            c.Abort()
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            c.JSON(http.StatusUnauthorized, models.ErrorResponse("Invalid authorization format", nil))
            c.Abort()
            return
        }
        
        claims, err := utils.ValidateToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, models.ErrorResponse("Invalid or expired token", nil))
            c.Abort()
            return
        }
        
        // Set user info in context
        c.Set("user_id", claims.UserID)
        c.Set("user_email", claims.Email)
        c.Set("user_role", claims.Role)
        
        c.Next()
    }
}
```

#### Admin Middleware (`middleware/admin.go`)

```go
package middleware

import (
    "net/http"
    "viskatera-api-go/models"
    
    "github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("user_role")
        if !exists || role != "admin" {
            c.JSON(http.StatusForbidden, models.ErrorResponse("Admin access required", nil))
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Password Hashing

#### Implementation (`utils/password.go`)

```go
package utils

import (
    "golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

---

## API Development

### Route Structure (`routes/routes.go`)

```go
func SetupRoutes() *gin.Engine {
    r := gin.Default()
    
    // CORS middleware
    r.Use(corsMiddleware())
    
    // Health check
    r.GET("/health", healthCheck)
    
    // Swagger documentation
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    
    // Public routes
    public := r.Group("/api/v1")
    {
        public.POST("/register", controllers.Register)
        public.POST("/login", controllers.Login)
        public.POST("/auth/request-otp", controllers.RequestOTP)
        public.POST("/auth/verify-otp", controllers.VerifyOTP)
        public.GET("/visas", controllers.GetVisas)
        public.GET("/visas/:id", controllers.GetVisaByID)
    }
    
    // Static files
    r.Static("/uploads", "./uploads")
    
    // Protected routes
    protected := r.Group("/api/v1")
    protected.Use(middleware.AuthMiddleware())
    {
        protected.POST("/purchases", controllers.PurchaseVisa)
        protected.GET("/purchases", controllers.GetUserPurchases)
        // ... more routes
    }
    
    // Admin routes
    admin := r.Group("/api/v1/admin")
    admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
    {
        admin.POST("/visas", controllers.CreateVisa)
        admin.PUT("/visas/:id", controllers.UpdateVisa)
        admin.DELETE("/visas/:id", controllers.DeleteVisa)
    }
    
    return r
}
```

### Controller Pattern

#### Example: Visa Controller (`controllers/visa.go`)

```go
func GetVisas(c *gin.Context) {
    var visas []models.Visa
    var total int64
    
    // Query parameters
    country := c.Query("country")
    visaType := c.Query("type")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
    
    // Build query
    query := config.DB.Model(&models.Visa{}).Where("is_active = ?", true)
    
    if country != "" {
        query = query.Where("country = ?", country)
    }
    if visaType != "" {
        query = query.Where("type = ?", visaType)
    }
    
    // Get total count
    query.Count(&total)
    
    // Pagination
    offset := (page - 1) * perPage
    if perPage > 100 {
        perPage = 100
    }
    
    // Execute query
    if err := query.Order("created_at DESC").Offset(offset).Limit(perPage).Find(&visas).Error; err != nil {
        c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to retrieve visas", nil))
        return
    }
    
    // Response with pagination metadata
    c.JSON(http.StatusOK, models.SuccessResponseWithMeta(
        "Visas retrieved successfully",
        visas,
        models.PaginationMeta{
            Page:       page,
            PerPage:    perPage,
            Total:      int(total),
            TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
        },
    ))
}
```

### Response Models (`models/response.go`)

```go
type APIResponse struct {
    Success   bool        `json:"success"`
    Message   string      `json:"message"`
    Data      interface{} `json:"data,omitempty"`
    Meta      interface{} `json:"meta,omitempty"`
    Error     interface{} `json:"error,omitempty"`
    Timestamp string      `json:"timestamp"`
}

func SuccessResponse(message string, data interface{}) APIResponse {
    return APIResponse{
        Success:   true,
        Message:   message,
        Data:      data,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }
}

func ErrorResponse(message string, err interface{}) APIResponse {
    return APIResponse{
        Success:   false,
        Message:   message,
        Error:     err,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
    }
}
```

---

## Performance Optimization

### Redis Caching

#### Cache Implementation (`config/cache.go`)

```go
func CacheGet(ctx context.Context, key string, dest interface{}) error {
    if !CacheEnabled || RedisClient == nil {
        return redis.Nil
    }
    
    val, err := RedisClient.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    
    return json.Unmarshal([]byte(val), dest)
}

func CacheSet(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    if !CacheEnabled || RedisClient == nil {
        return nil
    }
    
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return RedisClient.Set(ctx, key, data, ttl).Err()
}
```

#### Using Cache in Controllers

```go
func GetVisaByID(c *gin.Context) {
    id := c.Param("id")
    cacheKey := fmt.Sprintf("visa:%s", id)
    
    var visa models.Visa
    
    // Try to get from cache
    if err := config.CacheGet(c.Request.Context(), cacheKey, &visa); err == nil {
        c.JSON(http.StatusOK, models.SuccessResponse("Visa retrieved from cache", visa))
        return
    }
    
    // Get from database
    if err := config.DB.Preload("VisaOptions").First(&visa, id).Error; err != nil {
        c.JSON(http.StatusNotFound, models.ErrorResponse("Visa not found", nil))
        return
    }
    
    // Store in cache
    ttl := config.GetCacheTTL()
    config.CacheSet(c.Request.Context(), cacheKey, visa, ttl)
    
    c.JSON(http.StatusOK, models.SuccessResponse("Visa retrieved successfully", visa))
}
```

### Database Query Optimization

1. **Use Preload for Relationships**
```go
config.DB.Preload("VisaOptions").Preload("User").Find(&purchases)
```

2. **Select Specific Fields**
```go
config.DB.Select("id", "email", "name").Find(&users)
```

3. **Use Indexes Effectively**
```go
// Query uses index idx_visa_country_active
config.DB.Where("country = ? AND is_active = ?", country, true).Find(&visas)
```

4. **Pagination for Large Datasets**
```go
offset := (page - 1) * perPage
config.DB.Offset(offset).Limit(perPage).Find(&results)
```

---

## Security Implementation

### Input Validation

Using Gin's binding:

```go
type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    Name     string `json:"name" binding:"required,min=2"`
}
```

### Password Security

- Use bcrypt with cost 10 (default)
- Never store plain text passwords
- Enforce minimum password length (6 characters)

### JWT Security

- Use strong secret key (64+ characters)
- Set appropriate expiration time (24 hours)
- Validate token on every request
- Include user role in token for authorization

### SQL Injection Prevention

GORM automatically uses parameterized queries:

```go
// Safe - uses parameterized query
config.DB.Where("email = ?", email).First(&user)

// Unsafe - don't do this!
// config.DB.Where(fmt.Sprintf("email = '%s'", email)).First(&user)
```

### File Upload Security

```go
// Validate file type
allowedTypes := []string{"image/jpeg", "image/png", "image/jpg"}
if !contains(allowedTypes, fileHeader.Header.Get("Content-Type")) {
    return errors.New("invalid file type")
}

// Validate file size
if fileHeader.Size > maxSize {
    return errors.New("file too large")
}

// Sanitize filename
filename := sanitizeFilename(fileHeader.Filename)
```

---

## Testing & Deployment

### Development Testing

```bash
# Start development environment
./app.sh dev start

# Run migrations
./app.sh dev migrate

# Seed data
./app.sh dev seed

# Test API
curl http://localhost:8080/health
```

### Production Deployment

1. **Build Docker Image**
```bash
docker-compose -f docker-compose.prod.yml build
```

2. **Start Services**
```bash
docker-compose -f docker-compose.prod.yml up -d
```

3. **Setup Nginx**
```bash
sudo cp nginx.conf /etc/nginx/sites-available/viskatera-api
sudo ln -s /etc/nginx/sites-available/viskatera-api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

4. **Setup SSL**
```bash
sudo certbot --nginx -d api.ahmadcorp.com
```

### Monitoring

- Health check endpoint: `/health`
- Application logs: `docker-compose logs -f api`
- Database logs: `docker-compose logs -f postgres`
- Redis stats: `docker-compose exec redis redis-cli INFO stats`

---

## Best Practices Summary

1. **Code Organization**
   - Separate concerns (models, controllers, middleware)
   - Use dependency injection
   - Keep functions small and focused

2. **Error Handling**
   - Always handle errors explicitly
   - Return meaningful error messages
   - Log errors for debugging

3. **Database**
   - Use transactions for multi-step operations
   - Index frequently queried fields
   - Use connection pooling

4. **Security**
   - Validate all inputs
   - Use parameterized queries
   - Hash passwords with bcrypt
   - Use HTTPS in production

5. **Performance**
   - Cache frequently accessed data
   - Use pagination for large datasets
   - Optimize database queries
   - Monitor and profile application

6. **Documentation**
   - Document API with Swagger
   - Write clear code comments
   - Maintain README and guides

---

## RabbitMQ Integration

### Setup RabbitMQ Connection

Create `config/queue.go`:

```go
package config

import (
    "fmt"
    "log"
    "os"
    amqp "github.com/rabbitmq/amqp091-go"
)

var RabbitMQConn *amqp.Connection
var RabbitMQChan *amqp.Channel

func ConnectRabbitMQ() error {
    host := os.Getenv("RABBITMQ_HOST")
    port := os.Getenv("RABBITMQ_PORT")
    user := os.Getenv("RABBITMQ_USER")
    pass := os.Getenv("RABBITMQ_PASS")
    
    amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)
    
    conn, err := amqp.Dial(amqpURL)
    if err != nil {
        return err
    }
    
    ch, err := conn.Channel()
    if err != nil {
        return err
    }
    
    RabbitMQConn = conn
    RabbitMQChan = ch
    
    // Declare queues
    DeclareQueues()
    
    return nil
}
```

### Publishing Messages

```go
// Publish email job to queue
job := map[string]interface{}{
    "purchase_id": purchase.ID,
    "user_id":     user.ID,
    "email":       user.Email,
    "type":        "invoice",
}
jobJSON, _ := json.Marshal(job)
config.PublishMessage(config.QueueEmailInvoice, jobJSON)
```

### Worker Implementation

Workers process messages in parallel:

```go
func StartEmailWorker(concurrency int) error {
    // Set QoS for parallel processing
    ch.Qos(concurrency, 0, false)
    
    // Consume messages
    msgs, _ := ch.Consume(queueName, "", false, false, false, false, nil)
    
    // Start multiple workers
    for i := 0; i < concurrency; i++ {
        go func(workerID int) {
            for msg := range msgs {
                processEmail(msg, workerID)
            }
        }(i)
    }
    
    return nil
}
```

### PDF Generation

```go
func GenerateInvoicePDF(data InvoiceData) (string, error) {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    
    // Add content
    pdf.SetFont("Arial", "B", 16)
    pdf.Cell(40, 10, "Invoice")
    
    // ... add invoice details
    
    // Save PDF
    filename := fmt.Sprintf("invoice_%s.pdf", data.InvoiceNumber)
    return pdf.OutputFileAndClose(filename)
}
```

### Email with Attachment

```go
func SendEmailWithAttachment(to, subject, body, pdfPath string) error {
    // Read PDF file
    pdfData, _ := ioutil.ReadFile(pdfPath)
    encoded := base64.StdEncoding.EncodeToString(pdfData)
    
    // Build multipart email
    boundary := "----=_NextPart_"
    msg := buildMultipartEmail(to, subject, body, boundary, encoded)
    
    // Send email
    return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}
```

### Webhook Handler

```go
func XenditWebhook(c *gin.Context) {
    var payload XenditWebhookPayload
    c.ShouldBindJSON(&payload)
    
    // Update payment status
    payment.Status = payload.Status
    
    // If paid, send success email via RabbitMQ
    if payload.Status == "PAID" {
        job := map[string]interface{}{
            "purchase_id": purchase.ID,
            "user_id":     user.ID,
            "email":       user.Email,
            "type":        "payment_success",
        }
        config.PublishMessage(config.QueueEmailPaymentSuccess, jobJSON)
    }
}
```

---

**This guide provides a comprehensive foundation for building and maintaining the Viskatera API. For specific implementation details, refer to the source code and API documentation.**

