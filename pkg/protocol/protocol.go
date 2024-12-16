package protocol

import "github.com/shipeng101/singdns/pkg/types"

// BaseOutbound 基础出站配置
type BaseOutbound struct {
	Type           string   `json:"type"`
	Tag            string   `json:"tag"`
	Server         string   `json:"server"`
	ServerPort     int      `json:"server_port"`
	Network        []string `json:"network"`
	DomainStrategy string   `json:"domain_strategy"`
	Multiplex      bool     `json:"multiplex"`
	UDPOverTCP     bool     `json:"udp_over_tcp"`
}

// GetProtocol 获取节点协议配置
func GetProtocol(node *types.ProxyNode) (interface{}, error) {
	base := BaseOutbound{
		Tag:            "proxy",
		Server:         node.Server,
		ServerPort:     node.Port,
		Network:        []string{"tcp", "udp"},
		DomainStrategy: "prefer_ipv4",
		Multiplex:      true,
	}

	switch node.Type {
	case "ss":
		base.Type = "shadowsocks"
		return struct {
			BaseOutbound
			Method   string `json:"method"`
			Password string `json:"password"`
		}{
			BaseOutbound: base,
			Method:       node.Method,
			Password:     node.Password,
		}, nil

	case "vmess":
		base.Type = "vmess"
		return struct {
			BaseOutbound
			UUID     string `json:"uuid"`
			AlterId  int    `json:"alter_id"`
			Security string `json:"security"`
		}{
			BaseOutbound: base,
			UUID:         node.UUID,
			AlterId:      node.AlterId,
			Security:     node.Security,
		}, nil

	case "trojan":
		base.Type = "trojan"
		return struct {
			BaseOutbound
			Password string `json:"password"`
		}{
			BaseOutbound: base,
			Password:     node.Password,
		}, nil

	default:
		return nil, nil
	}
}
