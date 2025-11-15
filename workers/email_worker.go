package workers

import (
	"encoding/json"
	"fmt"
	"log"
	"viskatera-api-go/config"
	"viskatera-api-go/models"
	"viskatera-api-go/utils"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EmailInvoiceJob represents job data for sending invoice email
type EmailInvoiceJob struct {
	PurchaseID uint   `json:"purchase_id"`
	UserID     uint   `json:"user_id"`
	Email      string `json:"email"`
	Type       string `json:"type"` // "invoice" or "payment_success"
}

// StartEmailWorker starts the email worker with parallel processing
func StartEmailWorker(concurrency int) error {
	if concurrency < 1 {
		concurrency = 10 // Default to 10 parallel workers
	}

	log.Printf("[EMAIL-WORKER] Starting with %d parallel workers", concurrency)

	// Start workers for invoice email queue
	go func() {
		ch, err := config.RabbitMQConn.Channel()
		if err != nil {
			log.Printf("[EMAIL-WORKER] Failed to create channel for invoice queue: %v", err)
			return
		}
		defer ch.Close()

		// Set QoS to allow parallel processing
		err = ch.Qos(
			concurrency, // prefetch count - number of unacknowledged messages per worker
			0,           // prefetch size
			false,       // global
		)
		if err != nil {
			log.Printf("[EMAIL-WORKER] Failed to set QoS: %v", err)
			return
		}

		// Consume from email_invoice queue
		msgs, err := ch.Consume(
			config.QueueEmailInvoice,
			"",    // consumer tag
			false, // auto-ack
			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,   // args
		)
		if err != nil {
			log.Printf("[EMAIL-WORKER] Failed to consume from invoice queue: %v", err)
			return
		}

		// Process messages from invoice queue in parallel
		for i := 0; i < concurrency; i++ {
			go func(workerID int) {
				for msg := range msgs {
					processInvoiceEmail(msg, workerID)
				}
			}(i)
		}
	}()

	// Start workers for payment success email queue
	go func() {
		ch, err := config.RabbitMQConn.Channel()
		if err != nil {
			log.Printf("[EMAIL-WORKER] Failed to create channel for payment success queue: %v", err)
			return
		}
		defer ch.Close()

		// Set QoS to allow parallel processing
		err = ch.Qos(
			concurrency, // prefetch count
			0,           // prefetch size
			false,       // global
		)
		if err != nil {
			log.Printf("[EMAIL-WORKER] Failed to set QoS: %v", err)
			return
		}

		// Consume from email_payment_success queue
		msgsSuccess, err := ch.Consume(
			config.QueueEmailPaymentSuccess,
			"",    // consumer tag
			false, // auto-ack
			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,   // args
		)
		if err != nil {
			log.Printf("[EMAIL-WORKER] Failed to consume from payment success queue: %v", err)
			return
		}

		// Process messages from payment success queue in parallel
		for i := 0; i < concurrency; i++ {
			go func(workerID int) {
				for msg := range msgsSuccess {
					processPaymentSuccessEmail(msg, workerID)
				}
			}(i)
		}
	}()

	log.Printf("[EMAIL-WORKER] All workers started successfully")

	// Keep worker running
	select {}
}

// processInvoiceEmail processes invoice email job
func processInvoiceEmail(msg amqp.Delivery, workerID int) {
	var job EmailInvoiceJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		log.Printf("[EMAIL-WORKER-%d] Failed to unmarshal job: %v", workerID, err)
		msg.Nack(false, false) // Reject and don't requeue
		return
	}

	log.Printf("[EMAIL-WORKER-%d] Processing invoice email for purchase ID: %d", workerID, job.PurchaseID)

	// Get purchase and user data
	var purchase models.VisaPurchase
	if err := config.DB.Preload("Visa").Preload("VisaOption").First(&purchase, job.PurchaseID).Error; err != nil {
		log.Printf("[EMAIL-WORKER-%d] Failed to get purchase: %v", workerID, err)
		msg.Nack(false, true) // Reject and requeue
		return
	}

	var user models.User
	if err := config.DB.First(&user, job.UserID).Error; err != nil {
		log.Printf("[EMAIL-WORKER-%d] Failed to get user: %v", workerID, err)
		msg.Nack(false, true) // Reject and requeue
		return
	}

	// Get payment if exists
	var payment models.Payment
	config.DB.Where("purchase_id = ?", job.PurchaseID).Order("created_at DESC").First(&payment)

	// Generate invoice email body
	subject := fmt.Sprintf("Invoice for Visa Purchase - %s", purchase.Visa.Country)
	body := generateInvoiceEmailBody(purchase, user, payment)

	// Send email
	if err := utils.SendEmail(job.Email, subject, body); err != nil {
		log.Printf("[EMAIL-WORKER-%d] Failed to send email: %v", workerID, err)
		msg.Nack(false, true) // Reject and requeue
		return
	}

	log.Printf("[EMAIL-WORKER-%d] Invoice email sent successfully to %s", workerID, job.Email)
	msg.Ack(false) // Acknowledge message
}

// processPaymentSuccessEmail processes payment success email with PDF
func processPaymentSuccessEmail(msg amqp.Delivery, workerID int) {
	var job EmailInvoiceJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		log.Printf("[EMAIL-WORKER-%d] Failed to unmarshal job: %v", workerID, err)
		msg.Nack(false, false) // Reject and don't requeue
		return
	}

	log.Printf("[EMAIL-WORKER-%d] Processing payment success email for purchase ID: %d", workerID, job.PurchaseID)

	// Get purchase and user data
	var purchase models.VisaPurchase
	if err := config.DB.Preload("Visa").Preload("VisaOption").First(&purchase, job.PurchaseID).Error; err != nil {
		log.Printf("[EMAIL-WORKER-%d] Failed to get purchase: %v", workerID, err)
		msg.Nack(false, true) // Reject and requeue
		return
	}

	var user models.User
	if err := config.DB.First(&user, job.UserID).Error; err != nil {
		log.Printf("[EMAIL-WORKER-%d] Failed to get user: %v", workerID, err)
		msg.Nack(false, true) // Reject and requeue
		return
	}

	// Get payment
	var payment models.Payment
	if err := config.DB.Where("purchase_id = ?", job.PurchaseID).Order("created_at DESC").First(&payment).Error; err != nil {
		log.Printf("[EMAIL-WORKER-%d] Failed to get payment: %v", workerID, err)
		msg.Nack(false, true) // Reject and requeue
		return
	}

	// Generate PDF invoice
	pdfPath, err := utils.GeneratePurchaseInvoicePDF(purchase, user, payment)
	if err != nil {
		log.Printf("[EMAIL-WORKER-%d] Failed to generate PDF: %v", workerID, err)
		msg.Nack(false, true) // Reject and requeue
		return
	}

	// Generate email body
	subject := fmt.Sprintf("Payment Successful - Invoice #%d", purchase.ID)
	body := generatePaymentSuccessEmailBody(purchase, user, payment)

	// Send email with PDF attachment
	if err := utils.SendEmailWithAttachment(job.Email, subject, body, pdfPath); err != nil {
		log.Printf("[EMAIL-WORKER-%d] Failed to send email: %v", workerID, err)
		msg.Nack(false, true) // Reject and requeue
		return
	}

	log.Printf("[EMAIL-WORKER-%d] Payment success email with PDF sent to %s", workerID, job.Email)
	msg.Ack(false) // Acknowledge message
}

// generateInvoiceEmailBody generates HTML email body for invoice
func generateInvoiceEmailBody(purchase models.VisaPurchase, user models.User, payment models.Payment) string {
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; padding: 20px;">
			<h2>Invoice for Visa Purchase</h2>
			<p>Dear %s,</p>
			<p>Thank you for your purchase. Here are your order details:</p>
			<table style="border-collapse: collapse; width: 100%%; margin: 20px 0;">
				<tr>
					<td style="padding: 10px; border: 1px solid #ddd;"><strong>Visa</strong></td>
					<td style="padding: 10px; border: 1px solid #ddd;">%s - %s</td>
				</tr>
				<tr>
					<td style="padding: 10px; border: 1px solid #ddd;"><strong>Total Price</strong></td>
					<td style="padding: 10px; border: 1px solid #ddd;">Rp %.2f</td>
				</tr>
				<tr>
					<td style="padding: 10px; border: 1px solid #ddd;"><strong>Status</strong></td>
					<td style="padding: 10px; border: 1px solid #ddd;">%s</td>
				</tr>
			</table>
	`, user.Name, purchase.Visa.Country, purchase.Visa.Type, purchase.TotalPrice, purchase.Status)

	if payment.PaymentURL != "" {
		body += fmt.Sprintf(`
			<p><strong>Payment Link:</strong> <a href="%s">Click here to pay</a></p>
		`, payment.PaymentURL)
	}

	body += `
			<p>Best regards,<br>Viskatera Team</p>
		</body>
		</html>
	`

	return body
}

// generatePaymentSuccessEmailBody generates HTML email body for payment success
func generatePaymentSuccessEmailBody(purchase models.VisaPurchase, user models.User, payment models.Payment) string {
	return fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; padding: 20px;">
			<h2>Payment Successful!</h2>
			<p>Dear %s,</p>
			<p>Your payment has been successfully processed. Please find your invoice attached.</p>
			<table style="border-collapse: collapse; width: 100%%; margin: 20px 0;">
				<tr>
					<td style="padding: 10px; border: 1px solid #ddd;"><strong>Purchase ID</strong></td>
					<td style="padding: 10px; border: 1px solid #ddd;">#%d</td>
				</tr>
				<tr>
					<td style="padding: 10px; border: 1px solid #ddd;"><strong>Visa</strong></td>
					<td style="padding: 10px; border: 1px solid #ddd;">%s - %s</td>
				</tr>
				<tr>
					<td style="padding: 10px; border: 1px solid #ddd;"><strong>Amount Paid</strong></td>
					<td style="padding: 10px; border: 1px solid #ddd;">Rp %.2f</td>
				</tr>
				<tr>
					<td style="padding: 10px; border: 1px solid #ddd;"><strong>Payment Method</strong></td>
					<td style="padding: 10px; border: 1px solid #ddd;">%s</td>
				</tr>
			</table>
			<p>Your visa application is now being processed. We will notify you once it's ready.</p>
			<p>Best regards,<br>Viskatera Team</p>
		</body>
		</html>
	`, user.Name, purchase.ID, purchase.Visa.Country, purchase.Visa.Type, payment.Amount, payment.PaymentMethod)
}
