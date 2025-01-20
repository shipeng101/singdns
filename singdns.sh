#!/bin/bash

# 设置颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 安装目录
INSTALL_DIR="/usr/local/singdns"
BIN_DIR="$INSTALL_DIR/bin"
CONFIG_DIR="$INSTALL_DIR/configs/sing-box"
LOG_DIR="/var/log/singdns"
WEB_DIR="$INSTALL_DIR/web"  # 前端目录
DATA_DIR="$INSTALL_DIR/data"
BACKUP_DIR="$INSTALL_DIR/backups"

# 检查是否为root用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}错误：请使用root用户运行此脚本${NC}"
        exit 1
    fi
}

# 获取本机IP地址
get_ip() {
    local ip
    # 尝试获取本地IP
    ip=$(ip -4 addr show scope global | grep inet | awk '{print $2}' | cut -d'/' -f1 | head -n 1)
    
    # 如果本地IP获取失败，尝试获取公网IP
    if [ -z "$ip" ]; then
        ip=$(curl -s -m 5 ifconfig.me || wget -qO- -T 5 ifconfig.me)
    fi
    
    # 如果都失败了，返回错误
    if [ -z "$ip" ]; then
        echo "${RED}错误：无法获取IP地址${NC}" >&2
        return 1
    fi
    
    echo "$ip"
}

# 检查依赖
check_dependencies() {
    local deps=("curl" "wget" "iptables")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            echo -e "${YELLOW}安装 $dep...${NC}"
            install_package "$dep"
        fi
    done
}

# 检查服务状态
check_status() {
    local status=0
    local ip
    ip=$(get_ip) || ip="<ip>"
    
    if pgrep -f "singdns.*serve" > /dev/null; then
        echo -e "${GREEN}SingDNS 服务正在运行${NC}"
        echo -e "${GREEN}管理界面: http://${ip}:3000${NC}"
    else
        echo -e "${RED}SingDNS 服务未运行${NC}"
        status=1
    fi

    return $status
}

# 启动服务
start_service() {
    # 检查是否已经在运行
    if pgrep -f "singdns.*serve" > /dev/null; then
        echo -e "${YELLOW}SingDNS 服务已在运行${NC}"
        return
    fi

    # 创建必要的目录
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "$WEB_DIR"

    # 启动服务
    cd "$INSTALL_DIR" || exit 1
    nohup "$BIN_DIR/singdns" serve > "$LOG_DIR/singdns.log" 2>&1 &

    # 等待服务启动
    local timeout=30
    local counter=0
    while ! curl -s http://localhost:3000 > /dev/null && [ $counter -lt $timeout ]; do
        sleep 1
        ((counter++))
    done

    if [ $counter -eq $timeout ]; then
        echo -e "${RED}SingDNS 服务启动超时${NC}"
        exit 1
    fi

    # 检查服务状态
    if pgrep -f "singdns.*serve" > /dev/null; then
        echo -e "${GREEN}SingDNS 服务已启动${NC}"
        check_status
    else
        echo -e "${RED}SingDNS 服务启动失败${NC}"
        exit 1
    fi
}

# 停止服务
stop_service() {
    if ! pgrep -f "singdns.*serve" > /dev/null; then
        echo -e "${YELLOW}SingDNS 服务未运行${NC}"
        return
    fi

    pkill -f "singdns.*serve"
    sleep 2

    if ! pgrep -f "singdns.*serve" > /dev/null; then
        echo -e "${GREEN}SingDNS 服务已停止${NC}"
    else
        echo -e "${RED}SingDNS 服务停止失败${NC}"
        exit 1
    fi
}

# 配置备份
backup_config() {
    local backup_time=$(date +%Y%m%d_%H%M%S)
    local backup_path="$BACKUP_DIR/$backup_time"
    
    mkdir -p "$backup_path"
    
    echo -e "${YELLOW}备份配置文件...${NC}"
    cp -r "$CONFIG_DIR" "$backup_path/"
    cp -r "$DATA_DIR" "$backup_path/"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}配置已备份到: $backup_path${NC}"
        return 0
    else
        echo -e "${RED}配置备份失败${NC}"
        return 1
    fi
}

# 配置恢复
restore_config() {
    local backups=()
    local i=1
    
    echo -e "${YELLOW}可用的备份:${NC}"
    for backup in "$BACKUP_DIR"/*; do
        if [ -d "$backup" ]; then
            echo "$i) $(basename "$backup")"
            backups+=("$backup")
            ((i++))
        fi
    done
    
    if [ ${#backups[@]} -eq 0 ]; then
        echo -e "${RED}没有找到可用的备份${NC}"
        return 1
    fi
    
    read -p "请选择要恢复的备份 [1-${#backups[@]}]: " choice
    
    if [[ ! "$choice" =~ ^[0-9]+$ ]] || [ "$choice" -lt 1 ] || [ "$choice" -gt ${#backups[@]} ]; then
        echo -e "${RED}无效的选择${NC}"
        return 1
    fi
    
    local selected_backup="${backups[$((choice-1))]}"
    
    echo -e "${YELLOW}正在恢复配置...${NC}"
    cp -r "$selected_backup/sing-box"/* "$CONFIG_DIR/"
    cp -r "$selected_backup/data"/* "$DATA_DIR/"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}配置已恢复${NC}"
        return 0
    else
        echo -e "${RED}配置恢复失败${NC}"
        return 1
    fi
}

# 查看日志
view_logs() {
    if [ -f "$LOG_DIR/singdns.log" ]; then
        tail -f "$LOG_DIR/singdns.log"
    else
        echo -e "${RED}日志文件不存在${NC}"
    fi
}

# 查看最近日志
view_recent_logs() {
    if [ -f "$LOG_DIR/singdns.log" ]; then
        tail -n 20 "$LOG_DIR/singdns.log"
    else
        echo -e "${RED}日志文件不存在${NC}"
    fi
}

# 查看进程状态
view_process() {
    ps aux | grep "singdns" | grep -v grep
}

# 查看端口占用
check_ports() {
    echo -e "${BLUE}=== 端口占用情况 ===${NC}"
    echo -e "${YELLOW}管理界面端口 (3000)：${NC}"
    netstat -tunlp | grep ":3000 "
    
    echo -e "\n${YELLOW}所有 SingDNS 相关进程：${NC}"
    ps aux | grep "singdns" | grep -v grep
}

# 显示菜单
show_menu() {
    echo -e "${BLUE}=== SingDNS 管理面板 ===${NC}"
    echo "1. 启动服务"
    echo "2. 停止服务"
    echo "3. 查看状态"
    echo "4. 查看日志"
    echo "5. 查看进程状态"
    echo "6. 查看端口占用"
    echo "7. 备份配置"
    echo "8. 恢复配置"
    echo "0. 退出"
    echo -e "${BLUE}=====================${NC}"
}

# 主程序
main() {
    # 创建必要的目录
    mkdir -p $LOG_DIR
    mkdir -p $BACKUP_DIR
    
    # 检查是否为root用户
    check_root
    
    while true; do
        show_menu
        read -p "请选择操作 [0-8]: " choice
        
        case $choice in
            1)
                start_service
                ;;
            2)
                stop_service
                ;;
            3)
                check_status
                ;;
            4)
                view_logs
                ;;
            5)
                view_process
                ;;
            6)
                check_ports
                ;;
            7)
                backup_config
                ;;
            8)
                restore_config
                ;;
            0)
                echo -e "${GREEN}感谢使用！${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}无效的选项${NC}"
                ;;
        esac
        
        echo
        read -p "按回车键继续..."
    done
}

# 处理命令行参数
if [ $# -gt 0 ]; then
    case "$1" in
        start)
            check_root
            start_service
            ;;
        stop)
            check_root
            stop_service
            ;;
        status)
            check_status
            ;;
        logs)
            if [ "$2" = "recent" ]; then
                view_recent_logs
            else
                view_logs
            fi
            ;;
        backup)
            check_root
            backup_config
            ;;
        restore)
            check_root
            restore_config
            ;;
        *)
            echo "用法: singdns {start|stop|status|logs|backup|restore}"
            exit 1
            ;;
    esac
else
    main
fi