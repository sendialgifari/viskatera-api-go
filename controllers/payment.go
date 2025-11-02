package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
	"viskatera-api-go/config"
	"viskatera-api-go/models"
	"viskatera-api-go/utils"

	"github.com/gin-gonic/gin"
)

type XenditRequest struct {
	PurchaseID    uint    `json:"purchase_id" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=virtual_account qris"`
	Amount        float64 `json:"amount"`
	BankCode      string  `json:"bank_code,omitempty"` // For virtual account
	CustomerName  string  `json:"customer_name,omitempty"`
	CustomerEmail string  `json:"customer_email,omitempty"`
}

type XenditInvoiceRequest struct {
	ExternalID                     string                       `json:"external_id"`
	Amount                         float64                      `json:"amount"`
	Description                    string                       `json:"description"`
	Currency                       string                       `json:"currency"`
	Customer                       *XenditCustomer              `json:"customer,omitempty"`
	CustomerNotificationPreference XenditNotificationPreference `json:"customer_notification_preference,omitempty"`
	PaymentMethods                 []string                     `json:"payment_methods,omitempty"`
	SuccessRedirectURL             string                       `json:"success_redirect_url,omitempty"`
	FailureRedirectURL             string                       `json:"failure_redirect_url,omitempty"`
	Items                          []XenditItem                 `json:"items,omitempty"`
}

type XenditCustomer struct {
	GivenNames   string `json:"given_names"`
	Email        string `json:"email"`
	MobileNumber string `json:"mobile_number,omitempty"`
}

type XenditNotificationPreference struct {
	InvoiceCreated  []string `json:"invoice_created,omitempty"`
	InvoiceReminder []string `json:"invoice_reminder,omitempty"`
	InvoicePaid     []string `json:"invoice_paid,omitempty"`
}

type XenditItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type XenditResponse struct {
	ID                string   `json:"id"`
	ExternalID        string   `json:"external_id"`
	UserID            string   `json:"user_id"`
	Status            string   `json:"status"`
	MerchantName      string   `json:"merchant_name"`
	Amount            float64  `json:"amount"`
	Currency          string   `json:"currency"`
	InvoiceURL        string   `json:"invoice_url"`
	AvailableBanks    []Bank   `json:"available_banks,omitempty"`
	AvailableEwallets []Wallet `json:"available_ewallets,omitempty"`
}

type Bank struct {
	BankCode string `json:"bank_code"`
	Name     string `json:"name"`
}

type Wallet struct {
	EWalletType string `json:"ewallet_type"`
}

// CreatePayment godoc
// @Summary Create payment with Xendit
// @Description Create payment using virtual account or QRIS
// @Tags Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body XenditRequest true "Payment request"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /payments [post]
func CreatePayment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not authenticated",
			"UNAUTHORIZED",
			"Please login to create payment",
		))
		return
	}

	var req XenditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid request data",
			"VALIDATION_ERROR",
			err.Error(),
		))
		return
	}

	// Get purchase details
	var purchase models.VisaPurchase
	if err := config.DB.Where("id = ? AND user_id = ?", req.PurchaseID, userID).First(&purchase).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Purchase not found",
			"PURCHASE_NOT_FOUND",
			"Purchase with this ID does not exist or does not belong to you",
		))
		return
	}

	// Get user details
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to get user details",
			"DATABASE_ERROR",
			err.Error(),
		))
		return
	}

	// Get payment amount
	amount := req.Amount
	if amount == 0 {
		amount = purchase.TotalPrice
	}

	// Create invoice
	xenditSecretKey := os.Getenv("XENDIT_SECRET_KEY")
	xenditAPIURL := os.Getenv("XENDIT_API_URL")
	if xenditAPIURL == "" {
		xenditAPIURL = "https://api.xendit.co"
	}

	externalID := fmt.Sprintf("payment_%d_%d", purchase.ID, time.Now().Unix())
	description := fmt.Sprintf("Payment for visa purchase %d", purchase.ID)

	customerName := req.CustomerName
	if customerName == "" {
		customerName = user.Name
	}

	customerEmail := req.CustomerEmail
	if customerEmail == "" {
		customerEmail = user.Email
	}

	paymentMethods := []string{}
	if req.PaymentMethod == "virtual_account" {
		paymentMethods = append(paymentMethods, "BANK_TRANSFER")
		if req.BankCode != "" {
			paymentMethods = append(paymentMethods, req.BankCode)
		}
	} else if req.PaymentMethod == "qris" {
		paymentMethods = append(paymentMethods, "EWALLET")
	}

	invoiceReq := XenditInvoiceRequest{
		ExternalID:  externalID,
		Amount:      amount,
		Description: description,
		Currency:    "IDR",
		Customer: &XenditCustomer{
			GivenNames: customerName,
			Email:      customerEmail,
		},
		PaymentMethods: paymentMethods,
		Items: []XenditItem{
			{
				Name:     purchase.Visa.Country + " - " + purchase.Visa.Type,
				Quantity: 1,
				Price:    amount,
			},
		},
	}

	// Call Xendit API
	jsonData, err := json.Marshal(invoiceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to create payment request",
			"MARSHAL_ERROR",
			err.Error(),
		))
		return
	}

	xenditURL := fmt.Sprintf("%s/v2/invoices", xenditAPIURL)
	httpReq, err := http.NewRequest("POST", xenditURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to create payment request",
			"REQUEST_ERROR",
			err.Error(),
		))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(xenditSecretKey, "")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to connect to payment gateway",
			"PAYMENT_GATEWAY_ERROR",
			err.Error(),
		))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to read payment response",
			"READ_ERROR",
			err.Error(),
		))
		return
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Payment gateway error",
			"PAYMENT_ERROR",
			string(body),
		))
		return
	}

	var xenditResp XenditResponse
	if err := json.Unmarshal(body, &xenditResp); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to parse payment response",
			"PARSE_ERROR",
			err.Error(),
		))
		return
	}

	// Save payment record
	payment := models.Payment{
		UserID:        userID.(uint),
		PurchaseID:    purchase.ID,
		PaymentMethod: req.PaymentMethod,
		Amount:        amount,
		Status:        xenditResp.Status,
		XenditID:      xenditResp.ID,
		PaymentURL:    xenditResp.InvoiceURL,
	}

	if err := config.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to save payment record",
			"DATABASE_ERROR",
			err.Error(),
		))
		return
	}

	// Update purchase status to pending
	purchase.Status = "pending"
	config.DB.Save(&purchase)

	// Log activity
	logUserID := utils.GetUserIDFromContextWithDefault(c)
	entityName := "Payment #" + strconv.Itoa(int(payment.ID)) + " - " + payment.PaymentMethod
	utils.LogCreate(c, logUserID, models.EntityPayment, payment.ID, entityName, payment)

	responseData := gin.H{
		"payment":     payment,
		"payment_url": xenditResp.InvoiceURL,
		"xendit_id":   xenditResp.ID,
		"status":      xenditResp.Status,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Payment created successfully",
		responseData,
	))
}

// GetPaymentStatus godoc
// @Summary Get payment status
// @Description Get payment status from Xendit
// @Tags Payment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Payment ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /payments/{id}/status [get]
func GetPaymentStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not authenticated",
			"UNAUTHORIZED",
			"Please login",
		))
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid payment ID",
			"INVALID_ID",
			"Payment ID must be a valid number",
		))
		return
	}

	var payment models.Payment
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&payment).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Payment not found",
			"PAYMENT_NOT_FOUND",
			"Payment with this ID does not exist or does not belong to you",
		))
		return
	}

	// Call Xendit API to get latest status
	xenditSecretKey := os.Getenv("XENDIT_SECRET_KEY")
	xenditAPIURL := os.Getenv("XENDIT_API_URL")
	if xenditAPIURL == "" {
		xenditAPIURL = "https://api.xendit.co"
	}

	xenditURL := fmt.Sprintf("%s/v2/invoices/%s", xenditAPIURL, payment.XenditID)
	httpReq, err := http.NewRequest("GET", xenditURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to create request",
			"REQUEST_ERROR",
			err.Error(),
		))
		return
	}

	httpReq.SetBasicAuth(xenditSecretKey, "")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to connect to payment gateway",
			"PAYMENT_GATEWAY_ERROR",
			err.Error(),
		))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to read payment response",
			"READ_ERROR",
			err.Error(),
		))
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Payment gateway error",
			"PAYMENT_ERROR",
			string(body),
		))
		return
	}

	var xenditResp XenditResponse
	if err := json.Unmarshal(body, &xenditResp); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to parse payment response",
			"PARSE_ERROR",
			err.Error(),
		))
		return
	}

	// Store old status for logging
	oldStatus := payment.Status

	// Update payment status in database
	payment.Status = xenditResp.Status
	if xenditResp.Status == "PAID" || xenditResp.Status == "paid" {
		payment.Status = "paid"

		// Update purchase status
		var purchase models.VisaPurchase
		if err := config.DB.First(&purchase, payment.PurchaseID).Error; err == nil {
			oldPurchaseStatus := purchase.Status
			purchase.Status = "completed"
			config.DB.Save(&purchase)

			// Log purchase status update
			logUserID := utils.GetUserIDFromContextWithDefault(c)
			purchaseEntityName := "Purchase #" + strconv.Itoa(int(purchase.ID))
			oldPurchaseValues := map[string]interface{}{"status": oldPurchaseStatus}
			newPurchaseValues := map[string]interface{}{"status": purchase.Status}
			utils.LogUpdate(c, logUserID, models.EntityPurchase, purchase.ID, purchaseEntityName, oldPurchaseValues, newPurchaseValues)
		}
	} else if xenditResp.Status == "EXPIRED" || xenditResp.Status == "expired" {
		payment.Status = "expired"
	}

	config.DB.Save(&payment)

	config.DB.Preload("User").Preload("Purchase").First(&payment, payment.ID)

	// Log payment status update
	if oldStatus != payment.Status {
		logUserID := utils.GetUserIDFromContextWithDefault(c)
		entityName := "Payment #" + strconv.Itoa(int(payment.ID)) + " - " + payment.PaymentMethod
		oldValues := map[string]interface{}{"status": oldStatus}
		newValues := map[string]interface{}{"status": payment.Status}
		utils.LogUpdate(c, logUserID, models.EntityPayment, payment.ID, entityName, oldValues, newValues)
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Payment status retrieved successfully",
		payment,
	))
}
