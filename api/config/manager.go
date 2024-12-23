package config

import (
	"os/exec"
	"sync"
	"time"

	"singdns/api/models"

	"github.com/sirupsen/logrus"
)

// Manager implements the ConfigManager interface
type Manager struct {
	logger     *logrus.Logger
	mu         sync.RWMutex
	cmd        *exec.Cmd
	startTime  time.Time
	isRunning  bool
	version    string
	configPath string
	settings   *models.Settings
}

// NewManager creates a new config manager
func NewManager(logger *logrus.Logger, configPath string) *Manager {
	return &Manager{
		logger:     logger,
		configPath: configPath,
	}
}

// StartMosdns starts the mosdns service
func (m *Manager) StartMosdns() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return nil
	}

	m.cmd = exec.Command("mosdns", "start", "-c", m.configPath)
	if err := m.cmd.Start(); err != nil {
		m.logger.WithError(err).Error("Failed to start mosdns")
		return err
	}

	m.isRunning = true
	m.startTime = time.Now()
	m.logger.Info("mosdns started")

	// Start a goroutine to monitor the process
	go func() {
		if err := m.cmd.Wait(); err != nil {
			m.logger.WithError(err).Error("mosdns process exited with error")
		}
		m.mu.Lock()
		m.isRunning = false
		m.mu.Unlock()
	}()

	return nil
}

// StopMosdns stops the mosdns service
func (m *Manager) StopMosdns() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return nil
	}

	if m.cmd != nil && m.cmd.Process != nil {
		if err := m.cmd.Process.Kill(); err != nil {
			m.logger.WithError(err).Error("Failed to stop mosdns")
			return err
		}
	}

	m.isRunning = false
	m.logger.Info("mosdns stopped")
	return nil
}

// IsMosdnsRunning returns whether the mosdns service is running
func (m *Manager) IsMosdnsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}

// GetMosdnsVersion returns the version of the mosdns service
func (m *Manager) GetMosdnsVersion() string {
	if m.version == "" {
		out, err := exec.Command("mosdns", "version").Output()
		if err != nil {
			m.logger.WithError(err).Error("Failed to get mosdns version")
			return "unknown"
		}
		m.version = string(out)
	}
	return m.version
}

// GetMosdnsUptime returns the uptime of the mosdns service in seconds
func (m *Manager) GetMosdnsUptime() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.isRunning {
		return 0
	}
	return int64(time.Since(m.startTime).Seconds())
}

// GetSettings returns the current settings
func (m *Manager) GetSettings() *models.Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.settings == nil {
		m.settings = &models.Settings{
			Theme:            "dark",
			Language:         "zh-CN",
			ProxyPort:        7890,
			APIPort:          8080,
			DNSPort:          53,
			LogLevel:         "info",
			EnableAutoUpdate: true,
		}
	}
	return m.settings
}

// UpdateSettings updates the settings
func (m *Manager) UpdateSettings(settings *models.Settings) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = settings
	// TODO: Save settings to file
	return nil
}

// BackupConfig creates a backup of the current configuration
func (m *Manager) BackupConfig() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// TODO: Implement config backup
	return "", nil
}

// RestoreConfig restores configuration from a backup
func (m *Manager) RestoreConfig(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// TODO: Implement config restore
	return nil
}

// DeleteBackup deletes a configuration backup
func (m *Manager) DeleteBackup(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// TODO: Implement backup deletion
	return nil
}
