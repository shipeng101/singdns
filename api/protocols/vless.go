package protocols

import (
	"fmt"
	"net/url"
	"singdns/api/models"
)

// VLESSProtocol implements the VLESS protocol
type VLESSProtocol struct{}

func init() {
	Register("vless", &VLESSProtocol{})
}

// ParseURL parses a VLESS URL into a Node
// Format: vless://uuid@host:port?type=tcp&security=tls&flow=xtls-rprx-vision&sni=example.com#name
func (p *VLESSProtocol) ParseURL(u *url.URL) (*models.Node, error) {
	node := &models.Node{
		Type: "vless",
		Name: u.Fragment,
	}

	// Parse UUID
	if u.User == nil {
		return nil, fmt.Errorf("missing UUID")
	}
	node.UUID = u.User.Username()

	// Parse host and port
	if host, port, err := parseHostPort(u.Host); err != nil {
		return nil, err
	} else {
		node.Address = host
		node.Port = port
	}

	// Parse query parameters
	query := u.Query()

	// Network type
	if network := query.Get("type"); network != "" {
		node.Network = network

		switch network {
		case "ws":
			node.Path = query.Get("path")
			node.Host = query.Get("host")
		case "grpc":
			node.Path = query.Get("serviceName")
		}
	}

	// TLS
	if security := query.Get("security"); security == "tls" {
		node.TLS = true
		if sni := query.Get("sni"); sni != "" {
			node.Host = sni
		}
	}

	// Flow control
	if flow := query.Get("flow"); flow != "" {
		node.Flow = flow
	}

	return node, nil
}

// ToURL converts a Node to a VLESS URL
func (p *VLESSProtocol) ToURL(node *models.Node) (string, error) {
	if err := p.Validate(node); err != nil {
		return "", err
	}

	u := url.URL{
		Scheme:   "vless",
		User:     url.User(node.UUID),
		Host:     fmt.Sprintf("%s:%d", node.Address, node.Port),
		Fragment: node.Name,
	}

	// Add query parameters
	query := url.Values{}

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

	// TLS
	if node.TLS {
		query.Set("security", "tls")
		if node.Host != "" {
			query.Set("sni", node.Host)
		}
	}

	// Flow control
	if node.Flow != "" {
		query.Set("flow", node.Flow)
	}

	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	return u.String(), nil
}

// Validate validates a VLESS node configuration
func (p *VLESSProtocol) Validate(node *models.Node) error {
	if node.Type != "vless" {
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

	if node.Network != "" {
		switch node.Network {
		case "tcp", "ws", "grpc":
		default:
			return fmt.Errorf("unsupported network: %s", node.Network)
		}
	}

	if node.Flow != "" {
		switch node.Flow {
		case "xtls-rprx-vision", "xtls-rprx-vision-udp443":
		default:
			return fmt.Errorf("unsupported flow: %s", node.Flow)
		}
	}

	return nil
}
