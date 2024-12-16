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

// RegisterHandlers 注册代理服务相关路由
func RegisterHandlers(r *gin.RouterGroup, proxyService types.ProxyService) {
	h := NewHandler(proxyService)

	r.GET("/status", h.getStatus)
	r.GET("/rules", h.getRules)
	r.POST("/rules", h.updateRules)
	r.GET("/iptables", h.getIPTables)
	r.POST("/iptables", h.updateIPTables)
	r.GET("/nodes", h.listNodes)
	r.POST("/nodes", h.addNode)
	r.PUT("/nodes/:id", h.updateNode)
	r.DELETE("/nodes/:id", h.deleteNode)
	r.POST("/nodes/:id/test", h.testNode)
	r.POST("/nodes/:id/switch", h.switchNode)
	r.POST("/mode", h.setMode)
}

// getStatus 获取代理状态
func (h *Handler) getStatus(c *gin.Context) {
	stats, err := h.proxyService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// getRules 获取规则列表
func (h *Handler) getRules(c *gin.Context) {
	rules, err := h.proxyService.GetRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

// updateRules 更新规则列表
func (h *Handler) updateRules(c *gin.Context) {
	var rules []types.Rule
	if err := c.ShouldBindJSON(&rules); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则格式"})
		return
	}

	if err := h.proxyService.UpdateRules(rules); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "规则已更新"})
}

// getIPTables 获取iptables规则列表
func (h *Handler) getIPTables(c *gin.Context) {
	rules, err := h.proxyService.GetIPTables()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

// updateIPTables 更新iptables规则列表
func (h *Handler) updateIPTables(c *gin.Context) {
	var rules []types.IPTablesRule
	if err := c.ShouldBindJSON(&rules); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的规则格式"})
		return
	}

	if err := h.proxyService.UpdateIPTables(rules); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "规则已更新"})
}

// listNodes 获取节点列表
func (h *Handler) listNodes(c *gin.Context) {
	nodes, err := h.proxyService.ListNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

// addNode 添加节点
func (h *Handler) addNode(c *gin.Context) {
	var node types.ProxyNode
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的节点格式"})
		return
	}

	if err := h.proxyService.AddNode(&node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "节点已添加", "id": node.ID})
}

// updateNode 更新节点
func (h *Handler) updateNode(c *gin.Context) {
	id := c.Param("id")
	var node types.ProxyNode
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的节点格式"})
		return
	}

	node.ID = id
	if err := h.proxyService.UpdateNode(&node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "节点已更新"})
}

// deleteNode 删除节点
func (h *Handler) deleteNode(c *gin.Context) {
	id := c.Param("id")
	if err := h.proxyService.DeleteNode(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "节点已删除"})
}

// testNode 测试节点延迟
func (h *Handler) testNode(c *gin.Context) {
	id := c.Param("id")
	latency, err := h.proxyService.TestNode(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"latency": latency})
}

// switchNode 切换节点
func (h *Handler) switchNode(c *gin.Context) {
	id := c.Param("id")
	if err := h.proxyService.SwitchNode(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "节点已切换"})
}

// setMode 设置代理模式
func (h *Handler) setMode(c *gin.Context) {
	var req struct {
		Mode string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模式格式"})
		return
	}

	if err := h.proxyService.SetMode(req.Mode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "模式已更新"})
}
