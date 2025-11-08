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
    export $(cat .env | grep -v '^#' | xargs)
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
        # Try with "postgres" user, fallback to "$DB_USER" if needed
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

# Determine superuser/admin for first access
# If role "postgres" does NOT exist, fallback to "$DB_USER"
ADMIN_USER="postgres"
ROLE_EXISTS=$(docker-compose -f docker-compose.prod.yml exec -T postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='postgres'" || echo "")
if [ "$ROLE_EXISTS" != "1" ]; then
    echo "⚠️  Role \"postgres\" does not exist, using DB_USER ($DB_USER) as admin"
    ADMIN_USER="$DB_USER"
fi

echo "Creating user $DB_USER if not exists..."
docker-compose -f docker-compose.prod.yml exec -T postgres psql -h localhost -p 5432 -U "$ADMIN_USER" <<EOF
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_user WHERE usename = '$DB_USER') THEN
        CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
        RAISE NOTICE 'User $DB_USER created';
    ELSE
        RAISE NOTICE 'User $DB_USER already exists';
        ALTER USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
        RAISE NOTICE 'Password updated for user $DB_USER';
    END IF;
END
\$\$;
EOF

# Create database if not exists
echo "Creating database $DB_NAME if not exists..."
DB_EXISTS=$(docker-compose -f docker-compose.prod.yml exec -T postgres psql -h localhost -p 5432 -U "$ADMIN_USER" -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'")
if [ "$DB_EXISTS" != "1" ]; then
    docker-compose -f docker-compose.prod.yml exec -T postgres psql -h localhost -p 5432 -U "$ADMIN_USER" -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;"
    echo "Database $DB_NAME created."
else
    echo "Database $DB_NAME already exists."
fi

# Grant privileges on database
echo "Granting privileges on database..."
docker-compose -f docker-compose.prod.yml exec -T postgres psql -h localhost -p 5432 -U "$ADMIN_USER" -c "
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
ALTER DATABASE $DB_NAME OWNER TO $DB_USER;
"

# Grant schema privileges
echo "Granting schema privileges..."
docker-compose -f docker-compose.prod.yml exec -T postgres psql -h localhost -p 5432 -U "$ADMIN_USER" -d $DB_NAME -c "
GRANT ALL ON SCHEMA public TO $DB_USER;
ALTER SCHEMA public OWNER TO $DB_USER;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DB_USER;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;
"

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
    echo "psql: error: connection to server at \"localhost\" (::1), port 5432 failed: FATAL:  role \"$DB_USER\" does not exist"
    echo "  Please check that user \"$DB_USER\" exists and is correctly referenced in your .env and PostgreSQL instance."
    exit 1
fi
echo ""
echo "You can now restart the services:"
echo "  docker-compose -f docker-compose.prod.yml restart"
echo ""
