package routes

import (
	"net/http"
	"viskatera-api-go/controllers"
	"viskatera-api-go/middleware"
	"viskatera-api-go/models"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	// @Summary Health check
	// @Description Check API health and version
	// @Tags System
	// @Accept json
	// @Produce json
	// @Success 200 {object} models.APIResponse
	// @Router /health [get]
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, models.SuccessResponse(
			"Viskatera API is running",
			gin.H{
				"version": "1.0.0",
				"status":  "healthy",
			},
		))
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API Documentation redirect
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Public routes (no authentication required)
	public := r.Group("/api/v1")
	{
		// Auth routes
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
		public.POST("/auth/request-otp", controllers.RequestOTP)
		public.POST("/auth/verify-otp", controllers.VerifyOTP)
		public.POST("/auth/forgot-password", controllers.ForgotPassword)
		public.POST("/auth/reset-password", controllers.ResetPassword)
		public.GET("/auth/google/login", controllers.GoogleLogin)
		public.GET("/auth/google/callback", controllers.GoogleCallback)

		// Visa routes (public)
		public.GET("/visas", controllers.GetVisas)
		public.GET("/visas/:id", controllers.GetVisaByID)
	}

	// Expose file uploads so they can be accessed in the browser, e.g. /uploads/visas/visa_1_1761988882.png
	// This line makes files in ./uploads folder available at /uploads/*path on the server.
	r.Static("/uploads", "./uploads")

	// Protected routes (authentication required)
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		// Customer routes (both customer and admin can access)
		protected.POST("/purchases", controllers.PurchaseVisa)
		protected.GET("/purchases", controllers.GetUserPurchases)
		protected.GET("/purchases/:id", controllers.GetPurchaseByID)
		protected.PUT("/purchases/:id/status", controllers.UpdatePurchaseStatus)

		// Payment routes
		protected.POST("/payments", controllers.CreatePayment)
		protected.GET("/payments/:id/status", controllers.GetPaymentStatus)

		// Upload routes
		protected.POST("/uploads/avatar", controllers.UploadAvatar)
		protected.POST("/uploads/visa/:visa_id", controllers.UploadVisaDocument)
		// No need: protected.Static("/uploads", "./uploads") // Handled globally above

		// User profile
		protected.PUT("/user", controllers.UpdateUser)

		// Export routes
		protected.GET("/exports/visas/excel", controllers.ExportVisasExcel)
		protected.GET("/exports/visas/pdf", controllers.ExportVisasPDF)
		protected.GET("/exports/purchases/excel", controllers.ExportPurchasesExcel)
		protected.GET("/exports/purchases/pdf", controllers.ExportPurchasesPDF)

		// Activity log routes
		protected.GET("/activities", controllers.GetUserActivities)
		protected.GET("/activities/visa/:visa_id", controllers.GetVisaActivities)
		protected.GET("/activities/purchase/:purchase_id", controllers.GetPurchaseActivities)
		protected.GET("/activities/payment/:payment_id", controllers.GetPaymentActivities)
	}

	// Admin routes (admin authentication required)
	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		// Visa management (admin only)
		admin.POST("/visas", controllers.CreateVisa)
		admin.PUT("/visas/:id", controllers.UpdateVisa)
		admin.DELETE("/visas/:id", controllers.DeleteVisa)
	}

	return r
}
