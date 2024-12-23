package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Manager handles authentication and authorization
type Manager struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Secret   []byte `json:"secret"`
}

// NewManager creates a new auth manager
func NewManager(username, password string, secret []byte) *Manager {
	return &Manager{
		Username: username,
		Password: password,
		Secret:   secret,
	}
}

// Login authenticates a user and returns a JWT token
func (m *Manager) Login(username, password string) (string, error) {
	// Check if username matches
	if username != m.Username {
		return "", fmt.Errorf("invalid username or password")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(m.Password), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid username or password")
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(m.Secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

	return tokenString, nil
}

// UpdatePassword updates the admin password
func (m *Manager) UpdatePassword(oldPassword, newPassword string) error {
	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(m.Password), []byte(oldPassword)); err != nil {
		return fmt.Errorf("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	m.Password = string(hashedPassword)
	return nil
}

// VerifyToken verifies a JWT token
func (m *Manager) VerifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.Secret, nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if username, ok := claims["username"].(string); ok {
			return username, nil
		}
	}

	return "", fmt.Errorf("invalid token claims")
}
