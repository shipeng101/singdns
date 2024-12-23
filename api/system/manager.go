package system

import (
	"fmt"
	"os"
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

	mosdnsStatus := "stopped"
	if m.configManager.IsMosdnsRunning() {
		mosdnsStatus = "running"
	}

	return &models.SystemStatus{
		Services: []models.ServicesStatus{
			{
				Name:      "singbox",
				Status:    singboxStatus,
				IsRunning: m.proxyManager.IsRunning(),
				Version:   m.proxyManager.GetVersion(),
				Uptime:    int64(time.Since(m.proxyManager.GetStartTime()).Seconds()),
			},
			{
				Name:      "mosdns",
				Status:    mosdnsStatus,
				IsRunning: m.configManager.IsMosdnsRunning(),
				Version:   m.configManager.GetMosdnsVersion(),
				Uptime:    m.configManager.GetMosdnsUptime(),
			},
		},
	}, nil
}
