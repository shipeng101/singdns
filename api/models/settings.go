package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

// Settings represents system settings
type Settings struct {
	gorm.Model
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
}

// GetDashboard 获取仪表盘设置
func (s *Settings) GetDashboard() *DashboardSettings {
	if s.Dashboard == "" {
		// 如果仪表盘设置为空，设置默认值
		dashboard := &DashboardSettings{
			Type:   "yacd",
			Path:   "bin/web/yacd",
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
			Path:   "bin/web/yacd",
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
	Type   string `json:"type"`
	Path   string `json:"path"`
	Secret string `json:"secret"`
}
