package subscription

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"singdns/api/logger"
	"singdns/api/models"
	"strconv"
	"strings"
)

// parseShadowsocksURL parses a Shadowsocks URL into a Node
func parseShadowsocksURL(urlStr string) (*models.Node, error) {
	logger.LogDebug("开始解析 Shadowsocks URL: %s", urlStr)

	// Remove "ss://" prefix
	encoded := strings.TrimPrefix(urlStr, "ss://")

	// Split fragment (name) if exists
	var name string
	if idx := strings.Index(encoded, "#"); idx >= 0 {
		name = encoded[idx+1:]
		encoded = encoded[:idx]
	}

	// Decode Base64 if necessary
	var userInfo string
	if strings.Contains(encoded, "@") {
		parts := strings.SplitN(encoded, "@", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid Shadowsocks URL format")
		}
		decoded, err := base64.StdEncoding.DecodeString(parts[0])
		if err != nil {
			decoded, err = base64.URLEncoding.DecodeString(parts[0])
			if err != nil {
				return nil, fmt.Errorf("failed to decode Shadowsocks user info: %v", err)
			}
		}
		userInfo = string(decoded)
		encoded = fmt.Sprintf("%s@%s", userInfo, parts[1])
	} else {
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			decoded, err = base64.URLEncoding.DecodeString(encoded)
			if err != nil {
				return nil, fmt.Errorf("failed to decode Shadowsocks URL: %v", err)
			}
		}
		encoded = string(decoded)
	}

	// Parse URL
	uri, err := url.Parse("ss://" + encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Shadowsocks URL: %v", err)
	}

	// Extract method and password
	methodAndPassword := strings.SplitN(uri.User.String(), ":", 2)
	if len(methodAndPassword) != 2 {
		return nil, fmt.Errorf("invalid Shadowsocks user info format")
	}

	// Extract port
	port, err := strconv.Atoi(uri.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v", err)
	}

	// Create node
	node := &models.Node{
		Name:     name,
		Type:     "ss",
		Address:  uri.Hostname(),
		Port:     port,
		Method:   methodAndPassword[0],
		Password: methodAndPassword[1],
	}

	// 如果名称为空，生成一个
	if node.Name == "" {
		node.Name = fmt.Sprintf("SS-%s:%d", node.Address, node.Port)
	}

	// 尝试 URL 解码节点名称
	if decodedName, err := url.QueryUnescape(node.Name); err == nil {
		node.Name = decodedName
	}

	logger.LogInfo("成功解析 Shadowsocks 节点: %s (%s:%d)", node.Name, node.Address, node.Port)
	return node, nil
}

// parseShadowsocksSubscription parses a Shadowsocks subscription
func parseShadowsocksSubscription(content []byte) ([]*models.Node, error) {
	logger.LogDebug("开始解析 Shadowsocks 订阅")

	// 先尝试 Base64 解码
	decoded, err := base64.StdEncoding.DecodeString(string(content))
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 content: %v", err)
		}
	}

	// 按行分割
	lines := strings.Split(string(decoded), "\n")
	var nodes []*models.Node

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		logger.LogDebug("正在解析第 %d 行: %s", i+1, line)

		// 解析 SS URL
		if strings.HasPrefix(line, "ss://") {
			node, err := parseShadowsocksURL(line)
			if err != nil {
				logger.LogWarning("解析 SS 链接失败: %v", err)
				continue
			}
			nodes = append(nodes, node)
			logger.LogInfo("成功解析节点: %s", node.Name)
		}
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("未找到有效的节点")
	}

	logger.LogInfo("成功解析 %d 个 Shadowsocks 节点", len(nodes))
	return nodes, nil
}
