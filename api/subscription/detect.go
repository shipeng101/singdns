package subscription

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"singdns/api/logger"
)

// DetectSubscriptionType 自动检测订阅内容的类型
func DetectSubscriptionType(content []byte) string {
	logger.LogDebug("开始检测订阅类型")

	if len(content) == 0 {
		logger.LogDebug("订阅内容为空")
		return "unknown"
	}

	// 记录内容前缀用于调试
	previewLen := min(len(content), 50)
	logger.LogDebug("订阅内容前" + string(rune(previewLen)) + "个字符: " + string(content[:previewLen]))

	// 首先尝试检查原始内容
	contentStr := string(content)

	// 检查是否是 Clash 配置
	if strings.Contains(contentStr, "proxies:") && strings.Contains(contentStr, "proxy-groups:") {
		logger.LogDebug("检测到 Clash 配置")
		return "clash"
	}

	// 检查是否是 SingBox 配置
	var jsonObj map[string]interface{}
	if err := json.Unmarshal(content, &jsonObj); err == nil {
		if outbounds, ok := jsonObj["outbounds"].([]interface{}); ok && len(outbounds) > 0 {
			// 检查是否包含有效的出站配置
			for _, outbound := range outbounds {
				if out, ok := outbound.(map[string]interface{}); ok {
					if outType, hasType := out["type"]; hasType && outType != nil {
						// 确保是代理类型的出站
						if typeStr, ok := outType.(string); ok {
							switch typeStr {
							case "vmess", "vless", "shadowsocks", "trojan", "hysteria2", "tuic":
								logger.LogDebug("检测到 SingBox 配置")
								return "singbox"
							}
						}
					}
				}
			}
		}
	}

	// 检查是否包含直接的协议链接
	if strings.Contains(contentStr, "vmess://") {
		logger.LogDebug("检测到 VMess 链接")
		return "v2ray"
	}
	if strings.Contains(contentStr, "vless://") {
		logger.LogDebug("检测到 VLESS 链接")
		return "v2ray"
	}
	if strings.Contains(contentStr, "ss://") {
		logger.LogDebug("检测到 Shadowsocks 链接")
		return "v2ray"
	}
	if strings.Contains(contentStr, "trojan://") {
		logger.LogDebug("检测到 Trojan 链接")
		return "v2ray"
	}

	// 尝试 Base64 解码
	decoded, err := base64.StdEncoding.DecodeString(contentStr)
	if err == nil {
		decodedStr := string(decoded)
		previewLen = min(len(decodedStr), 50)
		logger.LogDebug("Base64 解码后内容前" + string(rune(previewLen)) + "个字符: " + decodedStr[:previewLen])

		// 检查解码后的内容
		if strings.Contains(decodedStr, "vmess://") ||
			strings.Contains(decodedStr, "vless://") ||
			strings.Contains(decodedStr, "ss://") ||
			strings.Contains(decodedStr, "trojan://") {
			logger.LogDebug("检测到 Base64 编码的代理链接")
			return "v2ray"
		}

		// 如果解码后的内容看起来像 Base64，继续尝试解码
		if !strings.Contains(decodedStr, "://") {
			if decoded2, err := base64.StdEncoding.DecodeString(decodedStr); err == nil {
				decodedStr = string(decoded2)
				previewLen = min(len(decodedStr), 50)
				logger.LogDebug("二次 Base64 解码后内容前" + string(rune(previewLen)) + "个字符: " + decodedStr[:previewLen])
				if strings.Contains(decodedStr, "vmess://") ||
					strings.Contains(decodedStr, "vless://") ||
					strings.Contains(decodedStr, "ss://") ||
					strings.Contains(decodedStr, "trojan://") {
					logger.LogDebug("检测到双重 Base64 编码的代理链接")
					return "v2ray"
				}
			}
		}
	}

	// 如果无法确定类型，默认尝试作为 v2ray 类型处理
	logger.LogDebug("无法确定订阅类型，默认使用 v2ray")
	return "v2ray"
}
