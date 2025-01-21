#!/bin/bash

# 设置变量
INSTALL_DIR="/usr/local/singdns"
LOG_DIR="/var/log/singdns"
PID_FILE="/var/run/singdns.pid"
FRONTEND_PORT=3000
FRONTEND_PID_FILE="/var/run/singdns-frontend.pid"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 检查是否为 root 用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo -e "${RED}错误: 请使用 root 用户运行此脚本${NC}"
        exit 1
    fi
}

# 显示菜单
show_menu() {
    clear
    echo -e "${GREEN}=== SingDNS 管理面板 ===${NC}"
    echo "1. 启动服务"
    echo "2. 停止服务"
    echo "3. 重启服务"
    echo "4. 查看状态"
    echo "5. 查看后端日志"
    echo "6. 查看前端日志"
    echo "7. 查看所有日志"
    echo "0. 退出"
    echo
    echo -n "请输入选项 [0-7]: "
}

# 启动后端服务
start_backend() {
    if pgrep -x "singdns" > /dev/null; then
        echo -e "${YELLOW}后端服务已经在运行中${NC}"
        return 0
    fi
    
    echo -e "${YELLOW}正在启动后端服务...${NC}"
    cd $INSTALL_DIR
    nohup ./singdns > "$LOG_DIR/backend.log" 2>&1 &
    echo $! > $PID_FILE
    
    sleep 2
    if pgrep -x "singdns" > /dev/null; then
        echo -e "${GREEN}后端服务启动成功${NC}"
    else
        echo -e "${RED}后端服务启动失败，请检查日志${NC}"
        exit 1
    fi
}

# 启动前端服务
start_frontend() {
    if pgrep -f "busybox httpd.*$FRONTEND_PORT" > /dev/null; then
        echo -e "${YELLOW}前端服务已经在运行中${NC}"
        return 0
    fi
    
    echo -e "${YELLOW}正在启动前端服务...${NC}"
    cd $INSTALL_DIR/web
    nohup busybox httpd -f -p $FRONTEND_PORT -h $INSTALL_DIR/web > "$LOG_DIR/frontend.log" 2>&1 &
    echo $! > $FRONTEND_PID_FILE
    
    sleep 2
    if pgrep -f "busybox httpd.*$FRONTEND_PORT" > /dev/null; then
        echo -e "${GREEN}前端服务启动成功 - 访问 http://localhost:${FRONTEND_PORT}${NC}"
    else
        echo -e "${RED}前端服务启动失败，请检查日志${NC}"
        exit 1
    fi
}

# 停止后端服务
stop_backend() {
    if ! pgrep -x "singdns" > /dev/null; then
        echo -e "${YELLOW}后端服务未在运行${NC}"
        return 0
    fi
    
    echo -e "${YELLOW}正在停止后端服务...${NC}"
    if [ -f $PID_FILE ]; then
        kill $(cat $PID_FILE)
        rm -f $PID_FILE
    else
        pkill singdns
    fi
    
    sleep 2
    if ! pgrep -x "singdns" > /dev/null; then
        echo -e "${GREEN}后端服务已停止${NC}"
    else
        echo -e "${RED}后端服务停止失败${NC}"
        exit 1
    fi
}

# 停止前端服务
stop_frontend() {
    if ! pgrep -f "busybox httpd.*$FRONTEND_PORT" > /dev/null; then
        echo -e "${YELLOW}前端服务未在运行${NC}"
        return 0
    fi
    
    echo -e "${YELLOW}正在停止前端服务...${NC}"
    if [ -f $FRONTEND_PID_FILE ]; then
        kill $(cat $FRONTEND_PID_FILE)
        rm -f $FRONTEND_PID_FILE
    else
        pkill -f "busybox httpd.*$FRONTEND_PORT"
    fi
    
    sleep 2
    if ! pgrep -f "busybox httpd.*$FRONTEND_PORT" > /dev/null; then
        echo -e "${GREEN}前端服务已停止${NC}"
    else
        echo -e "${RED}前端服务停止失败${NC}"
        exit 1
    fi
}

# 启动所有服务
start() {
    check_root
    start_backend
    start_frontend
}

# 停止所有服务
stop() {
    check_root
    stop_frontend
    stop_backend
}

# 重启所有服务
restart() {
    stop
    sleep 2
    start
}

# 查看所有服务状态
status() {
    echo -e "${BLUE}服务状态:${NC}"
    echo -e "------------------------"
    
    if pgrep -x "singdns" > /dev/null; then
        echo -e "后端服务: ${GREEN}运行中${NC}"
    else
        echo -e "后端服务: ${RED}未运行${NC}"
    fi
    
    if pgrep -f "busybox httpd.*$FRONTEND_PORT" > /dev/null; then
        echo -e "前端服务: ${GREEN}运行中${NC} (http://localhost:${FRONTEND_PORT})"
    else
        echo -e "前端服务: ${RED}未运行${NC}"
    fi
}

# 查看日志
logs() {
    case "$1" in
        backend)
            echo -e "${BLUE}查看后端日志:${NC}"
            tail -f "$LOG_DIR/backend.log"
            ;;
        frontend)
            echo -e "${BLUE}查看前端日志:${NC}"
            tail -f "$LOG_DIR/frontend.log"
            ;;
        *)
            echo -e "${BLUE}查看所有日志:${NC}"
            echo -e "------------------------"
            echo -e "${YELLOW}后端日志:${NC}"
            tail -n 50 "$LOG_DIR/backend.log"
            echo -e "\n${YELLOW}前端日志:${NC}"
            tail -n 50 "$LOG_DIR/frontend.log"
            ;;
    esac
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
                start
                ;;
            2)
                stop
                ;;
            3)
                restart
                ;;
            4)
                status
                ;;
            5)
                logs "backend"
                ;;
            6)
                logs "frontend"
                ;;
            7)
                logs
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