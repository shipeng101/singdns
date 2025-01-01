package config

import (
	"net/http"
	"singdns/api/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
func (h *ConfigHandler) RegisterRoutes(router gin.IRouter) {
	// 规则集管理
	router.GET("/config/rulesets", h.GetRuleSetsInfo)
	router.POST("/config/rulesets", h.CreateRuleSet)
	router.PUT("/config/rulesets/:id", h.UpdateRuleSet)
	router.DELETE("/config/rulesets/:id", h.DeleteRuleSet)
	router.POST("/config/rulesets/update", h.UpdateRuleSets)

	// SingBox 配置
	router.GET("/config/singbox", h.GetSingBoxConfig)
	router.PUT("/config/singbox", h.UpdateSingBoxConfig)
	router.POST("/config/singbox/outbounds", h.AddSingBoxOutbound)
	router.POST("/config/singbox/inbounds", h.AddSingBoxInbound)
	router.PUT("/config/singbox/inbounds/:tag", h.UpdateSingBoxInbound)
	router.DELETE("/config/singbox/inbounds/:tag", h.DeleteSingBoxInbound)

	// DNS 设置
	router.PUT("/config/dns", h.UpdateDNSConfig)

	// 配置生成
	router.POST("/config/generate", h.GenerateConfigs)
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

// GetRuleSetsInfo 获取规则集信息
func (h *ConfigHandler) GetRuleSetsInfo(c *gin.Context) {
	info, err := h.service.GetRuleSetsInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

// UpdateRuleSets 更新规则集
func (h *ConfigHandler) UpdateRuleSets(c *gin.Context) {
	if err := h.service.UpdateRuleSets(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "规则集更新成功"})
}

// CreateRuleSet 创建规则集
func (h *ConfigHandler) CreateRuleSet(c *gin.Context) {
	var ruleSet models.RuleSet
	if err := c.ShouldBindJSON(&ruleSet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成唯一ID
	ruleSet.ID = uuid.New().String()
	ruleSet.UpdatedAt = time.Now()

	// 保存规则集
	if err := h.service.SaveRuleSet(&ruleSet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ruleSet)
}

// UpdateRuleSet 更新规则集
func (h *ConfigHandler) UpdateRuleSet(c *gin.Context) {
	id := c.Param("id")
	var ruleSet models.RuleSet
	if err := c.ShouldBindJSON(&ruleSet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ruleSet.ID = id
	ruleSet.UpdatedAt = time.Now()

	if err := h.service.SaveRuleSet(&ruleSet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ruleSet)
}

// DeleteRuleSet 删除规则集
func (h *ConfigHandler) DeleteRuleSet(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteRuleSet(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
