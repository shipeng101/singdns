package proxy

import (
	"encoding/json"
	"fmt"

	"github.com/shipeng101/singdns/pkg/protocol"
	"github.com/shipeng101/singdns/pkg/types"
)

// ConfigGenerator 配置生成器
type ConfigGenerator struct{}

// NewConfigGenerator 创建配置生成器
func NewConfigGenerator() *ConfigGenerator {
	return &ConfigGenerator{}
}

// Generate 生成配置
func (g *ConfigGenerator) Generate(config types.ProxyConfig, currentNode *types.ProxyNode) ([]byte, error) {
	// 设置默认监听地址
	listen := config.Inbound.Listen
	if listen == "" {
		listen = "127.0.0.1"
	}

	// 基础配置
	cfg := map[string]interface{}{
		"log": map[string]interface{}{
			"level":     "info",
			"timestamp": true,
		},
		"dns": map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"tag":              "local",
					"address":          "223.5.5.5",
					"address_resolver": "dns-direct",
					"detour":           "direct",
				},
				{
					"tag":              "remote",
					"address":          "8.8.8.8",
					"address_resolver": "dns-direct",
					"detour":           "proxy",
				},
				{
					"tag":     "dns-direct",
					"address": "223.5.5.5",
				},
			},
			"rules": []map[string]interface{}{
				{
					"domain_suffix": []string{".cn"},
					"server":        "local",
				},
			},
			"final":    "remote",
			"strategy": "prefer_ipv4",
		},
		"inbounds": []map[string]interface{}{
			{
				"type":        "mixed",
				"tag":         "mixed-in",
				"listen":      listen,
				"listen_port": config.Inbound.Port,
				"sniff":       true,
			},
		},
		"outbounds": []map[string]interface{}{
			{
				"type": "direct",
				"tag":  "direct",
			},
			{
				"type": "block",
				"tag":  "block",
			},
			{
				"type": "dns",
				"tag":  "dns-out",
			},
		},
		"route": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"protocol": "dns",
					"outbound": "dns-out",
				},
			},
			"final": "direct",
		},
	}

	// 如果有当前节点，添加到出站
	if currentNode != nil {
		outbound, err := g.generateNodeOutbound(currentNode)
		if err != nil {
			return nil, fmt.Errorf("生成节点出站配置失败: %v", err)
		}

		outbounds := cfg["outbounds"].([]map[string]interface{})
		outbounds = append(outbounds, outbound)
		cfg["outbounds"] = outbounds

		// 修改路由规则
		route := cfg["route"].(map[string]interface{})
		if config.Mode == "global" {
			route["final"] = currentNode.Name
		} else if config.Mode == "rule" {
			rules := route["rules"].([]map[string]interface{})
			for _, rule := range config.Rules {
				if !rule.Enabled {
					continue
				}

				ruleConfig := map[string]interface{}{
					"outbound": currentNode.Name,
				}

				switch rule.Type {
				case "domain":
					ruleConfig["domain"] = []string{rule.Value}
				case "domain_suffix":
					ruleConfig["domain_suffix"] = []string{rule.Value}
				case "domain_keyword":
					ruleConfig["domain_keyword"] = []string{rule.Value}
				case "ip":
					ruleConfig["ip_cidr"] = []string{rule.Value}
				}

				rules = append(rules, ruleConfig)
			}
			route["rules"] = rules
		}
		cfg["route"] = route
	}

	return json.MarshalIndent(cfg, "", "  ")
}

// generateNodeOutbound 生成节点出站配置
func (g *ConfigGenerator) generateNodeOutbound(node *types.ProxyNode) (map[string]interface{}, error) {
	outbound, err := protocol.GetProtocol(node)
	if err != nil {
		return nil, err
	}

	// 转换为map
	data, err := json.Marshal(outbound)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// 添加TLS配置
	if node.TLS {
		result["tls"] = map[string]interface{}{
			"enabled":     true,
			"server_name": node.SNI,
			"alpn":        node.ALPN,
			"insecure":    node.AllowInsecure,
			"utls": map[string]interface{}{
				"enabled":     true,
				"fingerprint": "chrome",
			},
		}
	}

	return result, nil
}
