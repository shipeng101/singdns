#!/bin/sh

# 设置颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置变量
INSTALL_DIR="/usr/local/singdns"
LOG_DIR="/var/log/singdns"
TEMP_DIR="/tmp/singdns_temp"
MIN_DISK_SPACE=1024  # 需要的最小磁盘空间(MB)

# 检查是否为root用户
check_root() {
    if [ "$(id -u)" != "0" ]; then
        echo "${RED}错误：请使用root用户运行此脚本${NC}"
        return 1
    fi
    return 0
}

# 安装系统依赖
install_system_dependencies() {
    echo "${BLUE}安装系统依赖...${NC}"
    
    # 安装依赖包
    apk update
    apk add --no-cache curl wget git sqlite
    
    echo "${GREEN}系统依赖安装完成${NC}"
    return 0
}

# 安装SingDNS
install_singdns() {
    echo "${BLUE}开始安装 SingDNS...${NC}"
    
    # 创建必要的目录
    mkdir -p "$INSTALL_DIR/bin/web"
    mkdir -p "$INSTALL_DIR/web"
    mkdir -p "$INSTALL_DIR/configs/sing-box/rules"
    mkdir -p "$LOG_DIR"
    
    # 复制文件
    echo "${YELLOW}复制文件...${NC}"
    # 复制主程序
    cp singdns "$INSTALL_DIR/"
    # 复制 sing-box 到 bin 目录
    cp bin/sing-box "$INSTALL_DIR/bin/"
    # 复制 ClashAPI UI 面板到 bin/web 目录
    cp -r bin/web/* "$INSTALL_DIR/bin/web/"
    # 复制前端文件到 web 目录
    cp -r web/* "$INSTALL_DIR/web/"
    # 复制配置文件和规则文件
    cp configs/sing-box/config.json "$INSTALL_DIR/configs/sing-box/"
    cp configs/sing-box/rules/*.srs "$INSTALL_DIR/configs/sing-box/rules/"
    # 复制脚本和文档
    cp install.sh "$INSTALL_DIR/"
    cp singdns.sh "$INSTALL_DIR/"
    cp README.md "$INSTALL_DIR/"
    cp LICENSE "$INSTALL_DIR/" || true
    cp VERSION "$INSTALL_DIR/" || true
    
    # 设置权限
    echo "${YELLOW}设置文件权限...${NC}"
    find "$INSTALL_DIR" -type f -exec chmod 644 {} \;
    find "$INSTALL_DIR" -type d -exec chmod 755 {} \;
    chmod +x "$INSTALL_DIR/singdns"
    chmod +x "$INSTALL_DIR/singdns.sh"
    chmod +x "$INSTALL_DIR/bin/sing-box"
    chmod -R 755 "$INSTALL_DIR/bin"
    chmod -R 755 "$INSTALL_DIR/configs"
    chmod 755 "$LOG_DIR"
    
    # 创建符号链接
    ln -sf "$INSTALL_DIR/singdns.sh" "/usr/local/bin/singdns"
    
    echo "${GREEN}SingDNS 安装完成${NC}"
    return 0
}

# 卸载函数
uninstall_singdns() {
    echo "${YELLOW}开始卸载 SingDNS...${NC}"
    
    # 停止服务
    if command -v singdns > /dev/null; then
        singdns stop
        sleep 2
    fi
    
    # 删除文件
    rm -rf "$INSTALL_DIR"
    rm -f "/usr/local/bin/singdns"
    rm -rf "$LOG_DIR"
    
    echo "${GREEN}SingDNS 已成功卸载${NC}"
    return 0
}

# 主函数
main() {
    echo "${YELLOW}欢迎使用 SingDNS 安装程序${NC}"
    echo "请选择操作："
    echo "1. 安装 SingDNS"
    echo "2. 卸载 SingDNS"
    echo "3. 退出"
    
    printf "请输入选项 [1-3]: "
    read choice
    
    case $choice in
        1)
            check_root || exit 1
            install_system_dependencies || exit 1
            install_singdns || exit 1
            
            # 显示安装成功信息
            echo "${GREEN}SingDNS 安装成功！${NC}"
            echo "${GREEN}使用 'singdns' 命令来管理 SingDNS${NC}"
            echo "${YELLOW}提示: 使用 'singdns start' 启动服务${NC}"
            ;;
        2)
            check_root || exit 1
            uninstall_singdns
            ;;
        3)
            echo "${GREEN}感谢使用！${NC}"
            exit 0
            ;;
        *)
            echo "${RED}无效的选项${NC}"
            exit 1
            ;;
    esac
}

# 执行主函数
main 