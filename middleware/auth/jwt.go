package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pkgAuth "github.com/shipeng101/singdns/pkg/auth"
)

// JWTMiddleware 创建JWT认证中间件
func JWTMiddleware(authService pkgAuth.Service) gin.HandlerFunc {
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
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}
