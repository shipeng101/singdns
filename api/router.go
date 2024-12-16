package api

import (
	"github.com/gin-gonic/gin"
	authAPI "github.com/shipeng101/singdns/api/auth"
	"github.com/shipeng101/singdns/api/dns"
	"github.com/shipeng101/singdns/api/proxy"
	"github.com/shipeng101/singdns/api/subscription"
	authMiddleware "github.com/shipeng101/singdns/middleware/auth"
	"github.com/shipeng101/singdns/pkg/types"
)

// RegisterRoutes 注册路由
func RegisterRoutes(r *gin.Engine, services *types.Services) {
	// 认证中间件
	jwtMiddleware := authMiddleware.JWTMiddleware(services.AuthService)

	// API路由组
	api := r.Group("/api")
	{
		// 认证相关路由
		authAPI.RegisterHandlers(api, services.AuthService)

		// 需要认证的路由组
		authorized := api.Group("/", jwtMiddleware)
		{
			// DNS相关路由
			dns.RegisterHandlers(authorized.Group("/dns"), services.DNSService)

			// 订阅相关路由
			subscription.RegisterHandlers(authorized.Group("/subscription"), services.SubscriptionService)

			// 代理相关路由
			proxy.RegisterHandlers(authorized.Group("/proxy"), services.ProxyService)
		}
	}
}
