package services

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/Bgoodwin24/insightforge/internal/auth"
	"github.com/Bgoodwin24/insightforge/internal/database"
	"github.com/Bgoodwin24/insightforge/logger"
	"github.com/google/uuid"
)

type UserService struct {
	Repo *database.Repository
}

func NewUserService(repo *database.Repository) *UserService {
	return &UserService{Repo: repo}
}

func (s *UserService) CreateUser(username, email, password string) (*database.User, error) {
	// Hash password
	hashed, err := auth.HashPassword(password)
	if err != nil {
		logger.Logger.Println("ERROR: Failed to hash password:", err)
		return nil, err
	}

	// Use repo to create user
	createUserParams := database.CreateUserParams{
		ID:           uuid.New(),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Email:        email,
		Username:     username,
		PasswordHash: hashed,
	}

	// Save to database using repository
	user, err := s.Repo.Queries.CreateUser(context.Background(), createUserParams)
	if err != nil {
		logger.Logger.Println("ERROR: Error saving user to database:", err)
		return nil, fmt.Errorf("error saving user to database: %v", err)
	}

	// Return user data or error
	return &user, nil
}

func (s *UserService) LoginUser(email, password string) (*database.User, string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	const maxAttempts = 5
	const lockDuration = 15 * time.Minute

	// Get user by email (make query first)
	user, err := s.Repo.Queries.GetUserByEmail(context.Background(), email)
	if err != nil {
		logger.Logger.Println("ERROR: Invalid email or password:", err)
		return nil, "", fmt.Errorf("invalid email or password")
	}

	// Check if account is locked
	if user.LockedUntil.Valid && time.Now().Before(user.LockedUntil.Time) {
		lockedFor := time.Until(user.LockedUntil.Time).Round(time.Minute)
		logger.Logger.Println("ERROR: Account is locked, try again in", lockedFor)
		return nil, "", fmt.Errorf("account is locked, try again in %v", lockedFor)
	}

	// Verify password
	if !auth.VerifyPassword(password, user.PasswordHash) {
		attempts := user.FailedLoginAttempts.Int32 + 1

		// Create parameters for update
		user.LastFailedAttempt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}

		user.LockedUntil = sql.NullTime{Valid: false}

		// If exceeded max attempts
		if attempts >= int32(maxAttempts) {
			user.LockedUntil.Time = time.Now().Add(lockDuration)
			user.LockedUntil.Valid = true
		}

		// Update the failed login attempts in database
		err = s.Repo.Queries.UpdateLoginAttempts(context.Background(), database.UpdateLoginAttemptsParams{
			ID:                  user.ID,
			FailedLoginAttempts: sql.NullInt32{Int32: attempts, Valid: true},
			LastFailedAttempt:   user.LastFailedAttempt,
			LockedUntil:         user.LockedUntil,
		})

		if err != nil {
			logger.Logger.Println("ERROR: Error updating login attempts:", err)
			return nil, "", fmt.Errorf("error updating login attempts: %w", err)
		}

		logger.Logger.Println("ERROR: Invalid email or password")
		return nil, "", fmt.Errorf("invalid email or password")
	}

	// Password is correct - reset login attempts
	err = s.Repo.Queries.ResetLoginAttempts(context.Background(), user.ID)
	if err != nil {
		logger.Logger.Println("ERROR: Error resetting login attempts:", err)
		return nil, "", fmt.Errorf("error resetting login attempts: %w", err)
	}

	jwtManager := auth.NewJWTManager(jwtSecret, time.Hour*24)
	token, err := jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		logger.Logger.Println("ERROR: Failed to generate JWT:", err)
		return nil, "", fmt.Errorf("failed to generate JWT")
	}

	// Return authenticated user
	return &user, token, nil
}

func (s *UserService) ActivateUser(token string) error {
	pending, err := s.Repo.Queries.GetPendingUserByToken(context.Background(), token)
	if err != nil {
		logger.Logger.Println("ERROR: Invalid or expired token")
		return fmt.Errorf("invalid or expired token")
	}

	if time.Now().After(pending.ExpiresAt) {
		logger.Logger.Println("ERROR: Token has expired")
		return fmt.Errorf("token has expired")
	}

	// Check if user already exists before activation
	existingUser, err := s.Repo.Queries.GetUserByEmail(context.Background(), pending.Email)
	if err == nil && existingUser.ID != uuid.Nil {
		logger.Logger.Println("ERROR: User already exists")
		return fmt.Errorf("user already exists")
	} else if err != nil && err != sql.ErrNoRows {
		logger.Logger.Println("ERROR: Unexpected error checking for existing user:", err)
		return fmt.Errorf("error checking for existing user: %v", err)
	}

	_, err = s.Repo.Queries.CreateUser(context.Background(), database.CreateUserParams{
		ID:           pending.ID,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Email:        pending.Email,
		Username:     pending.Username,
		PasswordHash: pending.PasswordHash,
	})
	if err != nil {
		logger.Logger.Println("ERROR: Error creating user:", err)
		return fmt.Errorf("error creating user: %v", err)
	}

	err = s.Repo.Queries.DeletePendingUserByID(context.Background(), pending.ID)
	if err != nil {
		logger.Logger.Println("ERROR: Error deleting pending user:", err)
		return fmt.Errorf("error deleting pending user: %v", err)
	}

	return nil
}

func (s *UserService) RegisterPendingUser(username, email, password string) (*database.PendingUser, error) {
	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	token := uuid.New().String()

	expires_at := time.Now().UTC().Add(24 * time.Hour)

	// Create the Pending user params for SQLC
	params := database.CreatePendingUserParams{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		Token:        token,
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    expires_at,
	}

	pendingUser, err := s.Repo.Queries.CreatePendingUser(context.Background(), params)
	if err != nil {
		return nil, err
	}

	return &pendingUser, nil
}
