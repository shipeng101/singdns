package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shipeng101/singdns/pkg/auth"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string     `json:"token"`
	ExpiresIn int64      `json:"expires_in"` // 过期时间（秒）
	User      *auth.User `json:"user"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	Message string `json:"message"`
}

type Handler struct {
	authService auth.Service
}

// RegisterHandlers 注册认证相关路由
func RegisterHandlers(r *gin.RouterGroup, authService auth.Service) {
	h := &Handler{authService: authService}

	r.POST("/login", h.login)
	r.POST("/register", h.register)
}

// login 处理登录请求
func (h *Handler) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	user, token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token:     token,
		ExpiresIn: 24 * 60 * 60, // 24小时
		User:      user,
	})
}

// register 处理注册请求
func (h *Handler) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if err := h.authService.Register(req.Username, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{Message: "注册成功"})
}
