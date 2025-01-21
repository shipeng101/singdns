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

# 检查后端服务状态
check_backend() {
    # 检查8080端口是否被占用
    if netstat -tuln | grep -q ":8080 "; then
        # 获取占用8080端口的进程PID
        local pid=$(lsof -t -i:8080)
        if [ -n "$pid" ]; then
            echo "$pid" > "$PID_FILE"
            return 0
        fi
    fi
    
    # 如果端口未被占用，清理 PID 文件
    rm -f "$PID_FILE"
    return 1
}

# 检查前端服务状态
check_frontend() {
    # 检查进程
    local pid=$(ps aux | grep "[b]usybox.*httpd.*$FRONTEND_PORT" | grep -v "grep" | awk '{print $2}')
    if [ -n "$pid" ]; then
        echo "$pid" > "$FRONTEND_PID_FILE"
        return 0
    fi
    
    # 如果进程不存在，清理 PID 文件
    rm -f "$FRONTEND_PID_FILE"
    return 1
}

# 启动后端服务
start_backend() {
    if check_backend; then
        echo -e "${YELLOW}后端服务已经在运行中${NC}"
        return 0
    fi
    
    echo -e "${YELLOW}正在启动后端服务...${NC}"
    cd $INSTALL_DIR
    ./singdns > "$LOG_DIR/backend.log" 2>&1 &
    local pid=$!
    echo $pid > $PID_FILE
    
    sleep 2
    if check_backend; then
        echo -e "${GREEN}后端服务启动成功${NC}"
    else
        echo -e "${RED}后端服务启动失败，请检查日志${NC}"
        tail -n 10 "$LOG_DIR/backend.log"
        return 1
    fi
}

# 启动前端服务
start_frontend() {
    if check_frontend; then
        echo -e "${YELLOW}前端服务已经在运行中${NC}"
        return 0
    fi
    
    echo -e "${YELLOW}正在启动前端服务...${NC}"
    cd $INSTALL_DIR/web
    busybox httpd -f -p $FRONTEND_PORT -h $INSTALL_DIR/web > "$LOG_DIR/frontend.log" 2>&1 &
    local pid=$!
    echo $pid > $FRONTEND_PID_FILE
    
    sleep 2
    if check_frontend; then
        echo -e "${GREEN}前端服务启动成功 - 访问 http://localhost:${FRONTEND_PORT}${NC}"
    else
        echo -e "${RED}前端服务启动失败，请检查日志${NC}"
        tail -n 10 "$LOG_DIR/frontend.log"
        return 1
    fi
}

# 停止后端服务
stop_backend() {
    if ! check_backend; then
        echo -e "${YELLOW}后端服务未在运行${NC}"
        return 0
    fi
    
    echo -e "${YELLOW}正在停止后端服务...${NC}"
    local pid=$(lsof -t -i:8080)
    if [ -n "$pid" ]; then
        kill $pid
        rm -f $PID_FILE
    fi
    
    sleep 2
    if ! check_backend; then
        echo -e "${GREEN}后端服务已停止${NC}"
    else
        echo -e "${RED}后端服务停止失败，尝试强制终止${NC}"
        local pid=$(lsof -t -i:8080)
        if [ -n "$pid" ]; then
            kill -9 $pid
        fi
        rm -f $PID_FILE
        sleep 1
        if ! check_backend; then
            echo -e "${GREEN}后端服务已强制停止${NC}"
        else
            echo -e "${RED}后端服务无法停止${NC}"
            return 1
        fi
    fi
}

# 停止前端服务
stop_frontend() {
    if ! check_frontend; then
        echo -e "${YELLOW}前端服务未在运行${NC}"
        return 0
    fi
    
    echo -e "${YELLOW}正在停止前端服务...${NC}"
    pkill -f "busybox.*httpd.*$FRONTEND_PORT"
    rm -f $FRONTEND_PID_FILE
    
    sleep 2
    if ! check_frontend; then
        echo -e "${GREEN}前端服务已停止${NC}"
    else
        echo -e "${RED}前端服务停止失败，尝试强制终止${NC}"
        pkill -9 -f "busybox.*httpd.*$FRONTEND_PORT"
        rm -f $FRONTEND_PID_FILE
        sleep 1
        if ! check_frontend; then
            echo -e "${GREEN}前端服务已强制停止${NC}"
        else
            echo -e "${RED}前端服务无法停止${NC}"
            return 1
        fi
    fi
}

# 启动所有服务
start() {
    check_root
    start_backend || return 1
    start_frontend || return 1
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
    
    if check_backend; then
        local pid=$(cat "$PID_FILE" 2>/dev/null)
        echo -e "后端服务: ${GREEN}运行中${NC} (PID: $pid)"
    else
        echo -e "后端服务: ${RED}未运行${NC}"
    fi
    
    if check_frontend; then
        local pid=$(cat "$FRONTEND_PID_FILE" 2>/dev/null)
        echo -e "前端服务: ${GREEN}运行中${NC} (http://localhost:${FRONTEND_PORT}, PID: $pid)"
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