package dns

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shipeng101/singdns/pkg/types"
	dns_service "github.com/shipeng101/singdns/service/dns"
)

type Handler struct {
	dnsService dns_service.Service
}

func NewHandler(dnsService dns_service.Service) *Handler {
	return &Handler{
		dnsService: dnsService,
	}
}

// RegisterProtectedRoutes 注册需要认证的DNS路由
func (h *Handler) RegisterProtectedRoutes(router *gin.RouterGroup) {
	dns := router.Group("/dns")
	{
		dns.GET("/status", h.getStatus)
		dns.GET("/rules", h.getRules)
		dns.POST("/rules", h.updateRules)
		dns.POST("/config", h.updateConfig)
	}
}

// getStatus 获取DNS服务状态
func (h *Handler) getStatus(c *gin.Context) {
	status := h.dnsService.GetStatus()
	c.JSON(http.StatusOK, status)
}

// updateConfig 更新DNS配置
func (h *Handler) updateConfig(c *gin.Context) {
	var config types.DNSConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置格式"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则格式"})
		return
	}

	if err := h.dnsService.UpdateRules(rules); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "规则已更新"})
}

// RegisterHandlers 注册DNS相关路由
func RegisterHandlers(r *gin.RouterGroup, dnsService dns_service.Service) {
	h := NewHandler(dnsService)
	h.RegisterProtectedRoutes(r)
}
