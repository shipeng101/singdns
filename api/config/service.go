package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"singdns/api/models"
	"singdns/api/storage"
)

// ConfigService 配置服务
type ConfigService struct {
	singboxGenerator *SingBoxGenerator
	configDir        string
	dashboardManager *DashboardManager
	storage          storage.Storage
}

// NewConfigService 创建配置服务
func NewConfigService(configDir string, db storage.Storage) *ConfigService {
	// 获取设置
	settings, err := db.GetSettings()
	if err != nil {
		fmt.Printf("Warning: failed to get settings: %v\n", err)
	}

	// 如果没有设置，使用默认设置
	if settings == nil {
		settings = &models.Settings{
			ID:               "default",
			Theme:            "light",
			ThemeMode:        "system",
			Language:         "zh-CN",
			UpdateInterval:   24,
			ProxyPort:        1080,
			APIPort:          8080,
			DNSPort:          53,
			LogLevel:         "info",
			EnableAutoUpdate: true,
			UpdatedAt:        time.Now().Unix(),
		}
		// 设置默认仪表盘
		settings.SetDashboard(&models.DashboardSettings{
			Type: "yacd",
		})
	}

	// 初始化仪表盘管理器
	dashboardManager := NewDashboardManager("bin/web")

	// 创建生成器
	singboxGenerator := NewSingBoxGenerator(db)

	service := &ConfigService{
		singboxGenerator: singboxGenerator,
		configDir:        configDir,
		dashboardManager: dashboardManager,
		storage:          db,
	}

	return service
}

// GenerateConfigs 生成所有配置文件
func (s *ConfigService) GenerateConfigs() error {
	// 生成 SingBox 配置
	singboxConfig, err := s.singboxGenerator.GenerateConfig()
	if err != nil {
		return fmt.Errorf("generate singbox config: %w", err)
	}

	// 格式化 JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, singboxConfig, "", "  "); err != nil {
		return fmt.Errorf("format json: %w", err)
	}

	// 确保配置目录存在
	if err := os.MkdirAll(s.configDir, 0755); err != nil {
		return fmt.Errorf("create singbox directory: %w", err)
	}

	// 写入 SingBox 配置
	singboxPath := filepath.Join(s.configDir, "config.json")
	if err := os.WriteFile(singboxPath, prettyJSON.Bytes(), 0644); err != nil {
		return fmt.Errorf("write singbox config: %w", err)
	}

	return nil
}

// UpdateSingBoxConfig 更新 SingBox 配置
func (s *ConfigService) UpdateSingBoxConfig(config *SingBoxConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := s.singboxGenerator.ValidateConfig(data); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}
	return s.GenerateConfigs()
}

// GetSingBoxConfig 获取 SingBox 配置
func (s *ConfigService) GetSingBoxConfig() *SingBoxConfig {
	if s.singboxGenerator == nil {
		return nil
	}
	data, err := s.singboxGenerator.GenerateConfig()
	if err != nil {
		return nil
	}
	var config SingBoxConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil
	}
	return &config
}

// AddOutbound 添加出站配置
func (s *ConfigService) AddOutbound(outbound OutboundConfig) error {
	config := s.GetSingBoxConfig()
	if config == nil {
		return fmt.Errorf("config not initialized")
	}
	config.Outbounds = append(config.Outbounds, outbound)
	return s.UpdateSingBoxConfig(config)
}

// UpdateDNSConfig 更新 DNS 配置
func (s *ConfigService) UpdateDNSConfig(servers []DNSServerConfig, upstream DNSServerConfig) error {
	config := s.GetSingBoxConfig()
	if config == nil {
		return fmt.Errorf("config not initialized")
	}
	if config.DNS == nil {
		config.DNS = &DNSConfig{}
	}
	config.DNS.Servers = servers
	return s.UpdateSingBoxConfig(config)
}

// AddInbound 添加入站配置
func (s *ConfigService) AddInbound(inbound InboundConfig) error {
	config := s.GetSingBoxConfig()
	if config == nil {
		return fmt.Errorf("config not initialized")
	}
	config.Inbounds = append(config.Inbounds, inbound)
	return s.UpdateSingBoxConfig(config)
}

// UpdateInbound 更新入站配置
func (s *ConfigService) UpdateInbound(tag string, inbound InboundConfig) error {
	config := s.GetSingBoxConfig()
	if config == nil {
		return fmt.Errorf("config not initialized")
	}
	for i, in := range config.Inbounds {
		if in.Tag == tag {
			config.Inbounds[i] = inbound
			return s.UpdateSingBoxConfig(config)
		}
	}
	return fmt.Errorf("inbound not found: %s", tag)
}

// DeleteInbound 删除入站配置
func (s *ConfigService) DeleteInbound(tag string) error {
	config := s.GetSingBoxConfig()
	if config == nil {
		return fmt.Errorf("config not initialized")
	}
	for i, in := range config.Inbounds {
		if in.Tag == tag {
			config.Inbounds = append(config.Inbounds[:i], config.Inbounds[i+1:]...)
			return s.UpdateSingBoxConfig(config)
		}
	}
	return fmt.Errorf("inbound not found: %s", tag)
}
