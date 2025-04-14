package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword creates a bcrypt hash from a plaintext password with the default cost factor
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword compares a plaintext password against a hashed password
// Returns true if the match, false otherwise
// Could add cost parameter and adjust if implementation requires a different cost
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
