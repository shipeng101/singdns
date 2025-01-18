package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"singdns/api/models"

	"github.com/sirupsen/logrus"
)

// Manager implements the ProxyManager interface
type Manager struct {
	cmd        *exec.Cmd
	configPath string
	logger     *logrus.Logger
	workDir    string
	startTime  time.Time
	version    string
}

// NewManager creates a new proxy manager
func NewManager(logger *logrus.Logger, configPath string, workDir string) *Manager {
	return &Manager{
		logger:     logger,
		configPath: configPath,
		workDir:    workDir,
	}
}

// getLocalNetwork 获取本机网络信息
func (m *Manager) getLocalNetwork() (string, string, string, error) {
	// 获取默认路由的网卡和网关
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		return "", "", "", fmt.Errorf("get default route: %v", err)
	}
	fields := strings.Fields(string(output))
	if len(fields) < 3 {
		return "", "", "", fmt.Errorf("invalid route output: %s", string(output))
	}
	gateway := fields[2] // 网关地址
	iface := fields[4]

	// 获取网卡的 IP 地址和网段
	cmd = exec.Command("ip", "addr", "show", iface)
	output, err = cmd.Output()
	if err != nil {
		return "", "", "", fmt.Errorf("get interface addr: %v", err)
	}

	// 使用正则表达式匹配 IP/掩码
	re := regexp.MustCompile(`inet\s+(\d+\.\d+\.\d+\.\d+/\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "", "", "", fmt.Errorf("no IPv4 address found for interface %s", iface)
	}

	// 解析 IP 和网段
	ip, ipNet, err := net.ParseCIDR(matches[1])
	if err != nil {
		return "", "", "", fmt.Errorf("parse CIDR: %v", err)
	}

	return ip.String(), ipNet.String(), gateway, nil
}

// findAvailableTunInterface 获取可用的 TUN 接口名称
func (m *Manager) findAvailableTunInterface() string {
	// 尝试从 tun0 到 tun9
	for i := 0; i < 10; i++ {
		ifName := fmt.Sprintf("tun%d", i)
		// 检查接口是否已存在
		_, err := net.InterfaceByName(ifName)
		if err != nil {
			// 接口不存在,可以使用
			return ifName
		}
	}
	// 如果所有接口都被占用,返回一个新的名称
	return fmt.Sprintf("tun%d", time.Now().Unix()%100)
}

// getTunInterface 从配置文件中获取当前使用的 TUN 接口名称
func (m *Manager) getTunInterface() (string, error) {
	configData, err := os.ReadFile(m.configPath)
	if err != nil {
		return "", fmt.Errorf("read config file: %v", err)
	}

	var config struct {
		Inbounds []struct {
			Type          string `json:"type"`
			InterfaceName string `json:"interface_name"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal(configData, &config); err != nil {
		return "", fmt.Errorf("parse config: %v", err)
	}

	for _, inbound := range config.Inbounds {
		if inbound.Type == "tun" {
			if inbound.InterfaceName != "" {
				return inbound.InterfaceName, nil
			}
			break
		}
	}

	return "", fmt.Errorf("tun interface not found in config")
}

// setupFirewallRules 设置防火墙规则
func (m *Manager) setupFirewallRules(mode string) error {
	// 获取本地网络信息
	localIP, localNet, gateway, err := m.getLocalNetwork()
	if err != nil {
		return fmt.Errorf("get local network: %v", err)
	}

	// 首先清除现有规则
	flushCmd := exec.Command("nft", "flush", "ruleset")
	var flushStderr bytes.Buffer
	flushCmd.Stderr = &flushStderr
	if err = flushCmd.Run(); err != nil {
		m.logger.WithFields(logrus.Fields{
			"error":  err,
			"stderr": flushStderr.String(),
		}).Warn("Failed to flush nftables rules")
	}

	var rules string
	if mode == "redirect" {
		// Redirect TCP + TProxy UDP 模式的规则
		rules = fmt.Sprintf(`
table inet sing-box {
    chain input {
        type filter hook input priority filter; policy accept;
        # 放行本地回环
        iifname "lo" accept
        
        # 放行局域网设备访问本机的流量（目标是本机IP的流量）
        ip saddr %[1]s ip daddr %[3]s accept
    }

    chain forward_filter {
        type filter hook forward priority filter; policy accept;
        # 放行本机转发的流量
        ip saddr %[1]s accept
    }

    chain prerouting_mangle {
        type filter hook prerouting priority mangle; policy accept;
        # 放行本地回环
        iifname "lo" accept
        
        # 放行保留地址
        ip daddr { 0.0.0.0/8, 127.0.0.0/8, 169.254.0.0/16, 224.0.0.0/4, 240.0.0.0/4 } return
        
        # 放行访问本机的流量
        ip daddr %[3]s return
        
        # 放行本机访问网关的流量
        ip daddr %[2]s return
        
        # UDP 流量使用 TPROXY（包括 DNS）
        meta l4proto udp counter tproxy ip to :7893 mark 0x1
    }

    chain output_mangle {
        type route hook output priority mangle; policy accept;
        # 放行本地回环
        oifname "lo" accept
        
        # 放行保留地址
        ip daddr { 0.0.0.0/8, 127.0.0.0/8, 169.254.0.0/16, 224.0.0.0/4, 240.0.0.0/4 } return
        
        # 放行访问本机的流量
        ip daddr %[3]s return
        
        # 放行本机访问网关的流量
        ip daddr %[2]s return
        
        # 标记 UDP 流量
        meta l4proto udp counter mark 0x1
    }

    chain prerouting_dnat {
        type nat hook prerouting priority dstnat; policy accept;
        # 放行访问本机的流量
        ip daddr %[3]s return
        
        # 放行本机访问网关的流量
        ip daddr %[2]s return
        
        # DNS 查询重定向到 dns-in
        tcp dport 53 meta l4proto tcp counter redirect to :5353
        udp dport 53 meta l4proto udp counter redirect to :5353
        
        # TCP 流量重定向
        meta l4proto tcp counter redirect to :7892
    }

    chain postrouting_snat {
        type nat hook postrouting priority srcnat; policy accept;
        # 对其他设备的流量进行 MASQUERADE
        counter ip saddr %[1]s ip daddr != %[2]s masquerade
    }
}`, localNet, gateway, localIP)

		// 开启 IP 转发和 TProxy 支持
		if err := m.execCommand("echo 1 > /proc/sys/net/ipv4/ip_forward"); err != nil {
			return fmt.Errorf("enable ip forward: %w", err)
		}

		if err := m.execCommand("echo 1 > /proc/sys/net/ipv4/ip_nonlocal_bind"); err != nil {
			return fmt.Errorf("enable ip nonlocal bind: %w", err)
		}

		// 设置策略路由
		if err := m.execCommand("ip rule del fwmark 0x1 table 100 2>/dev/null || true"); err != nil {
			m.logger.WithError(err).Warn("Failed to delete existing ip rule")
		}

		if err := m.execCommand("ip route del local 0.0.0.0/0 dev lo table 100 2>/dev/null || true"); err != nil {
			m.logger.WithError(err).Warn("Failed to delete existing ip route")
		}

		if err := m.execCommand("ip rule add fwmark 0x1 lookup 100"); err != nil {
			return fmt.Errorf("add ip rule: %w", err)
		}

		if err := m.execCommand("ip route add local 0.0.0.0/0 dev lo table 100"); err != nil {
			return fmt.Errorf("add ip route: %w", err)
		}
	} else if mode == "tun" {
		// TUN 模式的规则
		rules = fmt.Sprintf(`
table ip nat {
    chain postrouting {
        type nat hook postrouting priority 100; policy accept;
        counter ip saddr %[1]s ip daddr != %[2]s oifname "tun0" masquerade
    }
}`, localNet, gateway)
	}

	// 写入临时文件
	var nftfile *os.File
	nftfile, err = os.CreateTemp("", "nftables-rules-*.nft")
	if err != nil {
		return fmt.Errorf("create temp file: %v", err)
	}
	defer os.Remove(nftfile.Name())

	if _, err = nftfile.Write([]byte(rules)); err != nil {
		return fmt.Errorf("write rules file: %v", err)
	}
	if err = nftfile.Close(); err != nil {
		return fmt.Errorf("close rules file: %v", err)
	}

	// 应用规则
	cmd := exec.Command("nft", "-f", nftfile.Name())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		output := stderr.String()
		m.logger.WithFields(logrus.Fields{
			"rules":  rules,
			"output": output,
		}).Error("Failed to apply nftables rules")
		return fmt.Errorf("apply nftables rules: %v, output: %s", err, output)
	}

	m.logger.Info("防火墙规则设置成功")
	return nil
}

// execCommand 执行命令
func (m *Manager) execCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %v (%s)", command, err, stderr.String())
	}
	return nil
}

// getDefaultInterface 获取默认网卡名称
func (m *Manager) getDefaultInterface() string {
	cmd := exec.Command("ip", "route", "show", "default")
	output, err := cmd.Output()
	if err != nil {
		return "eth0" // 默认值
	}
	fields := strings.Fields(string(output))
	if len(fields) < 5 {
		return "eth0"
	}
	return fields[4]
}

// Start starts the sing-box service
func (m *Manager) Start() error {
	m.logger.WithFields(logrus.Fields{
		"workDir":    m.workDir,
		"configPath": m.configPath,
	}).Info("Starting sing-box service")

	if m.IsRunning() {
		m.logger.Warn("Service is already running")
		return fmt.Errorf("service is already running")
	}

	// Kill any existing sing-box processes
	if err := exec.Command("pkill", "-f", "sing-box").Run(); err != nil {
		m.logger.WithError(err).Debug("No existing sing-box processes found")
	}

	// Check if sing-box binary exists
	binPath := fmt.Sprintf("%s/bin/sing-box", m.workDir)
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		m.logger.WithField("path", binPath).Error("sing-box binary not found")
		return fmt.Errorf("sing-box binary not found at %s", binPath)
	}

	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		m.logger.WithField("path", m.configPath).Error("Config file not found")
		return fmt.Errorf("config file not found at %s", m.configPath)
	}

	// Check file permissions
	binInfo, err := os.Stat(binPath)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get binary file info")
		return fmt.Errorf("failed to get binary file info: %v", err)
	}
	m.logger.WithField("permissions", binInfo.Mode().String()).Debug("Binary file permissions")

	configInfo, err := os.Stat(m.configPath)
	if err != nil {
		m.logger.WithError(err).Error("Failed to get config file info")
		return fmt.Errorf("failed to get config file info: %v", err)
	}
	m.logger.WithField("permissions", configInfo.Mode().String()).Debug("Config file permissions")

	// Enable IP forwarding
	if err := os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1\n"), 0644); err != nil {
		m.logger.WithError(err).Warn("Failed to enable IP forwarding by writing to /proc/sys/net/ipv4/ip_forward")
		// 尝试使用 sudo sysctl
		enableCmd := exec.Command("sudo", "sysctl", "-w", "net.ipv4.ip_forward=1")
		if err := enableCmd.Run(); err != nil {
			m.logger.WithError(err).Error("Failed to enable IP forwarding using sysctl")
		} else {
			m.logger.Info("IP forwarding enabled using sysctl")
		}
	} else {
		m.logger.Info("IP forwarding enabled by writing to /proc/sys/net/ipv4/ip_forward")
	}

	// 验证 IP 转发是否已开启
	if data, err := os.ReadFile("/proc/sys/net/ipv4/ip_forward"); err == nil {
		if string(bytes.TrimSpace(data)) == "1" {
			m.logger.Info("Verified IP forwarding is enabled")
		} else {
			m.logger.Warn("IP forwarding appears to be disabled")
		}
	}

	// Create command
	m.cmd = exec.Command(binPath, "run", "-c", m.configPath)
	m.cmd.Dir = m.workDir

	// Capture command output
	var stdout, stderr bytes.Buffer
	m.cmd.Stdout = &stdout
	m.cmd.Stderr = &stderr

	// Start the process
	m.startTime = time.Now()
	if err := m.cmd.Start(); err != nil {
		m.logger.WithError(err).Error("Failed to start service")
		return fmt.Errorf("failed to start service: %v", err)
	}

	// Wait a bit to ensure process started successfully
	time.Sleep(time.Second)
	if !m.IsRunning() {
		output := stdout.String()
		errOutput := stderr.String()
		m.logger.WithFields(logrus.Fields{
			"stdout": output,
			"stderr": errOutput,
		}).Error("Service failed to start")
		return fmt.Errorf("service failed to start. Stdout: %s, Stderr: %s", output, errOutput)
	}

	m.logger.Info("sing-box service started successfully")

	// 等待一会儿让 sing-box 完全启动并创建接口
	time.Sleep(time.Second * 2)

	// 读取配置文件以确定模式
	configData, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("read config file: %v", err)
	}

	var config struct {
		Inbounds []struct {
			Type string `json:"type"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("parse config: %v", err)
	}

	// 确定模式
	mode := "tun"
	for _, inbound := range config.Inbounds {
		if inbound.Type == "redirect" {
			mode = "redirect"
			break
		}
	}

	// 设置防火墙规则
	if err := m.setupFirewallRules(mode); err != nil {
		m.logger.WithError(err).Error("Failed to setup firewall rules")
		// 这里我们不返回错误,因为服务已经启动成功
	}

	return nil
}

// clearFirewallRules 清除防火墙规则
func (m *Manager) clearFirewallRules() error {
	cmd := exec.Command("nft", "flush", "ruleset")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errOutput := stderr.String()
		m.logger.WithFields(logrus.Fields{
			"error":  err,
			"stderr": errOutput,
		}).Error("Failed to clear nftables rules")
		return fmt.Errorf("clear nftables rules: %v (stderr: %s)", err, errOutput)
	}
	m.logger.Info("Firewall rules cleared")
	return nil
}

// Stop stops the sing-box service
func (m *Manager) Stop() error {
	if !m.IsRunning() {
		return fmt.Errorf("service is not running")
	}

	// 清除防火墙规则
	if err := m.clearFirewallRules(); err != nil {
		m.logger.WithError(err).Warn("Failed to clear firewall rules, continuing with service stop")
	}

	// Kill the process
	if err := m.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("kill process: %v", err)
	}

	m.cmd = nil
	m.logger.Info("Service stopped")
	return nil
}

// IsRunning returns whether the service is running
func (m *Manager) IsRunning() bool {
	if m.cmd == nil || m.cmd.Process == nil {
		return false
	}

	// 使用 pgrep 检查进程
	cmd := exec.Command("pgrep", "-f", "sing-box")
	output, err := cmd.Output()
	if err != nil {
		m.cmd = nil
		return false
	}

	// 检查进程是否真的在运行
	if len(output) == 0 {
		m.cmd = nil
		return false
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		m.cmd = nil
		return false
	}

	return true
}

// GetVersion returns the version of sing-box
func (m *Manager) GetVersion() string {
	if m.version != "" {
		return m.version
	}

	cmd := exec.Command(fmt.Sprintf("%s/bin/sing-box", m.workDir), "version")
	cmd.Dir = m.workDir
	out, err := cmd.Output()
	if err != nil {
		m.logger.WithError(err).Error("Failed to get version")
		return "unknown"
	}

	m.version = string(out)
	return m.version
}

// GetUptime returns the uptime in seconds
func (m *Manager) GetUptime() int64 {
	if !m.IsRunning() {
		return 0
	}
	return int64(time.Since(m.startTime).Seconds())
}

// StartService starts the specified service
func (m *Manager) StartService(service string) error {
	switch service {
	case "sing-box":
		return m.Start()
	default:
		return fmt.Errorf("unknown service: %s", service)
	}
}

// StopService stops the specified service
func (m *Manager) StopService(service string) error {
	switch service {
	case "sing-box":
		return m.Stop()
	default:
		return fmt.Errorf("unknown service: %s", service)
	}
}

// RestartService restarts the specified service
func (m *Manager) RestartService(service string) error {
	switch service {
	case "sing-box":
		if err := m.Stop(); err != nil {
			m.logger.WithError(err).Error("Failed to stop service during restart")
		}
		return m.Start()
	default:
		return fmt.Errorf("unknown service: %s", service)
	}
}

// GetNodeHealth returns the health status of a node
func (m *Manager) GetNodeHealth(id string) *models.HealthData {
	return &models.HealthData{
		Status:    "unknown",
		Latency:   0,
		LastCheck: time.Now(),
	}
}

// UpdateNodeHealth updates the health status of a node
func (m *Manager) UpdateNodeHealth(nodeID string, status string, latency int64, checkedAt time.Time, err string) {
	// No-op since we removed health tracking
}

// GetStartTime returns the start time of the service
func (m *Manager) GetStartTime() time.Time {
	return m.startTime
}
