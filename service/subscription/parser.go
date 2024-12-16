package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/shipeng101/singdns/pkg/types"
	"github.com/shipeng101/singdns/service/system"
	"go.uber.org/zap"
)

// Parser 订阅解析器接口
type Parser interface {
	// 解析订阅内容为节点列表
	Parse(content string) ([]*types.ProxyNode, error)
}

// parser 订阅解析器实现
type parser struct{}

// NewParser 创建订阅解析器
func NewParser() Parser {
	return &parser{}
}

// Parse 解析订阅内容为节点列表
func (p *parser) Parse(content string) ([]*types.ProxyNode, error) {
	// Base64解码
	decoded, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		// 尝试URL安全的Base64
		decoded, err = base64.URLEncoding.DecodeString(content)
		if err != nil {
			// 可能不是Base64编码
			decoded = []byte(content)
		}
	}

	// 按行分割
	lines := strings.Split(string(decoded), "\n")
	var nodes []*types.ProxyNode

	// 解析每一行
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 根据协议类型解析
		var node *types.ProxyNode
		var err error

		switch {
		case strings.HasPrefix(line, "vmess://"):
			node, err = p.parseVmess(line)
		case strings.HasPrefix(line, "ss://"):
			node, err = p.parseSS(line)
		case strings.HasPrefix(line, "ssr://"):
			node, err = p.parseSSR(line)
		case strings.HasPrefix(line, "trojan://"):
			node, err = p.parseTrojan(line)
		default:
			continue
		}

		if err != nil {
			system.Warn("解析节点失败",
				zap.String("line", line),
				zap.Error(err))
			continue
		}

		if node != nil {
			// 生成节点ID
			node.ID = fmt.Sprintf("%d", time.Now().UnixNano())
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// parseVmess 解析Vmess链接
func (p *parser) parseVmess(link string) (*types.ProxyNode, error) {
	// 移除前缀
	vmessStr := strings.TrimPrefix(link, "vmess://")

	// Base64解码
	decoded, err := base64.StdEncoding.DecodeString(vmessStr)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(vmessStr)
		if err != nil {
			return nil, err
		}
	}

	// 解析JSON
	var vmess map[string]interface{}
	if err := json.Unmarshal(decoded, &vmess); err != nil {
		return nil, err
	}

	// 创建节点
	node := &types.ProxyNode{
		Type:     "vmess",
		Name:     fmt.Sprintf("%v", vmess["ps"]),
		Server:   fmt.Sprintf("%v", vmess["add"]),
		Port:     int(vmess["port"].(float64)),
		UUID:     fmt.Sprintf("%v", vmess["id"]),
		AlterId:  int(vmess["aid"].(float64)),
		Security: fmt.Sprintf("%v", vmess["security"]),
	}

	// 可选参数
	if net, ok := vmess["net"]; ok {
		node.Network = fmt.Sprintf("%v", net)
	}
	if tls, ok := vmess["tls"]; ok {
		node.TLS = tls == "tls"
	}
	if path, ok := vmess["path"]; ok {
		node.Path = fmt.Sprintf("%v", path)
	}
	if host, ok := vmess["host"]; ok {
		node.Host = fmt.Sprintf("%v", host)
	}

	return node, nil
}

// parseSS 解析Shadowsocks链接
func (p *parser) parseSS(link string) (*types.ProxyNode, error) {
	// 移除前缀
	ssStr := strings.TrimPrefix(link, "ss://")

	// 解析名称
	hashIndex := strings.LastIndex(ssStr, "#")
	var name string
	if hashIndex != -1 {
		name, _ = url.QueryUnescape(ssStr[hashIndex+1:])
		ssStr = ssStr[:hashIndex]
	}

	// Base64解码
	decoded, err := base64.StdEncoding.DecodeString(ssStr)
	if err != nil {
		// 尝试URL安全的Base64
		decoded, err = base64.URLEncoding.DecodeString(ssStr)
		if err != nil {
			// 可能已经是解码后的格式
			parts := strings.Split(ssStr, "@")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid ss link format")
			}
			methodAndPassword, err := base64.StdEncoding.DecodeString(parts[0])
			if err != nil {
				methodAndPassword, err = base64.URLEncoding.DecodeString(parts[0])
				if err != nil {
					return nil, err
				}
			}
			decoded = []byte(string(methodAndPassword) + "@" + parts[1])
		}
	}

	// 解析地址和端口
	parts := strings.Split(string(decoded), "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ss link format")
	}

	// 解析加密方式和密码
	methodAndPassword := strings.Split(parts[0], ":")
	if len(methodAndPassword) != 2 {
		return nil, fmt.Errorf("invalid ss method and password format")
	}

	// 解析服务器和端口
	serverAndPort := strings.Split(parts[1], ":")
	if len(serverAndPort) != 2 {
		return nil, fmt.Errorf("invalid ss server and port format")
	}

	// 解析端口
	port := 0
	fmt.Sscanf(serverAndPort[1], "%d", &port)

	// 创建节点
	node := &types.ProxyNode{
		Type:     "ss",
		Name:     name,
		Server:   serverAndPort[0],
		Port:     port,
		Method:   methodAndPassword[0],
		Password: methodAndPassword[1],
	}

	return node, nil
}

// parseSSR 解析ShadowsocksR链接
func (p *parser) parseSSR(link string) (*types.ProxyNode, error) {
	// 移除前缀
	ssrStr := strings.TrimPrefix(link, "ssr://")

	// Base64解码
	decoded, err := base64.StdEncoding.DecodeString(ssrStr)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(ssrStr)
		if err != nil {
			return nil, err
		}
	}

	// 解析基本参数
	parts := strings.Split(string(decoded), ":")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid ssr link format")
	}

	// 解析服务器和端口
	server := parts[0]
	port := 0
	fmt.Sscanf(parts[1], "%d", &port)

	// 解析协议参数
	protocol := parts[2]
	method := parts[3]
	obfs := parts[4]

	// 解析密码和其他参数
	paramStr := strings.Split(parts[5], "/?")
	if len(paramStr) != 2 {
		return nil, fmt.Errorf("invalid ssr params format")
	}

	// 解析密码
	password, err := base64.StdEncoding.DecodeString(paramStr[0])
	if err != nil {
		return nil, err
	}

	// 解析其他参数
	params := make(map[string]string)
	for _, p := range strings.Split(paramStr[1], "&") {
		param := strings.Split(p, "=")
		if len(param) != 2 {
			continue
		}
		value, err := base64.StdEncoding.DecodeString(param[1])
		if err != nil {
			continue
		}
		params[param[0]] = string(value)
	}

	// 创建节点
	node := &types.ProxyNode{
		Type:     "ssr",
		Name:     params["remarks"],
		Server:   server,
		Port:     port,
		Protocol: protocol,
		Method:   method,
		Obfs:     obfs,
		Password: string(password),
	}

	return node, nil
}

// parseTrojan 解析Trojan链接
func (p *parser) parseTrojan(link string) (*types.ProxyNode, error) {
	// 解析URL
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	// 解析密码
	password := strings.TrimPrefix(u.User.String(), ":")

	// 解析端口
	port := u.Port()
	portNum := 0
	fmt.Sscanf(port, "%d", &portNum)

	// 创建节点
	node := &types.ProxyNode{
		Type:     "trojan",
		Name:     u.Fragment,
		Server:   u.Hostname(),
		Port:     portNum,
		Password: password,
	}

	// 解析查询参数
	query := u.Query()
	if sni := query.Get("sni"); sni != "" {
		node.SNI = sni
	}
	if alpn := query.Get("alpn"); alpn != "" {
		node.ALPN = strings.Split(alpn, ",")
	}
	if allowInsecure := query.Get("allowInsecure"); allowInsecure == "1" {
		node.AllowInsecure = true
	}

	return node, nil
}
