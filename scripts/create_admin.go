package main

import (
	"log"
	"viskatera-api-go/config"
	"viskatera-api-go/models"
	"viskatera-api-go/utils"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Connect to database
	config.ConnectDB()

	// Create admin user
	adminEmail := "admin@viskatera.com"
	adminPassword := "admin123"
	adminName := "System Administrator"

	// Check if admin already exists
	var existingAdmin models.User
	if err := config.DB.Where("email = ?", adminEmail).First(&existingAdmin).Error; err == nil {
		log.Printf("Admin user already exists: %s", adminEmail)
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(adminPassword)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Create admin user
	admin := models.User{
		Email:    adminEmail,
		Password: hashedPassword,
		Name:     adminName,
		Role:     models.RoleAdmin,
		IsActive: true,
	}

	if err := config.DB.Create(&admin).Error; err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	log.Printf("✅ Admin user created successfully!")
	log.Printf("   Email: %s", adminEmail)
	log.Printf("   Password: %s", adminPassword)
	log.Printf("   Role: %s", admin.Role)
	log.Printf("")
	log.Printf("You can now login with these credentials to access admin features.")
}
