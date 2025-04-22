package services_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/auth"
	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestLoginUser(t *testing.T) {
	db := setupDB()
	defer db.Close()

	testRepo := database.NewRepository(db)
	cleanDB(testRepo)
	logger.Init()

	userService := services.NewUserService(testRepo)

	email := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())
	password := "P@ssw0rd!"
	hashedPassword, _ := auth.HashPassword(password)
	userID := uuid.New()
	now := time.Now().UTC()

	_, err := testRepo.Queries.CreateUser(context.Background(), database.CreateUserParams{
		ID:           userID,
		CreatedAt:    now,
		UpdatedAt:    now,
		Email:        email,
		Username:     "testuser",
		PasswordHash: hashedPassword,
	})
	require.NoError(t, err)

	user, token, err := userService.LoginUser(email, password)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, email, user.Email)

	assert.NotEmpty(t, token, "JWT token should not be empty")

	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"), time.Hour*24)
	claims, err := jwtManager.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), claims["user_id"].(string), "JWT should contain correct user ID")
	expFloat := claims["exp"].(float64)
	expTime := time.Unix(int64(expFloat), 0)
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), expTime, time.Minute)
}

func TestActivateUser(t *testing.T) {
	db := setupDB()
	defer db.Close()

	testRepo := database.NewRepository(db)
	cleanDB(testRepo)
	logger.Init()

	userService := services.NewUserService(testRepo)

	pendingUser, err := userService.RegisterPendingUser("testuser", "test@example.com", "P@ssw0rd!")
	if err != nil {
		t.Logf("Failed to register pending user: %v", err)
	}

	err = userService.ActivateUser(pendingUser.Token)
	assert.NoError(t, err)

	var userExists bool
	query := `SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)`
	err = db.QueryRow(query, pendingUser.Email).Scan(&userExists)
	assert.NoError(t, err)
	assert.True(t, userExists, "User should be created")

	var pendingCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM pending_users WHERE id = $1`, pendingUser.ID).Scan(&pendingCount)
	assert.NoError(t, err)
	assert.Equal(t, 0, pendingCount, "Pending user should be deleted")
}

func TestRegisterPendingUser(t *testing.T) {
	username := "testuser"
	email := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())
	password := "P@ssw0rd!"
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		t.Logf("Failed to hash password: %v", err)
	}
	token := uuid.New().String()
	created_at := time.Now().UTC()
	expires_at := time.Now().UTC().Add(24 * time.Hour)
	id := uuid.New()

	_ = database.CreatePendingUserParams{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		Token:        token,
		CreatedAt:    created_at,
		ExpiresAt:    expires_at,
	}
	db := setupDB()
	defer db.Close()

	testRepo := database.NewRepository(db)
	cleanDB(testRepo)

	logger.Init()

	userService := services.NewUserService(testRepo)

	registered, err := userService.RegisterPendingUser(username, email, password)
	if err != nil {
		t.Log(err)
	}

	assert.Equal(t, username, registered.Username)
	assert.Equal(t, email, registered.Email)
	assert.True(t, auth.VerifyPassword(password, registered.PasswordHash))
	assert.NotEmpty(t, registered.Token)
	assert.WithinDuration(t, time.Now().UTC(), registered.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now().UTC().Add(24*time.Hour), registered.ExpiresAt, time.Second)
}

func cleanDB(repo *database.Repository) {
	_, err := repo.Exec("TRUNCATE users RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("failed to clean test database: %v", err)
	}

	_, err = repo.Exec("TRUNCATE pending_users RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("failed to clean test database: %v", err)
	}
}
