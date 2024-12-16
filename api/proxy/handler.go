package proxy

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shipeng101/singdns/pkg/types"
)

type Handler struct {
	proxyService types.ProxyService
}

func NewHandler(proxyService types.ProxyService) *Handler {
	return &Handler{
		proxyService: proxyService,
	}
}

// RegisterHandlers 注册代理相关路由
func RegisterHandlers(r *gin.RouterGroup, proxyService types.ProxyService) {
	h := NewHandler(proxyService)

	r.POST("/restart", h.restart)
}

// restart 重启代理服务
func (h *Handler) restart(c *gin.Context) {
	if err := h.proxyService.Restart(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "代理服务已重启"})
}
