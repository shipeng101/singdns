package models

import (
	"encoding/json"
)

// Settings represents system settings
type Settings struct {
	ID               string `json:"id" gorm:"primaryKey"`
	Theme            string `json:"theme"`
	ThemeMode        string `json:"theme_mode"`
	Language         string `json:"language"`
	UpdateInterval   int    `json:"update_interval"` // hours
	ProxyPort        int    `json:"proxy_port"`
	APIPort          int    `json:"api_port"`
	DNSPort          int    `json:"dns_port"`
	LogLevel         string `json:"log_level"`
	EnableAutoUpdate bool   `json:"enable_auto_update"`
	UpdatedAt        int64  `json:"updated_at"`
	Dashboard        string `json:"dashboard" gorm:"type:json"`
	SingboxMode      string `json:"singbox_mode" gorm:"type:json"`
	InboundMode      string `json:"inbound_mode" gorm:"type:json"`
	CreatedAt        int64  `json:"created_at"`
}

// GetDashboard 获取仪表盘设置
func (s *Settings) GetDashboard() *DashboardSettings {
	if s.Dashboard == "" {
		// 如果仪表盘设置为空，设置默认值
		dashboard := &DashboardSettings{
			Type:   "yacd",
			Secret: "",
		}
		// 保存默认设置
		if err := s.SetDashboard(dashboard); err != nil {
			return dashboard
		}
		return dashboard
	}
	var dashboard DashboardSettings
	if err := json.Unmarshal([]byte(s.Dashboard), &dashboard); err != nil {
		// 如果解析失败，返回默认设置
		return &DashboardSettings{
			Type:   "yacd",
			Secret: "",
		}
	}
	return &dashboard
}

// SetDashboard 设置仪表盘
func (s *Settings) SetDashboard(dashboard *DashboardSettings) error {
	data, err := json.Marshal(dashboard)
	if err != nil {
		return err
	}
	s.Dashboard = string(data)
	return nil
}

// DashboardSettings 仪表盘设置
type DashboardSettings struct {
	Type   string `json:"type"`   // yacd 或 metacubexd
	Secret string `json:"secret"` // API 密钥
	Path   string `json:"path"`   // 面板路径
}

// GetSingboxMode 获取 SingBox 模式设置
func (s *Settings) GetSingboxMode() *SingboxModeSettings {
	if s.SingboxMode == "" {
		// 如果 SingBox 模式设置为空，设置默认值
		mode := &SingboxModeSettings{
			Mode: "rule",
		}
		// 保存默认设置
		if err := s.SetSingboxMode(mode); err != nil {
			return mode
		}
		return mode
	}
	var mode SingboxModeSettings
	if err := json.Unmarshal([]byte(s.SingboxMode), &mode); err != nil {
		// 如果解析失败，返回默认设置
		return &SingboxModeSettings{
			Mode: "rule",
		}
	}
	return &mode
}

// SetSingboxMode 设置 SingBox 模式
func (s *Settings) SetSingboxMode(mode *SingboxModeSettings) error {
	data, err := json.Marshal(mode)
	if err != nil {
		return err
	}
	s.SingboxMode = string(data)
	return nil
}

// SingboxModeSettings SingBox 模式设置
type SingboxModeSettings struct {
	Mode string `json:"mode"`
}

// InboundModeSettings 入站模式设置
type InboundModeSettings struct {
	Mode string `json:"mode"`
}

// GetInboundMode 获取入站模式
func (s *Settings) GetInboundMode() string {
	if s.InboundMode == "" {
		return "tun" // 默认使用 TUN 模式
	}
	var mode InboundModeSettings
	if err := json.Unmarshal([]byte(s.InboundMode), &mode); err != nil {
		return "tun" // 如果解析失败，返回默认值
	}
	return mode.Mode
}

// SetInboundMode 设置入站模式
func (s *Settings) SetInboundMode(mode string) error {
	modeSettings := &InboundModeSettings{
		Mode: mode,
	}
	data, err := json.Marshal(modeSettings)
	if err != nil {
		return err
	}
	s.InboundMode = string(data)
	return nil
}

// GetDNSSettings 获取 DNS 设置
func (s *Settings) GetDNSSettings() *DNSSettings {
	// 从存储中获取 DNS 设置
	dnsSettings := &DNSSettings{
		ID:               "default",
		Domestic:         "223.5.5.5",
		DomesticType:     "udp",
		SingboxDNS:       "https://dns.google/dns-query",
		SingboxDNSType:   "doh",
		EDNSClientSubnet: "1.0.1.0",
	}
	return dnsSettings
}

// SetDNSSettings 设置 DNS 设置
func (s *Settings) SetDNSSettings(settings *DNSSettings) error {
	// 验证设置
	if err := settings.Validate(); err != nil {
		return err
	}
	return nil
}
