#!/bin/bash

BASE_URL="http://localhost:8080"

echo "Testing Viskatera API..."
echo "========================"

# Test health check
echo "1. Testing health check..."
curl -s "$BASE_URL/health" | jq .
echo ""

# Test register user
echo "2. Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }')
echo "$REGISTER_RESPONSE" | jq .
echo ""

# Test login
echo "3. Testing user login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')
echo "$LOGIN_RESPONSE" | jq .

# Extract JWT token
JWT_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
echo "JWT Token: $JWT_TOKEN"
echo ""

# Test get visas
echo "4. Testing get visas..."
curl -s "$BASE_URL/api/v1/visas" | jq .
echo ""

# Test get visa by ID
echo "5. Testing get visa by ID..."
curl -s "$BASE_URL/api/v1/visas/1" | jq .
echo ""

# Test purchase visa (protected)
echo "6. Testing purchase visa..."
curl -s -X POST "$BASE_URL/api/v1/purchases" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "visa_id": 1
  }' | jq .
echo ""

# Test get user purchases
echo "7. Testing get user purchases..."
curl -s -X GET "$BASE_URL/api/v1/purchases" \
  -H "Authorization: Bearer $JWT_TOKEN" | jq .
echo ""

echo "API testing completed!"
