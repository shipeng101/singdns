package subscription

import (
	"fmt"
	"net/url"
	"singdns/api/logger"
	"singdns/api/models"
	"strconv"
	"strings"
)

// parseHysteria2URL parses a Hysteria2 URL into a Node
func parseHysteria2URL(urlStr string) (*models.Node, error) {
	logger.LogDebug("开始解析 Hysteria2 URL: %s", urlStr)

	// Parse URL
	uri, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Hysteria2 URL: %v", err)
	}

	// Extract password
	password := uri.User.String()
	if password == "" {
		return nil, fmt.Errorf("missing password in Hysteria2 URL")
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
		Type:     "hy2",
		Address:  uri.Hostname(),
		Port:     port,
		Password: password,
		TLS:      true, // Hysteria2 always uses TLS
	}

	// Handle name
	if name := uri.Fragment; name != "" {
		node.Name = name
	} else {
		node.Name = fmt.Sprintf("HY2-%s:%d", node.Address, node.Port)
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

	// Handle bandwidth settings
	if up := query.Get("up"); up != "" {
		node.Up = up
	}
	if down := query.Get("down"); down != "" {
		node.Down = down
	}

	// Handle obfs settings
	if obfs := query.Get("obfs"); obfs != "" {
		node.Obfs = obfs
	}

	// Handle other parameters
	if alpn := query.Get("alpn"); alpn != "" {
		node.ALPN = models.StringSlice(strings.Split(alpn, ","))
	}
	if allowInsecure := query.Get("insecure"); allowInsecure == "1" {
		node.SkipCertVerify = true
	}

	logger.LogInfo("成功解析 Hysteria2 节点: %s (%s:%d)", node.Name, node.Address, node.Port)
	return node, nil
}
