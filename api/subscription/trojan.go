package subscription

import (
	"fmt"
	"net/url"
	"singdns/api/logger"
	"singdns/api/models"
	"strconv"
	"strings"
)

// parseTrojanURL parses a Trojan URL into a Node
func parseTrojanURL(urlStr string) (*models.Node, error) {
	logger.LogDebug("开始解析 Trojan URL: %s", urlStr)

	// Parse URL
	uri, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Trojan URL: %v", err)
	}

	// Extract password
	password := uri.User.String()
	if password == "" {
		return nil, fmt.Errorf("missing password in Trojan URL")
	}

	// Extract port
	port, err := strconv.Atoi(uri.Port())
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %v", err)
	}

	// Extract name from fragment
	name := uri.Fragment

	// Parse query parameters
	query := uri.Query()

	// Create node
	node := &models.Node{
		Type:     "trojan",
		Address:  uri.Hostname(),
		Port:     port,
		Password: password,
		TLS:      true, // Trojan always uses TLS
	}

	// Handle SNI
	if sni := query.Get("sni"); sni != "" {
		node.Host = sni
		logger.LogDebug("使用 SNI: %s", sni)
	} else {
		node.Host = uri.Hostname() // Use hostname as SNI if not specified
		logger.LogDebug("使用主机名作为 SNI: %s", node.Host)
	}

	// Handle name
	if name != "" {
		node.Name = name
	} else {
		node.Name = fmt.Sprintf("Trojan-%s:%d", node.Address, node.Port)
	}

	// Try to URL decode the name
	if decodedName, err := url.QueryUnescape(node.Name); err == nil {
		node.Name = decodedName
	}

	// Handle other parameters
	if alpn := query.Get("alpn"); alpn != "" {
		node.ALPN = models.StringSlice(strings.Split(alpn, ","))
	}
	if allowInsecure := query.Get("allowInsecure"); allowInsecure == "1" {
		node.SkipCertVerify = true
	}

	logger.LogInfo("成功解析 Trojan 节点: %s (%s:%d)", node.Name, node.Address, node.Port)
	return node, nil
}
