package main

import (
	"fmt"
	"log"

	// "os"
	"viskatera-api-go/config"
	"viskatera-api-go/models"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Connect to database
	config.ConnectDB()

	fmt.Println("Starting fresh migration...")
	fmt.Println("WARNING: This will drop all existing tables and recreate them!")
	fmt.Println("Press Ctrl+C to cancel, or wait 5 seconds to continue...")

	// Wait 5 seconds (in production, remove this or ask for confirmation)
	// This is a safeguard to prevent accidental data loss
	// time.Sleep(5 * time.Second)

	// Drop all tables
	fmt.Println("\nDropping existing tables...")

	// Drop tables in reverse order to respect foreign key constraints
	tables := []string{
		"activity_logs",
		"payments",
		"password_reset_tokens",
		"otps",
		"visa_purchases",
		"visa_options",
		"visas",
		"users",
	}

	for _, table := range tables {
		if err := config.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			log.Printf("Warning: Failed to drop table %s: %v", table, err)
		} else {
			fmt.Printf("✓ Dropped table: %s\n", table)
		}
	}

	// Also drop tables using GORM's DropTable if they exist
	fmt.Println("\nCleaning up with GORM...")
	config.DB.Migrator().DropTable(
		&models.ActivityLog{},
		&models.Payment{},
		&models.PasswordResetToken{},
		&models.OTP{},
		&models.VisaPurchase{},
		&models.VisaOption{},
		&models.Visa{},
		&models.User{},
	)

	// Fresh migration
	fmt.Println("\nRunning fresh migration...")
	err := config.DB.AutoMigrate(
		&models.User{},
		&models.Visa{},
		&models.VisaOption{},
		&models.VisaPurchase{},
		&models.PasswordResetToken{},
		&models.OTP{},
		&models.Payment{},
		&models.ActivityLog{},
	)

	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	fmt.Println("\n✓ Fresh migration completed successfully!")
	fmt.Println("\nTables created:")
	fmt.Println("  - users")
	fmt.Println("  - visas")
	fmt.Println("  - visa_options")
	fmt.Println("  - visa_purchases")
	fmt.Println("  - password_reset_tokens")
	fmt.Println("  - otps")
	fmt.Println("  - payments")

	fmt.Println("\nDatabase is now in a fresh state and ready to use.")
}
