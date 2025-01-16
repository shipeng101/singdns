package protocols

import (
	"fmt"
	"net/url"
	"singdns/api/models"
)

// TUICProtocol implements the TUIC protocol
type TUICProtocol struct{}

func init() {
	Register("tuic", &TUICProtocol{})
}

// ParseURL parses a TUIC URL into a Node
// Format: tuic://uuid:password@host:port?congestion_control=bbr&alpn=h3&sni=example.com#name
func (p *TUICProtocol) ParseURL(u *url.URL) (*models.Node, error) {
	node := &models.Node{
		Type: "tuic",
		Name: u.Fragment,
		TLS:  true, // TUIC always uses TLS
	}

	// Parse UUID and password
	if u.User == nil {
		return nil, fmt.Errorf("missing UUID and password")
	}
	node.UUID = u.User.Username()
	if password, ok := u.User.Password(); !ok {
		return nil, fmt.Errorf("missing password")
	} else {
		node.Password = password
	}

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
	if sni := query.Get("sni"); sni != "" {
		node.Host = sni
	} else {
		node.Host = node.Address
	}

	// Congestion control
	if cc := query.Get("congestion_control"); cc != "" {
		node.CC = cc
	}

	// ALPN
	if alpn := query.Get("alpn"); alpn != "" {
		node.ALPN = models.StringSlice{alpn}
	}

	return node, nil
}

// ToURL converts a Node to a TUIC URL
func (p *TUICProtocol) ToURL(node *models.Node) (string, error) {
	if err := p.Validate(node); err != nil {
		return "", err
	}

	u := url.URL{
		Scheme:   "tuic",
		User:     url.UserPassword(node.UUID, node.Password),
		Host:     fmt.Sprintf("%s:%d", node.Address, node.Port),
		Fragment: node.Name,
	}

	// Add query parameters
	query := url.Values{}

	// SNI
	if node.Host != "" && node.Host != node.Address {
		query.Set("sni", node.Host)
	}

	// Congestion control
	if node.CC != "" {
		query.Set("congestion_control", node.CC)
	}

	// ALPN
	if len(node.ALPN) > 0 {
		query.Set("alpn", node.ALPN[0])
	}

	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	return u.String(), nil
}

// Validate validates a TUIC node configuration
func (p *TUICProtocol) Validate(node *models.Node) error {
	if node.Type != "tuic" {
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

	if node.Password == "" {
		return fmt.Errorf("missing password")
	}

	if node.CC != "" {
		switch node.CC {
		case "cubic", "bbr", "new_reno":
		default:
			return fmt.Errorf("unsupported congestion control: %s", node.CC)
		}
	}

	return nil
}
