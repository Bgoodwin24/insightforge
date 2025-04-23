package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/auth"
	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

var TestDB *sql.DB

func SetupTestRepo() *database.Repository {
	db := SetupDB()
	return &database.Repository{
		DB:      db,
		Queries: database.New(db),
	}
}

func SetupDB() *sql.DB {
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

func CreateTestUser(t *testing.T, repo *database.Repository, email string) database.User {
	password, _ := auth.HashPassword("P@ssw0rd!")
	now := time.Now().UTC()
	username := fmt.Sprintf("testuser_%d", time.Now().UnixNano())

	user, err := repo.Queries.CreateUser(context.Background(), database.CreateUserParams{
		ID:           uuid.New(),
		Email:        email,
		Username:     username,
		PasswordHash: password,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	require.NoError(t, err)
	return user
}

func CreateTestDataset(t *testing.T, repo *database.Repository, userID uuid.UUID, name, description string) database.Dataset {
	id := uuid.New()
	now := time.Now().UTC()

	dataset, err := repo.Queries.CreateDataset(context.Background(), database.CreateDatasetParams{
		ID:          id,
		UserID:      userID,
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	require.NoError(t, err)

	return dataset
}

func GetAuthenticatedContext(userID uuid.UUID, email string) *gin.Context {
	secret := os.Getenv("JWT_SECRET")
	duration := time.Hour * 24
	jwtManager := auth.NewJWTManager(secret, duration)
	token, err := jwtManager.Generate(userID, email)
	if err != nil {
		log.Fatal("failed to generate JWT:" + err.Error())
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	c.Request = req

	auth.AuthMiddleware(jwtManager)(c)

	return c
}

func CleanDB(db *database.Repository) {
	_, err := db.Exec("TRUNCATE users, pending_users, datasets RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("failed to clean test database: %v", err)
	}
}
