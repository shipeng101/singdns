package types

import "time"

// ServiceStatus 服务状态
type ServiceStatus struct {
	Running    bool      `json:"running"`     // 运行状态
	Uptime     string    `json:"uptime"`      // 运行时间
	QueryCount int64     `json:"query_count"` // 查询次数
	StartTime  time.Time `json:"start_time"`  // 启动时间
}

// NodeStatus 节点状态
type NodeStatus struct {
	ID        string `json:"id"`         // 节点ID
	Connected bool   `json:"connected"`  // 是否连接
	UpSpeed   int64  `json:"up_speed"`   // 上传速度(B/s)
	DownSpeed int64  `json:"down_speed"` // 下载速度(B/s)
	UpTotal   int64  `json:"up_total"`   // 上传总量(B)
	DownTotal int64  `json:"down_total"` // 下载总量(B)
}

// ProxyStats 代理统计
type ProxyStats struct {
	Mode        string       `json:"mode"`        // 代理模式
	Running     bool         `json:"running"`     // 运行状态
	Uptime      int64        `json:"uptime"`      // 运行时间
	UpSpeed     int64        `json:"up_speed"`    // 上传速度
	DownSpeed   int64        `json:"down_speed"`  // 下载速度
	UpTotal     int64        `json:"up_total"`    // 总上传
	DownTotal   int64        `json:"down_total"`  // 总下载
	Connections int          `json:"connections"` // 连接数
	Nodes       []NodeStatus `json:"nodes"`       // 节点状态
}
