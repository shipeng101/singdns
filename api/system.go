package api

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemManager struct {
	startTime time.Time
	mu        sync.RWMutex
	services  map[string]*ServiceInfo
}

type ServiceInfo struct {
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
	Status  string `json:"status"`
}

type SystemInfo struct {
	CPU struct {
		Usage float64 `json:"usage"`
	} `json:"cpu"`
	Memory struct {
		Total float64 `json:"total"`
		Used  float64 `json:"used"`
	} `json:"memory"`
	Network struct {
		Speed struct {
			Up   float64 `json:"up"`
			Down float64 `json:"down"`
		} `json:"speed"`
	} `json:"network"`
	Singbox *ServiceInfo `json:"singbox"`
	Mosdns  *ServiceInfo `json:"mosdns"`
}

func NewSystemManager() *SystemManager {
	return &SystemManager{
		startTime: time.Now(),
		services: map[string]*ServiceInfo{
			"singbox": {Status: "stopped"},
			"mosdns":  {Status: "stopped"},
		},
	}
}

func (m *SystemManager) GetStatus() (*SystemInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := &SystemInfo{
		Singbox: m.services["singbox"],
		Mosdns:  m.services["mosdns"],
	}

	// CPU 使用率
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %v", err)
	}
	if len(cpuPercent) > 0 {
		status.CPU.Usage = cpuPercent[0]
	}

	// 内存使用情况
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %v", err)
	}
	status.Memory.Total = float64(memInfo.Total) / 1024 / 1024 / 1024 // 转换为 GB
	status.Memory.Used = float64(memInfo.Used) / 1024 / 1024 / 1024   // 转换为 GB

	return status, nil
}

func (m *SystemManager) GetServices() map[string]*ServiceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.services
}

func (m *SystemManager) StartService(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[name]
	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	if service.Status == "running" {
		return fmt.Errorf("service %s is already running", name)
	}

	var cmd *exec.Cmd
	switch name {
	case "singbox":
		cmd = exec.Command("./bin/sing-box", "run")
	case "mosdns":
		cmd = exec.Command("./bin/mosdns", "start")
	default:
		return fmt.Errorf("unknown service: %s", name)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %v", name, err)
	}

	service.Status = "running"
	service.Version = m.getServiceVersion(name)
	return nil
}

func (m *SystemManager) StopService(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[name]
	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	if service.Status == "stopped" {
		return fmt.Errorf("service %s is already stopped", name)
	}

	var cmd *exec.Cmd
	switch name {
	case "singbox":
		cmd = exec.Command("pkill", "sing-box")
	case "mosdns":
		cmd = exec.Command("pkill", "mosdns")
	default:
		return fmt.Errorf("unknown service: %s", name)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop %s: %v", name, err)
	}

	service.Status = "stopped"
	return nil
}

func (m *SystemManager) RestartService(name string) error {
	if err := m.StopService(name); err != nil {
		return fmt.Errorf("failed to stop service: %v", err)
	}
	time.Sleep(time.Second) // 等待服务完全停止
	if err := m.StartService(name); err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}
	return nil
}

func (m *SystemManager) getServiceVersion(name string) string {
	var cmd *exec.Cmd
	switch name {
	case "singbox":
		cmd = exec.Command("./bin/sing-box", "version")
	case "mosdns":
		cmd = exec.Command("./bin/mosdns", "version")
	default:
		return "unknown"
	}

	output, err := cmd.Output()
	if err != nil {
		LogWarning("Failed to get %s version: %v", name, err)
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func (m *SystemManager) formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d天 %d小时 %d分钟", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d小时 %d分钟", hours, minutes)
	}
	return fmt.Sprintf("%d分钟", minutes)
}
