package models

import "time"

// RuleSet represents a remote rule set
type RuleSet struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	URL         string    `json:"url" gorm:"not null"`
	Type        string    `json:"type" gorm:"not null"`
	Format      string    `json:"format" gorm:"not null"`
	Outbound    string    `json:"outbound" gorm:"not null"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
