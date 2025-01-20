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

# 检查是否为 root 用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}错误: 必须使用 root 用户运行此脚本${NC}"
        exit 1
    fi
}

# 启动服务
start_service() {
    echo -e "${BLUE}启动 SingDNS 服务...${NC}"
    
    # 检查服务是否已经运行
    if pgrep -f "singdns.*serve" > /dev/null; then
        echo -e "${YELLOW}SingDNS 服务已经在运行${NC}"
        return
    fi
    
    # 启动后端服务
    systemctl start singdns
    
    # 启动前端服务
    cd "$WEB_DIR" || exit 1
    nohup busybox httpd -f -p 3000 -h "$WEB_DIR" > "$LOG_DIR/frontend.log" 2>&1 &
    
    # 检查服务状态
    sleep 2
    if pgrep -f "singdns.*serve" > /dev/null; then
        echo -e "${GREEN}SingDNS 服务已启动${NC}"
        # 获取本机 IP
        ip=$(hostname -I | awk '{print $1}')
        echo -e "${GREEN}管理界面: http://${ip}:3000${NC}"
    else
        echo -e "${RED}SingDNS 服务启动失败${NC}"
    fi
}

# 停止服务
stop_service() {
    echo -e "${BLUE}停止 SingDNS 服务...${NC}"
    
    # 停止后端服务
    systemctl stop singdns
    
    # 停止前端服务
    pkill -f "busybox httpd -f -p 3000" || true
    
    # 检查服务是否已停止
    if pgrep -f "singdns.*serve" > /dev/null; then
        echo -e "${RED}SingDNS 服务停止失败${NC}"
    else
        echo -e "${GREEN}SingDNS 服务已停止${NC}"
    fi
}

# 重启服务
restart_service() {
    echo -e "${BLUE}重启 SingDNS 服务...${NC}"
    stop_service
    sleep 2
    start_service
}

# 检查服务状态
check_status() {
    echo -e "${BLUE}检查 SingDNS 服务状态...${NC}"
    
    # 检查服务进程
    local backend_running=false
    local frontend_running=false
    
    if pgrep -f "singdns.*serve" > /dev/null; then
        backend_running=true
    fi
    
    if pgrep -f "busybox httpd -f -p 3000" > /dev/null; then
        frontend_running=true
    fi
    
    if [ "$backend_running" = true ] && [ "$frontend_running" = true ]; then
        echo -e "${GREEN}SingDNS 服务正在运行${NC}"
        # 获取本机 IP
        ip=$(hostname -I | awk '{print $1}')
        echo -e "${GREEN}管理界面: http://${ip}:3000${NC}"
        
        # 检查端口状态
        echo -e "\n${YELLOW}管理界面端口 (3000)：${NC}"
        netstat -tunlp | grep ":3000 "
    else
        [ "$backend_running" = false ] && echo -e "${RED}后端服务未运行${NC}"
        [ "$frontend_running" = false ] && echo -e "${RED}前端服务未运行${NC}"
    fi
}

# 显示帮助信息
show_help() {
    echo -e "${BLUE}SingDNS 管理脚本使用说明${NC}"
    echo -e "用法: singdns <命令>"
    echo -e "\n命令列表:"
    echo -e "  start    启动服务"
    echo -e "  stop     停止服务"
    echo -e "  restart  重启服务"
    echo -e "  status   查看服务状态"
    echo -e "  help     显示帮助信息"
}

# 主函数
main() {
    # 检查是否为 root 用户
    check_root
    
    # 处理命令行参数
    case "$1" in
        start)
            start_service
            ;;
        stop)
            stop_service
            ;;
        restart)
            restart_service
            ;;
        status)
            check_status
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo -e "${RED}错误: 无效的命令${NC}"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@" 