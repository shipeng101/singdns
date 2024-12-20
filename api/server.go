package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Server represents the API server
type Server struct {
	router        *gin.Engine
	authManager   *AuthManager
	configManager *ConfigManager
	systemManager *SystemManager
	proxyManager  *ProxyManager
}

// NewServer creates a new API server
func NewServer(configManager *ConfigManager) *Server {
	jwtKey := []byte("your-secret-key")
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Create managers
	authManager := NewAuthManager(jwtKey)
	systemManager := NewSystemManager()
	proxyManager := NewProxyManager()

	server := &Server{
		router:        router,
		authManager:   authManager,
		configManager: configManager,
		systemManager: systemManager,
		proxyManager:  proxyManager,
	}

	// Register routes
	server.setupRoutes()

	return server
}

// Run starts the API server
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

// setupRoutes registers all API routes
func (s *Server) setupRoutes() {
	// Public routes
	s.router.POST("/api/auth/login", s.handleLogin)
	s.router.POST("/api/auth/register", s.handleRegister)

	// Protected routes
	api := s.router.Group("/api")
	api.Use(s.authMiddleware())

	// User routes
	api.GET("/user", s.handleGetUser)
	api.PUT("/user/password", s.handleUpdatePassword)

	// System routes
	api.GET("/system/status", s.handleGetSystemStatus)
	api.GET("/system/services", s.handleGetServices)
	api.POST("/system/services/:name/start", s.handleStartService)
	api.POST("/system/services/:name/stop", s.handleStopService)
	api.POST("/system/services/:name/restart", s.handleRestartService)

	// Node routes
	api.GET("/nodes", s.handleListNodes)
	api.POST("/nodes", s.handleCreateNode)
	api.PUT("/nodes/:id", s.handleUpdateNode)
	api.DELETE("/nodes/:id", s.handleDeleteNode)

	// Rule routes
	api.GET("/rules", s.handleListRules)
	api.POST("/rules", s.handleCreateRule)
	api.PUT("/rules/:id", s.handleUpdateRule)
	api.DELETE("/rules/:id", s.handleDeleteRule)

	// Subscription routes
	api.GET("/subscriptions", s.handleListSubscriptions)
	api.POST("/subscriptions", s.handleCreateSubscription)
	api.PUT("/subscriptions/:id", s.handleUpdateSubscription)
	api.DELETE("/subscriptions/:id", s.handleDeleteSubscription)
	api.POST("/subscriptions/:id/update", s.handleUpdateSubscriptionNodes)

	// Settings routes
	api.GET("/settings", s.handleGetSettings)
	api.PUT("/settings", s.handleUpdateSettings)
}

// authMiddleware validates JWT token
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := s.authManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// handleLogin handles user login
func (s *Server) handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := s.authManager.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// handleRegister handles user registration
func (s *Server) handleRegister(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.authManager.Register(req.Username, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// handleGetUser handles get user info
func (s *Server) handleGetUser(c *gin.Context) {
	username := c.GetString("username")
	user, err := s.authManager.GetUser(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// handleUpdateUser handles update user info
func (s *Server) handleUpdateUser(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.authManager.UpdateUser(userID, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// handleListNodes handles list nodes
func (s *Server) handleListNodes(c *gin.Context) {
	nodes := s.proxyManager.GetNodes()
	c.JSON(http.StatusOK, nodes)
}

// handleCreateNode handles create node
func (s *Server) handleCreateNode(c *gin.Context) {
	var node Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.proxyManager.CreateNode(&node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, node)
}

// handleUpdateNode handles update node
func (s *Server) handleUpdateNode(c *gin.Context) {
	nodeID := c.Param("id")
	var node Node
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	node.ID = nodeID
	if err := s.proxyManager.UpdateNode(&node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, node)
}

// handleDeleteNode handles delete node
func (s *Server) handleDeleteNode(c *gin.Context) {
	nodeID := c.Param("id")
	if err := s.proxyManager.DeleteNode(nodeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// handleListRules handles list rules
func (s *Server) handleListRules(c *gin.Context) {
	rules := s.proxyManager.GetRules()
	c.JSON(http.StatusOK, rules)
}

// handleCreateRule handles create rule
func (s *Server) handleCreateRule(c *gin.Context) {
	var rule Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.proxyManager.CreateRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// handleUpdateRule handles update rule
func (s *Server) handleUpdateRule(c *gin.Context) {
	ruleID := c.Param("id")
	var rule Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule.ID = ruleID
	if err := s.proxyManager.UpdateRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// handleDeleteRule handles delete rule
func (s *Server) handleDeleteRule(c *gin.Context) {
	ruleID := c.Param("id")
	if err := s.proxyManager.DeleteRule(ruleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// handleListSubscriptions handles list subscriptions
func (s *Server) handleListSubscriptions(c *gin.Context) {
	subscriptions := s.proxyManager.GetSubscriptions()
	c.JSON(http.StatusOK, subscriptions)
}

// handleCreateSubscription handles create subscription
func (s *Server) handleCreateSubscription(c *gin.Context) {
	var sub Subscription
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.proxyManager.CreateSubscription(&sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// handleUpdateSubscription handles update subscription
func (s *Server) handleUpdateSubscription(c *gin.Context) {
	subID := c.Param("id")
	var sub Subscription
	if err := c.ShouldBindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub.ID = subID
	if err := s.proxyManager.UpdateSubscription(&sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sub)
}

// handleDeleteSubscription handles delete subscription
func (s *Server) handleDeleteSubscription(c *gin.Context) {
	subID := c.Param("id")
	if err := s.proxyManager.DeleteSubscription(subID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// handleUpdateSubscriptionNodes handles update subscription nodes
func (s *Server) handleUpdateSubscriptionNodes(c *gin.Context) {
	subID := c.Param("id")
	if err := s.proxyManager.UpdateSubscriptionNodes(subID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// handleGetSettings handles get settings
func (s *Server) handleGetSettings(c *gin.Context) {
	settings := s.configManager.GetSettings()
	c.JSON(http.StatusOK, settings)
}

// handleUpdateSettings handles update settings
func (s *Server) handleUpdateSettings(c *gin.Context) {
	var settings Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.configManager.UpdateSettings(&settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// handleGetSystemStatus handles get system status
func (s *Server) handleGetSystemStatus(c *gin.Context) {
	status, err := s.systemManager.GetStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// handleGetServices handles get services status
func (s *Server) handleGetServices(c *gin.Context) {
	services := s.systemManager.GetServices()
	c.JSON(http.StatusOK, services)
}

// handleStartService handles start service
func (s *Server) handleStartService(c *gin.Context) {
	name := c.Param("name")
	if err := s.systemManager.StartService(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// handleStopService handles stop service
func (s *Server) handleStopService(c *gin.Context) {
	name := c.Param("name")
	if err := s.systemManager.StopService(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// handleRestartService handles restart service
func (s *Server) handleRestartService(c *gin.Context) {
	name := c.Param("name")
	if err := s.systemManager.RestartService(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

// handleGetTrafficStats handles get traffic statistics
func (s *Server) handleGetTrafficStats(c *gin.Context) {
	stats := s.proxyManager.GetTrafficStats()
	c.JSON(http.StatusOK, stats)
}

// handleGetRealtimeTraffic handles get realtime traffic
func (s *Server) handleGetRealtimeTraffic(c *gin.Context) {
	traffic := s.proxyManager.GetRealtimeTraffic()
	c.JSON(http.StatusOK, traffic)
}

// handleCreateBackup handles create backup
func (s *Server) handleCreateBackup(c *gin.Context) {
	backupFile, err := s.configManager.BackupConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"file": backupFile})
}

// handleListBackups handles list backups
func (s *Server) handleListBackups(c *gin.Context) {
	// TODO: Implement list backups
	c.JSON(http.StatusOK, []string{})
}

// handleRestoreBackup handles restore backup
func (s *Server) handleRestoreBackup(c *gin.Context) {
	id := c.Param("id")
	if err := s.configManager.RestoreConfig(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// handleDeleteBackup handles delete backup
func (s *Server) handleDeleteBackup(c *gin.Context) {
	backupID := c.Param("id")
	if err := s.configManager.DeleteBackup(backupID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// handleUpdatePassword handles update user password
func (s *Server) handleUpdatePassword(c *gin.Context) {
	username := c.GetString("username")
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.authManager.UpdatePassword(username, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
