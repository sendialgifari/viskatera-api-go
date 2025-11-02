package main

import (
	"log"
	"viskatera-api-go/config"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Connect to database
	config.ConnectDB()

	// Run migrations
	config.MigrateDB()

	log.Println("Database migration completed successfully!")
}
