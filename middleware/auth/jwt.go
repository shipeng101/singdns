package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shipeng101/singdns/pkg/types"
)

// JWTMiddleware 创建JWT认证中间件
func JWTMiddleware(authService types.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过认证的路径
		if c.Request.URL.Path == "/api/auth/login" ||
			c.Request.URL.Path == "/api/auth/register" {
			c.Next()
			return
		}

		// 获取token
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证token"})
			c.Abort()
			return
		}

		// 验证token
		if !authService.Validate(token) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			c.Abort()
			return
		}

		c.Next()
	}
}
