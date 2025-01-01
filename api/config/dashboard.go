package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// DashboardManager 面板管理器
type DashboardManager struct {
	baseDir string
}

// NewDashboardManager 创建面板管理器
func NewDashboardManager(baseDir string) *DashboardManager {
	return &DashboardManager{
		baseDir: baseDir,
	}
}

// DashboardInfo 面板信息
type DashboardInfo struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	LastModified string `json:"last_modified"`
	Size         int64  `json:"size"`
}

// ListDashboards 列出所有面板
func (m *DashboardManager) ListDashboards() ([]DashboardInfo, error) {
	var dashboards []DashboardInfo

	// 确保目录存在
	if err := os.MkdirAll(m.baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create directory: %w", err)
	}

	// 读取目录内容
	files, err := os.ReadDir(m.baseDir)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	// 遍历文件和目录
	for _, file := range files {
		// 只处理 yacd 和 metacubexd 目录
		if !file.IsDir() || (file.Name() != "yacd" && file.Name() != "metacubexd") {
			continue
		}

		// 获取目录信息
		info, err := file.Info()
		if err != nil {
			continue
		}

		dashboards = append(dashboards, DashboardInfo{
			Name:         file.Name(),
			Path:         filepath.Join(m.baseDir, file.Name()),
			LastModified: info.ModTime().Format("2006-01-02 15:04:05"),
			Size:         info.Size(),
		})
	}

	return dashboards, nil
}

// GetDashboardPath 获取面板路径
func (m *DashboardManager) GetDashboardPath(name string) string {
	return filepath.Join(m.baseDir, name)
}

// ValidateDashboard 验证面板是否存在
func (m *DashboardManager) ValidateDashboard(name string) error {
	path := m.GetDashboardPath(name)
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("validate dashboard: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", path)
	}

	// 检查是否存在 index.html
	indexPath := filepath.Join(path, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return fmt.Errorf("index.html not found: %w", err)
	}

	return nil
}

// GetDashboardURL 获取面板访问 URL
func (m *DashboardManager) GetDashboardURL(name string, baseURL string) string {
	return fmt.Sprintf("%s/%s/", baseURL, name)
}
