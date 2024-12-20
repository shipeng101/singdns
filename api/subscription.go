package api

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"singdns/api/logger"
	"singdns/api/models"
	"singdns/api/protocols"
	"singdns/api/subscription"
)

// SubscriptionManager manages node subscriptions
type SubscriptionManager struct {
	subscriptions map[string]*models.Subscription
	updateChan    chan string
	client        *http.Client
	checker       *protocols.NodeChecker
	converter     *subscription.SubscriptionConverter
	mu            sync.RWMutex
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager() *SubscriptionManager {
	m := &SubscriptionManager{
		subscriptions: make(map[string]*models.Subscription),
		updateChan:    make(chan string, 100),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		checker:   protocols.NewNodeChecker(),
		converter: subscription.NewSubscriptionConverter(),
	}

	// Start update worker
	go m.updateWorker()

	return m
}

// GetSubscriptions returns all subscriptions
func (m *SubscriptionManager) GetSubscriptions() []*models.Subscription {
	m.mu.RLock()
	defer m.mu.RUnlock()

	subs := make([]*models.Subscription, 0, len(m.subscriptions))
	for _, sub := range m.subscriptions {
		subs = append(subs, sub)
	}
	return subs
}

// CreateSubscription creates a new subscription
func (m *SubscriptionManager) CreateSubscription(sub *models.Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub.ID == "" {
		return fmt.Errorf("subscription ID is required")
	}
	if _, exists := m.subscriptions[sub.ID]; exists {
		return fmt.Errorf("subscription already exists: %s", sub.ID)
	}

	// Update nodes immediately
	if err := m.updateSubscriptionLocked(sub); err != nil {
		return fmt.Errorf("failed to update subscription nodes: %v", err)
	}

	m.subscriptions[sub.ID] = sub
	return nil
}

// UpdateSubscription updates a subscription
func (m *SubscriptionManager) UpdateSubscription(sub *models.Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.subscriptions[sub.ID]; !exists {
		return fmt.Errorf("subscription not found: %s", sub.ID)
	}

	m.subscriptions[sub.ID] = sub
	return nil
}

// DeleteSubscription deletes a subscription
func (m *SubscriptionManager) DeleteSubscription(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.subscriptions[id]; !exists {
		return fmt.Errorf("subscription not found: %s", id)
	}

	delete(m.subscriptions, id)
	return nil
}

// UpdateSubscriptionNodes updates nodes from subscription
func (m *SubscriptionManager) UpdateSubscriptionNodes(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sub, exists := m.subscriptions[id]
	if !exists {
		return fmt.Errorf("subscription not found: %s", id)
	}

	return m.updateSubscriptionLocked(sub)
}

// QueueUpdate queues a subscription for update
func (m *SubscriptionManager) QueueUpdate(id string) {
	select {
	case m.updateChan <- id:
	default:
		logger.LogWarning("Update queue is full: %s", id)
	}
}

// updateWorker processes subscription updates
func (m *SubscriptionManager) updateWorker() {
	for id := range m.updateChan {
		m.mu.Lock()
		sub, exists := m.subscriptions[id]
		if !exists {
			m.mu.Unlock()
			continue
		}

		if err := m.updateSubscriptionLocked(sub); err != nil {
			logger.LogError("Failed to update subscription: %v", err)
		}
		m.mu.Unlock()

		// Rate limit updates
		time.Sleep(time.Second)
	}
}

// updateSubscriptionLocked updates subscription nodes (must be called with lock held)
func (m *SubscriptionManager) updateSubscriptionLocked(sub *models.Subscription) error {
	// Download subscription content
	resp, err := m.client.Get(sub.URL)
	if err != nil {
		return fmt.Errorf("failed to download subscription: %v", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read subscription content: %v", err)
	}

	// Parse nodes based on subscription type
	var nodes []*protocols.Node
	switch sub.Type {
	case "clash":
		nodes, err = subscription.ParseClash(content)
	case "singbox":
		nodes, err = subscription.ParseSingbox(content)
	case "ss", "vmess", "trojan", "vless", "hy2", "tuic":
		// Try base64 decode
		if decoded, err := base64.StdEncoding.DecodeString(string(content)); err == nil {
			content = decoded
		}
		nodes, err = m.parseTextSubscription(content)
	default:
		return fmt.Errorf("unsupported subscription type: %s", sub.Type)
	}
	if err != nil {
		return fmt.Errorf("failed to parse subscription: %v", err)
	}

	// Convert protocol nodes to subscription nodes
	sub.Nodes = make([]*models.Node, len(nodes))
	for i, node := range nodes {
		sub.Nodes[i] = &models.Node{
			ID:        fmt.Sprintf("%s_%d", sub.ID, i),
			Name:      node.Name,
			Type:      node.Type,
			Address:   node.Address,
			Port:      node.Port,
			Method:    node.Method,
			Password:  node.Password,
			UUID:      node.UUID,
			AlterId:   node.AlterId,
			Network:   node.Network,
			Path:      node.Path,
			Host:      node.Host,
			TLS:       node.TLS,
			Flow:      node.Flow,
			Up:        node.Up,
			Down:      node.Down,
			CC:        node.CC,
			CreatedAt: time.Now(),
		}
	}

	// Check node availability
	results := m.checker.CheckNodes(nodes)
	for i, result := range results {
		if !result.Available {
			logger.LogWarning("Node is unavailable: %s (%s)", nodes[i].Name, result.Error)
		}
	}

	sub.UpdatedAt = time.Now()
	return nil
}

// ConvertSubscription converts subscription to different format
func (m *SubscriptionManager) ConvertSubscription(id string, format string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sub, exists := m.subscriptions[id]
	if !exists {
		return "", fmt.Errorf("subscription not found: %s", id)
	}

	// Convert nodes to protocol nodes
	nodes := make([]*protocols.Node, len(sub.Nodes))
	for i, node := range sub.Nodes {
		nodes[i] = &protocols.Node{
			Name:      node.Name,
			Type:      node.Type,
			Address:   node.Address,
			Port:      node.Port,
			Method:    node.Method,
			Password:  node.Password,
			UUID:      node.UUID,
			AlterId:   node.AlterId,
			Network:   node.Network,
			Path:      node.Path,
			Host:      node.Host,
			TLS:       node.TLS,
			Flow:      node.Flow,
			Up:        node.Up,
			Down:      node.Down,
			CC:        node.CC,
			CreatedAt: node.CreatedAt,
		}
	}

	// Convert to target format
	switch format {
	case "text":
		return m.converter.ConvertToText(nodes)
	case "clash":
		return m.converter.ConvertToClash(nodes)
	case "singbox":
		return m.converter.ConvertToSingbox(nodes)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// GetBestNode gets the best available node from subscription
func (m *SubscriptionManager) GetBestNode(id string) (*models.Node, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sub, exists := m.subscriptions[id]
	if !exists {
		return nil, fmt.Errorf("subscription not found: %s", id)
	}

	// Convert nodes to protocol nodes
	nodes := make([]*protocols.Node, len(sub.Nodes))
	for i, node := range sub.Nodes {
		nodes[i] = &protocols.Node{
			Name:      node.Name,
			Type:      node.Type,
			Address:   node.Address,
			Port:      node.Port,
			Method:    node.Method,
			Password:  node.Password,
			UUID:      node.UUID,
			AlterId:   node.AlterId,
			Network:   node.Network,
			Path:      node.Path,
			Host:      node.Host,
			TLS:       node.TLS,
			Flow:      node.Flow,
			Up:        node.Up,
			Down:      node.Down,
			CC:        node.CC,
			CreatedAt: node.CreatedAt,
		}
	}

	// Get best node
	best := m.checker.GetBestNode(nodes)
	if best == nil {
		return nil, fmt.Errorf("no available nodes")
	}

	// Convert back to subscription node
	return &models.Node{
		ID:        fmt.Sprintf("%s_best", id),
		Name:      best.Name,
		Type:      best.Type,
		Address:   best.Address,
		Port:      best.Port,
		Method:    best.Method,
		Password:  best.Password,
		UUID:      best.UUID,
		AlterId:   best.AlterId,
		Network:   best.Network,
		Path:      best.Path,
		Host:      best.Host,
		TLS:       best.TLS,
		Flow:      best.Flow,
		Up:        best.Up,
		Down:      best.Down,
		CC:        best.CC,
		CreatedAt: best.CreatedAt,
	}, nil
}

// parseTextSubscription parses text subscription (ss/vmess/trojan)
func (m *SubscriptionManager) parseTextSubscription(content []byte) ([]*protocols.Node, error) {
	lines := strings.Split(string(content), "\n")
	nodes := make([]*protocols.Node, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		node, err := protocols.ParseURL(line)
		if err != nil {
			logger.LogWarning("Failed to parse node URL: %s (%s)", line, err)
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}
