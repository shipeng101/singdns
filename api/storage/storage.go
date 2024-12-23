package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"singdns/api/models"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
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

	// Close database connection
	Close() error
}

// SQLiteStorage implements Storage interface using SQLite
type SQLiteStorage struct {
	db     *gorm.DB
	logger *logrus.Logger
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

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// 修改 rules 表的列类型
	if err := db.Exec("ALTER TABLE rules ADD COLUMN domains_new JSON").Error; err != nil {
		logger.Warnf("Failed to add domains_new column: %v", err)
	}
	if err := db.Exec("ALTER TABLE rules ADD COLUMN ips_new JSON").Error; err != nil {
		logger.Warnf("Failed to add ips_new column: %v", err)
	}

	// 迁移数据
	if err := db.Exec("UPDATE rules SET domains_new = domains, ips_new = ips").Error; err != nil {
		logger.Warnf("Failed to migrate data: %v", err)
	}

	// 删除旧列并重命名新列
	if err := db.Exec("ALTER TABLE rules DROP COLUMN domains").Error; err != nil {
		logger.Warnf("Failed to drop domains column: %v", err)
	}
	if err := db.Exec("ALTER TABLE rules DROP COLUMN ips").Error; err != nil {
		logger.Warnf("Failed to drop ips column: %v", err)
	}
	if err := db.Exec("ALTER TABLE rules RENAME COLUMN domains_new TO domains").Error; err != nil {
		logger.Warnf("Failed to rename domains column: %v", err)
	}
	if err := db.Exec("ALTER TABLE rules RENAME COLUMN ips_new TO ips").Error; err != nil {
		logger.Warnf("Failed to rename ips column: %v", err)
	}

	// 修改 node_groups 表的列类型
	if err := db.Exec("ALTER TABLE node_groups ADD COLUMN include_patterns_new JSON").Error; err != nil {
		logger.Warnf("Failed to add include_patterns_new column: %v", err)
	}
	if err := db.Exec("ALTER TABLE node_groups ADD COLUMN exclude_patterns_new JSON").Error; err != nil {
		logger.Warnf("Failed to add exclude_patterns_new column: %v", err)
	}

	// 迁移数据
	if err := db.Exec("UPDATE node_groups SET include_patterns_new = include_patterns, exclude_patterns_new = exclude_patterns").Error; err != nil {
		logger.Warnf("Failed to migrate node groups data: %v", err)
	}

	// 删除旧列并重命名新列
	if err := db.Exec("ALTER TABLE node_groups DROP COLUMN include_patterns").Error; err != nil {
		logger.Warnf("Failed to drop include_patterns column: %v", err)
	}
	if err := db.Exec("ALTER TABLE node_groups DROP COLUMN exclude_patterns").Error; err != nil {
		logger.Warnf("Failed to drop exclude_patterns column: %v", err)
	}
	if err := db.Exec("ALTER TABLE node_groups RENAME COLUMN include_patterns_new TO include_patterns").Error; err != nil {
		logger.Warnf("Failed to rename include_patterns column: %v", err)
	}
	if err := db.Exec("ALTER TABLE node_groups RENAME COLUMN exclude_patterns_new TO exclude_patterns").Error; err != nil {
		logger.Warnf("Failed to rename exclude_patterns column: %v", err)
	}

	// Auto migrate schema
	if err := db.AutoMigrate(
		&models.Node{},
		&models.Rule{},
		&models.Subscription{},
		&models.Settings{},
		&models.NodeGroup{},
		&models.RuleSet{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return &SQLiteStorage{
		db:     db,
		logger: logger,
	}, nil
}

// Node operations
func (s *SQLiteStorage) GetNodes() ([]models.Node, error) {
	var nodes []models.Node
	if err := s.db.Preload("NodeGroups").Find(&nodes).Error; err != nil {
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

	// Get all nodes with their groups
	if err := s.db.Preload("NodeGroups").Find(&nodes).Error; err != nil {
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

	// 如果不是新节点且提供了 GroupIDs，更新节点组关系
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

			// 检查排除规则
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
	return s.db.Delete(&models.Node{}, "id = ?", id).Error
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
	if sub.ID == "" {
		return s.db.Create(sub).Error
	}
	return s.db.Save(sub).Error
}

func (s *SQLiteStorage) DeleteSubscription(id string) error {
	return s.db.Delete(&models.Subscription{}, "id = ?", id).Error
}

// Settings operations
func (s *SQLiteStorage) GetSettings() (*models.Settings, error) {
	var settings models.Settings
	if err := s.db.First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default settings if not found
			return &models.Settings{
				Theme:          "light",
				Language:       "zh_CN",
				UpdateInterval: 24,
				ProxyPort:      1080,
				APIPort:        9090,
				LogLevel:       "info",
			}, nil
		}
		return nil, err
	}
	return &settings, nil
}

func (s *SQLiteStorage) SaveSettings(settings *models.Settings) error {
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
	// 设置默认值
	if group.Mode == "" {
		group.Mode = "select"
	}
	if !group.Active {
		group.Active = true
	}

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

	// 如果是新建或者更新
	if err := tx.Save(group).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 清除原有的节点关系
	if err := tx.Model(group).Association("Nodes").Clear(); err != nil {
		tx.Rollback()
		return err
	}

	// 获取所有节点
	var allNodes []models.Node
	if err := tx.Find(&allNodes).Error; err != nil {
		tx.Rollback()
		return err
	}

	s.logger.Infof("Processing node group %s with include patterns: %v, exclude patterns: %v",
		group.Name, group.IncludePatterns, group.ExcludePatterns)

	// 匹配节点
	var matchedNodes []models.Node
	for _, node := range allNodes {
		// 检查包含规则
		matched := false
		if len(group.IncludePatterns) > 0 {
			for _, pattern := range group.IncludePatterns {
				pattern = strings.TrimSpace(pattern)
				if pattern == "" {
					continue
				}

				// 分割多个关键字
				keywords := strings.Split(pattern, "|")
				for _, keyword := range keywords {
					keyword = strings.TrimSpace(keyword)
					if keyword == "" {
						continue
					}

					// 不区分大小写的字符串匹配
					nodeName := strings.ToLower(node.Name)
					keyword = strings.ToLower(keyword)

					// 使用 strings.Contains 进行匹配
					if strings.Contains(nodeName, keyword) {
						s.logger.Infof("Node %s matched include pattern %s", node.Name, keyword)
						matched = true
						break
					}
				}
				if matched {
					break
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

				// 分割多个关键字
				keywords := strings.Split(pattern, "|")
				for _, keyword := range keywords {
					keyword = strings.TrimSpace(keyword)
					if keyword == "" {
						continue
					}

					// 不区分大小写的字符串匹配
					nodeName := strings.ToLower(node.Name)
					keyword = strings.ToLower(keyword)

					// 使用 strings.Contains 进行匹配
					if strings.Contains(nodeName, keyword) {
						s.logger.Infof("Node %s matched exclude pattern %s", node.Name, keyword)
						matched = false
						break
					}
				}
				if !matched {
					break
				}
			}
		}

		if matched {
			s.logger.Infof("Adding node %s to group %s", node.Name, group.Name)
			matchedNodes = append(matchedNodes, node)
		}
	}

	// 添加匹配的节点到分组
	if len(matchedNodes) > 0 {
		if err := tx.Model(group).Association("Nodes").Replace(&matchedNodes); err != nil {
			tx.Rollback()
			return err
		}
	}

	// 更新节点数量
	group.NodeCount = len(matchedNodes)
	if err := tx.Save(group).Error; err != nil {
		tx.Rollback()
		return err
	}

	s.logger.Infof("Node group %s updated with %d nodes", group.Name, group.NodeCount)
	return tx.Commit().Error
}

func (s *SQLiteStorage) DeleteNodeGroup(id string) error {
	return s.db.Delete(&models.NodeGroup{}, "id = ?", id).Error
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
	result := s.db.Where("subscription_id = ?", subscriptionID).Delete(&models.Node{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete nodes: %v", result.Error)
	}
	return nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
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

// GetDB returns the underlying database connection
func (s *SQLiteStorage) GetDB() *gorm.DB {
	return s.db
}
