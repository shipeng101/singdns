package models

import "time"

// NodeGroup represents a group of nodes
type NodeGroup struct {
	ID              string      `json:"id" gorm:"primaryKey"`
	Name            string      `json:"name" gorm:"not null"`
	Description     string      `json:"description"`
	Tag             string      `json:"tag"`
	Mode            string      `json:"mode" gorm:"default:select"`
	Active          bool        `json:"active" gorm:"default:true"`
	IncludePatterns StringArray `json:"include_patterns" gorm:"type:json"`
	ExcludePatterns StringArray `json:"exclude_patterns" gorm:"type:json"`
	Nodes           []Node      `json:"nodes" gorm:"many2many:node_group_nodes;"`
	NodeCount       int         `json:"node_count" gorm:"-"`
	CreatedAt       time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}
