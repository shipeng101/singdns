package subscription

import (
	"encoding/json"
	"fmt"
	"log"
	"singdns/api/protocols"
)

// SingboxConfig represents sing-box configuration
type SingboxConfig struct {
	Outbounds []SingboxOutbound `json:"outbounds"`
}

// SingboxOutbound represents a sing-box outbound
type SingboxOutbound struct {
	Tag            string                 `json:"tag"`
	Type           string                 `json:"type"`
	Server         string                 `json:"server"`
	ServerPort     int                    `json:"server_port"`
	Method         string                 `json:"method,omitempty"`
	Password       string                 `json:"password,omitempty"`
	UUID           string                 `json:"uuid,omitempty"`
	AlterID        int                    `json:"alter_id,omitempty"`
	Network        string                 `json:"network,omitempty"`
	TLS            *SingboxTLS            `json:"tls,omitempty"`
	Transport      *SingboxTransport      `json:"transport,omitempty"`
	Multiplex      *SingboxMultiplex      `json:"multiplex,omitempty"`
	PacketEncoding string                 `json:"packet_encoding,omitempty"`
	Flow           string                 `json:"flow,omitempty"`
	UoT            bool                   `json:"udp_over_tcp,omitempty"`
	Options        map[string]interface{} `json:"options,omitempty"`
}

// SingboxTLS represents TLS configuration
type SingboxTLS struct {
	Enabled         bool     `json:"enabled"`
	ServerName      string   `json:"server_name,omitempty"`
	Insecure        bool     `json:"insecure,omitempty"`
	ALPN            []string `json:"alpn,omitempty"`
	ECH             bool     `json:"ech,omitempty"`
	UtlsEnabled     bool     `json:"utls,omitempty"`
	UtlsFingerprint string   `json:"utls_fingerprint,omitempty"`
}

// SingboxTransport represents transport configuration
type SingboxTransport struct {
	Type           string            `json:"type"`
	Path           string            `json:"path,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
	ServiceName    string            `json:"service_name,omitempty"`
	MaxConcurrency int               `json:"max_concurrent_streams,omitempty"`
}

// SingboxMultiplex represents multiplex configuration
type SingboxMultiplex struct {
	Enabled bool `json:"enabled"`
}

// ParseSingbox parses sing-box subscription content
func ParseSingbox(content []byte) ([]*protocols.Node, error) {
	var config SingboxConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("invalid sing-box config: %v", err)
	}

	nodes := make([]*protocols.Node, 0, len(config.Outbounds))
	for _, outbound := range config.Outbounds {
		node, err := convertSingboxOutbound(&outbound)
		if err != nil {
			log.Printf("Warning: %v", err)
			continue
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// convertSingboxOutbound converts a sing-box outbound to a Node
func convertSingboxOutbound(outbound *SingboxOutbound) (*protocols.Node, error) {
	node := &protocols.Node{
		Name:    outbound.Tag,
		Type:    outbound.Type,
		Address: outbound.Server,
		Port:    outbound.ServerPort,
	}

	// Set TLS configuration
	if outbound.TLS != nil && outbound.TLS.Enabled {
		node.TLS = true
		node.Host = outbound.TLS.ServerName
	}

	// Set transport configuration
	if outbound.Transport != nil {
		node.Network = outbound.Transport.Type
		switch outbound.Transport.Type {
		case "ws":
			node.Path = outbound.Transport.Path
			if h := outbound.Transport.Headers; h != nil {
				if host, ok := h["Host"]; ok {
					node.Host = host
				}
			}
		case "grpc":
			node.Path = outbound.Transport.ServiceName
		}
	}

	switch outbound.Type {
	case "shadowsocks":
		node.Type = "ss"
		node.Method = outbound.Method
		node.Password = outbound.Password

	case "vmess":
		node.UUID = outbound.UUID
		node.AlterId = outbound.AlterID

	case "trojan":
		node.Password = outbound.Password

	case "vless":
		node.Type = "vless"
		node.UUID = outbound.UUID
		if outbound.Flow != "" {
			node.Flow = outbound.Flow
		}

	case "hysteria2":
		node.Type = "hy2"
		node.Password = outbound.Password
		if opts := outbound.Options; opts != nil {
			if up, ok := opts["up"].(string); ok {
				node.Up = up
			}
			if down, ok := opts["down"].(string); ok {
				node.Down = down
			}
		}

	case "tuic":
		node.Type = "tuic"
		node.UUID = outbound.UUID
		node.Password = outbound.Password
		if opts := outbound.Options; opts != nil {
			if cc, ok := opts["congestion_control"].(string); ok {
				node.CC = cc
			}
		}

	default:
		return nil, fmt.Errorf("unsupported outbound type: %s", outbound.Type)
	}

	return node, nil
}
