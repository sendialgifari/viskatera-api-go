# Fresh Migration Guide

## Overview

Fresh migration allows you to completely drop all database tables and recreate them from scratch. This is useful for:
- Resetting the database in development
- Testing schema changes
- Starting with a clean slate

## WARNING ⚠️

**This will DELETE ALL DATA in your database!**
- All tables will be dropped
- All data will be lost
- This action cannot be undone

Use only in development or when you have a backup.

## Running Fresh Migration

### Method 1: Using the Shell Script (Recommended)

```bash
./scripts/fresh_migrate.sh
```

This script will:
1. Ask for confirmation
2. Drop all existing tables
3. Recreate tables from models
4. Show completion status

### Method 2: Using Go Run

```bash
go run scripts/fresh_migrate.go
```

### Method 3: Direct Execution

```bash
cd scripts
go run fresh_migrate.go
```

## What Gets Dropped and Recreated

The script drops these tables in order:
1. `payments`
2. `visa_purchases`
3. `visa_options`
4. `visas`
5. `users`

Then recreates them with:
- All columns from the models
- All relationships (foreign keys)
- All indexes
- All constraints

## Tables Created

After fresh migration, you'll have:

### users
- id, email, password, name
- avatar_url (NEW)
- role, is_active
- created_at, updated_at, deleted_at

### visas
- id, country, type, description
- price, duration
- visa_document_url (NEW)
- is_active
- created_at, updated_at, deleted_at

### visa_options
- id, visa_id (foreign key)
- name, description, price
- is_active
- created_at, updated_at, deleted_at

### visa_purchases
- id, user_id (foreign key)
- visa_id (foreign key)
- visa_option_id (foreign key)
- total_price, status
- created_at, updated_at, deleted_at

### payments (NEW)
- id, user_id (foreign key)
- purchase_id (foreign key)
- payment_method, amount, status
- xendit_id, payment_url
- created_at, updated_at, deleted_at

## Best Practices

### 1. Always Backup First

```bash
# Export data before migration
pg_dump -U postgres -d viskatera_db > backup.sql

# Restore after migration (if needed)
psql -U postgres -d viskatera_db < backup.sql
```

### 2. Use in Development Only

Fresh migration is designed for development environments. For production:
- Use proper migration tools
- Test migrations in staging
- Create database backups
- Use schema versioning

### 3. After Fresh Migration

Populate initial data:

```bash
# Seed admin user
go run scripts/create_admin.go

# Seed sample data
go run scripts/seed_data.go
```

### 4. Verify Tables

Check that all tables were created:

```sql
-- Connect to PostgreSQL
psql -U postgres -d viskatera_db

-- List tables
\dt

-- Check table structure
\d users
\d visas
\d visa_purchases
\d payments
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Fresh Migration

on:
  push:
    branches: [ main ]

jobs:
  migrate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.24'
      
      - name: Run fresh migration
        run: |
          go run scripts/fresh_migrate.go
        env:
          DB_HOST: ${{ secrets.DB_HOST }}
          DB_USER: ${{ secrets.DB_USER }}
          DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
          DB_NAME: ${{ secrets.DB_NAME }}
```

## Troubleshooting

### Issue: Foreign Key Constraint Error

```
ERROR: update or delete on table "users" violates foreign key constraint
```

**Solution**: The script drops tables in the correct order. If you still see this, run:

```bash
go run scripts/fresh_migrate.go
```

The CASCADE option should handle this.

### Issue: Database Connection Failed

```
FATAL: database connection failed
```

**Solution**: 
1. Check your `.env` file
2. Ensure PostgreSQL is running
3. Verify database credentials
4. Check network connectivity

### Issue: Tables Not Dropped

If some tables persist:

```sql
-- Connect to database
psql -U postgres -d viskatera_db

-- Manually drop tables
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS visa_purchases CASCADE;
DROP TABLE IF EXISTS visa_options CASCADE;
DROP TABLE IF EXISTS visas CASCADE;
DROP TABLE IF EXISTS users CASCADE;
```

### Issue: Migration Completed But No Tables

Check if database exists:

```bash
# List databases
psql -U postgres -l

# Create database if missing
createdb -U postgres viskatera_db
```

## Alternative: Manual Fresh Migration

If you prefer to do it manually:

```bash
# Connect to database
psql -U postgres -d viskatera_db

# Drop all tables
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS visa_purchases CASCADE;
DROP TABLE IF EXISTS visa_options CASCADE;
DROP TABLE IF EXISTS visas CASCADE;
DROP TABLE IF EXISTS users CASCADE;

# Exit psql
\q

# Run migration
go run main.go
```

## Comparison with AutoMigrate

### Regular Migration (AutoMigrate)
- Only creates new columns
- Only adds new tables
- Preserves existing data
- Safe for production

### Fresh Migration
- Drops all tables
- Recreates from scratch
- Loses all data
- Development only

## Development Workflow

Recommended workflow:

```bash
# 1. Start fresh
./scripts/fresh_migrate.sh

# 2. Seed data
go run scripts/create_admin.go
go run scripts/seed_data.go

# 3. Test application
go run main.go

# 4. Make changes to models
# (edit models/*.go)

# 5. Fresh migration again
./scripts/fresh_migrate.sh

# 6. Repeat as needed
```

## Production Workflow

For production, use proper migrations:

```bash
# 1. Backup database
pg_dump -U postgres -d viskatera_db > backup.sql

# 2. Create migration
go run scripts/migrate.go

# 3. Test migration
# (run in staging environment)

# 4. Apply to production
# (after thorough testing)
```

## Summary

Fresh migration is a powerful tool for development that:
- ✅ Resets database to clean state
- ✅ Recreates all tables
- ✅ Removes all data
- ✅ Useful for testing

Remember:
- ⚠️ Only use in development
- ⚠️ Always backup first
- ⚠️ Test thoroughly

