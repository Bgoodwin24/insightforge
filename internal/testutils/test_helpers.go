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

func InsertTestRecord(t *testing.T, repo *database.Repository, datasetID, fieldID uuid.UUID, value string) {
	recordID := uuid.New()
	now := time.Now()

	_, err := repo.DB.Exec(`
        INSERT INTO dataset_records (id, dataset_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4)
    `, recordID, datasetID, now, now)
	require.NoError(t, err)

	_, err = repo.DB.Exec(`
        INSERT INTO record_values (record_id, field_id, value)
        VALUES ($1, $2, $3)
    `, recordID, fieldID, value)
	require.NoError(t, err)
}

func InsertTestField(t *testing.T, repo *database.Repository, datasetID, fieldID uuid.UUID, name, dataType string) {
	t.Helper()
	err := repo.Queries.CreateDatasetField(context.Background(), database.CreateDatasetFieldParams{
		ID:        fieldID,
		DatasetID: datasetID,
		Name:      name,
		DataType:  dataType,
		CreatedAt: time.Now(),
	})
	require.NoError(t, err)
}

func InsertTestValueWithRecordID(t *testing.T, repo *database.Repository, datasetID, recordID, fieldID uuid.UUID, value string) {
	now := time.Now()

	if value == "" {
		fmt.Println("Empty value detected, skipping insertion")
		return
	}

	// Insert dataset_record only if it doesn't exist
	_, err := repo.DB.Exec(`
		INSERT INTO dataset_records (id, dataset_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`, recordID, datasetID, now, now)
	require.NoError(t, err)

	_, err = repo.DB.Exec(`
		INSERT INTO record_values (record_id, field_id, value)
		VALUES ($1, $2, $3)
	`, recordID, fieldID, value)
	require.NoError(t, err)
}

func InsertTestRecordMulti(t *testing.T, repo *database.Repository, datasetID uuid.UUID, values map[uuid.UUID]string) {
	t.Helper()
	recordID := uuid.New()
	now := time.Now()

	_, err := repo.DB.Exec(`
        INSERT INTO dataset_records (id, dataset_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4)
    `, recordID, datasetID, now, now)
	require.NoError(t, err)

	for fieldID, value := range values {
		_, err := repo.DB.Exec(`
            INSERT INTO record_values (record_id, field_id, value)
            VALUES ($1, $2, $3)
        `, recordID, fieldID, value)
		require.NoError(t, err)
	}
}
