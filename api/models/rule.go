package models

import "time"

// Rule represents a routing rule
type Rule struct {
	ID          string      `json:"id" gorm:"primaryKey"`
	Name        string      `json:"name" gorm:"not null"`
	Type        string      `json:"type" gorm:"not null"`
	Outbound    string      `json:"outbound" gorm:"not null"`
	Description string      `json:"description"`
	Enabled     bool        `json:"enabled" gorm:"default:true"`
	Priority    int         `json:"priority" gorm:"default:0"`
	Domains     StringArray `json:"domains" gorm:"type:json"`
	IPs         StringArray `json:"ips" gorm:"type:json"`
	CreatedAt   time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}
