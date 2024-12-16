package types

import "time"

// ProxyNode 代理节点
type ProxyNode struct {
	ID            string            `json:"id"`             // 节点ID
	Type          string            `json:"type"`           // 节点类型(ss/vmess/trojan/vless)
	Name          string            `json:"name"`           // 节点名称
	Server        string            `json:"server"`         // 服务器地址
	Port          int               `json:"port"`           // 服务器端口
	UUID          string            `json:"uuid"`           // UUID(vmess/vless)
	Password      string            `json:"password"`       // 密码(ss/trojan)
	Cipher        string            `json:"cipher"`         // 加密方式(ss)
	Network       string            `json:"network"`        // 传输协议(tcp/ws/grpc)
	TLS           bool              `json:"tls"`            // 是否启用TLS
	ALPN          []string          `json:"alpn"`           // ALPN
	SNI           string            `json:"sni"`            // SNI
	AlterId       int               `json:"alter_id"`       // AlterId(vmess)
	Security      string            `json:"security"`       // 加密方式(vmess)
	Flow          string            `json:"flow"`           // 流控(vless)
	Plugin        string            `json:"plugin"`         // 插件(ss)
	PluginOpts    map[string]string `json:"plugin_opts"`    // 插件选项
	Group         string            `json:"group"`          // 节点分组
	AllowInsecure bool              `json:"allow_insecure"` // 是否允许不安全连接
	Latency       int64             `json:"latency"`        // 延迟(毫秒)
	Path          string            `json:"path"`           // 路径(ws/http)
	Host          string            `json:"host"`           // 主机名(ws/http)
	ServiceName   string            `json:"service_name"`   // 服务名称(grpc)
	LastTest      time.Time         `json:"last_test"`      // 最后测试时间
	Method        string            `json:"method"`         // 加密方式(ss/ssr)
	UpMbps        int               `json:"up_mbps"`        // 上行带宽限制
	DownMbps      int               `json:"down_mbps"`      // 下行带宽限制
	Protocol      string            `json:"protocol"`       // 协议
	Obfs          string            `json:"obfs"`           // 混淆方式
	Congestion    string            `json:"congestion"`     // 拥塞控制
	UDPRelayMode  string            `json:"udp_relay_mode"` // UDP转发模式
}

// IPTablesRule iptables规则
type IPTablesRule struct {
	Type      string `json:"type"`      // 规则类型
	Table     string `json:"table"`     // 表名
	Chain     string `json:"chain"`     // 链名
	Interface string `json:"interface"` // 网卡
	Target    string `json:"target"`    // 目标
	Proto     string `json:"proto"`     // 协议
	Source    string `json:"source"`    // 源地址
	Dest      string `json:"dest"`      // 目标地址
	Sport     string `json:"sport"`     // 源端口
	Dport     string `json:"dport"`     // 目标端口
	ToPort    string `json:"to_port"`   // 转发端口
	Mark      string `json:"mark"`      // 标记
	Comment   string `json:"comment"`   // 注释
}

// Subscription 订阅信息
type Subscription struct {
	ID          string       `json:"id"`           // 订阅ID
	Name        string       `json:"name"`         // 订阅名称
	URL         string       `json:"url"`          // 订阅地址
	Format      string       `json:"format"`       // 订阅格式(base64/sip008/clash)
	Group       string       `json:"group"`        // 节点分组
	UserAgent   string       `json:"user_agent"`   // User-Agent
	Nodes       []*ProxyNode `json:"nodes"`        // 节点列表
	CreatedAt   time.Time    `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time    `json:"updated_at"`   // 更新时间
	NextUpdate  time.Time    `json:"next_update"`  // 下次更新时间
	AutoUpdate  bool         `json:"auto_update"`  // 是否自动更新
	UpdateHours int          `json:"update_hours"` // 更新间隔(小时)
}
