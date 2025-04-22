package auth_test

import (
	"os"
	"testing"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/auth"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	secretKey := os.Getenv("JWT_SECRET")
	duration := time.Hour * 24
	jwtManager := auth.NewJWTManager(secretKey, duration)
	userID := uuid.New()
	email := "example@example.com"

	token, err := jwtManager.Generate(userID, email)

	assert.NoError(t, err, "should not return an error")
	assert.NotEmpty(t, token, "should return a non-empty token")

	parsedClaims, err := jwtManager.Verify(token)
	assert.NoError(t, err, "should not return an error verifying the token")
	assert.Equal(t, userID, parsedClaims.UserID, "userID should match")
	assert.Equal(t, email, parsedClaims.Email, "email should match")

	assert.WithinDuration(t, time.Now().Add(duration), parsedClaims.ExpiresAt.Time, time.Second, "token expiration time should be correct")
}

func TestVerify(t *testing.T) {
	secretKey := os.Getenv("JWT_SECRET")
	duration := time.Hour * 24
	jwtManager := auth.NewJWTManager(secretKey, duration)

	userID := uuid.New()
	email := "example@example.com"
	token, _ := jwtManager.Generate(userID, email)

	t.Run("Valid Token", func(t *testing.T) {
		claims, err := jwtManager.Verify(token)

		assert.NoError(t, err, "should not return an error verifying the token")
		assert.Equal(t, userID, claims.UserID, "userID should match")
		assert.Equal(t, email, claims.Email, "email should match")
	})

	t.Run("Invalid Token", func(t *testing.T) {
		// Provide an invalid token
		invalidToken := "invalid.token.here"
		claims, err := jwtManager.Verify(invalidToken)

		assert.Error(t, err, "should return an error for an invalid token")
		assert.Nil(t, claims, "claims should be nil for an invalid token")
	})

	t.Run("Expired Token", func(t *testing.T) {
		// Generate a token with a past expiration
		expiredJWTManager := auth.NewJWTManager(secretKey, -time.Hour)
		expiredToken, err := expiredJWTManager.Generate(userID, email)

		assert.NoError(t, err, "should generate the expired token without errors")

		// Verify the expired token
		claims, err := jwtManager.Verify(expiredToken)

		assert.Error(t, err, "should return an error for an expired token")
		assert.Nil(t, claims, "claims should be nil for an expired token")
	})
}
