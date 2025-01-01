package protocols

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"singdns/api/logger"
	"singdns/api/models"
	"sync"
	"time"
)

// ProxyNodeWrapper wraps a models.Node for checking
type ProxyNodeWrapper struct {
	*models.Node
}

// CheckResult represents the result of a node check
type CheckResult struct {
	Node      *models.Node
	Available bool
	Latency   int64
	Error     string
	CheckedAt time.Time
}

// NodeChecker checks node availability and latency
type NodeChecker struct {
	results map[string]*CheckResult
	mu      sync.RWMutex
}

// NewNodeChecker creates a new node checker
func NewNodeChecker() *NodeChecker {
	return &NodeChecker{
		results: make(map[string]*CheckResult),
	}
}

// CheckNode checks a single node
func (c *NodeChecker) CheckNode(node *models.Node) *CheckResult {
	result := &CheckResult{
		Node:      node,
		CheckedAt: time.Now(),
	}

	logger.LogDebug("开始检测节点: %s (%s://%s:%d)", node.Name, node.Type, node.Address, node.Port)

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建 TCP 连接
	var d net.Dialer
	start := time.Now()

	// 根据不同协议类型进行不同的检测
	switch node.Type {
	case "vmess", "vless", "trojan", "ss", "shadowsocks":
		// 对于加密代理协议，先测试 TCP 连接
		conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", node.Address, node.Port))
		if err != nil {
			result.Error = fmt.Sprintf("TCP 连接失败: %v", err)
			logger.LogDebug("节点 %s 连接失败: %v", node.Name, err)
			return result
		}
		defer conn.Close()

		// 如果需要 TLS，进行 TLS 握手
		if node.TLS {
			tlsConfig := &tls.Config{
				ServerName:         node.Host,
				InsecureSkipVerify: node.SkipCertVerify,
			}
			tlsConn := tls.Client(conn, tlsConfig)
			if err := tlsConn.HandshakeContext(ctx); err != nil {
				result.Error = fmt.Sprintf("TLS 握手失败: %v", err)
				logger.LogDebug("节点 %s TLS 握手失败: %v", node.Name, err)
				return result
			}
			defer tlsConn.Close()
		}

		result.Available = true
		result.Latency = time.Since(start).Milliseconds()

	case "hysteria2", "tuic":
		// 对于基于 UDP 的协议，测试 UDP 可达性
		if node.TLS {
			// 先测试 TCP 连接以验证 TLS
			conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", node.Address, node.Port))
			if err != nil {
				result.Error = fmt.Sprintf("TCP 连接失败: %v", err)
				logger.LogDebug("节点 %s 连接失败: %v", node.Name, err)
				return result
			}
			defer conn.Close()

			tlsConfig := &tls.Config{
				ServerName:         node.Host,
				InsecureSkipVerify: node.SkipCertVerify,
			}
			tlsConn := tls.Client(conn, tlsConfig)
			if err := tlsConn.HandshakeContext(ctx); err != nil {
				result.Error = fmt.Sprintf("TLS 握手失败: %v", err)
				logger.LogDebug("节点 %s TLS 握手失败: %v", node.Name, err)
				return result
			}
			defer tlsConn.Close()
		}

		// 测试 UDP 连接
		udpConn, err := net.ListenUDP("udp", nil)
		if err != nil {
			result.Error = fmt.Sprintf("UDP 连接失败: %v", err)
			logger.LogDebug("节点 %s UDP 连接失败: %v", node.Name, err)
			return result
		}
		defer udpConn.Close()

		result.Available = true
		result.Latency = time.Since(start).Milliseconds()

	default:
		result.Error = fmt.Sprintf("不支持的协议类型: %s", node.Type)
		logger.LogDebug("节点 %s 使用了不支持的协议: %s", node.Name, node.Type)
		return result
	}

	logger.LogDebug("节点 %s 检测完成: 可用=%v, 延迟=%dms", node.Name, result.Available, result.Latency)

	// 存储结果
	c.mu.Lock()
	c.results[node.ID] = result
	c.mu.Unlock()

	return result
}

// CheckNodes checks multiple nodes concurrently
func (c *NodeChecker) CheckNodes(nodes []models.Node) []*CheckResult {
	results := make([]*CheckResult, len(nodes))
	var wg sync.WaitGroup
	wg.Add(len(nodes))

	// 限制并发数量
	semaphore := make(chan struct{}, 10)

	for i := range nodes {
		go func(i int, node *models.Node) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量
			results[i] = c.CheckNode(node)
		}(i, &nodes[i])
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
func (c *NodeChecker) StartPeriodicCheck(ctx context.Context, nodes []models.Node, interval time.Duration) {
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
func (c *NodeChecker) GetBestNode(nodes []models.Node) *models.Node {
	results := c.CheckNodes(nodes)
	var bestNode *models.Node
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
