package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// Protocol represents a proxy protocol
type Protocol interface {
	// Name returns the protocol name
	Name() string

	// ParseURL parses a URL into a Node
	ParseURL(u *url.URL) (*Node, error)

	// ToURL converts a Node to a URL string
	ToURL(node *Node) (string, error)

	// ToConfig converts a Node to protocol-specific configuration
	ToConfig(node *Node) (map[string]interface{}, error)
}

// Protocols is a map of supported protocols
var Protocols = map[string]Protocol{
	"ss":     &ShadowsocksProtocol{},
	"vmess":  &VMessProtocol{},
	"trojan": &TrojanProtocol{},
	"vless":  &VLESSProtocol{},
	"hy2":    &Hysteria2Protocol{},
	"tuic":   &TUICProtocol{},
}

// ShadowsocksProtocol implements the Shadowsocks protocol
type ShadowsocksProtocol struct{}

func (p *ShadowsocksProtocol) Name() string { return "ss" }

func (p *ShadowsocksProtocol) ParseURL(u *url.URL) (*Node, error) {
	if u.Scheme != "ss" {
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}

	// Parse userinfo (method:password)
	method, password, ok := strings.Cut(u.User.String(), ":")
	if !ok {
		return nil, fmt.Errorf("invalid userinfo format")
	}

	// Parse host and port
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid host: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	return &Node{
		Type:     "ss",
		Address:  host,
		Port:     port,
		Method:   method,
		Password: password,
	}, nil
}

func (p *ShadowsocksProtocol) ToURL(node *Node) (string, error) {
	if node.Type != "ss" {
		return "", fmt.Errorf("invalid node type: %s", node.Type)
	}

	userinfo := fmt.Sprintf("%s:%s", node.Method, node.Password)
	return fmt.Sprintf("ss://%s@%s:%d", base64.URLEncoding.EncodeToString([]byte(userinfo)), node.Address, node.Port), nil
}

func (p *ShadowsocksProtocol) ToConfig(node *Node) (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":     "shadowsocks",
		"address":  node.Address,
		"port":     node.Port,
		"method":   node.Method,
		"password": node.Password,
	}, nil
}

// VMessProtocol implements the VMess protocol
type VMessProtocol struct{}

func (p *VMessProtocol) Name() string { return "vmess" }

func (p *VMessProtocol) ParseURL(u *url.URL) (*Node, error) {
	if u.Scheme != "vmess" {
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}

	// Parse base64-encoded configuration
	data, err := base64.RawURLEncoding.DecodeString(u.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 encoding: %v", err)
	}

	var config struct {
		Version string `json:"v"`
		PS      string `json:"ps"`
		Add     string `json:"add"`
		Port    int    `json:"port"`
		ID      string `json:"id"`
		Aid     int    `json:"aid"`
		Net     string `json:"net"`
		Type    string `json:"type"`
		Host    string `json:"host"`
		Path    string `json:"path"`
		TLS     string `json:"tls"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	return &Node{
		Type:    "vmess",
		Name:    config.PS,
		Address: config.Add,
		Port:    config.Port,
		UUID:    config.ID,
		AlterId: config.Aid,
		Network: config.Net,
		Path:    config.Path,
		Host:    config.Host,
		TLS:     config.TLS == "tls",
	}, nil
}

func (p *VMessProtocol) ToURL(node *Node) (string, error) {
	if node.Type != "vmess" {
		return "", fmt.Errorf("invalid node type: %s", node.Type)
	}

	config := struct {
		Version string `json:"v"`
		PS      string `json:"ps"`
		Add     string `json:"add"`
		Port    int    `json:"port"`
		ID      string `json:"id"`
		Aid     int    `json:"aid"`
		Net     string `json:"net"`
		Type    string `json:"type"`
		Host    string `json:"host"`
		Path    string `json:"path"`
		TLS     string `json:"tls"`
	}{
		Version: "2",
		PS:      node.Name,
		Add:     node.Address,
		Port:    node.Port,
		ID:      node.UUID,
		Aid:     node.AlterId,
		Net:     node.Network,
		Type:    "none",
		Host:    node.Host,
		Path:    node.Path,
		TLS:     map[bool]string{true: "tls", false: ""}[node.TLS],
	}

	data, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %v", err)
	}

	return fmt.Sprintf("vmess://%s", base64.RawURLEncoding.EncodeToString(data)), nil
}

func (p *VMessProtocol) ToConfig(node *Node) (map[string]interface{}, error) {
	config := map[string]interface{}{
		"type":    "vmess",
		"address": node.Address,
		"port":    node.Port,
		"uuid":    node.UUID,
		"alterId": node.AlterId,
	}

	if node.Network != "" {
		config["network"] = node.Network
	}
	if node.Path != "" {
		config["path"] = node.Path
	}
	if node.Host != "" {
		config["host"] = node.Host
	}
	if node.TLS {
		config["tls"] = true
	}

	return config, nil
}

// TrojanProtocol implements the Trojan protocol
type TrojanProtocol struct{}

func (p *TrojanProtocol) Name() string { return "trojan" }

func (p *TrojanProtocol) ParseURL(u *url.URL) (*Node, error) {
	if u.Scheme != "trojan" {
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}

	password := u.User.String()
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid host: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	return &Node{
		Type:     "trojan",
		Address:  host,
		Port:     port,
		Password: password,
		TLS:      true, // Trojan always uses TLS
	}, nil
}

func (p *TrojanProtocol) ToURL(node *Node) (string, error) {
	if node.Type != "trojan" {
		return "", fmt.Errorf("invalid node type: %s", node.Type)
	}

	return fmt.Sprintf("trojan://%s@%s:%d", node.Password, node.Address, node.Port), nil
}

func (p *TrojanProtocol) ToConfig(node *Node) (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":     "trojan",
		"address":  node.Address,
		"port":     node.Port,
		"password": node.Password,
		"tls":      true,
	}, nil
}

// VLESSProtocol implements the VLESS protocol
type VLESSProtocol struct{}

func (p *VLESSProtocol) Name() string { return "vless" }

func (p *VLESSProtocol) ParseURL(u *url.URL) (*Node, error) {
	if u.Scheme != "vless" {
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}

	uuid := u.User.String()
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid host: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	query := u.Query()
	return &Node{
		Type:    "vless",
		Address: host,
		Port:    port,
		UUID:    uuid,
		Network: query.Get("type"),
		Path:    query.Get("path"),
		Host:    query.Get("host"),
		TLS:     query.Get("security") == "tls",
	}, nil
}

func (p *VLESSProtocol) ToURL(node *Node) (string, error) {
	if node.Type != "vless" {
		return "", fmt.Errorf("invalid node type: %s", node.Type)
	}

	query := url.Values{}
	if node.Network != "" {
		query.Set("type", node.Network)
	}
	if node.Path != "" {
		query.Set("path", node.Path)
	}
	if node.Host != "" {
		query.Set("host", node.Host)
	}
	if node.TLS {
		query.Set("security", "tls")
	}

	u := url.URL{
		Scheme:   "vless",
		User:     url.User(node.UUID),
		Host:     fmt.Sprintf("%s:%d", node.Address, node.Port),
		RawQuery: query.Encode(),
	}
	return u.String(), nil
}

func (p *VLESSProtocol) ToConfig(node *Node) (map[string]interface{}, error) {
	config := map[string]interface{}{
		"type":    "vless",
		"address": node.Address,
		"port":    node.Port,
		"uuid":    node.UUID,
	}

	if node.Network != "" {
		config["network"] = node.Network
	}
	if node.Path != "" {
		config["path"] = node.Path
	}
	if node.Host != "" {
		config["host"] = node.Host
	}
	if node.TLS {
		config["tls"] = true
	}

	return config, nil
}

// Hysteria2Protocol implements the Hysteria 2 protocol
type Hysteria2Protocol struct{}

func (p *Hysteria2Protocol) Name() string { return "hy2" }

func (p *Hysteria2Protocol) ParseURL(u *url.URL) (*Node, error) {
	if u.Scheme != "hy2" {
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}

	password := u.User.String()
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid host: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	return &Node{
		Type:     "hy2",
		Address:  host,
		Port:     port,
		Password: password,
		TLS:      true, // Hysteria 2 always uses TLS
	}, nil
}

func (p *Hysteria2Protocol) ToURL(node *Node) (string, error) {
	if node.Type != "hy2" {
		return "", fmt.Errorf("invalid node type: %s", node.Type)
	}

	return fmt.Sprintf("hy2://%s@%s:%d", node.Password, node.Address, node.Port), nil
}

func (p *Hysteria2Protocol) ToConfig(node *Node) (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":     "hysteria2",
		"address":  node.Address,
		"port":     node.Port,
		"password": node.Password,
		"tls":      true,
	}, nil
}

// TUICProtocol implements the TUIC protocol
type TUICProtocol struct{}

func (p *TUICProtocol) Name() string { return "tuic" }

func (p *TUICProtocol) ParseURL(u *url.URL) (*Node, error) {
	if u.Scheme != "tuic" {
		return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
	}

	uuid := u.User.String()
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid host: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %v", err)
	}

	query := u.Query()
	return &Node{
		Type:     "tuic",
		Address:  host,
		Port:     port,
		UUID:     uuid,
		Password: query.Get("password"),
		TLS:      true, // TUIC always uses TLS
	}, nil
}

func (p *TUICProtocol) ToURL(node *Node) (string, error) {
	if node.Type != "tuic" {
		return "", fmt.Errorf("invalid node type: %s", node.Type)
	}

	query := url.Values{}
	if node.Password != "" {
		query.Set("password", node.Password)
	}

	u := url.URL{
		Scheme:   "tuic",
		User:     url.User(node.UUID),
		Host:     fmt.Sprintf("%s:%d", node.Address, node.Port),
		RawQuery: query.Encode(),
	}
	return u.String(), nil
}

func (p *TUICProtocol) ToConfig(node *Node) (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":     "tuic",
		"address":  node.Address,
		"port":     node.Port,
		"uuid":     node.UUID,
		"password": node.Password,
		"tls":      true,
	}, nil
}
