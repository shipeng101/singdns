package protocols

import (
	"fmt"
	"net/url"
	"singdns/api/models"
)

// TrojanProtocol implements the Trojan protocol
type TrojanProtocol struct{}

func init() {
	Register("trojan", &TrojanProtocol{})
}

// ParseURL parses a Trojan URL into a Node
// Format: trojan://password@host:port?allowInsecure=1&peer=sni#name
func (p *TrojanProtocol) ParseURL(u *url.URL) (*models.Node, error) {
	node := &models.Node{
		Type: "trojan",
		Name: u.Fragment,
		TLS:  true, // Trojan always uses TLS
	}

	// Parse password
	if u.User == nil {
		return nil, fmt.Errorf("missing password")
	}
	node.Password = u.User.Username()

	// Parse host and port
	if host, port, err := parseHostPort(u.Host); err != nil {
		return nil, err
	} else {
		node.Address = host
		node.Port = port
	}

	// Parse query parameters
	query := u.Query()

	// SNI
	if sni := query.Get("peer"); sni != "" {
		node.Host = sni
	} else {
		node.Host = node.Address
	}

	// Allow insecure
	if query.Get("allowInsecure") == "1" {
		node.SkipCertVerify = true
	}

	// Parse other parameters
	if network := query.Get("type"); network != "" {
		node.Network = network

		switch network {
		case "ws":
			node.Path = query.Get("path")
			if host := query.Get("host"); host != "" {
				node.Host = host
			}
		case "grpc":
			node.Path = query.Get("serviceName")
		}
	}

	return node, nil
}

// ToURL converts a Node to a Trojan URL
func (p *TrojanProtocol) ToURL(node *models.Node) (string, error) {
	if err := p.Validate(node); err != nil {
		return "", err
	}

	u := url.URL{
		Scheme:   "trojan",
		User:     url.User(node.Password),
		Host:     fmt.Sprintf("%s:%d", node.Address, node.Port),
		Fragment: node.Name,
	}

	// Add query parameters
	query := url.Values{}

	// SNI
	if node.Host != "" && node.Host != node.Address {
		query.Set("peer", node.Host)
	}

	// Skip cert verify
	if node.SkipCertVerify {
		query.Set("allowInsecure", "1")
	}

	// Network specific parameters
	if node.Network != "" {
		query.Set("type", node.Network)

		switch node.Network {
		case "ws":
			if node.Path != "" {
				query.Set("path", node.Path)
			}
			if node.Host != "" {
				query.Set("host", node.Host)
			}
		case "grpc":
			if node.Path != "" {
				query.Set("serviceName", node.Path)
			}
		}
	}

	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	return u.String(), nil
}

// Validate validates a Trojan node configuration
func (p *TrojanProtocol) Validate(node *models.Node) error {
	if node.Type != "trojan" {
		return fmt.Errorf("invalid node type: %s", node.Type)
	}

	if node.Address == "" {
		return fmt.Errorf("missing address")
	}

	if node.Port <= 0 || node.Port > 65535 {
		return fmt.Errorf("invalid port: %d", node.Port)
	}

	if node.Password == "" {
		return fmt.Errorf("missing password")
	}

	if node.Network != "" {
		switch node.Network {
		case "tcp", "ws", "grpc":
		default:
			return fmt.Errorf("unsupported network: %s", node.Network)
		}
	}

	return nil
}
