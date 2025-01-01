package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"singdns/api/models"
	"singdns/api/storage"
	"strings"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// ConfigGenerator 配置生成器接口
type ConfigGenerator interface {
	// GenerateConfig 生成配置
	GenerateConfig() ([]byte, error)
	// ValidateConfig 验证配置
	ValidateConfig(config []byte) error
}

// MosDNSConfig MosDNS配置结构
type MosDNSConfig struct {
	Log struct {
		Level      string `yaml:"level"`
		Production bool   `yaml:"production"`
		File       string `yaml:"file"`
	} `yaml:"log"`
	Plugins []Plugin `yaml:"plugins"`
}

// Plugin MosDNS插件配置
type Plugin struct {
	Tag  string      `yaml:"tag"`
	Type string      `yaml:"type"`
	Args interface{} `yaml:"args"`
}

// ForwardRule MosDNS转发规则
type ForwardRule struct {
	DomainSuffix  string   `yaml:"domain-suffix,omitempty"`
	DomainKeyword string   `yaml:"domain-keyword,omitempty"`
	DomainRegex   string   `yaml:"domain-regex,omitempty"`
	GeoIP         string   `yaml:"geoip,omitempty"`
	GeoSite       string   `yaml:"geosite,omitempty"`
	IPCidr        []string `yaml:"ip-cidr,omitempty"`
	Forward       string   `yaml:"forward"`
}

// Experimental 实验性功能配置
type Experimental struct {
	CacheFile CacheFile `json:"cache_file"`
	ClashAPI  ClashAPI  `json:"clash_api"`
}

// CacheFile 缓存文件配置
type CacheFile struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
}

// ClashAPI Clash API配置
type ClashAPI struct {
	ExternalController string `json:"external_controller"`
	ExternalUI         string `json:"external_ui"`
	Secret             string `json:"secret"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level     string `json:"level"`
	Timestamp bool   `json:"timestamp"`
}

// DNSConfig DNS 配置
type DNSConfig struct {
	Servers        []DNSServerConfig `json:"servers"`
	Strategy       string            `json:"strategy"`
	ReverseMapping bool              `json:"reverse_mapping"`
	Rules          []DNSRuleConfig   `json:"rules"`
	Final          string            `json:"final"`
}

// DNSServerConfig DNS 服务器配置
type DNSServerConfig struct {
	Tag             string `json:"tag,omitempty"`
	Address         string `json:"address,omitempty"`
	AddressResolver string `json:"address_resolver,omitempty"`
	AddressStrategy string `json:"address_strategy,omitempty"`
	Strategy        string `json:"strategy,omitempty"`
	Detour          string `json:"detour,omitempty"`
	ClientSubnet    string `json:"client_subnet,omitempty"`
}

// DNSRuleConfig DNS 规则配置
type DNSRuleConfig struct {
	Domain []string `json:"domain,omitempty"`
	Server string   `json:"server"`
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

// InboundConfig 入站配置
type InboundConfig struct {
	Type                     string `json:"type"`
	Tag                      string `json:"tag"`
	Listen                   string `json:"listen"`
	ListenPort               int    `json:"listen_port"`
	DomainStrategy           string `json:"domain_strategy"`
	Sniff                    bool   `json:"sniff"`
	SniffOverrideDestination bool   `json:"sniff_override_destination"`
}

// OutboundConfig 出站配置
type OutboundConfig struct {
	Type       string           `json:"type"`
	Tag        string           `json:"tag"`
	Server     string           `json:"server,omitempty"`
	ServerPort int              `json:"server_port,omitempty"`
	UUID       string           `json:"uuid,omitempty"`
	AlterID    int              `json:"alter_id,omitempty"`
	Method     string           `json:"method,omitempty"`
	Password   string           `json:"password,omitempty"`
	Outbounds  []string         `json:"outbounds,omitempty"`
	TLS        *TLSConfig       `json:"tls,omitempty"`
	Transport  *TransportConfig `json:"transport,omitempty"`
	// VMess 特定配置
	Security            string `json:"security,omitempty"`
	GlobalPadding       bool   `json:"global_padding,omitempty"`
	AuthenticatedLength bool   `json:"authenticated_length,omitempty"`
	Network             string `json:"network,omitempty"`
	PacketEncoding      string `json:"packet_encoding,omitempty"`
	// URLTest 特定配置
	Interval  string `json:"interval,omitempty"`
	Tolerance int    `json:"tolerance,omitempty"`
}

// RouteConfig 路由配置
type RouteConfig struct {
	Rules   []RouteRule     `json:"rules,omitempty"`
	Final   string          `json:"final,omitempty"`
	RuleSet []RuleSetConfig `json:"rule_set,omitempty"`
}

// RouteRule 路由规则
type RouteRule struct {
	RuleSet                  []string     `json:"rule_set,omitempty"`
	Domain                   []string     `json:"domain,omitempty"`
	IPCidr                   []string     `json:"ip_cidr,omitempty"`
	IPIsPrivate              bool         `json:"ip_is_private,omitempty"`
	Protocol                 []string     `json:"protocol,omitempty"`
	RuleSetIPCidrMatchSource bool         `json:"rule_set_ipcidr_match_source,omitempty"`
	Outbound                 string       `json:"outbound,omitempty"`
	Inbound                  []string     `json:"inbound,omitempty"`
	Action                   string       `json:"action,omitempty"`
	DomainStrategy           string       `json:"domain_strategy,omitempty"`
	Sniff                    *SniffConfig `json:"sniff,omitempty"`
}

// RuleSetConfig 规则集配置
type RuleSetConfig struct {
	Tag    string `json:"tag"`
	Type   string `json:"type"`
	Format string `json:"format"`
	Path   string `json:"path,omitempty"`
}

// MosDNSGenerator MosDNS配置生成器
type MosDNSGenerator struct {
	Config  MosDNSConfig
	storage storage.Storage
}

// NewMosDNSGenerator 创建MosDNS配置生成器
func NewMosDNSGenerator() *MosDNSGenerator {
	g := &MosDNSGenerator{}
	g.Config.Log.Level = "info"
	g.Config.Log.Production = true

	g.Config.Plugins = []Plugin{
		// 匹配器
		{
			Tag:  "direct_domain",
			Type: "domain_set",
			Args: map[string]interface{}{
				"files": []string{
					"rules/china_domain_list.txt",
				},
			},
		},
		{
			Tag:  "direct_ip",
			Type: "ip_set",
			Args: map[string]interface{}{
				"files": []string{
					"rules/china_ip_list.txt",
				},
			},
		},
		// hosts 配置
		{
			Tag:  "hosts",
			Type: "hosts",
			Args: map[string]interface{}{
				"files": []string{
					"rules/hosts.txt",
				},
			},
		},
		// 缓存插件
		{
			Tag:  "cache",
			Type: "cache",
			Args: map[string]interface{}{
				"size":           65536,
				"lazy_cache_ttl": 86400,
				"dump_file":      "logs/cache.dump",
			},
		},
		// 远程转发插件
		{
			Tag:  "remote_forward",
			Type: "forward",
			Args: map[string]interface{}{
				"concurrent": 2,
				"upstreams": []map[string]interface{}{
					{
						"addr":         "127.0.0.1:7890",
						"idle_timeout": 30,
					},
				},
			},
		},
		// 本地转发插件
		{
			Tag:  "local_forward",
			Type: "forward",
			Args: map[string]interface{}{
				"concurrent": 2,
				"upstreams": []map[string]interface{}{
					{
						"addr":         "udp://223.5.5.5",
						"idle_timeout": 30,
					},
					{
						"addr":         "udp://119.29.29.29",
						"idle_timeout": 30,
					},
				},
			},
		},
		// 本地序列
		{
			Tag:  "local_sequence",
			Type: "sequence",
			Args: []map[string]interface{}{
				{
					"exec": "$local_forward",
				},
				{
					"matches": "resp_ip $direct_ip",
					"exec":    "accept",
				},
				{
					"exec": "drop_resp",
				},
			},
		},
		// 远程序列
		{
			Tag:  "remote_sequence",
			Type: "sequence",
			Args: []map[string]interface{}{
				{
					"exec": "$remote_forward",
				},
				{
					"exec": "accept",
				},
			},
		},
		// 故障转移
		{
			Tag:  "fallback",
			Type: "fallback",
			Args: map[string]interface{}{
				"primary":        "local_sequence",
				"secondary":      "remote_sequence",
				"threshold":      500,
				"always_standby": true,
			},
		},
		// 主序列
		{
			Tag:  "main_sequence",
			Type: "sequence",
			Args: []map[string]interface{}{
				{
					"exec": "$hosts",
				},
				{
					"matches": "has_resp",
					"exec":    "accept",
				},
				{
					"matches": "qtype 12 65",
					"exec":    "reject 0",
				},
				{
					"exec": "prefer_ipv4",
				},
				{
					"exec": "$cache",
				},
				{
					"matches": "has_resp",
					"exec":    "accept",
				},
				{
					"matches": "qname $direct_domain",
					"exec":    "$local_forward",
				},
				{
					"matches": "has_resp",
					"exec":    "accept",
				},
				{
					"exec": "$fallback",
				},
			},
		},
		// UDP 服务器
		{
			Tag:  "udp_server",
			Type: "udp_server",
			Args: map[string]interface{}{
				"entry":  "main_sequence",
				"listen": ":5354",
			},
		},
		// TCP 服务器
		{
			Tag:  "tcp_server",
			Type: "tcp_server",
			Args: map[string]interface{}{
				"entry":  "main_sequence",
				"listen": ":5354",
			},
		},
	}

	return g
}

// SetStorage 设置存储
func (g *MosDNSGenerator) SetStorage(storage storage.Storage) {
	g.storage = storage
}

// GenerateConfig 生成MosDNS配置
func (g *MosDNSGenerator) GenerateConfig() ([]byte, error) {
	// 使用已经在 NewMosDNSGenerator 中定义的配置
	data, err := yaml.Marshal(g.Config)
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	return data, nil
}

// ValidateConfig 验证MosDNS配置
func (g *MosDNSGenerator) ValidateConfig(config []byte) error {
	return yaml.Unmarshal(config, &g.Config)
}

// SingBoxGenerator sing-box 配置生成器
type SingBoxGenerator struct {
	settings   *models.Settings
	storage    storage.Storage
	configPath string
}

// NewSingBoxGenerator 创建一个新的 sing-box 配置生成器
func NewSingBoxGenerator(settings *models.Settings, storage storage.Storage, configPath string) *SingBoxGenerator {
	return &SingBoxGenerator{
		settings:   settings,
		storage:    storage,
		configPath: configPath,
	}
}

// SetStorage 设置数据库连接
func (g *SingBoxGenerator) SetStorage(storage storage.Storage) {
	g.storage = storage
	// 从数据库获取设置
	if settings, err := storage.GetSettings(); err == nil && settings != nil {
		// 确保设置了默认仪表盘
		dashboard := settings.GetDashboard()
		if dashboard == nil {
			dashboard = &models.DashboardSettings{
				Type:   "yacd",
				Path:   "bin/web/yacd",
				Secret: "",
			}
			settings.SetDashboard(dashboard)
		} else {
			// 根据仪表盘类型设置路径
			switch dashboard.Type {
			case "metacubexd":
				dashboard.Path = "bin/web/metacubexd"
			default:
				dashboard.Path = "bin/web/yacd"
			}
			settings.SetDashboard(dashboard)
		}
		g.settings = settings

		// 重新生成配置文件
		if err := g.regenerateConfig(); err != nil {
			fmt.Printf("failed to regenerate config: %v\n", err)
		}
	}
}

// UpdateRuleSets 更新规则集
func (g *SingBoxGenerator) UpdateRuleSets() error {
	// 创建规则集目录
	if err := os.MkdirAll("configs/sing-box/rules", 0755); err != nil {
		return fmt.Errorf("failed to create rules directory: %v", err)
	}

	// 读取规则文件夹
	files, err := os.ReadDir("configs/sing-box/rules")
	if err != nil {
		return fmt.Errorf("failed to read rules directory: %v", err)
	}

	// 更新规则集
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".srs") {
			// 获取规则集标签（文件名去掉 .srs 后缀）
			tag := strings.TrimSuffix(file.Name(), ".srs")

			// 获取规则集信息
			ruleSet, err := g.storage.GetRuleSetByTag(tag)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("failed to get rule set %s: %v", tag, err)
			}

			if ruleSet == nil {
				continue
			}

			// 下载规则集
			resp, err := http.Get(ruleSet.URL)
			if err != nil {
				return fmt.Errorf("failed to download rule set %s: %v", tag, err)
			}
			defer resp.Body.Close()

			// 保存规则集
			filePath := filepath.Join("configs/sing-box/rules", fmt.Sprintf("%s.srs", tag))
			file, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %v", filePath, err)
			}
			defer file.Close()

			if _, err := io.Copy(file, resp.Body); err != nil {
				return fmt.Errorf("failed to save rule set %s: %v", tag, err)
			}

			// 更新规则集记录
			ruleSet.UpdatedAt = time.Now()
			if err := g.storage.SaveRuleSet(ruleSet); err != nil {
				return fmt.Errorf("failed to update rule set %s: %v", tag, err)
			}
		}
	}

	return nil
}

// SaveRuleSet 保存规则集
func (g *SingBoxGenerator) SaveRuleSet(ruleSet *models.RuleSet) error {
	// 检查规则集是否存在
	existingRuleSet, err := g.storage.GetRuleSetByTag(ruleSet.Tag)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to get rule set: %v", err)
	}

	if existingRuleSet != nil {
		// 更新现有规则集的启用状态
		existingRuleSet.Enabled = ruleSet.Enabled
		existingRuleSet.UpdatedAt = time.Now()
		if err := g.storage.SaveRuleSet(existingRuleSet); err != nil {
			return fmt.Errorf("failed to update rule set: %v", err)
		}
	} else {
		// 创建新的规则集
		ruleSet.ID = uuid.New().String()
		ruleSet.UpdatedAt = time.Now()
		if err := g.storage.SaveRuleSet(ruleSet); err != nil {
			return fmt.Errorf("failed to save rule set: %v", err)
		}
	}

	// 从数据库获取最新设置
	if settings, err := g.storage.GetSettings(); err == nil && settings != nil {
		g.settings = settings
	}

	// 重新生成配置文件
	if err := g.regenerateConfig(); err != nil {
		return fmt.Errorf("failed to regenerate config: %v", err)
	}

	return nil
}

// DeleteRuleSet 删除规则集
func (g *SingBoxGenerator) DeleteRuleSet(id string) error {
	// 从数据库中获取规则集信息
	ruleSet, err := g.storage.GetRuleSetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get rule set: %v", err)
	}

	// 删除规则文件
	filePath := filepath.Join("configs/sing-box/rules", fmt.Sprintf("%s.srs", ruleSet.Tag))
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete rule set file: %v", err)
	}

	// 从数据库中删除规则集记录
	if err := g.storage.DeleteRuleSet(id); err != nil {
		return fmt.Errorf("failed to delete rule set from database: %v", err)
	}

	return nil
}

// GenerateConfig 生成配置
func (g *SingBoxGenerator) GenerateConfig() ([]byte, error) {
	if g.storage == nil {
		return nil, fmt.Errorf("storage is not initialized")
	}

	// 获取节点
	nodes, err := g.storage.GetNodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %v", err)
	}

	// 获取节点组
	groups, err := g.storage.GetNodeGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to get node groups: %v", err)
	}

	// 创建基础配置
	config := &SingBoxConfig{
		Log: &LogConfig{
			Level:     g.settings.LogLevel,
			Timestamp: true,
		},
		DNS: &DNSConfig{
			Servers: []DNSServerConfig{
				{
					Tag:             "google",
					Address:         "https://dns.google/dns-query",
					AddressResolver: "system",
					Strategy:        "ipv4_only",
					Detour:          "🚀 节点选择",
					ClientSubnet:    "39.85.255.234",
				},
				{
					Tag:             "local",
					Address:         "https://doh.pub/dns-query",
					AddressResolver: "system",
					Strategy:        "ipv4_only",
					Detour:          "🎯 全球直连",
					ClientSubnet:    "39.85.255.234",
				},
				{
					Tag:     "system",
					Address: "223.5.5.5",
					Detour:  "🎯 全球直连",
				},
			},
			Rules: []DNSRuleConfig{
				{
					Domain: []string{"geosite:cn"},
					Server: "local",
				},
			},
			Strategy: "ipv4_only",
			Final:    "google",
		},
		Inbounds: []InboundConfig{
			{
				Type:                     "mixed",
				Tag:                      "mixed-in",
				Listen:                   "::",
				ListenPort:               g.settings.ProxyPort,
				DomainStrategy:           "ipv4_only",
				Sniff:                    true,
				SniffOverrideDestination: true,
			},
		},
		Route: &RouteConfig{
			Rules: []RouteRule{
				{
					Protocol: []string{"dns"},
					Outbound: "🚀 节点选择",
				},
				{
					Protocol: []string{"quic"},
					Outbound: "🚀 节点选择",
				},
				{
					RuleSet:  []string{"geosite-category-ads"},
					Outbound: "🛑 广告拦截",
				},
				{
					RuleSet:  []string{"geosite-cn"},
					Outbound: "🎯 全球直连",
				},
				{
					RuleSet:  []string{"geoip-cn"},
					Outbound: "🎯 全球直连",
				},
			},
			Final: "🐠 漏网之鱼",
			RuleSet: []RuleSetConfig{
				{
					Tag:    "geoip-cn",
					Type:   "local",
					Format: "binary",
					Path:   "configs/sing-box/rules/geoip-cn.srs",
				},
				{
					Tag:    "geosite-cn",
					Type:   "local",
					Format: "binary",
					Path:   "configs/sing-box/rules/geosite-cn.srs",
				},
				{
					Tag:    "geosite-category-ads",
					Type:   "local",
					Format: "binary",
					Path:   "configs/sing-box/rules/geosite-category-ads.srs",
				},
			},
		},
		Experimental: &ExperimentalConfig{
			CacheFile: CacheFile{
				Enabled: true,
				Path:    "configs/sing-box/cache.db",
			},
			ClashAPI: ClashAPI{
				ExternalController: "127.0.0.1:9090",
				ExternalUI:         g.settings.GetDashboard().Path,
				Secret:             g.settings.GetDashboard().Secret,
			},
		},
	}

	// 1. 初始化基础出站配置
	baseOutbounds := []OutboundConfig{
		{
			Type: "direct",
			Tag:  "🎯 全球直连",
		},
		{
			Type: "block",
			Tag:  "🛑 广告拦截",
		},
	}

	// 2. 收集所有节点的出站配置
	nodeOutbounds := make([]OutboundConfig, 0)
	for _, node := range nodes {
		outbound := OutboundConfig{
			Type:       node.Type,
			Tag:        node.Name,
			Server:     node.Address,
			ServerPort: node.Port,
			UUID:       node.UUID,
			AlterID:    node.AlterID,
			Method:     node.Method,
			Password:   node.Password,
		}

		if node.TLS {
			outbound.TLS = &TLSConfig{
				Enabled:    true,
				ServerName: node.SNI,
				Insecure:   node.SkipCertVerify,
				ALPN:       node.ALPN,
			}
		}

		if node.Network != "" && node.Network != "tcp" {
			transport := &TransportConfig{
				Type: node.Network,
			}
			switch node.Network {
			case "ws", "websocket":
				transport.Path = node.Path
				if node.Host != "" {
					transport.Headers = map[string]string{
						"Host": node.Host,
					}
				}
			case "grpc":
				transport.ServiceName = node.Path
			}
			outbound.Transport = transport
		}

		if node.Type == "vmess" {
			outbound.Security = "auto"
			outbound.GlobalPadding = false
			outbound.AuthenticatedLength = true
			outbound.Network = "tcp"
			if node.UDP {
				outbound.PacketEncoding = "packetaddr"
			}
		}

		nodeOutbounds = append(nodeOutbounds, outbound)
	}

	// 3. 收集地区节点组
	regionGroups := make([]string, 0)
	regionGroupOutbounds := make([]OutboundConfig, 0)
	autoSelectOutbounds := make([]OutboundConfig, 0)

	// 创建节点到组的映射
	nodeGroupMap := make(map[string][]string) // key: 节点ID, value: 组名列表
	assignedNodes := make(map[string]bool)    // 记录已分配的节点

	// 按优先级顺序处理节点组（除了"其他节点"组）
	for _, group := range groups {
		if !group.Active || group.Tag == "others" {
			continue
		}

		groupNodes := make([]string, 0)
		for _, node := range nodes {
			// 跳过已分配的节点
			if assignedNodes[node.ID] {
				continue
			}

			// 检查节点是否属于当前组
			for _, nodeGroup := range node.NodeGroups {
				if nodeGroup.ID == group.ID {
					groupNodes = append(groupNodes, node.Name)
					assignedNodes[node.ID] = true
					nodeGroupMap[node.ID] = append(nodeGroupMap[node.ID], group.Name)
					break
				}
			}
		}

		if len(groupNodes) > 0 {
			// 添加自动选择组
			autoOutbound := OutboundConfig{
				Type:      "urltest",
				Tag:       group.Name + " - 自动选择",
				Outbounds: groupNodes,
				Interval:  "300s",
				Tolerance: 50,
			}
			autoSelectOutbounds = append(autoSelectOutbounds, autoOutbound)

			// 添加地区节点组
			groupOutboundNodes := []string{"🎯 全球直连"}
			groupOutboundNodes = append(groupOutboundNodes, group.Name+" - 自动选择") // 添加自动选择组
			groupOutboundNodes = append(groupOutboundNodes, groupNodes...)

			groupOutbound := OutboundConfig{
				Type:      "selector",
				Tag:       group.Name,
				Outbounds: groupOutboundNodes,
			}
			regionGroupOutbounds = append(regionGroupOutbounds, groupOutbound)
			regionGroups = append(regionGroups, group.Name)
		}
	}

	// 处理"其他节点"组
	for _, group := range groups {
		if !group.Active || group.Tag != "others" {
			continue
		}

		othersNodes := make([]string, 0)
		for _, node := range nodes {
			// 只添加未分配的节点
			if !assignedNodes[node.ID] {
				othersNodes = append(othersNodes, node.Name)
				assignedNodes[node.ID] = true
			}
		}

		if len(othersNodes) > 0 {
			// 添加自动选择组
			autoOutbound := OutboundConfig{
				Type:      "urltest",
				Tag:       group.Name + " - 自动选择",
				Outbounds: othersNodes,
				Interval:  "300s",
				Tolerance: 50,
			}
			autoSelectOutbounds = append(autoSelectOutbounds, autoOutbound)

			// 添加其他节点组
			othersOutboundNodes := []string{"🎯 全球直连"}
			othersOutboundNodes = append(othersOutboundNodes, group.Name+" - 自动选择") // 添加自动选择组
			othersOutboundNodes = append(othersOutboundNodes, othersNodes...)

			othersOutbound := OutboundConfig{
				Type:      "selector",
				Tag:       group.Name,
				Outbounds: othersOutboundNodes,
			}
			regionGroupOutbounds = append(regionGroupOutbounds, othersOutbound)
			regionGroups = append(regionGroups, group.Name)
		}
	}

	// 4. 按照指定顺序组装最终的出站配置
	outbounds := []string{"🎯 全球直连"}
	outbounds = append(outbounds, regionGroups...)

	// 创建节点选择器
	nodeSelector := OutboundConfig{
		Type:      "selector",
		Tag:       "🚀 节点选择",
		Outbounds: outbounds,
	}

	config.Outbounds = []OutboundConfig{
		// 1. 节点选择器
		nodeSelector,
	}

	// 2. 添加基础出站
	config.Outbounds = append(config.Outbounds, baseOutbounds...)

	// 3. 添加所有节点
	config.Outbounds = append(config.Outbounds, nodeOutbounds...)

	// 4. 添加地区节点组
	config.Outbounds = append(config.Outbounds, regionGroupOutbounds...)

	// 5. 添加自动选择组
	config.Outbounds = append(config.Outbounds, autoSelectOutbounds...)

	// 6. 漏网之鱼
	config.Outbounds = append(config.Outbounds, OutboundConfig{
		Type:      "selector",
		Tag:       "🐠 漏网之鱼",
		Outbounds: []string{"🚀 节点选择", "🎯 全球直连"},
	})

	// 添加内置规则
	config.Route.Rules = append(config.Route.Rules, []RouteRule{
		{
			RuleSet:  []string{"geosite-category-ads"},
			Outbound: "🛑 广告拦截",
		},
		{
			RuleSet:  []string{"geosite-cn"},
			Outbound: "🎯 全球直连",
		},
		{
			RuleSet:  []string{"geoip-cn"},
			Outbound: "🎯 全球直连",
		},
	}...)

	// 读取规则文件夹
	files, err := os.ReadDir("configs/sing-box/rules")
	if err != nil {
		return nil, fmt.Errorf("failed to read rules directory: %v", err)
	}

	// 添加自定义规则集
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".srs") {
			// 获取规则集标签（文件名去掉 .srs 后缀）
			tag := strings.TrimSuffix(file.Name(), ".srs")

			// 跳过内置规则集
			if tag == "geosite-cn" || tag == "geosite-category-ads" || tag == "geoip-cn" {
				continue
			}

			// 添加规则集配置
			config.Route.RuleSet = append(config.Route.RuleSet, RuleSetConfig{
				Tag:    tag,
				Type:   "local",
				Format: "binary",
				Path:   fmt.Sprintf("configs/sing-box/rules/%s", file.Name()),
			})

			// 为自定义规则创建选择器
			ruleSelector := OutboundConfig{
				Type:      "selector",
				Tag:       tag,
				Outbounds: append([]string{"🚀 节点选择"}, regionGroups...),
			}
			config.Outbounds = append(config.Outbounds, ruleSelector)

			// 添加路由规则
			config.Route.Rules = append(config.Route.Rules, RouteRule{
				RuleSet:  []string{tag},
				Outbound: tag,
			})
		}
	}

	// 将配置转换为 JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %v", err)
	}

	return data, nil
}

// regenerateConfig 重新生成配置文件
func (g *SingBoxGenerator) regenerateConfig() error {
	// 生成配置文件
	config, err := g.GenerateConfig()
	if err != nil {
		return fmt.Errorf("failed to generate config: %v", err)
	}

	// 创建配置目录
	if err := os.MkdirAll(filepath.Dir(g.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// 写入配置文件
	if err := os.WriteFile(g.configPath, config, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// ExperimentalConfig 实验性功能配置
type ExperimentalConfig struct {
	CacheFile CacheFile `json:"cache_file"`
	ClashAPI  ClashAPI  `json:"clash_api"`
}

// TLSConfig TLS 配置
type TLSConfig struct {
	Enabled    bool     `json:"enabled"`
	ServerName string   `json:"server_name,omitempty"`
	Insecure   bool     `json:"insecure,omitempty"`
	ALPN       []string `json:"alpn,omitempty"`
}

// TransportConfig 传输层配置
type TransportConfig struct {
	Type        string            `json:"type"`
	Path        string            `json:"path,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	ServiceName string            `json:"service_name,omitempty"`
}

// SniffConfig 嗅探配置
type SniffConfig struct {
	Enabled             bool     `json:"enabled"`
	OverrideDestination bool     `json:"override_destination"`
	Protocol            []string `json:"protocol"`
	ExcludeApplications []string `json:"exclude_applications"`
	ExcludeDomains      []string `json:"exclude_domains"`
	Ports               []uint16 `json:"ports"`
	PortRanges          []string `json:"port_ranges"`
	PortExclude         []uint16 `json:"port_exclude"`
	PortRangeExclude    []string `json:"port_range_exclude"`
}

// 添加规则集目录常量
const (
	ruleSetDir = "configs/sing-box/rules"
)

// ValidateConfig 验证sing-box配置
func (g *SingBoxGenerator) ValidateConfig(config []byte) error {
	var singBoxConfig SingBoxConfig
	if err := json.Unmarshal(config, &singBoxConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %v", err)
	}

	// 验证必要的配置项
	if singBoxConfig.Log == nil {
		return fmt.Errorf("log configuration is required")
	}

	if singBoxConfig.DNS == nil {
		return fmt.Errorf("DNS configuration is required")
	}

	if len(singBoxConfig.Inbounds) == 0 {
		return fmt.Errorf("at least one inbound is required")
	}

	if len(singBoxConfig.Outbounds) == 0 {
		return fmt.Errorf("at least one outbound is required")
	}

	if singBoxConfig.Route == nil {
		return fmt.Errorf("route configuration is required")
	}

	return nil
}

// UpdateSettings 更新设置
func (g *SingBoxGenerator) UpdateSettings(settings *models.Settings) error {
	g.settings = settings

	// 重新生成配置文件
	if err := g.regenerateConfig(); err != nil {
		return fmt.Errorf("failed to regenerate config: %v", err)
	}

	return nil
}

//
