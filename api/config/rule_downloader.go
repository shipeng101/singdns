package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"singdns/api/models"
	"singdns/api/storage"
)

// RuleSet 规则集信息
type RuleSet struct {
	Tag       string
	RemoteURL string
	LocalPath string
}

var ruleSets = []RuleSet{
	{
		Tag:       "geosite-category-ads",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/category-ads-all.srs",
		LocalPath: "configs/sing-box/rules/geosite-category-ads.srs",
	},
	{
		Tag:       "geosite-google",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/google.srs",
		LocalPath: "configs/sing-box/rules/geosite-google.srs",
	},
	{
		Tag:       "geoip-google",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geoip/google.srs",
		LocalPath: "configs/sing-box/rules/geoip-google.srs",
	},
	{
		Tag:       "geosite-telegram",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/telegram.srs",
		LocalPath: "configs/sing-box/rules/geosite-telegram.srs",
	},
	{
		Tag:       "geoip-telegram",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geoip/telegram.srs",
		LocalPath: "configs/sing-box/rules/geoip-telegram.srs",
	},
	{
		Tag:       "geosite-youtube",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/youtube.srs",
		LocalPath: "configs/sing-box/rules/geosite-youtube.srs",
	},
	{
		Tag:       "geosite-spotify",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/spotify.srs",
		LocalPath: "configs/sing-box/rules/geosite-spotify.srs",
	},
	{
		Tag:       "geosite-category-games",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/category-games.srs",
		LocalPath: "configs/sing-box/rules/geosite-category-games.srs",
	},
	{
		Tag:       "geosite-netflix",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/netflix.srs",
		LocalPath: "configs/sing-box/rules/geosite-netflix.srs",
	},
	{
		Tag:       "geoip-netflix",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geoip/netflix.srs",
		LocalPath: "configs/sing-box/rules/geoip-netflix.srs",
	},
	{
		Tag:       "geosite-disney",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/disney.srs",
		LocalPath: "configs/sing-box/rules/geosite-disney.srs",
	},
	{
		Tag:       "geosite-apple",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/apple.srs",
		LocalPath: "configs/sing-box/rules/geosite-apple.srs",
	},
	{
		Tag:       "geosite-microsoft",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/microsoft.srs",
		LocalPath: "configs/sing-box/rules/geosite-microsoft.srs",
	},
	{
		Tag:       "geosite-openai",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/openai.srs",
		LocalPath: "configs/sing-box/rules/geosite-openai.srs",
	},
	{
		Tag:       "geosite-twitter",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/twitter.srs",
		LocalPath: "configs/sing-box/rules/geosite-twitter.srs",
	},
	{
		Tag:       "geosite-facebook",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/facebook.srs",
		LocalPath: "configs/sing-box/rules/geosite-facebook.srs",
	},
	{
		Tag:       "geosite-instagram",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/instagram.srs",
		LocalPath: "configs/sing-box/rules/geosite-instagram.srs",
	},
	{
		Tag:       "geosite-amazon",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/amazon.srs",
		LocalPath: "configs/sing-box/rules/geosite-amazon.srs",
	},
	{
		Tag:       "geosite-github",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/github.srs",
		LocalPath: "configs/sing-box/rules/geosite-github.srs",
	},
	{
		Tag:       "geosite-category-porn",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/category-porn.srs",
		LocalPath: "configs/sing-box/rules/geosite-category-porn.srs",
	},
	{
		Tag:       "geosite-cn",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/cn.srs",
		LocalPath: "configs/sing-box/rules/geosite-cn.srs",
	},
	{
		Tag:       "geoip-cn",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geoip/cn.srs",
		LocalPath: "configs/sing-box/rules/geoip-cn.srs",
	},
	{
		Tag:       "geosite-geolocation-!cn",
		RemoteURL: "https://ghproxy.cn/https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/sing/geo/geosite/geolocation-!cn.srs",
		LocalPath: "configs/sing-box/rules/geosite-geolocation-!cn.srs",
	},
}

// DownloadRuleSets 下载所有规则集
func DownloadRuleSets() error {
	// 创建规则集目录
	rulesDir := "configs/sing-box/rules"
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return fmt.Errorf("failed to create rules directory: %v", err)
	}

	// 下载每个规则集
	for _, ruleSet := range ruleSets {
		// 检查文件是否已存在
		if _, err := os.Stat(ruleSet.LocalPath); err == nil {
			fmt.Printf("Rule set %s already exists, skipping download\n", ruleSet.Tag)
			continue
		}

		fmt.Printf("Downloading rule set %s...\n", ruleSet.Tag)

		// 创建目标文件
		out, err := os.Create(ruleSet.LocalPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %v", ruleSet.LocalPath, err)
		}
		defer out.Close()

		// 下载文件
		resp, err := http.Get(ruleSet.RemoteURL)
		if err != nil {
			return fmt.Errorf("failed to download %s: %v", ruleSet.Tag, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to download %s: status code %d", ruleSet.Tag, resp.StatusCode)
		}

		// 写入文件
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to save %s: %v", ruleSet.Tag, err)
		}

		fmt.Printf("Successfully downloaded rule set %s\n", ruleSet.Tag)
	}

	return nil
}

// EnsureRuleSetsExist 确保所有规则集文件存在
func EnsureRuleSetsExist() error {
	for _, ruleSet := range ruleSets {
		if _, err := os.Stat(ruleSet.LocalPath); os.IsNotExist(err) {
			return fmt.Errorf("rule set file %s does not exist", ruleSet.LocalPath)
		}
	}
	return nil
}

// InitializeRuleSets 初始化规则集到数据库
func InitializeRuleSets(storage storage.Storage) error {
	// 创建规则集目录
	rulesDir := "configs/sing-box/rules"
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return fmt.Errorf("failed to create rules directory: %v", err)
	}

	// 获取规则名称的中文名称
	getRuleName := func(name string) string {
		switch {
		case strings.Contains(name, "category-games"):
			return "游戏"
		case strings.Contains(name, "google"):
			return "Google"
		case strings.Contains(name, "telegram"):
			return "Telegram"
		case strings.Contains(name, "youtube"):
			return "YouTube"
		case strings.Contains(name, "netflix"):
			return "Netflix"
		case strings.Contains(name, "disney"):
			return "Disney+"
		case strings.Contains(name, "spotify"):
			return "Spotify"
		case strings.Contains(name, "apple"):
			return "Apple"
		case strings.Contains(name, "microsoft"):
			return "Microsoft"
		case strings.Contains(name, "openai"):
			return "OpenAI"
		case strings.Contains(name, "twitter"):
			return "Twitter"
		case strings.Contains(name, "facebook"):
			return "Facebook"
		case strings.Contains(name, "instagram"):
			return "Instagram"
		case strings.Contains(name, "amazon"):
			return "Amazon"
		case strings.Contains(name, "github"):
			return "GitHub"
		case strings.Contains(name, "category-porn"):
			return "成人网站"
		case strings.Contains(name, "category-ads"):
			return "广告"
		case strings.Contains(name, "cn"):
			return "中国大陆"
		case strings.Contains(name, "geolocation-!cn"):
			return "非中国大陆"
		default:
			return name
		}
	}

	// 获取规则类型的中文名称
	getRuleTypeName := func(ruleType string) string {
		switch ruleType {
		case "geosite":
			return "域名"
		case "geoip":
			return "IP"
		default:
			return ruleType
		}
	}

	// 初始化每个规则集
	for _, ruleSet := range ruleSets {
		ruleType := strings.Split(ruleSet.Tag, "-")[0] // geosite 或 geoip
		ruleName := getRuleName(ruleSet.Tag)
		ruleTypeName := getRuleTypeName(ruleType)

		// 设置默认出口
		var defaultOutbound string
		switch {
		case strings.Contains(ruleSet.Tag, "cn"):
			defaultOutbound = "direct"
		case strings.Contains(ruleSet.Tag, "category-ads"):
			defaultOutbound = "block"
		default:
			defaultOutbound = "节点选择"
		}

		// 构造规则集对象
		dbRuleSet := &models.RuleSet{
			ID:        ruleSet.Tag,
			Name:      fmt.Sprintf("%s-%s", ruleName, ruleTypeName),
			URL:       ruleSet.RemoteURL,
			Type:      ruleType,
			Format:    "binary",
			Enabled:   true,
			Outbound:  defaultOutbound,
			UpdatedAt: time.Now(),
		}

		// 保存到数据库
		if err := storage.SaveRuleSet(dbRuleSet); err != nil {
			return fmt.Errorf("failed to save rule set %s: %v", ruleSet.Tag, err)
		}
	}

	return nil
}
