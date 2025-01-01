package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"singdns/api"
	"singdns/api/config"
	"singdns/api/network"
	"singdns/api/proxy"
	"singdns/api/storage"
	"syscall"
	"time"

	"singdns/api/models"

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
			Path: "/web/public/yacd",
		}); err != nil {
			return fmt.Errorf("set dashboard: %w", err)
		}
		// 保存默认设置到数据库
		if err := db.SaveSettings(settings); err != nil {
			return fmt.Errorf("save settings: %w", err)
		}
	}

	switch configType {
	case "mosdns":
		g := config.NewMosDNSGenerator()
		g.SetStorage(db)
		generator = g
	case "singbox":
		g := config.NewSingBoxGenerator(settings, db, filepath.Join("configs", "sing-box", "config.json"))
		g.SetStorage(db)
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
	if configType == "mosdns" {
		configDir = filepath.Join("configs", "mosdns")
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// 写入配置文件
	configPath := filepath.Join(configDir, "config.json")
	if configType == "mosdns" {
		configPath = filepath.Join(configDir, "config.yaml")
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	fmt.Printf("Config file generated: %s\n", configPath)
	return nil
}

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel) // 设置为 Debug 级别
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Parse command line flags
	var (
		configPath string
		dbPath     string
		logLevel   string
		jwtSecret  string
	)

	flag.StringVar(&configPath, "config", "config.yaml", "Path to config file")
	flag.StringVar(&dbPath, "db", "data/singdns.db", "Path to SQLite database file")
	flag.StringVar(&logLevel, "log-level", "debug", "Log level (debug, info, warn, error)")
	flag.StringVar(&jwtSecret, "secret", "your-secret-key", "JWT secret key")
	flag.Parse()

	// Set log level from flag
	if level, err := logrus.ParseLevel(logLevel); err == nil {
		logger.SetLevel(level)
	}

	logger.Info("Starting SingDNS...")
	logger.Debugf("Config path: %s", configPath)
	logger.Debugf("Database path: %s", dbPath)
	logger.Debugf("Log level: %s", logLevel)

	// Initialize database
	logger.Info("Initializing database...")
	db, err := storage.NewSQLiteStorage(dbPath)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize database")
	}
	defer db.Close()
	logger.Info("Database initialized successfully")

	// 生成配置文件
	logger.Info("Generating configurations...")
	if err := generateConfig("singbox"); err != nil {
		logger.WithError(err).Fatal("Failed to generate sing-box config")
	}
	if err := generateConfig("mosdns"); err != nil {
		logger.WithError(err).Fatal("Failed to generate mosdns config")
	}
	logger.Info("Configurations generated successfully")

	// Initialize proxy manager
	logger.Info("Initializing proxy manager...")
	proxyManager := proxy.NewManager(logger, "configs/sing-box/config.json", db)
	logger.Info("Proxy manager initialized successfully")

	// Initialize network manager
	logger.Info("Initializing network manager...")
	netManager := network.NewManager()
	logger.Info("Network manager initialized successfully")

	// Initialize API server
	logger.Info("Initializing API server...")
	serverConfig := &api.Config{
		UpdateInterval: 24 * time.Hour,
	}
	server := api.NewServer(db, proxyManager, logger, serverConfig)
	logger.Info("API server initialized successfully")

	// Start API server in background
	go func() {
		logger.Info("Starting API server...")
		if err := server.Run(); err != nil {
			logger.WithError(err).Fatal("Failed to start API server")
		}
	}()

	// 设置 DNS 重定向
	logger.Info("Setting up DNS redirect...")
	if err := netManager.SetupDNSRedirect(); err != nil {
		logger.WithError(err).Warn("Failed to setup DNS redirect")
	}
	logger.Info("DNS redirect setup completed")

	// 启动 mosdns 服务
	logger.Info("Starting mosdns service...")
	if err := proxyManager.StartService("mosdns"); err != nil {
		logger.WithError(err).Fatal("Failed to start mosdns")
	}
	logger.Info("mosdns service started successfully")

	// 启动 sing-box 服务
	logger.Info("Starting sing-box service...")
	if err := proxyManager.StartService("singbox"); err != nil {
		logger.WithError(err).Fatal("Failed to start sing-box")
	}
	logger.Info("sing-box service started successfully")

	// 等待中断信号
	logger.Info("Waiting for interrupt signal...")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	logger.Info("Interrupt signal received, shutting down...")

	// 清理 DNS 重定向
	logger.Info("Cleaning up DNS redirect...")
	if err := netManager.CleanupDNSRedirect(); err != nil {
		logger.WithError(err).Warn("Failed to cleanup DNS redirect")
	}
	logger.Info("DNS redirect cleanup completed")

	// 停止 mosdns 服务
	logger.Info("Stopping mosdns service...")
	if err := proxyManager.StopService("mosdns"); err != nil {
		logger.WithError(err).Warn("Failed to stop mosdns")
	}
	logger.Info("mosdns service stopped")

	// 停止 sing-box 服务
	logger.Info("Stopping sing-box service...")
	if err := proxyManager.StopService("singbox"); err != nil {
		logger.WithError(err).Warn("Failed to stop sing-box")
	}
	logger.Info("sing-box service stopped")

	logger.Info("SingDNS shutdown completed")
}
