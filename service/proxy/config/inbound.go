package config

import (
	"github.com/shipeng101/singdns/pkg/types"
)

// InboundGenerator 入站配置生成器
type InboundGenerator struct {
	config *types.ProxyConfig
}

// NewInboundGenerator 创建入站配置生成器
func NewInboundGenerator(config *types.ProxyConfig) *InboundGenerator {
	return &InboundGenerator{
		config: config,
	}
}

// Generate 生成入站配置
func (g *InboundGenerator) Generate() ([]map[string]interface{}, error) {
	inbounds := []map[string]interface{}{}

	// 添加 Mixed 入站
	if mixed, err := g.generateMixedInbound(); err == nil {
		inbounds = append(inbounds, mixed)
	}

	// 添加 Socks 入站
	if socks, err := g.generateSocksInbound(); err == nil {
		inbounds = append(inbounds, socks)
	}

	// 添加 HTTP 入站
	if http, err := g.generateHTTPInbound(); err == nil {
		inbounds = append(inbounds, http)
	}

	// 添加 TUN 入站
	if g.config.Inbound.EnableTUN {
		if tun, err := g.generateTUNInbound(); err == nil {
			inbounds = append(inbounds, tun)
		}
	}

	// 添加 Transparent 入站
	if g.config.Inbound.EnableTransparent {
		if transparent, err := g.generateTransparentInbound(); err == nil {
			inbounds = append(inbounds, transparent)
		}
	}

	return inbounds, nil
}

// generateMixedInbound 生成 Mixed 入站配置
func (g *InboundGenerator) generateMixedInbound() (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":        "mixed",
		"tag":         "mixed-in",
		"listen":      g.config.Inbound.Listen,
		"listen_port": g.config.Inbound.Port,
		"users": []map[string]interface{}{
			{
				"username": g.config.Inbound.Username,
				"password": g.config.Inbound.Password,
			},
		},
		"set_system_proxy": g.config.Inbound.SetSystemProxy,
	}, nil
}

// generateSocksInbound 生成 Socks 入站配置
func (g *InboundGenerator) generateSocksInbound() (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":        "socks",
		"tag":         "socks-in",
		"listen":      g.config.Inbound.Listen,
		"listen_port": g.config.Inbound.SocksPort,
		"users": []map[string]interface{}{
			{
				"username": g.config.Inbound.Username,
				"password": g.config.Inbound.Password,
			},
		},
	}, nil
}

// generateHTTPInbound 生成 HTTP 入站配置
func (g *InboundGenerator) generateHTTPInbound() (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":        "http",
		"tag":         "http-in",
		"listen":      g.config.Inbound.Listen,
		"listen_port": g.config.Inbound.HTTPPort,
		"users": []map[string]interface{}{
			{
				"username": g.config.Inbound.Username,
				"password": g.config.Inbound.Password,
			},
		},
	}, nil
}

// generateTUNInbound 生成 TUN 入站配置
func (g *InboundGenerator) generateTUNInbound() (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":                     "tun",
		"tag":                      "tun-in",
		"interface_name":           g.config.Inbound.TUNInterface,
		"inet4_address":            g.config.Inbound.TUNAddress,
		"mtu":                      g.config.Inbound.TUNMTU,
		"auto_route":               g.config.Inbound.TUNAutoRoute,
		"strict_route":             g.config.Inbound.TUNStrictRoute,
		"stack":                    g.config.Inbound.TUNStack,
		"endpoint_independent_nat": g.config.Inbound.TUNEndpointIndependentNat,
	}, nil
}

// generateTransparentInbound 生成透明代理入站配置
func (g *InboundGenerator) generateTransparentInbound() (map[string]interface{}, error) {
	return map[string]interface{}{
		"type":        "redirect",
		"tag":         "redirect-in",
		"listen":      g.config.Inbound.Listen,
		"listen_port": g.config.Inbound.TransparentPort,
	}, nil
}
