package auth

import (
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
	"github.com/shipeng101/singdns/pkg/types"
)

type service struct {
	jwtSecret string
	users     map[string]string // username -> password map for simplicity
}

// NewService 创建认证服务
func NewService() types.AuthService {
	return &service{
		jwtSecret: "your-secret-key", // 在生产环境中应该从配置文件读取
		users:     make(map[string]string),
	}
}

// Login 用户登录
func (s *service) Login(username, password string) (bool, error) {
	if storedPassword, exists := s.users[username]; exists {
		if storedPassword == password { // 在生产环境中应该使用密码哈希
			return true, nil
		}
	}
	return false, errors.New("用户名或密码错误")
}

// Register 用户注册
func (s *service) Register(username, password string) error {
	if _, exists := s.users[username]; exists {
		return errors.New("用户名已存在")
	}
	s.users[username] = password // 在生产环境中应该使用密码哈希
	return nil
}

// Validate 验证令牌
func (s *service) Validate(token string) bool {
	claims := &jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	return err == nil
}
