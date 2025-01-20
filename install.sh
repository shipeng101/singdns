#!/bin/bash

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
REQUIRED_PORTS="3000"
GITHUB_REPO="shipeng101/singdns"
LATEST_VERSION="v1.0.7"  # 最新版本号

# 检查是否为root用户
check_root() {
    if [ "$(id -u)" != "0" ]; then
        echo "${RED}错误：请使用root用户运行此脚本${NC}"
        return 1
    fi
    return 0
}

# 检查磁盘空间
check_disk_space() {
    local available_space=$(df -m "$INSTALL_DIR" | awk 'NR==2 {print $4}')
    if [ "$available_space" -lt "$MIN_DISK_SPACE" ]; then
        echo "${RED}错误：磁盘空间不足，需要至少 ${MIN_DISK_SPACE}MB${NC}"
        return 1
    fi
    return 0
}

# 检查端口占用
check_ports() {
    for port in $REQUIRED_PORTS; do
        if netstat -tuln | grep -q ":$port "; then
            echo "${RED}错误：端口 $port 已被占用${NC}"
            return 1
        fi
    done
    return 0
}

# 检查现有安装
check_existing_installation() {
    if [ -d "$INSTALL_DIR" ]; then
        echo "${YELLOW}检测到已存在的安装，是否继续？[y/N]${NC}"
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            return 1
        fi
        # 备份现有配置
        backup_dir="/tmp/singdns_backup_$(date +%Y%m%d_%H%M%S)"
        mkdir -p "$backup_dir"
        if [ -d "$INSTALL_DIR/configs" ]; then
            cp -r "$INSTALL_DIR/configs" "$backup_dir/"
            echo "${GREEN}已备份现有配置到 $backup_dir${NC}"
        fi
    fi
    return 0
}

# 安装系统依赖
install_system_dependencies() {
    echo "${BLUE}安装系统依赖...${NC}"
    
    # 检查包管理器
    if command -v apk > /dev/null; then
        # Alpine Linux
        apk update
        apk add --no-cache curl wget git sqlite iptables ip6tables
    elif command -v apt-get > /dev/null; then
        # Debian/Ubuntu
        apt-get update
        apt-get install -y curl wget git sqlite3 iptables
    elif command -v yum > /dev/null; then
        # CentOS/RHEL
        yum install -y curl wget git sqlite iptables
    else
        echo "${RED}不支持的操作系统${NC}"
        return 1
    fi
    
    echo "${GREEN}系统依赖安装完成${NC}"
    return 0
}

# 下载安装包
download_package() {
    echo "${BLUE}下载安装包...${NC}"
    
    # 创建临时目录
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR" || exit 1
    
    # 下载最新版本
    local package_url="https://github.com/${GITHUB_REPO}/releases/download/${LATEST_VERSION}/singdns-linux-amd64.tar.gz"
    echo "${YELLOW}正在下载: ${package_url}${NC}"
    
    if ! wget -O singdns.tar.gz "$package_url"; then
        echo "${RED}下载失败${NC}"
        return 1
    fi
    
    # 解压安装包
    echo "${YELLOW}解压安装包...${NC}"
    if ! tar -xzf singdns.tar.gz; then
        echo "${RED}解压失败${NC}"
        return 1
    fi
    
    echo "${GREEN}安装包下载并解压完成${NC}"
    return 0
}

# 创建必要的目录结构
create_directories() {
    echo "${BLUE}创建目录结构...${NC}"
    
    # 创建主要目录
    mkdir -p "$INSTALL_DIR"/bin
    mkdir -p "$INSTALL_DIR"/web
    mkdir -p "$INSTALL_DIR"/configs/sing-box/rules
    mkdir -p "$INSTALL_DIR"/data
    mkdir -p "$INSTALL_DIR"/bin/web  # 面板目录
    mkdir -p "$LOG_DIR"
    
    if [ $? -ne 0 ]; then
        echo "${RED}创建目录失败${NC}"
        return 1
    fi
    
    echo "${GREEN}目录结构创建完成${NC}"
    return 0
}

# 复制文件
copy_files() {
    echo "${YELLOW}复制文件...${NC}"
    
    cd "$TEMP_DIR/singdns" || exit 1
    
    # 复制所有文件到安装目录
    cp -r * "$INSTALL_DIR/"
    
    # 设置权限
    find "$INSTALL_DIR" -type f -exec chmod 644 {} \;
    find "$INSTALL_DIR" -type d -exec chmod 755 {} \;
    chmod +x "$INSTALL_DIR/singdns"
    chmod +x "$INSTALL_DIR/singdns.sh"
    chmod +x "$INSTALL_DIR/bin/sing-box"
    
    # 创建符号链接
    ln -sf "$INSTALL_DIR/singdns.sh" "/usr/local/bin/singdns"
    
    echo "${GREEN}文件复制完成${NC}"
    return 0
}

# 清理临时文件
cleanup() {
    echo "${YELLOW}清理临时文件...${NC}"
    rm -rf "$TEMP_DIR"
    echo "${GREEN}清理完成${NC}"
}

# 主安装函数
install() {
    echo "${BLUE}开始安装 SingDNS...${NC}"
    
    # 检查是否为root用户
    check_root || exit 1
    
    # 检查磁盘空间
    check_disk_space || exit 1
    
    # 检查端口占用
    check_ports || exit 1
    
    # 检查现有安装
    check_existing_installation || exit 1
    
    # 安装系统依赖
    install_system_dependencies || exit 1
    
    # 下载安装包
    download_package || exit 1
    
    # 创建目录结构
    create_directories || exit 1
    
    # 复制文件
    copy_files || exit 1
    
    # 清理临时文件
    cleanup
    
    echo "${GREEN}SingDNS 安装完成！${NC}"
    echo "${GREEN}使用 'singdns' 命令管理服务${NC}"
    return 0
}

# 卸载函数
uninstall() {
    echo "${YELLOW}开始卸载 SingDNS...${NC}"
    
    # 检查是否为root用户
    check_root || exit 1
    
    # 停止服务
    if [ -f "$INSTALL_DIR/singdns.sh" ]; then
        "$INSTALL_DIR/singdns.sh" stop
    fi
    
    # 删除安装目录
    rm -rf "$INSTALL_DIR"
    
    # 删除日志目录
    rm -rf "$LOG_DIR"
    
    # 删除符号链接
    rm -f "/usr/local/bin/singdns"
    
    echo "${GREEN}SingDNS 卸载完成！${NC}"
    return 0
}

# 主函数
main() {
    if [ $# -eq 0 ]; then
        install
    else
        case "$1" in
            install)
                install
                ;;
            uninstall)
                uninstall
                ;;
            *)
                echo "用法: $0 {install|uninstall}"
                exit 1
                ;;
        esac
    fi
}

main "$@" 