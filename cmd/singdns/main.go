package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"singdns/api"
	"singdns/api/auth"
	"singdns/api/config"
	"singdns/api/proxy"
	"singdns/api/storage"
	"syscall"
	"time"

	"singdns/api/models"

	"net"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

func generateConfig(configType string) error {
	var generator config.ConfigGenerator

	// 初始化数据库
	db, err := storage.NewSQLiteStorage("data/singdns.db")
	if err != nil {
		return fmt.Errorf("init database: %w", err)
	}
	defer db.Close()

	// 获取设置
	settings, err := db.GetSettings()
	if err != nil {
		return fmt.Errorf("get settings: %w", err)
	}

	// 如果没有设置，使用默认设置
	if settings == nil {
		settings = &models.Settings{
			ID:             "1",
			Theme:          "light",
			ThemeMode:      "system",
			Language:       "zh-CN",
			UpdateInterval: 24,

			ProxyPort:        7890,
			APIPort:          8080,
			DNSPort:          53,
			LogLevel:         "info",
			EnableAutoUpdate: true,
			UpdatedAt:        time.Now().Unix(),
		}
		// 设置默认仪表盘
		if err := settings.SetDashboard(&models.DashboardSettings{
			Type: "yacd",
		}); err != nil {
			return fmt.Errorf("set dashboard: %w", err)
		}
		// 保存默认设置到数据库
		if err := db.SaveSettings(settings); err != nil {
			return fmt.Errorf("save settings: %w", err)
		}
	}

	switch configType {
	case "singbox":
		g := config.NewSingBoxGenerator(db)
		generator = g
	default:
		return fmt.Errorf("unknown config type: %s", configType)
	}

	data, err := generator.GenerateConfig()
	if err != nil {
		return err
	}

	// 确保配置目录存在
	configDir := filepath.Join("configs", "sing-box")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// 写入配置文件
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	fmt.Printf("Config file generated: %s\n", configPath)
	return nil
}

func enableIPForwarding() error {
	// 尝试直接写入
	if err := os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1\n"), 0644); err == nil {
		return nil
	}

	// 如果直接写入失败，尝试使用 sysctl 命令
	cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := cmd.Run(); err == nil {
		return nil
	}

	// 如果普通 sysctl 失败，尝试使用 sudo
	cmd = exec.Command("sudo", "sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %v", err)
	}

	return nil
}

// 获取本机网络信息
func getLocalNetwork() (string, string, error) {
	// 获取默认路由的网卡
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("get default route: %v", err)
	}
	fields := strings.Fields(string(output))
	if len(fields) < 5 {
		return "", "", fmt.Errorf("invalid route output: %s", string(output))
	}
	iface := fields[4]

	// 获取网卡的 IP 地址和网段
	cmd = exec.Command("ip", "addr", "show", iface)
	output, err = cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("get interface addr: %v", err)
	}

	// 使用正则表达式匹配 IP/掩码
	re := regexp.MustCompile(`inet\s+(\d+\.\d+\.\d+\.\d+/\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "", "", fmt.Errorf("no IPv4 address found for interface %s", iface)
	}

	// 解析 IP 和网段
	ip, ipNet, err := net.ParseCIDR(matches[1])
	if err != nil {
		return "", "", fmt.Errorf("parse CIDR: %v", err)
	}

	return ip.String(), ipNet.String(), nil
}

func setupNetwork() error {
	// 只开启 IP 转发
	if err := enableIPForwarding(); err != nil {
		return fmt.Errorf("enable IP forwarding: %v", err)
	}
	return nil
}

func main() {
	// 设置网络
	if err := setupNetwork(); err != nil {
		log.Printf("Warning: Failed to setup network: %v", err)
	} else {
		log.Println("Network setup completed")
	}

	// 初始化配置目录
	if err := os.MkdirAll("configs/sing-box", 0755); err != nil {
		log.Fatalf("Failed to create config directory: %v", err)
	}

	// 下载规则集
	if err := config.DownloadRuleSets(); err != nil {
		log.Fatalf("Failed to download rule sets: %v", err)
	}

	// 确保规则集存在
	if err := config.EnsureRuleSetsExist(); err != nil {
		log.Fatalf("Failed to verify rule sets: %v", err)
	}

	// 检查是否是生成配置命令
	if len(os.Args) > 1 && os.Args[1] == "config" && os.Args[2] == "generate" {
		if len(os.Args) <= 3 {
			fmt.Println("Usage: singdns config generate singbox")
			os.Exit(1)
		}
		if err := generateConfig(os.Args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Parse command line flags
	var (
		configPath string
		dbPath     string
		logLevel   string
		jwtSecret  string
	)

	flag.StringVar(&configPath, "config", "config.yaml", "Path to config file")
	flag.StringVar(&dbPath, "db", "data/singdns.db", "Path to SQLite database file")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&jwtSecret, "secret", "your-secret-key", "JWT secret key")
	flag.Parse()

	// 初始化日志
	logger := logrus.New()
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		fmt.Printf("Invalid log level: %s\n", logLevel)
		os.Exit(1)
	}
	logger.SetLevel(level)

	// 初始化数据库
	db, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		logger.WithError(err).Fatal("Failed to init database")
	}
	defer db.Close()

	// 初始化配置服务
	configService := config.NewConfigService("configs", db)
	if configService == nil {
		logger.Fatal("Failed to create config service")
	}

	// 初始化代理管理器
	proxyManager := proxy.NewManager(logger, "configs/sing-box/config.json", ".")

	// 初始化认证管理器
	authManager := auth.NewManager([]byte(jwtSecret), db)

	// 创建 API 服务器
	server := api.NewServer(db, proxyManager, logger, &api.Config{
		UpdateInterval: 24 * time.Hour,
		Auth:           authManager,
	})

	// 启动服务器
	if err := server.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}

	// 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// 关闭服务器
	server.Stop()
}
