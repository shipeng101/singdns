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

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	system.Info("订阅下载成功", zap.Int("size", len(body)))
	return string(body), nil
}

// Start 启动服务
func (s *service) Start() error {
	return s.loadConfig()
}

// Stop 停止服务
func (s *service) Stop() error {
	return nil
}
