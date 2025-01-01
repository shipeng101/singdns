package proxy

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"singdns/api/config"
	"singdns/api/models"
	"singdns/api/storage"

	"crypto/tls"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Manager implements the ProxyManager interface
type Manager struct {
	logger         *logrus.Logger
	mu             sync.RWMutex
	cmd            *exec.Cmd
	startTime      time.Time
	isRunning      bool
	version        string
	configPath     string
	backupPath     string
	storage        storage.Storage
	autoSave       bool
	retryConfig    *RetryConfig
	cache          *Cache
	batchUpdate    *BatchUpdate
	crypto         *Crypto
	monitor        *Monitor
	nodes          map[string]*models.Node
	rules          map[string]*models.Rule
	subs           map[string]*models.Subscription
	singboxCmd     *exec.Cmd
	mosdnsCmd      *exec.Cmd
	singboxTime    time.Time
	mosdnsTime     time.Time
	mosdnsVersion  string
	singboxVersion string
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxRetries  int
	RetryDelay  time.Duration
	MaxInterval time.Duration
}

// Cache represents the cache system
type Cache struct {
	mu           sync.RWMutex
	configCache  []byte
	lastModified time.Time
	ttl          time.Duration
}

// BatchUpdate represents the batch update system
type BatchUpdate struct {
	mu       sync.Mutex
	timer    *time.Timer
	interval time.Duration
	pending  bool
	changes  []func() error
	maxBatch int
}

// Crypto handles encryption and decryption
type Crypto struct {
	key []byte
}

// Monitor handles service monitoring
type Monitor struct {
	mu         sync.RWMutex
	interval   time.Duration
	healthData map[string]models.HealthData
}

// NewManager creates a new proxy manager
func NewManager(logger *logrus.Logger, configPath string, storage storage.Storage) *Manager {
	// Generate random encryption key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		logger.WithError(err).Fatal("Failed to generate encryption key")
	}

	m := &Manager{
		logger:     logger,
		configPath: configPath,
		backupPath: configPath + ".backup",
		storage:    storage,
		autoSave:   true,
		nodes:      make(map[string]*models.Node),
		rules:      make(map[string]*models.Rule),
		subs:       make(map[string]*models.Subscription),
		retryConfig: &RetryConfig{
			MaxRetries:  3,
			RetryDelay:  time.Second * 5,
			MaxInterval: time.Minute * 5,
		},
		cache: &Cache{
			ttl: time.Minute * 5,
		},
		batchUpdate: &BatchUpdate{
			interval: time.Second * 30,
			maxBatch: 100,
		},
		crypto: &Crypto{
			key: key,
		},
		monitor: &Monitor{
			interval:   time.Minute,
			healthData: make(map[string]models.HealthData),
		},
	}

	// Load initial data from storage
	if err := m.loadData(); err != nil {
		logger.WithError(err).Error("Failed to load initial data")
	}

	// Initialize default node groups
	if err := m.initDefaultNodeGroups(); err != nil {
		logger.WithError(err).Error("Failed to initialize default node groups")
	}

	// Start health check
	go m.startHealthCheck()

	return m
}

// loadData loads data from storage
func (m *Manager) loadData() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load nodes
	nodes, err := m.storage.GetNodes()
	if err != nil {
		return fmt.Errorf("failed to load nodes: %v", err)
	}
	for i := range nodes {
		m.nodes[nodes[i].ID] = &nodes[i]
	}

	// Load rules
	rules, err := m.storage.GetRules()
	if err != nil {
		return fmt.Errorf("failed to load rules: %v", err)
	}
	for i := range rules {
		m.rules[rules[i].ID] = &rules[i]
	}

	// Load subscriptions
	subs, err := m.storage.GetSubscriptions()
	if err != nil {
		return fmt.Errorf("failed to load subscriptions: %v", err)
	}
	for i := range subs {
		m.subs[subs[i].ID] = &subs[i]
	}

	return nil
}

// startHealthCheck starts periodic health checks
func (m *Manager) startHealthCheck() {
	ticker := time.NewTicker(m.monitor.interval)
	defer ticker.Stop()

	for range ticker.C {
		m.checkHealth()
	}
}

// checkHealth performs health checks on all services
func (m *Manager) checkHealth() {
	m.monitor.mu.Lock()
	defer m.monitor.mu.Unlock()

	// Check sing-box service
	start := time.Now()
	if m.isRunning {
		if err := m.pingService(); err != nil {
			m.monitor.healthData["sing-box"] = models.HealthData{
				Status:    "error",
				Latency:   time.Since(start).Milliseconds(),
				LastCheck: time.Now(),
				Error:     err.Error(),
			}
		} else {
			m.monitor.healthData["sing-box"] = models.HealthData{
				Status:    "healthy",
				Latency:   time.Since(start).Milliseconds(),
				LastCheck: time.Now(),
			}
		}
	} else {
		m.monitor.healthData["sing-box"] = models.HealthData{
			Status:    "stopped",
			LastCheck: time.Now(),
		}
	}

	// Check nodes
	for _, node := range m.GetNodes() {
		if latency, err := m.pingNode(node); err != nil {
			m.monitor.healthData[node.ID] = models.HealthData{
				Status:    "offline",
				Latency:   0,
				LastCheck: time.Now(),
				Error:     err.Error(),
			}

			// 更新数据库中的节点状态
			node.Status = "offline"
			node.Latency = 0
			node.CheckedAt = time.Now()
			if err := m.storage.SaveNode(node); err != nil {
				m.logger.WithError(err).Error("Failed to update node status")
			}
		} else {
			m.monitor.healthData[node.ID] = models.HealthData{
				Status:    "online",
				Latency:   latency,
				LastCheck: time.Now(),
			}

			// 更新数据库中的节点状态
			node.Status = "online"
			node.Latency = latency
			node.CheckedAt = time.Now()
			if err := m.storage.SaveNode(node); err != nil {
				m.logger.WithError(err).Error("Failed to update node status")
			}
		}
	}
}

// pingService checks if sing-box service is responsive
func (m *Manager) pingService() error {
	// TODO: Implement actual service health check
	// For now, just check if process is running
	if m.cmd == nil || m.cmd.Process == nil {
		return fmt.Errorf("service process not found")
	}
	return nil
}

// pingNode checks if a node is responsive
func (m *Manager) pingNode(node *models.Node) (int64, error) {
	if node == nil {
		return 0, fmt.Errorf("node is nil")
	}

	// Create TCP connection with timeout
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", node.Address, node.Port), time.Second*5)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	// Try to establish TLS connection for secure protocols
	if node.Type == "trojan" || node.TLS {
		tlsConn := tls.Client(conn, &tls.Config{
			ServerName:         node.Host,
			InsecureSkipVerify: true,
		})
		if err := tlsConn.Handshake(); err != nil {
			return 0, err
		}
		defer tlsConn.Close()
	}

	// Return latency in milliseconds
	return time.Since(start).Milliseconds(), nil
}

// withRetry executes a function with retry logic
func (m *Manager) withRetry(operation func() error) error {
	var lastErr error
	delay := m.retryConfig.RetryDelay

	for i := 0; i < m.retryConfig.MaxRetries; i++ {
		if err := operation(); err != nil {
			lastErr = err
			m.logger.WithError(err).Warnf("Operation failed (attempt %d/%d), retrying...", i+1, m.retryConfig.MaxRetries)
			time.Sleep(delay)
			delay *= 2
			if delay > m.retryConfig.MaxInterval {
				delay = m.retryConfig.MaxInterval
			}
			continue
		}
		return nil
	}

	return fmt.Errorf("operation failed after %d retries: %v", m.retryConfig.MaxRetries, lastErr)
}

// Start starts the proxy service with retry
func (m *Manager) Start() error {
	return m.withRetry(func() error {
		m.mu.Lock()
		defer m.mu.Unlock()

		if m.isRunning {
			return nil
		}

		m.logger.Info("Starting sing-box...")
		m.cmd = exec.Command("./bin/sing-box", "run", "-c", m.configPath)
		m.logger.Infof("Command: %s", m.cmd.String())

		// Set up pipes for stdout and stderr
		stdout, err := m.cmd.StdoutPipe()
		if err != nil {
			m.logger.WithError(err).Error("Failed to create stdout pipe")
			return err
		}
		stderr, err := m.cmd.StderrPipe()
		if err != nil {
			m.logger.WithError(err).Error("Failed to create stderr pipe")
			return err
		}

		m.logger.Info("Starting process...")
		// Start the process
		if err := m.cmd.Start(); err != nil {
			m.logger.WithError(err).Error("Failed to start sing-box")
			return err
		}
		m.logger.Info("Process started successfully")

		// Monitor stdout and stderr in background
		go func() {
			scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
			for scanner.Scan() {
				m.logger.Info("sing-box: " + scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				m.logger.WithError(err).Error("Error reading sing-box output")
			}
		}()

		m.isRunning = true
		m.startTime = time.Now()
		m.logger.Info("sing-box started successfully")

		go func() {
			if err := m.cmd.Wait(); err != nil {
				m.logger.WithError(err).Error("sing-box process exited with error")
			}
			m.mu.Lock()
			m.isRunning = false
			m.mu.Unlock()
			m.logger.Info("sing-box process stopped")
		}()

		return nil
	})
}

// Stop stops the proxy service with retry
func (m *Manager) Stop() error {
	return m.withRetry(func() error {
		m.mu.Lock()
		defer m.mu.Unlock()

		if !m.isRunning {
			return nil
		}

		// Try to kill the process we started
		if m.cmd != nil && m.cmd.Process != nil {
			if err := m.cmd.Process.Kill(); err != nil {
				m.logger.WithError(err).Error("Failed to stop sing-box")
				// If killing the process fails, try pkill as a fallback
				if err := exec.Command("pkill", "-f", "sing-box").Run(); err != nil {
					m.logger.WithError(err).Error("Failed to kill sing-box process")
					return fmt.Errorf("failed to kill sing-box process: %v", err)
				}
			}
		} else {
			// If we don't have a process handle, try pkill
			if err := exec.Command("pkill", "-f", "sing-box").Run(); err != nil {
				m.logger.WithError(err).Error("Failed to kill sing-box process")
				return fmt.Errorf("failed to kill sing-box process: %v", err)
			}
		}

		m.isRunning = false
		m.cmd = nil
		m.logger.Info("sing-box stopped")
		return nil
	})
}

// IsRunning returns whether the proxy service is running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}

// GetVersion returns the version of the proxy service
func (m *Manager) GetVersion() string {
	if m.version == "" {
		out, err := exec.Command("sh", "-c", "LANG=C ./bin/sing-box version").Output()
		if err != nil {
			m.logger.WithError(err).Error("Failed to get sing-box version")
			return "unknown"
		}
		m.version = string(out)
	}
	return m.version
}

// GetUptime returns the uptime of the proxy service in seconds
func (m *Manager) GetUptime() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isRunning {
		return 0
	}
	return int64(time.Since(m.startTime).Seconds())
}

// GetConnectionsCount returns the current number of connections
func (m *Manager) GetConnectionsCount() int {
	// Get connections count from sing-box API
	// TODO: Implement actual connections count collection from sing-box
	// For now, return mock data
	return 10
}

// GetNodes returns all proxy nodes
func (m *Manager) GetNodes() []*models.Node {
	nodes := make([]*models.Node, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNodesWithPagination returns paginated nodes and total count
func (m *Manager) GetNodesWithPagination(page, pageSize int) ([]*models.Node, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	nodes, total, err := m.storage.GetNodesWithPagination(page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get nodes: %v", err)
	}

	// Convert to pointer slice
	nodePointers := make([]*models.Node, len(nodes))
	for i := range nodes {
		nodePointers[i] = &nodes[i]

		// 如果数据库中没有状态数据，则从 monitor.healthData 中获取
		if nodePointers[i].Status == "" {
			if health, ok := m.monitor.healthData[nodes[i].ID]; ok {
				m.monitor.mu.RLock()
				nodePointers[i].Status = health.Status
				nodePointers[i].Latency = health.Latency
				nodePointers[i].CheckedAt = health.LastCheck
				m.monitor.mu.RUnlock()
			} else {
				nodePointers[i].Status = "offline"
				nodePointers[i].Latency = 0
				nodePointers[i].CheckedAt = time.Now()
			}
		}
	}

	return nodePointers, total, nil
}

// CreateNode creates a new proxy node
func (m *Manager) CreateNode(node *models.Node) error {
	if node.ID == "" {
		node.ID = uuid.New().String()
	}
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()
	m.nodes[node.ID] = node
	return nil
}

// UpdateNode updates an existing proxy node
func (m *Manager) UpdateNode(node *models.Node) error {
	node.UpdatedAt = time.Now()
	m.nodes[node.ID] = node
	return nil
}

// DeleteNode deletes a proxy node
func (m *Manager) DeleteNode(id string) error {
	delete(m.nodes, id)
	return nil
}

// updateConfig updates the sing-box configuration file
func (m *Manager) updateConfig() error {
	// Add to batch update queue
	return m.batchUpdate.Add(func() error {
		// Get nodes from database
		nodes, err := m.storage.GetNodes()
		if err != nil {
			return fmt.Errorf("failed to get nodes: %v", err)
		}

		// Get node groups from database
		groups, err := m.storage.GetNodeGroups()
		if err != nil {
			return fmt.Errorf("failed to get node groups: %v", err)
		}

		// Create outbounds configuration
		outbounds := []map[string]interface{}{
			{
				"type": "direct",
				"tag":  "direct",
			},
			{
				"type": "block",
				"tag":  "block",
			},
			{
				"type": "dns",
				"tag":  "dns-out",
			},
		}

		// Create node outbounds
		nodeOutbounds := make(map[string]map[string]interface{})
		for _, node := range nodes {
			outbound := m.createOutboundConfig(node)
			nodeOutbounds[node.ID] = outbound
			outbounds = append(outbounds, outbound)
		}

		// Create group outbounds
		for _, group := range groups {
			if !group.Active {
				continue
			}

			// Get nodes in this group
			groupNodes := make([]string, 0)
			for _, node := range nodes {
				for _, nodeGroup := range node.NodeGroups {
					if nodeGroup.ID == group.ID {
						groupNodes = append(groupNodes, node.ID)
						break
					}
				}
			}

			if len(groupNodes) > 0 {
				// Create selector outbound for group
				groupOutbound := map[string]interface{}{
					"type":      "selector",
					"tag":       group.Name,
					"outbounds": groupNodes,
				}
				outbounds = append(outbounds, groupOutbound)
			}
		}

		// Get rules from database
		rules, err := m.storage.GetRules()
		if err != nil {
			return fmt.Errorf("failed to get rules: %v", err)
		}

		// Create rules configuration
		rulesConfig := m.createRulesConfig(rules)

		// Get settings from database
		settings, err := m.storage.GetSettings()
		if err != nil {
			return fmt.Errorf("failed to get settings: %v", err)
		}

		// Create full configuration
		config := map[string]interface{}{
			"log": map[string]interface{}{
				"level":     settings.LogLevel,
				"timestamp": true,
			},
			"dns": map[string]interface{}{
				"servers": []map[string]interface{}{
					{
						"tag":     "local",
						"address": "223.5.5.5",
						"detour":  "direct",
					},
					{
						"tag":     "remote",
						"address": "8.8.8.8",
						"detour":  "proxy",
					},
				},
				"rules": []map[string]interface{}{
					{
						"domain":        []string{"cn", "localhost"},
						"server":        "local",
						"disable_cache": true,
					},
				},
				"strategy": "ipv4_only",
			},
			"inbounds": []map[string]interface{}{
				{
					"type":                       "mixed",
					"tag":                        "mixed-in",
					"listen":                     "::",
					"listen_port":                settings.ProxyPort,
					"sniff":                      true,
					"sniff_override_destination": true,
					"domain_strategy":            "ipv4_only",
				},
			},
			"outbounds": outbounds,
			"route": map[string]interface{}{
				"rules": append([]map[string]interface{}{
					{
						"protocol": "dns",
						"outbound": "dns-out",
					},
					{
						"geoip":    []string{"private"},
						"outbound": "direct",
					},
					{
						"geoip":    []string{"cn"},
						"outbound": "direct",
					},
					{
						"geosite":  []string{"cn"},
						"outbound": "direct",
					},
				}, rulesConfig...),
				"auto_detect_interface": true,
				"final":                 "proxy",
			},
			"experimental": map[string]interface{}{
				"cache_file": map[string]interface{}{
					"enabled": true,
					"path":    "cache.db",
				},
				"clash_api": map[string]interface{}{
					"external_controller": fmt.Sprintf("127.0.0.1:%d", settings.APIPort),
					"external_ui":         "ui",
					"secret":              "",
				},
			},
		}

		// Convert to JSON
		configJSON, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}

		// Write to file
		if err := os.WriteFile(m.configPath, configJSON, 0644); err != nil {
			return fmt.Errorf("failed to write config: %v", err)
		}

		// Create backup
		if err := os.WriteFile(m.backupPath, configJSON, 0644); err != nil {
			m.logger.WithError(err).Warn("Failed to create config backup")
		}

		// Restart service if running
		if m.isRunning {
			if err := m.Stop(); err != nil {
				return fmt.Errorf("failed to stop service: %v", err)
			}
			time.Sleep(time.Second)
			if err := m.Start(); err != nil {
				return fmt.Errorf("failed to restart service: %v", err)
			}
		}

		return nil
	})
}

// createOutboundConfig creates outbound configuration for a node
func (m *Manager) createOutboundConfig(node models.Node) map[string]interface{} {
	outbound := map[string]interface{}{
		"type": node.Type,
		"tag":  node.ID,
	}

	// Add protocol-specific settings
	switch node.Type {
	case "shadowsocks":
		outbound["server"] = node.Address
		outbound["server_port"] = node.Port
		outbound["method"] = node.Method
		outbound["password"] = node.Password
	case "vmess":
		outbound["server"] = node.Address
		outbound["server_port"] = node.Port
		outbound["uuid"] = node.UUID
		outbound["alter_id"] = node.AlterID
		if node.Network != "" {
			outbound["network"] = node.Network
		}
		if node.Path != "" {
			outbound["path"] = node.Path
		}
		if node.Host != "" {
			outbound["host"] = node.Host
		}
		if node.TLS {
			outbound["tls"] = map[string]interface{}{
				"enabled": true,
			}
		}
	case "trojan":
		outbound["server"] = node.Address
		outbound["server_port"] = node.Port
		outbound["password"] = node.Password
		outbound["tls"] = map[string]interface{}{
			"enabled":     true,
			"server_name": node.Host,
		}
	}

	return outbound
}

// createRulesConfig creates rules configuration
func (m *Manager) createRulesConfig(rules []models.Rule) []map[string]interface{} {
	rulesConfig := make([]map[string]interface{}, 0, len(rules))
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		ruleConfig := map[string]interface{}{
			"type":     rule.Type,
			"outbound": rule.Outbound,
		}

		switch rule.Type {
		case "domain":
			ruleConfig["domain"] = rule.Domains
		case "ip":
			ruleConfig["ip"] = rule.IPs
		}

		rulesConfig = append(rulesConfig, ruleConfig)
	}
	return rulesConfig
}

// Cache methods
func (c *Cache) Get() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if time.Since(c.lastModified) > c.ttl {
		return nil
	}
	return c.configCache
}

func (c *Cache) Set(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.configCache = make([]byte, len(data))
	copy(c.configCache, data)
	c.lastModified = time.Now()
}

// BatchUpdate methods
func (b *BatchUpdate) Add(change func() error) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Add change to queue
	b.changes = append(b.changes, change)
	if len(b.changes) > b.maxBatch {
		return fmt.Errorf("batch update queue is full")
	}

	// Start timer if not already running
	if !b.pending {
		b.pending = true
		b.timer = time.AfterFunc(b.interval, func() {
			b.flush()
		})
	}

	return nil
}

func (b *BatchUpdate) flush() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Execute all changes
	for _, change := range b.changes {
		if err := change(); err != nil {
			// Log error but continue with other changes
			logrus.WithError(err).Error("Failed to execute batch update change")
		}
	}

	// Clear changes
	b.changes = nil
	b.pending = false
}

// GetRules returns all routing rules
func (m *Manager) GetRules() []*models.Rule {
	rules := make([]*models.Rule, 0, len(m.rules))
	for _, rule := range m.rules {
		rules = append(rules, rule)
	}
	return rules
}

// CreateRule creates a new routing rule
func (m *Manager) CreateRule(rule *models.Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

// UpdateRule updates an existing routing rule
func (m *Manager) UpdateRule(rule *models.Rule) error {
	m.rules[rule.ID] = rule
	return nil
}

// DeleteRule deletes a routing rule
func (m *Manager) DeleteRule(id uint) error {
	delete(m.rules, strconv.FormatUint(uint64(id), 10))
	return nil
}

// GetSubscriptions returns all subscriptions
func (m *Manager) GetSubscriptions() []*models.Subscription {
	subs := make([]*models.Subscription, 0, len(m.subs))
	for _, sub := range m.subs {
		subs = append(subs, sub)
	}
	return subs
}

// CreateSubscription creates a new subscription
func (m *Manager) CreateSubscription(sub *models.Subscription) error {
	m.subs[sub.ID] = sub
	return nil
}

// UpdateSubscription updates an existing subscription
func (m *Manager) UpdateSubscription(sub *models.Subscription) error {
	m.subs[sub.ID] = sub
	return nil
}

// DeleteSubscription deletes a subscription and all its nodes
func (m *Manager) DeleteSubscription(id string) error {
	// Delete the subscription and its nodes
	if err := m.storage.DeleteSubscription(id); err != nil {
		return fmt.Errorf("failed to delete subscription: %v", err)
	}

	// Update configuration after deletion
	return m.updateConfig()
}

// GetNodeGroups returns all node groups
func (m *Manager) GetNodeGroups() []models.NodeGroup {
	groups, err := m.storage.GetNodeGroups()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get node groups")
		return nil
	}
	return groups
}

// CreateNodeGroup creates a new node group
func (m *Manager) CreateNodeGroup(group *models.NodeGroup) error {
	// Validate group
	if group.Name == "" {
		return fmt.Errorf("group name is required")
	}

	// Generate ID if not provided
	if group.ID == "" {
		group.ID = uuid.New().String()
	}

	// Save to database
	if err := m.storage.SaveNodeGroup(group); err != nil {
		return fmt.Errorf("failed to save node group: %v", err)
	}

	// Update configuration
	return m.updateConfig()
}

// UpdateNodeGroup updates an existing node group
func (m *Manager) UpdateNodeGroup(group *models.NodeGroup) error {
	// Validate group
	if group.Name == "" {
		return fmt.Errorf("group name is required")
	}

	// Save to database
	if err := m.storage.SaveNodeGroup(group); err != nil {
		return fmt.Errorf("failed to update node group: %v", err)
	}

	// Update configuration
	return m.updateConfig()
}

// DeleteNodeGroup deletes a node group
func (m *Manager) DeleteNodeGroup(id string) error {
	if err := m.storage.DeleteNodeGroup(id); err != nil {
		return fmt.Errorf("failed to delete node group: %v", err)
	}

	// Update configuration
	return m.updateConfig()
}

// TestNode tests a node's connection
func (m *Manager) TestNode(id string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find node
	node, exists := m.nodes[id]
	if !exists {
		return 0, fmt.Errorf("node not found")
	}

	// Test connection
	return m.pingNode(node)
}

// RefreshSubscription refreshes a subscription immediately
func (m *Manager) RefreshSubscription(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find subscription
	sub, exists := m.subs[id]
	if !exists {
		return fmt.Errorf("subscription not found")
	}

	// Update subscription nodes
	return m.UpdateSubscriptionNodes(sub)
}

// GetNodeHealth returns the health status of a node
func (m *Manager) GetNodeHealth(id string) *models.HealthData {
	m.monitor.mu.RLock()
	defer m.monitor.mu.RUnlock()

	if health, ok := m.monitor.healthData[id]; ok {
		return &models.HealthData{
			Status:    health.Status,
			Latency:   health.Latency,
			LastCheck: health.LastCheck,
			Error:     health.Error,
		}
	}

	// 如果没有健康数据，返回离线状态
	return &models.HealthData{
		Status:    "offline",
		Latency:   0,
		LastCheck: time.Now(),
	}
}

// GetServiceHealth returns the health status of the service
func (m *Manager) GetServiceHealth() *models.HealthData {
	m.monitor.mu.RLock()
	defer m.monitor.mu.RUnlock()

	if health, ok := m.monitor.healthData["sing-box"]; ok {
		return &health
	}
	return nil
}

// GetTrafficStats returns traffic statistics
func (m *Manager) GetTrafficStats() *models.TrafficStats {
	stats := &models.TrafficStats{
		NodeStats: make(map[string]models.Stats),
	}

	// Get traffic stats from sing-box API
	// TODO: Implement actual traffic stats collection from sing-box
	// For now, return mock data
	stats.Upload = 1024 * 1024 * 100   // 100 MB
	stats.Download = 1024 * 1024 * 200 // 200 MB

	// Mock node stats
	for _, node := range m.GetNodes() {
		stats.NodeStats[node.ID] = models.Stats{
			Upload:   1024 * 1024 * 10, // 10 MB
			Download: 1024 * 1024 * 20, // 20 MB
			Latency:  100,              // 100ms
		}
	}

	return stats
}

// GetRealtimeTraffic returns realtime traffic data
func (m *Manager) GetRealtimeTraffic() *models.TrafficStats {
	stats := &models.TrafficStats{
		NodeStats: make(map[string]models.Stats),
	}

	// Get realtime traffic from sing-box API
	// TODO: Implement actual realtime traffic collection from sing-box
	// For now, return mock data
	stats.Upload = 1024 * 100   // 100 KB/s
	stats.Download = 1024 * 200 // 200 KB/s

	// Mock node stats
	for _, node := range m.GetNodes() {
		stats.NodeStats[node.ID] = models.Stats{
			Upload:   1024 * 10, // 10 KB/s
			Download: 1024 * 20, // 20 KB/s
			Latency:  100,       // 100ms
		}
	}

	return stats
}

// UpdateSubscriptionNodes updates nodes from a subscription
func (m *Manager) UpdateSubscriptionNodes(sub *models.Subscription) error {
	// Download subscription content
	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Get(sub.URL)
	if err != nil {
		return fmt.Errorf("failed to download subscription: %v", err)
	}
	defer resp.Body.Close()

	// Read and decode content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read subscription content: %v", err)
	}

	// Try base64 decode first
	decoded, err := base64.StdEncoding.DecodeString(string(content))
	if err == nil {
		content = decoded
	}

	// Parse nodes
	var nodes []models.Node
	if err := json.Unmarshal(content, &nodes); err != nil {
		return fmt.Errorf("failed to parse subscription content: %v", err)
	}

	// Update subscription info
	sub.LastUpdate = time.Now()
	if err := m.storage.SaveSubscription(sub); err != nil {
		return fmt.Errorf("failed to update subscription: %v", err)
	}

	// Delete old nodes from this subscription
	oldNodes, err := m.storage.GetNodes()
	if err != nil {
		return fmt.Errorf("failed to get nodes: %v", err)
	}
	for _, node := range oldNodes {
		if node.SubscriptionID == sub.ID {
			if err := m.storage.DeleteNode(node.ID); err != nil {
				m.logger.WithError(err).Error("Failed to delete old node")
			}
		}
	}

	// Add new nodes
	for _, node := range nodes {
		// Decode node name from URL encoding
		if decodedName, err := url.QueryUnescape(node.Name); err == nil {
			node.Name = decodedName
		}

		// Limit name length to 30 characters
		if len(node.Name) > 30 {
			node.Name = node.Name[:27] + "..."
		}

		// Set subscription ID for the node
		node.ID = uuid.New().String()
		node.SubscriptionID = sub.ID
		node.CreatedAt = time.Now()
		node.UpdatedAt = time.Now()

		// Test node latency
		latency, err := m.pingNode(&node)
		if err != nil {
			m.logger.WithError(err).Warnf("Failed to test node latency: %s", node.Name)
		} else {
			// Update health data with latency
			m.monitor.mu.Lock()
			m.monitor.healthData[node.ID] = models.HealthData{
				Status:    "healthy",
				Latency:   latency,
				LastCheck: time.Now(),
			}
			m.monitor.mu.Unlock()
		}

		if err := m.storage.SaveNode(&node); err != nil {
			m.logger.WithError(err).Error("Failed to save node")
		}
	}

	return m.updateConfig()
}

// GetStartTime returns the start time of the service
func (m *Manager) GetStartTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.isRunning {
		return time.Time{} // Return zero time if service is not running
	}
	return m.startTime
}

// GetConfigPath returns the config path of the proxy manager
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// GetHealthData returns the health data of the proxy manager
func (m *Manager) GetHealthData() map[string]models.HealthData {
	m.monitor.mu.RLock()
	defer m.monitor.mu.RUnlock()
	return m.monitor.healthData
}

// Restart restarts the proxy manager
func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {
		return err
	}
	return m.Start()
}

// GetMosdnsVersion returns the version of mosdns
func (m *Manager) GetMosdnsVersion() string {
	if m.mosdnsVersion != "" {
		return m.mosdnsVersion
	}

	// Run with LANG=C to ensure English output
	cmd := exec.Command("sh", "-c", "LANG=C ./bin/mosdns version")
	out, err := cmd.Output()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get mosdns version")
		return "unknown"
	}

	// Clean the output and get the first line
	version := strings.TrimSpace(string(out))
	if version == "" {
		return "unknown"
	}

	// Store the version
	m.mosdnsVersion = version
	return version
}

// GetMosdnsUptime returns the uptime of mosdns in seconds
func (m *Manager) GetMosdnsUptime() int64 {
	if !m.IsMosdnsRunning() {
		return 0
	}
	return int64(time.Since(m.mosdnsTime).Seconds())
}

// IsMosdnsRunning returns whether mosdns is running
func (m *Manager) IsMosdnsRunning() bool {
	// First check if we have a process handle
	if m.mosdnsCmd == nil || m.mosdnsCmd.Process == nil {
		return false
	}

	// Then check if the process is actually running using ps
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", m.mosdnsCmd.Process.Pid))
	if err := cmd.Run(); err != nil {
		// Process not found
		m.mosdnsCmd = nil
		return false
	}

	return true
}

// IsSingboxRunning returns whether singbox is running
func (m *Manager) IsSingboxRunning() bool {
	// First check if we have a process handle
	if m.singboxCmd == nil || m.singboxCmd.Process == nil {
		return false
	}

	// Then check if the process is actually running using ps
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", m.singboxCmd.Process.Pid))
	if err := cmd.Run(); err != nil {
		// Process not found
		m.singboxCmd = nil
		return false
	}

	return true
}

// GetSingboxVersion returns the version of singbox
func (m *Manager) GetSingboxVersion() string {
	if m.singboxVersion != "" {
		return m.singboxVersion
	}

	// Run with LANG=C to ensure English output
	cmd := exec.Command("sh", "-c", "LANG=C bin/sing-box version")
	out, err := cmd.Output()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get sing-box version")
		return "unknown"
	}

	// Clean the output and get the first line
	version := strings.TrimSpace(string(out))
	if version == "" {
		return "unknown"
	}

	// Store the version
	m.singboxVersion = version
	return version
}

// GetSingboxUptime returns the uptime of singbox in seconds
func (m *Manager) GetSingboxUptime() int64 {
	if !m.IsSingboxRunning() {
		return 0
	}
	return int64(time.Since(m.singboxTime).Seconds())
}

// StartSingbox starts the sing-box service
func (m *Manager) StartSingbox() error {
	m.logger.Debug("Checking if sing-box is already running...")
	if m.singboxCmd != nil && m.singboxCmd.Process != nil {
		m.logger.Debug("sing-box is already running")
		return nil
	}

	m.logger.Debug("Creating sing-box command...")
	cmd := exec.Command("./bin/sing-box", "run", "-c", m.configPath)

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.logger.WithError(err).Error("Failed to create stdout pipe")
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		m.logger.WithError(err).Error("Failed to create stderr pipe")
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Start reading from stdout and stderr in goroutines
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m.logger.Debug("sing-box stdout: " + scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			m.logger.Debug("sing-box stderr: " + scanner.Text())
		}
	}()

	m.logger.Debug("Starting sing-box process...")
	if err := cmd.Start(); err != nil {
		m.logger.WithError(err).Error("Failed to start sing-box process")
		return fmt.Errorf("failed to start sing-box: %v", err)
	}

	m.singboxCmd = cmd
	m.singboxTime = time.Now() // 设置启动时间
	m.logger.Debug("sing-box process started successfully")

	// Monitor process status in a goroutine
	go func() {
		err := cmd.Wait()
		if err != nil {
			m.logger.WithError(err).Error("sing-box process exited with error")
		} else {
			m.logger.Info("sing-box process exited normally")
		}
		m.singboxCmd = nil
		m.singboxTime = time.Time{} // 进程退出时重置时间
	}()

	return nil
}

// StartMosdns starts the mosdns service
func (m *Manager) StartMosdns() error {
	m.logger.Debug("Checking if mosdns is already running...")
	if m.mosdnsCmd != nil && m.mosdnsCmd.Process != nil {
		m.logger.Debug("mosdns is already running")
		return nil
	}

	// Create logs directory
	if err := os.MkdirAll("logs", 0755); err != nil {
		m.logger.WithError(err).Error("Failed to create logs directory")
		return fmt.Errorf("failed to create logs directory: %v", err)
	}

	// Create mosdns config directory
	if err := os.MkdirAll("configs/mosdns", 0755); err != nil {
		m.logger.WithError(err).Error("Failed to create mosdns config directory")
		return fmt.Errorf("failed to create mosdns config directory: %v", err)
	}

	m.logger.Debug("Creating mosdns command...")
	cmd := exec.Command("./bin/mosdns", "start", "-d", "configs/mosdns")

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.logger.WithError(err).Error("Failed to create stdout pipe")
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		m.logger.WithError(err).Error("Failed to create stderr pipe")
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Start reading from stdout and stderr in goroutines
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			m.logger.Debug("mosdns stdout: " + scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			m.logger.Debug("mosdns stderr: " + scanner.Text())
		}
	}()

	m.logger.Debug("Starting mosdns process...")
	if err := cmd.Start(); err != nil {
		m.logger.WithError(err).Error("Failed to start mosdns process")
		return fmt.Errorf("failed to start mosdns: %v", err)
	}

	m.mosdnsCmd = cmd
	m.mosdnsTime = time.Now() // 设置启动时间
	m.logger.Debug("mosdns process started successfully")

	// Monitor process status in a goroutine
	go func() {
		err := cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				m.logger.WithError(err).Errorf("mosdns process exited with status: %d", exitErr.ExitCode())
			} else {
				m.logger.WithError(err).Error("mosdns process exited with error")
			}
		} else {
			m.logger.Info("mosdns process exited normally")
		}
		m.mosdnsCmd = nil
		m.mosdnsTime = time.Time{} // 进程退出时重置时间
	}()

	// Wait a bit to ensure process started successfully
	time.Sleep(time.Second)

	// Check if process is still running
	if m.mosdnsCmd == nil || m.mosdnsCmd.Process == nil {
		return fmt.Errorf("mosdns failed to start")
	}

	return nil
}

// SaveNode saves a node to the storage
func (m *Manager) SaveNode(node *models.Node) error {
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()
	return m.storage.SaveNode(node)
}

// SaveSubscription saves a subscription to the storage
func (m *Manager) SaveSubscription(sub *models.Subscription) error {
	sub.LastUpdate = time.Now()
	return m.storage.SaveSubscription(sub)
}

// DeleteNodesBySubscriptionID deletes all nodes belonging to a subscription
func (m *Manager) DeleteNodesBySubscriptionID(subscriptionID string) error {
	oldNodes, err := m.storage.GetNodes()
	if err != nil {
		return fmt.Errorf("failed to get nodes: %v", err)
	}
	for _, node := range oldNodes {
		if node.SubscriptionID == subscriptionID {
			if err := m.storage.DeleteNode(node.ID); err != nil {
				m.logger.WithError(err).Error("Failed to delete old node")
			}
		}
	}
	return nil
}

// UpdateNodeHealth updates the health status of a node
func (m *Manager) UpdateNodeHealth(nodeID string, status string, latency int64, checkedAt time.Time, err string) {
	m.monitor.mu.Lock()
	defer m.monitor.mu.Unlock()

	m.monitor.healthData[nodeID] = models.HealthData{
		Status:    status,
		Latency:   latency,
		LastCheck: checkedAt,
		Error:     err,
	}
}

// StartService starts a system service
func (m *Manager) StartService(name string) error {
	m.logger.Debugf("Starting service: %s", name)
	switch name {
	case "singbox":
		m.logger.Debug("Starting sing-box service...")
		if err := m.StartSingbox(); err != nil {
			m.logger.WithError(err).Error("Failed to start sing-box service")
			return fmt.Errorf("failed to start sing-box: %v", err)
		}
		m.logger.Info("sing-box service started successfully")
		m.isRunning = true
		m.startTime = time.Now()
		return nil
	case "mosdns":
		m.logger.Debug("Starting mosdns service...")
		if err := m.StartMosdns(); err != nil {
			m.logger.WithError(err).Error("Failed to start mosdns service")
			return fmt.Errorf("failed to start mosdns: %v", err)
		}
		m.logger.Info("mosdns service started successfully")
		m.isRunning = true
		m.startTime = time.Now()
		return nil
	default:
		return fmt.Errorf("unknown service: %s", name)
	}
}

// StopService stops a system service
func (m *Manager) StopService(name string) error {
	switch name {
	case "singbox":
		// First try to stop the process we started
		if m.singboxCmd != nil && m.singboxCmd.Process != nil {
			if err := m.singboxCmd.Process.Kill(); err != nil {
				m.logger.WithError(err).Error("Failed to stop sing-box")
				// If killing the process fails, try pkill as a fallback
				if err := exec.Command("pkill", "-f", "sing-box").Run(); err != nil {
					m.logger.WithError(err).Error("Failed to kill sing-box processes")
					return fmt.Errorf("failed to kill sing-box process: %v", err)
				}
			}
		} else {
			// If we don't have a process handle, try pkill
			if err := exec.Command("pkill", "-f", "sing-box").Run(); err != nil {
				m.logger.WithError(err).Error("Failed to kill sing-box processes")
				return fmt.Errorf("failed to kill sing-box process: %v", err)
			}
		}
		m.singboxCmd = nil
		m.singboxTime = time.Time{} // 重置启动时间
		return nil
	case "mosdns":
		if m.mosdnsCmd != nil && m.mosdnsCmd.Process != nil {
			if err := m.mosdnsCmd.Process.Kill(); err != nil {
				m.logger.WithError(err).Error("Failed to stop mosdns")
				// If killing the process fails, try pkill as a fallback
				if err := exec.Command("pkill", "-f", "mosdns").Run(); err != nil {
					m.logger.WithError(err).Error("Failed to kill mosdns processes")
					return fmt.Errorf("failed to kill mosdns process: %v", err)
				}
			}
		} else {
			// If we don't have a process handle, try pkill
			if err := exec.Command("pkill", "-f", "mosdns").Run(); err != nil {
				m.logger.WithError(err).Error("Failed to kill mosdns processes")
				return fmt.Errorf("failed to kill mosdns process: %v", err)
			}
		}
		m.mosdnsCmd = nil
		m.mosdnsTime = time.Time{} // 重置启动时间
		return nil
	default:
		return fmt.Errorf("unknown service: %s", name)
	}
}

// RestartService restarts a system service
func (m *Manager) RestartService(name string) error {
	switch name {
	case "singbox":
		// Stop the service
		if err := m.StopService("singbox"); err != nil {
			m.logger.WithError(err).Error("Failed to stop sing-box during restart")
			// Try to kill any remaining processes
			if err := exec.Command("pkill", "-f", "sing-box").Run(); err != nil {
				m.logger.WithError(err).Error("Failed to kill sing-box processes during restart")
			}
		}
		// Wait for the service to stop
		time.Sleep(time.Second)
		// Start the service
		return m.StartSingbox()
	case "mosdns":
		// Stop the service
		if err := m.StopService("mosdns"); err != nil {
			m.logger.WithError(err).Error("Failed to stop mosdns during restart")
			// Try to kill any remaining processes
			if err := exec.Command("pkill", "-f", "mosdns").Run(); err != nil {
				m.logger.WithError(err).Error("Failed to kill mosdns processes during restart")
			}
		}
		// Wait for the service to stop
		time.Sleep(time.Second)
		// Start the service
		return m.StartMosdns()
	default:
		return fmt.Errorf("unknown service: %s", name)
	}
}

// initDefaultNodeGroups initializes default node groups by country
func (m *Manager) initDefaultNodeGroups() error {
	defaultGroups := []struct {
		name            string
		tag             string
		includePatterns []string
		description     string
	}{
		{
			name:            "香港节点",
			tag:             "hk",
			includePatterns: []string{"香港", "HK", "Hong Kong", "HongKong"},
			description:     "香港地区的节点",
		},
		{
			name:            "台湾节点",
			tag:             "tw",
			includePatterns: []string{"台湾", "TW", "Taiwan"},
			description:     "台湾地区的节点",
		},
		{
			name:            "日本节点",
			tag:             "jp",
			includePatterns: []string{"日本", "JP", "Japan"},
			description:     "日本地区的节点",
		},
		{
			name:            "韩国节点",
			tag:             "kr",
			includePatterns: []string{"韩国", "KR", "Korea"},
			description:     "韩国地区的节点",
		},
		{
			name:            "新加坡节点",
			tag:             "sg",
			includePatterns: []string{"新加坡", "SG", "Singapore"},
			description:     "新加坡地区的节点",
		},
		{
			name:            "美国节点",
			tag:             "us",
			includePatterns: []string{"美国", "US", "United States", "USA"},
			description:     "美国地区的节点",
		},
	}

	// Get existing groups
	existingGroups, err := m.storage.GetNodeGroups()
	if err != nil {
		return fmt.Errorf("failed to get existing groups: %v", err)
	}

	// Create map for quick lookup
	existingGroupMap := make(map[string]bool)
	for _, group := range existingGroups {
		existingGroupMap[group.Name] = true
	}

	// Create default groups if they don't exist
	for _, dg := range defaultGroups {
		if !existingGroupMap[dg.name] {
			group := &models.NodeGroup{
				ID:              uuid.New().String(),
				Name:            dg.name,
				Tag:             dg.tag,
				Description:     dg.description,
				IncludePatterns: models.StringArray(dg.includePatterns),
				Mode:            "select",
				Active:          true,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}
			if err := m.storage.SaveNodeGroup(group); err != nil {
				m.logger.WithError(err).Errorf("Failed to create default group: %s", dg.name)
				continue
			}
		}
	}

	return nil
}

// UpdateNodeGroups updates node group assignments for all nodes
func (m *Manager) UpdateNodeGroups() error {
	return m.updateNodeGroupAssignments()
}

// updateNodeGroupAssignments updates node group assignments based on patterns
func (m *Manager) updateNodeGroupAssignments() error {
	// Get all nodes
	nodes, err := m.storage.GetNodes()
	if err != nil {
		return fmt.Errorf("failed to get nodes: %v", err)
	}

	// Get all groups
	groups, err := m.storage.GetNodeGroups()
	if err != nil {
		return fmt.Errorf("failed to get groups: %v", err)
	}

	// Find the "others" group and normal groups
	var othersGroup *models.NodeGroup
	var normalGroups []models.NodeGroup
	for i := range groups {
		if groups[i].Tag == "others" {
			othersGroup = &groups[i]
		} else {
			normalGroups = append(normalGroups, groups[i])
		}
	}

	// Start a database transaction
	tx := m.storage.GetDB().Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Step 1: Clear all node group associations
	if err := tx.Exec("DELETE FROM node_group_nodes").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to clear node group associations: %v", err)
	}

	// 创建一个 map 来跟踪已分配的节点
	assignedNodes := make(map[string]bool)

	// Step 2: Process normal groups in priority order
	for i := range normalGroups {
		if !normalGroups[i].Active {
			continue
		}

		matchedNodes := make([]models.Node, 0)
		for _, node := range nodes {
			// 跳过已分配的节点
			if assignedNodes[node.ID] {
				continue
			}

			// 检查节点是否匹配当前组的规则
			if m.nodeMatchesGroup(&node, &normalGroups[i]) {
				matchedNodes = append(matchedNodes, node)
				assignedNodes[node.ID] = true
				// Add node group association
				if err := tx.Exec("INSERT INTO node_group_nodes (node_group_id, node_id) VALUES (?, ?)",
					normalGroups[i].ID, node.ID).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to add node group association: %v", err)
				}
				m.logger.Debugf("Node %s assigned to group %s", node.Name, normalGroups[i].Name)
			}
		}

		// Update the node count for the group
		normalGroups[i].NodeCount = len(matchedNodes)
		if err := tx.Model(&normalGroups[i]).Update("node_count", len(matchedNodes)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update node count: %v", err)
		}
	}

	// Step 3: Process the "others" group
	if othersGroup != nil && othersGroup.Active {
		othersNodes := make([]models.Node, 0)
		for _, node := range nodes {
			// 只处理未分配的节点
			if !assignedNodes[node.ID] {
				othersNodes = append(othersNodes, node)
				// Add node group association
				if err := tx.Exec("INSERT INTO node_group_nodes (node_group_id, node_id) VALUES (?, ?)",
					othersGroup.ID, node.ID).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to add node group association: %v", err)
				}
				m.logger.Debugf("Node %s assigned to others group", node.Name)
			}
		}

		// Update the node count for the "others" group
		if err := tx.Model(othersGroup).Update("node_count", len(othersNodes)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update others group node count: %v", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Generate and save the configuration
	return m.updateSingboxConfig()
}

// nodeMatchesGroup checks if a node matches the group's patterns
func (m *Manager) nodeMatchesGroup(node *models.Node, group *models.NodeGroup) bool {
	// 如果是其他节点组，使用特殊的匹配逻辑
	if group.Tag == "others" {
		return true // 其他节点组总是匹配未分配的节点
	}

	// 对于非"其他"组，使用正常的匹配逻辑
	// 如果组没有包含规则，则不匹配
	if len(group.IncludePatterns) == 0 {
		return false
	}

	// 首先检查包含规则
	matched := false
	for _, pattern := range group.IncludePatterns {
		if pattern == "" {
			continue
		}
		if strings.Contains(strings.ToLower(node.Name), strings.ToLower(pattern)) {
			matched = true
			m.logger.Debugf("Node %s matched include pattern %s in group %s", node.Name, pattern, group.Name)
			break
		}
	}

	// 如果不匹配包含规则，直接返回 false
	if !matched {
		return false
	}

	// 如果匹配包含规则，再检查排除规则
	if len(group.ExcludePatterns) > 0 {
		for _, pattern := range group.ExcludePatterns {
			if pattern == "" {
				continue
			}
			if strings.Contains(strings.ToLower(node.Name), strings.ToLower(pattern)) {
				m.logger.Debugf("Node %s matched exclude pattern %s in group %s", node.Name, pattern, group.Name)
				return false
			}
		}
	}

	return true
}

// UpdateConfig updates the proxy configuration
func (m *Manager) UpdateConfig() error {
	// 使用 updateNodeGroupAssignments 来处理节点分组
	if err := m.updateNodeGroupAssignments(); err != nil {
		return fmt.Errorf("failed to update node groups: %v", err)
	}

	return nil
}

// updateSingboxConfig generates and saves the sing-box configuration
func (m *Manager) updateSingboxConfig() error {
	// Get settings
	settings, err := m.storage.GetSettings()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get settings")
		return err
	}

	// Generate new sing-box config
	generator := config.NewSingBoxGenerator(settings, m.storage, m.configPath)
	data, err := generator.GenerateConfig()
	if err != nil {
		m.logger.WithError(err).Error("Failed to generate sing-box config")
		return err
	}

	// Save config file
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		m.logger.WithError(err).Error("Failed to write sing-box config")
		return err
	}

	// If sing-box is running, restart service to apply new config
	if m.isRunning {
		if err := m.RestartService("singbox"); err != nil {
			m.logger.WithError(err).Error("Failed to restart sing-box")
			return err
		}
	}

	return nil
}

// UpdateNodeGroupAssignments updates the node group assignments
func (m *Manager) UpdateNodeGroupAssignments() error {
	// 获取所有节点和节点组
	nodes, err := m.storage.GetNodes()
	if err != nil {
		return fmt.Errorf("failed to get nodes: %v", err)
	}

	groups, err := m.storage.GetNodeGroups()
	if err != nil {
		return fmt.Errorf("failed to get node groups: %v", err)
	}

	// 创建一个 map 来跟踪已分配的节点
	assignedNodes := make(map[string]bool)

	// 遍历所有节点组，分配节点
	for _, group := range groups {
		for _, node := range nodes {
			if m.nodeMatchesGroup(&node, &group) {
				if err := m.storage.SaveNodeGroupAssignment(node.ID, group.ID); err != nil {
					return fmt.Errorf("failed to assign node to group: %v", err)
				}
				assignedNodes[node.ID] = true
			}
		}
	}

	return nil
}
