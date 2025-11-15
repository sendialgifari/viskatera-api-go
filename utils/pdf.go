package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"viskatera-api-go/models"

	"github.com/jung-kurt/gofpdf"
)

// InvoiceData represents data for invoice generation
type InvoiceData struct {
	InvoiceNumber string
	Date          time.Time
	CustomerName  string
	CustomerEmail string
	Items         []InvoiceItem
	Subtotal      float64
	Total         float64
	PaymentMethod string
	Status        string
}

// InvoiceItem represents an item in the invoice
type InvoiceItem struct {
	Description string
	Quantity    int
	Price       float64
	Total       float64
}

// GenerateInvoicePDF generates a PDF invoice
func GenerateInvoicePDF(data InvoiceData) (string, error) {
	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "VISKATERA")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Invoice")
	pdf.Ln(5)

	// Invoice details
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 8, fmt.Sprintf("Invoice Number: %s", data.InvoiceNumber))
	pdf.Ln(5)
	pdf.Cell(40, 8, fmt.Sprintf("Date: %s", data.Date.Format("January 2, 2006")))
	pdf.Ln(10)

	// Customer details
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Bill To:")
	pdf.Ln(5)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 8, data.CustomerName)
	pdf.Ln(5)
	pdf.Cell(40, 8, data.CustomerEmail)
	pdf.Ln(15)

	// Items table
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(100, 8, "Description", "1", 0, "L", false, 0, "")
	pdf.CellFormat(30, 8, "Quantity", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 8, "Price", "1", 0, "R", false, 0, "")
	pdf.CellFormat(30, 8, "Total", "1", 0, "R", false, 0, "")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	for _, item := range data.Items {
		pdf.CellFormat(100, 8, item.Description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("Rp %.2f", item.Price), "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("Rp %.2f", item.Total), "1", 0, "R", false, 0, "")
		pdf.Ln(8)
	}

	// Totals
	pdf.Ln(5)
	pdf.CellFormat(130, 8, "", "", 0, "", false, 0, "")
	pdf.CellFormat(30, 8, "Subtotal:", "1", 0, "R", false, 0, "")
	pdf.CellFormat(30, 8, fmt.Sprintf("Rp %.2f", data.Subtotal), "1", 0, "R", false, 0, "")
	pdf.Ln(8)

	pdf.CellFormat(130, 8, "", "", 0, "", false, 0, "")
	pdf.CellFormat(30, 8, "Total:", "1", 0, "R", false, 0, "")
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(30, 8, fmt.Sprintf("Rp %.2f", data.Total), "1", 0, "R", false, 0, "")
	pdf.Ln(15)

	// Payment info
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 8, fmt.Sprintf("Payment Method: %s", data.PaymentMethod))
	pdf.Ln(5)
	pdf.Cell(40, 8, fmt.Sprintf("Status: %s", data.Status))
	pdf.Ln(10)

	// Footer
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(0, 10, "Thank you for your business!", "", 0, "C", false, 0, "")

	// Save PDF
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	invoicesDir := filepath.Join(uploadDir, "invoices")
	if err := os.MkdirAll(invoicesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create invoices directory: %w", err)
	}

	filename := fmt.Sprintf("invoice_%s_%d.pdf", data.InvoiceNumber, time.Now().Unix())
	filepath := filepath.Join(invoicesDir, filename)

	if err := pdf.OutputFileAndClose(filepath); err != nil {
		return "", fmt.Errorf("failed to save PDF: %w", err)
	}

	return filepath, nil
}

// GeneratePurchaseInvoicePDF generates invoice PDF for a purchase
func GeneratePurchaseInvoicePDF(purchase models.VisaPurchase, user models.User, payment models.Payment) (string, error) {
	invoiceNumber := fmt.Sprintf("INV-%d-%d", purchase.ID, time.Now().Unix())

	items := []InvoiceItem{
		{
			Description: fmt.Sprintf("%s Visa - %s", purchase.Visa.Country, purchase.Visa.Type),
			Quantity:    1,
			Price:       purchase.Visa.Price,
			Total:       purchase.Visa.Price,
		},
	}

	if purchase.VisaOption != nil {
		items = append(items, InvoiceItem{
			Description: purchase.VisaOption.Name,
			Quantity:    1,
			Price:       purchase.VisaOption.Price,
			Total:       purchase.VisaOption.Price,
		})
	}

	data := InvoiceData{
		InvoiceNumber: invoiceNumber,
		Date:          time.Now(),
		CustomerName:  user.Name,
		CustomerEmail: user.Email,
		Items:         items,
		Subtotal:      purchase.TotalPrice,
		Total:         purchase.TotalPrice,
		PaymentMethod: payment.PaymentMethod,
		Status:        payment.Status,
	}

	return GenerateInvoicePDF(data)
}
