package config

import (
	"fmt"

	"github.com/shipeng101/singdns/pkg/protocol"
	"github.com/shipeng101/singdns/pkg/types"
)

// OutboundGenerator 出站配置生成器
type OutboundGenerator struct {
	nodes []*types.ProxyNode
}

// NewOutboundGenerator 创建出站配置生成器
func NewOutboundGenerator(nodes []*types.ProxyNode) *OutboundGenerator {
	return &OutboundGenerator{
		nodes: nodes,
	}
}

// Generate 生成出站配置
func (g *OutboundGenerator) Generate() ([]map[string]interface{}, error) {
	outbounds := []map[string]interface{}{
		{
			"type": "selector",
			"tag":  "proxy",
			"outbounds": []string{
				"auto",
				"direct",
			},
		},
		{
			"type":      "urltest",
			"tag":       "auto",
			"outbounds": []string{},
			"url":       "https://www.gstatic.com/generate_204",
			"interval":  "1m",
			"tolerance": 50,
		},
	}

	// 生成节点出站
	for _, node := range g.nodes {
		outbound, err := g.generateNodeOutbound(node)
		if err != nil {
			return nil, fmt.Errorf("generate node outbound failed: %v", err)
		}
		outbounds = append(outbounds, outbound)

		// 添加到自动选择组
		autoOutbound := outbounds[1]
		autoOutbound["outbounds"] = append(autoOutbound["outbounds"].([]string), node.Name)

		// 添加到手动选择组
		proxyOutbound := outbounds[0]
		proxyOutbound["outbounds"] = append(proxyOutbound["outbounds"].([]string), node.Name)
	}

	return outbounds, nil
}

// generateNodeOutbound 生成节点出站配置
func (g *OutboundGenerator) generateNodeOutbound(node *types.ProxyNode) (map[string]interface{}, error) {
	outbound, err := protocol.GetProtocol(node)
	if err != nil {
		return nil, fmt.Errorf("get protocol failed: %v", err)
	}

	// 添加TLS配置
	if node.TLS {
		result := outbound.(map[string]interface{})
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
		return result, nil
	}

	return outbound.(map[string]interface{}), nil
}
