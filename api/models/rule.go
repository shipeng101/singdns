package models

import (
	"fmt"
	"time"
)

// Rule 路由规则
type Rule struct {
	ID          string      `json:"id" gorm:"primaryKey"`
	Name        string      `json:"name" gorm:"not null"`
	Type        string      `json:"type" gorm:"not null"` // domain 或 ip
	Values      StringArray `json:"values" gorm:"type:json"`
	Outbound    string      `json:"outbound" gorm:"not null"`
	Description string      `json:"description"`
	Enabled     bool        `json:"enabled" gorm:"default:true"`
	Priority    int         `json:"priority" gorm:"default:0"`
	CreatedAt   time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// DNSRule DNS规则
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

// Validate 验证规则
func (r *Rule) Validate() error {
	if r.Type != "domain" && r.Type != "ip" {
		return fmt.Errorf("invalid rule type: %s", r.Type)
	}
	if r.Outbound == "" {
		return fmt.Errorf("outbound is required")
	}
	if len(r.Values) == 0 {
		return fmt.Errorf("rule values are required")
	}
	return nil
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
