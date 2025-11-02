#!/bin/bash

# Viskatera API Deployment Script
# Usage: ./deploy.sh [build|up|down|restart|logs|backup]

set -e

COMPOSE_FILE="docker-compose.prod.yml"
PROJECT_DIR=$(pwd)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if .env exists
check_env() {
    if [ ! -f .env ]; then
        print_error ".env file not found!"
        print_info "Please copy env.example to .env and configure it"
        exit 1
    fi
}

# Build images
build() {
    print_info "Building Docker images..."
    docker-compose -f $COMPOSE_FILE build --no-cache
    print_info "Build completed!"
}

# Start services
up() {
    check_env
    print_info "Starting services..."
    docker-compose -f $COMPOSE_FILE up -d
    print_info "Services started!"
    print_info "Checking health..."
    sleep 5
    docker-compose -f $COMPOSE_FILE ps
}

# Stop services
down() {
    print_info "Stopping services..."
    docker-compose -f $COMPOSE_FILE down
    print_info "Services stopped!"
}

# Restart services
restart() {
    print_info "Restarting services..."
    docker-compose -f $COMPOSE_FILE restart
    print_info "Services restarted!"
}

# View logs
logs() {
    SERVICE=${2:-""}
    if [ -z "$SERVICE" ]; then
        docker-compose -f $COMPOSE_FILE logs -f
    else
        docker-compose -f $COMPOSE_FILE logs -f $SERVICE
    fi
}

# Backup database
backup() {
    print_info "Creating database backup..."
    BACKUP_DIR="$PROJECT_DIR/backups"
    mkdir -p $BACKUP_DIR
    DATE=$(date +%Y%m%d_%H%M%S)
    FILENAME="backup_${DATE}.sql"
    
    docker-compose -f $COMPOSE_FILE exec -T postgres pg_dump -U ${DB_USER:-postgres} ${DB_NAME:-viskatera_db} > "${BACKUP_DIR}/${FILENAME}"
    
    # Compress
    gzip "${BACKUP_DIR}/${FILENAME}"
    
    print_info "Backup created: ${BACKUP_DIR}/${FILENAME}.gz"
}

# Restore database
restore() {
    if [ -z "$2" ]; then
        print_error "Please specify backup file: ./deploy.sh restore backups/backup_YYYYMMDD_HHMMSS.sql.gz"
        exit 1
    fi
    
    BACKUP_FILE=$2
    if [ ! -f "$BACKUP_FILE" ]; then
        print_error "Backup file not found: $BACKUP_FILE"
        exit 1
    fi
    
    print_warn "This will restore database from backup. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        print_info "Restoring database..."
        
        if [[ "$BACKUP_FILE" == *.gz ]]; then
            gunzip < "$BACKUP_FILE" | docker-compose -f $COMPOSE_FILE exec -T postgres psql -U ${DB_USER:-postgres} -d ${DB_NAME:-viskatera_db}
        else
            cat "$BACKUP_FILE" | docker-compose -f $COMPOSE_FILE exec -T postgres psql -U ${DB_USER:-postgres} -d ${DB_NAME:-viskatera_db}
        fi
        
        print_info "Database restored!"
    else
        print_info "Restore cancelled"
    fi
}

# Health check
health() {
    print_info "Checking service health..."
    docker-compose -f $COMPOSE_FILE ps
    echo ""
    print_info "Testing health endpoint..."
    curl -f http://localhost:8080/health || print_error "Health check failed!"
}

# Main
case "$1" in
    build)
        build
        ;;
    up|start)
        up
        ;;
    down|stop)
        down
        ;;
    restart)
        restart
        ;;
    logs)
        logs "$@"
        ;;
    backup)
        backup
        ;;
    restore)
        restore "$@"
        ;;
    health)
        health
        ;;
    *)
        echo "Usage: $0 {build|up|down|restart|logs|backup|restore|health}"
        echo ""
        echo "Commands:"
        echo "  build     - Build Docker images"
        echo "  up        - Start all services"
        echo "  down      - Stop all services"
        echo "  restart   - Restart all services"
        echo "  logs      - View logs (optionally: logs [service_name])"
        echo "  backup    - Create database backup"
        echo "  restore   - Restore database from backup"
        echo "  health    - Check service health"
        exit 1
        ;;
esac

exit 0

