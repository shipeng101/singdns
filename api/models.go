package api

import (
	"fmt"
	"time"
)

// Settings represents system settings
type Settings struct {
	Theme           string `json:"theme"`         // light, dark, auto
	Language        string `json:"language"`      // zh-CN, en-US
	LoopbackPort    int    `json:"loopback_port"` // Local proxy port
	LoopbackEnabled bool   `json:"loopback_enabled"`
	DNSServer       string `json:"dns_server"`
	DNSMode         string `json:"dns_mode"` // ipv4, ipv6, dual
	PTDirectEnabled bool   `json:"pt_direct_enabled"`
	ProxyName       string `json:"proxy_name"`
}

// Validate validates settings
func (s *Settings) Validate() error {
	// Validate theme
	switch s.Theme {
	case "light", "dark", "auto":
	default:
		return fmt.Errorf("invalid theme: %s", s.Theme)
	}

	// Validate language
	switch s.Language {
	case "zh-CN", "en-US":
	default:
		return fmt.Errorf("invalid language: %s", s.Language)
	}

	// Validate port
	if s.LoopbackPort < 1024 || s.LoopbackPort > 65535 {
		return fmt.Errorf("invalid port: %d", s.LoopbackPort)
	}

	// Validate DNS mode
	switch s.DNSMode {
	case "ipv4", "ipv6", "dual":
	default:
		return fmt.Errorf("invalid DNS mode: %s", s.DNSMode)
	}

	return nil
}

// Node represents a proxy node
type Node struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`     // ss, vmess, trojan, vless, hy2, tuic
	Address   string    `json:"address"`  // Server address
	Port      int       `json:"port"`     // Server port
	Method    string    `json:"method"`   // Encryption method (for ss)
	Password  string    `json:"password"` // Password (for ss/trojan)
	UUID      string    `json:"uuid"`     // UUID (for vmess/vless/tuic)
	AlterId   int       `json:"alter_id"` // Alter ID (for vmess)
	Network   string    `json:"network"`  // Network type (tcp/ws/grpc)
	Path      string    `json:"path"`     // WebSocket path
	Host      string    `json:"host"`     // SNI
	TLS       bool      `json:"tls"`      // Enable TLS
	Traffic   *Traffic  `json:"traffic"`  // Traffic statistics
	CreatedAt time.Time `json:"created_at"`
}

// Rule represents a routing rule
type Rule struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`    // domain, ip, port, process
	Mode      string    `json:"mode"`    // direct, proxy, block
	Pattern   string    `json:"pattern"` // Rule pattern
	CreatedAt time.Time `json:"created_at"`
}

// Subscription represents a node subscription
type Subscription struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Type      string    `json:"type"`  // clash, singbox, ss, v2ray
	Nodes     []*Node   `json:"nodes"` // Parsed nodes
	UpdatedAt time.Time `json:"updated_at"`
}

// Traffic represents traffic statistics
type Traffic struct {
	Total    int64  `json:"total"`     // Total traffic in bytes
	Used     int64  `json:"used"`      // Used traffic in bytes
	Unit     string `json:"unit"`      // Traffic unit (B, KB, MB, GB, TB)
	ExpireAt string `json:"expire_at"` // Expiration time
}

// SystemStatus represents the system status information
type SystemStatus struct {
	Platform  string  `json:"platform"`
	Arch      string  `json:"arch"`
	CPU       float64 `json:"cpu"`
	Memory    float64 `json:"memory"`
	DiskTotal uint64  `json:"disk_total"`
	DiskUsed  uint64  `json:"disk_used"`
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Version    string    `json:"version"`
	StartTime  time.Time `json:"start_time"`
	ConfigPath string    `json:"config_path"`
	Error      string    `json:"error,omitempty"`
}

// TrafficStats represents traffic statistics
type TrafficStats struct {
	Time      time.Time `json:"time"`
	Upload    int64     `json:"upload"`     // Upload speed in bytes/s
	Download  int64     `json:"download"`   // Download speed in bytes/s
	TotalUp   int64     `json:"total_up"`   // Total upload in bytes
	TotalDown int64     `json:"total_down"` // Total download in bytes
}
