package models

import (
	"regexp"
	"strings"
	"time"
)

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
	NodeCount       int         `json:"node_count" gorm:"default:0"`
	CreatedAt       time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// MatchNode checks if a node matches the group's patterns
func (g *NodeGroup) MatchNode(node *Node) bool {
	// "全部"分组特殊处理
	if g.Name == "全部 🌏" {
		return true
	}

	// If there are include patterns, the node must match at least one
	if len(g.IncludePatterns) > 0 {
		matched := false
		for _, pattern := range g.IncludePatterns {
			pattern = strings.TrimSpace(pattern)
			if pattern == "" {
				continue
			}

			// Try to compile as regex first
			re, err := regexp.Compile(pattern)
			if err == nil {
				// If it's a valid regex, use regex matching
				if re.MatchString(node.Name) {
					matched = true
					break
				}
			} else {
				// If not a valid regex, use keyword matching
				if strings.Contains(strings.ToLower(node.Name), strings.ToLower(pattern)) {
					matched = true
					break
				}
			}
		}
		if !matched {
			return false
		}
	}

	// If there are exclude patterns, the node must not match any
	if len(g.ExcludePatterns) > 0 {
		for _, pattern := range g.ExcludePatterns {
			pattern = strings.TrimSpace(pattern)
			if pattern == "" {
				continue
			}

			// Try to compile as regex first
			re, err := regexp.Compile(pattern)
			if err == nil {
				// If it's a valid regex, use regex matching
				if re.MatchString(node.Name) {
					return false
				}
			} else {
				// If not a valid regex, use keyword matching
				if strings.Contains(strings.ToLower(node.Name), strings.ToLower(pattern)) {
					return false
				}
			}
		}
	}

	return true
}
