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
GITHUB_API="https://api.github.com/repos/shipeng101/singdns/releases/latest"
GITHUB_DOWNLOAD="https://github.com/shipeng101/singdns/releases/download"

# 检查是否为root用户
check_root() {
    if [ "$(id -u)" != "0" ]; then
        echo "${RED}错误：请使用root用户运行此脚本${NC}"
        return 1
    fi
    return 0
}

# 下载最新版本
download_latest_release() {
    echo "${BLUE}正在获取最新版本信息...${NC}"
    
    # 获取最新版本号
    VERSION=$(curl -s $GITHUB_API | grep "tag_name" | cut -d'"' -f4)
    if [ -z "$VERSION" ]; then
        echo "${RED}获取版本信息失败${NC}"
        return 1
    fi
    
    echo "${BLUE}最新版本: ${VERSION}${NC}"
    
    # 创建临时目录
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR" || exit 1
    
    # 下载最新版本
    echo "${BLUE}下载安装包...${NC}"
    DOWNLOAD_URL="${GITHUB_DOWNLOAD}/${VERSION}/singdns-linux-amd64.tar.gz"
    if ! curl -L -o singdns.tar.gz "$DOWNLOAD_URL"; then
        echo "${RED}下载失败${NC}"
        return 1
    fi
    
    # 解压
    echo "${BLUE}解压安装包...${NC}"
    tar xzf singdns.tar.gz
    
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
    
    # 下载最新版本
    if ! download_latest_release; then
        echo "${RED}下载失败${NC}"
        return 1
    fi
    
    # 创建必要的目录
    mkdir -p "$INSTALL_DIR/bin/web"
    mkdir -p "$INSTALL_DIR/web"
    mkdir -p "$INSTALL_DIR/configs/sing-box/rules"
    mkdir -p "$LOG_DIR"
    
    # 复制文件
    echo "${YELLOW}复制文件...${NC}"
    cp -r "$TEMP_DIR"/singdns/* "$INSTALL_DIR/"
    
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
    
    # 清理临时文件
    rm -rf "$TEMP_DIR"
    
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
    
    # 检查root权限
    check_root || exit 1
    
    # 安装系统依赖
    install_system_dependencies || exit 1
    
    # 安装SingDNS
    install_singdns || exit 1
    
    # 显示安装成功信息
    echo "${GREEN}SingDNS 安装成功！${NC}"
    echo "${GREEN}使用 'singdns' 命令来管理 SingDNS${NC}"
    echo "${YELLOW}提示: 使用 'singdns start' 启动服务${NC}"
    echo "安装目录: ${INSTALL_DIR}"
    echo "日志目录: ${LOG_DIR}"
    
    # 询问是否立即启动服务
    printf "是否立即启动服务？[y/N] "
    read -r answer
    case $answer in
        [Yy]*)
            singdns start
            ;;
    esac
}

# 执行主函数
main 