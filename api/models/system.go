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

// Service represents a system service status
type Service struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  int64  `json:"uptime"`
}

// SystemStatus represents the system status
type SystemStatus struct {
	Services []Service   `json:"services"`
	System   *SystemInfo `json:"system"`
}
