#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查并修复文件格式
fix_file_format() {
    local file=$1
    if [ -f "$file" ]; then
        # 检查文件是否是Windows格式
        if file "$file" | grep -q "CRLF"; then
            echo -e "${YELLOW}修复文件格式: $file${NC}"
            sed -i 's/\r$//' "$file" 2>/dev/null
        fi
    fi
}

# 安装和初始化
install_service() {
    echo -e "${YELLOW}开始安装服务...${NC}"
    
    # 创建必要的目录
    mkdir -p bin/web
    mkdir -p configs/sing-box/rules
    mkdir -p configs/mosdns
    mkdir -p run
    mkdir -p logs
    
    # 修复可执行文件格式
    fix_file_format "bin/sing-box"
    fix_file_format "bin/mosdns"
    fix_file_format "bin/singdns"
    
    # 修复配置文件格式
    fix_file_format "configs/sing-box/config.json"
    fix_file_format "configs/mosdns/config.yaml"
    
    # 设置文件权限
    chmod +x bin/sing-box 2>/dev/null || echo -e "${YELLOW}警告: sing-box 可执行文件不存在${NC}"
    chmod +x bin/mosdns 2>/dev/null || echo -e "${YELLOW}警告: mosdns 可执行文件不存在${NC}"
    chmod +x bin/singdns 2>/dev/null || echo -e "${YELLOW}警告: singdns 可执行文件不存在${NC}"
    
    # 设置配置文件权限
    chmod 644 configs/sing-box/config.json 2>/dev/null || echo -e "${YELLOW}警告: sing-box 配置文件不存在${NC}"
    chmod 644 configs/mosdns/config.yaml 2>/dev/null || echo -e "${YELLOW}警告: mosdns 配置文件不存在${NC}"
    
    # 设置规则文件权限
    chmod 644 configs/sing-box/rules/*.srs 2>/dev/null
    
    # 设置日志目录权限
    chmod 755 logs
    chmod 755 run
    
    echo -e "${GREEN}安装完成！${NC}"
    echo -e "请确保以下文件存在并且可执行："
    echo -e "- bin/sing-box"
    echo -e "- bin/mosdns"
    echo -e "- bin/singdns"
    echo -e "\n请确保以下配置文件存在："
    echo -e "- configs/sing-box/config.json"
    echo -e "- configs/mosdns/config.yaml"
}

# 检查服务状态
check_service() {
    local name=$1
    local pid_file="run/${name}.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p "$pid" > /dev/null 2>&1; then
            echo -e "${GREEN}运行中${NC} (PID: $pid)"
            return 0
        fi
    fi
    echo -e "${RED}已停止${NC}"
    return 1
}

# 启动服务前检查文件格式
pre_start_check() {
    local name=$1
    local bin_file="bin/$name"
    
    if [ -f "$bin_file" ]; then
        fix_file_format "$bin_file"
    fi
}

# 启动服务
start_service() {
    local name=$1
    local command=$2
    local pid_file="run/${name}.pid"
    
    # 启动前检查文件格式
    pre_start_check "$name"
    
    if check_service "$name" > /dev/null; then
        echo -e "${YELLOW}$name 已经在运行中${NC}"
        return
    fi
    
    # 确保目录存在
    mkdir -p run
    mkdir -p logs
    
    # 启动服务
    eval "$command" > "logs/${name}.log" 2>&1 &
    echo $! > "$pid_file"
    echo -e "${GREEN}已启动 $name${NC}"
}

# 停止服务
stop_service() {
    local name=$1
    local pid_file="run/${name}.pid"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p "$pid" > /dev/null 2>&1; then
            kill -15 "$pid"
            sleep 1
            if ps -p "$pid" > /dev/null 2>&1; then
                kill -9 "$pid"
            fi
            rm "$pid_file"
            echo -e "${GREEN}已停止 $name${NC}"
            return
        fi
        rm "$pid_file"
    fi
    echo -e "${YELLOW}$name 未在运行${NC}"
}

# 重启服务
restart_service() {
    local name=$1
    local command=$2
    stop_service "$name"
    sleep 2
    start_service "$name" "$command"
}

# 显示服务状态
show_status() {
    echo "服务状态:"
    echo "---------------"
    echo -n "sing-box: "
    check_service "sing-box"
    echo -n "mosdns:  "
    check_service "mosdns"
    echo -n "singdns: "
    check_service "singdns"
    echo "---------------"
}

# 显示日志
show_logs() {
    local name=$1
    local lines=${2:-50}
    if [ -f "logs/${name}.log" ]; then
        tail -n "$lines" "logs/${name}.log"
    else
        echo -e "${RED}未找到 $name 的日志文件${NC}"
    fi
}

# 启动所有服务
start_all() {
    start_service "sing-box" "./bin/sing-box run -c configs/sing-box/config.json"
    start_service "mosdns" "./bin/mosdns start -d configs/mosdns"
    start_service "singdns" "./bin/singdns"
}

# 停止所有服务
stop_all() {
    stop_service "singdns"
    stop_service "mosdns"
    stop_service "sing-box"
}

# 重启所有服务
restart_all() {
    stop_all
    sleep 2
    start_all
}

# 显示菜单
show_menu() {
    clear
    echo -e "\n${YELLOW}SingDNS 管理菜单${NC}"
    echo "1. 启动所有服务"
    echo "2. 停止所有服务"
    echo "3. 重启所有服务"
    echo "4. 查看服务状态"
    echo "5. 查看 sing-box 日志"
    echo "6. 查看 mosdns 日志"
    echo "7. 查看 singdns 日志"
    echo "8. 启动 sing-box"
    echo "9. 启动 mosdns"
    echo "10. 启动 singdns"
    echo "11. 停止 sing-box"
    echo "12. 停止 mosdns"
    echo "13. 停止 singdns"
    echo "14. 重启 sing-box"
    echo "15. 重启 mosdns"
    echo "16. 重启 singdns"
    echo "17. 安装服务"
    echo "0. 退出"
    echo -n "请输入选项: "
}

# 主循环
main_loop() {
    while true; do
        show_menu
        read choice
        case $choice in
            1) start_all ;;
            2) stop_all ;;
            3) restart_all ;;
            4) show_status ;;
            5) show_logs "sing-box" ;;
            6) show_logs "mosdns" ;;
            7) show_logs "singdns" ;;
            8) start_service "sing-box" "./bin/sing-box run -c configs/sing-box/config.json" ;;
            9) start_service "mosdns" "./bin/mosdns start -d configs/mosdns" ;;
            10) start_service "singdns" "./bin/singdns" ;;
            11) stop_service "sing-box" ;;
            12) stop_service "mosdns" ;;
            13) stop_service "singdns" ;;
            14) restart_service "sing-box" "./bin/sing-box run -c configs/sing-box/config.json" ;;
            15) restart_service "mosdns" "./bin/mosdns start -d configs/mosdns" ;;
            16) restart_service "singdns" "./bin/singdns" ;;
            17) install_service ;;
            0) exit 0 ;;
            *) echo -e "${RED}无效的选项${NC}" ;;
        esac
        echo -e "\n按回车键继续..."
        read
    done
}

# 如果没有参数，启动所有服务并显示状态
if [ $# -eq 0 ]; then
    install_service
    start_all
    show_status
    main_loop
    exit 0
fi

# 处理命令行参数
case "$1" in
    "install")
        install_service
        ;;
    "start")
        start_all
        ;;
    "stop")
        stop_all
        ;;
    "restart")
        restart_all
        ;;
    "status")
        show_status
        ;;
    *)
        echo "用法: $0 {install|start|stop|restart|status}"
        exit 1
        ;;
esac 