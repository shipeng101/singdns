package protocols

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"singdns/api/models"
	"strings"
)

// Protocol defines the interface for proxy protocols
type Protocol interface {
	ParseURL(u *url.URL) (*models.Node, error)
	ToURL(node *models.Node) (string, error)
	Validate(node *models.Node) error
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
func ParseURL(rawURL string) (*models.Node, error) {
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
func ToURL(node *models.Node) (string, error) {
	protocol := Get(node.Type)
	if protocol == nil {
		return "", fmt.Errorf("unsupported protocol: %s", node.Type)
	}

	return protocol.ToURL(node)
}

// Validate validates a Node configuration
func Validate(node *models.Node) error {
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
