package dashboard

import (
	"fmt"

	"github.com/shipeng101/singdns/pkg/types"
)

type service struct {
	config *types.DashboardConfig
}

// NewService 创建仪表盘服务
func NewService() types.DashboardService {
	return &service{
		config: &types.DashboardConfig{
			Current:     "default",
			ConfigFile:  "config/dashboard.json",
			ResourceDir: "web/dashboard",
			APIEndpoint: "http://localhost:8080/api",
		},
	}
}

// Start 启动服务
func (s *service) Start() error {
	return nil
}

// Stop 停止服务
func (s *service) Stop() error {
	return nil
}

// GetConfig 获取配置
func (s *service) GetConfig() *types.DashboardConfig {
	return s.config
}

// UpdateConfig 更新配置
func (s *service) UpdateConfig(config *types.DashboardConfig) error {
	s.config = config
	return nil
}

// SwitchDashboard 切换面板
func (s *service) SwitchDashboard(dashboard string) error {
	s.config.Current = dashboard
	return nil
}

// GetCurrentDashboard 获取当前面板
func (s *service) GetCurrentDashboard() string {
	return s.config.Current
}

// GetDashboardURL 获取面板URL
func (s *service) GetDashboardURL() string {
	return fmt.Sprintf("%s/dashboard/%s", s.config.APIEndpoint, s.config.Current)
}
