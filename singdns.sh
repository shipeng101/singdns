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
    local deps=("nginx" "curl" "wget" "iptables")
    local missing=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" >/dev/null 2>&1; then
            missing+=("$dep")
        fi
    done
    
    if [ ${#missing[@]} -ne 0 ]; then
        echo -e "${RED}缺少依赖：${missing[*]}${NC}"
        echo -e "${YELLOW}正在安装缺失的依赖...${NC}"
        
        if command -v apk >/dev/null 2>&1; then
            apk add --no-cache "${missing[@]}"
        elif command -v apt-get >/dev/null 2>&1; then
            apt-get update && apt-get install -y "${missing[@]}"
        elif command -v yum >/dev/null 2>&1; then
            yum install -y "${missing[@]}"
        else
            echo -e "${RED}无法安装依赖：不支持的系统${NC}"
            return 1
        fi
    fi
    return 0
}

# 检查服务状态
check_status() {
    local status=0
    local ip
    ip=$(get_ip) || ip="<ip>"
    
    # 检查后端 API 服务
    if pgrep -f "singdns.*:8080" > /dev/null; then
        echo -e "${GREEN}后端 API 服务正在运行${NC}"
        echo -e "${GREEN}后端本机访问: http://localhost:8080${NC}"
        echo -e "${GREEN}后端远程访问: http://${ip}:8080${NC}"
    else
        echo -e "${RED}后端 API 服务未运行${NC}"
        status=1
    fi

    # 检查前端服务
    if pgrep -f "nginx.*master" > /dev/null; then
        echo -e "${GREEN}前端服务正在运行${NC}"
        echo -e "${GREEN}前端本机访问: http://localhost:3000${NC}"
        echo -e "${GREEN}前端远程访问: http://${ip}:3000${NC}"
        echo -e "${GREEN}面板本机访问: http://localhost:9090${NC}"
        echo -e "${GREEN}面板远程访问: http://${ip}:9090${NC}"
    else
        echo -e "${RED}前端服务未运行${NC}"
        status=1
    fi

    return $status
}

# 启动后端服务
start_backend() {
    echo -e "${YELLOW}启动后端服务...${NC}"
    cd $INSTALL_DIR || exit 1
    
    # 检查是否已运行
    if pgrep -f "singdns.*:8080" > /dev/null; then
        echo -e "${YELLOW}后端服务已在运行${NC}"
        return 0
    fi
    
    # 启动服务
    nohup ./singdns serve > $LOG_DIR/backend.log 2>&1 &
    
    # 等待服务启动
    local timeout=30
    local counter=0
    while ! curl -s http://localhost:8080/ping > /dev/null && [ $counter -lt $timeout ]; do
        sleep 1
        ((counter++))
    done
    
    if [ $counter -eq $timeout ]; then
        echo -e "${RED}后端服务启动超时${NC}"
        return 1
    fi
    
    echo -e "${GREEN}后端服务启动成功${NC}"
    return 0
}

# 启动前端服务
start_frontend() {
    echo -e "${YELLOW}启动前端服务...${NC}"
    
    # 检查 nginx 是否安装
    if ! command -v nginx &> /dev/null; then
        echo -e "${YELLOW}安装 nginx...${NC}"
        if ! check_dependencies; then
            return 1
        fi
    fi
    
    # 创建必要的目录
    mkdir -p /run/nginx
    
    # 创建 nginx 配置
    cat > /etc/nginx/http.d/singdns.conf << EOF
# 前端服务
server {
    listen 3000;
    server_name localhost;
    
    # 前端
    location / {
        root $WEB_DIR;
        index index.html;
        try_files \$uri \$uri/ /index.html;
    }
    
    # API 代理
    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }
}

# Clash API 面板配置
server {
    listen 9090;
    server_name localhost;
    
    location / {
        root $BIN_DIR/web;
        index index.html;
        try_files \$uri \$uri/ /index.html;
    }
}
EOF
    
    # 测试配置
    if ! nginx -t; then
        echo -e "${RED}Nginx 配置测试失败${NC}"
        return 1
    fi
    
    # 启动或重载 nginx
    if [ -f /run/nginx/nginx.pid ]; then
        nginx -s reload
    else
        nginx
    fi
    
    # 检查是否成功启动
    sleep 2
    if ! curl -s http://localhost:3000 > /dev/null; then
        echo -e "${RED}前端服务启动失败${NC}"
        return 1
    fi
    
    echo -e "${GREEN}前端服务启动成功${NC}"
    return 0
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

# 启动服务
start_service() {
    echo -e "${YELLOW}正在启动 SingDNS...${NC}"
    
    # 检查依赖
    check_dependencies || exit 1
    
    # 开启 IP 转发
    echo -e "${YELLOW}开启 IP 转发...${NC}"
    if ! sysctl -w net.ipv4.ip_forward=1 > /dev/null 2>&1; then
        echo -e "${RED}开启 IP 转发失败${NC}"
        return 1
    fi
    
    # 备份当前配置
    backup_config
    
    # 启动服务
    start_backend || return 1
    start_frontend || return 1
    
    # 检查服务状态
    if check_status > /dev/null; then
        echo -e "${GREEN}SingDNS 启动成功！${NC}"
        check_status
    else
        echo -e "${RED}SingDNS 启动失败${NC}"
        return 1
    fi
}

# 停止服务
stop_service() {
    echo -e "${YELLOW}正在停止 SingDNS...${NC}"
    
    # 停止后端
    if pgrep -f "singdns" > /dev/null; then
        pkill -f "singdns"
        sleep 1
        if pgrep -f "singdns" > /dev/null; then
            pkill -9 -f "singdns"
        fi
    fi
    
    # 停止前端
    if [ -f /run/nginx/nginx.pid ]; then
        nginx -s stop
    fi
    
    # 清理防火墙规则
    echo -e "${YELLOW}清理防火墙规则...${NC}"
    iptables -F
    iptables -X
    iptables -t nat -F
    iptables -t nat -X
    iptables -t mangle -F
    iptables -t mangle -X
    
    sleep 1
    
    if ! check_status > /dev/null; then
        echo -e "${GREEN}SingDNS 服务已停止${NC}"
        return 0
    else
        echo -e "${RED}SingDNS 服务停止失败${NC}"
        return 1
    fi
}

# 重启服务
restart_service() {
    stop_service
    sleep 2
    start_service
}

# 查看日志
view_logs() {
    if [ "$1" = "frontend" ]; then
        if [ -f "/var/log/nginx/access.log" ]; then
            tail -f /var/log/nginx/access.log
        else
            echo -e "${RED}前端日志文件不存在${NC}"
        fi
    elif [ "$1" = "backend" ]; then
        if [ -f "$LOG_DIR/backend.log" ]; then
            tail -f $LOG_DIR/backend.log
        else
            echo -e "${RED}后端日志文件不存在${NC}"
        fi
    else
        echo -e "${YELLOW}=== 前端日志 ===${NC}"
        if [ -f "/var/log/nginx/access.log" ]; then
            tail -n 20 /var/log/nginx/access.log
        else
            echo -e "${RED}前端日志文件不存在${NC}"
        fi
        echo -e "\n${YELLOW}=== 后端日志 ===${NC}"
        if [ -f "$LOG_DIR/backend.log" ]; then
            tail -n 20 $LOG_DIR/backend.log
        else
            echo -e "${RED}后端日志文件不存在${NC}"
        fi
    fi
}

# 查看端口占用
check_ports() {
    echo -e "${BLUE}=== 端口占用情况 ===${NC}"
    echo -e "${YELLOW}前端端口 (3000)：${NC}"
    netstat -tunlp | grep ":3000 "
    echo -e "\n${YELLOW}后端 API 端口 (8080)：${NC}"
    netstat -tunlp | grep ":8080 "
    echo -e "\n${YELLOW}面板端口 (9090)：${NC}"
    netstat -tunlp | grep ":9090 "
    
    echo -e "\n${YELLOW}所有 SingDNS 相关进程：${NC}"
    ps aux | grep -E "singdns|nginx" | grep -v grep
}

# 显示菜单
show_menu() {
    echo -e "${BLUE}=== SingDNS 管理面板 ===${NC}"
    echo "1. 启动服务"
    echo "2. 停止服务"
    echo "3. 重启服务"
    echo "4. 查看状态"
    echo "5. 查看前端日志"
    echo "6. 查看后端日志"
    echo "7. 查看端口占用"
    echo "8. 备份配置"
    echo "9. 恢复配置"
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
        read -p "请选择操作 [0-9]: " choice
        
        case $choice in
            1)
                start_service
                ;;
            2)
                stop_service
                ;;
            3)
                restart_service
                ;;
            4)
                check_status
                ;;
            5)
                view_logs frontend
                ;;
            6)
                view_logs backend
                ;;
            7)
                check_ports
                ;;
            8)
                backup_config
                ;;
            9)
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
        restart)
            check_root
            restart_service
            ;;
        status)
            check_status
            ;;
        logs)
            if [ "$2" = "frontend" ] || [ "$2" = "backend" ]; then
                view_logs "$2"
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
            echo "用法: singdns {start|stop|restart|status|logs|backup|restore}"
            exit 1
            ;;
    esac
else
    main
fi