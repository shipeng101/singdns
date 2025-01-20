#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 基本配置
INSTALL_DIR="/usr/local/singdns"
BIN_DIR="$INSTALL_DIR/bin"
CONFIG_DIR="$INSTALL_DIR/configs/sing-box"
LOG_DIR="/var/log/singdns"
WEB_DIR="$INSTALL_DIR/web"
REQUIRED_PORTS=("3000")

# 检查是否为 root 用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}错误: 必须使用 root 用户运行此脚本${NC}"
        exit 1
    fi
}

# 检查系统环境
check_environment() {
    echo -e "${BLUE}检查系统环境...${NC}"
    
    # 检查操作系统
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo -e "操作系统: ${GREEN}$NAME${NC}"
    else
        echo -e "${RED}无法确定操作系统类型${NC}"
        exit 1
    fi
    
    # 检查系统架构
    ARCH=$(uname -m)
    echo -e "系统架构: ${GREEN}$ARCH${NC}"
    if [[ $ARCH != "x86_64" ]]; then
        echo -e "${RED}错误: 仅支持 x86_64 架构${NC}"
        exit 1
    fi
    
    # 检查内存使用情况
    TOTAL_MEM=$(free -m | awk '/^Mem:/{print $2}')
    FREE_MEM=$(free -m | awk '/^Mem:/{print $4}')
    echo -e "总内存: ${GREEN}${TOTAL_MEM}MB${NC}"
    echo -e "可用内存: ${GREEN}${FREE_MEM}MB${NC}"
    
    # 检查磁盘空间
    DISK_SPACE=$(df -h / | awk 'NR==2 {print $4}')
    echo -e "可用磁盘空间: ${GREEN}${DISK_SPACE}${NC}"
    
    # 检查必要命令
    REQUIRED_COMMANDS=("tar" "systemctl" "busybox")
    for cmd in "${REQUIRED_COMMANDS[@]}"; do
        if ! command -v $cmd &> /dev/null; then
            echo -e "${RED}错误: 未找到命令 '$cmd'${NC}"
            exit 1
        fi
    done
    
    # 检查端口占用
    for port in "${REQUIRED_PORTS[@]}"; do
        if netstat -tuln | grep -q ":$port "; then
            echo -e "${RED}错误: 端口 $port 已被占用${NC}"
            exit 1
        fi
    done
    
    echo -e "${GREEN}系统环境检查通过${NC}"
}

# 安装系统依赖
install_dependencies() {
    echo -e "${BLUE}安装系统依赖...${NC}"
    
    if [ -f /etc/alpine-release ]; then
        apk update
        apk add --no-cache tar busybox
    elif [ -f /etc/debian_version ]; then
        apt-get update
        DEBIAN_FRONTEND=noninteractive apt-get install -y tar busybox
    elif [ -f /etc/redhat-release ]; then
        yum install -y tar busybox
    else
        echo -e "${RED}不支持的操作系统${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}系统依赖安装完成${NC}"
}

# 复制文件
copy_files() {
    echo -e "${BLUE}复制文件...${NC}"
    
    # 创建必要的目录
    mkdir -p "$INSTALL_DIR" "$BIN_DIR" "$CONFIG_DIR" "$LOG_DIR" "$WEB_DIR"
    
    # 设置权限
    chmod +x "$BIN_DIR/singdns"
    chmod +x "$INSTALL_DIR/singdns.sh"
    
    # 创建全局命令链接
    ln -sf "$INSTALL_DIR/singdns.sh" "/usr/local/bin/singdns"
    
    echo -e "${GREEN}文件复制完成${NC}"
}

# 配置系统服务
setup_service() {
    echo -e "${BLUE}配置系统服务...${NC}"
    
    # 创建 systemd 服务文件
    cat > /etc/systemd/system/singdns.service << EOF
[Unit]
Description=SingDNS Service
After=network.target

[Service]
Type=simple
ExecStart=$BIN_DIR/singdns serve
WorkingDirectory=$INSTALL_DIR
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
    
    # 重新加载 systemd
    systemctl daemon-reload
    systemctl enable singdns
    systemctl start singdns
    
    echo -e "${GREEN}系统服务配置完成${NC}"
}

# 卸载函数
uninstall() {
    echo -e "${YELLOW}开始卸载 SingDNS...${NC}"
    
    # 停止并禁用服务
    systemctl stop singdns
    systemctl disable singdns
    rm -f /etc/systemd/system/singdns.service
    systemctl daemon-reload
    
    # 停止前端服务
    pkill -f "busybox httpd -f -p 3000" || true
    
    # 删除安装目录和命令链接
    rm -rf "$INSTALL_DIR"
    rm -rf "$LOG_DIR"
    rm -f "/usr/local/bin/singdns"
    
    echo -e "${GREEN}SingDNS 已成功卸载${NC}"
}

# 显示菜单
show_menu() {
    echo -e "\n${BLUE}=== SingDNS 安装管理脚本 ===${NC}"
    echo -e "1. 安装 SingDNS"
    echo -e "2. 卸载 SingDNS"
    echo -e "3. 检查系统环境"
    echo -e "0. 退出脚本"
    echo -e "${BLUE}========================${NC}"
    echo -n "请输入选项 [0-3]: "
}

# 主函数
main() {
    # 检查命令行参数
    if [[ $1 == "install" ]]; then
        check_root
        check_environment
        install_dependencies
        copy_files
        setup_service
        echo -e "\n${GREEN}SingDNS 安装完成！${NC}"
        echo -e "使用 ${YELLOW}singdns help${NC} 查看使用说明"
        exit 0
    elif [[ $1 == "uninstall" ]]; then
        check_root
        uninstall
        exit 0
    fi
    
    # 显示菜单
    while true; do
        show_menu
        read -r choice
        case $choice in
            1)
                check_root
                check_environment
                install_dependencies
                copy_files
                setup_service
                echo -e "\n${GREEN}SingDNS 安装完成！${NC}"
                echo -e "使用 ${YELLOW}singdns help${NC} 查看使用说明"
                ;;
            2)
                check_root
                uninstall
                ;;
            3)
                check_environment
                ;;
            0)
                echo -e "${GREEN}感谢使用！${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}无效的选项，请重新选择${NC}"
                ;;
        esac
    done
}

# 执行主函数
main "$@" 