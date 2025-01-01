package config

import (
	"os"
	"os/exec"
	"sync"
	"time"

	"singdns/api/models"
	"singdns/api/storage"

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
	storage    storage.Storage
}

// NewManager creates a new config manager
func NewManager(logger *logrus.Logger, storage storage.Storage, configPath string) *Manager {
	return &Manager{
		logger:     logger,
		storage:    storage,
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

	m.cmd = exec.Command("bin/mosdns", "start", "-c", m.configPath)
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
		m.startTime = time.Time{}
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
	m.startTime = time.Time{}
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
		out, err := exec.Command("bin/mosdns", "version").Output()
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

	// 生成新的 sing-box 配置
	generator := NewSingBoxGenerator(settings, m.storage, m.configPath)
	config, err := generator.GenerateConfig()
	if err != nil {
		m.logger.WithError(err).Error("Failed to generate sing-box config")
		return err
	}

	// 保存配置文件
	if err := os.WriteFile(m.configPath, config, 0644); err != nil {
		m.logger.WithError(err).Error("Failed to write sing-box config")
		return err
	}

	// 如果 sing-box 正在运行，重启服务以应用新配置
	if m.isRunning {
		if err := m.RestartService(); err != nil {
			m.logger.WithError(err).Error("Failed to restart sing-box")
			return err
		}
	}

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

// RestartService restarts the sing-box service
func (m *Manager) RestartService() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果服务正在运行，先停止
	if m.isRunning {
		if m.cmd != nil && m.cmd.Process != nil {
			if err := m.cmd.Process.Kill(); err != nil {
				m.logger.WithError(err).Error("Failed to stop sing-box")
				return err
			}
		}
	}

	// 启动服务
	m.cmd = exec.Command("bin/sing-box", "run", "-c", m.configPath)
	if err := m.cmd.Start(); err != nil {
		m.logger.WithError(err).Error("Failed to start sing-box")
		return err
	}

	m.isRunning = true
	m.startTime = time.Now()
	m.logger.Info("sing-box restarted")

	// 监控进程
	go func() {
		if err := m.cmd.Wait(); err != nil {
			m.logger.WithError(err).Error("sing-box process exited with error")
		}
		m.mu.Lock()
		m.isRunning = false
		m.mu.Unlock()
	}()

	return nil
}

// UpdateConfig updates the configuration
func (m *Manager) UpdateConfig() error {
	// 获取设置
	settings, err := m.storage.GetSettings()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get settings")
		return err
	}

	// 生成新的 sing-box 配置
	generator := NewSingBoxGenerator(settings, m.storage, m.configPath)
	config, err := generator.GenerateConfig()
	if err != nil {
		m.logger.WithError(err).Error("Failed to generate sing-box config")
		return err
	}

	// 保存配置文件
	if err := os.WriteFile(m.configPath, config, 0644); err != nil {
		m.logger.WithError(err).Error("Failed to write sing-box config")
		return err
	}

	// 如果 sing-box 正在运行，重启服务以应用新配置
	if m.isRunning {
		if err := m.RestartService(); err != nil {
			m.logger.WithError(err).Error("Failed to restart sing-box")
			return err
		}
	}

	return nil
}
