package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"viskatera-api-go/config"
	"viskatera-api-go/models"
	"viskatera-api-go/utils"

	"github.com/gin-gonic/gin"
)

// XenditWebhookPayload represents Xendit webhook payload
type XenditWebhookPayload struct {
	ID         string  `json:"id"`
	ExternalID string  `json:"external_id"`
	Status     string  `json:"status"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Created    string  `json:"created"`
	Updated    string  `json:"updated"`
}

// XenditWebhook handles Xendit payment webhook
// @Summary Xendit payment webhook
// @Description Handle Xendit payment webhook notifications. Automatically updates payment status, purchase status, and sends payment success email with PDF invoice via RabbitMQ when payment is successful.
// @Tags Webhook
// @Accept json
// @Produce json
// @Param payload body XenditWebhookPayload true "Xendit webhook payload"
// @Success 200 {object} models.APIResponse "Webhook processed successfully"
// @Failure 400 {object} models.APIResponse "Invalid webhook payload"
// @Failure 404 {object} models.APIResponse "Payment not found"
// @Failure 500 {object} models.APIResponse "Internal server error"
// @Router /webhooks/xendit [post]
func XenditWebhook(c *gin.Context) {
	var payload XenditWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid webhook payload",
			"VALIDATION_ERROR",
			err.Error(),
		))
		return
	}

	log.Printf("[WEBHOOK] Received Xendit webhook: ID=%s, Status=%s, ExternalID=%s", payload.ID, payload.Status, payload.ExternalID)

	// Find payment by Xendit ID
	var payment models.Payment
	if err := config.DB.Where("xendit_id = ?", payload.ID).First(&payment).Error; err != nil {
		log.Printf("[WEBHOOK] Payment not found for Xendit ID: %s", payload.ID)
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Payment not found",
			"PAYMENT_NOT_FOUND",
			"Payment with this Xendit ID does not exist",
		))
		return
	}

	// Update payment status
	oldStatus := payment.Status
	payment.Status = payload.Status

	// Handle payment success
	if payload.Status == "PAID" || payload.Status == "paid" {
		payment.Status = "paid"

		// Update purchase status to completed
		var purchase models.VisaPurchase
		if err := config.DB.First(&purchase, payment.PurchaseID).Error; err == nil {
			oldPurchaseStatus := purchase.Status
			purchase.Status = "completed"
			config.DB.Save(&purchase)

			// Log purchase status update
			purchaseEntityName := "Purchase #" + strconv.Itoa(int(purchase.ID))
			oldPurchaseValues := map[string]interface{}{"status": oldPurchaseStatus}
			newPurchaseValues := map[string]interface{}{"status": purchase.Status}
			utils.LogUpdate(c, 0, models.EntityPurchase, purchase.ID, purchaseEntityName, oldPurchaseValues, newPurchaseValues)

			// Get user details
			var user models.User
			if err := config.DB.First(&user, payment.UserID).Error; err == nil {
				// Send payment success email with PDF via RabbitMQ
				job := map[string]interface{}{
					"purchase_id": purchase.ID,
					"user_id":     user.ID,
					"email":       user.Email,
					"type":        "payment_success",
				}
				jobJSON, _ := json.Marshal(job)
				if err := config.PublishMessage(config.QueueEmailPaymentSuccess, jobJSON); err != nil {
					log.Printf("[WEBHOOK] Failed to publish payment success email job: %v", err)
				} else {
					log.Printf("[WEBHOOK] Payment success email job published for purchase ID: %d", purchase.ID)
				}
			}
		}
	} else if payload.Status == "EXPIRED" || payload.Status == "expired" {
		payment.Status = "expired"
	}

	// Save payment status
	config.DB.Save(&payment)

	// Log payment status update
	if oldStatus != payment.Status {
		paymentEntityName := "Payment #" + strconv.Itoa(int(payment.ID)) + " - " + payment.PaymentMethod
		oldValues := map[string]interface{}{"status": oldStatus}
		newValues := map[string]interface{}{"status": payment.Status}
		utils.LogUpdate(c, 0, models.EntityPayment, payment.ID, paymentEntityName, oldValues, newValues)
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Webhook processed successfully",
		gin.H{
			"payment_id": payment.ID,
			"status":     payment.Status,
		},
	))
}
