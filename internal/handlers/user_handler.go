package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/auth"
	"github.com/Bgoodwin24/insightforge/internal/email"
	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	// Simple in-memory store for rate limiting
	ipAttemptMap = make(map[string]*IPAttempt)
	ipMutex      = &sync.Mutex{}
)

type IPAttempt struct {
	count    int
	lastSeen time.Time
}

type UserHandler struct {
	UserService *services.UserService
	Mailer      email.Mailer
	JWTManager  *auth.JWTManager
}

type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

func NewUserHandler(userService *services.UserService, mailer email.Mailer, jwtManager *auth.JWTManager) *UserHandler {
	return &UserHandler{
		UserService: userService,
		Mailer:      mailer,
		JWTManager:  jwtManager,
	}
}

func (h *UserHandler) RegisterUser(c *gin.Context) {
	API_URL := os.Getenv("API_URL")

	// Get IP address
	ip := c.ClientIP()

	// Check rate limiting
	if CheckIPRate(ip) {
		// Log the attempt
		logger.Logger.Printf("SECURITY: Rate limit exceeded for registration attempts from IP: %s", ip)
		c.JSON(400, gin.H{"error": "Too many registration attempts. Please try again later."})
		return
	}
	request := RegisterRequest{}

	// Parse request
	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Validation
	if request.Username == "" || request.Email == "" || request.Password == "" {
		c.JSON(400, gin.H{"error": "Username, email, and password are required"})
		return
	}

	// Email format validation
	if !strings.Contains(request.Email, "@") || !strings.Contains(request.Email, ".") {
		c.JSON(400, gin.H{"error": "Invalid email format: must contain '@' and '.'"})
		return
	}

	emailParts := strings.Split(request.Email, "@")
	if len(emailParts) != 2 || emailParts[0] == "" || emailParts[1] == "" {
		c.JSON(400, gin.H{"error": "Invalid email format: must have content before and after '@'"})
		return
	}

	domainParts := strings.Split(emailParts[1], ".")
	if len(domainParts) < 2 || domainParts[len(domainParts)-1] == "" {
		c.JSON(400, gin.H{"error": "Invalid email format: domain must have at least one '.' followed by a TLD"})
		return
	}

	// App Password strength check (skip if it's an app password)
	if !IsAppPassword(request.Password) {
		hasUppercase := false
		hasLowercase := false
		hasDigit := false
		hasSpecial := false
		hasLength := len(request.Password) >= 8

		for _, char := range request.Password {
			if char >= 'A' && char <= 'Z' {
				hasUppercase = true
			}

			if char >= 'a' && char <= 'z' {
				hasLowercase = true
			}

			if char >= '0' && char <= '9' {
				hasDigit = true
			}

			if (char >= '!' && char <= '/') || (char >= ':' && char <= '@') ||
				(char >= '[' && char <= '`') || (char >= '{' && char <= '~') {
				hasSpecial = true
			}
		}

		// Check all password requirements
		if !hasUppercase || !hasLowercase || !hasDigit || !hasSpecial || !hasLength {
			c.JSON(400, gin.H{"error": "Password must contain at least 1 uppercase letter, 1 lowercase letter, 1 digit, 1 special character, and be at least 8 characters long"})
			return
		}
	}

	// Create a pending user
	pendingUser, err := h.UserService.RegisterPendingUser(request.Username, request.Email, request.Password)
	if err != nil {
		// Log security event
		logger.Logger.Printf("SECURITY: Failed registration attempt from IP: %s, Error: %s",
			ip, err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Send verification email
	verificationLink := fmt.Sprintf("%s/user/verify?token=%s", API_URL, pendingUser.Token)
	err = h.Mailer.SendVerificationEmail(pendingUser.Email, pendingUser.Username, verificationLink)
	if err != nil {
		logger.Logger.Printf("ERROR: Failed to send verification email: %s", err.Error())
		c.JSON(500, gin.H{"error": "Failed to send verification email, please try again"})
		return
	}

	// Return Response
	c.JSON(201, gin.H{
		"message":    "Registration successful. Please check your email to verify your account.",
		"id":         pendingUser.ID,
		"username":   pendingUser.Username,
		"email":      pendingUser.Email,
		"created_at": pendingUser.CreatedAt,
	})
}

func CheckIPRate(ip string) bool {
	// Note: In a production environment, this would use Redis or a database
	// for distributed tracking and persistence, and would include more
	// sophisticated tracking like browser fingerprinting.
	ipMutex.Lock()
	defer ipMutex.Unlock()

	now := time.Now()
	windowDuration := 15 * time.Minute
	maxAttempts := 10

	// Clean up old entries while here
	for k, v := range ipAttemptMap {
		if now.Sub(v.lastSeen) > windowDuration {
			delete(ipAttemptMap, k)
		}
	}

	attempt, exists := ipAttemptMap[ip]
	if !exists {
		ipAttemptMap[ip] = &IPAttempt{count: 1, lastSeen: now}
		return false
	}

	// Reset if outside window
	if now.Sub(attempt.lastSeen) > windowDuration {
		ipAttemptMap[ip] = &IPAttempt{count: 1, lastSeen: now}
		return false
	}

	// Increment attempt count
	attempt.count++
	attempt.lastSeen = now

	// Check if over limit
	return attempt.count > maxAttempts
}

// Add a new handler for verification
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	FRONTEND_URL := os.Getenv("FRONTEND_URL")
	token := c.Query("token")
	if token == "" {
		c.JSON(400, gin.H{"error": "Invalid verification link"})
		return
	}

	user, err := h.UserService.Repo.Queries.GetPendingUserByToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(400, gin.H{"error": "Verification link is invalid or has expired"})
		return
	}

	// Activate user
	err = h.UserService.ActivateUser(token)
	if err != nil {
		logger.Logger.Printf("ERROR: Failed to activate user %s: %s", user.ID, err.Error())
		c.JSON(500, gin.H{"error": "Failed to verify account"})
		return
	}

	// Redirect or send success
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "text/html") {
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/user/login", FRONTEND_URL))
	} else {
		c.JSON(200, gin.H{"message": "Email verified successfully. You can now log in."})
	}
}

// Helper function to determine if the password is an app password
func IsAppPassword(password string) bool {
	// Remove spaces from the password
	password = strings.ReplaceAll(password, " ", "")

	// Check for the length of the password (Google app passwords are 16 characters long)
	return len(password) == 16
}

func (h *UserHandler) LoginUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
		return
	}

	user, token, err := h.UserService.LoginUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set JWT token as an HTTP-only cookie, expires in 1 hour
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Login Successful",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func (h *UserHandler) LogoutUser(c *gin.Context) {
	// Clear the cookie by setting MaxAge < 0 or expires in past
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *UserHandler) GetMyProfile(c *gin.Context) {
	rawID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	rawIDStr, ok := rawID.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID type"})
		return
	}

	userID, err := uuid.Parse(rawIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	user, err := h.UserService.Repo.Queries.GetUserByID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User no longer exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
