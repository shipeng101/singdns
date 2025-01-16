package api

import (
	"singdns/api/models"
	"time"
)

// ProxyManager defines the interface for proxy service management
type ProxyManager interface {
	Start() error
	Stop() error
	IsRunning() bool
	GetVersion() string
	GetStartTime() time.Time
	GetConfigPath() string
	GetHealthData() map[string]models.HealthData
	GetUptime() int64
	GetConnectionsCount() int
	GetNodes() []models.Node
	CreateNode(node *models.Node) error
	UpdateNode(node *models.Node) error
	DeleteNode(id string) error
	TestNode(id string) (int64, error)
	GetNodeHealth(id string) *models.HealthData
	GetNodeGroups() []models.NodeGroup
	CreateNodeGroup(group *models.NodeGroup) error
	UpdateNodeGroup(group *models.NodeGroup) error
	DeleteNodeGroup(id string) error
	GetRules() []models.Rule
	CreateRule(rule *models.Rule) error
	UpdateRule(rule *models.Rule) error
	DeleteRule(id string) error
	GetSubscriptions() []models.Subscription
	CreateSubscription(sub *models.Subscription) error
	UpdateSubscription(sub *models.Subscription) error
	DeleteSubscription(id string) error
	UpdateSubscriptionNodes(id string) error
	RefreshSubscription(id string) error
	GetTrafficStats() *models.TrafficStats
	GetRealtimeTraffic() *models.TrafficStats
}

// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	GetConfig() (*models.Config, error)
	SaveConfig(config *models.Config) error
	LoadConfig() error
	GetConfigPath() string
	GetConfigDir() string
	GetBinDir() string
	GetDataDir() string
	GetLogDir() string
	GetWorkDir() string
	GetVersion() string
	GetUptime() int64
	IsRunning() bool
	Start() error
	Stop() error
	Restart() error
}

// AuthManager defines the interface for authentication management
type AuthManager interface {
	Login(username, password string) (string, error)
	Register(username, password string) error
	ValidateToken(token string) (*models.User, error)
	GetUser(username string) (*models.User, error)
	UpdatePassword(username, oldPassword, newPassword string) error
}
