package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/Bgoodwin24/insightforge/internal/email"
	"github.com/Bgoodwin24/insightforge/internal/handlers"
	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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
	dbMaxUpload := os.Getenv("MAX_UPLOAD_SIZE")

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
	emailConfig := email.EmailConfig{
		SMTPServer:   os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
	}

	mailer := email.NewMailer(
		emailConfig.SMTPServer,
		emailConfig.SMTPPort,
		emailConfig.SMTPUsername,
		emailConfig.SMTPPassword,
		emailConfig.FromEmail,
	)

	router := gin.Default()

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})

	// Max dataset upload
	uploadLimit, err := strconv.ParseInt(dbMaxUpload, 10, 64)
	if err != nil {
		logger.Logger.Fatalf("Invalid MAX_UPLOAD_SIZE: %v", err)
	}

	if uploadLimit == 0 {
		uploadLimit = 50 << 20
	}

	router.MaxMultipartMemory = uploadLimit

	// User Service routes
	userService := services.NewUserService(repo)
	userHandler := handlers.NewUserHandler(userService, mailer)
	userGroup := router.Group("/user")
	{
		userGroup.POST("/register", userHandler.RegisterUser)
		userGroup.GET("/verify", userHandler.VerifyEmail)
		userGroup.POST("/login", userHandler.LoginUser)
	}

	// Dataset routes
	datasetService := services.NewDatasetService(repo)
	datasetHandler := handlers.NewDatasetHandler(datasetService)
	datasetGroup := router.Group("/datasets")
	{
		datasetGroup.POST("/upload", datasetHandler.UploadDataset)
		datasetGroup.GET("/", datasetHandler.ListDatasets)
		datasetGroup.GET("/:id", datasetHandler.GetDatasetByID)
		datasetGroup.DELETE("/:id", datasetHandler.DeleteDatasetsByID)
		datasetGroup.PUT("/:id", datasetHandler.UpdateDataset)
		datasetGroup.POST("/", datasetHandler.CreateDataset)
		datasetGroup.GET("/search", datasetHandler.SearchDataSets)
	}

	// Get the port from environment or default to 8080
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("InsightForge API running on port %s", port)

	// Start the Gin server
	if err := router.Run(":" + port); err != nil {
		logger.Logger.Fatalf("Failed to run server: %v", err)
	}

}
