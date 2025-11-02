#!/bin/bash

# Fresh Migration Script
# This script drops all tables and recreates them

set -e

echo "==========================================="
echo "  Viskatera API - Fresh Migration"
echo "==========================================="
echo ""
echo "⚠️  WARNING: This will DELETE ALL DATA!"
echo "⚠️  All existing tables will be dropped!"
echo ""

# Ask for confirmation
read -p "Are you sure you want to continue? (yes/no): " -r
echo
if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    echo "Migration cancelled."
    exit 0
fi

echo "Starting fresh migration..."
echo ""

# Run the migration
go run scripts/fresh_migrate.go

echo ""
echo "✅ Fresh migration completed!"
echo "You can now run the seed script to populate initial data."
echo ""
echo "To seed data, run:"
echo "  go run scripts/seed_data.go"

