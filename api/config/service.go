package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"singdns/api/models"
	"singdns/api/storage"
)

// ConfigService 配置服务
type ConfigService struct {
	mosdnsGenerator  *MosDNSGenerator
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
			ProxyPort:        7890,
			APIPort:          8080,
			DNSPort:          53,
			LogLevel:         "info",
			EnableAutoUpdate: true,
			UpdatedAt:        time.Now().Unix(),
		}
		// 设置默认仪表盘
		settings.SetDashboard(&models.DashboardSettings{
			Type: "yacd",
			Path: "/web/public/yacd",
		})
	}

	// 初始化仪表盘管理器
	dashboardManager := NewDashboardManager("bin/web")

	// 创建生成器
	mosdnsGenerator := NewMosDNSGenerator()
	mosdnsGenerator.SetStorage(db)

	singboxGenerator := NewSingBoxGenerator(settings, db, filepath.Join(configDir, "config.json"))

	service := &ConfigService{
		mosdnsGenerator:  mosdnsGenerator,
		singboxGenerator: singboxGenerator,
		configDir:        configDir,
		dashboardManager: dashboardManager,
		storage:          db,
	}

	return service
}

// GenerateConfigs 生成所有配置文件
func (s *ConfigService) GenerateConfigs() error {
	// 获取最新设置
	settings, err := s.storage.GetSettings()
	if err != nil {
		return fmt.Errorf("get settings: %w", err)
	}

	// 更新生成器的设置
	s.singboxGenerator = NewSingBoxGenerator(settings, s.storage, filepath.Join(s.configDir, "config.json"))

	// 生成 MosDNS 配置
	mosdnsConfig, err := s.mosdnsGenerator.GenerateConfig()
	if err != nil {
		return fmt.Errorf("generate mosdns config: %w", err)
	}

	// 生成 SingBox 配置
	singboxConfig, err := s.singboxGenerator.GenerateConfig()
	if err != nil {
		return fmt.Errorf("generate singbox config: %w", err)
	}

	// 确保配置目录存在
	mosdnsDir := filepath.Join("configs", "mosdns")
	if err := os.MkdirAll(mosdnsDir, 0755); err != nil {
		return fmt.Errorf("create mosdns directory: %w", err)
	}
	if err := os.MkdirAll(s.configDir, 0755); err != nil {
		return fmt.Errorf("create singbox directory: %w", err)
	}

	// 写入 MosDNS 配置
	mosdnsPath := filepath.Join(mosdnsDir, "config.yaml")
	if err := os.WriteFile(mosdnsPath, mosdnsConfig, 0644); err != nil {
		return fmt.Errorf("write mosdns config: %w", err)
	}

	// 写入 SingBox 配置
	singboxPath := filepath.Join(s.configDir, "config.json")
	if err := os.WriteFile(singboxPath, singboxConfig, 0644); err != nil {
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

// RuleSetFileInfo 规则集文件信息
type RuleSetFileInfo struct {
	ID          string    `json:"id"`
	Tag         string    `json:"tag"`
	Description string    `json:"description"`
	FilePath    string    `json:"file_path"`
	UpdatedAt   time.Time `json:"updated_at"`
	Enabled     bool      `json:"enabled"`
}

// GetRuleSetsInfo 获取规则集信息
func (s *ConfigService) GetRuleSetsInfo() ([]RuleSetFileInfo, error) {
	// 扫描规则集目录
	dir := "configs/sing-box/rules"
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules directory: %v", err)
	}

	ruleSets := make([]RuleSetFileInfo, 0)
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".srs") {
			filePath := filepath.Join(dir, file.Name())
			info, err := file.Info()
			if err != nil {
				continue
			}

			// 从文件名中提取标签（去掉 .srs 后缀）
			tag := strings.TrimSuffix(file.Name(), ".srs")

			// 生成描述
			description := tag
			switch tag {
			case "geosite-cn":
				description = "中国域名列表"
			case "geosite-category-ads":
				description = "广告域名列表"
			case "geosite-youtube":
				description = "YouTube 域名列表"
			}

			// 获取规则集状态
			ruleSetInfo, err := s.storage.GetRuleSetByTag(tag)
			enabled := true
			id := tag
			if err == nil && ruleSetInfo != nil {
				enabled = ruleSetInfo.Enabled
				id = ruleSetInfo.ID
			}

			fileInfo := RuleSetFileInfo{
				ID:          id,
				Tag:         tag,
				Description: description,
				FilePath:    filePath,
				UpdatedAt:   info.ModTime(),
				Enabled:     enabled,
			}
			ruleSets = append(ruleSets, fileInfo)
		}
	}

	return ruleSets, nil
}

// UpdateRuleSets 更新规则集
func (s *ConfigService) UpdateRuleSets() error {
	if s.singboxGenerator == nil {
		return fmt.Errorf("singbox generator not initialized")
	}
	return s.singboxGenerator.UpdateRuleSets()
}

// SaveRuleSet 保存规则集
func (s *ConfigService) SaveRuleSet(ruleSet *models.RuleSet) error {
	if s.singboxGenerator == nil {
		return fmt.Errorf("singbox generator not initialized")
	}
	return s.singboxGenerator.SaveRuleSet(ruleSet)
}

// DeleteRuleSet 删除规则集
func (s *ConfigService) DeleteRuleSet(id string) error {
	if s.singboxGenerator == nil {
		return fmt.Errorf("singbox generator not initialized")
	}
	return s.singboxGenerator.DeleteRuleSet(id)
}
