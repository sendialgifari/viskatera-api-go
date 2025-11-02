# Viskatera API Makefile
# Using app.sh for main operations - this Makefile provides convenience aliases

.PHONY: help install build run test clean dev prod start stop restart migrate fresh-migrate seed admin status

# Default target
help:
	@echo "Viskatera API - Available Commands:"
	@echo "===================================="
	@echo ""
	@echo "Development Commands:"
	@echo "  make dev           - Start in development mode"
	@echo "  make start         - Start application (dev mode)"
	@echo "  make restart       - Restart application"
	@echo "  make stop          - Stop application"
	@echo "  make migrate       - Run database migration"
	@echo "  make fresh-migrate - Fresh migration (dev only, deletes all data)"
	@echo "  make seed          - Seed database with sample data"
	@echo "  make admin         - Create admin user"
	@echo ""
	@echo "Production Commands:"
	@echo "  make prod          - Start in production mode"
	@echo "  make build         - Build production binary"
	@echo ""
	@echo "Utility Commands:"
	@echo "  make install       - Install Go dependencies"
	@echo "  make test          - Test the API endpoints"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make status       - Show application status"
	@echo ""
	@echo "Note: All commands use app.sh internally"
	@echo "For more options: ./app.sh help"
	@echo ""

# Development mode (default)
dev:
	@./app.sh dev start

# Production mode
prod:
	@./app.sh prod start

# Start application (dev mode)
start:
	@./app.sh dev start

# Stop application
stop:
	@./app.sh dev stop

# Restart application
restart:
	@./app.sh dev restart

# Run migration
migrate:
	@./app.sh dev migrate

# Fresh migration (dev only)
fresh-migrate:
	@./app.sh dev fresh-migrate

# Seed database
seed:
	@./app.sh dev seed

# Create admin user
admin:
	@./app.sh dev admin

# Build application
build:
	@./app.sh prod build

# Show status
status:
	@./app.sh dev status

# Install dependencies
install:
	@echo "Installing Go dependencies..."
	@go mod tidy
	@echo "✅ Dependencies installed"

# Test the API
test:
	@echo "Testing API endpoints..."
	@./test_api.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f viskatera-api
	@rm -f viskatera-api.pid
	@rm -f app.log
	@echo "✅ Clean completed"
