package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"singdns/api/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Storage defines the interface for data persistence
type Storage interface {
	// Node operations
	GetNodes() ([]models.Node, error)
	GetNodeByID(id string) (*models.Node, error)
	GetNodesWithPagination(page, pageSize int) ([]models.Node, int64, error)
	SaveNode(node *models.Node) error
	DeleteNode(id string) error
	DeleteNodesBySubscriptionID(subscriptionID string) error

	// Rule operations
	GetRules() ([]models.Rule, error)
	GetRuleByID(id string) (*models.Rule, error)
	SaveRule(rule *models.Rule) error
	DeleteRule(id string) error

	// Subscription operations
	GetSubscriptions() ([]models.Subscription, error)
	GetSubscriptionByID(id string) (*models.Subscription, error)
	SaveSubscription(sub *models.Subscription) error
	DeleteSubscription(id string) error

	// Settings operations
	GetSettings() (*models.Settings, error)
	SaveSettings(settings *models.Settings) error
	UpdateSettings(settings *models.Settings) error

	// NodeGroup operations
	GetNodeGroups() ([]models.NodeGroup, error)
	GetNodeGroupByID(id string) (*models.NodeGroup, error)
	SaveNodeGroup(group *models.NodeGroup) error
	DeleteNodeGroup(id string) error

	// RuleSet operations
	GetRuleSets() ([]models.RuleSet, error)
	GetRuleSetByID(id string) (*models.RuleSet, error)
	SaveRuleSet(ruleSet *models.RuleSet) error
	DeleteRuleSet(id string) error

	// User operations
	GetUser(username string) (*models.User, error)
	UpdateUser(user *models.User) error

	// DNS 规则
	GetDNSRules() ([]models.DNSRule, error)
	GetDNSRuleByID(id string) (*models.DNSRule, error)
	SaveDNSRule(rule *models.DNSRule) error
	DeleteDNSRule(id string) error

	// DNS 设置
	GetDNSSettings() (*models.DNSSettings, error)
	SaveDNSSettings(settings *models.DNSSettings) error

	// Close database connection
	Close() error
}

// SQLiteStorage implements Storage interface using SQLite
type SQLiteStorage struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// GetDB returns the underlying database connection
func (s *SQLiteStorage) GetDB() *gorm.DB {
	return s.db
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string) (Storage, error) {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	logger.SetOutput(os.Stdout)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	// Configure GORM
	config := &gorm.Config{
		Logger: NewGormLogger(logger),
	}

	// Open database
	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Auto migrate database schema
	if err := db.AutoMigrate(
		&models.Node{},
		&models.Rule{},
		&models.RuleSet{},
		&models.Subscription{},
		&models.Settings{},
		&models.NodeGroup{},
		&models.User{},
		&models.DNSRule{},
		&models.DNSSettings{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	storage := &SQLiteStorage{
		db:     db,
		logger: logger,
	}

	// Create default admin user if not exists
	var adminUser models.User
	err = db.Where("username = ?", "admin").First(&adminUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Hash default password
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
			if err != nil {
				return nil, fmt.Errorf("failed to hash password: %v", err)
			}

			// Create admin user
			adminUser = models.User{
				Username: "admin",
				Password: string(hashedPassword),
				Role:     "admin",
			}

			// Use Create instead of Save to ensure a new record is created
			if err := db.Create(&adminUser).Error; err != nil {
				return nil, fmt.Errorf("failed to create admin user: %v", err)
			}

			logger.Info("Created default admin user")
		} else {
			return nil, fmt.Errorf("failed to check admin user: %v", err)
		}
	}

	// Create default DNS settings if not exists
	var dnsSettings models.DNSSettings
	err = db.First(&dnsSettings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			dnsSettings = models.DNSSettings{
				ID:               "default",
				EnableAdBlock:    true,
				AdBlockRuleCount: 0,
				LastUpdate:       time.Now().Unix(),
				Domestic:         "223.5.5.5",
				DomesticType:     "udp",
				SingboxDNS:       "https://dns.google/dns-query",
				SingboxDNSType:   "doh",
				EDNSClientSubnet: "1.0.1.0",
			}
			if err := db.Create(&dnsSettings).Error; err != nil {
				return nil, fmt.Errorf("failed to create default DNS settings: %v", err)
			}
			logger.Info("Created default DNS settings")
		} else {
			return nil, fmt.Errorf("failed to check DNS settings: %v", err)
		}
	}

	// Create default node groups if not exist
	defaultGroups := []models.NodeGroup{
		{
			ID:              "all",
			Name:            "全部 🌏",
			Tag:             "ALL",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{".*"},
			Description:     "所有节点",
		},
		{
			ID:              "hk",
			Name:            "香港 🇭🇰",
			Tag:             "HK",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"香港", "HK", "Hong Kong", "HongKong"},
			Description:     "香港地区节点",
		},
		{
			ID:              "tw",
			Name:            "台湾 🇹🇼",
			Tag:             "TW",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"台湾", "TW", "Taiwan"},
			Description:     "台湾地区节点",
		},
		{
			ID:              "jp",
			Name:            "日本 🇯🇵",
			Tag:             "JP",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"日本", "JP", "Japan"},
			Description:     "日本地区节点",
		},
		{
			ID:              "sg",
			Name:            "新加坡 🇸🇬",
			Tag:             "SG",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"新加坡", "SG", "Singapore"},
			Description:     "新加坡地区节点",
		},
		{
			ID:              "us",
			Name:            "美国 🇺🇸",
			Tag:             "US",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"美国", "US", "United States", "USA"},
			Description:     "美国地区节点",
		},
		{
			ID:              "kr",
			Name:            "韩国 🇰🇷",
			Tag:             "KR",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"韩国", "KR", "Korea"},
			Description:     "韩国地区节点",
		},
		{
			ID:              "gb",
			Name:            "英国 🇬🇧",
			Tag:             "GB",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"英国", "GB", "United Kingdom", "UK"},
			Description:     "英国地区节点",
		},
		{
			ID:              "de",
			Name:            "德国 🇩🇪",
			Tag:             "DE",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"德国", "DE", "Germany"},
			Description:     "德国地区节点",
		},
		{
			ID:              "ca",
			Name:            "加拿大 🇨🇦",
			Tag:             "CA",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"加拿大", "CA", "Canada"},
			Description:     "加拿大地区节点",
		},
		{
			ID:              "au",
			Name:            "澳大利亚 🇦🇺",
			Tag:             "AU",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"澳大利亚", "AU", "Australia"},
			Description:     "澳大利亚地区节点",
		},
		{
			ID:              "in",
			Name:            "印度 🇮🇳",
			Tag:             "IN",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"印度", "IN", "India"},
			Description:     "印度地区节点",
		},
		{
			ID:              "ru",
			Name:            "俄罗斯 🇷🇺",
			Tag:             "RU",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"俄罗斯", "RU", "Russia"},
			Description:     "俄罗斯地区节点",
		},
		{
			ID:              "fr",
			Name:            "法国 🇫🇷",
			Tag:             "FR",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"法国", "FR", "France"},
			Description:     "法国地区节点",
		},
		{
			ID:              "nl",
			Name:            "荷兰 🇳🇱",
			Tag:             "NL",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"荷兰", "NL", "Netherlands"},
			Description:     "荷兰地区节点",
		},
		{
			ID:              "tr",
			Name:            "土耳其 🇹🇷",
			Tag:             "TR",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"土耳其", "TR", "Turkey"},
			Description:     "土耳其地区节点",
		},
		{
			ID:              "br",
			Name:            "巴西 🇧🇷",
			Tag:             "BR",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"巴西", "BR", "Brazil"},
			Description:     "巴西地区节点",
		},
		{
			ID:              "ar",
			Name:            "阿根廷 🇦🇷",
			Tag:             "AR",
			Mode:            "select",
			Active:          true,
			IncludePatterns: models.StringArray{"阿根廷", "AR", "Argentina"},
			Description:     "阿根廷地区节点",
		},
	}

	// 清理所有现有的节点组
	if err := db.Exec("DELETE FROM node_group_nodes").Error; err != nil {
		logger.WithError(err).Error("Failed to clean up node group associations")
		return nil, err
	}
	if err := db.Exec("DELETE FROM node_groups").Error; err != nil {
		logger.WithError(err).Error("Failed to clean up node groups")
		return nil, err
	}
	logger.Info("Cleaned up existing node groups")

	// 创建默认节点组
	for _, group := range defaultGroups {
		group.CreatedAt = time.Now()
		group.UpdatedAt = time.Now()
		if err := db.Create(&group).Error; err != nil {
			logger.WithError(err).Errorf("Failed to create default node group: %s", group.Name)
			continue
		}
		logger.Infof("Created default node group: %s", group.Name)
	}

	return storage, nil
}

// Node operations
func (s *SQLiteStorage) GetNodes() ([]models.Node, error) {
	var nodes []models.Node
	// 使用 Select("DISTINCT nodes.*") 来确保节点不重复
	if err := s.db.Select("DISTINCT nodes.*").
		Preload("NodeGroups").
		Find(&nodes).Error; err != nil {
		return nil, err
	}

	// 设置每个节点的 GroupIDs
	for i := range nodes {
		nodes[i].GroupIDs = make([]string, len(nodes[i].NodeGroups))
		for j, group := range nodes[i].NodeGroups {
			nodes[i].GroupIDs[j] = group.ID
		}
	}

	return nodes, nil
}

func (s *SQLiteStorage) GetNodeByID(id string) (*models.Node, error) {
	var node models.Node
	if err := s.db.First(&node, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (s *SQLiteStorage) GetNodesWithPagination(page, pageSize int) ([]models.Node, int64, error) {
	var nodes []models.Node
	var total int64

	// 使用 Select("DISTINCT nodes.*") 来确保节点不重复
	if err := s.db.Select("DISTINCT nodes.*").
		Preload("NodeGroups").
		Find(&nodes).Error; err != nil {
		return nil, 0, err
	}

	total = int64(len(nodes))

	// 设置每个节点的 GroupIDs
	for i := range nodes {
		nodes[i].GroupIDs = make([]string, len(nodes[i].NodeGroups))
		for j, group := range nodes[i].NodeGroups {
			nodes[i].GroupIDs[j] = group.ID
		}
	}

	// Calculate pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(nodes) {
		return []models.Node{}, total, nil
	}
	if end > len(nodes) {
		end = len(nodes)
	}

	return nodes[start:end], total, nil
}

func (s *SQLiteStorage) SaveNode(node *models.Node) error {
	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 如果是新建节点
	isNewNode := node.ID == ""
	if isNewNode {
		node.ID = uuid.New().String()
	}

	// 如果是更新现有节点，获取原有节点数据
	var existingNode models.Node
	if !isNewNode {
		err := tx.First(&existingNode, "id = ?", node.ID).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			tx.Rollback()
			return err
		}
		// 如果找到了现有节点，且新节点的 CheckedAt 为空，保留原有的 CheckedAt
		if err == nil && node.CheckedAt.IsZero() {
			node.CheckedAt = existingNode.CheckedAt
		}
	}

	// 保存节点基本信息
	if err := tx.Save(node).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 获取所有节点组
	var groups []models.NodeGroup
	if err := tx.Find(&groups).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 遍历所有节点组，检查是否匹配
	var matchedGroups []models.NodeGroup
	for _, group := range groups {
		// "全部"组特殊处理
		if group.Name == "全部 🌏" {
			matchedGroups = append(matchedGroups, group)
			continue
		}

		// 检查包含规则
		matched := false
		if len(group.IncludePatterns) > 0 {
			for _, pattern := range group.IncludePatterns {
				pattern = strings.TrimSpace(pattern)
				if pattern == "" {
					continue
				}

				// 尝试编译正则表达式
				re, err := regexp.Compile(pattern)
				if err == nil {
					// 如果是有效的正则表达式，使用正则匹配
					if re.MatchString(node.Name) {
						matched = true
						break
					}
				} else {
					// 如果不是有效的正则表达式，使用关键字匹配
					if strings.Contains(strings.ToLower(node.Name), strings.ToLower(pattern)) {
						matched = true
						break
					}
				}
			}
		} else {
			matched = true // 如果没有包含规则，默认匹配
		}

		// 检查排除规则
		if matched && len(group.ExcludePatterns) > 0 {
			for _, pattern := range group.ExcludePatterns {
				pattern = strings.TrimSpace(pattern)
				if pattern == "" {
					continue
				}

				// 尝试编译正则表达式
				re, err := regexp.Compile(pattern)
				if err == nil {
					// 如果是有效的正则表达式，使用正则匹配
					if re.MatchString(node.Name) {
						matched = false
						break
					}
				} else {
					// 如果不是有效的正则表达式，使用关键字匹配
					if strings.Contains(strings.ToLower(node.Name), strings.ToLower(pattern)) {
						matched = false
						break
					}
				}
			}
		}

		if matched {
			matchedGroups = append(matchedGroups, group)
		}
	}

	// 更新节点组关联
	if err := tx.Model(node).Association("NodeGroups").Replace(&matchedGroups); err != nil {
		tx.Rollback()
		return err
	}

	// 更新每个匹配组的节点数量
	for _, group := range matchedGroups {
		var count int64
		count = tx.Model(&group).Association("Nodes").Count()
		if err := tx.Model(&group).Update("node_count", count).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (s *SQLiteStorage) DeleteNode(id string) error {
	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	s.logger.Infof("开始删除节点 ID: %s", id)

	// 先删除节点与节点组的关联关系
	if err := tx.Exec("DELETE FROM node_group_nodes WHERE node_id = ?", id).Error; err != nil {
		s.logger.Errorf("删除节点关联关系失败: %v", err)
		tx.Rollback()
		return err
	}
	s.logger.Infof("已删除节点的组关联关系")

	// 删除节点
	result := tx.Delete(&models.Node{}, "id = ?", id)
	if result.Error != nil {
		s.logger.Errorf("删除节点失败: %v", result.Error)
		tx.Rollback()
		return result.Error
	}
	s.logger.Infof("已删除节点，影响行数: %d", result.RowsAffected)

	return tx.Commit().Error
}

// Rule operations
func (s *SQLiteStorage) GetRules() ([]models.Rule, error) {
	var rules []models.Rule
	if err := s.db.Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

func (s *SQLiteStorage) GetRuleByID(id string) (*models.Rule, error) {
	var rule models.Rule
	if err := s.db.First(&rule, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (s *SQLiteStorage) SaveRule(rule *models.Rule) error {
	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查是否已存在同ID的规则
	var existingRule models.Rule
	err := tx.First(&existingRule, "id = ?", rule.ID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return err
	}

	// 如果规则已存在，更新现有规则
	if err == nil {
		// 使用 Updates 而不是 Save，以避免创建新记录
		if err := tx.Model(&existingRule).Updates(map[string]interface{}{
			"name":        rule.Name,
			"type":        rule.Type,
			"outbound":    rule.Outbound,
			"description": rule.Description,
			"enabled":     rule.Enabled,
			"priority":    rule.Priority,
			"values":      rule.Values,
			"updated_at":  rule.UpdatedAt,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// 如果规则不存在，创建新规则
		if err := tx.Create(rule).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (s *SQLiteStorage) DeleteRule(id string) error {
	return s.db.Delete(&models.Rule{}, "id = ?", id).Error
}

// Subscription operations
func (s *SQLiteStorage) GetSubscriptions() ([]models.Subscription, error) {
	var subs []models.Subscription
	if err := s.db.Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

func (s *SQLiteStorage) GetSubscriptionByID(id string) (*models.Subscription, error) {
	var sub models.Subscription
	if err := s.db.First(&sub, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (s *SQLiteStorage) SaveSubscription(sub *models.Subscription) error {
	if sub.ID == "" {
		return s.db.Create(sub).Error
	}
	return s.db.Save(sub).Error
}

func (s *SQLiteStorage) DeleteSubscription(id string) error {
	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取该订阅下的所有节点
	var nodes []models.Node
	if err := tx.Where("subscription_id = ?", id).Find(&nodes).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 清除每个节点的节点组关联
	for _, node := range nodes {
		if err := tx.Model(&node).Association("NodeGroups").Clear(); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 删除该订阅下的所有节点
	if err := tx.Delete(&models.Node{}, "subscription_id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除订阅
	if err := tx.Delete(&models.Subscription{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Settings operations
func (s *SQLiteStorage) GetSettings() (*models.Settings, error) {
	var settings models.Settings
	err := s.db.First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			now := time.Now().Unix()
			// Create default settings
			settings = models.Settings{
				ID:               "1",
				Theme:            "light",
				ThemeMode:        "light",
				Language:         "zh-CN",
				UpdateInterval:   24,
				ProxyPort:        7890,
				APIPort:          8080,
				DNSPort:          53,
				LogLevel:         "info",
				EnableAutoUpdate: true,
				UpdatedAt:        now,
				CreatedAt:        now,
			}

			// 设置默认仪表盘
			dashboard := &models.DashboardSettings{
				Type: "yacd",
				Path: "bin/web/yacd",
			}
			dashboardJSON, err := json.Marshal(dashboard)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal dashboard: %v", err)
			}
			settings.Dashboard = string(dashboardJSON)

			// 设置默认 SingBox 模式
			mode := &models.SingboxModeSettings{
				Mode: "rule",
			}
			modeJSON, err := json.Marshal(mode)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal singbox mode: %v", err)
			}
			settings.SingboxMode = string(modeJSON)

			// 设置默认入站模式
			if err := settings.SetInboundMode("tun"); err != nil {
				return nil, fmt.Errorf("failed to set default inbound mode: %v", err)
			}

			// Use Create instead of Save to ensure a new record is created
			if err := s.db.Create(&settings).Error; err != nil {
				return nil, fmt.Errorf("failed to create default settings: %v", err)
			}

			s.logger.Info("Created default settings")
			return &settings, nil
		}
		return nil, fmt.Errorf("failed to get settings: %v", err)
	}

	// 如果仪表盘为空，设置默认值并保存
	if settings.Dashboard == "" {
		dashboard := &models.DashboardSettings{
			Type: "yacd",
			Path: "bin/web/yacd",
		}
		dashboardJSON, err := json.Marshal(dashboard)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal dashboard: %v", err)
		}
		settings.Dashboard = string(dashboardJSON)
	}

	// 如果 SingBox 模式为空，设置默认值并保存
	if settings.SingboxMode == "" {
		mode := &models.SingboxModeSettings{
			Mode: "rule",
		}
		modeJSON, err := json.Marshal(mode)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal singbox mode: %v", err)
		}
		settings.SingboxMode = string(modeJSON)
	}

	// 保存更新后的设置
	if settings.Dashboard == "" || settings.SingboxMode == "" {
		if err := s.db.Save(&settings).Error; err != nil {
			return nil, fmt.Errorf("failed to save settings: %v", err)
		}
	}

	return &settings, nil
}

func (s *SQLiteStorage) SaveSettings(settings *models.Settings) error {
	return s.db.Save(settings).Error
}

func (s *SQLiteStorage) UpdateSettings(settings *models.Settings) error {
	return s.db.Save(settings).Error
}

// NodeGroup operations
func (s *SQLiteStorage) GetNodeGroups() ([]models.NodeGroup, error) {
	var groups []models.NodeGroup
	if err := s.db.Preload("Nodes").Find(&groups).Error; err != nil {
		return nil, err
	}

	// 计算每个节点数量
	for i := range groups {
		groups[i].NodeCount = len(groups[i].Nodes)
		s.logger.Infof("Group %s has %d nodes", groups[i].Name, groups[i].NodeCount)
	}

	return groups, nil
}

func (s *SQLiteStorage) GetNodeGroupByID(id string) (*models.NodeGroup, error) {
	var group models.NodeGroup
	if err := s.db.Preload("Nodes").First(&group, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// 计算节点数量
	group.NodeCount = len(group.Nodes)

	return &group, nil
}

func (s *SQLiteStorage) SaveNodeGroup(group *models.NodeGroup) error {
	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 如果是新分组生成ID
	if group.ID == "" {
		group.ID = uuid.New().String()
	}

	// 获取所有节点
	var allNodes []models.Node
	if err := tx.Find(&allNodes).Error; err != nil {
		tx.Rollback()
		return err
	}

	s.logger.Infof("Processing node group %s with include patterns: %v, exclude patterns: %v", group.Name, group.IncludePatterns, group.ExcludePatterns)

	// 匹配节点
	var matchedNodes []models.Node
	for _, node := range allNodes {
		// 默认不匹配任何节点
		matched := false

		// "全部"分组特殊处理
		if group.Name == "全部 🌏" {
			matched = true
		} else {
			// 检查包含规则
			if len(group.IncludePatterns) > 0 {
				s.logger.Infof("Checking include patterns for node %s", node.Name)
				for _, pattern := range group.IncludePatterns {
					pattern = strings.TrimSpace(pattern)
					if pattern == "" {
						continue
					}

					// 尝试编译正则表达式
					re, err := regexp.Compile(pattern)
					if err == nil {
						// 如果是有效的正则表达式，使用正则匹配
						if re.MatchString(node.Name) {
							s.logger.Infof("Node %s matched regex pattern %s", node.Name, pattern)
							matched = true
							break
						}
					} else {
						// 如果不是有效的正则表达式，使用关键字匹配
						if strings.Contains(strings.ToLower(node.Name), strings.ToLower(pattern)) {
							s.logger.Infof("Node %s matched keyword %s", node.Name, pattern)
							matched = true
							break
						}
					}
				}
			}

			// 检查排除规则
			if matched && len(group.ExcludePatterns) > 0 {
				for _, pattern := range group.ExcludePatterns {
					pattern = strings.TrimSpace(pattern)
					if pattern == "" {
						continue
					}

					// 尝试编译正则表达式
					re, err := regexp.Compile(pattern)
					if err == nil {
						// 如果是有效的正则表达式，使用正则匹配
						if re.MatchString(node.Name) {
							s.logger.Infof("Node %s excluded by regex pattern %s", node.Name, pattern)
							matched = false
							break
						}
					} else {
						// 如果不是有效的正则表达式，使用关键字匹配
						if strings.Contains(strings.ToLower(node.Name), strings.ToLower(pattern)) {
							s.logger.Infof("Node %s excluded by keyword %s", node.Name, pattern)
							matched = false
							break
						}
					}
				}
			}
		}

		if matched {
			s.logger.Infof("Adding node %s to group %s", node.Name, group.Name)
			matchedNodes = append(matchedNodes, node)
		} else {
			s.logger.Infof("Node %s did not match group %s", node.Name, group.Name)
		}
	}

	// 更新节点数量
	group.NodeCount = len(matchedNodes)

	// 保存分组信息
	if err := tx.Save(group).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 清除旧的节点关联
	if err := tx.Model(group).Association("Nodes").Clear(); err != nil {
		tx.Rollback()
		return err
	}

	// 添加匹配的节点到分组
	if len(matchedNodes) > 0 {
		if err := tx.Model(group).Association("Nodes").Replace(&matchedNodes); err != nil {
			tx.Rollback()
			return err
		}
	}

	s.logger.Infof("Node group %s updated with %d nodes", group.Name, group.NodeCount)

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *SQLiteStorage) DeleteNodeGroup(id string) error {
	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 先清除节点组与节点关联关系
	if err := tx.Exec("DELETE FROM node_group_nodes WHERE node_group_id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除节点组
	if err := tx.Delete(&models.NodeGroup{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// RuleSet operations
func (s *SQLiteStorage) GetRuleSets() ([]models.RuleSet, error) {
	var ruleSets []models.RuleSet
	if err := s.db.Find(&ruleSets).Error; err != nil {
		return nil, err
	}
	return ruleSets, nil
}

func (s *SQLiteStorage) GetRuleSetByID(id string) (*models.RuleSet, error) {
	var ruleSet models.RuleSet
	if err := s.db.First(&ruleSet, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ruleSet, nil
}

func (s *SQLiteStorage) SaveRuleSet(ruleSet *models.RuleSet) error {
	if ruleSet.ID == "" {
		return s.db.Create(ruleSet).Error
	}
	return s.db.Save(ruleSet).Error
}

func (s *SQLiteStorage) DeleteRuleSet(id string) error {
	return s.db.Delete(&models.RuleSet{}, "id = ?", id).Error
}

// DeleteNodesBySubscriptionID deletes all nodes for a subscription
func (s *SQLiteStorage) DeleteNodesBySubscriptionID(subscriptionID string) error {
	// 开始事务
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取该订阅下的所有节点ID
	var nodeIDs []string
	if err := tx.Model(&models.Node{}).Where("subscription_id = ?", subscriptionID).Pluck("id", &nodeIDs).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 清除这些节点与节点组的关联关系
	if len(nodeIDs) > 0 {
		if err := tx.Exec("DELETE FROM node_group_nodes WHERE node_id IN (?)", nodeIDs).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 删除节点
	if err := tx.Delete(&models.Node{}, "subscription_id = ?", subscriptionID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// User operations
func (s *SQLiteStorage) GetUser(username string) (*models.User, error) {
	var user models.User
	err := s.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound && username == "admin" {
			// Create default admin user
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
			if err != nil {
				return nil, fmt.Errorf("failed to hash password: %v", err)
			}

			user = models.User{
				Username: "admin",
				Password: string(hashedPassword),
				Role:     "admin",
			}

			// Use Create instead of Save to ensure a new record is created
			if err := s.db.Create(&user).Error; err != nil {
				return nil, fmt.Errorf("failed to create admin user: %v", err)
			}

			s.logger.Info("Created default admin user")
			return &user, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *SQLiteStorage) UpdateUser(user *models.User) error {
	return s.db.Save(user).Error
}

// DNS 规则
func (s *SQLiteStorage) GetDNSRules() ([]models.DNSRule, error) {
	var rules []models.DNSRule
	if err := s.db.Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// GetDNSRuleByID returns a DNS rule by ID
func (s *SQLiteStorage) GetDNSRuleByID(id string) (*models.DNSRule, error) {
	var rule models.DNSRule
	if err := s.db.First(&rule, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// SaveDNSRule saves a DNS rule
func (s *SQLiteStorage) SaveDNSRule(rule *models.DNSRule) error {
	return s.db.Save(rule).Error
}

// DeleteDNSRule deletes a DNS rule
func (s *SQLiteStorage) DeleteDNSRule(id string) error {
	return s.db.Delete(&models.DNSRule{}, "id = ?", id).Error
}

// GetDNSSettings returns DNS settings
func (s *SQLiteStorage) GetDNSSettings() (*models.DNSSettings, error) {
	var settings models.DNSSettings
	err := s.db.First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		// 返回默认设置
		settings = models.DNSSettings{
			ID:               "default",
			EnableAdBlock:    true,
			AdBlockRuleCount: 0,
			LastUpdate:       0,
		}
		if err := s.db.Create(&settings).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &settings, nil
}

// SaveDNSSettings saves DNS settings
func (s *SQLiteStorage) SaveDNSSettings(settings *models.DNSSettings) error {
	return s.db.Save(settings).Error
}

// GormLogger adapts logrus logger to GORM logger interface
type GormLogger struct {
	logger *logrus.Logger
}

func NewGormLogger(logger *logrus.Logger) *GormLogger {
	return &GormLogger{logger: logger}
}

// LogMode implements gorm.Logger interface
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Info implements gorm.Logger interface
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Infof(msg, data...)
}

// Warn implements gorm.Logger interface
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Warnf(msg, data...)
}

// Error implements gorm.Logger interface
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.Errorf(msg, data...)
}

// Trace implements gorm.Logger interface
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		l.logger.Errorf("[%.3fms] [rows:%d] %s; error: %v", float64(elapsed.Nanoseconds())/1e6, rows, sql, err)
		return
	}

	if elapsed > time.Second {
		l.logger.Warnf("[%.3fms] [rows:%d] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
		return
	}

	l.logger.Debugf("[%.3fms] [rows:%d] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
}

// 创建默认设置
func createDefaultSettings(db *gorm.DB) error {
	settings := &models.Settings{
		ID:               "default",
		Theme:            "light",
		ThemeMode:        "system",
		Language:         "zh-CN",
		UpdateInterval:   24,
		ProxyPort:        1080,
		APIPort:          8080,
		DNSPort:          53,
		LogLevel:         "info",
		EnableAutoUpdate: true,
		UpdatedAt:        time.Now().Unix(),
	}

	// 设置默认仪表盘
	dashboard := &models.DashboardSettings{
		Type: "yacd",
		Path: "bin/web/yacd",
	}
	dashboardJSON, err := json.Marshal(dashboard)
	if err != nil {
		return fmt.Errorf("failed to marshal dashboard: %v", err)
	}
	settings.Dashboard = string(dashboardJSON)

	// 设置默认 SingBox 模式
	mode := &models.SingboxModeSettings{
		Mode: "rule",
	}
	modeJSON, err := json.Marshal(mode)
	if err != nil {
		return fmt.Errorf("failed to marshal singbox mode: %v", err)
	}
	settings.SingboxMode = string(modeJSON)

	// 设置默认入站模式
	if err := settings.SetInboundMode("tun"); err != nil {
		return fmt.Errorf("failed to set default inbound mode: %v", err)
	}

	// 保存设置
	if err := db.Create(settings).Error; err != nil {
		return fmt.Errorf("create settings: %w", err)
	}

	return nil
}
