package ruleset

import (
	"fmt"
	"singdns/api/models"
	"singdns/api/storage"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Updater handles automatic updates of rule sets
type Updater struct {
	storage storage.Storage
	logger  *logrus.Logger
	ticker  *time.Ticker
	done    chan bool
	mutex   sync.Mutex
}

// NewUpdater creates a new rule set updater
func NewUpdater(storage storage.Storage, logger *logrus.Logger) *Updater {
	return &Updater{
		storage: storage,
		logger:  logger,
		done:    make(chan bool),
	}
}

// Start starts the automatic update process
func (u *Updater) Start(interval time.Duration) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.ticker != nil {
		return
	}

	u.ticker = time.NewTicker(interval)
	go u.run()
}

// Stop stops the automatic update process
func (u *Updater) Stop() {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.ticker != nil {
		u.ticker.Stop()
		u.ticker = nil
		u.done <- true
	}
}

// run is the main update loop
func (u *Updater) run() {
	for {
		select {
		case <-u.ticker.C:
			u.updateAll()
		case <-u.done:
			return
		}
	}
}

// updateAll updates all enabled rule sets
func (u *Updater) updateAll() {
	ruleSets, err := u.storage.GetRuleSets()
	if err != nil {
		u.logger.Errorf("Failed to get rule sets: %v", err)
		return
	}

	for _, ruleSet := range ruleSets {
		if !ruleSet.Enabled {
			continue
		}

		if err := u.updateOne(&ruleSet); err != nil {
			u.logger.Errorf("Failed to update rule set %s: %v", ruleSet.ID, err)
		}
	}
}

// updateOne updates a single rule set
func (u *Updater) updateOne(ruleSet *models.RuleSet) error {
	// Download rule set
	data, err := DownloadRuleSet(ruleSet.URL)
	if err != nil {
		return fmt.Errorf("failed to download rule set: %v", err)
	}

	// Parse rule set
	parser := NewParser(ruleSet.Type)
	if parser == nil {
		return fmt.Errorf("unsupported rule set type")
	}

	rules, err := parser.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse rule set: %v", err)
	}

	// Create or update rule
	rule := &models.Rule{
		ID:          fmt.Sprintf("ruleset-%s", ruleSet.ID),
		Name:        ruleSet.Name,
		Type:        ruleSet.Format,
		Outbound:    ruleSet.Outbound,
		Description: ruleSet.Description,
		Enabled:     ruleSet.Enabled,
		Priority:    100, // Default priority for rule sets
		UpdatedAt:   time.Now(),
	}

	// Set values
	rule.Values = rules

	if err := u.storage.SaveRule(rule); err != nil {
		return fmt.Errorf("failed to save rule: %v", err)
	}

	// Update rule set timestamp
	ruleSet.UpdatedAt = time.Now()
	if err := u.storage.SaveRuleSet(ruleSet); err != nil {
		return fmt.Errorf("failed to update rule set: %v", err)
	}

	u.logger.Infof("Updated rule set %s successfully", ruleSet.ID)
	return nil
}

// UpdateRuleSet updates a rule set from remote URL
func (u *Updater) UpdateRuleSet(ruleSet *models.RuleSet) error {
	// Download rule set
	data, err := DownloadRuleSet(ruleSet.URL)
	if err != nil {
		return fmt.Errorf("failed to download rule set: %v", err)
	}

	// Parse rule set
	parser := NewParser(ruleSet.Type)
	if parser == nil {
		return fmt.Errorf("unsupported rule set type")
	}

	rules, err := parser.Parse(data)
	if err != nil {
		return fmt.Errorf("failed to parse rule set: %v", err)
	}

	// Create or update rule
	rule := &models.Rule{
		ID:          fmt.Sprintf("ruleset-%s", ruleSet.ID),
		Name:        ruleSet.Name,
		Type:        ruleSet.Format,
		Outbound:    ruleSet.Outbound,
		Description: ruleSet.Description,
		Enabled:     ruleSet.Enabled,
		Priority:    100, // Default priority for rule sets
		UpdatedAt:   time.Now(),
	}

	// Set values
	rule.Values = models.StringArray(rules)

	if err := u.storage.SaveRule(rule); err != nil {
		return fmt.Errorf("failed to save rule: %v", err)
	}

	// Update rule set timestamp
	ruleSet.UpdatedAt = time.Now()
	if err := u.storage.SaveRuleSet(ruleSet); err != nil {
		return fmt.Errorf("failed to update rule set: %v", err)
	}

	return nil
}
