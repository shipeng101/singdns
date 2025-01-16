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
WEB_DIR="$INSTALL_DIR/web"

# 检查是否为root用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}错误：请使用root用户运行此脚本${NC}"
        exit 1
    fi
}

# 获取本机IP地址
get_ip() {
    ip addr | grep 'state UP' -A2 | grep 'inet ' | awk '{print $2}' | cut -f1 -d'/' | head -n 1
}

# 检查服务状态
check_status() {
    local status=0
    # 检查后端 API 服务
    if pgrep -f "singdns.*:8080" > /dev/null; then
        echo -e "${GREEN}后端 API 服务正在运行${NC}"
        local_ip=$(get_ip)
        if [ -n "$local_ip" ]; then
            echo -e "${GREEN}后端本机访问: http://localhost:8080${NC}"
            echo -e "${GREEN}后端远程访问: http://${local_ip}:8080${NC}"
        else
            echo -e "${GREEN}后端访问地址: http://<ip>:8080${NC}"
        fi
    else
        echo -e "${RED}后端 API 服务未运行${NC}"
        status=1
    fi

    # 检查前端服务
    if pgrep -f "nginx.*master" > /dev/null; then
        echo -e "${GREEN}前端服务正在运行${NC}"
        local_ip=$(get_ip)
        if [ -n "$local_ip" ]; then
            echo -e "${GREEN}前端本机访问: http://localhost:3000${NC}"
            echo -e "${GREEN}前端远程访问: http://${local_ip}:3000${NC}"
        else
            echo -e "${GREEN}前端访问地址: http://<ip>:3000${NC}"
        fi
    else
        echo -e "${RED}前端服务未运行${NC}"
        status=1
    fi

    return $status
}

# 启动后端服务
start_backend() {
    echo -e "${YELLOW}启动后端服务...${NC}"
    cd $INSTALL_DIR
    nohup ./singdns serve > $LOG_DIR/backend.log 2>&1 &
    sleep 2
}

# 启动前端服务
start_frontend() {
    echo -e "${YELLOW}启动前端服务...${NC}"
    
    # 检查 nginx 是否安装
    if ! command -v nginx &> /dev/null; then
        echo -e "${YELLOW}安装 nginx...${NC}"
        apk add --no-cache nginx
    fi
    
    # 创建必要的目录
    mkdir -p /run/nginx
    
    # 创建 nginx 配置
    cat > /etc/nginx/http.d/singdns.conf << EOF
server {
    listen 3000;
    server_name localhost;
    
    location / {
        root $WEB_DIR;
        index index.html;
        try_files \$uri \$uri/ /index.html;
    }
}
EOF
    
    # 启动 nginx
    if [ -f /run/nginx/nginx.pid ]; then
        nginx -s reload
    else
        nginx
    fi
    sleep 2
}

# 启动服务
start_service() {
    echo -e "${YELLOW}正在启动 SingDNS...${NC}"
    
    # 启动后端
    start_backend
    
    # 启动前端
    start_frontend
    
    if check_status > /dev/null; then
        echo -e "${GREEN}SingDNS 启动成功！${NC}"
        local_ip=$(get_ip)
        if [ -n "$local_ip" ]; then
            echo -e "${GREEN}前端本机访问: http://localhost:3000${NC}"
            echo -e "${GREEN}前端远程访问: http://${local_ip}:3000${NC}"
            echo -e "${GREEN}后端本机访问: http://localhost:8080${NC}"
            echo -e "${GREEN}后端远程访问: http://${local_ip}:8080${NC}"
        else
            echo -e "${GREEN}前端访问地址: http://<ip>:3000${NC}"
            echo -e "${GREEN}后端访问地址: http://<ip>:8080${NC}"
        fi
    else
        echo -e "${RED}SingDNS 启动失败，请查看日志${NC}"
        echo -e "${YELLOW}使用 'singdns logs' 查看详细日志${NC}"
    fi
}

# 停止服务
stop_service() {
    echo -e "${YELLOW}正在停止 SingDNS...${NC}"
    
    # 停止后端
    pkill -f "singdns"
    
    # 停止前端
    nginx -s stop
    
    sleep 1
    
    if ! check_status > /dev/null; then
        echo -e "${GREEN}SingDNS 服务已停止${NC}"
    else
        echo -e "${RED}SingDNS 服务停止失败${NC}"
        echo -e "${YELLOW}尝试强制停止...${NC}"
        pkill -9 -f "singdns"
        pkill -9 -f "nginx"
    fi
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
    
    echo -e "\n${YELLOW}所有 SingDNS 相关进程：${NC}"
    ps aux | grep -E "singdns|nginx" | grep -v grep
}

# 显示菜单
show_menu() {
    echo -e "${BLUE}=== SingDNS 管理面板 ===${NC}"
    echo "1. 启动服务"
    echo "2. 停止服务"
    echo "3. 查看状态"
    echo "4. 查看前端日志"
    echo "5. 查看后端日志"
    echo "6. 查看端口"
    echo "0. 退出"
    echo -e "${BLUE}=====================${NC}"
}

# 主程序
main() {
    mkdir -p $LOG_DIR
    
    while true; do
        show_menu
        read -p "请选择操作 [0-6]: " choice
        
        case $choice in
            1)
                check_root
                start_service
                ;;
            2)
                check_root
                stop_service
                ;;
            3)
                check_status
                ;;
            4)
                view_logs frontend
                ;;
            5)
                view_logs backend
                ;;
            6)
                check_ports
                ;;
            0)
                echo -e "${GREEN}感谢使用 SingDNS 管理面板！${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}无效的选项，请重新选择${NC}"
                ;;
        esac
        
        echo
        read -p "按回车键继续..."
    done
}

# 处理命令行参数
if [ $# -gt 0 ]; then
    case $1 in
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
            if [ "$2" = "frontend" ] || [ "$2" = "backend" ]; then
                view_logs $2
            else
                view_logs
            fi
            ;;
        ports)
            check_ports
            ;;
        *)
            echo -e "${RED}无效的命令${NC}"
            echo "用法: singdns {start|stop|status|logs|ports}"
            exit 1
            ;;
    esac
else
    main
fi