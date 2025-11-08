package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"viskatera-api-go/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	var err error

	// Database connection string
	// SSL mode: disable (default), require, verify-full
	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable" // Default untuk development
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		sslMode,
	)

	// Determine log level based on environment
	logLevel := logger.Info
	if os.Getenv("GIN_MODE") == "release" {
		logLevel = logger.Error
	}

	// Connect to database with optimized configuration
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool for production
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}

	// Optimized connection pool settings for production
	// SetMaxIdleConns: Keep more idle connections to reduce connection overhead
	sqlDB.SetMaxIdleConns(25)

	// SetMaxOpenConns: Increase max connections for high traffic
	// Formula: ((core_count * 2) + effective_spindle_count)
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime: Reuse connections efficiently
	sqlDB.SetConnMaxLifetime(time.Hour)

	// SetConnMaxIdleTime: Close idle connections after 10 minutes
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	log.Println("Database connected successfully with connection pooling configured!")
}

func MigrateDB() {
	err := DB.AutoMigrate(
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
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed!")
}

// CloseDB closes the database connection gracefully
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
