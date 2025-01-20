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
REQUIRED_PORTS="8080 3000 9090"

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
        apk add --no-cache curl wget git sqlite nginx iptables ip6tables
    elif command -v apt-get > /dev/null; then
        # Debian/Ubuntu
        apt-get update
        apt-get install -y curl wget git sqlite3 nginx iptables
    elif command -v yum > /dev/null; then
        # CentOS/RHEL
        yum install -y curl wget git sqlite nginx iptables
    else
        echo "${RED}不支持的操作系统${NC}"
        return 1
    fi
    
    echo "${GREEN}系统依赖安装完成${NC}"
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
    
    # 复制主程序
    cp singdns "$INSTALL_DIR/"
    
    # 复制二进制文件和面板
    cp bin/sing-box "$INSTALL_DIR/bin/"
    cp -r bin/web/* "$INSTALL_DIR/bin/web/"  # 复制面板文件
    
    # 复制前端文件
    cp -r web/* "$INSTALL_DIR/web/"  # 复制前端文件到根目录的 web
    
    # 复制配置文件
    cp -r configs/* "$INSTALL_DIR/configs/"
    
    # 复制管理脚本
    cp singdns.sh "$INSTALL_DIR/"
    cp install.sh "$INSTALL_DIR/"
    
    # 复制文档
    cp README.md "$INSTALL_DIR/" || true
    cp LICENSE "$INSTALL_DIR/" || true
    cp VERSION "$INSTALL_DIR/" || true
    
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

# 验证安装
verify_installation() {
    echo "${BLUE}验证安装...${NC}"
    
    # 检查关键文件
    local required_files=(
        "$INSTALL_DIR/singdns"
        "$INSTALL_DIR/bin/sing-box"
        "$INSTALL_DIR/configs/sing-box/config.json"
        "$INSTALL_DIR/singdns.sh"
    )
    
    for file in "${required_files[@]}"; do
        if [ ! -f "$file" ]; then
            echo "${RED}错误：文件不存在 $file${NC}"
            return 1
        fi
    done
    
    # 验证可执行权限
    if [ ! -x "$INSTALL_DIR/singdns" ] || [ ! -x "$INSTALL_DIR/bin/sing-box" ]; then
        echo "${RED}错误：可执行文件权限设置失败${NC}"
        return 1
    fi
    
    echo "${GREEN}安装验证完成${NC}"
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
    
    # 清理防火墙规则
    iptables -F
    iptables -X
    iptables -t nat -F
    iptables -t nat -X
    iptables -t mangle -F
    iptables -t mangle -X
    
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
    echo "请选择操作："
    echo "1. 安装 SingDNS"
    echo "2. 卸载 SingDNS"
    echo "3. 退出"
    
    printf "请输入选项 [1-3]: "
    read choice
    
    case $choice in
        1)
            check_root || exit 1
            check_disk_space || exit 1
            check_ports || exit 1
            check_existing_installation || exit 1
            install_system_dependencies || exit 1
            create_directories || exit 1
            copy_files || exit 1
            verify_installation || exit 1
            
            echo "${GREEN}SingDNS 安装成功！${NC}"
            echo "${GREEN}使用 'singdns' 命令来管理 SingDNS${NC}"
            echo "${YELLOW}提示: 使用 'singdns start' 启动服务${NC}"
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