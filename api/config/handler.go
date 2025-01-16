package config

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	service *ConfigService
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(service *ConfigService) *ConfigHandler {
	return &ConfigHandler{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *ConfigHandler) RegisterRoutes(router *gin.Engine) {
	config := router.Group("/api/config")
	{
		// SingBox 配置
		config.GET("/singbox", h.GetSingBoxConfig)
		config.PUT("/singbox", h.UpdateSingBoxConfig)
		config.POST("/singbox/outbounds", h.AddSingBoxOutbound)
		config.POST("/singbox/inbounds", h.AddSingBoxInbound)
		config.PUT("/singbox/inbounds/:tag", h.UpdateSingBoxInbound)
		config.DELETE("/singbox/inbounds/:tag", h.DeleteSingBoxInbound)

		// DNS 设置
		config.PUT("/dns", h.UpdateDNSConfig)

		// 配置生成
		config.POST("/generate", h.GenerateConfigs)
	}
}

// GetSingBoxConfig 获取 SingBox 配置
func (h *ConfigHandler) GetSingBoxConfig(c *gin.Context) {
	config := h.service.GetSingBoxConfig()
	if config == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "config not initialized"})
		return
	}
	c.JSON(http.StatusOK, config)
}

// UpdateSingBoxConfig 更新 SingBox 配置
func (h *ConfigHandler) UpdateSingBoxConfig(c *gin.Context) {
	var config SingBoxConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateSingBoxConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SingBox config updated successfully"})
}

// AddSingBoxOutbound 添加 SingBox 出站配置
func (h *ConfigHandler) AddSingBoxOutbound(c *gin.Context) {
	var outbound OutboundConfig
	if err := c.ShouldBindJSON(&outbound); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddOutbound(outbound); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SingBox outbound added successfully"})
}

// UpdateDNSConfig 更新 DNS 配置
func (h *ConfigHandler) UpdateDNSConfig(c *gin.Context) {
	var config struct {
		Servers  []DNSServerConfig `json:"servers"`
		Upstream DNSServerConfig   `json:"upstream"`
	}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateDNSConfig(config.Servers, config.Upstream); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "DNS config updated successfully"})
}

// GenerateConfigs 生成配置
func (h *ConfigHandler) GenerateConfigs(c *gin.Context) {
	if err := h.service.GenerateConfigs(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configs generated successfully"})
}

// AddSingBoxInbound 添加入站配置
func (h *ConfigHandler) AddSingBoxInbound(c *gin.Context) {
	var inbound InboundConfig
	if err := c.ShouldBindJSON(&inbound); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddInbound(inbound); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inbound added successfully"})
}

// UpdateSingBoxInbound 更新入站配置
func (h *ConfigHandler) UpdateSingBoxInbound(c *gin.Context) {
	tag := c.Param("tag")
	var inbound InboundConfig
	if err := c.ShouldBindJSON(&inbound); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateInbound(tag, inbound); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inbound updated successfully"})
}

// DeleteSingBoxInbound 删除入站配置
func (h *ConfigHandler) DeleteSingBoxInbound(c *gin.Context) {
	tag := c.Param("tag")
	if err := h.service.DeleteInbound(tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inbound deleted successfully"})
}
