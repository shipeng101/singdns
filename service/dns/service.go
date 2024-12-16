package dns

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/shipeng101/singdns/pkg/types"
	"github.com/shipeng101/singdns/service/system"
	"go.uber.org/zap"
)

// Service DNS服务接口
type Service interface {
	// 启动服务
	Start() error
	// 停止服务
	Stop() error

	// 配置管理
	GetConfig() types.DNSConfig
	UpdateConfig(config types.DNSConfig) error

	// 规则管理
	GetRules() []types.Rule
	UpdateRules(rules []types.Rule) error

	// 状态查询
	GetStatus() types.ServiceStatus
}

// DNSCache DNS缓存条目
type DNSCacheEntry struct {
	Msg      *dns.Msg
	ExpireAt time.Time
}

type service struct {
	config     types.DNSConfig
	rules      []types.Rule
	ruleSets   []types.RuleSet
	status     types.ServiceStatus
	startTime  time.Time
	queryCount int64
	server     *dns.Server
	cache      map[string]*DNSCacheEntry
	cacheMu    sync.RWMutex
	mu         sync.Mutex
}

// NewService 创建DNS服务实例
func NewService(config types.DNSConfig) Service {
	// 添加默认的广告拦截规则集
	defaultRuleSet := types.RuleSet{
		ID:          "adblock",
		Name:        "AdBlock",
		Description: "默认广告拦截规则",
		Type:        "domain",
		Rules: []types.Rule{
			{
				Type:    "domain_suffix",
				Value:   "doubleclick.net",
				Target:  "reject",
				Enabled: true,
			},
			{
				Type:    "domain_suffix",
				Value:   "googleadservices.com",
				Target:  "reject",
				Enabled: true,
			},
			{
				Type:    "domain_suffix",
				Value:   "googlesyndication.com",
				Target:  "reject",
				Enabled: true,
			},
			{
				Type:    "domain_suffix",
				Value:   "google-analytics.com",
				Target:  "reject",
				Enabled: true,
			},
		},
		Enabled:        true,
		Priority:       100,
		AutoUpdate:     false,
		UpdateInterval: 24,
		Tags:           []string{"ads", "tracker"},
	}

	// 如果配置中没有规则集，添加默认规则集
	if len(config.RuleSets) == 0 {
		config.RuleSets = []types.RuleSet{defaultRuleSet}
	}

	return &service{
		config:    config,
		rules:     []types.Rule{}, // 不再使用独立规则，统一使用规则集
		ruleSets:  config.RuleSets,
		startTime: time.Now(),
		cache:     make(map[string]*DNSCacheEntry),
	}
}

// Start 启动DNS服务
func (s *service) Start() error {
	s.status = types.ServiceStatus{
		Running:    true,
		StartTime:  s.startTime,
		QueryCount: 0,
	}

	// 启动缓存清理
	if s.config.Cache {
		go s.cleanCache()
	}

	// 创建DNS服务器
	handler := &dnsHandler{service: s}
	s.server = &dns.Server{
		Addr:    s.config.Listen,
		Net:     "udp",
		Handler: handler,
	}

	// 启动DNS服务器
	go func() {
		system.Info("正在启动DNS服务", zap.String("addr", s.config.Listen))
		if err := s.server.ListenAndServe(); err != nil {
			system.Error("DNS服务启动失败", zap.Error(err))
		}
	}()

	system.Info("DNS服务已启动", zap.String("addr", s.config.Listen))
	return nil
}

// cleanCache 定期清理过期缓存
func (s *service) cleanCache() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.cacheMu.Lock()
		now := time.Now()
		for key, entry := range s.cache {
			if now.After(entry.ExpireAt) {
				delete(s.cache, key)
			}
		}
		s.cacheMu.Unlock()
	}
}

// getCacheKey 生成缓存键
func (s *service) getCacheKey(q dns.Question) string {
	return fmt.Sprintf("%s-%d", strings.ToLower(q.Name), q.Qtype)
}

// getFromCache 从缓存获取响应
func (s *service) getFromCache(q dns.Question) *dns.Msg {
	if !s.config.Cache {
		return nil
	}

	key := s.getCacheKey(q)
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	if entry, ok := s.cache[key]; ok {
		if time.Now().Before(entry.ExpireAt) {
			// 克隆消息以避免并发修改
			return entry.Msg.Copy()
		}
		// 删除过期条目
		delete(s.cache, key)
	}
	return nil
}

// addToCache 添加响应到缓存
func (s *service) addToCache(q dns.Question, msg *dns.Msg) {
	if !s.config.Cache {
		return
	}

	// 计算TTL
	minTTL := uint32(3600) // 默认1小时
	for _, answer := range msg.Answer {
		if answer.Header().Ttl < minTTL {
			minTTL = answer.Header().Ttl
		}
	}

	// 如果TTL太小，不缓存
	if minTTL < 10 {
		return
	}

	key := s.getCacheKey(q)
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.cache[key] = &DNSCacheEntry{
		Msg:      msg.Copy(),
		ExpireAt: time.Now().Add(time.Duration(minTTL) * time.Second),
	}
}

// Stop 停止DNS服务
func (s *service) Stop() error {
	if s.server != nil {
		if err := s.server.Shutdown(); err != nil {
			return fmt.Errorf("关闭DNS服务失败: %v", err)
		}
	}
	s.status.Running = false
	return nil
}

// GetStatus 获取服务状态
func (s *service) GetStatus() types.ServiceStatus {
	s.status.Uptime = time.Since(s.startTime).String()
	s.status.QueryCount = s.queryCount
	return s.status
}

// UpdateConfig 更新配置
func (s *service) UpdateConfig(config types.DNSConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 更新配置
	s.config = config
	s.ruleSets = config.RuleSets

	// 重启服务
	if err := s.Stop(); err != nil {
		return fmt.Errorf("停止DNS服务失败: %v", err)
	}

	if err := s.Start(); err != nil {
		return fmt.Errorf("重启DNS服务失败: %v", err)
	}

	return nil
}

// GetRules 获取规则
func (s *service) GetRules() []types.Rule {
	var allRules []types.Rule
	for _, ruleSet := range s.ruleSets {
		if ruleSet.Enabled {
			allRules = append(allRules, ruleSet.Rules...)
		}
	}
	return allRules
}

// UpdateRules 更新规则
func (s *service) UpdateRules(rules []types.Rule) error {
	// 将规则添加到默认规则集
	defaultRuleSet := types.RuleSet{
		ID:             "default",
		Name:           "Default",
		Description:    "默认规则集",
		Type:           "mixed",
		Rules:          rules,
		Enabled:        true,
		Priority:       999,
		AutoUpdate:     false,
		UpdateInterval: 0,
		Tags:           []string{"default"},
	}

	// 更新或添加默认规则集
	found := false
	for i, rs := range s.ruleSets {
		if rs.ID == "default" {
			s.ruleSets[i] = defaultRuleSet
			found = true
			break
		}
	}
	if !found {
		s.ruleSets = append(s.ruleSets, defaultRuleSet)
	}

	return nil
}

// GetConfig 获取DNS配置
func (s *service) GetConfig() types.DNSConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.config
}

// dnsHandler DNS请求处理器
type dnsHandler struct {
	service *service
}

// checkRules 检查域名是否匹配任何规则
func (h *dnsHandler) checkRules(name string) (string, bool) {
	// 按优先级排序规则集
	ruleSets := make([]types.RuleSet, len(h.service.ruleSets))
	copy(ruleSets, h.service.ruleSets)
	sort.Slice(ruleSets, func(i, j int) bool {
		return ruleSets[i].Priority < ruleSets[j].Priority
	})

	// 遍历所有启用的规则集
	for _, ruleSet := range ruleSets {
		if !ruleSet.Enabled {
			continue
		}

		system.Info("正在检查规则集",
			zap.String("name", ruleSet.Name),
			zap.Int("priority", ruleSet.Priority))

		// 遍历规则集中的规则
		for _, rule := range ruleSet.Rules {
			if rule.Enabled && h.matchDomain(name, rule.Value) {
				system.Info("命中规则",
					zap.String("domain", name),
					zap.String("rule_set", ruleSet.Name),
					zap.String("rule", rule.Value),
					zap.String("target", rule.Target))
				return rule.Target, true
			}
		}
	}

	return "", false
}

// ServeDNS 处理DNS请求
func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	h.service.queryCount++
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true
	msg.RecursionAvailable = true

	for _, q := range r.Question {
		name := strings.TrimSuffix(q.Name, ".")
		system.Info("收到DNS查询",
			zap.String("domain", name),
			zap.String("type", dns.TypeToString[q.Qtype]))

		// 检查缓存
		if cached := h.service.getFromCache(q); cached != nil {
			cached.SetReply(r)
			w.WriteMsg(cached)
			system.Info("使用缓存响应",
				zap.String("domain", name))
			return
		}

		// 检查规则
		if target, matched := h.checkRules(name); matched {
			switch target {
			case "reject":
				msg.Rcode = dns.RcodeNameError
				w.WriteMsg(msg)
				system.Info("域名已拦截",
					zap.String("domain", name))
				return
			case "direct":
				// 使用中国DNS服务器
				if len(h.service.config.ChinaDNS) > 0 {
					resp := h.forwardToUpstream(w, r, h.service.config.ChinaDNS[0])
					if resp != nil {
						// 添加到缓存
						h.service.addToCache(q, resp)
					}
					return
				}
			}
		}

		// 默认转发到上游DNS
		if len(h.service.config.Upstream) > 0 {
			system.Info("使用默认上游DNS服务器",
				zap.String("domain", name),
				zap.String("upstream", h.service.config.Upstream[0]))
			resp := h.forwardToUpstream(w, r, h.service.config.Upstream[0])
			if resp != nil {
				// 添加到缓存
				h.service.addToCache(q, resp)
			}
		} else {
			system.Error("未配置上游DNS服务器")
			msg.Rcode = dns.RcodeServerFailure
			w.WriteMsg(msg)
		}
	}
}

// matchDomain 检查域名是否匹配规则
func (h *dnsHandler) matchDomain(name, pattern string) bool {
	system.Info("正在匹配域名规则",
		zap.String("domain", name),
		zap.String("pattern", pattern))

	// 移除末尾的点号
	name = strings.TrimSuffix(name, ".")

	switch {
	case pattern == "*":
		return true
	case strings.HasPrefix(pattern, "*."):
		suffix := pattern[2:]
		match := strings.HasSuffix(name, suffix)
		system.Info("通配符匹配结果",
			zap.String("domain", name),
			zap.String("suffix", suffix),
			zap.Bool("match", match))
		return match
	case strings.HasSuffix(pattern, ".*"):
		prefix := pattern[:len(pattern)-2]
		match := strings.HasPrefix(name, prefix)
		system.Info("前缀匹配结果",
			zap.String("domain", name),
			zap.String("prefix", prefix),
			zap.Bool("match", match))
		return match
	default:
		// 直接后缀匹配
		match := strings.HasSuffix(name, pattern)
		system.Info("后缀匹配结果",
			zap.String("domain", name),
			zap.String("pattern", pattern),
			zap.Bool("match", match))
		return match
	}
}

// forwardToUpstream 转发请求到上游DNS服务器
func (h *dnsHandler) forwardToUpstream(w dns.ResponseWriter, r *dns.Msg, upstream string) *dns.Msg {
	// 移除协议前缀
	upstream = strings.TrimPrefix(upstream, "tcp://")
	upstream = strings.TrimPrefix(upstream, "udp://")

	system.Info("正在转发DNS查询",
		zap.String("domain", r.Question[0].Name),
		zap.String("upstream", upstream))

	c := new(dns.Client)
	c.Net = "tcp"               // 强制使用TCP，避免UDP包大小限制
	c.Timeout = 5 * time.Second // 设置超时时间

	// 设置递归查询标志
	r.RecursionDesired = true

	resp, rtt, err := c.Exchange(r, upstream)
	if err != nil {
		system.Error("上游DNS查询失败",
			zap.String("upstream", upstream),
			zap.Error(err))

		var alternateUpstreams []string
		if strings.Contains(upstream, "223.5.5.5") || strings.Contains(upstream, "119.29.29.29") {
			alternateUpstreams = h.service.config.ChinaDNS[1:]
		} else {
			alternateUpstreams = h.service.config.Upstream[1:]
		}

		// 尝试其他上游服务器
		for _, u := range alternateUpstreams {
			u = strings.TrimPrefix(u, "tcp://")
			u = strings.TrimPrefix(u, "udp://")
			system.Info("尝试备用DNS服务器",
				zap.String("domain", r.Question[0].Name),
				zap.String("upstream", u))

			resp, rtt, err = c.Exchange(r, u)
			if err == nil {
				system.Info("备用DNS查询成功",
					zap.String("upstream", u),
					zap.Duration("rtt", rtt))
				break
			}
			system.Error("备用DNS查询失败",
				zap.String("upstream", u),
				zap.Error(err))
		}
	} else {
		system.Info("DNS查询成功",
			zap.String("upstream", upstream),
			zap.Duration("rtt", rtt))
	}

	if err != nil {
		// 所有上游都失败
		system.Error("所有上游DNS查询均失败",
			zap.String("domain", r.Question[0].Name))
		m := new(dns.Msg)
		m.SetReply(r)
		m.RecursionAvailable = true
		m.Rcode = dns.RcodeServerFailure
		w.WriteMsg(m)
		return nil
	}

	resp.RecursionAvailable = true
	w.WriteMsg(resp)
	return resp
}

// handleDNSRequest 处理DNS请求
func (s *service) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	// 创建响应
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	// 处理每个问题
	for _, q := range r.Question {
		switch q.Qtype {
		case dns.TypeA:
			// 处理A记录查询
			if ip := s.resolveA(q.Name); ip != nil {
				rr := &dns.A{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					A: ip,
				}
				m.Answer = append(m.Answer, rr)
			}
		}
	}

	// 发送响应
	if err := w.WriteMsg(m); err != nil {
		system.Error("发送DNS响应失败",
			zap.Error(err))
	}
}

// shouldProxy 检查是否需要代理
func (s *service) shouldProxy(name string) bool {
	// 检查规则
	if target, matched := (&dnsHandler{service: s}).checkRules(name); matched {
		return target == "proxy"
	}
	return false
}

// resolveA 解析A记录
func (s *service) resolveA(name string) net.IP {
	// 移除末尾的点
	name = strings.TrimSuffix(name, ".")

	// 检查是否需要代理
	if s.shouldProxy(name) {
		return net.ParseIP("127.0.0.1")
	}

	return nil
}
