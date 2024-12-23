package subscription

import (
	"encoding/base64"
	"fmt"
	"singdns/api/logger"
	"singdns/api/models"
	"strings"
)

// ParseSubscription parses subscription content and returns nodes
func ParseSubscription(content []byte, subType string) ([]*models.Node, error) {
	logger.LogInfo("开始解析订阅内容，类型: %s", subType)
	logger.LogDebug("订阅内容预览: %s", string(content[:min(len(content), 100)]))

	// 如果没有指定类型，自动检测
	if subType == "" {
		subType = DetectSubscriptionType(content)
		logger.LogInfo("自动检测到订阅类型: %s", subType)
	}

	// 根据类型解析内容
	switch subType {
	case "clash":
		return ParseClash(content)
	case "v2ray", "vless", "vmess":
		// 尝试解析 Base64 编码的内容
		decoded := content
		decodedStr := string(content)

		// 如果内容看起来像 Base64（不包含常见的协议前缀），尝试解码
		if !strings.Contains(decodedStr, "vmess://") &&
			!strings.Contains(decodedStr, "vless://") &&
			!strings.Contains(decodedStr, "ss://") &&
			!strings.Contains(decodedStr, "trojan://") {

			logger.LogDebug("尝试 Base64 解码")
			if dec, err := base64.StdEncoding.DecodeString(decodedStr); err == nil {
				decoded = dec
				logger.LogDebug("标准 Base64 解码成功")
			} else if dec, err := base64.URLEncoding.DecodeString(decodedStr); err == nil {
				decoded = dec
				logger.LogDebug("URL-safe Base64 解码成功")
			} else {
				logger.LogDebug("Base64 解码失败，使用原始内容")
			}
		}

		// 按行分割并解析每一行
		lines := strings.Split(string(decoded), "\n")
		var nodes []*models.Node
		logger.LogDebug("开始解析 %d 行内容", len(lines))

		for i, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			logger.LogDebug("正在解析第 %d 行: %s", i+1, line)

			var node *models.Node
			var err error

			// 根据协议类型解析
			switch {
			case strings.HasPrefix(line, "vmess://"):
				node, err = parseVMessURL(line)
			case strings.HasPrefix(line, "vless://"):
				node, err = parseVLESSURL(line)
			case strings.HasPrefix(line, "ss://"):
				node, err = parseShadowsocksURL(line)
			case strings.HasPrefix(line, "trojan://"):
				node, err = parseTrojanURL(line)
			case strings.HasPrefix(line, "hy2://"):
				node, err = parseHysteria2URL(line)
			case strings.HasPrefix(line, "tuic://"):
				node, err = parseTUICURL(line)
			default:
				logger.LogDebug("跳过不支持的协议: %s", line)
				continue
			}

			if err != nil {
				logger.LogWarning("解析节点失败: %v", err)
				continue
			}

			if node != nil {
				nodes = append(nodes, node)
				logger.LogInfo("成功解析节点: %s (%s)", node.Name, node.Type)
			}
		}

		if len(nodes) == 0 {
			return nil, fmt.Errorf("未找到有效的节点")
		}

		logger.LogInfo("成功解析 %d 个节点", len(nodes))
		return nodes, nil

	case "shadowsocks":
		return parseShadowsocksSubscription(content)
	case "singbox":
		return parseSingboxSubscription(content)
	default:
		return nil, fmt.Errorf("不支持的订阅类型: %s", subType)
	}
}
