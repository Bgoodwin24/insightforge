package auth_test

import (
	"testing"

	"github.com/Bgoodwin24/insightforge/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	password := "StrongPassword123!"

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		t.Logf("Error hashing password: %v", err)
	}

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.GreaterOrEqual(t, len(hashedPassword), 60)
}

func TestVerifyPassword(t *testing.T) {
	password := "StrongPassword123!"

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		t.Logf("Error hashing password: %v", err)
	}

	assert.NoError(t, err)

	t.Run("Correct Password", func(t *testing.T) {
		result := auth.VerifyPassword(password, hashedPassword)
		assert.True(t, result, "Password verification should succeed with the correct password")
	})

	t.Run("Incorrect Password", func(t *testing.T) {
		incorrectPassword := "IncorrectPassword123!"
		result := auth.VerifyPassword(incorrectPassword, hashedPassword)
		assert.False(t, result, "Password verification should fail with an incorrect password")
	})
}
