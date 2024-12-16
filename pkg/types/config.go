package types

// Config 全局配置
type Config struct {
	Server  ServerConfig `json:"server"`   // 服务器配置
	Log     LogConfig    `json:"log"`      // 日志配置
	LogFile string       `json:"log_file"` // 日志文件路径
	Debug   bool         `json:"debug"`    // 是否开启调试模式
	DNS     DNSConfig    `json:"dns"`      // DNS配置
	Proxy   ProxyConfig  `json:"proxy"`    // 代理配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `json:"host"` // 监听地址
	Port int    `json:"port"` // 监听端口
	Mode string `json:"mode"` // 运行模式
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level"`       // 日志级别
	File       string `json:"file"`        // 日志文件
	MaxSize    int    `json:"max_size"`    // 单个文件最大大小(MB)
	MaxBackups int    `json:"max_backups"` // 最大备份数量
	MaxAge     int    `json:"max_age"`     // 最大保留天数
	Compress   bool   `json:"compress"`    // 是否压缩
}

// DNSConfig DNS配置
type DNSConfig struct {
	Listen   string    `json:"listen"`   // 监听地址
	Port     int       `json:"port"`     // 监听端口
	Cache    bool      `json:"cache"`    // 是否启用缓存
	Upstream []string  `json:"upstream"` // 上游DNS服务器
	ChinaDNS []string  `json:"chinadns"` // 中国DNS服务器
	RuleSets []RuleSet `json:"rulesets"` // 规则集
}

// RuleSet 规则集
type RuleSet struct {
	ID             string   `json:"id"`              // 规则集ID
	Name           string   `json:"name"`            // 规则集名称
	Description    string   `json:"description"`     // 规则集描述
	Type           string   `json:"type"`            // 规则集类型
	Rules          []Rule   `json:"rules"`           // 规则列表
	Enabled        bool     `json:"enabled"`         // 是否启用
	Priority       int      `json:"priority"`        // 优先级
	AutoUpdate     bool     `json:"auto_update"`     // 是否自动更新
	UpdateInterval int      `json:"update_interval"` // 更新间隔(小时)
	Tags           []string `json:"tags"`            // 标签
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	Mode    string        `json:"mode"`    // 代理模式：rule/global/direct
	Inbound InboundConfig `json:"inbound"` // 入站配置
	Rules   []Rule        `json:"rules"`   // 代理规则
	Nodes   []*ProxyNode  `json:"nodes"`   // 代理节点
}

// InboundConfig 入站配置
type InboundConfig struct {
	Listen            string `json:"listen"`             // 监听地址
	Port              int    `json:"port"`               // 混合端口
	SocksPort         int    `json:"socks_port"`         // Socks端口
	HTTPPort          int    `json:"http_port"`          // HTTP端口
	Username          string `json:"username"`           // 认证用户名
	Password          string `json:"password"`           // 认证密码
	EnableTUN         bool   `json:"enable_tun"`         // 启用TUN
	EnableTransparent bool   `json:"enable_transparent"` // 启用透明代理
	SetSystemProxy    bool   `json:"set_system_proxy"`   // 设置系统代理
	TransparentPort   int    `json:"transparent_port"`   // 透明代理端口

	// TUN相关配置
	TUNInterface              string `json:"tun_interface"`                // TUN接口名称
	TUNAddress                string `json:"tun_address"`                  // TUN地址
	TUNMTU                    int    `json:"tun_mtu"`                      // TUN MTU
	TUNAutoRoute              bool   `json:"tun_auto_route"`               // TUN自动路由
	TUNStrictRoute            bool   `json:"tun_strict_route"`             // TUN严格路由
	TUNStack                  string `json:"tun_stack"`                    // TUN协议栈
	TUNEndpointIndependentNat bool   `json:"tun_endpoint_independent_nat"` // TUN端点独立NAT
}

// Rule 代理规则
type Rule struct {
	Type    string `json:"type"`    // 规则类型(domain/ip/port/domain_suffix/domain_keyword)
	Value   string `json:"value"`   // 规则值
	Target  string `json:"target"`  // 目标(proxy/direct/reject)
	Enabled bool   `json:"enabled"` // 是否启用
}

// DashboardConfig 仪表盘配置
type DashboardConfig struct {
	Current     string `json:"current"`      // 当前面板
	ConfigFile  string `json:"config_file"`  // 配置文件路径
	ResourceDir string `json:"resource_dir"` // 资源目录
	APIEndpoint string `json:"api_endpoint"` // API端点
}

// NewDefaultConfig 创建默认配置
func NewDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
			Mode: "release",
		},
		Log: LogConfig{
			Level:      "info",
			File:       "logs/singdns.log",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     7,
			Compress:   true,
		},
		DNS: DNSConfig{
			Listen: "127.0.0.1",
			Port:   53,
			Cache:  true,
			Upstream: []string{
				"tcp://8.8.8.8:53",
				"tcp://8.8.4.4:53",
			},
			ChinaDNS: []string{
				"tcp://223.5.5.5:53",
				"tcp://119.29.29.29:53",
			},
		},
		Proxy: ProxyConfig{
			Mode: "rule",
			Inbound: InboundConfig{
				Listen:         "127.0.0.1",
				Port:           1080,
				SocksPort:      1080,
				HTTPPort:       8118,
				SetSystemProxy: false,
			},
		},
		Debug: false,
	}
}
