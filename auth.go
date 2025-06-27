package main

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Config holds security parameters
type Config struct {
	MinPasswordLength int
	MaxPasswordLength int
	BcryptCost        int
}

// DefaultConfig provides secure default values
var DefaultConfig = Config{
	MinPasswordLength: 6,
	MaxPasswordLength: 50, // bcrypt max length
	BcryptCost:        bcrypt.DefaultCost,
}

// Validation errors
var (
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrPasswordTooShort  = errors.New("password too short")
	ErrPasswordTooLong   = errors.New("password too long")
	ErrPasswordWeak      = errors.New("password too weak")
	ErrEmailExists       = errors.New("email already exists")
	ErrInvalidCharacters = errors.New("password contains invalid characters")
	ErrDatabaseError     = errors.New("database error")
)

// CreateUser validates and creates a new user account
func (s *PostgresStore) CreateUser(email, password string) (string, error) {
	// Sanitize inputs
	config := DefaultConfig
	email = strings.TrimSpace(strings.ToLower(email))
	password = strings.TrimSpace(password)

	// Validate email format
	if !isValidEmail(email) {
		return "", ErrInvalidEmail
	}

	// Check if email already exists
	if err := checkEmailExists(s.db, email); err != nil {
		return "", err
	}

	// Validate password
	if err := validatePassword(password, config); err != nil {
		return "", err
	}

	// Hash password
	hashedPassword, err := hashPassword(password, config.BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return hashedPassword, nil
}

// isValidEmail checks email format using regex
func isValidEmail(email string) bool {
	// Basic email regex pattern
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email) && len(email) <= 254
}

// checkEmailExists verifies if email is already registered
func checkEmailExists(db *sql.DB, email string) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM worker WHERE email = $1", email).Scan(&count)
	if err != nil {
		return ErrDatabaseError
	}
	if count > 0 {
		return ErrEmailExists
	}
	return nil
}

// validatePassword checks password strength
func validatePassword(password string, config Config) error {
	// Check length
	if len(password) < config.MinPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > config.MaxPasswordLength {
		fmt.Println("password len = ", password)
		return ErrPasswordTooLong
	}

	// Check for invalid characters
	if strings.ContainsAny(password, "\x00\t\n\r") {
		return ErrInvalidCharacters
	}

	// Check password strength
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return ErrPasswordWeak
	}

	return nil
}

// hashPassword creates a secure password hash
func hashPassword(password string, cost int) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
