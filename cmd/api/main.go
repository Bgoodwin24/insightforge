package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/auth"
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

	// Initialize JWT Manager
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		logger.Logger.Fatal("JWT_SECRET is not set in environment variables")
	}
	tokenDuration := 24 * time.Hour
	jwtManager := &auth.JWTManager{
		SecretKey:     jwtSecret,
		TokenDuration: tokenDuration,
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
	userHandler := handlers.NewUserHandler(userService, mailer, jwtManager)
	userGroup := router.Group("/user")
	{
		userGroup.POST("/register", userHandler.RegisterUser)
		userGroup.GET("/verify", userHandler.VerifyEmail)
		userGroup.POST("/login", userHandler.LoginUser)
		userGroup.GET("/profile", auth.AuthMiddleware(jwtManager), userHandler.GetMyProfile)
	}

	// Dataset routes
	datasetService := services.NewDatasetService(repo)
	datasetHandler := handlers.NewDatasetHandler(datasetService)
	datasetGroup := router.Group("/datasets")
	datasetGroup.Use(auth.AuthMiddleware(jwtManager))
	{
		datasetGroup.POST("/upload", datasetHandler.UploadDataset)
		datasetGroup.GET("/", datasetHandler.ListDatasets)
		datasetGroup.GET("/:id", datasetHandler.GetDatasetByID)
		datasetGroup.DELETE("/:id", datasetHandler.DeleteDatasetsByID)
		datasetGroup.PUT("/:id", datasetHandler.UpdateDataset)
		datasetGroup.POST("/", datasetHandler.CreateDataset)
		datasetGroup.GET("/search", datasetHandler.SearchDataSets)
	}

	// Analytics routes
	analyticsGroup := router.Group("/analytics")
	analyticsGroup.Use(auth.AuthMiddleware(jwtManager))
	{
		// Descriptives
		descriptivesGroup := analyticsGroup.Group("/descriptives")

		descriptivesGroup.POST("/mean", handlers.MeanHandler)
		descriptivesGroup.POST("/median", handlers.MedianHandler)
		descriptivesGroup.POST("/mode", handlers.ModeHandler)
		descriptivesGroup.POST("/stddev", handlers.StdDevHandler)
		descriptivesGroup.POST("/variance", handlers.VarianceHandler)
		descriptivesGroup.POST("/min", handlers.MinHandler)
		descriptivesGroup.POST("/max", handlers.MaxHandler)
		descriptivesGroup.POST("/range", handlers.RangeHandler)
		descriptivesGroup.POST("/sum", handlers.SumHandler)
		descriptivesGroup.POST("/count", handlers.CountHandler)

		// Correlation
		correlationGroup := analyticsGroup.Group("/correlation")

		correlationGroup.POST("/pearson-correlation", handlers.PearsonHandler)
		correlationGroup.POST("/spearman-correlation", handlers.SpearmanHandler)
		correlationGroup.POST("/correlation-matrix", func(c *gin.Context) {
			handlers.CorrelationMatrixHandler(c, datasetService)
		})

		// Outliers
		outliersGroup := analyticsGroup.Group("/outliers")

		outliersGroup.POST("/zscore-outliers", handlers.ZScoreOutliersHandler)
		outliersGroup.POST("/iqr-outliers", handlers.IQROutliersHandler)
		outliersGroup.POST("/boxplot", handlers.BoxPlotHandler)

		// Aggregation
		aggregationGroup := analyticsGroup.Group("/aggregation")

		aggregationGroup.POST("/group-by", handlers.GroupByHandler)
		aggregationGroup.POST("/grouped-sum", handlers.GroupedSumHandler)
		aggregationGroup.POST("/grouped-mean", handlers.GroupedMeanHandler)
		aggregationGroup.POST("/grouped-count", handlers.GroupedCountHandler)
		aggregationGroup.POST("/grouped-min", handlers.GroupedMinHandler)
		aggregationGroup.POST("/grouped-max", handlers.GroupedMaxHandler)
		aggregationGroup.POST("/grouped-median", handlers.GroupedMedianHandler)
		aggregationGroup.POST("/grouped-stddev", handlers.GroupedStdDevHandler)
		aggregationGroup.GET("/:datasetID/pivot", handlers.PivotTableHandler)

		// Distribution
		distributionGroup := analyticsGroup.Group("/distribution")

		distributionGroup.POST("/histogram", handlers.HistogramHandler)
		distributionGroup.POST("/kde", handlers.KDEHandler)

		// Filter/Sort
		filtersortGroup := analyticsGroup.Group("/filtersort")

		filtersortGroup.POST("/filter-sort", handlers.FilterSortHandler)

		// Cleaning
		cleaningGroup := analyticsGroup.Group("/cleaning")

		cleaningGroup.POST("/drop-rows-with-missing/:datasetID", func(c *gin.Context) {
			handlers.DropRowsWithMissingHandler(c, datasetService)
		})
		cleaningGroup.POST("/fill-missing-with/:datasetID", func(c *gin.Context) {
			handlers.FillMissingWithHandler(c, datasetService)
		})
		cleaningGroup.POST("/apply-log-transformation/:datasetID", func(c *gin.Context) {
			handlers.ApplyLogTransformationHandler(c, datasetService)
		})
		cleaningGroup.POST("/normalize-column/:datasetID", func(c *gin.Context) {
			handlers.NormalizeColumnHandler(c, datasetService)
		})
		cleaningGroup.POST("/standardize-column/:datasetID", func(c *gin.Context) {
			handlers.StandardizeColumnHandler(c, datasetService)
		})
		cleaningGroup.POST("/drop-columns/:datasetID", func(c *gin.Context) {
			handlers.DropColumnsHandler(c, datasetService)
		})
		cleaningGroup.POST("/rename-columns/:datasetID", func(c *gin.Context) {
			handlers.RenameColumnsHandler(c, datasetService)
		})
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
