package subscription

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/shipeng101/singdns/pkg/types"
	"github.com/shipeng101/singdns/service/system"
	"go.uber.org/zap"
)

type service struct {
	nodes      []*types.ProxyNode
	parser     Parser
	configPath string
	mu         sync.RWMutex
}

// NewService 创建订阅服务
func NewService() types.SubscriptionService {
	return &service{
		nodes:      make([]*types.ProxyNode, 0),
		configPath: "config/subscription.json",
		parser:     NewParser(),
	}
}

// GetNodes 获取节点列表
func (s *service) GetNodes() ([]*types.ProxyNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 如果节点列表为空，尝试从配置文件加载
	if len(s.nodes) == 0 {
		if err := s.loadConfig(); err != nil {
			return nil, err
		}
	}

	return s.nodes, nil
}

// UpdateSubscription 更新订阅
func (s *service) UpdateSubscription(url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 下载订阅内容
	content, err := s.downloadSubscription(url)
	if err != nil {
		return fmt.Errorf("下载订阅失败: %v", err)
	}

	// 解析节点
	nodes, err := s.parser.Parse(content)
	if err != nil {
		return fmt.Errorf("解析订阅失败: %v", err)
	}

	// 更新节点列表
	s.nodes = nodes

	// 保存配置
	return s.saveConfig()
}

// SelectNode 选择节点
func (s *service) SelectNode(id string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 如果节点列表为空，尝试从配置文件加载
	if len(s.nodes) == 0 {
		if err := s.loadConfig(); err != nil {
			return err
		}
	}

	// 查找节点
	var selectedNode *types.ProxyNode
	for _, node := range s.nodes {
		if node.ID == id {
			selectedNode = node
			break
		}
	}

	if selectedNode == nil {
		return fmt.Errorf("节点不存在")
	}

	// 生成节点配置
	config := map[string]interface{}{
		"log": map[string]interface{}{
			"level":     "info",
			"timestamp": true,
		},
		"dns": map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"tag":              "local",
					"address":          "223.5.5.5",
					"detour":           "direct",
					"address_resolver": "dns-direct",
				},
				{
					"tag":              "remote",
					"address":          "8.8.8.8",
					"detour":           "proxy",
					"address_resolver": "dns-direct",
				},
				{
					"tag":     "dns-direct",
					"address": "223.5.5.5",
				},
			},
			"rules": []map[string]interface{}{
				{
					"domain_suffix": []string{".cn"},
					"server":        "local",
				},
			},
			"strategy": "prefer_ipv4",
			"final":    "remote",
		},
		"inbounds": []map[string]interface{}{
			{
				"type":        "mixed",
				"tag":         "mixed-in",
				"listen":      "127.0.0.1",
				"listen_port": 7890,
				"sniff":       true,
			},
		},
		"outbounds": []map[string]interface{}{
			{
				"type":        selectedNode.Type,
				"tag":         "proxy",
				"server":      selectedNode.Server,
				"server_port": selectedNode.Port,
				"method":      selectedNode.Method,
				"password":    selectedNode.Password,
				"network":     "tcp",
			},
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
		},
		"route": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"protocol": "dns",
					"outbound": "dns-out",
				},
			},
			"final": "proxy",
		},
	}

	// 保存配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("生成配置失败: %v", err)
	}

	// 保存到 core/config.json
	if err := os.WriteFile("core/config.json", data, 0644); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	return nil
}

// loadConfig 加载配置
func (s *service) loadConfig() error {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.saveConfig()
		}
		return err
	}

	var config struct {
		Nodes []*types.ProxyNode `json:"nodes"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	s.nodes = config.Nodes
	return nil
}

// saveConfig 保存配置
func (s *service) saveConfig() error {
	// 创建配置目录
	if err := os.MkdirAll(filepath.Dir(s.configPath), 0755); err != nil {
		return err
	}

	config := struct {
		Nodes []*types.ProxyNode `json:"nodes"`
	}{
		Nodes: s.nodes,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.configPath, data, 0644)
}

// downloadSubscription 下载订阅内容
func (s *service) downloadSubscription(url string) (string, error) {
	system.Info("正在下载订阅", zap.String("url", url))

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Start 启动服务
func (s *service) Start() error {
	// 启动时加载配置
	return s.loadConfig()
}

// Stop 停止服务
func (s *service) Stop() error {
	// 停止时保存配置
	return s.saveConfig()
}
