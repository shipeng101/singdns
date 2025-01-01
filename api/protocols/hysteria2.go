package protocols

import (
	"fmt"
	"net/url"
	"singdns/api/models"
)

// Hysteria2Protocol implements the Hysteria2 protocol
type Hysteria2Protocol struct{}

func init() {
	Register("hy2", &Hysteria2Protocol{})
}

// ParseURL parses a Hysteria2 URL into a Node
// Format: hy2://password@host:port?insecure=1&sni=example.com&up=100&down=500#name
func (p *Hysteria2Protocol) ParseURL(u *url.URL) (*models.Node, error) {
	node := &models.Node{
		Type: "hy2",
		Name: u.Fragment,
		TLS:  true, // Hysteria2 always uses TLS
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
	if sni := query.Get("sni"); sni != "" {
		node.Host = sni
	} else {
		node.Host = node.Address
	}

	// Allow insecure
	if query.Get("insecure") == "1" {
		node.SkipCertVerify = true
	}

	return node, nil
}

// ToURL converts a Node to a Hysteria2 URL
func (p *Hysteria2Protocol) ToURL(node *models.Node) (string, error) {
	if err := p.Validate(node); err != nil {
		return "", err
	}

	u := url.URL{
		Scheme:   "hy2",
		User:     url.User(node.Password),
		Host:     fmt.Sprintf("%s:%d", node.Address, node.Port),
		Fragment: node.Name,
	}

	// Add query parameters
	query := url.Values{}

	// SNI
	if node.Host != "" && node.Host != node.Address {
		query.Set("sni", node.Host)
	}

	// Skip cert verify
	if node.SkipCertVerify {
		query.Set("insecure", "1")
	}

	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	return u.String(), nil
}

// Validate validates a Hysteria2 node configuration
func (p *Hysteria2Protocol) Validate(node *models.Node) error {
	if node.Type != "hy2" {
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

	return nil
}

// isValidBandwidth checks if the bandwidth format is valid
func isValidBandwidth(_ string) bool {
	// TODO: Implement bandwidth format validation
	// Format: number + unit (mbps, gb)
	return true
}

// HandleConnection handles a Hysteria2 connection
func (p *Hysteria2Protocol) HandleConnection(_s *Session) error {
	// TODO: Implement Hysteria2 connection handling
	return fmt.Errorf("hysteria2 protocol not implemented")
}
