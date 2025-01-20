#!/bin/sh

# 设置颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置变量
INSTALL_DIR="/usr/local/singdns"
LOG_DIR="/var/log/singdns"
TEMP_DIR="/tmp/singdns_temp"
MIN_DISK_SPACE=1024  # 需要的最小磁盘空间(MB)

# 检查是否为root用户
check_root() {
    if [ "$(id -u)" != "0" ]; then
        echo "${RED}错误：请使用root用户运行此脚本${NC}"
        return 1
    fi
    return 0
}

# 检测系统类型
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
    elif [ -f /etc/debian_version ]; then
        OS="debian"
    elif [ -f /etc/redhat-release ]; then
        OS="rhel"
    else
        OS="unknown"
    fi
    echo $OS
}

# 安装系统依赖
install_system_dependencies() {
    echo "${BLUE}安装系统依赖...${NC}"
    
    OS=$(detect_os)
    case $OS in
        "alpine")
            apk update
            apk add --no-cache curl wget git sqlite busybox iptables
            ;;
        "ubuntu"|"debian")
            apt-get update
            apt-get install -y curl wget git sqlite3 busybox iptables
            ;;
        "centos"|"rhel"|"fedora")
            yum -y update
            yum -y install curl wget git sqlite busybox iptables
            ;;
        *)
            echo "${RED}不支持的操作系统${NC}"
            return 1
            ;;
    esac
    
    # 检查 busybox 是否安装成功
    if ! command -v busybox > /dev/null; then
        echo "${RED}错误：busybox 安装失败${NC}"
        return 1
    fi
    
    # 检查 iptables 是否安装成功
    if ! command -v iptables > /dev/null; then
        echo "${RED}错误：iptables 安装失败${NC}"
        return 1
    fi
    
    echo "${GREEN}系统依赖安装完成${NC}"
    return 0
}

# 检查系统要求
check_system_requirements() {
    echo "${BLUE}检查系统要求...${NC}"
    
    # 检查磁盘空间
    AVAILABLE_SPACE=$(df -m "$INSTALL_DIR" | awk 'NR==2 {print $4}')
    if [ "$AVAILABLE_SPACE" -lt "$MIN_DISK_SPACE" ]; then
        echo "${RED}错误：磁盘空间不足，需要至少 ${MIN_DISK_SPACE}MB 可用空间${NC}"
        return 1
    fi
    
    # 检查必要命令
    for cmd in curl wget git busybox iptables; do
        if ! command -v $cmd >/dev/null 2>&1; then
            echo "${YELLOW}警告：未找到命令 '$cmd'，将尝试安装${NC}"
        fi
    done
    
    # 检查系统服务
    if [ -f "/proc/sys/net/ipv4/ip_forward" ]; then
        # 检查 IP 转发是否启用
        if [ "$(cat /proc/sys/net/ipv4/ip_forward)" != "1" ]; then
            echo "${YELLOW}警告：IP 转发未启用，将尝试启用${NC}"
            echo "1" > /proc/sys/net/ipv4/ip_forward 2>/dev/null || {
                echo "${RED}错误：无法启用 IP 转发${NC}"
                return 1
            }
        fi
    else
        echo "${RED}错误：系统不支持 IP 转发${NC}"
        return 1
    fi
    
    return 0
}

# 安装SingDNS
install_singdns() {
    echo "${BLUE}开始安装 SingDNS...${NC}"
    
    # 检查系统要求
    check_system_requirements || return 1
    
    # 创建必要的目录
    mkdir -p "$INSTALL_DIR/bin/web"
    mkdir -p "$INSTALL_DIR/web"
    mkdir -p "$INSTALL_DIR/configs/sing-box/rules"
    mkdir -p "$LOG_DIR"
    
    # 复制文件
    echo "${YELLOW}复制文件...${NC}"
    if [ ! -f "singdns" ]; then
        echo "${RED}错误：未找到 singdns 主程序${NC}"
        return 1
    fi
    
    # 复制主程序
    cp singdns "$INSTALL_DIR/"
    
    # 检查并复制 sing-box
    if [ ! -f "bin/sing-box" ]; then
        echo "${RED}错误：未找到 sing-box 程序${NC}"
        return 1
    fi
    cp bin/sing-box "$INSTALL_DIR/bin/"
    
    # 复制其他文件
    if [ -d "bin/web" ]; then
        cp -r bin/web/* "$INSTALL_DIR/bin/web/"
    else
        echo "${YELLOW}警告：未找到 ClashAPI UI 面板文件${NC}"
    fi
    
    if [ -d "web" ]; then
        cp -r web/* "$INSTALL_DIR/web/"
    else
        echo "${YELLOW}警告：未找到前端文件${NC}"
    fi
    
    # 复制配置文件
    if [ -f "configs/sing-box/config.json" ]; then
        cp configs/sing-box/config.json "$INSTALL_DIR/configs/sing-box/"
    else
        echo "${RED}错误：未找到配置文件${NC}"
        return 1
    fi
    
    # 复制规则文件
    cp configs/sing-box/rules/*.srs "$INSTALL_DIR/configs/sing-box/rules/" 2>/dev/null || echo "${YELLOW}警告：未找到规则文件${NC}"
    
    # 复制其他文件
    for file in install.sh singdns.sh README.md LICENSE VERSION; do
        [ -f "$file" ] && cp "$file" "$INSTALL_DIR/"
    done
    
    # 设置权限
    echo "${YELLOW}设置文件权限...${NC}"
    find "$INSTALL_DIR" -type f -exec chmod 644 {} \;
    find "$INSTALL_DIR" -type d -exec chmod 755 {} \;
    chmod +x "$INSTALL_DIR/singdns"
    chmod +x "$INSTALL_DIR/singdns.sh"
    chmod +x "$INSTALL_DIR/bin/sing-box"
    chmod -R 755 "$INSTALL_DIR/bin"
    chmod -R 755 "$INSTALL_DIR/configs"
    chmod 755 "$LOG_DIR"
    
    # 创建符号链接
    ln -sf "$INSTALL_DIR/singdns.sh" "/usr/local/bin/singdns"
    
    echo "${GREEN}SingDNS 安装完成${NC}"
    return 0
}

# 卸载函数
uninstall_singdns() {
    echo "${YELLOW}开始卸载 SingDNS...${NC}"
    
    # 停止服务
    if command -v singdns > /dev/null; then
        singdns stop
        sleep 2
    fi
    
    # 删除文件
    rm -rf "$INSTALL_DIR"
    rm -f "/usr/local/bin/singdns"
    rm -rf "$LOG_DIR"
    
    echo "${GREEN}SingDNS 已成功卸载${NC}"
    return 0
}

# 主函数
main() {
    echo "${YELLOW}欢迎使用 SingDNS 安装程序${NC}"
    echo "系统类型: $(detect_os)"
    echo "请选择操作："
    echo "1. 安装 SingDNS"
    echo "2. 卸载 SingDNS"
    echo "3. 退出"
    
    printf "请输入选项 [1-3]: "
    read choice
    
    case $choice in
        1)
            check_root || exit 1
            install_system_dependencies || exit 1
            install_singdns || exit 1
            
            # 显示安装成功信息
            echo "${GREEN}SingDNS 安装成功！${NC}"
            echo "${GREEN}使用 'singdns' 命令来管理 SingDNS${NC}"
            echo "${YELLOW}提示: 使用 'singdns start' 启动服务${NC}"
            echo "安装目录: ${INSTALL_DIR}"
            echo "日志目录: ${LOG_DIR}"
            ;;
        2)
            check_root || exit 1
            uninstall_singdns
            ;;
        3)
            echo "${GREEN}感谢使用！${NC}"
            exit 0
            ;;
        *)
            echo "${RED}无效的选项${NC}"
            exit 1
            ;;
    esac
}

# 执行主函数
main 