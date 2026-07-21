package main

import (
	"log"
	"os"

	"auth-service/config"
	"auth-service/controllers"
	"auth-service/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Connect to MongoDB
	config.ConnectDB()

	// Initialize Gin router
	r := gin.Default()

	// CORS Middleware (simplified for development)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	// Health and Root Routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the AI PDF Assistant Auth Service!",
			"status":  "running",
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// Routes
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/signup", controllers.Signup)
			auth.POST("/login", controllers.Login)
			auth.GET("/verify", controllers.VerifyEmail)
			auth.POST("/forgot-password", controllers.ForgotPassword)
			auth.POST("/reset-password", controllers.ResetPassword)
		}
		
		user := api.Group("/user")
		user.Use(middleware.RequireAuth)
		{
			user.GET("/profile", controllers.GetProfile)
			user.PUT("/profile", controllers.UpdateProfile)
			user.PUT("/password", controllers.ChangePassword)
		}
	}

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}
	log.Printf("Starting Auth Service on port %s...\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
