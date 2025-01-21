#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 安装目录
INSTALL_DIR="/usr/local/singdns"
LOG_DIR="/var/log/singdns"
SERVICE_NAME="singdns"

# 检查是否为root用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}错误: 请使用 root 用户运行此脚本${NC}"
        exit 1
    fi
}

# 检查系统依赖
check_dependencies() {
    echo -e "${YELLOW}检查系统依赖...${NC}"
    local deps_missing=0
    
    # 检查必要的命令
    for cmd in curl wget systemctl busybox; do
        if ! command -v $cmd &> /dev/null; then
            echo -e "${RED}未找到命令: $cmd${NC}"
            deps_missing=1
        fi
    done
    
    if [ $deps_missing -eq 1 ]; then
        echo -e "${YELLOW}正在安装缺失的依赖...${NC}"
        if command -v apt &> /dev/null; then
            apt update
            apt install -y curl wget systemd busybox
        elif command -v yum &> /dev/null; then
            yum install -y curl wget systemd busybox
        else
            echo -e "${RED}不支持的系统，请手动安装依赖${NC}"
            exit 1
        fi
    else
        echo -e "${GREEN}所有依赖已满足${NC}"
    fi
}

# 检查必需文件
check_files() {
    echo -e "${YELLOW}检查必需文件...${NC}"
    local missing_files=0
    
    # 检查主程序
    if [ ! -f "singdns" ]; then
        echo -e "${RED}错误: 未找到主程序 'singdns'${NC}"
        missing_files=1
    fi
    
    # 检查管理脚本
    if [ ! -f "singdns.sh" ]; then
        echo -e "${RED}错误: 未找到管理脚本 'singdns.sh'${NC}"
        missing_files=1
    fi
    
    # 检查 bin 目录
    if [ ! -d "bin" ] || [ ! -f "bin/sing-box" ]; then
        echo -e "${RED}错误: 未找到 bin 目录或 sing-box 程序${NC}"
        missing_files=1
    fi
    
    # 检查 configs 目录
    if [ ! -d "configs" ]; then
        echo -e "${RED}错误: 未找到 configs 目录${NC}"
        missing_files=1
    fi
    
    # 检查 web 目录
    if [ ! -d "web" ]; then
        echo -e "${RED}错误: 未找到 web 目录${NC}"
        missing_files=1
    fi
    
    if [ $missing_files -eq 1 ]; then
        echo -e "${RED}错误: 缺少必需文件，安装终止${NC}"
        exit 1
    fi
}

# 安装服务
install_service() {
    echo -e "${YELLOW}开始安装 SingDNS...${NC}"
    
    # 检查文件
    check_files
    
    # 停止现有服务
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        echo -e "${YELLOW}停止现有服务...${NC}"
        systemctl stop ${SERVICE_NAME}
    fi
    
    # 创建目录
    echo "创建目录..."
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$LOG_DIR"
    
    # 复制文件
    echo "复制文件..."
    # 复制主程序
    cp singdns "$INSTALL_DIR/"
    
    # 复制管理脚本
    cp singdns.sh "$INSTALL_DIR/"
    
    # 复制 bin 目录
    cp -r bin "$INSTALL_DIR/"
    
    # 复制 configs 目录
    cp -r configs "$INSTALL_DIR/"
    
    # 复制 web 目录
    cp -r web "$INSTALL_DIR/"
    
    # 设置权限
    echo "设置权限..."
    chmod +x "$INSTALL_DIR/singdns"
    chmod +x "$INSTALL_DIR/singdns.sh"
    chmod +x "$INSTALL_DIR/bin/sing-box"
    chown -R root:root "$INSTALL_DIR"
    chown -R root:root "$LOG_DIR"
    
    # 创建软链接
    ln -sf "$INSTALL_DIR/singdns.sh" /usr/local/bin/singdns
    
    # 创建服务文件
    cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
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
    echo -e "使用以下命令管理服务："
    echo -e "直接运行管理面板: ${YELLOW}singdns${NC}"
    echo -e "系统服务管理:"
    echo -e "  启动服务: ${YELLOW}systemctl start ${SERVICE_NAME}${NC}"
    echo -e "  停止服务: ${YELLOW}systemctl stop ${SERVICE_NAME}${NC}"
    echo -e "  查看状态: ${YELLOW}systemctl status ${SERVICE_NAME}${NC}"
    echo -e "  开机启动: ${YELLOW}systemctl enable ${SERVICE_NAME}${NC}"
}

# 卸载服务
uninstall_service() {
    echo -e "${YELLOW}开始卸载 SingDNS...${NC}"
    
    # 停止服务
    systemctl stop ${SERVICE_NAME}
    systemctl disable ${SERVICE_NAME}
    
    # 删除文件
    rm -f /etc/systemd/system/${SERVICE_NAME}.service
    rm -f /usr/local/bin/singdns
    rm -rf "$INSTALL_DIR"
    rm -rf "$LOG_DIR"
    
    # 重新加载 systemd
    systemctl daemon-reload
    
    echo -e "${GREEN}卸载完成！${NC}"
}

# 显示菜单
show_menu() {
    clear
    echo -e "${GREEN}=== SingDNS 安装管理器 ===${NC}"
    echo "1. 安装 SingDNS"
    echo "2. 卸载 SingDNS"
    echo "0. 退出"
    echo
    echo -n "请选择操作 [0-2]: "
}

# 主循环
main() {
    check_root
    
    while true; do
        show_menu
        read -r choice
        echo
        
        case $choice in
            1)
                check_dependencies
                install_service
                ;;
            2)
                uninstall_service
                ;;
            0)
                echo -e "${GREEN}再见！${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}无效的选项，请重试${NC}"
                ;;
        esac
        
        echo
        read -n 1 -s -r -p "按任意键继续..."
    done
}

main 