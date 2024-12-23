package protocols

import (
	"encoding/json"
	"fmt"
	"net/url"
	"singdns/api/models"
)

// VMessProtocol implements the VMess protocol
type VMessProtocol struct{}

// VMessConfig represents VMess configuration
type VMessConfig struct {
	Version string `json:"v"`
	Name    string `json:"ps"`
	Address string `json:"add"`
	Port    int    `json:"port"`
	UUID    string `json:"id"`
	AlterID int    `json:"aid"`
	Network string `json:"net"`
	Type    string `json:"type"`
	Host    string `json:"host"`
	Path    string `json:"path"`
	TLS     string `json:"tls"`
}

func init() {
	Register("vmess", &VMessProtocol{})
}

// ParseURL parses a VMess URL into a Node
// Format: vmess://base64(json)
func (p *VMessProtocol) ParseURL(u *url.URL) (*models.Node, error) {
	// Remove "vmess://" prefix
	encoded := u.String()[8:]

	// Decode base64
	decoded, err := decodeBase64(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 encoding: %v", err)
	}

	// Parse JSON
	var config VMessConfig
	if err := json.Unmarshal([]byte(decoded), &config); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	// Create node
	node := &models.Node{
		Type:    "vmess",
		Name:    config.Name,
		Address: config.Address,
		Port:    config.Port,
		UUID:    config.UUID,
		AlterID: config.AlterID,
		Network: config.Network,
		Path:    config.Path,
		Host:    config.Host,
		TLS:     config.TLS == "tls",
	}

	return node, nil
}

// ToURL converts a Node to a VMess URL
func (p *VMessProtocol) ToURL(node *models.Node) (string, error) {
	if err := p.Validate(node); err != nil {
		return "", err
	}

	// Create config
	config := VMessConfig{
		Version: "2",
		Name:    node.Name,
		Address: node.Address,
		Port:    node.Port,
		UUID:    node.UUID,
		AlterID: node.AlterID,
		Network: node.Network,
		Type:    "none",
		Host:    node.Host,
		Path:    node.Path,
	}

	if node.TLS {
		config.TLS = "tls"
	}

	// Convert to JSON
	data, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	// Encode as base64
	return "vmess://" + encodeBase64(string(data)), nil
}

// Validate validates a VMess node configuration
func (p *VMessProtocol) Validate(node *models.Node) error {
	if node.Type != "vmess" {
		return fmt.Errorf("invalid node type: %s", node.Type)
	}

	if node.Address == "" {
		return fmt.Errorf("missing address")
	}

	if node.Port <= 0 || node.Port > 65535 {
		return fmt.Errorf("invalid port: %d", node.Port)
	}

	if node.UUID == "" {
		return fmt.Errorf("missing UUID")
	}

	if node.AlterID < 0 {
		return fmt.Errorf("invalid alter ID: %d", node.AlterID)
	}

	switch node.Network {
	case "", "tcp", "ws", "http", "h2", "quic", "grpc":
	default:
		return fmt.Errorf("unsupported network: %s", node.Network)
	}

	return nil
}
