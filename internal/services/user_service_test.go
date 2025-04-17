package services_test

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq"
)

func setupDB() *sql.DB {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
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
		log.Fatalf("failed to connect to test database: %v", err)
	}

	return db
}

func TestCreateUser_Success(t *testing.T) {
	db := setupDB()
	defer db.Close()

	testRepo := database.NewRepository(db)
	cleanDB(testRepo)

	logger.Init()
	userService := services.NewUserService(testRepo)

	user, err := userService.CreateUser("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.WithinDuration(t, time.Now().UTC(), user.CreatedAt.UTC(), time.Second*2)
}

func TestCreateUser_Error(t *testing.T) {
	db := setupDB()
	defer db.Close()

	testRepo := database.NewRepository(db)
	cleanDB(testRepo)

	logger.Init()
	userService := services.NewUserService(testRepo)

	firstUser, firstError := userService.CreateUser("testuser", "dupe@example.com", "password123")
	assert.NoError(t, firstError)
	assert.NotNil(t, firstUser)

	user, err := userService.CreateUser("another", "dupe@example.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, user)
}

func cleanDB(db *database.Repository) {
	_, err := db.Exec("TRUNCATE users RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("failed to clean test database: %v", err)
	}
}
