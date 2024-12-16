package core

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/shipeng101/singdns/service/system"
	"go.uber.org/zap"
)

const (
	// 核心文件目录
	CoreDir = "bin"
	// 默认版本
	DefaultSingBoxVersion = "1.8.0-rc.1"
	DefaultMosDNSVersion  = "v5.3.1"
)

// Service 核心文件管理服务接口
type Service interface {
	// 初始化核心文件
	Init() error
	// 检查更新
	CheckUpdate() error
	// 获取核心文件路径
	GetSingBoxPath() string
	GetMosDNSPath() string
}

type service struct {
	coreDir     string
	singBoxPath string
	mosDNSPath  string
	osType      string
	archType    string
	mu          sync.Mutex
}

// NewService 创建核心文件管理服务
func NewService() Service {
	osType := runtime.GOOS
	archType := runtime.GOARCH
	return &service{
		coreDir:     CoreDir,
		osType:      osType,
		archType:    archType,
		singBoxPath: filepath.Join(CoreDir, "sing-box"),
		mosDNSPath:  filepath.Join(CoreDir, "mosdns"),
	}
}

// Init 初始化核心文件
func (s *service) Init() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 创建核心文件目录
	if err := os.MkdirAll(s.coreDir, 0755); err != nil {
		return fmt.Errorf("创建核心文件目录失败: %v", err)
	}

	// 检查并下载 sing-box
	s.singBoxPath = filepath.Join(s.coreDir, s.getSingBoxName())
	if !s.checkFileExists(s.singBoxPath) {
		if err := s.downloadSingBox(); err != nil {
			return fmt.Errorf("下载sing-box失败: %v", err)
		}
	}

	// 检查并下载 mosdns
	s.mosDNSPath = filepath.Join(s.coreDir, s.getMosDNSName())
	if !s.checkFileExists(s.mosDNSPath) {
		if err := s.downloadMosDNS(); err != nil {
			return fmt.Errorf("下载mosdns失败: %v", err)
		}
	}

	// 设置执行权限
	if err := s.setExecutable(s.singBoxPath); err != nil {
		return fmt.Errorf("设置sing-box执行权限失败: %v", err)
	}
	if err := s.setExecutable(s.mosDNSPath); err != nil {
		return fmt.Errorf("设置mosdns执行权限失败: %v", err)
	}

	return nil
}

// CheckUpdate 检查更新
func (s *service) CheckUpdate() error {
	// TODO: 实现版本检查和更新逻辑
	return nil
}

// GetSingBoxPath 获取sing-box路径
func (s *service) GetSingBoxPath() string {
	return s.singBoxPath
}

// GetMosDNSPath 获取mosdns路径
func (s *service) GetMosDNSPath() string {
	return s.mosDNSPath
}

// getSingBoxName 获取sing-box文件名
func (s *service) getSingBoxName() string {
	ext := ""
	if s.osType == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("sing-box%s", ext)
}

// getMosDNSName 获取mosdns文件名
func (s *service) getMosDNSName() string {
	ext := ""
	if s.osType == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("mosdns%s", ext)
}

// checkFileExists 检查文件是否存在
func (s *service) checkFileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// setExecutable 设置文件可执行权限
func (s *service) setExecutable(path string) error {
	if s.osType != "windows" {
		return os.Chmod(path, 0755)
	}
	return nil
}

// downloadSingBox 下载sing-box
func (s *service) downloadSingBox() error {
	version := DefaultSingBoxVersion
	url := fmt.Sprintf(
		"https://github.com/SagerNet/sing-box/releases/download/v%s/sing-box-%s-%s-%s.tar.gz",
		version,
		version,
		s.osType,
		s.archType,
	)

	system.Info("正在下载sing-box",
		zap.String("version", version),
		zap.String("url", url))

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "sing-box-*.tar.gz")
	if err != nil {
		return fmt.Errorf("创建临时文件��败: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	// 写入临时文件
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	// 创建目标目录
	targetDir := filepath.Dir(s.singBoxPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	// 解压文件
	cmd := exec.Command("tar", "xzf", tmpFile.Name(), "-C", targetDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("解压文件失败: %v, 输出: %s", err, string(output))
	}

	// 移动文件到目标位置
	extractedPath := filepath.Join(targetDir, fmt.Sprintf("sing-box-%s-%s-%s/sing-box", version, s.osType, s.archType))
	if err := os.Rename(extractedPath, s.singBoxPath); err != nil {
		return fmt.Errorf("移动文件失败: %v", err)
	}

	// 清理解压目录
	extractedDir := filepath.Join(targetDir, fmt.Sprintf("sing-box-%s-%s-%s", version, s.osType, s.archType))
	if err := os.RemoveAll(extractedDir); err != nil {
		system.Warn("清理解压目录失败", zap.Error(err))
	}

	system.Info("sing-box下载完成",
		zap.String("path", s.singBoxPath))

	return nil
}

// downloadMosDNS 下载mosdns
func (s *service) downloadMosDNS() error {
	version := DefaultMosDNSVersion
	// 新版本的下载 URL 格式 (v5.x.x 版本使用新的下载地址)
	url := fmt.Sprintf(
		"https://github.com/IrineSistiana/mosdns/releases/download/%s/mosdns-%s-%s.zip",
		version,
		s.osType,
		s.archType,
	)

	system.Info("正在下载mosdns",
		zap.String("version", version),
		zap.String("url", url))

	return s.downloadAndExtract(url, "mosdns", s.mosDNSPath)
}

// downloadAndExtract 下载并解压文件
func (s *service) downloadAndExtract(url, prefix, destPath string) error {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "download-*.zip")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	// 写入临时文件
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	// 解压文件
	zipReader, err := zip.OpenReader(tmpFile.Name())
	if err != nil {
		return fmt.Errorf("打开zip文件失败: %v", err)
	}
	defer zipReader.Close()

	// 查找目标文件
	var targetFile *zip.File
	for _, file := range zipReader.File {
		if strings.Contains(file.Name, prefix) && !strings.HasSuffix(file.Name, ".zip") {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return fmt.Errorf("在zip文件中未找到%s", prefix)
	}

	// 解压目标文件
	src, err := targetFile.Open()
	if err != nil {
		return fmt.Errorf("打开zip内文件失败: %v", err)
	}
	defer src.Close()

	dest, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, src); err != nil {
		return fmt.Errorf("解压文件失败: %v", err)
	}

	system.Info("文件下载完成",
		zap.String("path", destPath))

	return nil
}
