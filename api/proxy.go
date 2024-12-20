package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ProxyManager manages proxy nodes and subscriptions
type ProxyManager struct {
	nodes         map[string]*Node
	rules         map[string]*Rule
	subscriptions map[string]*Subscription
	trafficStats  *TrafficStats
	mu            sync.RWMutex
}

// NewProxyManager creates a new proxy manager
func NewProxyManager() *ProxyManager {
	return &ProxyManager{
		nodes:         make(map[string]*Node),
		rules:         make(map[string]*Rule),
		subscriptions: make(map[string]*Subscription),
		trafficStats: &TrafficStats{
			Time: time.Now(),
		},
	}
}

// GetNodes returns all nodes
func (m *ProxyManager) GetNodes() []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	nodes := make([]*Node, 0, len(m.nodes))
	for _, node := range m.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// CreateNode creates a new node
func (m *ProxyManager) CreateNode(node *Node) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate node
	if node.ID == "" {
		return fmt.Errorf("node ID is required")
	}
	if _, exists := m.nodes[node.ID]; exists {
		return fmt.Errorf("node already exists: %s", node.ID)
	}

	// Set creation time
	node.CreatedAt = time.Now()

	// Initialize traffic stats
	node.Traffic = &Traffic{
		Unit: "B",
	}

	m.nodes[node.ID] = node
	return nil
}

// UpdateNode updates a node
func (m *ProxyManager) UpdateNode(node *Node) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.nodes[node.ID]; !exists {
		return fmt.Errorf("node not found: %s", node.ID)
	}

	m.nodes[node.ID] = node
	return nil
}

// DeleteNode deletes a node
func (m *ProxyManager) DeleteNode(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.nodes[id]; !exists {
		return fmt.Errorf("node not found: %s", id)
	}

	delete(m.nodes, id)
	return nil
}

// GetRules returns all rules
func (m *ProxyManager) GetRules() []*Rule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rules := make([]*Rule, 0, len(m.rules))
	for _, rule := range m.rules {
		rules = append(rules, rule)
	}
	return rules
}

// CreateRule creates a new rule
func (m *ProxyManager) CreateRule(rule *Rule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}
	if _, exists := m.rules[rule.ID]; exists {
		return fmt.Errorf("rule already exists: %s", rule.ID)
	}

	rule.CreatedAt = time.Now()
	m.rules[rule.ID] = rule
	return nil
}

// UpdateRule updates a rule
func (m *ProxyManager) UpdateRule(rule *Rule) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rules[rule.ID]; !exists {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}

	m.rules[rule.ID] = rule
	return nil
}

// DeleteRule deletes a rule
func (m *ProxyManager) DeleteRule(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rules[id]; !exists {
		return fmt.Errorf("rule not found: %s", id)
	}

	delete(m.rules, id)
	return nil
}

// GetSubscriptions returns all subscriptions
func (m *ProxyManager) GetSubscriptions() []*Subscription {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subs := make([]*Subscription, 0, len(m.subscriptions))
	for _, sub := range m.subscriptions {
		subs = append(subs, sub)
	}
	return subs
}

// CreateSubscription creates a new subscription
func (m *ProxyManager) CreateSubscription(sub *Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub.ID == "" {
		return fmt.Errorf("subscription ID is required")
	}
	if _, exists := m.subscriptions[sub.ID]; exists {
		return fmt.Errorf("subscription already exists: %s", sub.ID)
	}

	// Update nodes immediately
	if err := m.updateSubscriptionNodesLocked(sub); err != nil {
		return fmt.Errorf("failed to update subscription nodes: %v", err)
	}

	m.subscriptions[sub.ID] = sub
	return nil
}

// UpdateSubscription updates a subscription
func (m *ProxyManager) UpdateSubscription(sub *Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.subscriptions[sub.ID]; !exists {
		return fmt.Errorf("subscription not found: %s", sub.ID)
	}

	m.subscriptions[sub.ID] = sub
	return nil
}

// DeleteSubscription deletes a subscription
func (m *ProxyManager) DeleteSubscription(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.subscriptions[id]; !exists {
		return fmt.Errorf("subscription not found: %s", id)
	}

	delete(m.subscriptions, id)
	return nil
}

// UpdateSubscriptionNodes updates nodes from subscription
func (m *ProxyManager) UpdateSubscriptionNodes(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sub, exists := m.subscriptions[id]
	if !exists {
		return fmt.Errorf("subscription not found: %s", id)
	}

	return m.updateSubscriptionNodesLocked(sub)
}

// updateSubscriptionNodesLocked updates subscription nodes (must be called with lock held)
func (m *ProxyManager) updateSubscriptionNodesLocked(sub *Subscription) error {
	// Download subscription content
	resp, err := http.Get(sub.URL)
	if err != nil {
		return fmt.Errorf("failed to download subscription: %v", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read subscription content: %v", err)
	}

	// Parse nodes based on subscription type
	var nodes []*Node
	switch sub.Type {
	case "clash":
		nodes, err = m.parseClashSubscription(content)
	case "singbox":
		nodes, err = m.parseSingboxSubscription(content)
	case "ss":
		nodes, err = m.parseSSSubscription(content)
	case "v2ray":
		nodes, err = m.parseV2raySubscription(content)
	default:
		return fmt.Errorf("unsupported subscription type: %s", sub.Type)
	}
	if err != nil {
		return fmt.Errorf("failed to parse subscription: %v", err)
	}

	// Update subscription
	sub.Nodes = nodes
	sub.UpdatedAt = time.Now()

	return nil
}

// parseClashSubscription parses Clash subscription
func (m *ProxyManager) parseClashSubscription(content []byte) ([]*Node, error) {
	var config struct {
		Proxies []map[string]interface{} `yaml:"proxies"`
	}
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("invalid Clash config: %v", err)
	}

	nodes := make([]*Node, 0, len(config.Proxies))
	for _, proxy := range config.Proxies {
		node := &Node{
			Name:      proxy["name"].(string),
			Type:      proxy["type"].(string),
			Address:   proxy["server"].(string),
			Port:      int(proxy["port"].(float64)),
			CreatedAt: time.Now(),
		}

		// Set protocol-specific fields
		switch node.Type {
		case "ss":
			node.Method = proxy["cipher"].(string)
			node.Password = proxy["password"].(string)
		case "vmess":
			node.UUID = proxy["uuid"].(string)
			if alterId, ok := proxy["alterId"]; ok {
				node.AlterId = int(alterId.(float64))
			}
		case "trojan":
			node.Password = proxy["password"].(string)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// parseSingboxSubscription parses sing-box subscription
func (m *ProxyManager) parseSingboxSubscription(content []byte) ([]*Node, error) {
	var config struct {
		Outbounds []map[string]interface{} `json:"outbounds"`
	}
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("invalid sing-box config: %v", err)
	}

	nodes := make([]*Node, 0, len(config.Outbounds))
	for _, outbound := range config.Outbounds {
		node := &Node{
			Name:      outbound["tag"].(string),
			Type:      outbound["type"].(string),
			Address:   outbound["server"].(string),
			Port:      int(outbound["server_port"].(float64)),
			CreatedAt: time.Now(),
		}

		// Set protocol-specific fields
		switch node.Type {
		case "shadowsocks":
			node.Type = "ss"
			node.Method = outbound["method"].(string)
			node.Password = outbound["password"].(string)
		case "vmess":
			node.UUID = outbound["uuid"].(string)
			if alterId, ok := outbound["alter_id"]; ok {
				node.AlterId = int(alterId.(float64))
			}
		case "trojan":
			node.Password = outbound["password"].(string)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// parseSSSubscription parses ShadowSocks subscription
func (m *ProxyManager) parseSSSubscription(content []byte) ([]*Node, error) {
	lines := strings.Split(string(content), "\n")
	nodes := make([]*Node, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if u, err := url.Parse(line); err == nil && u.Scheme == "ss" {
			if node, err := Protocols["ss"].ParseURL(u); err == nil {
				node.CreatedAt = time.Now()
				nodes = append(nodes, node)
			}
		}
	}

	return nodes, nil
}

// parseV2raySubscription parses V2Ray subscription
func (m *ProxyManager) parseV2raySubscription(content []byte) ([]*Node, error) {
	lines := strings.Split(string(content), "\n")
	nodes := make([]*Node, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if u, err := url.Parse(line); err == nil {
			var node *Node
			var err error

			switch u.Scheme {
			case "vmess":
				node, err = Protocols["vmess"].ParseURL(u)
			case "vless":
				node, err = Protocols["vless"].ParseURL(u)
			case "trojan":
				node, err = Protocols["trojan"].ParseURL(u)
			}

			if err == nil && node != nil {
				node.CreatedAt = time.Now()
				nodes = append(nodes, node)
			}
		}
	}

	return nodes, nil
}

// GetTrafficStats returns traffic statistics
func (m *ProxyManager) GetTrafficStats() *TrafficStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.trafficStats
}

// GetRealtimeTraffic returns realtime traffic statistics
func (m *ProxyManager) GetRealtimeTraffic() *TrafficStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.trafficStats
}

// UpdateTrafficStats updates traffic statistics
func (m *ProxyManager) UpdateTrafficStats(upload, download int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.trafficStats.Time = time.Now()
	m.trafficStats.Upload = upload
	m.trafficStats.Download = download
	m.trafficStats.TotalUp += upload
	m.trafficStats.TotalDown += download
}
