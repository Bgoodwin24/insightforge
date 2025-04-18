package email_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/Bgoodwin24/insightforge/internal/email"
	"github.com/Bgoodwin24/insightforge/internal/handlers"
	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

type Mailer interface {
	SendVerificationEmail(email, username, verificationLink string) error
}

var testDB *sql.DB

func TestMain(m *testing.M) {
	testDB = setupDB()
	defer testDB.Close()

	testRepo := database.NewRepository(testDB)
	cleanDB(testRepo)

	os.Exit(m.Run())
}

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

func TestSendVerificationEmail(t *testing.T) {
	testRepo := database.NewRepository(testDB)
	cleanDB(testRepo)

	mailer := email.NewMailer(
		os.Getenv("SMTP_SERVER"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("INSIGHTFORGE_SMTP_PASSWORD"),
		"noreply@insightforge.com",
	)

	logger.Init()

	userService := services.NewUserService(testRepo)
	userHandler := handlers.NewUserHandler(userService, mailer)

	router := gin.Default()
	router.POST("/register", userHandler.RegisterUser)

	body := `{"username":"realmailertest","email":"noreply.insightforge@gmail.com","password":"StrongPass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Optionally print body or log something
	t.Log("Response:", w.Body.String())
}

func cleanDB(db *database.Repository) {
	_, err := db.Exec("TRUNCATE users RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("failed to clean test database: %v", err)
	}

	_, err = testDB.Exec("TRUNCATE pending_users RESTART IDENTITY CASCADE")
	if err != nil {
		log.Fatalf("failed to clean test database: %v", err)
	}
}
