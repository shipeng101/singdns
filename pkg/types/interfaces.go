package types

// Services 服务集合
type Services struct {
	AuthService         AuthService
	DNSService          DNSService
	ProxyService        ProxyService
	SubscriptionService SubscriptionService
	CoreService         CoreService
	DashboardService    DashboardService
	ConfigService       ConfigService
}

// AuthService 认证服务接口
type AuthService interface {
	// Login 用户登录
	Login(username, password string) (bool, error)
	// Register 用户注册
	Register(username, password string) error
	// Validate 验证令牌
	Validate(token string) bool
}

// DNSService DNS服务接口
type DNSService interface {
	// 启动服务
	Start() error
	// 停止服务
	Stop() error
	// 获取配置
	GetConfig() DNSConfig
	// 更新配置
	UpdateConfig(config DNSConfig) error
	// 获取规则
	GetRules() []Rule
	// 更新规则
	UpdateRules(rules []Rule) error
	// 获取状态
	GetStatus() ServiceStatus
}

// ProxyService 代理服务接口
type ProxyService interface {
	// 启动服务
	Start() error
	// 停止服务
	Stop() error
	// 重启服务
	Restart() error
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
	// 选择节点
	SelectNode(id string) error
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
