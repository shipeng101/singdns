package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/shipeng101/singdns/pkg/types"
	"github.com/shipeng101/singdns/service/system"
	"go.uber.org/zap"
)

type service struct {
	config              types.ProxyConfig
	stats               *types.ProxyStats
	running             bool
	currentNode         string
	startTime           time.Time
	mu                  sync.RWMutex
	ctx                 context.Context
	cancel              context.CancelFunc
	subscriptionService types.SubscriptionService
	configPath          string
	iptablesService     *iptablesService
	configGenerator     *ConfigGenerator
}

// NewService 创建代理服务
func NewService(cfg types.ProxyConfig, subscriptionService types.SubscriptionService) types.ProxyService {
	ctx, cancel := context.WithCancel(context.Background())
	return &service{
		config:              cfg,
		stats:               &types.ProxyStats{},
		running:             false,
		ctx:                 ctx,
		cancel:              cancel,
		subscriptionService: subscriptionService,
		configPath:          "config/proxy.json",
		iptablesService:     newIPTablesService(),
		configGenerator:     NewConfigGenerator(),
	}
}

// Start 启动服务
func (s *service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("服务已在运行")
	}

	// 加载配置
	if err := s.loadConfig(); err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 生成sing-box配置
	if err := s.generateConfig(); err != nil {
		return fmt.Errorf("生成配置失败: %v", err)
	}

	// 启动sing-box进程
	if err := s.startProcess(); err != nil {
		return fmt.Errorf("启动进程失败: %v", err)
	}

	// 设置系统代理
	if s.config.Inbound.SetSystemProxy {
		if err := s.setSystemProxy(); err != nil {
			system.Warn("设置系统代理失败", zap.Error(err))
		}
	}

	// 设置iptables规则
	if runtime.GOOS == "linux" && s.config.Inbound.EnableTransparent {
		if err := s.setupIPTables(); err != nil {
			system.Warn("设置iptables规则失败", zap.Error(err))
		}
	}

	s.running = true
	s.startTime = time.Now()

	system.Info("代理服务启动成功",
		zap.String("mode", s.config.Mode))

	return nil
}

// Stop 停止服务
func (s *service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// 取消系统代理
	if s.config.Inbound.SetSystemProxy {
		if err := s.unsetSystemProxy(); err != nil {
			system.Warn("取消系统代理失败", zap.Error(err))
		}
	}

	// 清理iptables规则
	if runtime.GOOS == "linux" && s.config.Inbound.EnableTransparent {
		if err := s.cleanIPTables(); err != nil {
			system.Warn("清理iptables规则失败", zap.Error(err))
		}
	}

	// 停止sing-box进程
	s.cancel()

	s.running = false
	return nil
}

// Restart 重启服务
func (s *service) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.Start()
}

// AddNode 添加节点
func (s *service) AddNode(node *types.ProxyNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证节点配置
	if err := s.validateNode(node); err != nil {
		return err
	}

	// 添加节点
	s.config.Nodes = append(s.config.Nodes, node)

	// 保存配置
	return s.saveConfig()
}

// UpdateNode 更新节点
func (s *service) UpdateNode(node *types.ProxyNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 验证节点配置
	if err := s.validateNode(node); err != nil {
		return err
	}

	// 查找并更新节点
	for i, n := range s.config.Nodes {
		if n.ID == node.ID {
			s.config.Nodes[i] = node
			return s.saveConfig()
		}
	}

	return fmt.Errorf("节点不存在")
}

// DeleteNode 删除节点
func (s *service) DeleteNode(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并删除节点
	for i, node := range s.config.Nodes {
		if node.ID == id {
			s.config.Nodes = append(s.config.Nodes[:i], s.config.Nodes[i+1:]...)
			return s.saveConfig()
		}
	}

	return fmt.Errorf("节点不存在")
}

// GetNode 获取节点
func (s *service) GetNode(id string) (*types.ProxyNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, node := range s.config.Nodes {
		if node.ID == id {
			return node, nil
		}
	}

	return nil, fmt.Errorf("节点不存在")
}

// ListNodes 获取节点列表
func (s *service) ListNodes() ([]*types.ProxyNode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config.Nodes, nil
}

// TestNode 测试节点延迟
func (s *service) TestNode(id string) (int64, error) {
	node, err := s.GetNode(id)
	if err != nil {
		return 0, err
	}

	// 测试延迟
	start := time.Now()
	cmd := exec.Command("ping", "-c", "1", node.Server)
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("测试失败: %v", err)
	}

	latency := time.Since(start).Milliseconds()
	node.Latency = latency
	node.LastTest = time.Now()

	return latency, nil
}

// TestAllNodes 测试所有节点
func (s *service) TestAllNodes() error {
	nodes, err := s.ListNodes()
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if _, err := s.TestNode(node.ID); err != nil {
			system.Warn("节点测试失败",
				zap.String("node", node.Name),
				zap.Error(err))
		}
	}

	return nil
}

// SwitchNode 切换节点
func (s *service) SwitchNode(id string) error {
	// 检查节点是否存在
	if _, err := s.GetNode(id); err != nil {
		return err
	}

	s.mu.Lock()
	s.currentNode = id
	s.mu.Unlock()

	// 重新生成配置并重启服务
	return s.Restart()
}

// GetStats 获取统计信息
func (s *service) GetStats() (*types.ProxyStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.stats, nil
}

// SetMode 设置代理模式
func (s *service) SetMode(mode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if mode != "rule" && mode != "global" && mode != "direct" {
		return fmt.Errorf("不支持的代理模式")
	}

	s.config.Mode = mode
	return s.saveConfig()
}

// UpdateRules 更新规则
func (s *service) UpdateRules(rules []types.Rule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config.Rules = rules
	return s.saveConfig()
}

// GetRules 获取规则
func (s *service) GetRules() ([]types.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.config.Rules, nil
}

// UpdateIPTables 更新iptables规则
func (s *service) UpdateIPTables(rules []types.IPTablesRule) error {
	return s.iptablesService.UpdateRules(rules)
}

// GetIPTables 获取iptables规则
func (s *service) GetIPTables() ([]types.IPTablesRule, error) {
	return s.iptablesService.GetRules(), nil
}

// 内部辅助方法

// loadConfig 加载配置
func (s *service) loadConfig() error {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.saveConfig()
		}
		return err
	}

	return json.Unmarshal(data, &s.config)
}

// saveConfig 保存配置
func (s *service) saveConfig() error {
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}

	// 创建配置目录
	if err := os.MkdirAll(filepath.Dir(s.configPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(s.configPath, data, 0644)
}

// generateConfig 生成sing-box配置
func (s *service) generateConfig() error {
	// 获取当前节点
	var currentNode *types.ProxyNode
	if s.currentNode != "" {
		for _, node := range s.config.Nodes {
			if node.ID == s.currentNode {
				currentNode = node
				break
			}
		}
	}

	// 生成配置
	config, err := s.configGenerator.Generate(s.config, currentNode)
	if err != nil {
		return err
	}

	// 保存配置文件
	return os.WriteFile("config/sing-box.json", config, 0644)
}

// startProcess 启动sing-box进程
func (s *service) startProcess() error {
	cmd := exec.CommandContext(s.ctx, "core/sing-box", "run", "-c", "config/sing-box.json")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

// validateNode 验证节点配置
func (s *service) validateNode(node *types.ProxyNode) error {
	if node.Server == "" {
		return fmt.Errorf("服务器地址不能为空")
	}
	if node.Port == 0 {
		return fmt.Errorf("服务器端口不能为空")
	}
	if node.Type == "" {
		return fmt.Errorf("节点类型不能为空")
	}
	return nil
}

// setSystemProxy 设置系统代理
func (s *service) setSystemProxy() error {
	switch runtime.GOOS {
	case "darwin":
		return s.setMacOSProxy()
	case "linux":
		return s.setLinuxProxy()
	case "windows":
		return s.setWindowsProxy()
	default:
		return fmt.Errorf("不支持的操作系统")
	}
}

// unsetSystemProxy 取消系统代理
func (s *service) unsetSystemProxy() error {
	switch runtime.GOOS {
	case "darwin":
		return s.unsetMacOSProxy()
	case "linux":
		return s.unsetLinuxProxy()
	case "windows":
		return s.unsetWindowsProxy()
	default:
		return fmt.Errorf("不支持的操作系统")
	}
}

// setMacOSProxy 设置macOS系统代理
func (s *service) setMacOSProxy() error {
	cmd := exec.Command("networksetup", "-setwebproxy", "Wi-Fi", "127.0.0.1", fmt.Sprint(s.config.Inbound.HTTPPort))
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("networksetup", "-setsecurewebproxy", "Wi-Fi", "127.0.0.1", fmt.Sprint(s.config.Inbound.HTTPPort))
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("networksetup", "-setsocksfirewallproxy", "Wi-Fi", "127.0.0.1", fmt.Sprint(s.config.Inbound.SocksPort))
	return cmd.Run()
}

// unsetMacOSProxy 取消macOS系统代理
func (s *service) unsetMacOSProxy() error {
	cmd := exec.Command("networksetup", "-setwebproxystate", "Wi-Fi", "off")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("networksetup", "-setsecurewebproxystate", "Wi-Fi", "off")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("networksetup", "-setsocksfirewallproxystate", "Wi-Fi", "off")
	return cmd.Run()
}

// setLinuxProxy 设置Linux系统代理
func (s *service) setLinuxProxy() error {
	// 设置环境变量
	os.Setenv("http_proxy", fmt.Sprintf("http://127.0.0.1:%d", s.config.Inbound.HTTPPort))
	os.Setenv("https_proxy", fmt.Sprintf("http://127.0.0.1:%d", s.config.Inbound.HTTPPort))
	os.Setenv("all_proxy", fmt.Sprintf("socks5://127.0.0.1:%d", s.config.Inbound.SocksPort))
	return nil
}

// unsetLinuxProxy 取消Linux系统代理
func (s *service) unsetLinuxProxy() error {
	os.Unsetenv("http_proxy")
	os.Unsetenv("https_proxy")
	os.Unsetenv("all_proxy")
	return nil
}

// setWindowsProxy 设置Windows系统代理
func (s *service) setWindowsProxy() error {
	// TODO: 实现Windows系统代理设置
	return fmt.Errorf("暂不支持Windows系统代理设置")
}

// unsetWindowsProxy 取消Windows系统代理
func (s *service) unsetWindowsProxy() error {
	// TODO: 实现Windows系统代理取消
	return fmt.Errorf("暂不支持Windows系统代理取消")
}

// setupIPTables 设置iptables规则
func (s *service) setupIPTables() error {
	rules := []types.IPTablesRule{
		{
			Table:  "nat",
			Chain:  "PREROUTING",
			Target: "REDIRECT",
			ToPort: fmt.Sprint(s.config.Inbound.Port),
		},
	}
	return s.iptablesService.UpdateRules(rules)
}

// cleanIPTables 清理iptables规则
func (s *service) cleanIPTables() error {
	return s.iptablesService.UpdateRules(nil)
}

// iptablesService iptables服务实现
type iptablesService struct {
	rules []types.IPTablesRule
}

// newIPTablesService 创建iptables服务
func newIPTablesService() *iptablesService {
	return &iptablesService{
		rules: make([]types.IPTablesRule, 0),
	}
}

// UpdateRules 更新规则
func (s *iptablesService) UpdateRules(rules []types.IPTablesRule) error {
	// 在macOS上跳过iptables操作
	if runtime.GOOS == "darwin" {
		return nil
	}

	// 清理旧规则
	for _, rule := range s.rules {
		if err := s.deleteRule(rule); err != nil {
			return fmt.Errorf("清理规则失败: %v", err)
		}
	}

	// 添加新规则
	for _, rule := range rules {
		if err := s.addRule(rule); err != nil {
			return fmt.Errorf("添加规则失败: %v", err)
		}
	}

	s.rules = rules
	return nil
}

// GetRules 获取规则
func (s *iptablesService) GetRules() []types.IPTablesRule {
	return s.rules
}

// addRule 添加单条规则
func (s *iptablesService) addRule(rule types.IPTablesRule) error {
	args := s.buildArgs(rule)
	cmd := exec.Command("iptables", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("执行iptables命令失败: %w, 输出: %s", err, string(output))
	}
	return nil
}

// deleteRule 删除单条规则
func (s *iptablesService) deleteRule(rule types.IPTablesRule) error {
	args := append([]string{"-D"}, s.buildArgs(rule)[1:]...)
	cmd := exec.Command("iptables", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("执行iptables命令失败: %w, 输出: %s", err, string(output))
	}
	return nil
}

// buildArgs 构建iptables命令参数
func (s *iptablesService) buildArgs(rule types.IPTablesRule) []string {
	args := []string{"-A"}

	if rule.Table != "" {
		args = append(args, "-t", rule.Table)
	}

	if rule.Chain != "" {
		args = append(args, rule.Chain)
	}

	if rule.Interface != "" {
		args = append(args, "-i", rule.Interface)
	}

	if rule.Proto != "" {
		args = append(args, "-p", rule.Proto)
	}

	if rule.Source != "" {
		args = append(args, "-s", rule.Source)
	}

	if rule.Dest != "" {
		args = append(args, "-d", rule.Dest)
	}

	if rule.Sport != "" {
		args = append(args, "--sport", rule.Sport)
	}

	if rule.Dport != "" {
		args = append(args, "--dport", rule.Dport)
	}

	if rule.Target != "" {
		args = append(args, "-j", rule.Target)
	}

	if rule.ToPort != "" {
		args = append(args, "--to-port", rule.ToPort)
	}

	if rule.Mark != "" {
		args = append(args, "-m", "mark", "--mark", rule.Mark)
	}

	if rule.Comment != "" {
		args = append(args, "-m", "comment", "--comment", rule.Comment)
	}

	return args
}
