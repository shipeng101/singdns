package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"singdns/api/auth"
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
func NewServer(storage storage.Storage, manager *proxy.Manager, logger *logrus.Logger, config *Config) *Server {
	server := &Server{
		router:       gin.Default(),
		manager:      manager,
		storage:      storage,
		logger:       logger,
		config:       config,
		networkStats: &NetworkStats{Timestamp: time.Now()}, // Initialize network stats
		proxy:        manager,                              // Initialize proxy field with the manager
	}

	// Configure CORS
	server.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add auth middleware
	server.router.Use(server.authMiddleware())

	// Create rule set updater
	server.updater = ruleset.NewUpdater(storage, logger)

	// Setup routes
	server.setupRoutes()

	// Start subscription auto-updater
	server.startSubscriptionUpdater()

	return server
}

// Start starts the API server
func (s *Server) Start() error {
	// Start rule set updater
	s.updater.Start(s.config.UpdateInterval)

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
			"name":       "singbox",
			"is_running": s.manager.IsSingboxRunning(),
			"version":    s.manager.GetSingboxVersion(),
			"uptime":     s.manager.GetSingboxUptime(),
		},
		{
			"name":       "mosdns",
			"is_running": s.manager.IsMosdnsRunning(),
			"version":    s.manager.GetMosdnsVersion(),
			"uptime":     s.manager.GetMosdnsUptime(),
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
			// 如果没有健康状态数据，保持数据库中的 CheckedAt 不变
		}
		s.logger.Debugf("Node %s status: %s, latency: %d, checked at: %v",
			allNodes[i].Name, allNodes[i].Status, allNodes[i].Latency, allNodes[i].CheckedAt)
	}

	// 按延迟排序，离线节点放在最后
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

	node.ID = uuid.New().String()
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()

	if err := s.storage.SaveNode(&node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	node.ID = id
	node.UpdatedAt = time.Now()

	if err := s.storage.SaveNode(&node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
	c.Status(http.StatusNoContent)
}

// handleGetNodeGroups handles GET /api/node-groups
func (s *Server) handleGetNodeGroups(c *gin.Context) {
	groups, err := s.storage.GetNodeGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	// 设置默认值
	group.ID = uuid.New().String()
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()
	if group.Mode == "" {
		group.Mode = "select"
	}
	if !group.Active {
		group.Active = true
	}

	// 保存点组
	if err := s.storage.SaveNodeGroup(&group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	// 获取现有的节点组
	existingGroup, err := s.storage.GetNodeGroupByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node group not found"})
		return
	}

	// 保留原有的节点关系
	group.ID = id
	group.CreatedAt = existingGroup.CreatedAt
	group.UpdatedAt = time.Now()
	group.Nodes = existingGroup.Nodes

	// 设置默认值
	if group.Mode == "" {
		group.Mode = "select"
	}
	if !group.Active {
		group.Active = true
	}

	// 保存更新后的节点组
	if err := s.storage.SaveNodeGroup(&group); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载节点组以获取最新状态
	updatedGroup, err := s.storage.GetNodeGroupByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reload node group"})
		return
	}

	c.JSON(http.StatusOK, updatedGroup)
}

// handleDeleteNodeGroup handles DELETE /api/node-groups/:id
func (s *Server) handleDeleteNodeGroup(c *gin.Context) {
	id := c.Param("id")
	if err := s.storage.DeleteNodeGroup(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	rule.ID = uuid.New().String()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	if err := s.storage.SaveRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	rule.ID = id
	rule.UpdatedAt = time.Now()

	if err := s.storage.SaveRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
	c.Status(http.StatusNoContent)
}

// handleGetRuleSets handles GET /api/rulesets
func (s *Server) handleGetRuleSets(c *gin.Context) {
	ruleSets, err := s.storage.GetRuleSets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	ruleSet.ID = uuid.New().String()
	ruleSet.UpdatedAt = time.Now()

	if err := s.storage.SaveRuleSet(&ruleSet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	ruleSet.ID = id
	ruleSet.UpdatedAt = time.Now()

	if err := s.storage.SaveRuleSet(&ruleSet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ruleSet)
}

// handleDeleteRuleSet handles DELETE /api/rulesets/:id
func (s *Server) handleDeleteRuleSet(c *gin.Context) {
	id := c.Param("id")
	if err := s.storage.DeleteRuleSet(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	// Download rule set
	data, err := ruleset.DownloadRuleSet(ruleSet.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to download rule set: %v", err)})
		return
	}

	// Parse rule set
	parser := ruleset.NewParser(ruleSet.Type)
	if parser == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unsupported rule set type"})
		return
	}

	rules, err := parser.Parse(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to parse rule set: %v", err)})
		return
	}

	// Create or update rule
	rule := &models.Rule{
		ID:          fmt.Sprintf("ruleset-%s", ruleSet.ID),
		Name:        ruleSet.Name,
		Type:        ruleSet.Format,
		Outbound:    ruleSet.Outbound,
		Description: ruleSet.Description,
		Enabled:     ruleSet.Enabled,
		Priority:    100, // Default priority for rule sets
		UpdatedAt:   time.Now(),
	}

	// Set domains or IPs based on format
	if ruleSet.Format == "domain" {
		rule.Domains = rules
	} else {
		rule.IPs = rules
	}

	// 使用事务保存规则和规则集
	storage := s.storage.(*storage.SQLiteStorage)
	tx := storage.GetDB().Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to start transaction: %v", tx.Error)})
		return
	}

	// 保存规则
	if err := tx.Save(rule).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to save rule: %v", err)})
		return
	}

	// 更新规则集时间戳
	ruleSet.UpdatedAt = time.Now()
	if err := tx.Save(ruleSet).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update rule set: %v", err)})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to commit transaction: %v", err)})
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

// handleUpdatePassword handles POST /api/auth/password
func (s *Server) handleUpdatePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.config.Auth.UpdatePassword(req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
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

		// Store username in context for later use
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

		// 设置节点的订阅 ID
		node.ID = uuid.New().String()
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
		node.CheckedAt = result.CheckedAt

		if err := s.storage.SaveNode(node); err != nil {
			s.logger.Errorf("保存节点失败: %v", err)
			continue
		}
	}

	s.logger.Infof("订阅刷新完成，成功保存 %d 个节点", len(nodes))
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

	// 如果订阅是激活的，立即刷新一次
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

	// 如果订阅是激活的，立即刷新一次
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

	// 先删除该订阅下的所有节点
	if err := s.storage.DeleteNodesBySubscriptionID(id); err != nil {
		s.logger.Errorf("Failed to delete subscription nodes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to delete subscription nodes: %v", err)})
		return
	}

	// 然后删除订阅
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

	// 并发测试所有节点
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

// ... rest of the code ...
