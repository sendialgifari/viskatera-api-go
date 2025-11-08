#!/bin/bash
# Script untuk memperbaiki user database PostgreSQL
# Digunakan ketika DB_USER di .env berbeda dengan user yang sudah ada di database

set -e

echo "=========================================="
echo "Fix PostgreSQL User Configuration"
echo "=========================================="
echo ""

# Load environment variables
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
else
    echo "Error: .env file not found!"
    exit 1
fi

DB_USER=${DB_USER:-viskatera_user}
DB_NAME=${DB_NAME:-viskatera_db}
DB_PASSWORD=${DB_PASSWORD}
DB_PORT=${DB_PORT:-5433}

if [ -z "$DB_PASSWORD" ]; then
    echo "Error: DB_PASSWORD not set in .env file!"
    exit 1
fi

echo "Configuration:"
echo "  DB_USER: $DB_USER"
echo "  DB_NAME: $DB_NAME"
echo "  DB_PORT: $DB_PORT"
echo ""

# Detect container name and its exposed host port
CONTAINER_NAME="viskatera_postgres_prod"
HOST_PORT=$(docker-compose -f docker-compose.prod.yml port postgres 5432 | awk -F: '{print $2}' | tr -d '[:space:]')
if [ -z "$HOST_PORT" ]; then
    HOST_PORT=$DB_PORT
fi

# Check if the PostgreSQL container is running
if ! docker-compose -f docker-compose.prod.yml ps postgres | grep -q "Up"; then
    echo "Starting PostgreSQL container..."
    docker-compose -f docker-compose.prod.yml up -d postgres
    echo "Waiting for PostgreSQL to be ready..."
    maxtries=15
    for i in $(seq 1 $maxtries); do
        if docker-compose -f docker-compose.prod.yml exec -T postgres pg_isready -h localhost -p 5432 -U postgres > /dev/null 2>&1; then
            break
        elif docker-compose -f docker-compose.prod.yml exec -T postgres pg_isready -h localhost -p 5432 -U "$DB_USER" > /dev/null 2>&1; then
            break
        fi
        sleep 2
    done
else
    for i in {1..10}; do
        if docker-compose -f docker-compose.prod.yml exec -T postgres pg_isready -h localhost -p 5432 -U postgres > /dev/null 2>&1; then
            break
        elif docker-compose -f docker-compose.prod.yml exec -T postgres pg_isready -h localhost -p 5432 -U "$DB_USER" > /dev/null 2>&1; then
            break
        fi
        sleep 2
    done
fi

# Try admin as "postgres", if fails fallback to "DB_USER"
ADMIN_USER="postgres"
# Test if can connect as 'postgres' and database 'postgres' exists
CAN_CONNECT=$(docker-compose -f docker-compose.prod.yml exec -T postgres psql -U "$ADMIN_USER" -d postgres -c '\q' 2>&1 || true)
if echo "$CAN_CONNECT" | grep -qi "does not exist"; then
    # Try using DB_USER as admin, but ensure DB_NAME exists for connecting
    echo "psql: error: connection to server on socket \"/var/run/postgresql/.s.PGSQL.5432\" failed: FATAL:  role \"postgres\" does not exist"
    echo "⚠️  Role \"postgres\" does not exist, using DB_USER ($DB_USER) as admin"
    ADMIN_USER="$DB_USER"
    # Special init: If DB_NAME does not exist, create it using template1
    DB_EXISTS=$(docker-compose -f docker-compose.prod.yml exec -T postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" -U "$ADMIN_USER" -d template1 2>/dev/null || true)
    if [ "$DB_EXISTS" != "1" ]; then
        echo "Database \"$DB_NAME\" does not exist; creating it from template1..."
        docker-compose -f docker-compose.prod.yml exec -T postgres psql -U "$ADMIN_USER" -d template1 -c "CREATE DATABASE $DB_NAME;" || true
    fi
fi

# Now try to create the user if not exists. Use template1 as fallback database context for creation.
echo "Creating user $DB_USER if not exists..."
USER_EXISTS=$(docker-compose -f docker-compose.prod.yml exec -T postgres psql -qAt -U "$ADMIN_USER" -d "$DB_NAME" -c "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER';" 2>/dev/null || true)
if [ "$USER_EXISTS" != "1" ]; then
    # User not exists, create
    docker-compose -f docker-compose.prod.yml exec -T postgres psql -U "$ADMIN_USER" -d "$DB_NAME" -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" >/dev/null 2>&1 && \
        echo "User $DB_USER created."
else
    # User exists, update password
    docker-compose -f docker-compose.prod.yml exec -T postgres psql -U "$ADMIN_USER" -d "$DB_NAME" -c "ALTER USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" >/dev/null 2>&1 && \
        echo "User $DB_USER already exists (password updated)."
fi

# Create database if not exists
echo "Creating database $DB_NAME if not exists..."
DB_EXISTS=$(docker-compose -f docker-compose.prod.yml exec -T postgres psql -U "$ADMIN_USER" -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" -d template1 2>/dev/null || true)
if [ "$DB_EXISTS" != "1" ]; then
    docker-compose -f docker-compose.prod.yml exec -T postgres psql -U "$ADMIN_USER" -d template1 -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" && \
    echo "Database $DB_NAME created."
else
    echo "Database $DB_NAME already exists."
fi

# Grant privileges on database
echo "Granting privileges on database..."
docker-compose -f docker-compose.prod.yml exec -T postgres psql -U "$ADMIN_USER" -d "$DB_NAME" -c "
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
ALTER DATABASE $DB_NAME OWNER TO $DB_USER;
" >/dev/null 2>&1

# Grant schema privileges
echo "Granting schema privileges..."
docker-compose -f docker-compose.prod.yml exec -T postgres psql -U "$ADMIN_USER" -d "$DB_NAME" -c "
GRANT ALL ON SCHEMA public TO $DB_USER;
ALTER SCHEMA public OWNER TO $DB_USER;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;
" >/dev/null 2>&1

echo ""
echo "=========================================="
echo "Database user configuration completed!"
echo "=========================================="
echo ""
echo "Testing connection..."

# Test connection to DB using host mapping via docker exec
docker-compose -f docker-compose.prod.yml exec -T postgres psql -h localhost -p 5432 -U "$DB_USER" -d "$DB_NAME" -c "SELECT version();" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Connection test successful!"
else
    echo "❌ Connection test failed!"
    echo "psql: error: connection to server at \"localhost\" (::1), port 5432 failed: FATAL:  database \"$DB_NAME\" does not exist"
    echo "  Please check that database \"$DB_NAME\" exists and user \"$DB_USER\" is correctly referenced in your .env and PostgreSQL instance."
    exit 1
fi
echo ""
echo "You can now restart the services:"
echo "  docker-compose -f docker-compose.prod.yml restart"
echo ""
