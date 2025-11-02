package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"viskatera-api-go/config"
	"viskatera-api-go/routes"

	_ "viskatera-api-go/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title Viskatera API
// @version 1.0.0
// @description Comprehensive Visa Management API with role-based authentication, OTP login, payment processing, and document management
// @termsOfService http://swagger.io/terms/

// @contact.name Viskatera API Support
// @contact.url https://viskatera.com/support
// @contact.email support@viskatera.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		env := os.Getenv("ENVIRONMENT")
		if env == "production" || env == "prod" {
			gin.SetMode(gin.ReleaseMode)
		}
	} else {
		gin.SetMode(os.Getenv("GIN_MODE"))
	}

	// Connect to database
	config.ConnectDB()
	config.MigrateDB()

	// Setup routes
	r := routes.SetupRoutes()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server with timeouts for production
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s (mode: %s)", port, gin.Mode())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Close database connection
	if err := config.CloseDB(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server exited gracefully")
}
