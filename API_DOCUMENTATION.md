# Viskatera API Documentation

> **📚 Interactive API Documentation**: For the most up-to-date and interactive API documentation, visit **[Swagger UI](http://localhost:8080/swagger/index.html)** or **[API Docs](http://localhost:8080/docs)**

## Overview

Viskatera API is a comprehensive visa management system with role-based authentication. It provides endpoints for visa management, user authentication, OTP login, payment processing, and document management.

## Base URL
```
http://localhost:8080
```

## Authentication

The API uses JWT (JSON Web Token) for authentication. Include the token in the Authorization header:

```
Authorization: Bearer <jwt_token>
```

## User Roles

### Customer
- Can view visas
- Can purchase visas
- Can manage their own purchases

### Admin
- All customer permissions
- Can create, update, and delete visas
- Can manage all purchases

## API Response Format

All API responses follow the international JSON standard:

### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 100,
    "total_pages": 10
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error message",
  "error": {
    "code": "ERROR_CODE",
    "message": "Error message",
    "details": "Additional error details"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Endpoints

### Public Endpoints (No Authentication Required)

#### 1. Health Check
```http
GET /health
```

**Response:**
```json
{
  "success": true,
  "message": "Viskatera API is running",
  "data": {
    "version": "1.0.0",
    "status": "healthy"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 2. User Registration
```http
POST /api/v1/register
Content-Type: application/json
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe",
  "role": "customer"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "John Doe",
      "role": "customer"
    }
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 3. User Login
```http
POST /api/v1/login
Content-Type: application/json
```

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "John Doe",
      "role": "customer"
    }
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 4. Get All Visas
```http
GET /api/v1/visas
```

**Query Parameters:**
- `country` (optional): Filter by country
- `type` (optional): Filter by visa type
- `page` (optional): Page number (default: 1)
- `per_page` (optional): Items per page (default: 10, max: 100)

**Response:**
```json
{
  "success": true,
  "message": "Visas retrieved successfully",
  "data": [
    {
      "id": 1,
      "country": "Japan",
      "type": "Tourist",
      "description": "Tourist visa for Japan with 30 days validity",
      "price": 500000,
      "duration": 30,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 1,
    "total_pages": 1
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 5. Get Visa by ID
```http
GET /api/v1/visas/{id}
```

**Response:**
```json
{
  "success": true,
  "message": "Visa retrieved successfully",
  "data": {
    "visa": {
      "id": 1,
      "country": "Japan",
      "type": "Tourist",
      "description": "Tourist visa for Japan with 30 days validity",
      "price": 500000,
      "duration": 30,
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    "options": [
      {
        "id": 1,
        "visa_id": 1,
        "name": "Express Processing",
        "description": "Fast processing within 3-5 business days",
        "price": 200000,
        "is_active": true,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Protected Endpoints (Authentication Required)

#### 6. Purchase Visa
```http
POST /api/v1/purchases
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "visa_id": 1,
  "visa_option_id": 1
}
```

**Response:**
```json
{
  "success": true,
  "message": "Visa purchase created successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "visa_id": 1,
    "visa_option_id": 1,
    "total_price": 700000,
    "status": "pending",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "visa": {
      "id": 1,
      "country": "Japan",
      "type": "Tourist",
      "description": "Tourist visa for Japan with 30 days validity",
      "price": 500000,
      "duration": 30,
      "is_active": true
    },
    "visa_option": {
      "id": 1,
      "name": "Express Processing",
      "description": "Fast processing within 3-5 business days",
      "price": 200000,
      "is_active": true
    }
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 7. Get User Purchases
```http
GET /api/v1/purchases
Authorization: Bearer <jwt_token>
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `per_page` (optional): Items per page (default: 10, max: 100)

**Response:**
```json
{
  "success": true,
  "message": "Purchases retrieved successfully",
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "visa_id": 1,
      "visa_option_id": 1,
      "total_price": 700000,
      "status": "pending",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "visa": {
        "id": 1,
        "country": "Japan",
        "type": "Tourist",
        "description": "Tourist visa for Japan with 30 days validity",
        "price": 500000,
        "duration": 30,
        "is_active": true
      },
      "visa_option": {
        "id": 1,
        "name": "Express Processing",
        "description": "Fast processing within 3-5 business days",
        "price": 200000,
        "is_active": true
      }
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 1,
    "total_pages": 1
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 8. Get Purchase by ID
```http
GET /api/v1/purchases/{id}
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "success": true,
  "message": "Purchase retrieved successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "visa_id": 1,
    "visa_option_id": 1,
    "total_price": 700000,
    "status": "pending",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "visa": {
      "id": 1,
      "country": "Japan",
      "type": "Tourist",
      "description": "Tourist visa for Japan with 30 days validity",
      "price": 500000,
      "duration": 30,
      "is_active": true
    },
    "visa_option": {
      "id": 1,
      "name": "Express Processing",
      "description": "Fast processing within 3-5 business days",
      "price": 200000,
      "is_active": true
    }
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 9. Update Purchase Status
```http
PUT /api/v1/purchases/{id}/status
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "status": "completed"
}
```

**Valid status values:** `pending`, `completed`, `cancelled`

**Response:**
```json
{
  "success": true,
  "message": "Purchase status updated successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "visa_id": 1,
    "visa_option_id": 1,
    "total_price": 700000,
    "status": "completed",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "visa": {
      "id": 1,
      "country": "Japan",
      "type": "Tourist",
      "description": "Tourist visa for Japan with 30 days validity",
      "price": 500000,
      "duration": 30,
      "is_active": true
    },
    "visa_option": {
      "id": 1,
      "name": "Express Processing",
      "description": "Fast processing within 3-5 business days",
      "price": 200000,
      "is_active": true
    }
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Admin Endpoints (Admin Authentication Required)

#### 10. Create Visa
```http
POST /api/v1/admin/visas
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "country": "Thailand",
  "type": "Tourist",
  "description": "Tourist visa for Thailand with 30 days validity",
  "price": 300000,
  "duration": 30,
  "is_active": true
}
```

**Response:**
```json
{
  "success": true,
  "message": "Visa created successfully",
  "data": {
    "id": 5,
    "country": "Thailand",
    "type": "Tourist",
    "description": "Tourist visa for Thailand with 30 days validity",
    "price": 300000,
    "duration": 30,
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 11. Update Visa
```http
PUT /api/v1/admin/visas/{id}
Authorization: Bearer <admin_jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "country": "Thailand",
  "type": "Tourist",
  "description": "Updated description",
  "price": 350000,
  "duration": 30,
  "is_active": true
}
```

**Response:**
```json
{
  "success": true,
  "message": "Visa updated successfully",
  "data": {
    "id": 5,
    "country": "Thailand",
    "type": "Tourist",
    "description": "Updated description",
    "price": 350000,
    "duration": 30,
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

#### 12. Delete Visa
```http
DELETE /api/v1/admin/visas/{id}
Authorization: Bearer <admin_jwt_token>
```

**Response:**
```json
{
  "success": true,
  "message": "Visa deleted successfully",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `USER_EXISTS` | User with email already exists |
| `INVALID_CREDENTIALS` | Invalid email or password |
| `UNAUTHORIZED` | Authentication required |
| `ACCESS_DENIED` | Admin privileges required |
| `VISA_NOT_FOUND` | Visa not found |
| `PURCHASE_NOT_FOUND` | Purchase not found |
| `DATABASE_ERROR` | Database operation failed |
| `TOKEN_EXPIRED` | JWT token expired |
| `INVALID_TOKEN` | Invalid JWT token |

## Testing

### Using curl

#### Register User
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

#### Get Visas
```bash
curl http://localhost:8080/api/v1/visas
```

#### Purchase Visa
```bash
curl -X POST http://localhost:8080/api/v1/purchases \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "visa_id": 1
  }'
```

### Using Test Script
```bash
./test_api.sh
```

## Documentation Access

- **Swagger UI** (Interactive): http://localhost:8080/swagger/index.html
- **API Docs** (Redirect): http://localhost:8080/docs
- **Health Check**: http://localhost:8080/health

> **Note**: All endpoints are fully documented in Swagger UI with request/response examples, authentication requirements, and error codes.

## Default Admin Credentials

- **Email**: admin@viskatera.com
- **Password**: admin123
- **Role**: admin

## Rate Limiting

Currently no rate limiting is implemented. Consider implementing rate limiting for production use.

## CORS

CORS is enabled for all origins. Configure appropriately for production use.

## Database

- **Host**: localhost
- **Port**: 5433
- **Database**: viskatera_db
- **Username**: postgres
- **Password**: password
- **Admin Panel**: http://localhost:8081
