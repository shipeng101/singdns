package config

import (
	"encoding/json"
	"fmt"
	"singdns/api/models"
	"singdns/api/storage"
)

// ConfigGenerator 配置生成器接口
type ConfigGenerator interface {
	// GenerateConfig 生成配置
	GenerateConfig() ([]byte, error)
	// ValidateConfig 验证配置
	ValidateConfig(config []byte) error
}

// LogConfig 日志配置
type LogConfig struct {
	Level     string `json:"level"`
	Timestamp bool   `json:"timestamp"`
	Output    string `json:"output"`
	Disabled  bool   `json:"disabled"`
}

// ExperimentalConfig 实验性功能配置
type ExperimentalConfig struct {
	ClashAPI  *ClashAPIConfig  `json:"clash_api,omitempty"`
	CacheFile *CacheFileConfig `json:"cache_file,omitempty"`
}

// ClashAPIConfig Clash API配置
type ClashAPIConfig struct {
	ExternalController       string `json:"external_controller"`
	ExternalUI               string `json:"external_ui"`
	Secret                   string `json:"secret"`
	ExternalUIDownloadURL    string `json:"external_ui_download_url,omitempty"`
	ExternalUIDownloadDetour string `json:"external_ui_download_detour,omitempty"`
	DefaultMode              string `json:"default_mode"`
}

// CacheFileConfig 缓存文件配置
type CacheFileConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

// DNSConfig DNS配置
type DNSConfig struct {
	Servers          []DNSServerConfig `json:"servers,omitempty"`
	Rules            []DNSRule         `json:"rules,omitempty"`
	Final            string            `json:"final,omitempty"`
	Strategy         string            `json:"strategy,omitempty"`
	IndependentCache bool              `json:"independent_cache,omitempty"`
	ReverseMapping   bool              `json:"reverse_mapping,omitempty"`
}

// DNSServerConfig DNS服务器配置
type DNSServerConfig struct {
	Tag             string `json:"tag,omitempty"`
	Address         string `json:"address,omitempty"`
	AddressResolver string `json:"address_resolver,omitempty"`
	AddressStrategy string `json:"address_strategy,omitempty"`
	Strategy        string `json:"strategy,omitempty"`
	Detour          string `json:"detour,omitempty"`
	ClientSubnet    string `json:"client_subnet,omitempty"`
}

// DNSRule DNS规则配置
type DNSRule struct {
	RuleSet      []string `json:"rule_set,omitempty"`
	Domain       []string `json:"domain,omitempty"`
	Server       string   `json:"server,omitempty"`
	Outbound     string   `json:"outbound,omitempty"`
	DisableCache bool     `json:"disable_cache,omitempty"`
}

// EDNSConfig EDNS配置
type EDNSConfig struct {
	ClientSubnet string `json:"client_subnet,omitempty"`
}

// InboundConfig 入站配置
type InboundConfig struct {
	Type                     string          `json:"type,omitempty"`
	Tag                      string          `json:"tag,omitempty"`
	InterfaceName            string          `json:"interface_name,omitempty"`
	Inet4Address             string          `json:"inet4_address,omitempty"`
	AutoRoute                bool            `json:"auto_route,omitempty"`
	StrictRoute              bool            `json:"strict_route,omitempty"`
	Stack                    string          `json:"stack,omitempty"`
	Sniff                    bool            `json:"sniff,omitempty"`
	SniffOverrideDestination bool            `json:"sniff_override_destination,omitempty"`
	DomainStrategy           string          `json:"domain_strategy,omitempty"`
	Listen                   string          `json:"listen,omitempty"`
	ListenPort               int             `json:"listen_port,omitempty"`
	Platform                 *PlatformConfig `json:"platform,omitempty"`
	EndpointIndependentNat   bool            `json:"endpoint_independent_nat,omitempty"`
	UDPTimeout               int             `json:"udp_timeout,omitempty"`
}

// PlatformConfig 平台特定配置
type PlatformConfig struct {
	HTTPProxy *HTTPProxyConfig `json:"http_proxy,omitempty"`
}

// HTTPProxyConfig HTTP代理配置
type HTTPProxyConfig struct {
	Enabled    bool   `json:"enabled"`
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
}

// NodeOutboundConfig 节点出站配置
type NodeOutboundConfig struct {
	Type       string `json:"type"`
	Tag        string `json:"tag"`
	Server     string `json:"server,omitempty"`
	ServerPort int    `json:"server_port,omitempty"`
	UUID       string `json:"uuid,omitempty"`
	AlterID    int    `json:"alter_id,omitempty"`
	Method     string `json:"method,omitempty"`
	Password   string `json:"password,omitempty"`
	Security   string `json:"security,omitempty"`
	Network    string `json:"network,omitempty"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	Enabled     bool           `json:"enabled,omitempty"`
	ServerName  string         `json:"server_name,omitempty"`
	Insecure    bool           `json:"insecure,omitempty"`
	ALPN        []string       `json:"alpn,omitempty"`
	MinVersion  string         `json:"min_version,omitempty"`
	MaxVersion  string         `json:"max_version,omitempty"`
	Certificate string         `json:"certificate,omitempty"`
	PrivateKey  string         `json:"private_key,omitempty"`
	ECH         *ECHConfig     `json:"ech,omitempty"`
	UTLS        *UTLSConfig    `json:"utls,omitempty"`
	Reality     *RealityConfig `json:"reality,omitempty"`
}

// ECHConfig ECH配置
type ECHConfig struct {
	Enabled     bool     `json:"enabled,omitempty"`
	PEMKey      string   `json:"pem_key,omitempty"`
	Configs     []string `json:"configs,omitempty"`
	DynamicPath string   `json:"dynamic_path,omitempty"`
}

// UTLSConfig UTLS配置
type UTLSConfig struct {
	Enabled     bool   `json:"enabled,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// RealityConfig Reality配置
type RealityConfig struct {
	Enabled   bool   `json:"enabled,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	ShortID   string `json:"short_id,omitempty"`
}

// TransportConfig 传输层配置
type TransportConfig struct {
	Type        string            `json:"type,omitempty"`
	Path        string            `json:"path,omitempty"`
	ServiceName string            `json:"service_name,omitempty"`
	MaxIdleTime string            `json:"max_idle_time,omitempty"`
	PingTimeout string            `json:"ping_timeout,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	EarlyData   *EarlyDataConfig  `json:"early_data,omitempty"`
}

// EarlyDataConfig 早期数据配置
type EarlyDataConfig struct {
	Header_name string `json:"header_name,omitempty"`
	Bytes       int    `json:"bytes,omitempty"`
}

// SelectorConfig 选择器配置
type SelectorConfig struct {
	Outbounds []string `json:"outbounds,omitempty"`
	Default   string   `json:"default,omitempty"`
}

// URLTestConfig URL测试配置
type URLTestConfig struct {
	Outbounds []string `json:"outbounds,omitempty"`
	URL       string   `json:"url,omitempty"`
	Interval  string   `json:"interval,omitempty"`
	Tolerance int      `json:"tolerance,omitempty"`
}

// OutboundConfig 出站配置
type OutboundConfig struct {
	Type           string           `json:"type"`
	Tag            string           `json:"tag"`
	Server         string           `json:"server,omitempty"`
	ServerPort     int              `json:"server_port,omitempty"`
	UUID           string           `json:"uuid,omitempty"`
	Password       string           `json:"password,omitempty"`
	Security       string           `json:"security,omitempty"`
	Method         string           `json:"method,omitempty"`
	Network        string           `json:"network,omitempty"`
	Flow           string           `json:"flow,omitempty"`
	TLS            *TLSConfig       `json:"tls,omitempty"`
	Transport      *TransportConfig `json:"transport,omitempty"`
	Multiplex      *MultiplexConfig `json:"multiplex,omitempty"`
	PacketAddr     bool             `json:"packet_addr,omitempty"`
	DomainStrategy string           `json:"domain_strategy,omitempty"`
	FallbackDelay  string           `json:"fallback_delay,omitempty"`
	Detour         string           `json:"detour,omitempty"`
	Outbounds      []string         `json:"outbounds,omitempty"`
	Default        string           `json:"default,omitempty"`
	URL            string           `json:"url,omitempty"`
	Interval       string           `json:"interval,omitempty"`
	Tolerance      int              `json:"tolerance,omitempty"`
}

// MultiplexConfig 多路复用配置
type MultiplexConfig struct {
	Enabled        bool   `json:"enabled,omitempty"`
	Protocol       string `json:"protocol,omitempty"`
	MaxConnections int    `json:"max_connections,omitempty"`
	MinStreams     int    `json:"min_streams,omitempty"`
	MaxStreams     int    `json:"max_streams,omitempty"`
}

// FilterRule 过滤规则
type FilterRule struct {
	Action   string   `json:"action"`
	Keywords []string `json:"keywords"`
}

// RouteConfig 路由配置
type RouteConfig struct {
	AutoDetectInterface bool            `json:"auto_detect_interface"`
	Final               string          `json:"final"`
	Rules               []RouteRule     `json:"rules"`
	RuleSet             []RuleSetConfig `json:"rule_set"`
}

// RouteRule 路由规则
type RouteRule struct {
	Type        string      `json:"type,omitempty"`
	Mode        string      `json:"mode,omitempty"`
	Domain      []string    `json:"domain,omitempty"`
	IPIsPrivate bool        `json:"ip_is_private,omitempty"`
	Protocol    []string    `json:"protocol,omitempty"`
	Port        int         `json:"port,omitempty"`
	RuleSet     []string    `json:"rule_set,omitempty"`
	Outbound    string      `json:"outbound"`
	Rules       []RouteRule `json:"rules,omitempty"`
	ClashMode   string      `json:"clash_mode,omitempty"`
}

// RuleSetConfig 规则集配置
type RuleSetConfig struct {
	Tag            string `json:"tag"`
	Type           string `json:"type"`
	Format         string `json:"format"`
	Path           string `json:"path,omitempty"`
	URL            string `json:"url,omitempty"`
	DownloadDetour string `json:"download_detour,omitempty"`
}

// SingBoxConfig sing-box 配置
type SingBoxConfig struct {
	Log          *LogConfig          `json:"log,omitempty"`
	DNS          *DNSConfig          `json:"dns,omitempty"`
	Inbounds     []InboundConfig     `json:"inbounds,omitempty"`
	Outbounds    []OutboundConfig    `json:"outbounds,omitempty"`
	Route        *RouteConfig        `json:"route,omitempty"`
	Experimental *ExperimentalConfig `json:"experimental,omitempty"`
}

// SingBoxGenerator sing-box 配置生成器
type SingBoxGenerator struct {
	storage storage.Storage
}

// NewSingBoxGenerator 创建 sing-box 配置生成器
func NewSingBoxGenerator(storage storage.Storage) *SingBoxGenerator {
	return &SingBoxGenerator{
		storage: storage,
	}
}

// RuleGroup 规则组定义
type RuleGroup struct {
	Name     string
	RuleSets []string
	Outbound string
}

// 预定义规则组
var defaultRuleGroups = []RuleGroup{
	{
		Name: "广告拦截",
		RuleSets: []string{
			"geosite-category-ads",
		},
		Outbound: "block",
	},
	{
		Name: "🎬 流媒体",
		RuleSets: []string{
			"geosite-netflix",
			"geosite-disney",
			"geosite-youtube",
			"geosite-spotify",
		},
		Outbound: "🎬 流媒体",
	},
	{
		Name: "💬 社交媒体",
		RuleSets: []string{
			"geosite-telegram",
			"geoip-telegram",
			"geosite-twitter",
			"geosite-facebook",
			"geosite-instagram",
		},
		Outbound: "💬 社交媒体",
	},
	{
		Name: "🔍 谷歌服务",
		RuleSets: []string{
			"geosite-google",
			"geoip-google",
		},
		Outbound: "🔍 谷歌服务",
	},
	{
		Name: "💻 开发服务",
		RuleSets: []string{
			"geosite-github",
			"geosite-microsoft",
			"geosite-apple",
			"geosite-openai",
		},
		Outbound: "💻 开发服务",
	},
	{
		Name: "国内网站",
		RuleSets: []string{
			"geosite-cn",
			"geoip-cn",
		},
		Outbound: "direct-out",
	},
	{
		Name: "国外网站",
		RuleSets: []string{
			"geosite-geolocation-!cn",
		},
		Outbound: "节点选择",
	},
}

// GenerateConfig 生成配置
func (g *SingBoxGenerator) GenerateConfig() ([]byte, error) {
	// 获取设置
	settings, err := g.storage.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}

	// 获取仪表盘设置
	dashboard := settings.GetDashboard()
	if dashboard == nil {
		dashboard = &models.DashboardSettings{
			Type: "yacd",
		}
	}

	// 根据面板类型生成路径
	dashboardPath := fmt.Sprintf("bin/web/%s", dashboard.Type)

	// 生成出站配置
	var outbounds []OutboundConfig

	// 添加直连出站
	outbounds = append(outbounds, OutboundConfig{
		Type: "direct",
		Tag:  "direct-out",
	})

	// 添加拦截出站
	outbounds = append(outbounds, OutboundConfig{
		Type: "block",
		Tag:  "block",
	})

	// 添加DNS出站
	outbounds = append(outbounds, OutboundConfig{
		Type: "dns",
		Tag:  "dns-out",
	})

	// 获取节点组
	nodeGroups, err := g.storage.GetNodeGroups()
	if err != nil {
		return nil, fmt.Errorf("get node groups: %w", err)
	}

	// 获取所有节点
	nodes, err := g.storage.GetNodes()
	if err != nil {
		return nil, fmt.Errorf("get nodes: %w", err)
	}

	// 为每个节点组匹配节点
	for i := range nodeGroups {
		nodeGroups[i].Nodes = nil // 清空现有节点
		if nodeGroups[i].Name == "全部" {
			// "全部"分组包含所有节点
			nodeGroups[i].Nodes = append(nodeGroups[i].Nodes, nodes...)
		} else {
			// 其他分组按照匹配规则添加节点
			for _, node := range nodes {
				if nodeGroups[i].MatchNode(&node) {
					nodeGroups[i].Nodes = append(nodeGroups[i].Nodes, node)
				}
			}
		}
	}

	// 创建节点组出站
	var groupOutbounds []string
	var autoGroupOutbounds []string                 // 存储自动选择的节点组
	var nodeOutboundMap = make(map[string][]string) // 存储每个组的节点出站

	// 1. 首先收集所有节点出站
	for _, group := range nodeGroups {
		if !group.Active {
			continue
		}

		// 收集组内节点
		var nodeOutbounds []string
		for _, node := range group.Nodes {
			// 生成唯一的出站标签
			nodeTag := fmt.Sprintf("%s-%s", group.ID, node.Name)
			outbound, err := g.generateOutbound(&node)
			if err != nil {
				return nil, fmt.Errorf("generate outbound for node %s: %w", node.Name, err)
			}
			outbound.Tag = nodeTag // 使用唯一的标签
			outbounds = append(outbounds, outbound)
			nodeOutbounds = append(nodeOutbounds, nodeTag)
		}

		if len(nodeOutbounds) > 0 {
			nodeOutboundMap[group.Name] = nodeOutbounds
		}
	}

	// 2. 创建所有手动节点组
	for _, group := range nodeGroups {
		if !group.Active {
			continue
		}

		nodeOutbounds := nodeOutboundMap[group.Name]
		if len(nodeOutbounds) > 0 {
			// 创建普通选择器
			groupOutbound := OutboundConfig{
				Type:      "selector",
				Tag:       group.Name,
				Outbounds: nodeOutbounds,
				Default:   nodeOutbounds[0],
			}
			outbounds = append(outbounds, groupOutbound)
			groupOutbounds = append(groupOutbounds, group.Name)
		}
	}

	// 3. 创建所有自动节点组
	for _, group := range nodeGroups {
		if !group.Active {
			continue
		}

		nodeOutbounds := nodeOutboundMap[group.Name]
		if len(nodeOutbounds) > 0 {
			// 创建自动测试版本
			autoGroupName := fmt.Sprintf("%s自动", group.Name)
			autoGroupOutbound := OutboundConfig{
				Type:      "urltest",
				Tag:       autoGroupName,
				Outbounds: nodeOutbounds,
				URL:       "http://www.gstatic.com/generate_204",
				Interval:  "300s",
				Tolerance: 50,
			}
			outbounds = append(outbounds, autoGroupOutbound)
			autoGroupOutbounds = append(autoGroupOutbounds, autoGroupName)
		}
	}

	// 添加节点选择器
	outbounds = append(outbounds, OutboundConfig{
		Type: "selector",
		Tag:  "节点选择",
		Outbounds: []string{
			"全部 🌏",
			"新加坡 🇸🇬",
			"全部 🌏自动",
			"新加坡 🇸🇬自动",
			"direct-out",
			"block",
		},
		Default: "全部 🌏",
	})

	// 添加规则组选择器
	outbounds = append(outbounds, OutboundConfig{
		Type: "selector",
		Tag:  "🎬 流媒体",
		Outbounds: []string{
			"节点选择",
			"全部 🌏",
			"新加坡 🇸🇬",
			"全部 🌏自动",
			"新加坡 🇸🇬自动",
			"direct-out",
		},
		Default: "节点选择",
	})

	outbounds = append(outbounds, OutboundConfig{
		Type: "selector",
		Tag:  "💬 社交媒体",
		Outbounds: []string{
			"节点选择",
			"全部 🌏",
			"新加坡 🇸🇬",
			"全部 🌏自动",
			"新加坡 🇸🇬自动",
			"direct-out",
		},
		Default: "节点选择",
	})

	outbounds = append(outbounds, OutboundConfig{
		Type: "selector",
		Tag:  "🔍 谷歌服务",
		Outbounds: []string{
			"节点选择",
			"全部 🌏",
			"新加坡 🇸🇬",
			"全部 🌏自动",
			"新加坡 🇸🇬自动",
			"direct-out",
		},
		Default: "节点选择",
	})

	outbounds = append(outbounds, OutboundConfig{
		Type: "selector",
		Tag:  "💻 开发服务",
		Outbounds: []string{
			"节点选择",
			"全部 🌏",
			"新加坡 🇸🇬",
			"全部 🌏自动",
			"新加坡 🇸🇬自动",
			"direct-out",
		},
		Default: "节点选择",
	})

	// 生成规则集配置
	var ruleSetConfigs []RuleSetConfig
	var rules []RouteRule

	// 添加所有规则集配置
	ruleSetMap := make(map[string]bool)
	for _, group := range defaultRuleGroups {
		for _, ruleSetID := range group.RuleSets {
			if !ruleSetMap[ruleSetID] {
				ruleSetMap[ruleSetID] = true

				// 添加规则集配置
				ruleSetConfigs = append(ruleSetConfigs, RuleSetConfig{
					Tag:    ruleSetID,
					Type:   "local",
					Format: "binary",
					Path:   fmt.Sprintf("./configs/sing-box/rules/%s.srs", ruleSetID),
				})
			}
		}
	}

	// 添加基本路由规则
	rules = append(rules, RouteRule{
		Type:     "logical",
		Mode:     "or",
		Outbound: "dns-out",
		Rules: []RouteRule{
			{Protocol: []string{"dns"}},
			{Port: 53},
		},
	})

	// 按规则组添加规则
	for _, group := range defaultRuleGroups {
		if len(group.RuleSets) > 0 {
			rules = append(rules, RouteRule{
				RuleSet:  group.RuleSets,
				Outbound: group.Outbound,
			})
		}
	}

	// 构建 DNS 配置
	dnsSettings, err := g.storage.GetDNSSettings()
	if err != nil {
		return nil, fmt.Errorf("get dns settings: %w", err)
	}

	dnsConfig := &DNSConfig{
		Servers: []DNSServerConfig{
			{
				Tag:     "dns-direct",
				Address: "223.5.5.5",
				Detour:  "direct-out",
			},
			{
				Tag:             "alidns",
				Address:         dnsSettings.Domestic,
				AddressResolver: "dns-direct",
				AddressStrategy: "prefer_ipv4",
				Strategy:        "ipv4_only",
				Detour:          "direct-out",
			},
			{
				Tag:             "google",
				Address:         dnsSettings.SingboxDNS,
				AddressResolver: "dns-direct",
				AddressStrategy: "prefer_ipv4",
				Strategy:        "ipv4_only",
				Detour:          "节点选择",
				ClientSubnet:    dnsSettings.EDNSClientSubnet,
			},
		},
		Rules: []DNSRule{
			{
				RuleSet:      []string{"geosite-cn"},
				Server:       "alidns",
				DisableCache: false,
			},
			{
				RuleSet:      []string{"geosite-geolocation-!cn"},
				Server:       "google",
				DisableCache: false,
			},
			{
				Server:   "alidns",
				Outbound: "direct-out",
			},
			{
				Server:   "google",
				Outbound: "节点选择",
			},
		},
		Final:            "dns-direct",
		Strategy:         "ipv4_only",
		IndependentCache: true,
		ReverseMapping:   true,
	}

	// 生成入站配置
	inbounds := generateInbounds(settings)

	// 构建实验性配置
	experimentalConfig := &ExperimentalConfig{
		ClashAPI: &ClashAPIConfig{
			ExternalController: "0.0.0.0:9090",
			ExternalUI:         dashboardPath,
			Secret:             "",
			DefaultMode:        "rule",
		},
		CacheFile: &CacheFileConfig{
			Enabled: false,
			Path:    "cache.db",
		},
	}

	// 构建最终配置
	config := &SingBoxConfig{
		Log: &LogConfig{
			Level:     "info",
			Timestamp: true,
			Output:    "",
			Disabled:  false,
		},
		DNS:       dnsConfig,
		Inbounds:  inbounds,
		Outbounds: outbounds,
		Route: &RouteConfig{
			AutoDetectInterface: true,
			Final:               "节点选择",
			Rules:               rules,
			RuleSet:             ruleSetConfigs,
		},
		Experimental: experimentalConfig,
	}

	// 序列化配置（使用格式化输出）
	return json.MarshalIndent(config, "", "  ")
}

// ValidateConfig 验证配置
func (g *SingBoxGenerator) ValidateConfig(config []byte) error {
	var cfg SingBoxConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}
	return nil
}

func (g *SingBoxGenerator) generateOutbound(node *models.Node) (OutboundConfig, error) {
	outbound := OutboundConfig{
		Type:       node.Type,
		Tag:        node.Name,
		Server:     node.Address,
		ServerPort: node.Port,
	}

	switch node.Type {
	case "vmess":
		outbound.UUID = node.UUID
		outbound.Security = "auto"
		if node.TLS {
			outbound.TLS = &TLSConfig{
				Enabled:    true,
				ServerName: node.SNI,
				ALPN:       node.ALPN,
				Insecure:   node.SkipCertVerify,
			}
		}
		if node.Network != "" && node.Network != "tcp" {
			outbound.Transport = &TransportConfig{
				Type: node.Network,
				Path: node.Path,
				Headers: map[string]string{
					"Host": node.Host,
				},
			}
			if node.Network == "grpc" {
				outbound.Transport.ServiceName = node.ServiceName
			}
		}
	case "vless":
		outbound.UUID = node.UUID
		if node.TLS {
			outbound.TLS = &TLSConfig{
				Enabled:    true,
				ServerName: node.SNI,
				ALPN:       node.ALPN,
				Insecure:   node.SkipCertVerify,
			}
			if node.Flow != "" {
				outbound.Flow = node.Flow
			}
			if node.PublicKey != "" {
				outbound.TLS.Reality = &RealityConfig{
					Enabled:   true,
					PublicKey: node.PublicKey,
					ShortID:   node.ShortID,
				}
			}
		}
		if node.Network != "" && node.Network != "tcp" {
			outbound.Transport = &TransportConfig{
				Type: node.Network,
				Path: node.Path,
				Headers: map[string]string{
					"Host": node.Host,
				},
			}
			if node.Network == "grpc" {
				outbound.Transport.ServiceName = node.ServiceName
			}
		}
	case "trojan":
		outbound.Password = node.Password
		if node.TLS {
			outbound.TLS = &TLSConfig{
				Enabled:    true,
				ServerName: node.SNI,
				ALPN:       node.ALPN,
				Insecure:   node.SkipCertVerify,
			}
		}
		if node.Network != "" && node.Network != "tcp" {
			outbound.Transport = &TransportConfig{
				Type: node.Network,
				Path: node.Path,
				Headers: map[string]string{
					"Host": node.Host,
				},
			}
			if node.Network == "grpc" {
				outbound.Transport.ServiceName = node.ServiceName
			}
		}
	case "shadowsocks", "ss":
		outbound.Type = "shadowsocks"
		outbound.Method = node.Method
		outbound.Password = node.Password
		if node.Plugin != "" {
			// TODO: 添加插件支持
		}
	case "direct", "block", "dns":
		// 无需额外配置
	case "selector":
		outbound.Type = "selector"
		outbound.Outbounds = node.GroupIDs
		if len(node.GroupIDs) > 0 {
			outbound.Default = node.GroupIDs[0]
		}
	case "urltest":
		outbound.Type = "urltest"
		outbound.Outbounds = node.GroupIDs
		outbound.URL = "http://www.gstatic.com/generate_204"
		outbound.Interval = "300s"
		outbound.Tolerance = 50
	default:
		return outbound, fmt.Errorf("unsupported node type: %s", node.Type)
	}

	return outbound, nil
}

// getOutboundType 转换出站类型
func getOutboundType(mode string) string {
	switch mode {
	case "select":
		return "selector"
	case "urltest":
		return "urltest"
	default:
		return "selector"
	}
}

// 生成入站配置
func generateInbounds(settings *models.Settings) []InboundConfig {
	var inbounds []InboundConfig

	// 根据入站模式生成配置
	switch settings.GetInboundMode() {
	case "tun":
		// TUN 模式配置
		inbounds = append(inbounds, InboundConfig{
			Type:                     "tun",
			Tag:                      "tun-in",
			InterfaceName:            "tun",
			Inet4Address:             "172.19.0.1/30",
			AutoRoute:                true,
			StrictRoute:              true,
			Stack:                    "system",
			Sniff:                    true,
			SniffOverrideDestination: true,
			DomainStrategy:           "ipv4_only",
			EndpointIndependentNat:   true,
			UDPTimeout:               300,
		})
	case "redirect":
		// Redirect TCP + TProxy UDP 模式配置
		inbounds = append(inbounds,
			InboundConfig{
				Type:                     "redirect",
				Tag:                      "redirect-in",
				Listen:                   "::",
				Sniff:                    true,
				SniffOverrideDestination: true,
				DomainStrategy:           "ipv4_only",
			},
			InboundConfig{
				Type:                     "tproxy",
				Tag:                      "tproxy-in",
				Listen:                   "::",
				Sniff:                    true,
				SniffOverrideDestination: true,
				DomainStrategy:           "ipv4_only",
			},
		)
	default:
		// 默认使用 TUN 模式
		inbounds = append(inbounds, InboundConfig{
			Type:                     "tun",
			Tag:                      "tun-in",
			InterfaceName:            "tun",
			Inet4Address:             "172.19.0.1/30",
			AutoRoute:                true,
			StrictRoute:              true,
			Stack:                    "system",
			Sniff:                    true,
			SniffOverrideDestination: true,
			DomainStrategy:           "ipv4_only",
			EndpointIndependentNat:   true,
			UDPTimeout:               300,
		})
	}

	return inbounds
}
