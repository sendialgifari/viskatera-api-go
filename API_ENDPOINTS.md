# API Endpoints Quick Reference

> **📚 Full Interactive Documentation**: Visit **[Swagger UI](http://localhost:8080/swagger/index.html)** for complete API documentation with interactive testing

## Base URL
```
http://localhost:8080
```

## Authentication
Protected endpoints require JWT token in Authorization header:
```
Authorization: Bearer <jwt_token>
```

---

## Public Endpoints (No Authentication Required)

### 1. Health Check
```http
GET /health
```
**Response:**
```json
{
  "status": "ok",
  "message": "Viskatera API is running"
}
```

### 2. User Registration
```http
POST /api/v1/register
Content-Type: application/json
```
**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```
**Response:**
```json
{
  "message": "User registered successfully",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

### 3. User Login
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
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

### 4. Get All Visas
```http
GET /api/v1/visas
```
**Query Parameters:**
- `country` (optional): Filter by country
- `type` (optional): Filter by visa type

**Examples:**
```http
GET /api/v1/visas
GET /api/v1/visas?country=Japan
GET /api/v1/visas?type=Tourist
GET /api/v1/visas?country=Japan&type=Tourist
```

**Response:**
```json
{
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
  ]
}
```

### 5. Get Visa by ID
```http
GET /api/v1/visas/{id}
```

**Response:**
```json
{
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
  }
}
```

---

## Protected Endpoints (Authentication Required)

### 6. Purchase Visa
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
**Note:** `visa_option_id` is optional

**Response:**
```json
{
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
  }
}
```

### 7. Get User Purchases
```http
GET /api/v1/purchases
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
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
  ]
}
```

### 8. Get Purchase by ID
```http
GET /api/v1/purchases/{id}
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
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
  }
}
```

### 9. Update Purchase Status
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
  "message": "Purchase status updated successfully",
  "data": {
    "id": 1,
    "user_id": 1,
    "visa_id": 1,
    "visa_option_id": 1,
    "total_price": 700000,
    "status": "completed",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

## Admin Endpoints (For Testing/Management)

### 10. Create Visa
```http
POST /api/v1/admin/visas
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

### 11. Update Visa
```http
PUT /api/v1/admin/visas/{id}
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

### 12. Delete Visa
```http
DELETE /api/v1/admin/visas/{id}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request data"
}
```

### 401 Unauthorized
```json
{
  "error": "Authorization header required"
}
```

### 404 Not Found
```json
{
  "error": "Resource not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```

---

## Testing with curl

### Register User
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Get Visas
```bash
curl http://localhost:8080/api/v1/visas
```

### Purchase Visa
```bash
curl -X POST http://localhost:8080/api/v1/purchases \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "visa_id": 1
  }'
```
