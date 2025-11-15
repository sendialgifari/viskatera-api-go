package workers

import (
	"log"
	"os"
	"strconv"
	"viskatera-api-go/config"
)

// StartWorkers starts all background workers
func StartWorkers() {
	// Get concurrency from environment or use default
	concurrencyStr := os.Getenv("WORKER_CONCURRENCY")
	concurrency := 10 // Default to 10 parallel workers
	if concurrencyStr != "" {
		if c, err := strconv.Atoi(concurrencyStr); err == nil && c > 0 {
			concurrency = c
		}
	}

	log.Printf("[WORKERS] Starting workers with concurrency: %d", concurrency)

	// Start email worker
	go func() {
		if err := StartEmailWorker(concurrency); err != nil {
			log.Fatalf("[WORKERS] Failed to start email worker: %v", err)
		}
	}()

	log.Println("[WORKERS] All workers started successfully")
}

// InitializeWorkers initializes workers (called from main.go)
func InitializeWorkers() {
	// Connect to RabbitMQ if not already connected
	if config.RabbitMQConn == nil {
		if err := config.ConnectRabbitMQ(); err != nil {
			log.Fatalf("[WORKERS] Failed to connect to RabbitMQ: %v", err)
		}
	}

	// Start workers
	StartWorkers()
}
