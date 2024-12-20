package models

import "time"

// Node represents a subscription node
type Node struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`     // ss, vmess, trojan, vless, hy2, tuic
	Address   string    `json:"address"`  // Server address
	Port      int       `json:"port"`     // Server port
	Method    string    `json:"method"`   // Encryption method (for ss)
	Password  string    `json:"password"` // Password (for ss/trojan/hy2)
	UUID      string    `json:"uuid"`     // UUID (for vmess/vless/tuic)
	AlterId   int       `json:"alter_id"` // Alter ID (for vmess)
	Network   string    `json:"network"`  // Network type (tcp/ws/grpc)
	Path      string    `json:"path"`     // WebSocket path or gRPC service name
	Host      string    `json:"host"`     // SNI or WebSocket host
	TLS       bool      `json:"tls"`      // Enable TLS
	Flow      string    `json:"flow"`     // VLESS flow control
	Up        string    `json:"up"`       // Hysteria2 uplink capacity
	Down      string    `json:"down"`     // Hysteria2 downlink capacity
	CC        string    `json:"cc"`       // TUIC congestion control
	CreatedAt time.Time `json:"created_at"`
}

// Subscription represents a node subscription
type Subscription struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // clash, singbox, ss, vmess, trojan
	URL       string    `json:"url"`
	Nodes     []*Node   `json:"nodes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
