package utils

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// SendEmail sends email using SMTP (supports MailHog for development)
func SendEmail(to, subject, body string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	// Default to MailHog if SMTP not configured (development)
	if host == "" {
		host = "localhost"
		port = "1025" // MailHog SMTP port
		from = "noreply@viskatera.com"
		user = "" // MailHog doesn't require auth
		pass = ""
		log.Printf("[EMAIL] Using MailHog at %s:%s", host, port)
	}

	if port == "" {
		port = "1025" // Default MailHog port
	}

	if from == "" {
		from = "noreply@viskatera.com"
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"utf-8\"\r\n\r\n" +
		body)

	var auth smtp.Auth
	if user != "" && pass != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}

	err := smtp.SendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		log.Printf("[EMAIL-ERROR] Failed to send email: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[EMAIL] Email sent successfully to %s", to)
	return nil
}
