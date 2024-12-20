package protocols

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Node represents a proxy node
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

// Protocol defines the interface for proxy protocols
type Protocol interface {
	ParseURL(u *url.URL) (*Node, error)
	ToURL(node *Node) (string, error)
	Validate(node *Node) error
}

// protocols maps protocol names to their implementations
var protocols = make(map[string]Protocol)

// Register registers a protocol implementation
func Register(name string, protocol Protocol) {
	protocols[name] = protocol
}

// Get returns a protocol implementation by name
func Get(name string) Protocol {
	return protocols[name]
}

// ParseURL parses a URL into a Node using the appropriate protocol
func ParseURL(rawURL string) (*Node, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	protocol := Get(u.Scheme)
	if protocol == nil {
		return nil, fmt.Errorf("unsupported protocol: %s", u.Scheme)
	}

	return protocol.ParseURL(u)
}

// ToURL converts a Node to its URL representation
func ToURL(node *Node) (string, error) {
	protocol := Get(node.Type)
	if protocol == nil {
		return "", fmt.Errorf("unsupported protocol: %s", node.Type)
	}

	return protocol.ToURL(node)
}

// Validate validates a Node configuration
func Validate(node *Node) error {
	protocol := Get(node.Type)
	if protocol == nil {
		return fmt.Errorf("unsupported protocol: %s", node.Type)
	}

	return protocol.Validate(node)
}

// decodeBase64 decodes a base64 string
func decodeBase64(s string) (string, error) {
	if i := len(s) % 4; i != 0 {
		s += strings.Repeat("=", 4-i)
	}
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(s)
	}
	return string(decoded), err
}

// encodeBase64 encodes a string to base64
func encodeBase64(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}
