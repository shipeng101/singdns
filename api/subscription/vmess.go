package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"singdns/api/logger"
	"singdns/api/models"
	"strconv"
	"strings"
)

// parseVMessURL parses a VMess URL into a Node
func parseVMessURL(urlStr string) (*models.Node, error) {
	logger.LogDebug("开始解析 VMess URL: %s", urlStr)

	// Remove "vmess://" prefix
	encoded := strings.TrimPrefix(urlStr, "vmess://")
	logger.LogDebug("去除前缀后的内容: %s", encoded)

	// 处理可能的填充问题
	if padding := len(encoded) % 4; padding > 0 {
		encoded += strings.Repeat("=", 4-padding)
		logger.LogDebug("添加填充后的内容: %s", encoded)
	}

	// Decode Base64
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		logger.LogDebug("标准 Base64 解码失败，尝试 URL-safe 解码")
		decoded, err = base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			logger.LogWarning("Base64 解码失败: %v", err)
			return nil, fmt.Errorf("failed to decode VMess URL: %v", err)
		}
	}
	logger.LogDebug("Base64 解码后的内容: %s", string(decoded))

	// 首先尝试解析为原始 JSON
	var rawConfig map[string]interface{}
	decodedStr := string(decoded)
	decodedStr = strings.ReplaceAll(decodedStr, "'", "\"")
	decodedStr = strings.ReplaceAll(decodedStr, "\n", "")
	decodedStr = strings.TrimSpace(decodedStr)

	if err := json.Unmarshal([]byte(decodedStr), &rawConfig); err != nil {
		logger.LogWarning("JSON 解析失败: %v", err)
		return nil, fmt.Errorf("failed to parse VMess config: %v", err)
	}

	// 创建节点配置
	node := &models.Node{
		Type: "vmess",
	}

	// 解析各个字段
	if v, ok := rawConfig["v"].(string); ok {
		node.Version = v
	}
	if ps, ok := rawConfig["ps"].(string); ok {
		node.Name = ps
	}
	if add, ok := rawConfig["add"].(string); ok {
		node.Address = add
	}
	// 处理端口，支持字符串和数字格式
	if port, ok := rawConfig["port"].(float64); ok {
		node.Port = int(port)
	} else if portStr, ok := rawConfig["port"].(string); ok {
		if port, err := strconv.Atoi(portStr); err == nil {
			node.Port = port
		}
	}
	if id, ok := rawConfig["id"].(string); ok {
		node.UUID = id
	}
	// 处理 AlterID，支持字符串和数字格式
	if aid, ok := rawConfig["aid"].(float64); ok {
		node.AlterID = int(aid)
	} else if aidStr, ok := rawConfig["aid"].(string); ok {
		if aid, err := strconv.Atoi(aidStr); err == nil {
			node.AlterID = aid
		}
	}
	if net, ok := rawConfig["net"].(string); ok {
		node.Network = net
	}
	if path, ok := rawConfig["path"].(string); ok {
		node.Path = path
	}
	if host, ok := rawConfig["host"].(string); ok {
		node.Host = host
	}
	if tls, ok := rawConfig["tls"].(string); ok {
		node.TLS = tls == "tls"
	}
	if security, ok := rawConfig["security"].(string); ok {
		node.Security = security
	}

	// 验证必要字段
	if node.Address == "" || node.Port == 0 || node.UUID == "" {
		logger.LogWarning("缺少必要字段 - 地址: %s, 端口: %d, UUID: %s", node.Address, node.Port, node.UUID)
		return nil, fmt.Errorf("missing required fields")
	}

	// 设置默认值
	if node.Security == "" {
		node.Security = "auto"
	}
	if node.Network == "" {
		node.Network = "tcp"
	}

	// 如果名称为空，生成一个
	if node.Name == "" {
		node.Name = fmt.Sprintf("VMess-%s:%d", node.Address, node.Port)
	}

	logger.LogInfo("成功解析 VMess 节点: %s (%s://%s:%d)", node.Name, node.Network, node.Address, node.Port)
	logger.LogDebug("节点详细信息 - UUID: %s, AlterID: %d, Network: %s, TLS: %v, Security: %s",
		node.UUID, node.AlterID, node.Network, node.TLS, node.Security)
	return node, nil
}
