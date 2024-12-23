package subscription

import (
	"encoding/base64"
	"fmt"
	"singdns/api/logger"
	"singdns/api/models"
	"strings"

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
func ParseClash(content []byte) ([]*models.Node, error) {
	logger.LogInfo("Starting to parse subscription content")
	logger.LogDebug("Content length: %d bytes", len(content))

	// First try to decode Base64 content
	decodedContent, err := base64.StdEncoding.DecodeString(string(content))
	if err == nil {
		logger.LogInfo("Successfully decoded Base64 content")
		content = decodedContent
	} else {
		logger.LogDebug("Content is not Base64 encoded or already decoded: %v", err)
	}

	// Try to parse as YAML format
	var config struct {
		Proxies []ClashProxy `yaml:"proxies"`
		Proxy   []ClashProxy `yaml:"proxy"`
	}

	if err := yaml.Unmarshal(content, &config); err != nil {
		logger.LogWarning("Failed to parse YAML: %v", err)
		logger.LogDebug("Content preview: %s", string(content[:min(len(content), 200)]))
		return nil, fmt.Errorf("no valid nodes found in subscription")
	}

	// Merge proxies and proxy fields
	allProxies := append(config.Proxies, config.Proxy...)
	logger.LogInfo("Found %d proxies in Clash config", len(allProxies))

	if len(allProxies) == 0 {
		return nil, fmt.Errorf("no proxies found in Clash config")
	}

	nodes := make([]*models.Node, 0, len(allProxies))
	for i, proxy := range allProxies {
		logger.LogDebug("Converting proxy %d: %s (%s)", i+1, proxy.Name, proxy.Type)
		node, err := convertClashProxy(&proxy)
		if err != nil {
			logger.LogWarning("Failed to convert Clash proxy: %v", err)
			continue
		}
		nodes = append(nodes, node)
		logger.LogInfo("Successfully converted node: %s (%s)", node.Name, node.Type)
	}

	if len(nodes) == 0 {
		logger.LogWarning("No valid nodes found after conversion")
		return nil, fmt.Errorf("no valid nodes found in subscription")
	}

	logger.LogInfo("Successfully parsed %d nodes from config", len(nodes))
	return nodes, nil
}

// convertClashProxy converts a Clash proxy to a Node
func convertClashProxy(proxy *ClashProxy) (*models.Node, error) {
	logger.LogDebug("Converting Clash proxy: %s (%s)", proxy.Name, proxy.Type)

	if proxy.Type == "" || proxy.Server == "" || proxy.Port == 0 {
		return nil, fmt.Errorf("missing required proxy fields - type: %s, server: %s, port: %d", proxy.Type, proxy.Server, proxy.Port)
	}

	node := &models.Node{
		Name:           proxy.Name,
		Type:           strings.ToLower(proxy.Type),
		Address:        proxy.Server,
		Port:           proxy.Port,
		TLS:            proxy.TLS,
		Host:           proxy.SNI,
		SkipCertVerify: proxy.SkipCertVerify,
		UDP:            proxy.UDP,
	}

	proxyType := strings.ToLower(proxy.Type)
	logger.LogDebug("Processing proxy type: %s", proxyType)

	switch proxyType {
	case "ss", "shadowsocks":
		node.Type = "ss"
		if proxy.Cipher == "" || proxy.Password == "" {
			return nil, fmt.Errorf("missing required shadowsocks fields - cipher: %s, password: %s", proxy.Cipher, proxy.Password)
		}
		node.Method = proxy.Cipher
		node.Password = proxy.Password
		logger.LogDebug("Converted Shadowsocks proxy - method: %s", node.Method)

	case "vmess":
		if proxy.UUID == "" {
			return nil, fmt.Errorf("missing required vmess field - uuid: %s", proxy.UUID)
		}
		node.UUID = proxy.UUID
		node.AlterID = proxy.AlterID
		node.Network = proxy.Network
		if proxy.Network == "ws" {
			node.Path = proxy.WSPath
			if proxy.WSHeaders != nil {
				node.Host = proxy.WSHeaders["Host"]
			}
			logger.LogDebug("Configured WebSocket - path: %s, host: %s", node.Path, node.Host)
		} else if proxy.Network == "grpc" && proxy.GRPC != nil {
			if serviceName, ok := proxy.GRPC["serviceName"].(string); ok {
				node.Path = serviceName
			}
			logger.LogDebug("Configured gRPC - service name: %s", node.Path)
		}
		logger.LogDebug("Converted VMess proxy - uuid: %s, network: %s", node.UUID, node.Network)

	case "trojan":
		if proxy.Password == "" {
			return nil, fmt.Errorf("missing required trojan field - password: %s", proxy.Password)
		}
		node.Password = proxy.Password
		node.TLS = true // Trojan always uses TLS
		logger.LogDebug("Converted Trojan proxy - tls: %v", node.TLS)

	case "vless":
		if proxy.UUID == "" {
			return nil, fmt.Errorf("missing required vless field - uuid: %s", proxy.UUID)
		}
		node.UUID = proxy.UUID
		node.Network = proxy.Network
		if proxy.Network == "ws" {
			node.Path = proxy.WSPath
			if proxy.WSHeaders != nil {
				node.Host = proxy.WSHeaders["Host"]
			}
			logger.LogDebug("Configured WebSocket - path: %s, host: %s", node.Path, node.Host)
		} else if proxy.Network == "grpc" && proxy.GRPC != nil {
			if serviceName, ok := proxy.GRPC["serviceName"].(string); ok {
				node.Path = serviceName
			}
			logger.LogDebug("Configured gRPC - service name: %s", node.Path)
		}
		logger.LogDebug("Converted VLESS proxy - uuid: %s, network: %s", node.UUID, node.Network)

	default:
		return nil, fmt.Errorf("unsupported proxy type: %s", proxy.Type)
	}

	logger.LogInfo("Successfully converted proxy to node: %s (%s)", node.Name, node.Type)
	return node, nil
}
