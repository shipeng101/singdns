package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shipeng101/singdns/pkg/types"
)

type Handler struct {
	configService types.ConfigService
}

func NewHandler(configService types.ConfigService) *Handler {
	return &Handler{
		configService: configService,
	}
}

// RegisterHandlers 注册系统相关路由
func RegisterHandlers(r *gin.RouterGroup, configService types.ConfigService) {
	h := NewHandler(configService)

	r.GET("/config", h.getConfig)
	r.POST("/config", h.updateConfig)
}

// getConfig 获取系统配置
func (h *Handler) getConfig(c *gin.Context) {
	config := h.configService.Get()
	c.JSON(http.StatusOK, config)
}

// updateConfig 更新系统配置
func (h *Handler) updateConfig(c *gin.Context) {
	var config types.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置格式"})
		return
	}

	if err := h.configService.Update(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置已更新"})
}
