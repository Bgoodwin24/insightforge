package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/services"
	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/gin-gonic/gin"
)

var (
	// Simple in-memory store for rate limiting
	ipAttemptMap = make(map[string]*ipAttempt)
	ipMutex      = &sync.Mutex{}
)

type ipAttempt struct {
	count    int
	lastSeen time.Time
}

type UserHandler struct {
	userService *services.UserService
	emailConfig struct {
		smtpServer   string
		smtpPort     string
		smtpUsername string
		smtpPassword string
		fromEmail    string
	}
}

type RegisterRequest struct {
	Username string
	Email    string
	Password string
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	SMTP_SERVER := os.Getenv("SMTP_SERVER")
	SMTP_PORT := os.Getenv("SMTP_PORT")
	SMTP_USERNAME := os.Getenv("SMTP_USERNAME")
	SMTP_PASSWORD := os.Getenv("SMTP_PASSWORD")
	FROM_EMAIL := os.Getenv("FROM_EMAIL")

	return &UserHandler{
		userService: userService,
		emailConfig: struct {
			smtpServer   string
			smtpPort     string
			smtpUsername string
			smtpPassword string
			fromEmail    string
		}{
			smtpServer:   SMTP_SERVER,
			smtpPort:     SMTP_PORT,
			smtpUsername: SMTP_USERNAME,
			smtpPassword: SMTP_PASSWORD,
			fromEmail:    FROM_EMAIL,
		},
	}
}

func (h *UserHandler) RegisterUser(c *gin.Context) {
	SMTP_SERVER := os.Getenv("SMTP_SERVER")

	// Get IP address
	ip := c.ClientIP()

	// Check rate limiting
	if checkIPRate(ip) {
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

	// Password strength
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

	// Create a pending user
	pendingUser, err := h.userService.RegisterPendingUser(request.Username, request.Email, request.Password)
	if err != nil {
		// Log security event
		logger.Logger.Printf("SECURITY: Failed registration attempt from IP: %s, Email: %s, Error: %s",
			ip, request.Email, err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Send verification email
	verificationLink := fmt.Sprintf("https://%s/verify?token=%s", SMTP_SERVER, pendingUser.Token)
	err = h.sendVerificationEmail(pendingUser.Email, pendingUser.Username, verificationLink)
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

func checkIPRate(ip string) bool {
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
		ipAttemptMap[ip] = &ipAttempt{count: 1, lastSeen: now}
		return false
	}

	// Reset if outside window
	if now.Sub(attempt.lastSeen) > windowDuration {
		ipAttemptMap[ip] = &ipAttempt{count: 1, lastSeen: now}
		return false
	}

	// Increment attempt count
	attempt.count++
	attempt.lastSeen = now

	// Check if over limit
	return attempt.count > maxAttempts
}

// Send an email with the verification link
func (h *UserHandler) sendVerificationEmail(email, username, verificationLink string) error {
	// Email message
	subject := "Verify your account"
	body := fmt.Sprintf(`
<html>
	<body>
		<p>Hello %s,</p>
		<p>Please verify your account by clicking the link below:</p>
		<p><a href="%s">%s</a></p>
		<p>The link will expire in 24 hours.</p>
		<p>If you didn't register, please ignore this email.</p>
		<br>
		<p>— The InsightForge Team</p>
	</body>
</html>
`, username, verificationLink, verificationLink)
	message := []byte("To: " + email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\";\r\n" +
		"\r\n" +
		body)

	// Connect to SMTP server
	auth := smtp.PlainAuth("", h.emailConfig.smtpUsername, h.emailConfig.smtpPassword, h.emailConfig.smtpServer)
	addr := h.emailConfig.smtpServer + ":" + h.emailConfig.smtpPort

	// Send email
	err := smtp.SendMail(addr, auth, h.emailConfig.fromEmail, []string{email}, message)
	if err != nil {
		return err
	}

	return nil
}

// Add a new handler for verification
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(400, gin.H{"error": "Invalid verification link"})
		return
	}

	// Check if token exists and is valid
	user, err := h.userService.Repo.Queries.GetPendingUserByToken(context.Background(), token)
	if err != nil {
		c.JSON(400, gin.H{"error": "Verification link is invalid or has expired"})
		return
	}

	// Activate the user
	err = h.userService.ActivateUser(token)
	if err != nil {
		logger.Logger.Printf("ERROR: Failed to activate user %s: %s", user.ID, err.Error())
		c.JSON(500, gin.H{"error": "Failed to verify account"})
		return
	}

	// Success - redirect to login page or return success message
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "text/html") {
		c.Redirect(http.StatusFound, "/login")
	} else {
		c.JSON(200, gin.H{"message": "Email verified successfully. You can now log in."})
	}
}
