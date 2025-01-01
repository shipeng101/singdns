package api

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

// SystemManager handles system-related operations
type SystemManager struct {
	logger        *logrus.Logger
	startTime     time.Time
	proxyManager  ProxyManager
	configManager ConfigManager
}

// NewSystemManager creates a new system manager
func NewSystemManager(logger *logrus.Logger, proxyManager ProxyManager, configManager ConfigManager) *SystemManager {
	return &SystemManager{
		logger:        logger,
		startTime:     time.Now(),
		proxyManager:  proxyManager,
		configManager: configManager,
	}
}

// GetSystemInfo returns current system information
func (s *SystemManager) GetSystemInfo() (*models.SystemInfo, error) {
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		s.logger.WithError(err).Error("Failed to get hostname")
		hostname = "unknown"
	}

	// Get CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get CPU usage")
		cpuPercent = []float64{0}
	}

	// Get memory info
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		s.logger.WithError(err).Error("Failed to get memory info")
		memInfo = &mem.VirtualMemoryStat{}
	}

	// Get network stats
	netStats, err := net.IOCounters(false)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get network stats")
		netStats = []net.IOCountersStat{{}}
	}

	// Get uptime
	hostInfo, err := host.Info()
	if err != nil {
		s.logger.WithError(err).Error("Failed to get host info")
		hostInfo = &host.InfoStat{}
	}

	// Get connections count from proxy manager
	connections := s.proxyManager.GetConnectionsCount()

	return &models.SystemInfo{
		Version:         s.proxyManager.GetVersion(),
		StartTime:       s.proxyManager.GetStartTime(),
		IsRunning:       s.proxyManager.IsRunning(),
		ConfigPath:      s.proxyManager.GetConfigPath(),
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
func (s *SystemManager) GetSystemStatus() (*models.SystemStatus, error) {
	singboxStatus := "stopped"
	if s.proxyManager.IsRunning() {
		singboxStatus = "running"
	}

	mosdnsStatus := "stopped"
	if s.configManager.IsMosdnsRunning() {
		mosdnsStatus = "running"
	}

	return &models.SystemStatus{
		Services: []models.ServicesStatus{
			{
				Name:      "singbox",
				Status:    singboxStatus,
				IsRunning: s.proxyManager.IsRunning(),
				Version:   s.proxyManager.GetVersion(),
				Uptime:    int64(time.Since(s.proxyManager.GetStartTime()).Seconds()),
			},
			{
				Name:      "mosdns",
				Status:    mosdnsStatus,
				IsRunning: s.configManager.IsMosdnsRunning(),
				Version:   s.configManager.GetMosdnsVersion(),
				Uptime:    s.configManager.GetMosdnsUptime(),
			},
		},
	}, nil
}

// StartService starts a system service
func (s *SystemManager) StartService(name string) error {
	s.logger.WithField("service", name).Info("Starting service")

	switch name {
	case "singbox":
		return s.proxyManager.Start()
	case "mosdns":
		return s.configManager.StartMosdns()
	default:
		return fmt.Errorf("unknown service: %s", name)
	}
}

// StopService stops a system service
func (s *SystemManager) StopService(name string) error {
	s.logger.WithField("service", name).Info("Stopping service")

	switch name {
	case "singbox":
		return s.proxyManager.Stop()
	case "mosdns":
		return s.configManager.StopMosdns()
	default:
		return fmt.Errorf("unknown service: %s", name)
	}
}

// RestartService restarts a system service
func (s *SystemManager) RestartService(name string) error {
	s.logger.WithField("service", name).Info("Restarting service")

	switch name {
	case "singbox":
		if err := s.proxyManager.Stop(); err != nil {
			return err
		}
		time.Sleep(time.Second) // Wait for service to stop
		return s.proxyManager.Start()
	case "mosdns":
		if err := s.configManager.StopMosdns(); err != nil {
			return err
		}
		time.Sleep(time.Second) // Wait for service to stop
		return s.configManager.StartMosdns()
	default:
		return fmt.Errorf("unknown service: %s", name)
	}
}
