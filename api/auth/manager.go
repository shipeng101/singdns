package auth

import (
	"errors"
	"fmt"
	"time"

	"singdns/api/storage"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Manager handles authentication and authorization
type Manager struct {
	jwtSecret []byte
	storage   storage.Storage
}

// NewManager creates a new auth manager
func NewManager(jwtSecret []byte, storage storage.Storage) *Manager {
	return &Manager{
		jwtSecret: jwtSecret,
		storage:   storage,
	}
}

// GenerateToken generates a JWT token for the given username
func (m *Manager) GenerateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(m.jwtSecret)
}

// ValidateToken validates a JWT token and returns the username
func (m *Manager) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if username, ok := claims["username"].(string); ok {
			return username, nil
		}
	}

	return "", fmt.Errorf("invalid token claims")
}

// Login validates user credentials and returns a JWT token
func (m *Manager) Login(username, password string) (string, error) {
	user, err := m.storage.GetUser(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("用户不存在")
		}
		return "", fmt.Errorf("获取用户信息失败: %v", err)
	}

	if !m.verifyPassword(password, user.Password) {
		return "", errors.New("密码错误")
	}

	return m.GenerateToken(username)
}

// UpdatePassword updates the user's password
func (m *Manager) UpdatePassword(username, oldPassword, newPassword string) error {
	user, err := m.storage.GetUser(username)
	if err != nil {
		return err
	}

	if !m.verifyPassword(oldPassword, user.Password) {
		return errors.New("invalid password")
	}

	hashedPassword, err := m.hashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	return m.storage.UpdateUser(user)
}

// VerifyToken verifies a JWT token
func (m *Manager) VerifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
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

// ValidatePassword validates the user's password
func (m *Manager) ValidatePassword(username, password string) error {
	user, err := m.storage.GetUser(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("用户不存在")
		}
		return fmt.Errorf("获取用户信息失败: %v", err)
	}

	if !m.verifyPassword(password, user.Password) {
		return errors.New("密码错误")
	}

	return nil
}

// ChangePassword changes the user's password
func (m *Manager) ChangePassword(username, currentPassword, newPassword string) error {
	// First validate current password
	if err := m.ValidatePassword(username, currentPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := m.hashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update user's password
	user, err := m.storage.GetUser(username)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	return m.storage.UpdateUser(user)
}

// hashPassword hashes a password using bcrypt
func (m *Manager) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword verifies a password against a hash
func (m *Manager) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
