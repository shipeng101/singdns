package models

import "time"

// SystemInfo represents system information
type SystemInfo struct {
	Version         string    `json:"version"`
	StartTime       time.Time `json:"start_time"`
	IsRunning       bool      `json:"is_running"`
	ConfigPath      string    `json:"config_path"`
	Hostname        string    `json:"hostname"`
	Platform        string    `json:"platform"`
	Arch            string    `json:"arch"`
	Uptime          int64     `json:"uptime"`
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryTotal     uint64    `json:"memory_total"`
	MemoryUsed      uint64    `json:"memory_used"`
	NetworkUpload   uint64    `json:"network_upload"`
	NetworkDownload uint64    `json:"network_download"`
	Connections     int       `json:"connections"`
}

// SystemStatus represents system status
type SystemStatus struct {
	Services []ServicesStatus `json:"services"`
}

// ServicesStatus represents the status of system services
type ServicesStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	IsRunning bool   `json:"is_running"`
	Version   string `json:"version"`
	Uptime    int64  `json:"uptime"`
}
