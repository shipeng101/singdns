package network

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Manager 网络管理器
type Manager struct {
	os string
}

// NewManager 创建网络管理器
func NewManager() *Manager {
	return &Manager{
		os: runtime.GOOS,
	}
}

// SetupDNSRedirect 设置 DNS 重定向
func (m *Manager) SetupDNSRedirect() error {
	switch m.os {
	case "linux":
		return m.setupLinuxDNSRedirect()
	case "darwin":
		return m.setupDarwinDNSRedirect()
	default:
		return fmt.Errorf("不支持的操作系统: %s", m.os)
	}
}

// CleanupDNSRedirect 清理 DNS 重定向
func (m *Manager) CleanupDNSRedirect() error {
	switch m.os {
	case "linux":
		return m.cleanupLinuxDNSRedirect()
	case "darwin":
		return m.cleanupDarwinDNSRedirect()
	default:
		return nil
	}
}

// setupLinuxDNSRedirect 设置 Linux DNS 重定向
func (m *Manager) setupLinuxDNSRedirect() error {
	// 检查是否有 iptables
	if _, err := exec.LookPath("iptables"); err != nil {
		return fmt.Errorf("未找到 iptables: %v", err)
	}

	// 添加 NAT 规则
	cmds := [][]string{
		// 清除已有的 DNS 重定向规则
		{"iptables", "-t", "nat", "-D", "OUTPUT", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-port", "5354"},
		{"iptables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-port", "5354"},
		{"iptables", "-t", "nat", "-D", "OUTPUT", "-p", "tcp", "--dport", "53", "-j", "REDIRECT", "--to-port", "5354"},
		{"iptables", "-t", "nat", "-D", "PREROUTING", "-p", "tcp", "--dport", "53", "-j", "REDIRECT", "--to-port", "5354"},

		// 添加新的 DNS 重定向规则
		{"iptables", "-t", "nat", "-A", "OUTPUT", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-port", "5354"},
		{"iptables", "-t", "nat", "-A", "PREROUTING", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-port", "5354"},
		{"iptables", "-t", "nat", "-A", "OUTPUT", "-p", "tcp", "--dport", "53", "-j", "REDIRECT", "--to-port", "5354"},
		{"iptables", "-t", "nat", "-A", "PREROUTING", "-p", "tcp", "--dport", "53", "-j", "REDIRECT", "--to-port", "5354"},
	}

	for _, cmd := range cmds {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			// 忽略删除不存在规则的错误
			if len(cmd) > 2 && cmd[2] == "-D" {
				continue
			}
			return fmt.Errorf("执行命令失败 %v: %v", cmd, err)
		}
	}

	return nil
}

// cleanupLinuxDNSRedirect 清理 Linux DNS 重定向
func (m *Manager) cleanupLinuxDNSRedirect() error {
	// 检查是否有 iptables
	if _, err := exec.LookPath("iptables"); err != nil {
		return fmt.Errorf("未找到 iptables: %v", err)
	}

	// 删除 NAT 规则
	cmds := [][]string{
		{"iptables", "-t", "nat", "-D", "OUTPUT", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-port", "5354"},
		{"iptables", "-t", "nat", "-D", "PREROUTING", "-p", "udp", "--dport", "53", "-j", "REDIRECT", "--to-port", "5354"},
		{"iptables", "-t", "nat", "-D", "OUTPUT", "-p", "tcp", "--dport", "53", "-j", "REDIRECT", "--to-port", "5354"},
		{"iptables", "-t", "nat", "-D", "PREROUTING", "-p", "tcp", "--dport", "53", "-j", "REDIRECT", "--to-port", "5354"},
	}

	for _, cmd := range cmds {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			// 忽略删除不存在规则的错误
			continue
		}
	}

	return nil
}

// setupDarwinDNSRedirect 设置 macOS DNS 重定向
func (m *Manager) setupDarwinDNSRedirect() error {
	// 检查是否有 pfctl
	if _, err := exec.LookPath("pfctl"); err != nil {
		return fmt.Errorf("未找到 pfctl: %v", err)
	}

	// 创建 PF 规则
	rules := `
rdr pass on lo0 inet proto udp from any to any port 53 -> 127.0.0.1 port 5354
rdr pass on lo0 inet proto tcp from any to any port 53 -> 127.0.0.1 port 5354
`

	// 写入临时文件
	if err := exec.Command("sh", "-c", fmt.Sprintf("echo '%s' | sudo pfctl -ef -", rules)).Run(); err != nil {
		return fmt.Errorf("设置 PF 规则失败: %v", err)
	}

	return nil
}

// cleanupDarwinDNSRedirect 清理 macOS DNS 重定向
func (m *Manager) cleanupDarwinDNSRedirect() error {
	// 检查是否有 pfctl
	if _, err := exec.LookPath("pfctl"); err != nil {
		return fmt.Errorf("未找到 pfctl: %v", err)
	}

	// 禁用 PF
	if err := exec.Command("sudo", "pfctl", "-d").Run(); err != nil {
		return fmt.Errorf("禁用 PF 失败: %v", err)
	}

	return nil
}
