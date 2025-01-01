package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"singdns/api/logger"
	"singdns/api/models"
	"strings"
)

// parseVMessURL parses a VMess URL into a Node
func parseVMessURL(urlStr string) (*models.Node, error) {
	logger.LogDebug("开始解析 VMess URL: %s", urlStr)

	// Remove "vmess://" prefix
	encoded := strings.TrimPrefix(urlStr, "vmess://")

	// Decode Base64
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode VMess URL: %v", err)
		}
	}

	// Parse JSON
	var config struct {
		Version string `json:"v"`
		Name    string `json:"ps"`
		Address string `json:"add"`
		Port    int    `json:"port"`
		UUID    string `json:"id"`
		AlterID int    `json:"aid"`
		Network string `json:"net"`
		Type    string `json:"type"`
		Host    string `json:"host"`
		Path    string `json:"path"`
		TLS     string `json:"tls"`
	}

	if err := json.Unmarshal(decoded, &config); err != nil {
		return nil, fmt.Errorf("failed to parse VMess config: %v", err)
	}

	node := &models.Node{
		Type:    "vmess",
		Name:    config.Name,
		Address: config.Address,
		Port:    config.Port,
		UUID:    config.UUID,
		AlterID: config.AlterID,
		Network: config.Network,
		Path:    config.Path,
		Host:    config.Host,
		TLS:     config.TLS == "tls",
	}

	// 如果名称为空，生成一个
	if node.Name == "" {
		node.Name = fmt.Sprintf("VMess-%s:%d", node.Address, node.Port)
	}

	logger.LogInfo("成功解析 VMess 节点: %s (%s://%s:%d)", node.Name, node.Network, node.Address, node.Port)
	return node, nil
}
