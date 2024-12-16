package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	apiAuth "github.com/shipeng101/singdns/api/auth"
	apiDNS "github.com/shipeng101/singdns/api/dns"
	apiProxy "github.com/shipeng101/singdns/api/proxy"
	apiSubscription "github.com/shipeng101/singdns/api/subscription"
	apiSystem "github.com/shipeng101/singdns/api/system"
	authMiddleware "github.com/shipeng101/singdns/middleware/auth"
	svcAuth "github.com/shipeng101/singdns/service/auth"
	"github.com/shipeng101/singdns/service/config"
	"github.com/shipeng101/singdns/service/core"
	"github.com/shipeng101/singdns/service/dns"
	"github.com/shipeng101/singdns/service/proxy"
	"github.com/shipeng101/singdns/service/subscription"
	"github.com/shipeng101/singdns/service/system"
	"go.uber.org/zap"
)

var (
	configFile = flag.String("config", "config/config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 创建必要的目录
	dirs := []string{
		"logs",
		"config",
		"config/rules",
		"config/mosdns",
		"core",
		"web/dashboard",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("创建目录失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 初始化配置服务
	cfgService := config.NewService(*configFile)
	if err := cfgService.Load(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	cfg := cfgService.Get()

	// 初始化日志
	if err := system.InitLogger(cfg.Log); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化核心文件
	coreService := core.NewService()
	if err := coreService.Init(); err != nil {
		system.Error("初始化核心文件失败", zap.Error(err))
	}

	// 创建服务实例
	subscriptionSrv := subscription.NewService()
	dnsSrv := dns.NewService(cfg.DNS)
	proxySrv := proxy.NewService(cfg.Proxy, subscriptionSrv)
	authSrv := svcAuth.NewService()

	// 启动DNS服务
	if err := dnsSrv.Start(); err != nil {
		system.Error("启动DNS服务失败", zap.Error(err))
	}

	// 启动代理服务
	if err := proxySrv.Start(); err != nil {
		system.Error("启动代理服务失败", zap.Error(err))
	}

	// 创建HTTP服务器
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(authMiddleware.JWTMiddleware(authSrv))

	// 注册API路由
	apiGroup := r.Group("/api")
	{
		apiAuth.RegisterHandlers(apiGroup.Group("/auth"), authSrv)
		apiDNS.RegisterHandlers(apiGroup.Group("/dns"), dnsSrv)
		apiProxy.RegisterHandlers(apiGroup.Group("/proxy"), proxySrv)
		apiSubscription.RegisterHandlers(apiGroup.Group("/subscription"), subscriptionSrv)
		apiSystem.RegisterHandlers(apiGroup.Group("/system"), cfgService)
	}

	// 启动HTTP服务器
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		system.Info("HTTP服务器启动", zap.String("addr", addr))
		if err := r.Run(addr); err != nil {
			system.Error("HTTP服务器启动失败", zap.Error(err))
			os.Exit(1)
		}
	}()

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 关闭服务
	system.Info("正在关闭服务...")
	proxySrv.Stop()
	dnsSrv.Stop()
	system.Info("服务已关闭")
}
