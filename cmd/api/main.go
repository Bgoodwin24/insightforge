package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/Bgoodwin24/insightforge/internal/handlers"
	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Mailer interface {
	SendVerificationEmail(email, username, verificationLink string) error
}

func main() {
	// Initialize logger
	logger.Init()

	// Load .env file
	err := godotenv.Load(".env")
	if err != nil {
		logger.Logger.Fatalf("Failed to load .env file: %v", err)
	}

	//Retrieve environment variables
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	// Create DSN string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	// Open the database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Ping the database to ensure it's reachable
	if err := db.Ping(); err != nil {
		logger.Logger.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Database connected successfully:", db != nil)

	// Initialize repository with DB connection
	repo := &database.Repository{
		DB: db,
	}

	// Set up email configuration (you can load these from environment variables)
	emailConfig := handlers.EmailConfig{
		SMTPServer:   os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
	}

	userService := services.NewUserService(repo)

	userHandler := handlers.NewUserHandler(userService, &emailConfig)

	router := gin.Default()

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	// Registration and email verification routes
	router.POST("/register", userHandler.RegisterUser)
	router.GET("/verify", userHandler.VerifyEmail)

	// Get the port from environment or default to 8080
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	// Start the Gin server
	if err := router.Run(":" + port); err != nil {
		logger.Logger.Fatalf("Failed to run server: %v", err)
	}

}
