package handlers_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/auth"
	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/Bgoodwin24/insightforge/internal/handlers"
	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/Bgoodwin24/insightforge/internal/testutils"
	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq"
)

type MockMailer struct {
	Called              bool
	LastEmail           string
	LastUsername        string
	LastVerificationURL string
}

func (m *MockMailer) SendVerificationEmail(email, username, verificationLink string) error {
	// Mock the email sending without actually sending it
	// You can log or validate the parameters here as needed
	if email == "" || username == "" || verificationLink == "" {
		return fmt.Errorf("invalid parameters")
	}
	m.Called = true
	m.LastEmail = email
	m.LastUsername = username
	m.LastVerificationURL = verificationLink
	return nil
}

var testDB *sql.DB

func TestMain(m *testing.M) {
	testDB = testutils.SetupDB()
	defer testDB.Close()

	testRepo := database.NewRepository(testDB)
	testutils.CleanDB(testRepo)

	os.Exit(m.Run())
}

func TestRegisterUser_Success(t *testing.T) {
	testRepo := database.NewRepository(testDB)
	testutils.CleanDB(testRepo)

	// Mock mailer so no real email is sent
	mockMailer := &MockMailer{}

	// Initialize JWT Token
	jwtManager := auth.NewJWTManager("test-secret", time.Hour)

	// Initialize logger and services
	logger.Init()
	userService := services.NewUserService(testRepo)
	userHandler := handlers.NewUserHandler(userService, mockMailer, jwtManager)

	// Setup test router
	router := gin.Default()
	router.POST("/register", userHandler.RegisterUser)

	// Prepare request
	body := `{"username":"testuser","email":"testuser@example.com","password":"Password123!"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions on response
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, strings.ToLower(w.Body.String()), "check your email")

	// Assertions on pending_users DB entry
	var email string
	err := testDB.QueryRow("SELECT email FROM pending_users WHERE email = $1", "testuser@example.com").Scan(&email)
	assert.NoError(t, err)
	assert.Equal(t, "testuser@example.com", email)
}

func TestRegisterUser_IPRateLimit(t *testing.T) {
	ip := "127.0.0.1"

	// Simulate max attempts
	for i := 0; i < 10; i++ {
		exceeded := handlers.CheckIPRate(ip)
		assert.False(t, exceeded, "Rate limit should not be exceeded yet")
	}

	// 11th attempt should trigger the limit
	exceeded := handlers.CheckIPRate(ip)
	assert.True(t, exceeded, "Rate limit should be exceeded on 11th attempt")
}

func TestVerifyEmail_Success(t *testing.T) {
	testRepo := database.NewRepository(testDB)
	testutils.CleanDB(testRepo)

	logger.Init()
	userService := services.NewUserService(testRepo)

	// Register a pending user directly via service
	pendingUser, err := userService.RegisterPendingUser("testuser", "noreply.insightforge@gmail.com", "Password123!")
	if err != nil {
		t.Fatal(err)
	}

	// Now verify the user by token
	err = userService.ActivateUser(pendingUser.Token)
	if err != nil {
		t.Fatal(err)
	}

	// Assert user was moved to `users` table
	var activeUser database.User
	err = testDB.QueryRow("SELECT id, username, email FROM users WHERE email = $1", "noreply.insightforge@gmail.com").Scan(&activeUser.ID, &activeUser.Username, &activeUser.Email)
	if err != nil {
		t.Fatal("Failed to find active user:", err)
	}

	assert.Equal(t, "noreply.insightforge@gmail.com", activeUser.Email)
	assert.Equal(t, "testuser", activeUser.Username)

	// Ensure user was removed from pending_users
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM pending_users WHERE email = $1", "noreply.insightforge@gmail.com").Scan(&count)
	if err != nil {
		t.Fatal("Failed to count pending users:", err)
	}
	assert.Equal(t, 0, count)
}
