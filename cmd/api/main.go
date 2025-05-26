package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"net/http"

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

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

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
		DB:      db,
		Queries: database.New(db),
	}

	// Set up email configuration (you can load these from environment variables)
	emailConfig := email.EmailConfig{
		SMTPServer:   os.Getenv("SMTP_SERVER"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("INSIGHTFORGE_SMTP_PASSWORD"),
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
	router.Use(CORSMiddleware())

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
		userGroup.POST("/verify", userHandler.VerifyEmail)
		userGroup.POST("/set-verify-cookie", userHandler.SetVerifyTokenCookie)
		userGroup.POST("/login", userHandler.LoginUser)
		userGroup.POST("/logout", userHandler.LogoutUser)
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
	analyticsHandler := &handlers.AnalyticsHandler{
		Service:        datasetService,
		DatasetService: datasetService,
	}
	analyticsGroup := router.Group("/analytics")
	analyticsGroup.Use(auth.AuthMiddleware(jwtManager))
	{
		// Descriptives
		descriptivesGroup := analyticsGroup.Group("/descriptives")

		descriptivesGroup.GET("/mean", analyticsHandler.MeanHandler)
		descriptivesGroup.GET("/median", analyticsHandler.MedianHandler)
		descriptivesGroup.GET("/mode", analyticsHandler.ModeHandler)
		descriptivesGroup.GET("/stddev", analyticsHandler.StdDevHandler)
		descriptivesGroup.GET("/variance", analyticsHandler.VarianceHandler)
		descriptivesGroup.GET("/min", analyticsHandler.MinHandler)
		descriptivesGroup.GET("/max", analyticsHandler.MaxHandler)
		descriptivesGroup.GET("/range", analyticsHandler.RangeHandler)
		descriptivesGroup.GET("/sum", analyticsHandler.SumHandler)
		descriptivesGroup.GET("/count", analyticsHandler.CountHandler)

		// Correlation
		correlationGroup := analyticsGroup.Group("/correlation")

		correlationGroup.GET("/pearson-correlation", datasetHandler.PearsonHandler)
		correlationGroup.GET("/spearman-correlation", datasetHandler.SpearmanHandler)
		correlationGroup.GET("/correlation-matrix", datasetHandler.CorrelationMatrixHandler)

		// Outliers
		outliersGroup := analyticsGroup.Group("/outliers")

		outliersGroup.GET("/zscore-outliers", analyticsHandler.ZScoreOutliersHandler)
		outliersGroup.GET("/iqr-outliers", analyticsHandler.IQROutliersHandler)
		outliersGroup.GET("/boxplot", analyticsHandler.BoxPlotHandler)

		// Aggregation
		aggregationGroup := analyticsGroup.Group("/aggregation")

		aggregationGroup.GET("/grouped-sum", analyticsHandler.GroupedSumHandler)
		aggregationGroup.GET("/grouped-mean", analyticsHandler.GroupedMeanHandler)
		aggregationGroup.GET("/grouped-count", analyticsHandler.GroupedCountHandler)
		aggregationGroup.GET("/grouped-min", analyticsHandler.GroupedMinHandler)
		aggregationGroup.GET("/grouped-max", analyticsHandler.GroupedMaxHandler)
		aggregationGroup.GET("/grouped-median", analyticsHandler.GroupedMedianHandler)
		aggregationGroup.GET("/grouped-stddev", analyticsHandler.GroupedStdDevHandler)
		aggregationGroup.GET("/pivot-sum", analyticsHandler.PivotTableHandler)
		aggregationGroup.GET("/pivot-mean", analyticsHandler.PivotTableHandler)
		aggregationGroup.GET("/pivot-count", analyticsHandler.PivotTableHandler)
		aggregationGroup.GET("/pivot-min", analyticsHandler.PivotTableHandler)
		aggregationGroup.GET("/pivot-max", analyticsHandler.PivotTableHandler)
		aggregationGroup.GET("/pivot-median", analyticsHandler.PivotTableHandler)
		aggregationGroup.GET("/pivot-stddev", analyticsHandler.PivotTableHandler)

		// Distribution
		distributionGroup := analyticsGroup.Group("/distribution")

		distributionGroup.GET("/histogram", analyticsHandler.HistogramHandler)
		distributionGroup.GET("/kde", analyticsHandler.KDEHandler)

		// Filter/Sort
		filtersortGroup := analyticsGroup.Group("/filtersort")

		filtersortGroup.GET("/filter-sort", datasetHandler.FilterSortHandler)

		// Cleaning
		cleaningGroup := analyticsGroup.Group("/cleaning")

		cleaningGroup.POST("/drop-rows-with-missing", datasetHandler.DropRowsWithMissingHandler)
		cleaningGroup.POST("/fill-missing-with", datasetHandler.FillMissingWithHandler)
		cleaningGroup.POST("/apply-log-transformation", datasetHandler.ApplyLogTransformationHandler)
		cleaningGroup.POST("/normalize-column", datasetHandler.NormalizeColumnHandler)
		cleaningGroup.POST("/standardize-column", datasetHandler.StandardizeColumnHandler)
		cleaningGroup.POST("/drop-columns/:datasetID", datasetHandler.DropColumnsHandler)
		cleaningGroup.POST("/rename-columns/:datasetID", datasetHandler.RenameColumnsHandler)
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
