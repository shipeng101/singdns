package system

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/sirupsen/logrus"

	"singdns/api/models"
)

// Manager handles system-related operations
type Manager struct {
	logger        *logrus.Logger
	startTime     time.Time
	proxyManager  ProxyManager
	configManager ConfigManager
	singboxCmd    *exec.Cmd
	routeManager  *RouteManager
}

// ProxyManager interface defines the methods required from proxy manager
type ProxyManager interface {
	GetVersion() string
	GetStartTime() time.Time
	IsRunning() bool
	GetConfigPath() string
	GetConnectionsCount() int
}

// ConfigManager interface defines the methods required from config manager
type ConfigManager interface {
	GetMosdnsVersion() string
	GetMosdnsUptime() int64
	IsMosdnsRunning() bool
}

// NewManager creates a new system manager
func NewManager(logger *logrus.Logger, proxyManager ProxyManager, configManager ConfigManager) *Manager {
	return &Manager{
		logger:        logger,
		startTime:     time.Now(),
		proxyManager:  proxyManager,
		configManager: configManager,
		routeManager:  NewRouteManager(),
	}
}

// GetSystemInfo returns current system information
func (m *Manager) GetSystemInfo() (*models.SystemInfo, error) {
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get hostname")
		hostname = "unknown"
	}

	// Get CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get CPU usage")
		cpuPercent = []float64{0}
	}

	// Get memory info
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get memory info")
		memInfo = &mem.VirtualMemoryStat{}
	}

	// Get network stats
	netStats, err := net.IOCounters(false)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get network stats")
		netStats = []net.IOCountersStat{{}}
	}

	// Get uptime
	hostInfo, err := host.Info()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get host info")
		hostInfo = &host.InfoStat{}
	}

	// Get connections count from proxy manager
	connections := m.proxyManager.GetConnectionsCount()

	return &models.SystemInfo{
		Version:         m.proxyManager.GetVersion(),
		StartTime:       m.proxyManager.GetStartTime(),
		IsRunning:       m.proxyManager.IsRunning(),
		ConfigPath:      m.proxyManager.GetConfigPath(),
		Hostname:        hostname,
		Platform:        fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH),
		Arch:            runtime.GOARCH,
		Uptime:          int64(hostInfo.Uptime),
		CPUUsage:        cpuPercent[0],
		MemoryTotal:     uint64(memInfo.Total),
		MemoryUsed:      uint64(memInfo.Used),
		NetworkUpload:   uint64(netStats[0].BytesSent),
		NetworkDownload: uint64(netStats[0].BytesRecv),
		Connections:     connections,
	}, nil
}

// GetSystemStatus returns current system and services status
func (m *Manager) GetSystemStatus() (*models.SystemStatus, error) {
	singboxStatus := "stopped"
	if m.proxyManager.IsRunning() {
		singboxStatus = "running"
	}

	services := []models.Service{
		{
			Name:    "sing-box",
			Status:  singboxStatus,
			Version: m.proxyManager.GetVersion(),
			Uptime:  int64(time.Since(m.proxyManager.GetStartTime()).Seconds()),
		},
	}

	// Get system info
	info, err := m.GetSystemInfo()
	if err != nil {
		return nil, err
	}

	return &models.SystemStatus{
		Services: services,
		System:   info,
	}, nil
}

// StartSingbox starts the singbox service
func (m *Manager) StartSingbox() error {
	if m.singboxCmd != nil && m.singboxCmd.Process != nil {
		return fmt.Errorf("singbox is already running")
	}

	m.singboxCmd = exec.Command("bin/sing-box", "run", "-c", "configs/sing-box/config.json")

	// Set required environment variables
	m.singboxCmd.Env = append(os.Environ(),
		"ENABLE_DEPRECATED_GEOIP=true",
		"ENABLE_DEPRECATED_GEOSITE=true",
	)

	if err := m.singboxCmd.Start(); err != nil {
		m.singboxCmd = nil
		return fmt.Errorf("failed to start singbox: %v", err)
	}

	// Monitor process state in background
	go func() {
		if err := m.singboxCmd.Wait(); err != nil {
			m.logger.WithError(err).Error("singbox process exited with error")
		}
		m.singboxCmd = nil
	}()

	return nil
}

// StopSingbox stops the singbox service
func (m *Manager) StopSingbox() error {
	if m.singboxCmd == nil || m.singboxCmd.Process == nil {
		// Try to find and kill any running sing-box process
		if err := exec.Command("pkill", "-f", "sing-box").Run(); err != nil {
			m.logger.WithError(err).Error("Failed to kill sing-box process")
			return fmt.Errorf("failed to kill sing-box process: %v", err)
		}
		return nil
	}

	// Try to kill the process we started
	if err := m.singboxCmd.Process.Kill(); err != nil {
		m.logger.WithError(err).Error("Failed to stop sing-box")
		// If killing the process fails, try pkill as a fallback
		if err := exec.Command("pkill", "-f", "sing-box").Run(); err != nil {
			m.logger.WithError(err).Error("Failed to kill sing-box process")
			return fmt.Errorf("failed to kill sing-box process: %v", err)
		}
	}

	m.singboxCmd = nil
	return nil
}

func (m *Manager) Start() error {
	// 设置路由规则
	if err := m.routeManager.Setup(); err != nil {
		return fmt.Errorf("failed to setup routing: %v", err)
	}

	return nil
}

func (m *Manager) Stop() error {
	// 清理路由规则
	if err := m.routeManager.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup routing: %v", err)
	}

	return nil
}
