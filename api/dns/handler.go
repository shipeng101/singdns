package dns

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shipeng101/singdns/pkg/types"
)

type Handler struct {
	dnsService types.DNSService
}

// RegisterHandlers 注册DNS相关路由
func RegisterHandlers(r *gin.RouterGroup, dnsService types.DNSService) {
	h := &Handler{dnsService: dnsService}

	r.GET("/status", h.getStatus)
	r.GET("/config", h.getConfig)
	r.PUT("/config", h.updateConfig)
	r.GET("/rules", h.getRules)
	r.PUT("/rules", h.updateRules)
}

// getStatus 获取DNS服务状态
func (h *Handler) getStatus(c *gin.Context) {
	status := h.dnsService.GetStatus()
	c.JSON(http.StatusOK, status)
}

// getConfig 获取DNS配置
func (h *Handler) getConfig(c *gin.Context) {
	config := h.dnsService.GetConfig()
	c.JSON(http.StatusOK, config)
}

// updateConfig 更新DNS配置
func (h *Handler) updateConfig(c *gin.Context) {
	var config types.DNSConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置参数"})
		return
	}

	if err := h.dnsService.UpdateConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置已更新"})
}

// getRules 获取DNS规则
func (h *Handler) getRules(c *gin.Context) {
	rules := h.dnsService.GetRules()
	c.JSON(http.StatusOK, rules)
}

// updateRules 更新DNS规则
func (h *Handler) updateRules(c *gin.Context) {
	var rules []types.Rule
	if err := c.ShouldBindJSON(&rules); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则参数"})
		return
	}

	if err := h.dnsService.UpdateRules(rules); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "规则已更新"})
}
