package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// User 用户模型
type User struct {
	ID        string    `json:"id"`         // 用户ID
	Username  string    `json:"username"`   // 用户名
	Password  string    `json:"-"`          // 密码
	Role      string    `json:"role"`       // 角色
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// UserRole 用户角色
const (
	RoleAdmin = "admin" // 管理员
	RoleUser  = "user"  // 普通用户
	RoleGuest = "guest" // 访客
)

// Claims JWT令牌的声明
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

// Service 认证服务接口
type Service interface {
	GenerateToken(user *User) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	Login(username, password string) (*User, string, error)
	Register(username, password string) error
}

// Config 认证服务配置
type Config struct {
	JWTSecret        string `json:"jwt_secret"`         // JWT密钥
	TokenExpireHours int    `json:"token_expire_hours"` // Token过期时间(小时)
	AllowRegister    bool   `json:"allow_register"`     // 是否允许注册
	DefaultRole      string `json:"default_role"`       // 默认角色
}
