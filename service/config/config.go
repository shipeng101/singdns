package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/shipeng101/singdns/pkg/types"
	"gopkg.in/yaml.v2"
)

type service struct {
	config     *types.Config
	configPath string
	mu         sync.RWMutex
}

// NewService 创建配置服务
func NewService(configPath string) types.ConfigService {
	return &service{
		config:     types.NewDefaultConfig(),
		configPath: configPath,
	}
}

// Load 加载配置
func (s *service) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建配置目录
	configDir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.Save()
		}
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := yaml.Unmarshal(data, s.config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	return nil
}

// Save 保存配置
func (s *service) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建配置目录
	configDir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	data, err := yaml.Marshal(s.config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	return os.WriteFile(s.configPath, data, 0644)
}

// Get 获取配置
func (s *service) Get() *types.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// Update 更新配置
func (s *service) Update(config *types.Config) error {
	s.mu.Lock()
	s.config = config
	s.mu.Unlock()
	return s.Save()
}
