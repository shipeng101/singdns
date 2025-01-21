#!/bin/bash

# 设置变量
INSTALL_DIR="/usr/local/singdns"
LOG_DIR="/var/log/singdns"
SERVICE_NAME="singdns"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 检查是否为 root 用户
check_root() {
    if [ "$(id -u)" != "0" ]; then
        echo -e "${RED}错误: 必须使用 root 权限运行此脚本${NC}"
        exit 1
    fi
}

# 检查系统类型
check_system() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
    else
        echo -e "${RED}无法确定操作系统类型${NC}"
        exit 1
    fi
}

# 检查并安装依赖
check_dependencies() {
    echo -e "${YELLOW}检查系统依赖...${NC}"
    local missing_deps=()
    
    # 检查必要的命令
    local deps=("curl" "tar" "systemctl" "busybox")
    for dep in "${deps[@]}"; do
        if ! command -v $dep &> /dev/null; then
            missing_deps+=($dep)
        fi
    done
    
    # 如果有缺失的依赖，尝试安装
    if [ ${#missing_deps[@]} -ne 0 ]; then
        echo -e "${YELLOW}安装缺失的依赖: ${missing_deps[*]}${NC}"
        case $OS in
            ubuntu|debian)
                apt update
                apt install -y ${missing_deps[@]}
                ;;
            centos|rhel|fedora)
                yum install -y ${missing_deps[@]}
                ;;
            *)
                echo -e "${RED}不支持的操作系统${NC}"
                exit 1
                ;;
        esac
    fi
    
    echo -e "${GREEN}所有依赖已满足${NC}"
}

# 安装函数
install() {
    echo -e "${YELLOW}开始安装 SingDNS...${NC}"
    
    # 创建必要的目录
    echo -e "${YELLOW}创建目录...${NC}"
    mkdir -p $INSTALL_DIR
    mkdir -p $LOG_DIR
    
    # 复制文件
    echo -e "${YELLOW}复制文件...${NC}"
    cp -r bin/* $INSTALL_DIR/
    cp -r configs $INSTALL_DIR/
    cp -r web $INSTALL_DIR/
    cp singdns $INSTALL_DIR/
    cp singdns.sh /usr/local/bin/singdns
    
    # 设置权限
    echo -e "${YELLOW}设置权限...${NC}"
    chmod +x $INSTALL_DIR/singdns
    chmod +x /usr/local/bin/singdns
    chown -R root:root $INSTALL_DIR
    chown -R root:root $LOG_DIR
    
    # 创建 systemd 服务
    echo -e "${YELLOW}创建系统服务...${NC}"
    cat > /etc/systemd/system/$SERVICE_NAME.service << EOF
[Unit]
Description=SingDNS Service
After=network.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/singdns
WorkingDirectory=$INSTALL_DIR
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOF
    
    # 重新加载 systemd
    systemctl daemon-reload
    
    echo -e "${GREEN}安装完成！${NC}"
    echo -e "${GREEN}使用以下命令管理服务：${NC}"
    echo -e "${YELLOW}启动服务: ${NC}systemctl start $SERVICE_NAME"
    echo -e "${YELLOW}停止服务: ${NC}systemctl stop $SERVICE_NAME"
    echo -e "${YELLOW}查看状态: ${NC}systemctl status $SERVICE_NAME"
    echo -e "${YELLOW}开机启动: ${NC}systemctl enable $SERVICE_NAME"
}

# 卸载函数
uninstall() {
    echo -e "${YELLOW}开始卸载 SingDNS...${NC}"
    
    # 停止服务
    if systemctl is-active --quiet $SERVICE_NAME; then
        echo -e "${YELLOW}停止服务...${NC}"
        systemctl stop $SERVICE_NAME
    fi
    
    # 禁用服务
    if systemctl is-enabled --quiet $SERVICE_NAME; then
        echo -e "${YELLOW}禁用服务...${NC}"
        systemctl disable $SERVICE_NAME
    fi
    
    # 删除服务文件
    echo -e "${YELLOW}删除服务文件...${NC}"
    rm -f /etc/systemd/system/$SERVICE_NAME.service
    systemctl daemon-reload
    
    # 删除安装文件
    echo -e "${YELLOW}删除安装文件...${NC}"
    rm -rf $INSTALL_DIR
    rm -rf $LOG_DIR
    rm -f /usr/local/bin/singdns
    
    echo -e "${GREEN}卸载完成！${NC}"
}

# 显示菜单
show_menu() {
    echo -e "${GREEN}=== SingDNS 安装管理器 ===${NC}"
    echo -e "${YELLOW}1.${NC} 安装 SingDNS"
    echo -e "${YELLOW}2.${NC} 卸载 SingDNS"
    echo -e "${YELLOW}3.${NC} 退出"
    echo
    read -p "请选择操作 [1-3]: " choice
    
    case $choice in
        1)
            check_root
            check_system
            check_dependencies
            install
            ;;
        2)
            check_root
            uninstall
            ;;
        3)
            echo -e "${GREEN}再见！${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}无效的选择${NC}"
            exit 1
            ;;
    esac
}

# 主程序
show_menu 