#!/usr/bin/env bash

# Viskatera API - Main Application Runner
# Usage: ./app.sh [mode] [action]
# Modes: dev, prod
# Actions: start, stop, restart, migrate, fresh-migrate, seed

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Handle help command early
if [ "$1" = "help" ] || [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    MODE=""
    ACTION="help"
else
    # Default values
    MODE="${1:-dev}"
    ACTION="${2:-start}"
fi

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/.env"
BINARY_NAME="viskatera-api"
PID_FILE="$SCRIPT_DIR/$BINARY_NAME.pid"

# Functions
print_header() {
    local mode_upper=$(echo "$MODE" | tr '[:lower:]' '[:upper:]')
    echo -e "${BLUE}════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  Viskatera API - ${mode_upper} Mode${NC}"
    echo -e "${BLUE}════════════════════════════════════════════════${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

check_prerequisites() {
    print_info "Checking prerequisites..."
    
    local missing=0
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21+"
        missing=1
    fi
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        missing=1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed"
        missing=1
    fi
    
    if [ $missing -eq 1 ]; then
        exit 1
    fi
    
    print_success "All prerequisites installed"
}

setup_env() {
    if [ ! -f "$ENV_FILE" ]; then
        print_info "Creating .env file from template..."
        cp "$SCRIPT_DIR/env.example" "$ENV_FILE"
        print_success ".env file created"
    fi
    
    # Load environment variables
    set -a
    source "$ENV_FILE"
    set +a
}

check_database() {
    print_info "Checking database connection..."
    
    if [ "$MODE" = "dev" ]; then
        # Check if Docker container is running
        if ! docker-compose ps postgres | grep -q "Up"; then
            print_info "Starting PostgreSQL database..."
            docker-compose up -d postgres
            print_info "Waiting for database to be ready..."
            sleep 5
        fi
        
        # Wait for database to be ready
        local retries=30
        local count=0
        while ! docker-compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; do
            count=$((count + 1))
            if [ $count -ge $retries ]; then
                print_error "Database is not ready after $retries retries"
                exit 1
            fi
            sleep 1
        done
        
        # Create database if it doesn't exist
        docker-compose exec -T postgres psql -U postgres -c "SELECT 1 FROM pg_database WHERE datname='viskatera_db'" | grep -q 1 || \
        docker-compose exec -T postgres psql -U postgres -c "CREATE DATABASE viskatera_db;" > /dev/null 2>&1
        
        print_success "Database is ready"
    else
        # Production: just check connection
        print_info "Using production database configuration"
    fi
}

install_dependencies() {
    print_info "Installing Go dependencies..."
    cd "$SCRIPT_DIR"
    go mod tidy
    print_success "Dependencies installed"
}

run_migration() {
    print_info "Running database migration..."
    cd "$SCRIPT_DIR"
    go run scripts/migrate.go
    print_success "Migration completed"
}

run_fresh_migration() {
    print_warning "This will DELETE ALL DATA in the database!"
    read -p "Are you sure? Type 'yes' to continue: " confirm
    
    if [ "$confirm" != "yes" ]; then
        print_info "Cancelled"
        exit 0
    fi
    
    print_info "Running fresh migration..."
    cd "$SCRIPT_DIR"
    go run scripts/fresh_migrate.go
    print_success "Fresh migration completed"
}

seed_database() {
    print_info "Seeding database..."
    cd "$SCRIPT_DIR"
    go run scripts/seed_data.go
    print_success "Database seeded"
}

create_admin() {
    print_info "Creating admin user..."
    cd "$SCRIPT_DIR"
    go run scripts/create_admin.go
    print_success "Admin user created"
}

build_application() {
    print_info "Building application..."
    cd "$SCRIPT_DIR"
    
    if [ "$MODE" = "prod" ]; then
        # Production build with optimizations
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o "$BINARY_NAME" main.go
        print_success "Production binary built: $BINARY_NAME"
    else
        # Development build
        go build -o "$BINARY_NAME" main.go
        print_success "Development binary built: $BINARY_NAME"
    fi
}

start_application() {
    cd "$SCRIPT_DIR"
    
    # Check if already running
    if [ -f "$PID_FILE" ] && ps -p "$(cat $PID_FILE)" > /dev/null 2>&1; then
        print_warning "Application is already running (PID: $(cat $PID_FILE))"
        return
    fi
    
    if [ "$MODE" = "prod" ]; then
        # Production: run binary
        if [ ! -f "$BINARY_NAME" ]; then
            build_application
        fi
        
        print_info "Starting application in production mode..."
        nohup "./$BINARY_NAME" > "$SCRIPT_DIR/app.log" 2>&1 &
        echo $! > "$PID_FILE"
        print_success "Application started (PID: $(cat $PID_FILE))"
        print_info "Logs: tail -f $SCRIPT_DIR/app.log"
    else
        # Development: run with live reload
        print_info "Starting application in development mode..."
        
        if command -v air >/dev/null 2>&1; then
            print_info "Using Air for live reload..."
            air -c .air.toml 2>&1 || {
                print_warning "Air config not found, using basic Go run..."
                go run main.go
            }
        elif command -v reflex >/dev/null 2>&1; then
            print_info "Using Reflex for live reload..."
            reflex -r '\.go$' -- sh -c 'go run main.go'
        else
            print_info "Running without live reload..."
            go run main.go
        fi
    fi
}

stop_application() {
    if [ ! -f "$PID_FILE" ]; then
        print_warning "PID file not found. Application may not be running."
        return
    fi
    
    local pid=$(cat "$PID_FILE")
    
    if ! ps -p "$pid" > /dev/null 2>&1; then
        print_warning "Process $pid not found. Removing stale PID file."
        rm -f "$PID_FILE"
        return
    fi
    
    print_info "Stopping application (PID: $pid)..."
    kill "$pid" 2>/dev/null || true
    
    # Wait for process to stop
    local count=0
    while ps -p "$pid" > /dev/null 2>&1 && [ $count -lt 10 ]; do
        sleep 1
        count=$((count + 1))
    done
    
    if ps -p "$pid" > /dev/null 2>&1; then
        print_warning "Process did not stop gracefully, forcing kill..."
        kill -9 "$pid" 2>/dev/null || true
    fi
    
    rm -f "$PID_FILE"
    print_success "Application stopped"
}

restart_application() {
    stop_application
    sleep 2
    start_application
}

show_status() {
    echo ""
    print_info "Application Status:"
    echo "────────────────────────────────────────"
    
    if [ -f "$PID_FILE" ] && ps -p "$(cat $PID_FILE)" > /dev/null 2>&1; then
        print_success "Status: Running (PID: $(cat $PID_FILE))"
    else
        print_warning "Status: Not running"
    fi
    
    if [ "$MODE" = "dev" ]; then
        if docker-compose ps postgres | grep -q "Up"; then
            print_success "Database: Running"
        else
            print_warning "Database: Stopped"
        fi
    fi
    
    echo ""
    print_info "Access Points:"
    echo "  API Server: http://localhost:${PORT:-8080}"
    echo "  Health Check: http://localhost:${PORT:-8080}/health"
    echo "  Swagger Docs: http://localhost:${PORT:-8080}/docs"
    if [ "$MODE" = "dev" ]; then
        echo "  Database Admin: http://localhost:8081"
        echo "  MailHog UI: http://localhost:8025"
    fi
    echo ""
}

show_help() {
    echo "Viskatera API - Application Runner"
    echo ""
    echo "Usage: ./app.sh [mode] [action]"
    echo ""
    echo "Modes:"
    echo "  dev   - Development mode (default)"
    echo "  prod  - Production mode"
    echo ""
    echo "Actions:"
    echo "  start         - Start the application (default)"
    echo "  stop          - Stop the application"
    echo "  restart       - Restart the application"
    echo "  migrate       - Run database migration"
    echo "  fresh-migrate - Run fresh migration (DANGER: deletes all data)"
    echo "  seed          - Seed database with sample data"
    echo "  admin         - Create admin user"
    echo "  build         - Build the application binary"
    echo "  status        - Show application status"
    echo "  help          - Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./app.sh dev start          # Start in development mode"
    echo "  ./app.sh prod start         # Start in production mode"
    echo "  ./app.sh dev migrate        # Run migration in dev"
    echo "  ./app.sh dev fresh-migrate  # Fresh migration (dev only)"
    echo ""
}

# Main execution
main() {
    # Handle help early
    if [ "$ACTION" = "help" ] || [ -z "$MODE" ]; then
        show_help
        exit 0
    fi
    
    print_header
    
    # Validate mode
    if [ "$MODE" != "dev" ] && [ "$MODE" != "prod" ]; then
        print_error "Invalid mode: $MODE"
        echo "Use 'dev' or 'prod'"
        show_help
        exit 1
    fi
    
    # Setup
    setup_env
    check_prerequisites
    
    # Execute action
    case "$ACTION" in
        start)
            check_database
            install_dependencies
            start_application
            show_status
            ;;
        stop)
            stop_application
            ;;
        restart)
            check_database
            restart_application
            show_status
            ;;
        migrate)
            check_database
            install_dependencies
            run_migration
            ;;
        fresh-migrate)
            if [ "$MODE" != "dev" ]; then
                print_error "Fresh migration is only allowed in development mode"
                exit 1
            fi
            check_database
            install_dependencies
            run_fresh_migration
            ;;
        seed)
            if [ "$MODE" != "dev" ]; then
                print_error "Database seeding is only allowed in development mode"
                exit 1
            fi
            check_database
            install_dependencies
            seed_database
            ;;
        admin)
            check_database
            install_dependencies
            create_admin
            ;;
        build)
            install_dependencies
            build_application
            ;;
        status)
            setup_env
            show_status
            ;;
        *)
            print_error "Unknown action: $ACTION"
            show_help
            exit 1
            ;;
    esac
}

main

