package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/shipeng101/singdns/pkg/auth"
)

type service struct {
	config auth.Config
	users  map[string]*auth.User
}

// NewService 创建认证服务
func NewService() auth.Service {
	return &service{
		config: auth.Config{
			JWTSecret:        "your-secret-key",
			TokenExpireHours: 24,
			AllowRegister:    true,
			DefaultRole:      auth.RoleUser,
		},
		users: make(map[string]*auth.User),
	}
}

// GenerateToken 生成JWT令牌
func (s *service) GenerateToken(user *auth.User) (string, error) {
	claims := &auth.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(s.config.TokenExpireHours) * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "singdns",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// ValidateToken 验证JWT令牌
func (s *service) ValidateToken(tokenString string) (*auth.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*auth.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// Login 用户登录
func (s *service) Login(username, password string) (*auth.User, string, error) {
	user, ok := s.users[username]
	if !ok || user.Password != password { // 实际应该使用密码哈希
		return nil, "", errors.New("invalid username or password")
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Register 用户注册
func (s *service) Register(username, password string) error {
	if !s.config.AllowRegister {
		return errors.New("registration is disabled")
	}

	if _, exists := s.users[username]; exists {
		return errors.New("username already exists")
	}

	s.users[username] = &auth.User{
		ID:        generateID(), // 使用一个函数生成唯一ID
		Username:  username,
		Password:  password, // 实际应该使用密码哈希
		Role:      s.config.DefaultRole,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return nil
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
