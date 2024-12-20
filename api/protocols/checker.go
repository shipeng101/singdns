package protocols

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// CheckResult represents a node check result
type CheckResult struct {
	Node      *Node     `json:"node"`
	Available bool      `json:"available"`
	Latency   int64     `json:"latency"` // in milliseconds
	Error     string    `json:"error,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

// NodeChecker checks node availability and latency
type NodeChecker struct {
	client    *http.Client
	transport *http.Transport
	results   map[string]*CheckResult
	mu        sync.RWMutex
}

// NewNodeChecker creates a new node checker
func NewNodeChecker() *NodeChecker {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &NodeChecker{
		client: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
		transport: transport,
		results:   make(map[string]*CheckResult),
	}
}

// CheckNode checks a single node
func (c *NodeChecker) CheckNode(node *Node) *CheckResult {
	result := &CheckResult{
		Node:      node,
		CheckedAt: time.Now(),
	}

	// Test TCP connection
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", node.Address, node.Port), 5*time.Second)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer conn.Close()

	result.Available = true
	result.Latency = time.Since(start).Milliseconds()

	// Store result
	c.mu.Lock()
	c.results[node.ID] = result
	c.mu.Unlock()

	return result
}

// CheckNodes checks multiple nodes concurrently
func (c *NodeChecker) CheckNodes(nodes []*Node) []*CheckResult {
	results := make([]*CheckResult, len(nodes))
	var wg sync.WaitGroup
	wg.Add(len(nodes))

	for i, node := range nodes {
		go func(i int, node *Node) {
			defer wg.Done()
			results[i] = c.CheckNode(node)
		}(i, node)
	}

	wg.Wait()
	return results
}

// GetResult gets the check result for a node
func (c *NodeChecker) GetResult(nodeID string) *CheckResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.results[nodeID]
}

// GetResults gets all check results
func (c *NodeChecker) GetResults() []*CheckResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make([]*CheckResult, 0, len(c.results))
	for _, result := range c.results {
		results = append(results, result)
	}
	return results
}

// ClearResults clears all check results
func (c *NodeChecker) ClearResults() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = make(map[string]*CheckResult)
}

// StartPeriodicCheck starts periodic node checking
func (c *NodeChecker) StartPeriodicCheck(ctx context.Context, nodes []*Node, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.CheckNodes(nodes)
		}
	}
}

// GetBestNode gets the node with lowest latency
func (c *NodeChecker) GetBestNode(nodes []*Node) *Node {
	results := c.CheckNodes(nodes)
	var bestNode *Node
	var bestLatency int64 = -1

	for _, result := range results {
		if !result.Available {
			continue
		}
		if bestLatency == -1 || result.Latency < bestLatency {
			bestNode = result.Node
			bestLatency = result.Latency
		}
	}

	return bestNode
}
