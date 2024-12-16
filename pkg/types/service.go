package types

import "time"

// ServiceStatus 服务状态
type ServiceStatus struct {
	Running    bool      `json:"running"`     // 运行状态
	Uptime     string    `json:"uptime"`      // 运行时间
	QueryCount int64     `json:"query_count"` // 查询次数
	StartTime  time.Time `json:"start_time"`  // 启动时间
}

// ProxyService 代理服务接口
type ProxyService interface {
	// 启动服务
	Start() error
	// 停止服务
	Stop() error
	// 重启服务
	Restart() error

	// 节点管理
	AddNode(node *ProxyNode) error
	UpdateNode(node *ProxyNode) error
	DeleteNode(id string) error
	GetNode(id string) (*ProxyNode, error)
	ListNodes() ([]*ProxyNode, error)
	TestNode(id string) (int64, error)
	TestAllNodes() error
	SwitchNode(id string) error

	// 代理控制
	GetStats() (*ProxyStats, error)
	SetMode(mode string) error

	// 规则管理
	UpdateRules(rules []Rule) error
	GetRules() ([]Rule, error)

	// iptables管理
	UpdateIPTables(rules []IPTablesRule) error
	GetIPTables() ([]IPTablesRule, error)
}

// DNSService DNS服务接口
type DNSService interface {
	// 启动服务
	Start() error
	// 停止服务
	Stop() error

	// 配置管理
	GetConfig() DNSConfig
	UpdateConfig(config DNSConfig) error

	// 规则管理
	GetRules() []Rule
	UpdateRules(rules []Rule) error

	// 状态查询
	GetStatus() ServiceStatus
}

// SubscriptionService 订阅服务接口
type SubscriptionService interface {
	// 启动服务
	Start() error
	// 停止服务
	Stop() error
	// 获取节点列表
	GetNodes() ([]*ProxyNode, error)
	// 更新订阅
	UpdateSubscription(url string) error
}

// CoreService 核心服务接口
type CoreService interface {
	// 初始化核心文件
	Init() error
	// 检查更新
	CheckUpdate() error
	// 获取核心文件路径
	GetSingBoxPath() string
	GetMosDNSPath() string
}

// DashboardService 仪表盘服务接口
type DashboardService interface {
	// 启动服务
	Start() error
	// 停止服务
	Stop() error
	// 获取配置
	GetConfig() *DashboardConfig
	// 更新配置
	UpdateConfig(config *DashboardConfig) error
	// 切换面板
	SwitchDashboard(dashboard string) error
	// 获取当前面板
	GetCurrentDashboard() string
	// 获取面板URL
	GetDashboardURL() string
}

// DashboardConfig 仪表盘配置
type DashboardConfig struct {
	Current     string `json:"current"`      // 当前面板
	ConfigFile  string `json:"config_file"`  // 配置文件路径
	ResourceDir string `json:"resource_dir"` // 资源目录
	APIEndpoint string `json:"api_endpoint"` // API端点
}

// ConfigService 配置服务接口
type ConfigService interface {
	// 加载配置
	Load() error
	// 保存配置
	Save() error
	// 获取配置
	Get() *Config
	// 更新配置
	Update(config *Config) error
}
