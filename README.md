# SingDNS

SingDNS 是一个基于 Go 语言开发的高性能 DNS 代理工具，提供了友好的 Web 界面，支持多种代理协议，支持规则分流，支持订阅管理。

## 功能特点

- 🚀 高性能 DNS 代理
- 🌐 支持多种代理协议
- 📱 美观的 Web 管理界面
- 🔄 支持订阅管理
- 📊 实时流量监控
- 🎯 规则分流
- 🔒 安全认证
- 🌓 支持深色模式

## 前端功能

### 仪表盘
- 系统状态监控
- 流量统计图表
- 节点状态概览
- 快速操作面板

### 节点管理
- 节点列表查看
- 添加/编辑/删除节点
- 节点分组管理
- 节点延迟测试
- 节点排序和筛选

### 规则管理
- 规则列表查看
- 添加/编辑/删除规则
- 规则优先级调整
- 规则导入导出
- 规则测试

### 订阅管理
- 订阅列表查看
- 添加/编辑/删除订阅
- 订阅更新和转换
- 自动更新设置
- 订阅分组管理

### 系统设置
- 基本设置
- DNS 设置
- 安全设置
- 主题设置
- 备份还原

## 技术栈

### 后端
- Go
- Gin Web Framework
- SQLite/MySQL
- JWT Authentication

### 前端
- React 18
- Material-UI (MUI)
- Axios
- React Router
- Redux Toolkit
- Recharts (图表)
- Day.js (时间处理)

## 项目结构

```
singdns/
├── api/            # 后端 API 实现
├── cmd/            # 命令行工具
├── configs/        # 配置文件
├── web/           # 前端代码
│   ├── src/
│   │   ├── components/  # React 组件
│   │   ├── pages/      # 页面组件
│   │   │   ├── Dashboard.js    # 仪表盘
│   │   │   ├── Login.js        # 登录页面
│   │   │   ├── Settings.js     # 设置页面
│   │   │   ├── Subscriptions.js # 订阅管理
│   │   │   ├── Nodes.js        # 节点管理
│   │   │   └── Rules.js        # 规则管理
│   │   ├── services/   # API 服务
│   │   ├── hooks/      # 自定义 Hooks
│   │   └── styles/     # 样式文件
│   └── public/
├── bin/           # 二进制文件
└── data/          # 数据文件
```

## 安装部署

### 环境要求
- Linux 系统
- curl 或 wget（用于下载）
- iptables（用于流量转发）

### 快速安装（推荐）
```bash
# 下载安装脚本
curl -O https://raw.githubusercontent.com/shipeng101/singdns/main/install.sh

# 添加执行权限
chmod +x install.sh

# 安装
sudo ./install.sh

# 卸载
sudo ./install.sh uninstall
```

### 手动部署
```bash
# 克隆项目
git clone https://github.com/shipeng101/singdns.git

# 进入项目目录
cd singdns

# 使用安装脚本
chmod +x install.sh
./install.sh
```

### 服务管理
```bash
# 启动服务
singdns start

# 停止服务
singdns stop

# 查看状态
singdns status

# 查看日志
singdns logs

# 备份配置
singdns backup

# 恢复配置
singdns restore
```

## 访问管理界面

安装完成后，可以通过以下地址访问管理界面：

```
http://<服务器IP>:3000
```

## 配置说明

### 配置文件
配置文件位于 `configs/sing-box/config.json`，包含以下主要配置项：

```json
{
  "dns": {
    "servers": [
      {
        "tag": "google",
        "address": "8.8.8.8",
        "detour": "direct"
      }
    ],
    "rules": []
  },
  "inbounds": [],
  "outbounds": [
    {
      "type": "direct",
      "tag": "direct"
    }
  ]
}
```

## 更新日志

### v1.0.8
- 移除 nginx 依赖，使用内置服务提供前端访问
- 简化服务管理逻辑
- 优化安装脚本

### v1.0.7
- 修复安装脚本问题
- 更新管理界面显示

### v1.0.6
- 添加自动构建功能
- 优化项目结构

### v1.0.5
- 初始版本发布
- 基础功能实现

## 开源协议

本项目采用 MIT 协议开源，详见 [LICENSE](LICENSE) 文件。 