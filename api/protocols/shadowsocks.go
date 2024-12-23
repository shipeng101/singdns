package protocols

import (
	"fmt"
	"net/url"
	"singdns/api/models"
	"strconv"
	"strings"
)

// ShadowsocksProtocol implements the Shadowsocks protocol
type ShadowsocksProtocol struct{}

// Supported encryption methods
var supportedMethods = map[string]bool{
	"aes-128-gcm":                   true,
	"aes-256-gcm":                   true,
	"chacha20-ietf-poly1305":        true,
	"2022-blake3-aes-128-gcm":       true,
	"2022-blake3-aes-256-gcm":       true,
	"2022-blake3-chacha20-poly1305": true,
}

func init() {
	Register("ss", &ShadowsocksProtocol{})
}

// ParseURL parses a Shadowsocks URL into a Node
// Format: ss://method:password@host:port#name
func (p *ShadowsocksProtocol) ParseURL(u *url.URL) (*models.Node, error) {
	node := &models.Node{
		Type: "ss",
		Name: u.Fragment,
	}

	// Parse userinfo (method:password)
	if u.User == nil {
		// Try base64 encoded format
		if parts := strings.SplitN(u.Host, "@", 2); len(parts) == 2 {
			decoded, err := decodeBase64(parts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid base64 encoding: %v", err)
			}
			methodParts := strings.SplitN(decoded, ":", 2)
			if len(methodParts) != 2 {
				return nil, fmt.Errorf("invalid userinfo format")
			}
			node.Method = methodParts[0]
			node.Password = methodParts[1]
			u.Host = parts[1]
		} else {
			return nil, fmt.Errorf("missing userinfo")
		}
	} else {
		password, _ := u.User.Password()
		node.Method = u.User.Username()
		node.Password = password
	}

	// Parse host and port
	if host, port, err := parseHostPort(u.Host); err != nil {
		return nil, err
	} else {
		node.Address = host
		node.Port = port
	}

	// Validate method
	if !supportedMethods[node.Method] {
		return nil, fmt.Errorf("unsupported encryption method: %s", node.Method)
	}

	// Parse plugin options
	if plugin := u.Query().Get("plugin"); plugin != "" {
		// TODO: Implement plugin support
	}

	return node, nil
}

// ToURL converts a Node to a Shadowsocks URL
func (p *ShadowsocksProtocol) ToURL(node *models.Node) (string, error) {
	if err := p.Validate(node); err != nil {
		return "", err
	}

	userinfo := fmt.Sprintf("%s:%s", node.Method, node.Password)
	encoded := encodeBase64(userinfo)

	u := url.URL{
		Scheme:   "ss",
		Host:     fmt.Sprintf("%s@%s:%d", encoded, node.Address, node.Port),
		Fragment: node.Name,
	}

	// Add plugin if configured
	// TODO: Implement plugin support

	return u.String(), nil
}

// Validate validates a Shadowsocks node configuration
func (p *ShadowsocksProtocol) Validate(node *models.Node) error {
	if node.Type != "ss" {
		return fmt.Errorf("invalid node type: %s", node.Type)
	}

	if node.Address == "" {
		return fmt.Errorf("missing address")
	}

	if node.Port <= 0 || node.Port > 65535 {
		return fmt.Errorf("invalid port: %d", node.Port)
	}

	if !supportedMethods[node.Method] {
		return fmt.Errorf("unsupported encryption method: %s", node.Method)
	}

	if node.Password == "" {
		return fmt.Errorf("missing password")
	}

	return nil
}

// parseHostPort parses a host:port string
func parseHostPort(s string) (string, int, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid host:port format")
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port: %v", err)
	}

	return parts[0], port, nil
}
