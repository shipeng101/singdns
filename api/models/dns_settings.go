package models

// DNSSettings DNS 设置
type DNSSettings struct {
	ID               string `json:"id" gorm:"primaryKey"`
	Final            string `json:"final"`
	Strategy         string `json:"strategy"`
	DisableCache     bool   `json:"disable_cache"`
	DisableExpire    bool   `json:"disable_expire"`
	IndependentCache bool   `json:"independent_cache"`
	ReverseMapping   bool   `json:"reverse_mapping"`
	EnableAdBlock    bool   `json:"enable_ad_block" gorm:"default:true"`
	AdBlockRuleCount int    `json:"ad_block_rule_count" gorm:"default:0"`
	LastUpdate       int64  `json:"last_update" gorm:"default:0"`

	// DNS 服务器设置
	Domestic         string `json:"domestic" gorm:"column:domestic"`
	DomesticType     string `json:"domestic_type" gorm:"column:domestic_type"`
	SingboxDNS       string `json:"singbox_dns" gorm:"column:singbox_dns"`
	SingboxDNSType   string `json:"singbox_dns_type" gorm:"column:singbox_dns_type"`
	EDNSClientSubnet string `json:"edns_client_subnet" gorm:"column:edns_client_subnet"`
}

// Validate 验证 DNS 设置
func (s *DNSSettings) Validate() error {
	// 验证 Final 字段
	if s.Final == "" {
		s.Final = "google" // 默认使用 google DNS
	}

	// 验证 Strategy 字段
	if s.Strategy == "" {
		s.Strategy = "ipv4_only" // 默认使用 ipv4_only
	}

	return nil
}
