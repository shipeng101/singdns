package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigManager manages configuration and supports hot reload
type ConfigManager struct {
	configDir string
	settings  *Settings
	watcher   *fsnotify.Watcher
	handlers  map[string]func([]byte) error
	mu        sync.RWMutex
}

// NewConfigManager creates a new config manager
func NewConfigManager(configDir string) (*ConfigManager, error) {
	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher failed: %v", err)
	}

	cm := &ConfigManager{
		configDir: configDir,
		settings:  &Settings{},
		watcher:   watcher,
		handlers:  make(map[string]func([]byte) error),
	}

	// Register default handlers
	cm.handlers["settings.json"] = cm.handleSettingsChange
	cm.handlers["sing-box/config.json"] = cm.handleSingboxChange
	cm.handlers["mosdns/config.yaml"] = cm.handleMosdnsChange

	// Start watching
	go cm.watch()

	// Watch config directory
	if err := watcher.Add(configDir); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("watch config dir failed: %v", err)
	}

	// Watch subdirectories
	subdirs := []string{"sing-box", "mosdns"}
	for _, subdir := range subdirs {
		path := filepath.Join(configDir, subdir)
		if err := os.MkdirAll(path, 0755); err != nil {
			continue
		}
		if err := watcher.Add(path); err != nil {
			LogWarning("Failed to watch directory", "path", path, "error", err)
		}
	}

	return cm, nil
}

// Close closes the config manager
func (cm *ConfigManager) Close() error {
	return cm.watcher.Close()
}

// watch monitors file changes
func (cm *ConfigManager) watch() {
	for {
		select {
		case event, ok := <-cm.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Get relative path
				relPath, err := filepath.Rel(cm.configDir, event.Name)
				if err != nil {
					LogError(err, "Failed to get relative path", "path", event.Name)
					continue
				}

				// Find handler
				handler := cm.handlers[relPath]
				if handler == nil {
					continue
				}

				// Read file
				data, err := os.ReadFile(event.Name)
				if err != nil {
					LogError(err, "Failed to read config file", "path", event.Name)
					continue
				}

				// Handle change
				if err := handler(data); err != nil {
					LogError(err, "Failed to handle config change", "path", event.Name)
					continue
				}

				LogInfo("Config reloaded", "path", relPath)
			}
		case err, ok := <-cm.watcher.Errors:
			if !ok {
				return
			}
			LogError(err, "Watcher error")
		}
	}
}

// handleSettingsChange handles settings.json changes
func (cm *ConfigManager) handleSettingsChange(data []byte) error {
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return err
	}

	cm.mu.Lock()
	cm.settings = &settings
	cm.mu.Unlock()

	return nil
}

// handleSingboxChange handles sing-box config changes
func (cm *ConfigManager) handleSingboxChange(data []byte) error {
	// Validate JSON format
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	// TODO: Restart sing-box service with new config
	return nil
}

// handleMosdnsChange handles mosdns config changes
func (cm *ConfigManager) handleMosdnsChange(data []byte) error {
	// TODO: Validate YAML format and restart mosdns service
	return nil
}

// GetSettings returns current settings
func (cm *ConfigManager) GetSettings() *Settings {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.settings
}

// UpdateSettings updates settings
func (cm *ConfigManager) UpdateSettings(settings *Settings) error {
	// Validate settings
	if err := settings.Validate(); err != nil {
		return err
	}

	// Save to file
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(cm.configDir, "settings.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil
}

// RegisterHandler registers a config change handler
func (cm *ConfigManager) RegisterHandler(relPath string, handler func([]byte) error) {
	cm.handlers[relPath] = handler
}

// BackupConfig creates a backup of current configuration
func (cm *ConfigManager) BackupConfig() (string, error) {
	// Create backup directory
	backupDir := filepath.Join(cm.configDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	// Create backup file
	timestamp := time.Now().Format("20060102-150405")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("config-%s.zip", timestamp))

	// TODO: Implement backup logic
	return backupFile, nil
}

// RestoreConfig restores configuration from backup
func (cm *ConfigManager) RestoreConfig(backupFile string) error {
	// TODO: Implement restore logic
	return nil
}

// DeleteBackup deletes a backup file
func (m *ConfigManager) DeleteBackup(backupID string) error {
	backupPath := fmt.Sprintf("%s/backups/%s", m.configDir, backupID)
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup: %v", err)
	}
	return nil
}
