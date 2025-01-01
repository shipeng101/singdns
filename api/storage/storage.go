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
	// Database operations
	GetDB() *gorm.DB

	// Node operations
	GetNodes() ([]models.Node, error)
	GetNodeByID(id string) (*models.Node, error)
	GetNodesWithPagination(page, pageSize int) ([]models.Node, int64, error)
	SaveNode(node *models.Node) error
	DeleteNode(id string) error
	DeleteNodes(ids []string) error
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
	GetRuleSetByTag(tag string) (*models.RuleSet, error)

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

	// SaveNodeGroupAssignment saves the assignment of a node to a group
	SaveNodeGroupAssignment(nodeID, groupID string) error
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
	// 确保数据库目录存在
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	logger := logrus.New()
	gormLogger := NewGormLogger(logger)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	storage := &SQLiteStorage{
		db:     db,
		logger: logger,
	}

	// 自动迁移数据库结构
	if err := db.AutoMigrate(
		&models.Node{},
		&models.Rule{},
		&models.Subscription{},
		&models.Settings{},
		&models.NodeGroup{},
		&models.User{},
		&models.DNSRule{},
		&models.DNSSettings{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// 手动处理 RuleSet 表的迁移
	if db.Migrator().HasTable("rule_sets") {
		// 创建临时表
		if err := db.Exec(`
			CREATE TABLE rule_sets_new (
				id TEXT PRIMARY KEY,
				tag TEXT UNIQUE,
				name TEXT NOT NULL,
				url TEXT NOT NULL,
				type TEXT NOT NULL,
				format TEXT NOT NULL,
				outbound TEXT NOT NULL,
				description TEXT,
				enabled BOOLEAN DEFAULT true,
				updated_at DATETIME
			)
		`).Error; err != nil {
			return nil, fmt.Errorf("create new table: %w", err)
		}

		// 复制数据，使用规则名称的小写形式作为 tag
		if err := db.Exec(`
			WITH numbered_rules AS (
				SELECT 
					id,
					name,
					url,
					type,
					format,
					outbound,
					description,
					enabled,
					updated_at,
					ROW_NUMBER() OVER (PARTITION BY LOWER(REPLACE(REPLACE(REPLACE(name, ' ', '-'), '.', '-'), '_', '-')) ORDER BY id) as row_num
				FROM rule_sets
			)
			INSERT INTO rule_sets_new (id, tag, name, url, type, format, outbound, description, enabled, updated_at)
			SELECT 
				id,
				CASE 
					WHEN row_num = 1 THEN LOWER(REPLACE(REPLACE(REPLACE(name, ' ', '-'), '.', '-'), '_', '-'))
					ELSE LOWER(REPLACE(REPLACE(REPLACE(name, ' ', '-'), '.', '-'), '_', '-')) || '-' || row_num
				END as tag,
				name, url, type, format, outbound, description, enabled, updated_at 
			FROM numbered_rules
		`).Error; err != nil {
			return nil, fmt.Errorf("copy data: %w", err)
		}

		// 删除旧表
		if err := db.Exec(`DROP TABLE rule_sets`).Error; err != nil {
			return nil, fmt.Errorf("drop old table: %w", err)
		}

		// 重命名新表
		if err := db.Exec(`ALTER TABLE rule_sets_new RENAME TO rule_sets`).Error; err != nil {
			return nil, fmt.Errorf("rename table: %w", err)
		}
	} else {
		// 如果表不存在，直接创建
		if err := db.AutoMigrate(&models.RuleSet{}); err != nil {
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
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
			}
			if err := db.Create(&dnsSettings).Error; err != nil {
				return nil, fmt.Errorf("failed to create default DNS settings: %v", err)
			}
			logger.Info("Created default DNS settings")
		} else {
			return nil, fmt.Errorf("failed to check DNS settings: %v", err)
		}
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

	// 如果不节点且提供了 GroupIDs，更新节点组关系
	if !isNewNode && len(node.GroupIDs) > 0 {
		// 清除原有的节点组关系
		if err := tx.Model(node).Association("NodeGroups").Clear(); err != nil {
			tx.Rollback()
			return err
		}

		// 添加新的节点组关系
		var groups []models.NodeGroup
		if err := tx.Where("id IN ?", node.GroupIDs).Find(&groups).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Model(node).Association("NodeGroups").Replace(groups); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 如果是新节点，检查是否有匹配的节点组
	if isNewNode {
		var groups []models.NodeGroup
		if err := tx.Find(&groups).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 遍历所有节点组，检查是否匹配
		var matchedGroups []models.NodeGroup
		for _, group := range groups {
			// 检查包含规则
			matched := false
			if len(group.IncludePatterns) > 0 {
				for _, pattern := range group.IncludePatterns {
					if strings.Contains(strings.ToLower(node.Name), strings.ToLower(pattern)) {
						matched = true
						break
					}
				}
			} else {
				matched = true // 如果没有包含规则，默认匹配
			}

			// 检查除则
			if matched && len(group.ExcludePatterns) > 0 {
				for _, pattern := range group.ExcludePatterns {
					if strings.Contains(strings.ToLower(node.Name), strings.ToLower(pattern)) {
						matched = false
						break
					}
				}
			}

			if matched {
				matchedGroups = append(matchedGroups, group)
			}
		}

		// 添加节点到匹配的组
		if len(matchedGroups) > 0 {
			if err := tx.Model(node).Association("NodeGroups").Replace(&matchedGroups); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// 提交事务
	return tx.Commit().Error
}

func (s *SQLiteStorage) DeleteNode(id string) error {
	// Start a transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Delete node group associations
	if err := tx.Exec("DELETE FROM node_group_nodes WHERE node_id = ?", id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete node group associations: %v", err)
	}

	// Delete the node
	if err := tx.Exec("DELETE FROM nodes WHERE id = ?", id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete node: %v", err)
	}

	// Commit transaction
	return tx.Commit().Error
}

func (s *SQLiteStorage) DeleteNodes(ids []string) error {
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

	// 清除节点组关联
	if err := tx.Exec("DELETE FROM node_group_nodes WHERE node_id IN (?)", ids).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete node group associations: %v", err)
	}

	// 删除节点
	result := tx.Exec("DELETE FROM nodes WHERE id IN (?)", ids)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete nodes: %v", result.Error)
	}

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
			"domains":     rule.Domains,
			"ips":         rule.IPs,
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

	// 如果是新订阅生成ID
	if sub.ID == "" {
		sub.ID = uuid.New().String()
	}

	// 保存订阅信息
	if err := tx.Save(sub).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 获取所有节点
	var allNodes []models.Node
	if err := tx.Find(&allNodes).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 获取所有节点组
	var groups []models.NodeGroup
	if err := tx.Preload("Nodes").Find(&groups).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 创建已分配节点的映射
	assignedNodes := make(map[string]bool)

	// 遍历所有组，重新匹配节点
	for _, group := range groups {
		// 清除当前组的节点关联
		if err := tx.Exec("DELETE FROM node_group_nodes WHERE node_group_id = ?", group.ID).Error; err != nil {
			tx.Rollback()
			return err
		}

		// 匹配节点
		var nodesToAssign []models.Node
		for _, node := range allNodes {
			// 跳过已分配的节点
			if assignedNodes[node.ID] {
				continue
			}

			matched := false
			if len(group.IncludePatterns) > 0 {
				for _, pattern := range group.IncludePatterns {
					pattern = strings.TrimSpace(pattern)
					if pattern == "" {
						continue
					}

					nodeName := strings.ToLower(node.Name)
					pattern = strings.ToLower(pattern)

					// 尝试作为正则表达式匹配
					if strings.HasPrefix(pattern, "^") && strings.HasSuffix(pattern, "$") {
						if regexp, err := regexp.Compile(pattern); err == nil {
							if regexp.MatchString(nodeName) {
								matched = true
								break
							}
						}
					} else {
						// 作为关键字匹配
						keywords := strings.Split(pattern, "|")
						for _, keyword := range keywords {
							keyword = strings.TrimSpace(keyword)
							if keyword != "" && strings.Contains(nodeName, strings.ToLower(keyword)) {
								matched = true
								break
							}
						}
					}
				}

				// 如果匹配了包含规则，检查排除规则
				if matched && len(group.ExcludePatterns) > 0 {
					for _, pattern := range group.ExcludePatterns {
						pattern = strings.TrimSpace(pattern)
						if pattern == "" {
							continue
						}

						nodeName := strings.ToLower(node.Name)
						pattern = strings.ToLower(pattern)

						// 尝试作为正则表达式匹配
						if strings.HasPrefix(pattern, "^") && strings.HasSuffix(pattern, "$") {
							if regexp, err := regexp.Compile(pattern); err == nil {
								if regexp.MatchString(nodeName) {
									matched = false
									break
								}
							}
						} else {
							// 作为关键字匹配
							keywords := strings.Split(pattern, "|")
							for _, keyword := range keywords {
								keyword = strings.TrimSpace(keyword)
								if keyword != "" && strings.Contains(nodeName, strings.ToLower(keyword)) {
									matched = false
									break
								}
							}
						}
					}
				}
			}

			if matched {
				nodesToAssign = append(nodesToAssign, node)
				assignedNodes[node.ID] = true
				s.logger.Infof("Adding node %s to group %s", node.Name, group.Name)
			}
		}

		// 批量插入节点关联
		if len(nodesToAssign) > 0 {
			valueStrings := make([]string, 0, len(nodesToAssign))
			valueArgs := make([]interface{}, 0, len(nodesToAssign)*2)
			for _, node := range nodesToAssign {
				valueStrings = append(valueStrings, "(?, ?)")
				valueArgs = append(valueArgs, group.ID, node.ID)
			}

			stmt := fmt.Sprintf("INSERT INTO node_group_nodes (node_group_id, node_id) VALUES %s",
				strings.Join(valueStrings, ","))

			if err := tx.Exec(stmt, valueArgs...).Error; err != nil {
				tx.Rollback()
				return err
			}
		}

		// 更新节点数量
		nodeCount := len(nodesToAssign)
		if err := tx.Model(&group).Update("node_count", nodeCount).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	return tx.Commit().Error
}

// DeleteSubscription deletes a subscription and all related data
func (s *SQLiteStorage) DeleteSubscription(id string) error {
	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get all nodes for this subscription
	var nodeIDs []string
	if err := tx.Model(&models.Node{}).Where("subscription_id = ?", id).Pluck("id", &nodeIDs).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get subscription nodes: %v", err)
	}

	// Delete node group associations first
	if len(nodeIDs) > 0 {
		if err := tx.Exec("DELETE FROM node_group_nodes WHERE node_id IN (?)", nodeIDs).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete node group associations: %v", err)
		}
	}

	// Delete all nodes belonging to this subscription
	if err := tx.Where("subscription_id = ?", id).Delete(&models.Node{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete subscription nodes: %v", err)
	}

	// Delete the subscription
	if err := tx.Where("id = ?", id).Delete(&models.Subscription{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete subscription: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// Settings operations
func (s *SQLiteStorage) GetSettings() (*models.Settings, error) {
	var settings models.Settings
	err := s.db.First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
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
				UpdatedAt:        time.Now().Unix(),
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

		// 保存更新后的设置
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

	// 保存分组基本信息
	if err := tx.Save(group).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 获取所有节点
	var allNodes []models.Node
	if err := tx.Find(&allNodes).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 获取所有其他组的节点分配情况
	var otherGroups []models.NodeGroup
	if err := tx.Preload("Nodes").Where("id != ?", group.ID).Find(&otherGroups).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 创建已分配节点的映射
	assignedNodes := make(map[string]bool)

	// 记录其他组已分配的节点
	for _, g := range otherGroups {
		for _, node := range g.Nodes {
			assignedNodes[node.ID] = true
		}
	}

	// 清除当前组的节点关联
	if err := tx.Exec("DELETE FROM node_group_nodes WHERE node_group_id = ?", group.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 匹配节点
	var nodesToAssign []models.Node

	if group.Tag == "others" {
		// 对于"其他节点"组，只添加未被其他组分配的节点
		for _, node := range allNodes {
			if !assignedNodes[node.ID] {
				nodesToAssign = append(nodesToAssign, node)
				s.logger.Infof("Adding unmatched node %s to others group", node.Name)
			}
		}
	} else {
		// 对于其他组，使用正常的匹配规则
		for _, node := range allNodes {
			// 跳过已分配给其他组的节点
			if assignedNodes[node.ID] {
				s.logger.Infof("Skipping node %s as it's already assigned to another group", node.Name)
				continue
			}

			matched := false
			if len(group.IncludePatterns) > 0 {
				for _, pattern := range group.IncludePatterns {
					pattern = strings.TrimSpace(pattern)
					if pattern == "" {
						continue
					}

					nodeName := strings.ToLower(node.Name)
					pattern = strings.ToLower(pattern)

					// 尝试作为正则表达式匹配
					if strings.HasPrefix(pattern, "^") && strings.HasSuffix(pattern, "$") {
						if regexp, err := regexp.Compile(pattern); err == nil {
							if regexp.MatchString(nodeName) {
								matched = true
								break
							}
						}
					} else {
						// 作为关键字匹配
						keywords := strings.Split(pattern, "|")
						for _, keyword := range keywords {
							keyword = strings.TrimSpace(keyword)
							if keyword != "" && strings.Contains(nodeName, strings.ToLower(keyword)) {
								matched = true
								break
							}
						}
					}
				}

				// 如果匹配了包含规则，检查排除规则
				if matched && len(group.ExcludePatterns) > 0 {
					for _, pattern := range group.ExcludePatterns {
						pattern = strings.TrimSpace(pattern)
						if pattern == "" {
							continue
						}

						nodeName := strings.ToLower(node.Name)
						pattern = strings.ToLower(pattern)

						// 尝试作为正则表达式匹配
						if strings.HasPrefix(pattern, "^") && strings.HasSuffix(pattern, "$") {
							if regexp, err := regexp.Compile(pattern); err == nil {
								if regexp.MatchString(nodeName) {
									matched = false
									break
								}
							}
						} else {
							// 作为关键字匹配
							keywords := strings.Split(pattern, "|")
							for _, keyword := range keywords {
								keyword = strings.TrimSpace(keyword)
								if keyword != "" && strings.Contains(nodeName, strings.ToLower(keyword)) {
									matched = false
									break
								}
							}
						}
					}
				}
			}

			if matched {
				nodesToAssign = append(nodesToAssign, node)
				assignedNodes[node.ID] = true
				s.logger.Infof("Adding node %s to group %s", node.Name, group.Name)
			}
		}
	}

	// 批量插入节点关联
	if len(nodesToAssign) > 0 {
		valueStrings := make([]string, 0, len(nodesToAssign))
		valueArgs := make([]interface{}, 0, len(nodesToAssign)*2)
		for _, node := range nodesToAssign {
			valueStrings = append(valueStrings, "(?, ?)")
			valueArgs = append(valueArgs, group.ID, node.ID)
		}

		stmt := fmt.Sprintf("INSERT INTO node_group_nodes (node_group_id, node_id) VALUES %s",
			strings.Join(valueStrings, ","))

		if err := tx.Exec(stmt, valueArgs...).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 更新节点数量
	group.NodeCount = len(nodesToAssign)
	if err := tx.Model(group).Update("node_count", group.NodeCount).Error; err != nil {
		tx.Rollback()
		return err
	}

	s.logger.Infof("Node group %s updated with %d nodes", group.Name, group.NodeCount)

	// 提交事务
	return tx.Commit().Error
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
	// 如果是新规则集
	if ruleSet.ID == "" {
		ruleSet.ID = uuid.New().String()
	}

	// 如果 tag 为空，使用规则集名称的小写形式作为 tag
	if ruleSet.Tag == "" {
		// 移除特殊字符，将空格替换为短横线
		tag := strings.ToLower(ruleSet.Name)
		tag = regexp.MustCompile(`[^a-z0-9\-]+`).ReplaceAllString(tag, "-")
		tag = regexp.MustCompile(`-+`).ReplaceAllString(tag, "-")
		tag = strings.Trim(tag, "-")
		ruleSet.Tag = tag
	}

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

	// 如果有节点，清除节点组关联
	if len(nodeIDs) > 0 {
		// 使用 IN 查询一次性删除所有关联
		if err := tx.Exec("DELETE FROM node_group_nodes WHERE node_id IN (?)", nodeIDs).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 删除该订阅下的所有节点
	if err := tx.Exec("DELETE FROM nodes WHERE subscription_id = ?", subscriptionID).Error; err != nil {
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

	// 忽略 "record not found" 错误的日志记录
	if err == gorm.ErrRecordNotFound {
		return
	}

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

// GetRuleSetByTag returns a rule set by its tag
func (s *SQLiteStorage) GetRuleSetByTag(tag string) (*models.RuleSet, error) {
	var ruleSet models.RuleSet
	err := s.db.Where("tag = ?", tag).First(&ruleSet).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &ruleSet, nil
}

// SaveNodeGroupAssignment saves the assignment of a node to a group
func (s *SQLiteStorage) SaveNodeGroupAssignment(nodeID, groupID string) error {
	// Start a transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Insert the node group assignment
	if err := tx.Exec("INSERT INTO node_group_nodes (node_id, node_group_id) VALUES (?, ?)", nodeID, groupID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to save node group assignment: %v", err)
	}

	// Commit transaction
	return tx.Commit().Error
}
