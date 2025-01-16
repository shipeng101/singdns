package subscription

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

type SubscriptionManager struct {
	subscriptions map[string]*Subscription
}

type Subscription struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"` // clash, singbox
	URL        string `json:"url"`
	Interval   int    `json:"interval"` // hours
	LastUpdate string `json:"lastUpdate"`
}

func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subscriptions: make(map[string]*Subscription),
	}
}

func (m *SubscriptionManager) AddSubscription(sub *Subscription) error {
	if sub.ID == "" {
		return fmt.Errorf("subscription ID cannot be empty")
	}
	m.subscriptions[sub.ID] = sub
	return nil
}

func (m *SubscriptionManager) GetSubscription(id string) (*Subscription, error) {
	sub, ok := m.subscriptions[id]
	if !ok {
		return nil, fmt.Errorf("subscription not found: %s", id)
	}
	return sub, nil
}

func (m *SubscriptionManager) ListSubscriptions() []*Subscription {
	subs := make([]*Subscription, 0, len(m.subscriptions))
	for _, sub := range m.subscriptions {
		subs = append(subs, sub)
	}
	return subs
}

func (m *SubscriptionManager) UpdateSubscription(sub *Subscription) error {
	if sub.ID == "" {
		return fmt.Errorf("subscription ID cannot be empty")
	}
	if _, ok := m.subscriptions[sub.ID]; !ok {
		return fmt.Errorf("subscription not found: %s", sub.ID)
	}
	m.subscriptions[sub.ID] = sub
	return nil
}

func (m *SubscriptionManager) DeleteSubscription(id string) error {
	if _, ok := m.subscriptions[id]; !ok {
		return fmt.Errorf("subscription not found: %s", id)
	}
	delete(m.subscriptions, id)
	return nil
}

// GetBase64URL returns the base64 encoded subscription URL
func (m *SubscriptionManager) GetBase64URL(id string) (string, error) {
	sub, err := m.GetSubscription(id)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString([]byte(sub.URL)), nil
}

// GetBase64QRCode returns the base64 encoded QR code content for the subscription
func (m *SubscriptionManager) GetBase64QRCode(id string) (string, error) {
	sub, err := m.GetSubscription(id)
	if err != nil {
		return "", err
	}

	// Generate QR code content based on subscription type
	var content string
	switch strings.ToLower(sub.Type) {
	case "clash":
		content = fmt.Sprintf("clash://%s", url.QueryEscape(sub.URL))
	case "singbox":
		content = fmt.Sprintf("sb://%s", url.QueryEscape(sub.URL))
	default:
		content = sub.URL
	}

	return base64.URLEncoding.EncodeToString([]byte(content)), nil
}

// ParseBase64URL decodes a base64 encoded subscription URL
func (m *SubscriptionManager) ParseBase64URL(encodedURL string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 URL: %v", err)
	}
	return string(decoded), nil
}
