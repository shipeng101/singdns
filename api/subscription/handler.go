package subscription

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shipeng101/singdns/pkg/types"
)

type Handler struct {
	subscriptionService types.SubscriptionService
}

func NewHandler(subscriptionService types.SubscriptionService) *Handler {
	return &Handler{
		subscriptionService: subscriptionService,
	}
}

// RegisterHandlers 注册订阅相关路由
func RegisterHandlers(r *gin.RouterGroup, subscriptionService types.SubscriptionService) {
	h := NewHandler(subscriptionService)

	r.GET("/nodes", h.getNodes)
	r.POST("/update", h.updateSubscription)
}

// getNodes 获取节点列表
func (h *Handler) getNodes(c *gin.Context) {
	nodes, err := h.subscriptionService.GetNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

// updateSubscription 更新订阅
func (h *Handler) updateSubscription(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	if err := h.subscriptionService.UpdateSubscription(req.URL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "订阅已更新"})
}
