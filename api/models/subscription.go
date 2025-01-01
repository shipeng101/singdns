package models

import (
	"fmt"
	"time"
)

// Subscription represents a subscription
type Subscription struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	Name           string    `json:"name" gorm:"not null"`
	URL            string    `json:"url" gorm:"not null"`
	Type           string    `json:"type" gorm:"not null"`
	NodeCount      int       `json:"node_count" gorm:"default:0"`
	Active         bool      `json:"active" gorm:"default:true"`
	AutoUpdate     bool      `json:"auto_update" gorm:"default:false"`
	UpdateInterval int64     `json:"update_interval" gorm:"default:3600"`
	LastUpdate     time.Time `json:"last_update"`
	ExpireTime     time.Time `json:"expire_time"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Validate validates the subscription
func (s *Subscription) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("name is required")
	}
	if s.URL == "" {
		return fmt.Errorf("url is required")
	}
	if s.Type == "" {
		return fmt.Errorf("type is required")
	}
	if s.UpdateInterval < 0 {
		return fmt.Errorf("update interval must be non-negative")
	}
	return nil
}
