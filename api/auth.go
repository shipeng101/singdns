package api

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Claims represents JWT claims
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthManager manages user authentication
type AuthManager struct {
	users  map[string]*User
	mu     sync.RWMutex
	jwtKey []byte
}

// NewAuthManager creates a new auth manager
func NewAuthManager(jwtKey []byte) *AuthManager {
	return &AuthManager{
		users:  make(map[string]*User),
		jwtKey: jwtKey,
	}
}

// Register registers a new user
func (m *AuthManager) Register(username, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[username]; exists {
		return fmt.Errorf("user already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	m.users[username] = &User{
		Username:  username,
		Password:  string(hash),
		Role:      "user",
		CreatedAt: time.Now(),
	}

	return nil
}

// Login authenticates a user and returns a JWT token
func (m *AuthManager) Login(username, password string) (string, error) {
	m.mu.RLock()
	user, exists := m.users[username]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid password")
	}

	claims := &Claims{
		Username: username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.jwtKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (m *AuthManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return claims, nil
}

// GetUser returns a user by ID
func (m *AuthManager) GetUser(id string) (*User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

// UpdateUser updates a user's password
func (m *AuthManager) UpdateUser(id string, password string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find user
	var user *User
	for _, u := range m.users {
		if u.ID == id {
			user = u
			break
		}
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	user.Password = string(hash)
	return nil
}

// UpdatePassword updates user's password
func (m *AuthManager) UpdatePassword(username, oldPassword, newPassword string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verify old password
	if _, err := m.Login(username, oldPassword); err != nil {
		return fmt.Errorf("invalid old password")
	}

	// Get user
	user, exists := m.users[username]
	if !exists {
		return fmt.Errorf("user not found")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// Update password
	user.Password = string(hashedPassword)
	m.users[username] = user

	return nil
}
