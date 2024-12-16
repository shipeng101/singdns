package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/shipeng101/singdns/api"
	"github.com/shipeng101/singdns/pkg/types"
	"github.com/shipeng101/singdns/service/auth"
	"github.com/shipeng101/singdns/service/config"
	"github.com/shipeng101/singdns/service/core"
	"github.com/shipeng101/singdns/service/dashboard"
	"github.com/shipeng101/singdns/service/dns"
	"github.com/shipeng101/singdns/service/proxy"
	"github.com/shipeng101/singdns/service/subscription"
	"github.com/shipeng101/singdns/service/system"
	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	logConfig := types.LogConfig{
		Level:      "info",
		File:       "logs/singdns.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     7,
		Compress:   true,
	}
	system.InitLogger(logConfig)

	// 创建服务
	services := &types.Services{
		AuthService:         auth.NewService(),
		DNSService:          dns.NewService(),
		ProxyService:        proxy.NewService(),
		SubscriptionService: subscription.NewService(),
		CoreService:         core.NewService(),
		DashboardService:    dashboard.NewService(),
		ConfigService:       config.NewService("config/config.yaml"),
	}

	// 启动服务
	if err := startServices(services); err != nil {
		system.Fatal("启动服务失败", zap.Error(err))
	}

	// 创建路由
	r := gin.Default()
	api.RegisterRoutes(r, services)

	// 启动HTTP服务
	go func() {
		if err := r.Run(":8080"); err != nil {
			system.Fatal("启动HTTP服务失败", zap.Error(err))
		}
	}()

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 停止服务
	if err := stopServices(services); err != nil {
		system.Fatal("停止服务失败", zap.Error(err))
	}

	system.Info("服务已停止")
}

// startServices 启动所有服务
func startServices(services *types.Services) error {
	// 初始化核心服务
	if err := services.CoreService.Init(); err != nil {
		return fmt.Errorf("初始化核心服务失败: %v", err)
	}

	// 启动配置服务
	if err := services.ConfigService.Load(); err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	// 启动DNS服务
	if err := services.DNSService.Start(); err != nil {
		return fmt.Errorf("启动DNS服务失败: %v", err)
	}

	// 启动代理服务
	if err := services.ProxyService.Start(); err != nil {
		return fmt.Errorf("启动代理服务失败: %v", err)
	}

	// 启动订阅服务
	if err := services.SubscriptionService.Start(); err != nil {
		return fmt.Errorf("启动订阅服务失败: %v", err)
	}

	// 启动仪表盘服务
	if err := services.DashboardService.Start(); err != nil {
		return fmt.Errorf("启动仪表盘服务失败: %v", err)
	}

	return nil
}

// stopServices 停止所有服务
func stopServices(services *types.Services) error {
	// 停止仪表盘服务
	if err := services.DashboardService.Stop(); err != nil {
		return fmt.Errorf("停止仪表盘服务失败: %v", err)
	}

	// 停止订阅服务
	if err := services.SubscriptionService.Stop(); err != nil {
		return fmt.Errorf("停止订阅服务失败: %v", err)
	}

	// 停止代理服务
	if err := services.ProxyService.Stop(); err != nil {
		return fmt.Errorf("停止代理服务失败: %v", err)
	}

	// 停止DNS服务
	if err := services.DNSService.Stop(); err != nil {
		return fmt.Errorf("停止DNS服务失败: %v", err)
	}

	return nil
}
