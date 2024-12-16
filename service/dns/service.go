package dns

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/shipeng101/singdns/pkg/types"
	"github.com/shipeng101/singdns/service/system"
	"gopkg.in/yaml.v2"
)

type service struct {
	config  types.DNSConfig
	cmd     *exec.Cmd
	mu      sync.Mutex
	running bool
}

// NewService 创建DNS服务
func NewService() types.DNSService {
	return &service{
		config: types.DNSConfig{
			Listen:   "127.0.0.1",
			Port:     5354,
			Cache:    true,
			Upstream: []string{"udp://8.8.8.8:53", "udp://8.8.4.4:53"},
			ChinaDNS: []string{"udp://223.5.5.5:53", "udp://119.29.29.29:53"},
		},
	}
}

// generateConfig 生成 mosdns 配置
func (s *service) generateConfig() error {
	// mosdns v5.x.x 配置格式
	config := map[string]interface{}{
		"log": map[string]interface{}{
			"level": "info",
		},
		"plugins": []map[string]interface{}{
			{
				"tag":  "forward_local",
				"type": "forward",
				"args": map[string]interface{}{
					"concurrent": 2,
					"upstreams": []map[string]interface{}{
						{
							"addr":            "udp://223.5.5.5:53",
							"enable_pipeline": true,
						},
						{
							"addr":            "udp://119.29.29.29:53",
							"enable_pipeline": true,
						},
					},
				},
			},
			{
				"tag":  "udp_server",
				"type": "udp_server",
				"args": map[string]interface{}{
					"entry":  "forward_local",
					"listen": fmt.Sprintf("%s:%d", s.config.Listen, s.config.Port),
				},
			},
			{
				"tag":  "tcp_server",
				"type": "tcp_server",
				"args": map[string]interface{}{
					"entry":  "forward_local",
					"listen": fmt.Sprintf("%s:%d", s.config.Listen, s.config.Port),
				},
			},
		},
	}

	// 将配置写入文件
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile("config/mosdns.yaml", data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// Start 启动服务
func (s *service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	// 生成配置文件
	if err := s.generateConfig(); err != nil {
		return err
	}

	// 获取 mosdns 路径
	mosdnsPath := "bin/mosdns"
	if _, err := os.Stat(mosdnsPath); err != nil {
		return fmt.Errorf("mosdns not found: %v", err)
	}

	// 检查配置文件
	configPath := "config/mosdns.yaml"
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("配置文件不存在: %v", err)
	}

	// 启动 mosdns
	cmd := exec.Command(mosdnsPath, "start", "-c", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 mosdns 失败: %v", err)
	}

	s.cmd = cmd
	s.running = true

	system.Info("DNS服务已启动")
	return nil
}

// Stop 停止服务
func (s *service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.cmd != nil && s.cmd.Process != nil {
		if err := s.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("停止 mosdns 失败: %v", err)
		}
		s.cmd.Wait()
	}

	s.cmd = nil
	s.running = false

	system.Info("DNS服务已停止")
	return nil
}

// GetConfig 获取配置
func (s *service) GetConfig() types.DNSConfig {
	return s.config
}

// UpdateConfig 更新配置
func (s *service) UpdateConfig(config types.DNSConfig) error {
	s.config = config

	// 如果服务正在运行，重启服务以应用新配置
	if s.running {
		if err := s.Stop(); err != nil {
			return err
		}
		return s.Start()
	}

	return nil
}

// GetRules 获取规则
func (s *service) GetRules() []types.Rule {
	return nil
}

// UpdateRules 更新规则
func (s *service) UpdateRules(rules []types.Rule) error {
	return nil
}

// GetStatus 获取状态
func (s *service) GetStatus() types.ServiceStatus {
	return types.ServiceStatus{
		Running: s.running,
	}
}
