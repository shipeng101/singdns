package subscription

import (
	"fmt"
	"singdns/api/logger"
	"singdns/api/protocols"

	"gopkg.in/yaml.v3"
)

// ClashConfig represents Clash configuration
type ClashConfig struct {
	Proxies []ClashProxy `yaml:"proxies"`
}

// ClashProxy represents a Clash proxy
type ClashProxy struct {
	Name           string                 `yaml:"name"`
	Type           string                 `yaml:"type"`
	Server         string                 `yaml:"server"`
	Port           int                    `yaml:"port"`
	Password       string                 `yaml:"password,omitempty"`
	Cipher         string                 `yaml:"cipher,omitempty"`
	UUID           string                 `yaml:"uuid,omitempty"`
	AlterID        int                    `yaml:"alterId,omitempty"`
	Network        string                 `yaml:"network,omitempty"`
	WSPath         string                 `yaml:"ws-path,omitempty"`
	WSHeaders      map[string]string      `yaml:"ws-headers,omitempty"`
	GRPC           map[string]interface{} `yaml:"grpc,omitempty"`
	TLS            bool                   `yaml:"tls,omitempty"`
	SNI            string                 `yaml:"sni,omitempty"`
	SkipCertVerify bool                   `yaml:"skip-cert-verify,omitempty"`
	UDP            bool                   `yaml:"udp,omitempty"`
}

// ParseClash parses Clash subscription content
func ParseClash(content []byte) ([]*protocols.Node, error) {
	var config ClashConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("invalid Clash config: %v", err)
	}

	nodes := make([]*protocols.Node, 0, len(config.Proxies))
	for _, proxy := range config.Proxies {
		node, err := convertClashProxy(&proxy)
		if err != nil {
			logger.LogWarning("Failed to convert Clash proxy: %v", err)
			continue
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// convertClashProxy converts a Clash proxy to a Node
func convertClashProxy(proxy *ClashProxy) (*protocols.Node, error) {
	node := &protocols.Node{
		Name:    proxy.Name,
		Type:    proxy.Type,
		Address: proxy.Server,
		Port:    proxy.Port,
		TLS:     proxy.TLS,
		Host:    proxy.SNI,
	}

	switch proxy.Type {
	case "ss", "shadowsocks":
		node.Type = "ss"
		node.Method = proxy.Cipher
		node.Password = proxy.Password

	case "vmess":
		node.UUID = proxy.UUID
		node.AlterId = proxy.AlterID
		if proxy.Network != "" {
			node.Network = proxy.Network
			switch proxy.Network {
			case "ws":
				node.Path = proxy.WSPath
				if h := proxy.WSHeaders; h != nil {
					if host, ok := h["Host"]; ok {
						node.Host = host
					}
				}
			case "grpc":
				if g := proxy.GRPC; g != nil {
					if sn, ok := g["service-name"].(string); ok {
						node.Path = sn
					}
				}
			}
		}

	case "trojan":
		node.Password = proxy.Password
		if proxy.Network != "" {
			node.Network = proxy.Network
			switch proxy.Network {
			case "ws":
				node.Path = proxy.WSPath
				if h := proxy.WSHeaders; h != nil {
					if host, ok := h["Host"]; ok {
						node.Host = host
					}
				}
			case "grpc":
				if g := proxy.GRPC; g != nil {
					if sn, ok := g["service-name"].(string); ok {
						node.Path = sn
					}
				}
			}
		}

	default:
		return nil, fmt.Errorf("unsupported proxy type: %s", proxy.Type)
	}

	return node, nil
}
