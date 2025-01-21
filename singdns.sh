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
DATA_DIR="$INSTALL_DIR/data"

# 配置变量
SINGDNS_PID="/var/run/singdns.pid"
SINGBOX_PID="/var/run/singbox.pid"
SINGDNS_LOG="$LOG_DIR/singdns.log"
SINGBOX_LOG="$LOG_DIR/singbox.log"

# 检查是否为root用户
check_root() {
    if [ "$(id -u)" != "0" ]; then
        echo "${RED}错误：请使用root用户运行此脚本${NC}"
        return 1
    fi
    return 0
}

# 获取本机IP地址
get_ip() {
    local ip=""
    # 尝试多种方法获取IP
    if command -v ip >/dev/null 2>&1; then
        ip=$(ip addr | grep 'state UP' -A2 | grep 'inet ' | awk '{print $2}' | cut -f1 -d'/' | head -n 1)
    elif command -v ifconfig >/dev/null 2>&1; then
        ip=$(ifconfig | grep 'inet ' | grep -v '127.0.0.1' | awk '{print $2}' | head -n 1)
    fi
    
    if [ -z "$ip" ]; then
        # 尝试通过网络连接获取
        ip=$(curl -s ifconfig.me || wget -qO- ifconfig.me)
    fi
    
    echo "$ip"
}

# 检查进程是否运行
check_process() {
    if [ -f "$1" ]; then
        if kill -0 "$(cat "$1")" >/dev/null 2>&1; then
            return 0
        fi
    fi
    return 1
}

# 检查服务状态
check_status() {
    local status=0
    
    # 检查必要目录
    if [ ! -d "$INSTALL_DIR" ] || [ ! -d "$CONFIG_DIR" ]; then
        echo -e "${RED}错误：安装目录不完整${NC}"
        return 1
    fi
    
    # 检查必要文件
    if [ ! -f "$INSTALL_DIR/singdns" ] || [ ! -f "$BIN_DIR/sing-box" ]; then
        echo -e "${RED}错误：关键程序文件缺失${NC}"
        return 1
    fi
    
    # 检查配置文件
    if [ ! -f "$CONFIG_DIR/config.json" ]; then
        echo -e "${RED}错误：配置文件不存在${NC}"
        return 1
    fi
    
    # 检查后端服务
    if check_process "$SINGDNS_PID"; then
        echo -e "${GREEN}SingDNS 后端 - 运行中 (PID: $(cat "$SINGDNS_PID"))${NC}"
        status=1
    else
        echo -e "${RED}SingDNS 后端 - 未运行${NC}"
        status=1
    fi
    
    # 检查 sing-box 服务
    if check_process "$SINGBOX_PID"; then
        echo -e "${GREEN}Sing-Box - 运行中 (PID: $(cat "$SINGBOX_PID"))${NC}"
        status=1
    else
        echo -e "${RED}Sing-Box - 未运行${NC}"
        status=1
    fi

    # 检查前端服务
    if pgrep -f "busybox httpd -f -p 3000" > /dev/null; then
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

    # 检查系统资源
    check_system_resources

    return $status
}

# 检查系统资源
check_system_resources() {
    echo -e "\n${BLUE}系统资源使用情况：${NC}"
    
    # CPU 使用率
    if command -v top >/dev/null 2>&1; then
        echo -e "${YELLOW}CPU 使用率：${NC}"
        top -bn1 | grep "Cpu(s)" | awk '{print $2 + $4}' | awk '{print $1"%"}'
    fi
    
    # 内存使用情况
    if command -v free >/dev/null 2>&1; then
        echo -e "${YELLOW}内存使用情况：${NC}"
        free -h | grep "Mem:"
    fi
    
    # 磁盘使用情况
    echo -e "${YELLOW}磁盘使用情况：${NC}"
    df -h "$INSTALL_DIR"
}

# 启动SingDNS后端
start_backend() {
    echo "${BLUE}启动 SingDNS 后端...${NC}"
    
    if check_process "$SINGDNS_PID"; then
        echo "${YELLOW}SingDNS 后端已在运行${NC}"
        return 0
    fi
    
    # 启动后端服务
    nohup "$INSTALL_DIR/singdns" -c "$INSTALL_DIR/configs/sing-box/config.json" > "$SINGDNS_LOG" 2>&1 &
    echo $! > "$SINGDNS_PID"
    
    # 等待服务启动
    sleep 2
    if check_process "$SINGDNS_PID"; then
        echo "${GREEN}SingDNS 后端启动成功${NC}"
        return 0
    else
        echo "${RED}SingDNS 后端启动失败${NC}"
        return 1
    fi
}

# 启动Sing-Box
start_singbox() {
    echo "${BLUE}启动 Sing-Box...${NC}"
    
    if check_process "$SINGBOX_PID"; then
        echo "${YELLOW}Sing-Box 已在运行${NC}"
        return 0
    fi
    
    # 启动Sing-Box
    nohup "$INSTALL_DIR/bin/sing-box" run -c "$INSTALL_DIR/configs/sing-box/config.json" > "$SINGBOX_LOG" 2>&1 &
    echo $! > "$SINGBOX_PID"
    
    # 等待服务启动
    sleep 2
    if check_process "$SINGBOX_PID"; then
        echo "${GREEN}Sing-Box 启动成功${NC}"
        return 0
    else
        echo "${RED}Sing-Box 启动失败${NC}"
        return 1
    fi
}

# 启动前端服务
start_frontend() {
    echo -e "${YELLOW}启动前端服务...${NC}"
    
    # 检查是否已经运行
    if pgrep -f "busybox httpd -f -p 3000" > /dev/null; then
        echo -e "${YELLOW}前端服务已在运行${NC}"
        return 0
    fi
    
    # 检查前端目录
    if [ ! -d "$WEB_DIR" ]; then
        echo -e "${RED}错误：前端目录不存在${NC}"
        return 1
    fi
    
    # 检查 busybox
    if ! command -v busybox >/dev/null 2>&1; then
        echo -e "${RED}错误：未找到 busybox${NC}"
        return 1
    fi
    
    # 启动 busybox httpd
    cd "$WEB_DIR" || exit 1
    nohup busybox httpd -f -p 3000 -h "$WEB_DIR" > "$LOG_DIR/frontend.log" 2>&1 &
    
    # 等待服务启动
    local count=0
    while ! pgrep -f "busybox httpd -f -p 3000" > /dev/null && [ $count -lt 5 ]; do
        sleep 1
        count=$((count + 1))
    done
    
    if pgrep -f "busybox httpd -f -p 3000" > /dev/null; then
        echo -e "${GREEN}前端服务启动成功${NC}"
        return 0
    else
        echo -e "${RED}前端服务启动失败${NC}"
        return 1
    fi
}

# 启动所有服务
start() {
    check_root || return 1
    
    # 创建日志目录
    mkdir -p "$LOG_DIR"
    chmod 755 "$LOG_DIR"
    
    # 启动服务
    start_backend || return 1
    start_singbox || return 1
    start_frontend || return 1
    
    echo "${GREEN}所有服务启动完成${NC}"
    return 0
}

# 停止服务
stop() {
    check_root || return 1
    
    stop_service "$SINGDNS_PID" "SingDNS后端"
    stop_service "$SINGBOX_PID" "Sing-Box"
    
    echo "${GREEN}所有服务已停止${NC}"
    return 0
}

# 重启服务
restart() {
    check_root || return 1
    
    stop
    sleep 2
    start
}

# 停止前端服务
stop_frontend() {
    local pid
    pid=$(pgrep -f "busybox httpd -f -p 3000")
    if [ -n "$pid" ]; then
        kill "$pid"
        sleep 1
        if kill -0 "$pid" 2>/dev/null; then
            kill -9 "$pid"
        fi
    fi
}

# 停止服务
stop_service() {
    local pid_file="$1"
    local service_name="$2"
    
    if [ -f "$pid_file" ]; then
        echo "${BLUE}停止 $service_name...${NC}"
        if kill "$(cat "$pid_file")" 2>/dev/null; then
            rm -f "$pid_file"
            echo "${GREEN}$service_name 已停止${NC}"
        else
            echo "${RED}停止 $service_name 失败${NC}"
            return 1
        fi
    else
        echo "${YELLOW}$service_name 未运行${NC}"
    fi
    return 0
}

# 查看日志
logs() {
    local log_file="$1"
    if [ -f "$log_file" ]; then
        tail -f "$log_file" | cat
    else
        echo "${RED}日志文件不存在: $log_file${NC}"
        return 1
    fi
}

# 查看端口占用
check_ports() {
    echo -e "${BLUE}=== 端口占用情况 ===${NC}"
    
    # 检查端口是否被占用的函数
    check_port() {
        local port=$1
        local name=$2
        echo -e "${YELLOW}${name} (${port})：${NC}"
        if command -v lsof >/dev/null 2>&1; then
            lsof -i ":$port" || echo "端口未被占用"
        elif command -v netstat >/dev/null 2>&1; then
            netstat -tunlp | grep ":$port " || echo "端口未被占用"
        else
            echo "无法检查端口状态（需要 lsof 或 netstat）"
        fi
    }
    
    check_port 3000 "前端端口"
    echo
    check_port 8080 "后端 API 端口"
    
    echo -e "\n${YELLOW}所有 SingDNS 相关进程：${NC}"
    ps aux | grep -E "singdns|busybox httpd" | grep -v grep || echo "没有相关进程运行"
}

# 备份配置
backup_config() {
    local backup_dir="$INSTALL_DIR/backups"
    local backup_file="singdns_backup_$(date +%Y%m%d_%H%M%S).tar.gz"
    
    echo -e "${YELLOW}开始备份配置...${NC}"
    
    # 创建备份目录
    mkdir -p "$backup_dir"
    
    # 创建临时目录
    local temp_dir="/tmp/singdns_backup_temp"
    mkdir -p "$temp_dir"
    
    # 复制配置文件
    cp -r "$CONFIG_DIR" "$temp_dir/"
    [ -d "$DATA_DIR" ] && cp -r "$DATA_DIR" "$temp_dir/"
    
    # 创建备份文件
    tar -czf "$backup_dir/$backup_file" -C "$temp_dir" .
    
    # 清理临时目录
    rm -rf "$temp_dir"
    
    if [ -f "$backup_dir/$backup_file" ]; then
        echo -e "${GREEN}备份成功：$backup_dir/$backup_file${NC}"
        echo -e "${YELLOW}备份文件大小：$(du -h "$backup_dir/$backup_file" | cut -f1)${NC}"
    else
        echo -e "${RED}备份失败${NC}"
        return 1
    fi
}

# 恢复配置
restore_config() {
    local backup_dir="$INSTALL_DIR/backups"
    
    if [ ! -d "$backup_dir" ]; then
        echo -e "${RED}错误：备份目录不存在${NC}"
        return 1
    fi
    
    # 列出所有备份文件
    echo -e "${YELLOW}可用的备份文件：${NC}"
    local backup_files=("$backup_dir"/*.tar.gz)
    if [ ${#backup_files[@]} -eq 0 ]; then
        echo -e "${RED}没有找到备份文件${NC}"
        return 1
    fi
    
    local i=1
    for file in "${backup_files[@]}"; do
        echo "$i) $(basename "$file") ($(du -h "$file" | cut -f1))"
        i=$((i + 1))
    done
    
    echo
    read -p "请选择要恢复的备份文件编号 [1-${#backup_files[@]}]: " choice
    
    if [[ ! "$choice" =~ ^[0-9]+$ ]] || [ "$choice" -lt 1 ] || [ "$choice" -gt ${#backup_files[@]} ]; then
        echo -e "${RED}无效的选择${NC}"
        return 1
    fi
    
    local selected_file="${backup_files[$((choice-1))]}"
    
    echo -e "${YELLOW}正在恢复配置...${NC}"
    
    # 停止服务
    stop
    
    # 创建临时目录
    local temp_dir="/tmp/singdns_restore_temp"
    mkdir -p "$temp_dir"
    
    # 解压备份文件
    tar -xzf "$selected_file" -C "$temp_dir"
    
    # 恢复配置文件
    if [ -d "$temp_dir/sing-box" ]; then
        rm -rf "$CONFIG_DIR"
        mv "$temp_dir/sing-box" "$CONFIG_DIR"
    fi
    
    if [ -d "$temp_dir/data" ]; then
        rm -rf "$DATA_DIR"
        mv "$temp_dir/data" "$DATA_DIR"
    fi
    
    # 清理临时目录
    rm -rf "$temp_dir"
    
    echo -e "${GREEN}配置恢复完成${NC}"
    echo -e "${YELLOW}请使用 'singdns start' 重新启动服务${NC}"
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
    echo "7. 备份配置"
    echo "8. 恢复配置"
    echo "0. 退出"
    echo -e "${BLUE}=====================${NC}"
}

# 主程序
main() {
    mkdir -p "$LOG_DIR"
    
    while true; do
        show_menu
        read -p "请选择操作 [0-8]: " choice
        
        case $choice in
            1)
                check_root
                start
                ;;
            2)
                check_root
                stop
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
            7)
                check_root
                backup_config
                ;;
            8)
                check_root
                restore_config
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
            start
            ;;
        stop)
            check_root
            stop
            ;;
        restart)
            check_root
            restart
            ;;
        status)
            check_status
            ;;
        logs)
            case "$2" in
                backend)
                    logs "$SINGDNS_LOG"
                    ;;
                singbox)
                    logs "$SINGBOX_LOG"
                    ;;
                *)
                    echo "用法: $0 logs {backend|singbox}"
                    exit 1
                    ;;
            esac
            ;;
        ports)
            check_ports
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
            echo -e "${RED}无效的命令${NC}"
            echo "用法: singdns {start|stop|restart|status|logs|ports|backup|restore}"
            exit 1
            ;;
    esac
else
    main
fi