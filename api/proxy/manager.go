package proxy

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"singdns/api/models"

	"github.com/sirupsen/logrus"
)

// Manager implements the ProxyManager interface
type Manager struct {
	cmd        *exec.Cmd
	configPath string
	logger     *logrus.Logger
	workDir    string
	startTime  time.Time
	version    string
}

// NewManager creates a new proxy manager
func NewManager(logger *logrus.Logger, configPath string, workDir string) *Manager {
	return &Manager{
		logger:     logger,
		configPath: configPath,
		workDir:    workDir,
	}
}

// Start starts the sing-box service
func (m *Manager) Start() error {
	m.logger.WithFields(logrus.Fields{
		"workDir":    m.workDir,
		"configPath": m.configPath,
	}).Info("Starting sing-box service")

	if m.IsRunning() {
		m.logger.Warn("Service is already running")
		return fmt.Errorf("service is already running")
	}

	// Kill any existing sing-box processes
	if err := exec.Command("pkill", "-f", "sing-box").Run(); err != nil {
		m.logger.WithError(err).Debug("No existing sing-box processes found")
	}

	// Check if sing-box binary exists
	binPath := fmt.Sprintf("%s/bin/sing-box", m.workDir)
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		m.logger.WithField("path", binPath).Error("sing-box binary not found")
		return fmt.Errorf("sing-box binary not found at %s", binPath)
	}

	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		m.logger.WithField("path", m.configPath).Error("Config file not found")
		return fmt.Errorf("config file not found at %s", m.configPath)
	}

	// Check file permissions
	binInfo, err := os.Stat(binPath)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get binary file info")
		return fmt.Errorf("failed to get binary file info: %v", err)
	}
	m.logger.WithField("permissions", binInfo.Mode().String()).Debug("Binary file permissions")

	configInfo, err := os.Stat(m.configPath)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get config file info")
		return fmt.Errorf("failed to get config file info: %v", err)
	}
	m.logger.WithField("permissions", configInfo.Mode().String()).Debug("Config file permissions")

	// Create command
	m.cmd = exec.Command(binPath, "run", "-c", m.configPath)
	m.cmd.Dir = m.workDir

	// Capture command output
	var stdout, stderr bytes.Buffer
	m.cmd.Stdout = &stdout
	m.cmd.Stderr = &stderr

	// Start the process
	m.startTime = time.Now()
	if err := m.cmd.Start(); err != nil {
		m.logger.WithError(err).Error("Failed to start service")
		return fmt.Errorf("failed to start service: %v", err)
	}

	// Wait a bit to ensure process started successfully
	time.Sleep(time.Second)
	if !m.IsRunning() {
		output := stdout.String()
		errOutput := stderr.String()
		m.logger.WithFields(logrus.Fields{
			"stdout": output,
			"stderr": errOutput,
		}).Error("Service failed to start")
		return fmt.Errorf("service failed to start. Stdout: %s, Stderr: %s", output, errOutput)
	}

	m.logger.Info("sing-box service started successfully")
	return nil
}

// Stop stops the sing-box service
func (m *Manager) Stop() error {
	if !m.IsRunning() {
		return nil
	}

	if err := m.cmd.Process.Kill(); err != nil {
		m.logger.WithError(err).Error("Failed to stop service")
		// If killing the process fails, try pkill as a fallback
		if err := exec.Command("pkill", "-f", "sing-box").Run(); err != nil {
			m.logger.WithError(err).Error("Failed to kill processes")
			return fmt.Errorf("failed to kill process: %v", err)
		}
	}
	m.cmd = nil
	return nil
}

// IsRunning returns whether the service is running
func (m *Manager) IsRunning() bool {
	if m.cmd == nil || m.cmd.Process == nil {
		return false
	}

	// 使用 pgrep 检查进程
	cmd := exec.Command("pgrep", "-f", "sing-box")
	output, err := cmd.Output()
	if err != nil {
		m.cmd = nil
		return false
	}

	// 检查进程是否真的在运行
	if len(output) == 0 {
		m.cmd = nil
		return false
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		m.cmd = nil
		return false
	}

	return true
}

// GetVersion returns the version of sing-box
func (m *Manager) GetVersion() string {
	if m.version != "" {
		return m.version
	}

	cmd := exec.Command(fmt.Sprintf("%s/bin/sing-box", m.workDir), "version")
	cmd.Dir = m.workDir
	out, err := cmd.Output()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get version")
		return "unknown"
	}

	m.version = string(out)
	return m.version
}

// GetUptime returns the uptime in seconds
func (m *Manager) GetUptime() int64 {
	if !m.IsRunning() {
		return 0
	}
	return int64(time.Since(m.startTime).Seconds())
}

// StartService starts the specified service
func (m *Manager) StartService(service string) error {
	switch service {
	case "sing-box":
		return m.Start()
	default:
		return fmt.Errorf("unknown service: %s", service)
	}
}

// StopService stops the specified service
func (m *Manager) StopService(service string) error {
	switch service {
	case "sing-box":
		return m.Stop()
	default:
		return fmt.Errorf("unknown service: %s", service)
	}
}

// RestartService restarts the specified service
func (m *Manager) RestartService(service string) error {
	switch service {
	case "sing-box":
		if err := m.Stop(); err != nil {
			m.logger.WithError(err).Error("Failed to stop service during restart")
		}
		return m.Start()
	default:
		return fmt.Errorf("unknown service: %s", service)
	}
}

// GetNodeHealth returns the health status of a node
func (m *Manager) GetNodeHealth(id string) *models.HealthData {
	return &models.HealthData{
		Status:    "unknown",
		Latency:   0,
		LastCheck: time.Now(),
	}
}

// UpdateNodeHealth updates the health status of a node
func (m *Manager) UpdateNodeHealth(nodeID string, status string, latency int64, checkedAt time.Time, err string) {
	// No-op since we removed health tracking
}

// GetStartTime returns the start time of the service
func (m *Manager) GetStartTime() time.Time {
	return m.startTime
}
