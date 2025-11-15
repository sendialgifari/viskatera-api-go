package controllers

import (
	"net/http"
	"viskatera-api-go/config"
	"viskatera-api-go/models"

	"github.com/gin-gonic/gin"
)

// GetQueueStats godoc
// @Summary Get RabbitMQ queue statistics
// @Description Get statistics for all RabbitMQ queues including message counts for email_invoice, email_payment_success, and generate_pdf queues
// @Tags Monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse "Queue statistics retrieved successfully"
// @Failure 401 {object} models.APIResponse "Unauthorized"
// @Failure 500 {object} models.APIResponse "Failed to get queue statistics"
// @Router /monitoring/queues [get]
func GetQueueStats(c *gin.Context) {
	stats, err := config.GetAllQueueStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to get queue statistics",
			"QUEUE_ERROR",
			err.Error(),
		))
		return
	}

	// Get detailed stats for each queue
	detailedStats := make(map[string]interface{})
	for queueName, messageCount := range stats {
		detailedStats[queueName] = gin.H{
			"messages":       messageCount,
			"messages_ready": messageCount,
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Queue statistics retrieved successfully",
		gin.H{
			"queues":       detailedStats,
			"total_queues": len(stats),
		},
	))
}

// GetQueueHealth godoc
// @Summary Get RabbitMQ health status
// @Description Get health status of RabbitMQ connection including connection status and total messages in all queues
// @Tags Monitoring
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse "RabbitMQ is healthy"
// @Failure 401 {object} models.APIResponse "Unauthorized"
// @Failure 503 {object} models.APIResponse "RabbitMQ connection is not available"
// @Failure 500 {object} models.APIResponse "Failed to get queue statistics"
// @Router /monitoring/queues/health [get]
func GetQueueHealth(c *gin.Context) {
	if config.RabbitMQConn == nil || config.RabbitMQConn.IsClosed() {
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse(
			"RabbitMQ connection is not available",
			"QUEUE_UNAVAILABLE",
			"RabbitMQ connection is closed or not initialized",
		))
		return
	}

	stats, err := config.GetAllQueueStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to get queue statistics",
			"QUEUE_ERROR",
			err.Error(),
		))
		return
	}

	// Calculate total messages
	totalMessages := 0
	for _, count := range stats {
		totalMessages += count
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"RabbitMQ is healthy",
		gin.H{
			"status":         "healthy",
			"connected":      true,
			"total_messages": totalMessages,
			"queues":         stats,
		},
	))
}
