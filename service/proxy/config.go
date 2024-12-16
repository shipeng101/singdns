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
func (g *ConfigGenerator) Generate(cfg types.ProxyConfig, node *types.ProxyNode) ([]byte, error) {
	// 创建基础配置
	config := map[string]interface{}{
		"log": map[string]interface{}{
			"level":     "info",
			"timestamp": true,
		},
		"inbounds": []map[string]interface{}{
			{
				"type":        "mixed",
				"tag":         "mixed-in",
				"listen":      cfg.Inbound.Listen,
				"listen_port": cfg.Inbound.Port,
				"sniff":       true,
			},
		},
		"outbounds": []map[string]interface{}{},
		"route": map[string]interface{}{
			"rules": []map[string]interface{}{
				{
					"protocol": []string{"dns"},
					"outbound": "dns-out",
				},
			},
			"auto_detect_interface": true,
			"final":                 "direct",
		},
		"dns": map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"tag":     "local",
					"address": "223.5.5.5",
					"detour":  "direct",
				},
				{
					"tag":     "remote",
					"address": "8.8.8.8",
					"detour":  "proxy",
				},
			},
			"rules": []map[string]interface{}{
				{
					"domain_suffix": []string{".cn"},
					"server":        "local",
				},
			},
			"strategy": "prefer_ipv4",
		},
	}

	// 添加出站配置
	outbounds := []map[string]interface{}{}

	// 如果有节点配置，添加到出站
	if node != nil {
		outbound, err := g.generateNodeOutbound(node)
		if err != nil {
			return nil, fmt.Errorf("生成节点出站配置失败: %w", err)
		}
		outbounds = append(outbounds, outbound)
	}

	// 添加其他基础出站
	outbounds = append(outbounds,
		map[string]interface{}{
			"type": "direct",
			"tag":  "direct",
		},
		map[string]interface{}{
			"type": "block",
			"tag":  "block",
		},
		map[string]interface{}{
			"type": "dns",
			"tag":  "dns-out",
		},
	)

	config["outbounds"] = outbounds

	// 序列化配置
	return json.MarshalIndent(config, "", "  ")
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
