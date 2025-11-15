package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	RabbitMQConn *amqp.Connection
	RabbitMQChan *amqp.Channel
	queueOnce    sync.Once
)

// Queue names
const (
	QueueEmailInvoice        = "email_invoice"
	QueueEmailPaymentSuccess = "email_payment_success"
	QueueGeneratePDF         = "generate_pdf"
)

// ConnectRabbitMQ connects to RabbitMQ server
func ConnectRabbitMQ() error {
	var err error

	host := os.Getenv("RABBITMQ_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("RABBITMQ_PORT")
	if port == "" {
		port = "5672"
	}

	user := os.Getenv("RABBITMQ_USER")
	if user == "" {
		user = "admin"
	}

	pass := os.Getenv("RABBITMQ_PASS")
	if pass == "" {
		pass = "admin123"
	}

	vhost := os.Getenv("RABBITMQ_VHOST")
	if vhost == "" {
		vhost = "/"
	}

	// Build connection URL
	amqpURL := fmt.Sprintf("amqp://%s:%s@%s:%s%s", user, pass, host, port, vhost)

	// Connect to RabbitMQ
	RabbitMQConn, err = amqp.Dial(amqpURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	RabbitMQChan, err = RabbitMQConn.Channel()
	if err != nil {
		RabbitMQConn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	log.Println("RabbitMQ connected successfully!")

	// Declare queues
	if err := DeclareQueues(); err != nil {
		return fmt.Errorf("failed to declare queues: %w", err)
	}

	return nil
}

// DeclareQueues declares all required queues
func DeclareQueues() error {
	queues := []string{
		QueueEmailInvoice,
		QueueEmailPaymentSuccess,
		QueueGeneratePDF,
	}

	for _, queueName := range queues {
		_, err := RabbitMQChan.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			amqp.Table{
				"x-message-ttl": int32(3600000), // 1 hour TTL
			}, // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queueName, err)
		}
	}

	log.Println("All queues declared successfully!")
	return nil
}

// PublishMessage publishes a message to a queue
func PublishMessage(queueName string, message []byte) error {
	if RabbitMQChan == nil {
		return fmt.Errorf("RabbitMQ channel not initialized")
	}

	err := RabbitMQChan.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Make message persistent
			Body:         message,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Message published to queue: %s", queueName)
	return nil
}

// CloseRabbitMQ closes RabbitMQ connection
func CloseRabbitMQ() error {
	if RabbitMQChan != nil {
		RabbitMQChan.Close()
	}
	if RabbitMQConn != nil {
		return RabbitMQConn.Close()
	}
	return nil
}

// GetQueueStats returns queue statistics
func GetQueueStats(queueName string) (int, error) {
	if RabbitMQChan == nil {
		return 0, fmt.Errorf("RabbitMQ channel not initialized")
	}

	queue, err := RabbitMQChan.QueueInspect(queueName)
	if err != nil {
		return 0, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return queue.Messages, nil
}

// GetAllQueueStats returns statistics for all queues
func GetAllQueueStats() (map[string]int, error) {
	stats := make(map[string]int)
	queues := []string{
		QueueEmailInvoice,
		QueueEmailPaymentSuccess,
		QueueGeneratePDF,
	}

	for _, queueName := range queues {
		count, err := GetQueueStats(queueName)
		if err != nil {
			return nil, err
		}
		stats[queueName] = count
	}

	return stats, nil
}
