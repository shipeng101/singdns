package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// StringSlice is a custom type for storing string slices in SQLite
type StringSlice []string

// Scan implements the sql.Scanner interface
func (ss *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*ss = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal StringSlice value")
	}
	return json.Unmarshal(bytes, ss)
}

// Value implements the driver.Valuer interface
func (ss StringSlice) Value() (driver.Value, error) {
	if ss == nil {
		return nil, nil
	}
	return json.Marshal(ss)
}

// Node represents a proxy node
type Node struct {
	ID               string      `json:"id" gorm:"primaryKey"`
	Name             string      `json:"name" gorm:"not null"`
	Type             string      `json:"type" gorm:"not null"`
	Address          string      `json:"address" gorm:"not null"`
	Port             int         `json:"port" gorm:"not null"`
	Password         string      `json:"password"`
	UUID             string      `json:"uuid"`
	AlterID          int         `json:"alter_id"`
	Security         string      `json:"security"`
	Network          string      `json:"network"`
	Path             string      `json:"path"`
	Host             string      `json:"host"`
	TLS              bool        `json:"tls"`
	SNI              string      `json:"sni"`
	ALPN             StringSlice `json:"alpn" gorm:"type:text"`
	Fingerprint      string      `json:"fingerprint"`
	PublicKey        string      `json:"public_key"`
	ShortID          string      `json:"short_id"`
	SpiderX          string      `json:"spider_x"`
	Flow             string      `json:"flow"`
	Version          string      `json:"version"`
	Method           string      `json:"method"`
	SkipCertVerify   bool        `json:"skip_cert_verify"`
	CC               string      `json:"cc"`
	Up               string      `json:"up"`
	Down             string      `json:"down"`
	ServiceName      string      `json:"service_name"`
	Obfs             string      `json:"obfs"`
	UDP              bool        `json:"udp"`
	Plugin           string      `json:"plugin"`
	PluginOpts       string      `json:"plugin_opts"`
	Status           string      `json:"status" gorm:"not null;default:'offline'"`
	Latency          int64       `json:"latency" gorm:"not null;default:0"`
	SubscriptionID   string      `json:"subscription_id" gorm:"index"`
	SubscriptionName string      `json:"subscription_name" gorm:"-"`
	NodeGroups       []NodeGroup `json:"node_groups" gorm:"many2many:node_group_nodes;"`
	GroupIDs         []string    `json:"group_ids" gorm:"-"`
	CreatedAt        time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
	CheckedAt        time.Time   `json:"checked_at"`
}
