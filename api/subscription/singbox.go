package subscription

import (
	"encoding/json"
	"fmt"
	"singdns/api/logger"
	"singdns/api/models"
)

// parseSingboxSubscription parses a SingBox subscription
func parseSingboxSubscription(content []byte) ([]*models.Node, error) {
	logger.LogDebug("开始解析 SingBox 配置")

	var config struct {
		Outbounds []struct {
			Tag        string `json:"tag"`
			Type       string `json:"type"`
			Server     string `json:"server"`
			ServerPort int    `json:"server_port"`
			Password   string `json:"password"`
			Method     string `json:"method"`
			UUID       string `json:"uuid"`
			AlterID    int    `json:"alter_id"`
			Network    string `json:"network"`
			TLS        struct {
				Enabled     bool     `json:"enabled"`
				ServerName  string   `json:"server_name"`
				Insecure    bool     `json:"insecure"`
				ALPN        []string `json:"alpn"`
				Fingerprint string   `json:"fingerprint"`
			} `json:"tls"`
			Transport struct {
				Type        string            `json:"type"`
				Path        string            `json:"path"`
				Host        string            `json:"host"`
				ServiceName string            `json:"service_name"`
				Headers     map[string]string `json:"headers"`
			} `json:"transport"`
			Flow string `json:"flow"`
		} `json:"outbounds"`
	}

	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse SingBox config: %v", err)
	}

	var nodes []*models.Node
	for i, outbound := range config.Outbounds {
		logger.LogDebug("正在解析第 %d 个出站: %s (%s)", i+1, outbound.Tag, outbound.Type)

		// 只处理支持的代理类型
		switch outbound.Type {
		case "shadowsocks", "vmess", "trojan", "vless":
			node := &models.Node{
				Name:     outbound.Tag,
				Type:     outbound.Type,
				Address:  outbound.Server,
				Port:     outbound.ServerPort,
				Method:   outbound.Method,
				Password: outbound.Password,
				UUID:     outbound.UUID,
				AlterID:  outbound.AlterID,
				Network:  outbound.Network,
				Flow:     outbound.Flow,
			}

			// 处理 TLS 设置
			if outbound.TLS.Enabled {
				node.TLS = true
				node.SNI = outbound.TLS.ServerName
				node.SkipCertVerify = outbound.TLS.Insecure
				node.ALPN = outbound.TLS.ALPN
				node.Fingerprint = outbound.TLS.Fingerprint
			}

			// 处理传输层设置
			if outbound.Transport.Type != "" {
				node.Network = outbound.Transport.Type
				node.Path = outbound.Transport.Path
				node.Host = outbound.Transport.Host
				node.ServiceName = outbound.Transport.ServiceName

				// 处理 WebSocket 的 Host 头
				if node.Network == "ws" && outbound.Transport.Headers != nil {
					if host, ok := outbound.Transport.Headers["Host"]; ok {
						node.Host = host
					}
				}
			}

			// 如果名称为空，生成一个
			if node.Name == "" {
				node.Name = fmt.Sprintf("%s-%s:%d", outbound.Type, node.Address, node.Port)
			}

			nodes = append(nodes, node)
			logger.LogInfo("成功解析节点: %s (%s)", node.Name, node.Type)
		default:
			logger.LogDebug("跳过不支持的出站类型: %s", outbound.Type)
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("未找到有效的节点")
	}

	logger.LogInfo("成功解析 %d 个节点", len(nodes))
	return nodes, nil
}
