package dashboard

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/shipeng101/singdns/service/system"
	"go.uber.org/zap"
)

// DashboardType 面板类型
type DashboardType string

const (
	DashboardYacd      DashboardType = "yacd"
	DashboardMetacubex DashboardType = "metacubexd"
)

// Config 面板配置
type Config struct {
	// 当前面板
	Current DashboardType `json:"current"`
	// 配置文件路径
	ConfigFile string `json:"config_file"`
	// 面板资源目录
	ResourceDir string `json:"resource_dir"`
	// API端点
	APIEndpoint string `json:"api_endpoint"`
}

// Service 面板服务接口
type Service interface {
	// 启动服务
	Start() error
	// 停止服务
	Stop() error
	// 获取配置
	GetConfig() *Config
	// 更新配置
	UpdateConfig(config *Config) error
	// 切换面板
	SwitchDashboard(dashboard DashboardType) error
	// 获取当前面板
	GetCurrentDashboard() DashboardType
	// 获取面板URL
	GetDashboardURL() string
}

type service struct {
	config *Config
	mu     sync.Mutex
}

// NewService 创建面板服务
func NewService() Service {
	return &service{
		config: &Config{
			Current:     DashboardYacd,
			ConfigFile:  "config/dashboard.json",
			ResourceDir: "web/dashboard",
			APIEndpoint: "http://127.0.0.1:10802",
		},
	}
}

// Start 启动服务
func (s *service) Start() error {
	// 加载配置
	if err := s.loadConfig(); err != nil {
		system.Warn("加载面板配置失败，使用默认配置",
			zap.Error(err))
	}

	// 检查资源目录
	if err := s.checkResources(); err != nil {
		return fmt.Errorf("检查面板资源失败: %v", err)
	}

	system.Info("面板服务启动成功",
		zap.String("dashboard", string(s.config.Current)))

	return nil
}

// Stop 停止服务
func (s *service) Stop() error {
	return nil
}

// GetConfig 获取配置
func (s *service) GetConfig() *Config {
	return s.config
}

// UpdateConfig 更新配置
func (s *service) UpdateConfig(config *Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config

	// 保存配置
	return s.saveConfig()
}

// SwitchDashboard 切换面板
func (s *service) SwitchDashboard(dashboard DashboardType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if dashboard != DashboardYacd && dashboard != DashboardMetacubex {
		return fmt.Errorf("不支持的面板类型")
	}

	s.config.Current = dashboard

	// 保存配置
	if err := s.saveConfig(); err != nil {
		return err
	}

	system.Info("面板切换成功",
		zap.String("dashboard", string(dashboard)))

	return nil
}

// GetCurrentDashboard 获取当前面板
func (s *service) GetCurrentDashboard() DashboardType {
	return s.config.Current
}

// GetDashboardURL 获取面板URL
func (s *service) GetDashboardURL() string {
	switch s.config.Current {
	case DashboardYacd:
		return fmt.Sprintf("%s/yacd/", s.config.ResourceDir)
	case DashboardMetacubex:
		return fmt.Sprintf("%s/metacubexd/", s.config.ResourceDir)
	default:
		return ""
	}
}

// loadConfig 加载配置
func (s *service) loadConfig() error {
	data, err := os.ReadFile(s.config.ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，保存默认配置
			return s.saveConfig()
		}
		return err
	}

	return json.Unmarshal(data, s.config)
}

// saveConfig 保存配置
func (s *service) saveConfig() error {
	// 创建配置目录
	if err := os.MkdirAll(filepath.Dir(s.config.ConfigFile), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.config.ConfigFile, data, 0644)
}

// checkResources 检查面板资源
func (s *service) checkResources() error {
	// 检查资源目录
	if err := os.MkdirAll(s.config.ResourceDir, 0755); err != nil {
		return fmt.Errorf("创建资源目录失败: %v", err)
	}

	// 检查yacd目录
	yacdDir := filepath.Join(s.config.ResourceDir, "yacd")
	if _, err := os.Stat(yacdDir); os.IsNotExist(err) {
		system.Info("正在下载yacd面板资源...")
		if err := s.downloadYacd(yacdDir); err != nil {
			return fmt.Errorf("下载yacd面板失败: %v", err)
		}
	}

	// 检查metacubexd目录
	metacubexDir := filepath.Join(s.config.ResourceDir, "metacubexd")
	if _, err := os.Stat(metacubexDir); os.IsNotExist(err) {
		system.Info("正在下载metacubexd面板资源...")
		if err := s.downloadMetacubex(metacubexDir); err != nil {
			return fmt.Errorf("下载metacubexd面板失败: %v", err)
		}
	}

	return nil
}

// downloadYacd 下载yacd面板
func (s *service) downloadYacd(dir string) error {
	// 创建目录
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 下载最新release
	resp, err := http.Get("https://api.github.com/repos/haishanh/yacd/releases/latest")
	if err != nil {
		return fmt.Errorf("获取yacd最新版本失败: %v", err)
	}
	defer resp.Body.Close()

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("解析yacd版本信息失败: %v", err)
	}

	// 查找dist.zip资源
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == "dist.zip" {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("未找到yacd下载资源")
	}

	// 下载zip文件
	resp, err = http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("下载yacd失败: %v", err)
	}
	defer resp.Body.Close()

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "yacd-*.zip")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 保存zip文件
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("保存yacd文件失败: %v", err)
	}

	// 解压文件
	if err := s.unzip(tmpFile.Name(), dir); err != nil {
		return fmt.Errorf("解压yacd文件失败: %v", err)
	}

	return nil
}

// downloadMetacubex 下载metacubexd面板
func (s *service) downloadMetacubex(dir string) error {
	// 创建目录
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 下载最新release
	resp, err := http.Get("https://api.github.com/repos/MetaCubeX/metacubexd/releases/latest")
	if err != nil {
		return fmt.Errorf("获取metacubexd最新版本失败: %v", err)
	}
	defer resp.Body.Close()

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("解析metacubexd版本信息失败: %v", err)
	}

	// 查找dist.zip资源
	var downloadURL string
	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".zip") {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("未找到metacubexd下载资源")
	}

	// 下载zip文件
	resp, err = http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("下载metacubexd失败: %v", err)
	}
	defer resp.Body.Close()

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "metacubexd-*.zip")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 保存zip文件
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("保存metacubexd文件失败: %v", err)
	}

	// 解压文件
	if err := s.unzip(tmpFile.Name(), dir); err != nil {
		return fmt.Errorf("解压metacubexd文件失败: %v", err)
	}

	return nil
}

// unzip 解压zip文件
func (s *service) unzip(zipFile string, dest string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	// 创建目标目录
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// 清理目标目录中的现有文件
	if err := os.RemoveAll(dest); err != nil {
		return err
	}

	// 重新创建目标目录
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// 检查路径是否在目标目录内
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法的文件路径: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
