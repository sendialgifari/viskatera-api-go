package utils

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
)

// SendEmail sends email using SMTP (supports MailHog for development)
func SendEmail(to, subject, body string) error {
	return SendEmailWithAttachment(to, subject, body, "")
}

// SendEmailWithAttachment sends email with optional PDF attachment
func SendEmailWithAttachment(to, subject, body, pdfPath string) error {
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

	// Build email message
	boundary := "----=_NextPart_" + fmt.Sprintf("%d", os.Getpid())

	msg := "To: " + to + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/mixed; boundary=\"" + boundary + "\"\r\n\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/html; charset=\"utf-8\"\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n\r\n" +
		body + "\r\n"

	// Add PDF attachment if provided
	if pdfPath != "" {
		pdfData, err := ioutil.ReadFile(pdfPath)
		if err != nil {
			log.Printf("[EMAIL-ERROR] Failed to read PDF file: %v", err)
			return fmt.Errorf("failed to read PDF file: %w", err)
		}

		filename := filepath.Base(pdfPath)
		encoded := base64.StdEncoding.EncodeToString(pdfData)

		msg += "--" + boundary + "\r\n" +
			"Content-Type: application/pdf\r\n" +
			"Content-Transfer-Encoding: base64\r\n" +
			"Content-Disposition: attachment; filename=\"" + filename + "\"\r\n\r\n" +
			encoded + "\r\n"
	}

	msg += "--" + boundary + "--"

	var auth smtp.Auth
	if user != "" && pass != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}

	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("[EMAIL-ERROR] Failed to send email: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[EMAIL] Email sent successfully to %s", to)
	return nil
}
