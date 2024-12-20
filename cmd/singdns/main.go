package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"singdns/api"
)

var (
	configDir   = flag.String("config", "configs", "配置文件目录")
	apiAddr     = flag.String("api", ":8080", "API 服务地址")
	pprofAddr   = flag.String("pprof", ":6060", "pprof 服务地址")
	metricsAddr = flag.String("metrics", ":9090", "Prometheus metrics 服务地址")
	logConfig   = &api.LogConfig{
		Level:      "info",
		Filename:   "logs/singdns.log",
		MaxSize:    100,  // 100MB
		MaxBackups: 30,   // 保留30个备份
		MaxAge:     7,    // 保留7天
		Compress:   true, // 压缩旧日志
		Console:    true, // 同时输出到控制台
	}
)

func main() {
	flag.Parse()

	// 初始化日志系统
	if err := api.InitLogger(logConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// 创建必要的目录
	dirs := []string{
		*configDir,
		filepath.Dir(logConfig.Filename),
		filepath.Join(*configDir, "sing-box"),
		filepath.Join(*configDir, "mosdns"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			api.LogError(err, "Failed to create directory", "path", dir)
			os.Exit(1)
		}
	}

	// 创建配置管理器
	configManager, err := api.NewConfigManager(*configDir)
	if err != nil {
		api.LogError(err, "Failed to create config manager")
		os.Exit(1)
	}
	defer configManager.Close()

	// 启动性能监控
	monitor, err := api.NewMonitor(5 * time.Second)
	if err != nil {
		api.LogError(err, "Failed to create monitor")
		os.Exit(1)
	}
	defer monitor.Stop()

	// 启动 pprof 服务
	if err := api.StartProfiling(*pprofAddr); err != nil {
		api.LogError(err, "Failed to start pprof server")
		os.Exit(1)
	}

	// 启动 Prometheus metrics 服务
	if err := api.StartMetrics(*metricsAddr); err != nil {
		api.LogError(err, "Failed to start metrics server")
		os.Exit(1)
	}

	// 创建 API 服务器
	server := api.NewServer(configManager)

	// 启动 API 服务器
	go func() {
		api.LogInfo("API server starting", "address", *apiAddr)
		if err := server.Run(*apiAddr); err != nil {
			api.LogError(err, "Failed to start API server")
			os.Exit(1)
		}
	}()

	// 等待中断信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	api.LogInfo("Shutting down...")
}
