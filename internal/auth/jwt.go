package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTManager struct {
	SecretKey     string
	TokenDuration time.Duration
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

var ErrUnauthorized = errors.New("unauthorized")

const ContextUserIDKey = "userID"

func NewJWTManager(secretKey string, duration time.Duration) *JWTManager {
	return &JWTManager{
		SecretKey:     secretKey,
		TokenDuration: duration,
	}
}

func GenerateSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func (j *JWTManager) Generate(userID uuid.UUID, email string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.TokenDuration)),
			Issuer:    "InsightForge",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

func (j *JWTManager) Verify(accessToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.SecretKey), nil
	})
	if err != nil {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}

func AuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err != nil {
			log.Println("No token cookie found:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication token missing"})
			return
		}

		// Verify the token string
		claims, err := jwtManager.Verify(token)
		if err != nil {
			log.Println("Invalid token:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// Store claims info in context for handlers to use
		c.Set("user_id", claims.UserID.String())
		c.Set("userEmail", claims.Email)

		c.Next()
	}
}

func (m *JWTManager) Parse(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.SecretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return claims, nil
}
