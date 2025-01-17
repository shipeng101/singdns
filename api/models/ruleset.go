package models

import "time"

// RuleSet represents a remote rule set
type RuleSet struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	URL         string    `json:"url" gorm:"not null"`
	Type        string    `json:"type" gorm:"not null"`
	Format      string    `json:"format" gorm:"not null"`
	Path        string    `json:"path" gorm:"not null"`
	Outbound    string    `json:"outbound" gorm:"not null"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DefaultCloudflareRuleSet 返回默认的 Cloudflare 规则集
func DefaultCloudflareRuleSet() *RuleSet {
	return &RuleSet{
		ID:          "geoip-cloudflare",
		Name:        "Cloudflare IP",
		URL:         "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geoip/cloudflare.srs",
		Type:        "remote",
		Format:      "binary",
		Path:        "configs/sing-box/rules/geoip-cloudflare.srs",
		Outbound:    "direct",
		Description: "Cloudflare IP 规则集",
		Enabled:     true,
	}
}
