package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"singdns/api/auth"
	"singdns/api/config"
	"singdns/api/middleware"
	"singdns/api/models"
	"singdns/api/protocols"
	"singdns/api/proxy"
	"singdns/api/ruleset"
	"singdns/api/storage"
	"singdns/api/subscription"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/sirupsen/logrus"
)

// Config represents the server configuration
type Config struct {
	UpdateInterval time.Duration `json:"update_interval"`
	Auth           *auth.Manager `json:"auth"`
}

// NetworkStats stores network statistics
type NetworkStats struct {
	RxBytes   uint64
	TxBytes   uint64
	Timestamp time.Time
}

// Server represents the API server
type Server struct {
	router       *gin.Engine
	manager      *proxy.Manager
	storage      storage.Storage
	logger       *logrus.Logger
	config       *Config
	updater      *ruleset.Updater
	networkStats *NetworkStats // Add network stats cache
	proxy        *proxy.Manager
}

// NewServer creates a new API server
func NewServer(storage storage.Storage, manager *proxy.Manager, logger *logrus.Logger, cfg *Config) *Server {
	// 根据环境变量设置Gin模式
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger)) // 添加日志中间件

	// Configure CORS
	router.Use(func() gin.HandlerFunc {
		config := cors.Config{
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-CSRF-Token"},
			ExposeHeaders:    []string{"Content-Length", "Content-Type", "Authorization"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
			AllowOriginFunc: func(origin string) bool {
				return true // 允许所有来源
			},
		}
		return cors.New(config)
	}())

	// Create auth manager
	cfg.Auth = auth.NewManager([]byte("your-jwt-secret"), storage)

	// Create config service and handler
	configService := config.NewConfigService("configs/sing-box", storage)
	configHandler := config.NewConfigHandler(configService)

	server := &Server{
		router:       router,
		manager:      manager,
		storage:      storage,
		logger:       logger,
		config:       cfg,
		networkStats: &NetworkStats{Timestamp: time.Now()},
		proxy:        manager,
	}

	// Add auth middleware
	server.router.Use(server.authMiddleware())

	// Create rule set updater
	server.updater = ruleset.NewUpdater(storage, logger)

	// Register config routes
	configHandler.RegisterRoutes(router)

	// Setup routes
	server.setupRoutes()

	// Start subscription auto-updater
	server.startSubscriptionUpdater()

	return server
}

// Start starts the API server
func (s *Server) Start() error {
	// 下载规则集文件
	if err := config.DownloadRuleSets(); err != nil {
		return fmt.Errorf("failed to download rule sets: %v", err)
	}

	// 初始化规则集到数据库
	if err := config.InitializeRuleSets(s.storage); err != nil {
		return fmt.Errorf("failed to initialize rule sets: %v", err)
	}

	// 启动服务器
	return s.router.Run()
}

// Stop stops the API server
func (s *Server) Stop() {
	// Stop rule set updater
	s.updater.Stop()
}

// setupRoutes sets up the API routes
func (s *Server) setupRoutes() {
	// Auth routes
	s.router.POST("/api/auth/login", s.handleLogin)
	s.router.POST("/api/auth/logout", s.handleLogout)
	s.router.POST("/api/auth/password", s.handleUpdatePassword)

	// System routes
	s.router.GET("/api/system/info", s.handleGetSystemInfo)
	s.router.GET("/api/system/status", s.handleGetSystemStatus)
	s.router.POST("/api/system/services/:name/start", s.handleStartService)
	s.router.POST("/api/system/services/:name/stop", s.handleStopService)
	s.router.POST("/api/system/services/:name/restart", s.handleRestartService)

	// Node routes
	s.router.GET("/api/nodes", s.handleGetNodes)
	s.router.GET("/api/nodes/:id", s.handleGetNode)
	s.router.POST("/api/nodes", s.handleCreateNode)
	s.router.PUT("/api/nodes/:id", s.handleUpdateNode)
	s.router.DELETE("/api/nodes/:id", s.handleDeleteNode)
	s.router.POST("/api/nodes/test", s.handleTestNodes)

	// Node group routes
	s.router.GET("/api/node-groups", s.handleGetNodeGroups)
	s.router.GET("/api/node-groups/:id", s.handleGetNodeGroup)
	s.router.POST("/api/node-groups", s.handleCreateNodeGroup)
	s.router.PUT("/api/node-groups/:id", s.handleUpdateNodeGroup)
	s.router.DELETE("/api/node-groups/:id", s.handleDeleteNodeGroup)

	// Rule routes
	s.router.GET("/api/rules", s.handleGetRules)
	s.router.GET("/api/rules/:id", s.handleGetRule)
	s.router.POST("/api/rules", s.handleCreateRule)
	s.router.PUT("/api/rules/:id", s.handleUpdateRule)
	s.router.DELETE("/api/rules/:id", s.handleDeleteRule)

	// Rule set routes
	s.router.GET("/api/rulesets", s.handleGetRuleSets)
	s.router.GET("/api/rulesets/:id", s.handleGetRuleSet)
	s.router.POST("/api/rulesets", s.handleCreateRuleSet)
	s.router.PUT("/api/rulesets/:id", s.handleUpdateRuleSet)
	s.router.DELETE("/api/rulesets/:id", s.handleDeleteRuleSet)
	s.router.POST("/api/rulesets/:id/update", s.handleUpdateRuleSetRules)

	// Subscription routes
	s.router.GET("/api/subscriptions", s.handleGetSubscriptions)
	s.router.GET("/api/subscriptions/:id", s.handleGetSubscription)
	s.router.POST("/api/subscriptions", s.handleCreateSubscription)
	s.router.PUT("/api/subscriptions/:id", s.handleUpdateSubscription)
	s.router.DELETE("/api/subscriptions/:id", s.handleDeleteSubscription)
	s.router.POST("/api/subscriptions/:id/refresh", s.handleRefreshSubscription)

	// Settings routes
	s.router.GET("/api/settings", s.handleGetSettings)
	s.router.PUT("/api/settings", s.handleUpdateSettings)
	s.router.PUT("/api/settings/password", s.handleUpdatePassword)

	// DNS 规则路由
	s.router.GET("/api/dns/rules", s.handleGetDNSRules)
	s.router.POST("/api/dns/rules", s.handleCreateDNSRule)
	s.router.PUT("/api/dns/rules/:id", s.handleUpdateDNSRule)
	s.router.DELETE("/api/dns/rules/:id", s.handleDeleteDNSRule)
	s.router.PUT("/api/dns/settings", s.handleUpdateDNSSettings)
}

// handleGetSystemInfo handles GET /api/system/info
func (s *Server) handleGetSystemInfo(c *gin.Context) {
	info := gin.H{
		"version": "0.1.0",
		"uptime":  0,
		"system": gin.H{
			"platform": runtime.GOOS,
			"arch":     runtime.GOARCH,
			"cpu": gin.H{
				"cores":    runtime.NumCPU(),
				"usage":    0.0,
				"load_avg": []float64{0, 0, 0},
			},
			"memory": gin.H{
				"total":     uint64(0),
				"used":      uint64(0),
				"available": uint64(0),
			},
			"network": gin.H{
				"rx_bytes":    uint64(0),
				"tx_bytes":    uint64(0),
				"rx_rate":     uint64(0),
				"tx_rate":     uint64(0),
				"connections": 0,
			},
		},
	}

	// Get CPU info
	if cpuPercent, err := cpu.Percent(0, false); err == nil && len(cpuPercent) > 0 {
		info["system"].(gin.H)["cpu"].(gin.H)["usage"] = cpuPercent[0]
	} else {
		s.logger.Warnf("Failed to get CPU usage: %v", err)
	}

	// Get load average
	if loadAvg, err := load.Avg(); err == nil {
		info["system"].(gin.H)["cpu"].(gin.H)["load_avg"] = []float64{loadAvg.Load1, loadAvg.Load5, loadAvg.Load15}
	} else {
		s.logger.Warnf("Failed to get load average: %v", err)
	}

	// Get memory info
	if memInfo, err := mem.VirtualMemory(); err == nil {
		info["system"].(gin.H)["memory"].(gin.H)["total"] = memInfo.Total
		info["system"].(gin.H)["memory"].(gin.H)["used"] = memInfo.Used
		info["system"].(gin.H)["memory"].(gin.H)["available"] = memInfo.Available
	} else {
		s.logger.Warnf("Failed to get memory info: %v", err)
	}

	// Get network info and calculate rates
	if netIOCounters, err := net.IOCounters(false); err == nil && len(netIOCounters) > 0 {
		currentTime := time.Now()
		timeDiff := currentTime.Sub(s.networkStats.Timestamp).Seconds()

		// Calculate network rates
		var rxRate, txRate uint64
		if timeDiff > 0 {
			rxDiff := netIOCounters[0].BytesRecv - s.networkStats.RxBytes
			txDiff := netIOCounters[0].BytesSent - s.networkStats.TxBytes
			rxRate = uint64(float64(rxDiff) / timeDiff)
			txRate = uint64(float64(txDiff) / timeDiff)
		}

		// Update network stats cache
		s.networkStats.RxBytes = netIOCounters[0].BytesRecv
		s.networkStats.TxBytes = netIOCounters[0].BytesSent
		s.networkStats.Timestamp = currentTime

		// Update info
		info["system"].(gin.H)["network"].(gin.H)["rx_bytes"] = netIOCounters[0].BytesRecv
		info["system"].(gin.H)["network"].(gin.H)["tx_bytes"] = netIOCounters[0].BytesSent
		info["system"].(gin.H)["network"].(gin.H)["rx_rate"] = rxRate
		info["system"].(gin.H)["network"].(gin.H)["tx_rate"] = txRate
	} else {
		s.logger.Warnf("Failed to get network info: %v", err)
	}

	// Get host info
	if hostInfo, err := host.Info(); err == nil {
		info["uptime"] = hostInfo.Uptime
		info["system"].(gin.H)["platform"] = hostInfo.Platform
		info["system"].(gin.H)["arch"] = hostInfo.KernelArch
	} else {
		s.logger.Warnf("Failed to get host info: %v", err)
		// Fallback to runtime info
		info["system"].(gin.H)["platform"] = runtime.GOOS
		info["system"].(gin.H)["arch"] = runtime.GOARCH
	}

	// Get network connections count
	if connections, err := net.Connections("all"); err == nil {
		info["system"].(gin.H)["network"].(gin.H)["connections"] = len(connections)
	} else {
		s.logger.Warnf("Failed to get network connections: %v", err)
	}

	c.JSON(http.StatusOK, info)
}

// handleGetSystemStatus handles GET /api/system/status
func (s *Server) handleGetSystemStatus(c *gin.Context) {
	// Get services status
	services := []gin.H{
		{
			"name":       "sing-box",
			"is_running": s.manager.IsRunning(),
			"version":    s.manager.GetVersion(),
			"uptime":     s.manager.GetUptime(),
		},
	}

	// Get subscriptions status
	subscriptions, err := s.storage.GetSubscriptions()
	if err != nil {
		s.logger.Warnf("Failed to get subscriptions: %v", err)
		subscriptions = []models.Subscription{}
	}

	// Convert subscriptions to map for response
	subsResponse := make([]gin.H, 0, len(subscriptions))
	for _, sub := range subscriptions {
		subsResponse = append(subsResponse, gin.H{
			"id":              sub.ID,
			"name":            sub.Name,
			"url":             sub.URL,
			"node_count":      sub.NodeCount,
			"last_update":     sub.LastUpdate,
			"expire_time":     sub.ExpireTime,
			"auto_update":     sub.AutoUpdate,
			"update_interval": sub.UpdateInterval,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"services":      services,
		"subscriptions": subsResponse,
	})
}

// handleGetNodes handles GET /api/nodes
func (s *Server) handleGetNodes(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 获取所有节点
	allNodes, err := s.storage.GetNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取每个节点的健康状态
	for i := range allNodes {
		health := s.proxy.GetNodeHealth(allNodes[i].ID)
		if health != nil {
			allNodes[i].Status = health.Status
			allNodes[i].Latency = health.Latency
			// 只有当健康检查的时间比数据库中的时间更新时，才更新 CheckedAt
			if health.LastCheck.After(allNodes[i].CheckedAt) {
				allNodes[i].CheckedAt = health.LastCheck
			}
		} else {
			allNodes[i].Status = "offline"
			allNodes[i].Latency = 0
			// 如果没有健康状态数据保持数据库中的 CheckedAt 不变
		}
		s.logger.Debugf("Node %s status: %s, latency: %d, checked at: %v",
			allNodes[i].Name, allNodes[i].Status, allNodes[i].Latency, allNodes[i].CheckedAt)
	}

	// 延迟排序，离线节点放在最后
	sort.Slice(allNodes, func(i, j int) bool {
		if allNodes[i].Status == "offline" && allNodes[j].Status == "offline" {
			return false
		}
		if allNodes[i].Status == "offline" {
			return false
		}
		if allNodes[j].Status == "offline" {
			return true
		}
		return allNodes[i].Latency < allNodes[j].Latency
	})

	// 计算分页
	total := int64(len(allNodes))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(allNodes) {
		c.JSON(http.StatusOK, gin.H{
			"nodes": []models.Node{},
			"pagination": gin.H{
				"total":    total,
				"page":     page,
				"pageSize": pageSize,
			},
		})
		return
	}
	if end > len(allNodes) {
		end = len(allNodes)
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": allNodes[start:end],
		"pagination": gin.H{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
	})
}

// handleCreateNode handles POST /api/nodes
func (s *Server) handleCreateNode(c *gin.Context) {
	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置新节点的 ID 和时间戳
	node.ID = uuid.New().String()
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()

	if err := s.storage.SaveNode(&node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("Failed to regenerate config: %v", err)
	}

	c.JSON(http.StatusOK, node)
}

// handleGetNode handles GET /api/nodes/:id
func (s *Server) handleGetNode(c *gin.Context) {
	id := c.Param("id")
	node, err := s.storage.GetNodeByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}
	c.JSON(http.StatusOK, node)
}

// handleUpdateNode handles PUT /api/nodes/:id
func (s *Server) handleUpdateNode(c *gin.Context) {
	id := c.Param("id")
	var node models.Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置节点 ID 和更新时间
	node.ID = id
	node.UpdatedAt = time.Now()

	if err := s.storage.SaveNode(&node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("Failed to regenerate config: %v", err)
	}

	c.JSON(http.StatusOK, node)
}

// handleDeleteNode handles DELETE /api/nodes/:id
func (s *Server) handleDeleteNode(c *gin.Context) {
	id := c.Param("id")
	if err := s.storage.DeleteNode(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("Failed to regenerate config: %v", err)
	}

	c.Status(http.StatusNoContent)
}

// handleGetNodeGroups handles GET /api/node-groups
func (s *Server) handleGetNodeGroups(c *gin.Context) {
	groups, err := s.storage.GetNodeGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 确保每个组都有正确的节点数量
	for i := range groups {
		s.logger.Infof("Group %s has %d nodes", groups[i].Name, groups[i].NodeCount)
	}

	c.JSON(http.StatusOK, groups)
}

// handleCreateNodeGroup handles POST /api/node-groups
func (s *Server) handleCreateNodeGroup(c *gin.Context) {
	var group models.NodeGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置新节点组的 ID 和时间戳
	group.ID = uuid.New().String()
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()

	if err := s.storage.SaveNodeGroup(&group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("Failed to regenerate config: %v", err)
	}

	c.JSON(http.StatusOK, group)
}

// handleGetNodeGroup handles GET /api/node-groups/:id
func (s *Server) handleGetNodeGroup(c *gin.Context) {
	id := c.Param("id")
	group, err := s.storage.GetNodeGroupByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node group not found"})
		return
	}
	c.JSON(http.StatusOK, group)
}

// handleUpdateNodeGroup handles PUT /api/node-groups/:id
func (s *Server) handleUpdateNodeGroup(c *gin.Context) {
	id := c.Param("id")
	var group models.NodeGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置节点组 ID 和更新时间
	group.ID = id
	group.UpdatedAt = time.Now()

	if err := s.storage.SaveNodeGroup(&group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("Failed to regenerate config: %v", err)
	}

	c.JSON(http.StatusOK, group)
}

// handleDeleteNodeGroup handles DELETE /api/node-groups/:id
func (s *Server) handleDeleteNodeGroup(c *gin.Context) {
	id := c.Param("id")
	if err := s.storage.DeleteNodeGroup(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("Failed to regenerate config: %v", err)
	}

	c.Status(http.StatusNoContent)
}

// handleGetRules handles GET /api/rules
func (s *Server) handleGetRules(c *gin.Context) {
	rules, err := s.storage.GetRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rules)
}

// handleCreateRule handles POST /api/rules
func (s *Server) handleCreateRule(c *gin.Context) {
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置规则 ID 和创建时间
	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	// 验证规则类型和值
	switch rule.Type {
	case "domain":
		if len(rule.Values) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "domains are required for domain rule"})
			return
		}
	case "ip":
		if len(rule.Values) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "IPs are required for IP rule"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported rule type: %s", rule.Type)})
		return
	}

	if err := s.storage.SaveRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("Failed to regenerate config: %v", err)
	}

	c.JSON(http.StatusOK, rule)
}

// handleGetRule handles GET /api/rules/:id
func (s *Server) handleGetRule(c *gin.Context) {
	id := c.Param("id")
	rule, err := s.storage.GetRuleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule not found"})
		return
	}
	c.JSON(http.StatusOK, rule)
}

// handleUpdateRule handles PUT /api/rules/:id
func (s *Server) handleUpdateRule(c *gin.Context) {
	id := c.Param("id")
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置规则 ID 和更新时间
	rule.ID = id
	rule.UpdatedAt = time.Now()

	// 验证规则类型和值
	switch rule.Type {
	case "domain":
		if len(rule.Values) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "domains are required for domain rule"})
			return
		}
	case "ip":
		if len(rule.Values) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "IPs are required for IP rule"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported rule type: %s", rule.Type)})
		return
	}

	if err := s.storage.SaveRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("Failed to regenerate config: %v", err)
	}

	c.JSON(http.StatusOK, rule)
}

// handleDeleteRule handles DELETE /api/rules/:id
func (s *Server) handleDeleteRule(c *gin.Context) {
	id := c.Param("id")
	if err := s.storage.DeleteRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("Failed to regenerate config: %v", err)
	}

	c.Status(http.StatusNoContent)
}

// handleGetRuleSets handles GET /api/rulesets
func (s *Server) handleGetRuleSets(c *gin.Context) {
	// 从数据库获取规则集列表
	ruleSets, err := s.storage.GetRuleSets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果数据库中没有规则集，尝试初始化
	if len(ruleSets) == 0 {
		if err := config.InitializeRuleSets(s.storage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to initialize rule sets: %v", err)})
			return
		}
		// 重新获取规则集列表
		ruleSets, err = s.storage.GetRuleSets()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, ruleSets)
}

// handleCreateRuleSet handles POST /api/rulesets
func (s *Server) handleCreateRuleSet(c *gin.Context) {
	var ruleSet models.RuleSet
	if err := c.ShouldBindJSON(&ruleSet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 使用前端传入的ID
	if ruleSet.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rule set ID is required"})
		return
	}
	ruleSet.UpdatedAt = time.Now()

	// 创建规则集目录
	rulesDir := "configs/sing-box/rules"
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create rules directory: %v", err)})
		return
	}

	// 如果是 geosite 或 geoip 类型的规则集，下载规则文件
	if ruleSet.Type == "geosite" || ruleSet.Type == "geoip" {
		// 确保使用正确的文件路径
		ruleSet.Path = fmt.Sprintf("configs/sing-box/rules/%s.srs", ruleSet.ID)
		out, err := os.Create(ruleSet.Path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create file: %v", err)})
			return
		}
		defer out.Close()

		// 下载文件
		resp, err := http.Get(ruleSet.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to download rule set: %v", err)})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to download rule set: status code %d", resp.StatusCode)})
			return
		}

		// 写入文件
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to save rule set: %v", err)})
			return
		}
	}

	if err := s.storage.SaveRuleSet(&ruleSet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("重新生成配置文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("重新生成配置文件失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, ruleSet)
}

// handleGetRuleSet handles GET /api/rulesets/:id
func (s *Server) handleGetRuleSet(c *gin.Context) {
	id := c.Param("id")
	ruleSet, err := s.storage.GetRuleSetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule set not found"})
		return
	}
	c.JSON(http.StatusOK, ruleSet)
}

// handleUpdateRuleSet handles PUT /api/rulesets/:id
func (s *Server) handleUpdateRuleSet(c *gin.Context) {
	id := c.Param("id")
	var ruleSet models.RuleSet
	if err := c.ShouldBindJSON(&ruleSet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查规则集是否存在
	existingRuleSet, err := s.storage.GetRuleSetByID(id)
	if err != nil {
		// 如果规则集不存在，创建新的规则集
		ruleSet.ID = id
		ruleSet.UpdatedAt = time.Now()

		// 确保规则集目录存在
		rulesDir := "configs/sing-box/rules"
		if err := os.MkdirAll(rulesDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create rules directory: %v", err)})
			return
		}

		// 设置文件路径
		ruleSet.Path = fmt.Sprintf("configs/sing-box/rules/%s.srs", ruleSet.ID)

		// 如果是远程规则集，下载规则文件
		if ruleSet.Type == "geosite" || ruleSet.Type == "geoip" {
			// 创建目标文件
			out, err := os.Create(ruleSet.Path)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create file: %v", err)})
				return
			}
			defer out.Close()

			// 下载文件
			resp, err := http.Get(ruleSet.URL)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to download rule set: %v", err)})
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to download rule set: status code %d", resp.StatusCode)})
				return
			}

			// 写入文件
			_, err = io.Copy(out, resp.Body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to save rule set: %v", err)})
				return
			}
		}

		if err := s.storage.SaveRuleSet(&ruleSet); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// 如果规则集存在，更新现有规则集
		ruleSet.ID = id
		ruleSet.UpdatedAt = time.Now()
		ruleSet.Path = existingRuleSet.Path // 保持原有的文件路径

		if err := s.storage.SaveRuleSet(&ruleSet); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("重新生成配置文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("重新生成配置文件失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, ruleSet)
}

// handleDeleteRuleSet handles DELETE /api/rulesets/:id
func (s *Server) handleDeleteRuleSet(c *gin.Context) {
	id := c.Param("id")

	// 删除规则文件
	filePath := fmt.Sprintf("configs/sing-box/rules/%s.srs", id)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		s.logger.Warnf("删除规则文件失败: %v", err)
	}

	if err := s.storage.DeleteRuleSet(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("重新生成配置文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("重新生成配置文件失败: %v", err)})
		return
	}

	c.Status(http.StatusNoContent)
}

// handleUpdateRuleSetRules handles POST /api/rulesets/:id/update
func (s *Server) handleUpdateRuleSetRules(c *gin.Context) {
	id := c.Param("id")
	ruleSet, err := s.storage.GetRuleSetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "rule set not found"})
		return
	}

	// 创建规则集目录
	rulesDir := "configs/sing-box/rules"
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create rules directory: %v", err)})
		return
	}

	// 下载规则集文件
	filePath := fmt.Sprintf("%s/%s.srs", rulesDir, ruleSet.ID)
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create file: %v", err)})
		return
	}
	defer out.Close()

	// 下载文件
	resp, err := http.Get(ruleSet.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to download rule set: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to download rule set: status code %d", resp.StatusCode)})
		return
	}

	// 写入文件
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to save rule set: %v", err)})
		return
	}

	// 更新规则集时间戳
	ruleSet.UpdatedAt = time.Now()
	if err := s.storage.SaveRuleSet(ruleSet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update rule set: %v", err)})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("重新生成配置文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("重新生成配置文件失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rule set updated successfully"})
}

// handleLogin handles POST /api/auth/login
func (s *Server) handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := s.config.Auth.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// handleLogout handles POST /api/auth/logout
func (s *Server) handleLogout(c *gin.Context) {
	// Currently, we just return success as we're using stateless JWT tokens
	c.Status(http.StatusNoContent)
}

// handleUpdatePassword handles PUT /api/settings/password
func (s *Server) handleUpdatePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user
	user := c.MustGet("user").(*models.User)

	// Update password
	if err := s.config.Auth.ChangePassword(user.Username, req.CurrentPassword, req.NewPassword); err != nil {
		if err.Error() == "密码错误" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "当前密码错误"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "密码修改成功",
		"requireRelogin": true,
	})
}

// authMiddleware verifies the JWT token in the Authorization header
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for login route
		if c.Request.URL.Path == "/api/auth/login" {
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		username, err := s.config.Auth.VerifyToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Get user from storage
		user, err := s.storage.GetUser(username)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		// Store user in context for later use
		c.Set("user", user)
		c.Set("username", username)
		c.Next()
	}
}

// handleGetSubscriptions handles GET /api/subscriptions
func (s *Server) handleGetSubscriptions(c *gin.Context) {
	subscriptions, err := s.storage.GetSubscriptions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subscriptions)
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// refreshSubscription refreshes a subscription
func (s *Server) refreshSubscription(sub *models.Subscription) error {
	s.logger.Infof("开始刷新订阅: %s (%s)", sub.Name, sub.ID)
	s.logger.Debugf("订阅详情: url=%s, active=%v", sub.URL, sub.Active)

	// 下载订阅内容
	s.logger.Debugf("从 URL 下载订阅内容: %s", sub.URL)
	resp, err := http.Get(sub.URL)
	if err != nil {
		s.logger.Errorf("下载订阅失败: %v", err)
		return fmt.Errorf("下载订阅失败: %v", err)
	}
	defer resp.Body.Close()

	s.logger.Debugf("订阅下载响应: status=%d, content-type=%s",
		resp.StatusCode, resp.Header.Get("Content-Type"))

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorf("订阅 URL 返回非 200 状态码: %d", resp.StatusCode)
		return fmt.Errorf("订阅 URL 返回状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Errorf("读取订阅内容失败: %v", err)
		return fmt.Errorf("读取订阅内容失败: %v", err)
	}

	s.logger.Debugf("成功下载订阅内容 (长度: %d 字节)", len(body))
	s.logger.Debugf("内容预览: %s", string(body[:min(len(body), 200)]))

	// 检测订阅类型
	detectedType := subscription.DetectSubscriptionType(body)
	s.logger.Infof("检测订阅类型: %s", detectedType)

	// 解析订阅内容
	nodes, err := subscription.ParseSubscription(body, detectedType)
	if err != nil {
		s.logger.Errorf("解析订阅内容失败: %v", err)
		return fmt.Errorf("解析订阅内容失败: %v", err)
	}

	s.logger.Infof("成功解析 %d 个节点", len(nodes))

	// 更新订阅信息
	sub.LastUpdate = time.Now()
	sub.NodeCount = len(nodes)
	sub.Type = detectedType // 更新订阅类型为检测到的类型
	if err := s.storage.SaveSubscription(sub); err != nil {
		return fmt.Errorf("保存订阅信息失败: %v", err)
	}

	// 删除旧节点
	if err := s.storage.DeleteNodesBySubscriptionID(sub.ID); err != nil {
		s.logger.Errorf("删除旧节点失败: %v", err)
		return fmt.Errorf("删除旧节点失败: %v", err)
	}

	// 添加新节点
	for _, node := range nodes {
		// URL 解码节点名称
		if decodedName, err := url.QueryUnescape(node.Name); err == nil {
			node.Name = decodedName
		}

		// 限制名称长度
		if len(node.Name) > 50 {
			node.Name = node.Name[:47] + "..."
		}

		// 设置节点的订阅 ID 和其他必要字段
		node.ID = "" // 确保是新节点，这样会触发自动匹配逻辑
		node.SubscriptionID = sub.ID
		node.CreatedAt = time.Now()
		node.UpdatedAt = time.Now()

		// 测试节点延迟
		checker := protocols.NewNodeChecker()
		result := checker.CheckNode(node)
		if result.Available {
			node.Status = "online"
			node.Latency = result.Latency
		} else {
			node.Status = "offline"
			node.Latency = 0
		}
		node.CheckedAt = time.Now()

		// 保存节点
		if err := s.storage.SaveNode(node); err != nil {
			s.logger.Errorf("保存节点失败: %v", err)
			return fmt.Errorf("保存节点失败: %v", err)
		}
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("重新生成配置文件失败: %v", err)
		return fmt.Errorf("重新生成配置文件失败: %v", err)
	}

	s.logger.Infof("订阅刷新完成，成功添加 %d 个节点", len(nodes))
	return nil
}

// handleCreateSubscription handles POST /api/subscriptions
func (s *Server) handleCreateSubscription(c *gin.Context) {
	var subscription models.Subscription
	if err := c.ShouldBindJSON(&subscription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证订阅数据
	if err := subscription.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription.ID = uuid.New().String()
	subscription.CreatedAt = time.Now()
	subscription.UpdatedAt = time.Now()

	// 如果没指定 active 段，默认为 true
	if !subscription.Active {
		subscription.Active = true
	}

	// 保存订阅
	if err := s.storage.SaveSubscription(&subscription); err != nil {
		s.logger.Errorf("Failed to save subscription: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to save subscription: %v", err)})
		return
	}

	// 如果订阅是激活的，即刷新一次
	if subscription.Active {
		if err := s.refreshSubscription(&subscription); err != nil {
			s.logger.Warnf("Failed to refresh subscription: %v", err)
		}
	}

	c.JSON(http.StatusOK, subscription)
}

// handleGetSubscription handles GET /api/subscriptions/:id
func (s *Server) handleGetSubscription(c *gin.Context) {
	id := c.Param("id")
	subscription, err := s.storage.GetSubscriptionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// handleUpdateSubscription handles PUT /api/subscriptions/:id
func (s *Server) handleUpdateSubscription(c *gin.Context) {
	id := c.Param("id")
	var subscription models.Subscription
	if err := c.ShouldBindJSON(&subscription); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证订阅数据
	if err := subscription.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取有订阅
	existingSubscription, err := s.storage.GetSubscriptionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	subscription.ID = id
	subscription.CreatedAt = existingSubscription.CreatedAt
	subscription.UpdatedAt = time.Now()
	subscription.LastUpdate = existingSubscription.LastUpdate
	subscription.NodeCount = existingSubscription.NodeCount

	if err := s.storage.SaveSubscription(&subscription); err != nil {
		s.logger.Errorf("Failed to save subscription: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to save subscription: %v", err)})
		return
	}

	// 如果订阅是激活的，即刷新一次
	if subscription.Active {
		if err := s.refreshSubscription(&subscription); err != nil {
			s.logger.Warnf("Failed to refresh subscription: %v", err)
		}
	}

	c.JSON(http.StatusOK, subscription)
}

// handleDeleteSubscription handles DELETE /api/subscriptions/:id
func (s *Server) handleDeleteSubscription(c *gin.Context) {
	id := c.Param("id")

	// 获取该订阅下的所有节点ID
	nodes, err := s.storage.GetNodes()
	if err != nil {
		s.logger.Errorf("Failed to get nodes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get nodes: %v", err)})
		return
	}

	// 收集要删除的节点ID
	nodeIDs := make([]string, 0)
	for _, node := range nodes {
		if node.SubscriptionID == id {
			nodeIDs = append(nodeIDs, node.ID)
		}
	}

	// 获取所有节点组
	groups, err := s.storage.GetNodeGroups()
	if err != nil {
		s.logger.Errorf("Failed to get node groups: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get node groups: %v", err)})
		return
	}

	// 从节点组中移除要删除的节点
	for i := range groups {
		needsUpdate := false
		updatedNodes := make([]models.Node, 0)

		for _, node := range groups[i].Nodes {
			// 检查节点是否属于要删除的订阅
			shouldKeep := true
			for _, idToDelete := range nodeIDs {
				if node.ID == idToDelete {
					shouldKeep = false
					needsUpdate = true
					break
				}
			}
			if shouldKeep {
				updatedNodes = append(updatedNodes, node)
			}
		}

		// 如果节点组需要更新
		if needsUpdate {
			groups[i].Nodes = updatedNodes
			groups[i].NodeCount = len(updatedNodes)
			if err := s.storage.SaveNodeGroup(&groups[i]); err != nil {
				s.logger.Errorf("Failed to update node group: %v", err)
				// 继续处理其他组，不中断流程
			}
		}
	}

	// 删除该订阅下的所有节点
	if err := s.storage.DeleteNodesBySubscriptionID(id); err != nil {
		s.logger.Errorf("Failed to delete subscription nodes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to delete subscription nodes: %v", err)})
		return
	}

	// 删除订阅
	if err := s.storage.DeleteSubscription(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// startSubscriptionUpdater starts the subscription auto-update goroutine
func (s *Server) startSubscriptionUpdater() {
	go func() {
		ticker := time.NewTicker(s.config.UpdateInterval)
		defer ticker.Stop()

		for range ticker.C {
			s.logger.Info("Starting auto-update of subscriptions")
			subscriptions, err := s.storage.GetSubscriptions()
			if err != nil {
				s.logger.Errorf("Failed to get subscriptions: %v", err)
				continue
			}

			for i := range subscriptions {
				if !subscriptions[i].AutoUpdate {
					continue
				}

				if err := s.refreshSubscription(&subscriptions[i]); err != nil {
					s.logger.Errorf("Failed to refresh subscription %s: %v", subscriptions[i].ID, err)
				}
			}
		}
	}()
}

// handleRefreshSubscription handles POST /api/subscriptions/:id/refresh
func (s *Server) handleRefreshSubscription(c *gin.Context) {
	id := c.Param("id")
	subscription, err := s.storage.GetSubscriptionByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
		return
	}

	if err := s.refreshSubscription(subscription); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to refresh subscription: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription refreshed successfully"})
}

// handleTestNodes handles POST /api/nodes/test
func (s *Server) handleTestNodes(c *gin.Context) {
	// 获取所有节点
	nodes, err := s.storage.GetNodes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 创建节点检查器
	checker := protocols.NewNodeChecker()

	// 发测试所有节点
	results := checker.CheckNodes(nodes)

	// 准备测速结果
	testResults := make([]map[string]interface{}, 0)

	// 记录当前时间作为测速时间
	testTime := time.Now()

	// 更新节点状态和延迟
	for _, result := range results {
		if result.Node != nil {
			// 更新节点状态
			node := result.Node
			if result.Available {
				node.Status = "online"
				node.Latency = result.Latency
			} else {
				node.Status = "offline"
				node.Latency = 0
			}
			node.CheckedAt = testTime
			node.UpdatedAt = testTime

			// 保存节点状态到数据库
			if err := s.storage.SaveNode(node); err != nil {
				s.logger.Warnf("Failed to save node status: %v", err)
			}

			// 更新健康状态缓存
			s.proxy.UpdateNodeHealth(node.ID, node.Status, node.Latency, node.CheckedAt, result.Error)

			// 记录测速结果
			testResult := map[string]interface{}{
				"id":        node.ID,
				"name":      node.Name,
				"type":      node.Type,
				"address":   node.Address,
				"port":      node.Port,
				"status":    node.Status,
				"latency":   node.Latency,
				"error":     result.Error,
				"checkedAt": node.CheckedAt,
			}
			testResults = append(testResults, testResult)
		}
	}

	// 按延迟排序，离线节点放在最后
	sort.Slice(testResults, func(i, j int) bool {
		statusI := testResults[i]["status"].(string)
		statusJ := testResults[j]["status"].(string)
		if statusI == "offline" && statusJ == "offline" {
			return false
		}
		if statusI == "offline" {
			return false
		}
		if statusJ == "offline" {
			return true
		}
		return testResults[i]["latency"].(int64) < testResults[j]["latency"].(int64)
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "nodes tested successfully",
		"results": testResults,
	})
}

// handleGetSettings handles GET /api/settings
func (s *Server) handleGetSettings(c *gin.Context) {
	settings, err := s.storage.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// handleUpdateSettings handles PUT /api/settings
func (s *Server) handleUpdateSettings(c *gin.Context) {
	var req struct {
		ThemeMode   string                      `json:"theme_mode"`
		SingboxMode *models.SingboxModeSettings `json:"singbox_mode"`
		DNS         *models.DNSSettings         `json:"dns"`
		Dashboard   *models.DashboardSettings   `json:"dashboard"`
		InboundMode string                      `json:"inbound_mode"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := s.storage.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.ThemeMode != "" {
		settings.ThemeMode = req.ThemeMode
	}

	if req.SingboxMode != nil {
		if err := settings.SetSingboxMode(req.SingboxMode); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if req.InboundMode != "" {
		if err := settings.SetInboundMode(req.InboundMode); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if req.Dashboard != nil {
		if err := settings.SetDashboard(req.Dashboard); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if req.DNS != nil {
		// 保存 DNS 设置
		if err := s.storage.SaveDNSSettings(req.DNS); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// 保存设置
	if err := s.storage.UpdateSettings(settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("重新生成配置文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("重新生成配置文件失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// handleStartService handles POST /api/system/services/:name/start
func (s *Server) handleStartService(c *gin.Context) {
	name := c.Param("name")
	s.logger.WithField("service", name).Info("Starting service")

	if err := s.proxy.StartService(name); err != nil {
		s.logger.WithError(err).Error("Failed to start service")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("Failed to start service: %v", err),
			"service": name,
		})
		return
	}

	c.Status(http.StatusOK)
}

// handleStopService handles POST /api/system/services/:name/stop
func (s *Server) handleStopService(c *gin.Context) {
	name := c.Param("name")
	if err := s.manager.StopService(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s stopped successfully", name)})
}

// handleRestartService handles POST /api/system/services/:name/restart
func (s *Server) handleRestartService(c *gin.Context) {
	name := c.Param("name")
	if err := s.manager.RestartService(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s restarted successfully", name)})
}

// regenerateConfig regenerates the sing-box config file
func (s *Server) regenerateConfig() error {
	// 生成 sing-box 配置
	generator := config.NewSingBoxGenerator(s.storage)
	data, err := generator.GenerateConfig()
	if err != nil {
		return fmt.Errorf("generate sing-box config: %w", err)
	}

	// 确保配置目录存在
	configDir := filepath.Join("configs", "sing-box")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// 写入配置文件
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	s.logger.Infof("sing-box config file generated: %s", configPath)
	return nil
}

// handleGetDNSRules handles GET /api/dns/rules
func (s *Server) handleGetDNSRules(c *gin.Context) {
	// 获取 DNS 规则
	rules, err := s.storage.GetDNSRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取 DNS 设置
	settings, err := s.storage.GetDNSSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules":    rules,
		"settings": settings,
	})
}

// handleCreateDNSRule handles POST /api/dns/rules
func (s *Server) handleCreateDNSRule(c *gin.Context) {
	var rule models.DNSRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证规则
	if err := rule.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置规则 ID
	rule.ID = uuid.New().String()

	if err := s.storage.SaveDNSRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("重新生成配置文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("重新生成配置文件失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// handleUpdateDNSRule handles PUT /api/dns/rules/:id
func (s *Server) handleUpdateDNSRule(c *gin.Context) {
	id := c.Param("id")
	var rule models.DNSRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证规则
	if err := rule.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置规则 ID
	rule.ID = id

	if err := s.storage.SaveDNSRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("重新生成配置文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("重新生成配置文件失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// handleDeleteDNSRule handles DELETE /api/dns/rules/:id
func (s *Server) handleDeleteDNSRule(c *gin.Context) {
	id := c.Param("id")
	if err := s.storage.DeleteDNSRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("重新生成配置文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("重新生成配置文件失败: %v", err)})
		return
	}

	c.Status(http.StatusNoContent)
}

// handleUpdateDNSSettings handles PUT /api/dns/settings
func (s *Server) handleUpdateDNSSettings(c *gin.Context) {
	var settings models.DNSSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.storage.SaveDNSSettings(&settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新生成配置文件
	if err := s.regenerateConfig(); err != nil {
		s.logger.Errorf("重新生成配置文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("重新生成配置文件失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// ... rest of the code ...
