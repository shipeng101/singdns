package models

import (
	"fmt"
	"time"
)

// Rule represents a routing rule
type Rule struct {
	ID          string      `json:"id" gorm:"primaryKey"`
	Name        string      `json:"name" gorm:"not null"`
	Type        string      `json:"type" gorm:"not null"`
	Domains     StringArray `json:"domains" gorm:"type:json"`
	IPs         StringArray `json:"ips" gorm:"type:json"`
	Outbound    string      `json:"outbound" gorm:"not null"`
	Description string      `json:"description"`
	Enabled     bool        `json:"enabled" gorm:"default:true"`
	Priority    int         `json:"priority" gorm:"default:0"`
	CreatedAt   time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// DNSRule DNS 分流规则
type DNSRule struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Type        string    `json:"type" gorm:"not null"` // domain 或 ip
	Value       string    `json:"value" gorm:"not null"`
	Action      string    `json:"action" gorm:"not null"` // direct, remote 或 block
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DNSSettings DNS 设置
type DNSSettings struct {
	ID               string `json:"id" gorm:"primaryKey"`
	EnableAdBlock    bool   `json:"enable_ad_block" gorm:"default:true"`
	AdBlockRuleCount int    `json:"ad_block_rule_count" gorm:"default:0"`
	LastUpdate       int64  `json:"last_update" gorm:"default:0"`
}

// Validate 验证规则
func (r *DNSRule) Validate() error {
	if r.Type != "domain" && r.Type != "ip" {
		return fmt.Errorf("invalid rule type: %s", r.Type)
	}
	if r.Action != "direct" && r.Action != "remote" && r.Action != "block" {
		return fmt.Errorf("invalid rule action: %s", r.Action)
	}
	if r.Value == "" {
		return fmt.Errorf("rule value is required")
	}
	return nil
}
