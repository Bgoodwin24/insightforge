package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	logger.Init()

	if err := godotenv.Load(); err != nil {
		logger.Logger.Fatalf("Failed to load .env file")
	}

	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Logger.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Database connected successfully:", db != nil)

	queries := db.New(db)

	ctx := context.Background()

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "OK"})
	})
	router.GET("/verify", userHandler.VerifyEmail)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		logger.Logger.Fatalf("Failed to run server: %v", err)
	}

}
