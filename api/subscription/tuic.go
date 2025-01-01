package subscription

import (
	"fmt"
	"net/url"
	"singdns/api/logger"
	"singdns/api/models"
	"strconv"
	"strings"
)

// parseTUICURL parses a TUIC URL into a Node
func parseTUICURL(urlStr string) (*models.Node, error) {
	logger.LogDebug("开始解析 TUIC URL: %s", urlStr)

	// Parse URL
	uri, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TUIC URL: %v", err)
	}

	// Extract UUID and password
	userInfo := strings.Split(uri.User.String(), ":")
	if len(userInfo) != 2 {
		return nil, fmt.Errorf("invalid TUIC user info format, expected uuid:password")
	}

	uuid := userInfo[0]
	password := userInfo[1]

	if uuid == "" {
		return nil, fmt.Errorf("missing UUID in TUIC URL")
	}
	if password == "" {
		return nil, fmt.Errorf("missing password in TUIC URL")
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
		Type:     "tuic",
		Address:  uri.Hostname(),
		Port:     port,
		UUID:     uuid,
		Password: password,
		TLS:      true, // TUIC always uses TLS
	}

	// Handle name
	if name := uri.Fragment; name != "" {
		node.Name = name
	} else {
		node.Name = fmt.Sprintf("TUIC-%s:%d", node.Address, node.Port)
	}

	// Try to URL decode the name
	if decodedName, err := url.QueryUnescape(node.Name); err == nil {
		node.Name = decodedName
	}

	// Handle SNI
	if sni := query.Get("sni"); sni != "" {
		node.Host = sni
		logger.LogDebug("使用 SNI: %s", sni)
	} else {
		node.Host = uri.Hostname() // Use hostname as SNI if not specified
		logger.LogDebug("使用主机名作为 SNI: %s", node.Host)
	}

	// Handle congestion control
	if cc := query.Get("congestion_control"); cc != "" {
		node.CC = cc
	}

	// Handle other parameters
	if alpn := query.Get("alpn"); alpn != "" {
		node.ALPN = models.StringSlice(strings.Split(alpn, ","))
	}
	if allowInsecure := query.Get("allowInsecure"); allowInsecure == "1" {
		node.SkipCertVerify = true
	}
	if udp := query.Get("udp"); udp == "1" {
		node.UDP = true
	}

	logger.LogInfo("成功解析 TUIC 节点: %s (%s:%d)", node.Name, node.Address, node.Port)
	return node, nil
}
