package subscription

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"singdns/api/logger"
	"singdns/api/models"
)

func parseVLESSURL(vlessURL string) (*models.Node, error) {
	logger.LogDebug("开始解析 VLESS URL: %s", vlessURL)

	// Parse URL
	uri, err := url.Parse(vlessURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse VLESS URL: %v", err)
	}

	// Extract UUID
	if uri.User == nil {
		return nil, fmt.Errorf("missing UUID in VLESS URL")
	}
	uuid := uri.User.String()
	if uuid == "" {
		return nil, fmt.Errorf("invalid UUID")
	}

	// Extract port
	port, err := strconv.Atoi(uri.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v", err)
	}

	// Parse query parameters
	query := uri.Query()

	// Create node
	node := &models.Node{
		Type:     "vless",
		Address:  uri.Hostname(),
		Port:     port,
		UUID:     uuid,
		Network:  "tcp", // 默认使用 TCP
		Security: "none",
		Flow:     "",
	}

	// Handle name
	if name := uri.Fragment; name != "" {
		node.Name = name
	} else {
		node.Name = fmt.Sprintf("VLESS-%s:%d", node.Address, node.Port)
	}

	// Try to URL decode the name
	if decodedName, err := url.QueryUnescape(node.Name); err == nil {
		node.Name = decodedName
	}

	// Handle security
	if security := query.Get("security"); security != "" {
		node.Security = security
	}

	// Handle network type
	if network := query.Get("type"); network != "" {
		node.Network = network
	}

	// Handle flow
	if flow := query.Get("flow"); flow != "" {
		node.Flow = flow
	}

	// Handle TLS settings
	if node.Security == "tls" {
		node.TLS = true
		if sni := query.Get("sni"); sni != "" {
			node.SNI = sni
		}
		if fp := query.Get("fp"); fp != "" {
			node.Fingerprint = fp
		}
		if alpn := query.Get("alpn"); alpn != "" {
			node.ALPN = models.StringSlice(strings.Split(alpn, ","))
		}
	}

	// Handle transport settings
	switch node.Network {
	case "ws":
		if path := query.Get("path"); path != "" {
			node.Path = path
		}
		if host := query.Get("host"); host != "" {
			node.Host = host
		}
	case "grpc":
		if serviceName := query.Get("serviceName"); serviceName != "" {
			node.ServiceName = serviceName
		}
	}

	logger.LogInfo("成功解析 VLESS 节点: %s (%s://%s:%d)", node.Name, node.Network, node.Address, node.Port)
	return node, nil
}
