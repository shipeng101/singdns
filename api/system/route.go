package system

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// RouteManager 路由管理器
type RouteManager struct {
	enabled bool
	osType  string
}

// NewRouteManager 创建路由管理器
func NewRouteManager() *RouteManager {
	return &RouteManager{
		osType: detectOSType(),
	}
}

// detectOSType 检测操作系统类型
func detectOSType() string {
	// 检查是否是 Alpine
	if _, err := os.Stat("/etc/alpine-release"); err == nil {
		return "alpine"
	}

	// 检查其他 Linux 发行版
	if out, err := exec.Command("cat", "/etc/os-release").Output(); err == nil {
		output := string(out)
		if strings.Contains(output, "ID=debian") || strings.Contains(output, "ID=ubuntu") {
			return "debian"
		}
		if strings.Contains(output, "ID=centos") || strings.Contains(output, "ID=rhel") {
			return "redhat"
		}
	}

	return "unknown"
}

// ensureDependencies 确保必要的依赖已安装
func (m *RouteManager) ensureDependencies() error {
	var installCmd *exec.Cmd

	switch m.osType {
	case "alpine":
		// Alpine 使用 apk
		if err := exec.Command("which", "iptables").Run(); err != nil {
			fmt.Println("installing iptables on Alpine...")
			installCmd = exec.Command("apk", "add", "--no-cache", "iptables", "ip6tables")
		}
	case "debian":
		// Debian/Ubuntu 使用 apt
		if err := exec.Command("which", "iptables").Run(); err != nil {
			fmt.Println("installing iptables on Debian/Ubuntu...")
			installCmd = exec.Command("apt-get", "update")
			installCmd.Run()
			installCmd = exec.Command("apt-get", "install", "-y", "iptables", "iproute2")
		}
	case "redhat":
		// CentOS/RHEL 使用 yum
		if err := exec.Command("which", "iptables").Run(); err != nil {
			fmt.Println("installing iptables on CentOS/RHEL...")
			installCmd = exec.Command("yum", "install", "-y", "iptables", "iptables-services", "iproute")
		}
	default:
		return fmt.Errorf("unsupported operating system")
	}

	if installCmd != nil {
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %v", err)
		}

		// 对于特定系统，可能需要额外的设置
		if m.osType == "redhat" {
			// 启用 iptables 服务
			if err := exec.Command("systemctl", "enable", "iptables").Run(); err != nil {
				return fmt.Errorf("failed to enable iptables service: %v", err)
			}
			if err := exec.Command("systemctl", "start", "iptables").Run(); err != nil {
				return fmt.Errorf("failed to start iptables service: %v", err)
			}
		}
	}

	return nil
}

// Setup 设置路由规则
func (m *RouteManager) Setup() error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("windows system is not supported")
	}

	// 检查是否为 root
	if err := checkRoot(); err != nil {
		return err
	}

	// 确保依赖已安装
	if err := m.ensureDependencies(); err != nil {
		return fmt.Errorf("failed to ensure dependencies: %v", err)
	}

	// 启用 IP 转发
	if err := m.enableIPForward(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding: %v", err)
	}

	// 设置路由表
	if err := m.setupRoutingTable(); err != nil {
		return fmt.Errorf("failed to setup routing table: %v", err)
	}

	// 设置 iptables 规则
	if err := m.setupIPTables(); err != nil {
		return fmt.Errorf("failed to setup iptables rules: %v", err)
	}

	m.enabled = true
	return nil
}

// Cleanup 清理路由规则
func (m *RouteManager) Cleanup() error {
	if !m.enabled {
		return nil
	}

	if err := m.cleanupIPTables(); err != nil {
		return fmt.Errorf("failed to cleanup iptables rules: %v", err)
	}

	if err := m.cleanupRoutingTable(); err != nil {
		return fmt.Errorf("failed to cleanup routing table: %v", err)
	}

	m.enabled = false
	return nil
}

// enableIPForward 启用 IP 转发
func (m *RouteManager) enableIPForward() error {
	var cmd *exec.Cmd

	switch m.osType {
	case "alpine":
		// Alpine 特定的 sysctl 设置
		cmd = exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
		if err := cmd.Run(); err != nil {
			return err
		}

		// 持久化设置
		return m.persistSysctl("net.ipv4.ip_forward=1")

	default:
		// 其他 Linux 发行版的通用设置
		cmd = exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
		if err := cmd.Run(); err != nil {
			return err
		}

		// 持久化设置
		return m.persistSysctl("net.ipv4.ip_forward=1")
	}
}

// setupRoutingTable 设置路由表
func (m *RouteManager) setupRoutingTable() error {
	cmds := [][]string{
		{"ip", "route", "add", "local", "default", "dev", "lo", "table", "100"},
		{"ip", "rule", "add", "fwmark", "100", "table", "100"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		if err := cmd.Run(); err != nil {
			// 忽略 "File exists" 错误
			if !strings.Contains(err.Error(), "File exists") {
				return fmt.Errorf("failed to run command %v: %v", cmdArgs, err)
			}
		}
	}

	return nil
}

// setupIPTables 设置 iptables 规则
func (m *RouteManager) setupIPTables() error {
	rules := [][]string{
		{"-t", "mangle", "-N", "SING_BOX"},
		{"-t", "mangle", "-A", "SING_BOX", "-d", "127.0.0.0/8", "-j", "RETURN"},
		{"-t", "mangle", "-A", "SING_BOX", "-d", "224.0.0.0/4", "-j", "RETURN"},
		{"-t", "mangle", "-A", "SING_BOX", "-d", "255.255.255.255/32", "-j", "RETURN"},
		{"-t", "mangle", "-A", "SING_BOX", "-d", "192.168.0.0/16", "-j", "RETURN"},
		{"-t", "mangle", "-A", "SING_BOX", "-d", "10.0.0.0/8", "-j", "RETURN"},
		{"-t", "mangle", "-A", "SING_BOX", "-d", "172.16.0.0/12", "-j", "RETURN"},
		{"-t", "mangle", "-A", "SING_BOX", "-j", "MARK", "--set-mark", "100"},
		{"-t", "mangle", "-A", "PREROUTING", "-j", "SING_BOX"},
		{"-t", "mangle", "-N", "SING_BOX_SELF"},
		{"-t", "mangle", "-A", "SING_BOX_SELF", "-j", "MARK", "--set-mark", "100"},
		{"-t", "mangle", "-A", "OUTPUT", "-j", "SING_BOX_SELF"},
	}

	for _, rule := range rules {
		cmd := exec.Command("iptables", rule...)
		if err := cmd.Run(); err != nil {
			// 忽略链已存在的错误
			if !strings.Contains(err.Error(), "Chain already exists") {
				return fmt.Errorf("failed to setup iptables rule %v: %v", rule, err)
			}
		}
	}

	return nil
}

// cleanupIPTables 清理 iptables 规则
func (m *RouteManager) cleanupIPTables() error {
	rules := [][]string{
		{"-t", "mangle", "-D", "PREROUTING", "-j", "SING_BOX"},
		{"-t", "mangle", "-D", "OUTPUT", "-j", "SING_BOX_SELF"},
		{"-t", "mangle", "-F", "SING_BOX"},
		{"-t", "mangle", "-X", "SING_BOX"},
		{"-t", "mangle", "-F", "SING_BOX_SELF"},
		{"-t", "mangle", "-X", "SING_BOX_SELF"},
	}

	for _, rule := range rules {
		cmd := exec.Command("iptables", rule...)
		_ = cmd.Run() // 忽略错误，因为规则可能不存在
	}

	return nil
}

// cleanupRoutingTable 清理路由表
func (m *RouteManager) cleanupRoutingTable() error {
	cmds := [][]string{
		{"ip", "rule", "del", "fwmark", "100", "table", "100"},
		{"ip", "route", "del", "local", "default", "dev", "lo", "table", "100"},
	}

	for _, cmdArgs := range cmds {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		_ = cmd.Run() // 忽略错误，因为规则可能不存在
	}

	return nil
}

// persistSysctl 持久化 sysctl 设置
func (m *RouteManager) persistSysctl(setting string) error {
	var sysctlPath string

	switch m.osType {
	case "alpine":
		sysctlPath = "/etc/sysctl.d/99-singdns.conf"
	default:
		sysctlPath = "/etc/sysctl.d/99-singdns.conf"
	}

	// 检查文件是否存在
	if _, err := os.Stat(sysctlPath); os.IsNotExist(err) {
		// 创建文件
		f, err := os.Create(sysctlPath)
		if err != nil {
			return err
		}
		f.Close()
	}

	// 追加设置
	return exec.Command("sh", "-c", fmt.Sprintf("echo '%s' >> %s", setting, sysctlPath)).Run()
}

// checkRoot 检查是否为 root 用户
func checkRoot() error {
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check root: %v", err)
	}

	if strings.TrimSpace(string(output)) != "0" {
		return fmt.Errorf("this function requires root privileges")
	}

	return nil
}
