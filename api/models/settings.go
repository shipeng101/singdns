package models

import (
	"gorm.io/gorm"
)

// Settings represents system settings
type Settings struct {
	gorm.Model
	ID               string `json:"id" gorm:"primaryKey"`
	Theme            string `json:"theme"`
	Language         string `json:"language"`
	UpdateInterval   int    `json:"update_interval"` // hours
	ProxyPort        int    `json:"proxy_port"`
	APIPort          int    `json:"api_port"`
	DNSPort          int    `json:"dns_port"`
	LogLevel         string `json:"log_level"`
	EnableAutoUpdate bool   `json:"enable_auto_update"`
	UpdatedAt        int64  `json:"updated_at"`
}
