package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shipeng101/singdns/pkg/types"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
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
	authService types.AuthService
}

// RegisterHandlers 注册认证相关路由
func RegisterHandlers(r *gin.RouterGroup, authService types.AuthService) {
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

	success, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Success: success,
		Message: "登录成功",
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
		c.JSON(http.StatusBadRequest, RegisterResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{Message: "注册成功"})
}
