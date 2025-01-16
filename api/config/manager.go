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

// Manager 配置管理器
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

// NewManager 创建配置管理器
func NewManager(logger *logrus.Logger, configPath string, storage storage.Storage) *Manager {
	return &Manager{
		logger:     logger,
		configPath: configPath,
		storage:    storage,
	}
}

// GetSettings 获取设置
func (m *Manager) GetSettings() *models.Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings
}

// UpdateSettings 更新设置
func (m *Manager) UpdateSettings(settings *models.Settings) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = settings

	// 生成新的 sing-box 配置
	generator := NewSingBoxGenerator(m.storage)
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

// IsRunning returns whether the sing-box service is running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}

// GetVersion returns the sing-box version
func (m *Manager) GetVersion() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.version
}

// GetUptime returns the sing-box uptime in seconds
func (m *Manager) GetUptime() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.isRunning {
		return 0
	}
	return time.Since(m.startTime).Milliseconds() / 1000
}

// StartService starts the sing-box service
func (m *Manager) StartService() error {
	return m.RestartService()
}

// StopService stops the sing-box service
func (m *Manager) StopService() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return nil
	}

	if m.cmd != nil && m.cmd.Process != nil {
		if err := m.cmd.Process.Kill(); err != nil {
			m.logger.WithError(err).Error("Failed to stop sing-box")
			return err
		}
	}

	m.isRunning = false
	return nil
}
