package subscription

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"singdns/api/logger"
	"singdns/api/models"
	"singdns/api/protocols"
	"strings"

	"gopkg.in/yaml.v3"
)

// SubscriptionConverter converts between different subscription formats
type SubscriptionConverter struct{}

// NewSubscriptionConverter creates a new subscription converter
func NewSubscriptionConverter() *SubscriptionConverter {
	return &SubscriptionConverter{}
}

// ConvertToText converts nodes to text subscription format
func (c *SubscriptionConverter) ConvertToText(nodes []*models.Node) (string, error) {
	var urls []string
	for _, node := range nodes {
		url, err := protocols.ToURL(node)
		if err != nil {
			logger.LogWarning("Failed to convert node to URL: %v", err)
			continue
		}
		urls = append(urls, url)
	}

	content := strings.Join(urls, "\n")
	return base64.StdEncoding.EncodeToString([]byte(content)), nil
}

// ConvertToClash converts nodes to Clash subscription format
func (c *SubscriptionConverter) ConvertToClash(nodes []*models.Node) (string, error) {
	config := map[string]interface{}{
		"port":                7890,
		"socks-port":          7891,
		"allow-lan":           true,
		"mode":                "rule",
		"log-level":           "info",
		"external-controller": "127.0.0.1:9090",
		"proxies":             make([]map[string]interface{}, 0, len(nodes)),
		"proxy-groups": []map[string]interface{}{
			{
				"name":     "PROXY",
				"type":     "url-test",
				"proxies":  make([]string, 0, len(nodes)),
				"url":      "http://www.gstatic.com/generate_204",
				"interval": 300,
			},
		},
	}

	proxyGroup := config["proxy-groups"].([]map[string]interface{})[0]
	proxies := proxyGroup["proxies"].([]string)

	for _, node := range nodes {
		proxy, err := convertNodeToClashProxy(node)
		if err != nil {
			logger.LogWarning("Failed to convert node to Clash proxy: %v", err)
			continue
		}

		config["proxies"] = append(config["proxies"].([]map[string]interface{}), proxy)
		proxies = append(proxies, node.Name)
	}

	proxyGroup["proxies"] = proxies

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(config); err != nil {
		return "", fmt.Errorf("failed to encode Clash config: %v", err)
	}

	return buf.String(), nil
}

// ConvertToSingbox converts nodes to sing-box subscription format
func (c *SubscriptionConverter) ConvertToSingbox(nodes []*models.Node) (string, error) {
	config := map[string]interface{}{
		"log": map[string]interface{}{
			"level":  "info",
			"output": "box.log",
		},
		"dns": map[string]interface{}{
			"servers": []map[string]interface{}{
				{"tag": "google", "address": "8.8.8.8"},
				{"tag": "local", "address": "local", "detour": "direct"},
			},
			"rules": []map[string]interface{}{
				{"geosite": "cn", "server": "local"},
			},
		},
		"inbounds": []map[string]interface{}{
			{
				"type":        "mixed",
				"tag":         "mixed-in",
				"listen":      "::",
				"listen_port": 2080,
			},
		},
		"outbounds": make([]map[string]interface{}, 0, len(nodes)+2),
		"route": map[string]interface{}{
			"rules": []map[string]interface{}{
				{"geosite": "cn", "outbound": "direct"},
				{"geoip": "cn", "outbound": "direct"},
			},
			"auto_detect_interface": true,
		},
	}

	// Add direct and block outbounds
	config["outbounds"] = append(config["outbounds"].([]map[string]interface{}),
		map[string]interface{}{"type": "direct", "tag": "direct"},
		map[string]interface{}{"type": "block", "tag": "block"},
	)

	for _, node := range nodes {
		outbound, err := convertNodeToSingboxOutbound(node)
		if err != nil {
			logger.LogWarning("Failed to convert node to sing-box outbound: %v", err)
			continue
		}

		config["outbounds"] = append(config["outbounds"].([]map[string]interface{}), outbound)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to encode sing-box config: %v", err)
	}

	return string(data), nil
}

// convertNodeToClashProxy converts a Node to Clash proxy configuration
func convertNodeToClashProxy(node *models.Node) (map[string]interface{}, error) {
	proxy := map[string]interface{}{
		"name":   node.Name,
		"type":   node.Type,
		"server": node.Address,
		"port":   node.Port,
	}

	switch node.Type {
	case "ss":
		proxy["cipher"] = node.Method
		proxy["password"] = node.Password

	case "vmess":
		proxy["uuid"] = node.UUID
		proxy["alterId"] = node.AlterID
		if node.Network != "" {
			proxy["network"] = node.Network
			switch node.Network {
			case "ws":
				proxy["ws-path"] = node.Path
				if node.Host != "" {
					proxy["ws-headers"] = map[string]string{
						"Host": node.Host,
					}
				}
			case "grpc":
				proxy["grpc-opts"] = map[string]interface{}{
					"grpc-service-name": node.Path,
				}
			}
		}

	case "trojan":
		proxy["password"] = node.Password
		if node.Network != "" {
			proxy["network"] = node.Network
			switch node.Network {
			case "ws":
				proxy["ws-path"] = node.Path
				if node.Host != "" {
					proxy["ws-headers"] = map[string]string{
						"Host": node.Host,
					}
				}
			case "grpc":
				proxy["grpc-opts"] = map[string]interface{}{
					"grpc-service-name": node.Path,
				}
			}
		}

	case "vless":
		proxy["uuid"] = node.UUID
		if node.Flow != "" {
			proxy["flow"] = node.Flow
		}

	case "hy2":
		proxy["password"] = node.Password

	case "tuic":
		proxy["uuid"] = node.UUID
		proxy["password"] = node.Password
		if node.CC != "" {
			proxy["congestion-controller"] = node.CC
		}

	default:
		return nil, fmt.Errorf("unsupported proxy type: %s", node.Type)
	}

	if node.TLS {
		proxy["tls"] = true
		if node.Host != "" {
			proxy["servername"] = node.Host
		}
	}

	return proxy, nil
}

// convertNodeToSingboxOutbound converts a Node to sing-box outbound configuration
func convertNodeToSingboxOutbound(node *models.Node) (map[string]interface{}, error) {
	outbound := map[string]interface{}{
		"type":        node.Type,
		"tag":         node.Name,
		"server":      node.Address,
		"server_port": node.Port,
	}

	switch node.Type {
	case "ss":
		outbound["method"] = node.Method
		outbound["password"] = node.Password

	case "vmess":
		outbound["uuid"] = node.UUID
		outbound["alter_id"] = node.AlterID
		outbound["security"] = "auto"

	case "trojan":
		outbound["password"] = node.Password

	case "vless":
		outbound["uuid"] = node.UUID
		if node.Flow != "" {
			outbound["flow"] = node.Flow
		}

	case "hysteria2":
		outbound["password"] = node.Password
		if node.Up != "" || node.Down != "" {
			outbound["bandwidth"] = map[string]interface{}{
				"up":   node.Up,
				"down": node.Down,
			}
		}

	case "tuic":
		outbound["uuid"] = node.UUID
		outbound["password"] = node.Password
		if node.CC != "" {
			outbound["congestion_control"] = node.CC
		}

	default:
		return nil, fmt.Errorf("unsupported outbound type: %s", node.Type)
	}

	if node.TLS {
		outbound["tls"] = map[string]interface{}{
			"enabled": true,
		}
		if node.Host != "" {
			outbound["tls"].(map[string]interface{})["server_name"] = node.Host
		}
	}

	if node.Network != "" {
		outbound["transport"] = map[string]interface{}{
			"type": node.Network,
		}
		switch node.Network {
		case "ws":
			transport := outbound["transport"].(map[string]interface{})
			if node.Path != "" {
				transport["path"] = node.Path
			}
			if node.Host != "" {
				transport["headers"] = map[string]string{
					"Host": node.Host,
				}
			}
		case "grpc":
			transport := outbound["transport"].(map[string]interface{})
			if node.Path != "" {
				transport["service_name"] = node.Path
			}
		}
	}

	return outbound, nil
}
