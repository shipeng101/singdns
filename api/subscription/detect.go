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
	previewLen := min(len(content), 100)
	logger.LogDebug("订阅内容前 %d 个字符: %s", previewLen, string(content[:previewLen]))

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
		logger.LogDebug("成功解析为 JSON 格式")
		if outbounds, ok := jsonObj["outbounds"].([]interface{}); ok && len(outbounds) > 0 {
			logger.LogDebug("找到 outbounds 字段，包含 %d 个出站配置", len(outbounds))
			// 检查是否包含有效的出站配置
			for i, outbound := range outbounds {
				if out, ok := outbound.(map[string]interface{}); ok {
					if outType, hasType := out["type"]; hasType && outType != nil {
						// 确保是代理类型的出站
						if typeStr, ok := outType.(string); ok {
							logger.LogDebug("出站 %d 类型: %s", i+1, typeStr)
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
	} else {
		logger.LogDebug("不是有效的 JSON 格式: %v", err)
	}

	// 尝试多次 Base64 解码
	decodedStr := contentStr
	maxAttempts := 3
	for i := 0; i < maxAttempts; i++ {
		logger.LogDebug("尝试第 %d 次 Base64 解码", i+1)

		// 处理可能的填充问题
		if padding := len(decodedStr) % 4; padding > 0 {
			decodedStr += strings.Repeat("=", 4-padding)
			logger.LogDebug("添加填充后的内容: %s", decodedStr[:min(len(decodedStr), 100)])
		}

		// 检查是否包含直接的协议链接
		if strings.Contains(decodedStr, "vmess://") {
			logger.LogDebug("检测到 VMess 链接")
			return "v2ray"
		}
		if strings.Contains(decodedStr, "vless://") {
			logger.LogDebug("检测到 VLESS 链接")
			return "v2ray"
		}
		if strings.Contains(decodedStr, "ss://") {
			logger.LogDebug("检测到 Shadowsocks 链接")
			return "v2ray"
		}
		if strings.Contains(decodedStr, "trojan://") {
			logger.LogDebug("检测到 Trojan 链接")
			return "v2ray"
		}

		// 尝试 Base64 解码
		if decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(decodedStr)); err == nil {
			decodedStr = string(decoded)
			logger.LogDebug("Base64 解码成功，解码后内容: %s", decodedStr[:min(len(decodedStr), 100)])
			continue
		} else if decoded, err := base64.URLEncoding.DecodeString(strings.TrimSpace(decodedStr)); err == nil {
			decodedStr = string(decoded)
			logger.LogDebug("URL-safe Base64 解码成功，解码后内容: %s", decodedStr[:min(len(decodedStr), 100)])
			continue
		} else {
			logger.LogDebug("Base64 解码失败: %v", err)
		}

		// 如果无法继续解码，跳出循环
		logger.LogDebug("无法继续 Base64 解码")
		break
	}

	// 如果无法确定类型，默认尝试作为 v2ray 类型处理
	logger.LogDebug("无法确定订阅类型，默认使用 v2ray")
	return "v2ray"
}
