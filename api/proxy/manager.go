package proxy

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"singdns/api/models"
	"singdns/api/storage"

	"crypto/tls"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Manager implements the ProxyManager interface
type Manager struct {
	logger      *logrus.Logger
	mu          sync.RWMutex
	cmd         *exec.Cmd
	startTime   time.Time
	isRunning   bool
	version     string
	configPath  string
	backupPath  string
	storage     storage.Storage
	autoSave    bool
	saveTimer   *time.Timer
	retryConfig *RetryConfig
	cache       *Cache
	batchUpdate *BatchUpdate
	crypto      *Crypto
	monitor     *Monitor
	nodes       map[string]*models.Node
	rules       map[string]*models.Rule
	subs        map[string]*models.Subscription
	singboxCmd  *exec.Cmd
	mosdnsCmd   *exec.Cmd
	singboxTime time.Time
	mosdnsTime  time.Time
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
	lastCheck  time.Time
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

// encryptPassword encrypts sensitive information
func (m *Manager) encryptPassword(password string) (string, error) {
	block, err := aes.NewCipher(m.crypto.key)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptPassword decrypts sensitive information
func (m *Manager) decryptPassword(encrypted string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(m.crypto.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// sanitizeLog removes sensitive information from log messages
func (m *Manager) sanitizeLog(message string) string {
	// Remove passwords
	message = regexp.MustCompile(`password["\s:]+[^"\s,}]+`).ReplaceAllString(message, "password: [REDACTED]")

	// Remove UUIDs
	message = regexp.MustCompile(`uuid["\s:]+[^"\s,}]+`).ReplaceAllString(message, "uuid: [REDACTED]")

	// Remove tokens
	message = regexp.MustCompile(`token["\s:]+[^"\s,}]+`).ReplaceAllString(message, "token: [REDACTED]")

	return message
}

// loadConfig loads configuration from file
func (m *Manager) loadConfig() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	var config struct {
		Nodes []models.Node         `json:"nodes"`
		Rules []models.Rule         `json:"rules"`
		Subs  []models.Subscription `json:"subscriptions"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	m.mu.Lock()
	// Convert slices to maps
	m.nodes = make(map[string]*models.Node)
	for i := range config.Nodes {
		m.nodes[config.Nodes[i].ID] = &config.Nodes[i]
	}
	m.rules = make(map[string]*models.Rule)
	for i := range config.Rules {
		m.rules[config.Rules[i].ID] = &config.Rules[i]
	}
	m.subs = make(map[string]*models.Subscription)
	for i := range config.Subs {
		m.subs[config.Subs[i].ID] = &config.Subs[i]
	}
	m.mu.Unlock()

	return nil
}

// loadBackupConfig loads configuration from backup file
func (m *Manager) loadBackupConfig() error {
	data, err := os.ReadFile(m.backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %v", err)
	}

	var config struct {
		Nodes []models.Node         `json:"nodes"`
		Rules []models.Rule         `json:"rules"`
		Subs  []models.Subscription `json:"subscriptions"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse backup file: %v", err)
	}

	m.mu.Lock()
	// Convert slices to maps
	m.nodes = make(map[string]*models.Node)
	for i := range config.Nodes {
		m.nodes[config.Nodes[i].ID] = &config.Nodes[i]
	}
	m.rules = make(map[string]*models.Rule)
	for i := range config.Rules {
		m.rules[config.Rules[i].ID] = &config.Rules[i]
	}
	m.subs = make(map[string]*models.Subscription)
	for i := range config.Subs {
		m.subs[config.Subs[i].ID] = &config.Subs[i]
	}
	m.mu.Unlock()

	return nil
}

// saveConfig saves current configuration to file
func (m *Manager) saveConfig() error {
	m.mu.RLock()
	// Convert maps to slices
	nodes := make([]models.Node, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, *node)
	}
	rules := make([]models.Rule, 0, len(m.rules))
	for _, rule := range m.rules {
		rules = append(rules, *rule)
	}
	subs := make([]models.Subscription, 0, len(m.subs))
	for _, sub := range m.subs {
		subs = append(subs, *sub)
	}

	config := struct {
		Nodes []models.Node         `json:"nodes"`
		Rules []models.Rule         `json:"rules"`
		Subs  []models.Subscription `json:"subscriptions"`
	}{
		Nodes: nodes,
		Rules: rules,
		Subs:  subs,
	}
	m.mu.RUnlock()

	// Create backup first
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(m.backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %v", err)
	}

	// Write to main config file
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// startAutoSave starts the auto-save timer
func (m *Manager) startAutoSave() {
	if !m.autoSave {
		return
	}

	m.saveTimer = time.AfterFunc(time.Minute*5, func() {
		if err := m.saveConfig(); err != nil {
			m.logger.WithError(err).Error("Failed to auto-save configuration")
		}
		m.startAutoSave() // Reschedule
	})
}

// stopAutoSave stops the auto-save timer
func (m *Manager) stopAutoSave() {
	if m.saveTimer != nil {
		m.saveTimer.Stop()
		m.saveTimer = nil
	}
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

		m.cmd = exec.Command("./bin/sing-box", "run", "-c", m.configPath)
		if err := m.cmd.Start(); err != nil {
			m.logger.WithError(err).Error("Failed to start sing-box")
			return err
		}

		m.isRunning = true
		m.startTime = time.Now()
		m.logger.Info("sing-box started")

		go func() {
			if err := m.cmd.Wait(); err != nil {
				m.logger.WithError(err).Error("sing-box process exited with error")
			}
			m.mu.Lock()
			m.isRunning = false
			m.mu.Unlock()
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

		if m.cmd != nil && m.cmd.Process != nil {
			if err := m.cmd.Process.Kill(); err != nil {
				m.logger.WithError(err).Error("Failed to stop sing-box")
				return err
			}
		}

		m.isRunning = false
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
		out, err := exec.Command("./bin/sing-box", "version").Output()
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

// validateNode validates a node's configuration
func (m *Manager) validateNode(node *models.Node) error {
	if node.Address == "" {
		return fmt.Errorf("server address is required")
	}
	if node.Port <= 0 || node.Port > 65535 {
		return fmt.Errorf("invalid port number")
	}

	switch node.Type {
	case "shadowsocks":
		if node.Method == "" {
			return fmt.Errorf("encryption method is required for shadowsocks")
		}
		if node.Password == "" {
			return fmt.Errorf("password is required for shadowsocks")
		}
	case "vmess":
		if node.UUID == "" {
			return fmt.Errorf("UUID is required for vmess")
		}
	case "trojan":
		if node.Password == "" {
			return fmt.Errorf("password is required for trojan")
		}
		if node.Host == "" {
			node.Host = node.Address
		}
	default:
		return fmt.Errorf("unsupported node type: %s", node.Type)
	}

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

		// Create outbounds configuration
		outbounds := make([]map[string]interface{}, len(nodes))
		for i, node := range nodes {
			// Decrypt password for configuration
			if node.Password != "" {
				if decrypted, err := m.decryptPassword(node.Password); err == nil {
					node.Password = decrypted
				} else {
					m.logger.WithError(err).Error("Failed to decrypt node password")
				}
			}
			outbound := m.createOutboundConfig(node)
			outbounds[i] = outbound
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
			"outbounds": append([]map[string]interface{}{
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
			}, outbounds...),
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

		// Check cache
		if data := m.cache.Get(); data != nil {
			// Compare with new config
			newData, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal config: %v", err)
			}
			if bytes.Equal(data, newData) {
				return nil // No changes
			}
		}

		// Marshal to JSON
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %v", err)
		}

		// Log sanitized configuration
		m.logger.WithField("config", m.sanitizeLog(string(data))).Debug("Updating configuration")

		// Update cache
		m.cache.Set(data)

		// Write to file
		if err := os.WriteFile(m.configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config: %v", err)
		}

		// Write backup
		if err := os.WriteFile(m.backupPath, data, 0644); err != nil {
			m.logger.WithError(err).Error("Failed to write backup config")
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

// validateRule validates a rule's configuration
func (m *Manager) validateRule(rule *models.Rule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	switch rule.Type {
	case "domain":
		if len(rule.Domains) == 0 {
			return fmt.Errorf("domains are required for domain rule")
		}
	case "ip":
		if len(rule.IPs) == 0 {
			return fmt.Errorf("IPs are required for IP rule")
		}
	default:
		return fmt.Errorf("unsupported rule type: %s", rule.Type)
	}

	// Validate outbound exists
	outboundExists := false
	for _, node := range m.GetNodes() {
		if node.ID == rule.Outbound {
			outboundExists = true
			break
		}
	}
	if !outboundExists {
		return fmt.Errorf("outbound node not found: %s", rule.Outbound)
	}

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

// DeleteSubscription deletes a subscription
func (m *Manager) DeleteSubscription(id uint) error {
	// Delete all nodes belonging to this subscription first
	nodes, err := m.storage.GetNodes()
	if err != nil {
		return fmt.Errorf("failed to get nodes: %v", err)
	}

	for _, node := range nodes {
		if node.SubscriptionID == strconv.FormatUint(uint64(id), 10) {
			if err := m.storage.DeleteNode(node.ID); err != nil {
				m.logger.WithError(err).Error("Failed to delete old node")
			}
		}
	}

	// Then delete the subscription
	if err := m.storage.DeleteSubscription(strconv.FormatUint(uint64(id), 10)); err != nil {
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

// GetStartTime returns the start time of the proxy manager
func (m *Manager) GetStartTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
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
	return "v1.0.0" // TODO: implement actual version retrieval
}

// GetMosdnsUptime returns the uptime of mosdns in seconds
func (m *Manager) GetMosdnsUptime() int64 {
	return 0 // TODO: implement actual uptime tracking
}

// IsMosdnsRunning returns whether mosdns is running
func (m *Manager) IsMosdnsRunning() bool {
	return false // TODO: implement actual status check
}

// IsSingboxRunning returns whether singbox is running
func (m *Manager) IsSingboxRunning() bool {
	return m.singboxCmd != nil && m.singboxCmd.Process != nil
}

// GetSingboxVersion returns the version of singbox
func (m *Manager) GetSingboxVersion() string {
	// TODO: Implement version detection
	return "unknown"
}

// GetSingboxUptime returns the uptime of singbox in seconds
func (m *Manager) GetSingboxUptime() int64 {
	if !m.IsSingboxRunning() {
		return 0
	}
	return int64(time.Since(m.singboxTime).Seconds())
}

// StartSingbox starts the singbox service
func (m *Manager) StartSingbox() error {
	if m.IsSingboxRunning() {
		return fmt.Errorf("singbox is already running")
	}

	m.singboxCmd = exec.Command("singbox", "run")
	m.singboxTime = time.Now()

	if err := m.singboxCmd.Start(); err != nil {
		return fmt.Errorf("failed to start singbox: %v", err)
	}

	return nil
}

// StartMosdns starts the mosdns service
func (m *Manager) StartMosdns() error {
	if m.IsMosdnsRunning() {
		return fmt.Errorf("mosdns is already running")
	}

	m.mosdnsCmd = exec.Command("mosdns", "start")
	m.mosdnsTime = time.Now()

	if err := m.mosdnsCmd.Start(); err != nil {
		return fmt.Errorf("failed to start mosdns: %v", err)
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

// Download subscription content
func (m *Manager) downloadSubscription(sub *models.Subscription) ([]byte, error) {
	resp, err := http.Get(sub.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download subscription: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription URL returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
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

// parseClashSubscription parses a Clash subscription
func (m *Manager) parseClashSubscription(data []byte) ([]*models.Node, error) {
	var config struct {
		Proxies []struct {
			Name           string            `yaml:"name"`
			Type           string            `yaml:"type"`
			Server         string            `yaml:"server"`
			Port           int               `yaml:"port"`
			Password       string            `yaml:"password"`
			UUID           string            `yaml:"uuid"`
			Cipher         string            `yaml:"cipher"`
			UDP            bool              `yaml:"udp"`
			Network        string            `yaml:"network"`
			TLS            bool              `yaml:"tls"`
			SNI            string            `yaml:"sni"`
			AlterID        int               `yaml:"alterId"`
			WS_PATH        string            `yaml:"ws-path"`
			WS_HEADERS     map[string]string `yaml:"ws-headers"`
			Plugin         string            `yaml:"plugin"`
			PluginOpts     map[string]string `yaml:"plugin-opts"`
			ALPN           []string          `yaml:"alpn"`
			SkipCertVerify bool              `yaml:"skip-cert-verify"`
		} `yaml:"proxies"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		// If YAML parsing fails, try JSON as fallback
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse Clash config: %v", err)
		}
	}

	var nodes []*models.Node
	for _, proxy := range config.Proxies {
		switch strings.ToLower(proxy.Type) {
		case "ss", "shadowsocks", "vmess", "trojan":
			// Convert PluginOpts to JSON string if it exists
			var pluginOptsStr string
			if len(proxy.PluginOpts) > 0 {
				pluginOptsBytes, err := json.Marshal(proxy.PluginOpts)
				if err != nil {
					m.logger.Warnf("Failed to marshal plugin opts for node %s: %v", proxy.Name, err)
					pluginOptsStr = "{}"
				} else {
					pluginOptsStr = string(pluginOptsBytes)
				}
			}

			node := &models.Node{
				Name:           proxy.Name,
				Type:           proxy.Type,
				Address:        proxy.Server,
				Port:           proxy.Port,
				Password:       proxy.Password,
				UUID:           proxy.UUID,
				Method:         proxy.Cipher,
				Network:        proxy.Network,
				TLS:            proxy.TLS,
				AlterID:        proxy.AlterID,
				UDP:            proxy.UDP,
				SkipCertVerify: proxy.SkipCertVerify,
				Plugin:         proxy.Plugin,
				PluginOpts:     pluginOptsStr,
			}

			// Handle WebSocket configuration
			if proxy.Network == "ws" {
				node.Path = proxy.WS_PATH
				if host, ok := proxy.WS_HEADERS["Host"]; ok {
					node.Host = host
				}
			}

			// Handle ALPN
			if len(proxy.ALPN) > 0 {
				node.ALPN = proxy.ALPN
			}

			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// parseV2raySubscription parses a V2Ray subscription
func (m *Manager) parseV2raySubscription(data []byte) ([]*models.Node, error) {
	// Try to decode Base64 if the content looks like Base64
	if !strings.Contains(string(data), "vmess://") {
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			decoded, err = base64.URLEncoding.DecodeString(string(data))
			if err != nil {
				return nil, fmt.Errorf("failed to decode subscription content: %v", err)
			}
		}
		data = decoded
	}

	lines := strings.Split(string(data), "\n")
	var nodes []*models.Node

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var node *models.Node
		var err error

		switch {
		case strings.HasPrefix(line, "vmess://"):
			node, err = m.parseVMessLink(line)
		case strings.HasPrefix(line, "ss://"):
			node, err = m.parseShadowsocksLink(line)
		case strings.HasPrefix(line, "trojan://"):
			node, err = m.parseTrojanLink(line)
		default:
			m.logger.Debugf("Skipping unsupported protocol: %s", line)
			continue
		}

		if err != nil {
			m.logger.Warnf("Failed to parse node: %v", err)
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// parseShadowsocksSubscription parses a Shadowsocks subscription
func (m *Manager) parseShadowsocksSubscription(data []byte) ([]*models.Node, error) {
	// Shadowsocks subscriptions are typically handled the same way as V2Ray subscriptions
	return m.parseV2raySubscription(data)
}

// parseVMessLink parses a VMess link
func (m *Manager) parseVMessLink(link string) (*models.Node, error) {
	// Remove "vmess://" prefix
	link = strings.TrimPrefix(link, "vmess://")

	// Decode Base64
	decoded, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(link)
		if err != nil {
			return nil, fmt.Errorf("failed to decode VMess link: %v", err)
		}
	}

	// Parse JSON
	var config struct {
		Version string `json:"v"`
		Name    string `json:"ps"`
		Address string `json:"add"`
		Port    int    `json:"port"`
		UUID    string `json:"id"`
		AlterID int    `json:"aid"`
		Network string `json:"net"`
		Type    string `json:"type"`
		Host    string `json:"host"`
		Path    string `json:"path"`
		TLS     string `json:"tls"`
	}

	if err := json.Unmarshal(decoded, &config); err != nil {
		return nil, fmt.Errorf("failed to parse VMess config: %v", err)
	}

	return &models.Node{
		Name:    config.Name,
		Type:    "vmess",
		Address: config.Address,
		Port:    config.Port,
		UUID:    config.UUID,
		AlterID: config.AlterID,
		Network: config.Network,
		Host:    config.Host,
		Path:    config.Path,
		TLS:     config.TLS == "tls",
	}, nil
}

// parseShadowsocksLink parses a Shadowsocks link
func (m *Manager) parseShadowsocksLink(link string) (*models.Node, error) {
	// Remove "ss://" prefix
	uri, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Shadowsocks URL: %v", err)
	}

	// Extract method and password
	methodAndPassword := strings.SplitN(uri.User.String(), ":", 2)
	if len(methodAndPassword) != 2 {
		return nil, fmt.Errorf("invalid Shadowsocks user info format")
	}

	// Extract port
	port, err := strconv.Atoi(uri.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v", err)
	}

	return &models.Node{
		Name:     uri.Fragment,
		Type:     "shadowsocks",
		Address:  uri.Hostname(),
		Port:     port,
		Method:   methodAndPassword[0],
		Password: methodAndPassword[1],
	}, nil
}

// parseTrojanLink parses a Trojan link
func (m *Manager) parseTrojanLink(link string) (*models.Node, error) {
	// Remove "trojan://" prefix
	uri, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Trojan URL: %v", err)
	}

	// Extract password
	password := uri.User.String()

	// Extract port
	port, err := strconv.Atoi(uri.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v", err)
	}

	// Extract name from fragment
	name := uri.Fragment

	// Parse query parameters
	query := uri.Query()
	sni := query.Get("sni")
	if sni == "" {
		sni = uri.Hostname() // Use hostname as SNI if not specified
	}

	return &models.Node{
		Name:     name,
		Type:     "trojan",
		Address:  uri.Hostname(),
		Port:     port,
		Password: password,
		Host:     sni,
		TLS:      true, // Trojan always uses TLS
	}, nil
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
