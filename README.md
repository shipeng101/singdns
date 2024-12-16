# SingDNS

SingDNS 是一个基于 Sing-box 的代理管理工具，提供直观的 Web 界面来管理代理配置、订阅、规则和 DNS 设置。

## 功能特性

### 代理管理
- 自动配置 iptables/nftables 代理转发
- 支持全局代理和规则代理模式切换
- 可视化节点管理界面
- 规则编辑与自动重启
- 流量监控和统计

### 订阅管理
- 支持多种订阅格式转换为 sing-box 格式
- 自动更新订阅
- 节点分组管理
- 灵活的过滤选项

### DNS 管理
- 基于 mosdns 的 DNS 解析
- 自动更新规则
- 广告过滤功能
- 自定义 DNS 服务器配置

### 系统功能
- 支持暗色/亮色主题
- 系统监控和日志
- 配置导入导出
- 用户友好的 Web 界面

## 技术栈

### 后端
- Go 1.19+
- Sing-box 代理处理
- mosdns DNS 管理
- Gin Web 框架

### 前端
- Vue 3
- TypeScript
- Element Plus UI 框架
- Vite 构建工具

## 快速开始

### 环境要求
- Go 1.19 或更高版本
- Node.js 16 或更高版本
- sing-box
- mosdns

### 安装步骤

1. 克隆仓库:
```bash
git clone https://github.com/shipeng101/singdns.git
cd singdns
```

2. 安装前端依赖:
```bash
cd web
npm install
```

3. 构建前端:
```bash
npm run build
```

4. 构建后端:
```bash
cd ..
go mod tidy
go build
```

### 配置说明

主要配置文件位置:
- `config/`: 主配置目录
- `config/config.yaml`: 核心配置文件
- `config/mosdns/`: DNS 配置文件
- `config/rules/`: 规则配置文件

### 运行服务

1. 启动服务:
```bash
./singdns
```

2. 访问 Web 界面:
打开浏览器访问 `http://localhost:3000`

## 项目结构

```
singdns/
├── api/                # API 处理器
│   ├── proxy/         # 代理相关 API
│   ├── subscription/  # 订阅 API
│   ├── dns/          # DNS API
│   └── system/       # 系统 API
├── service/          # 业务逻辑
│   ├── proxy/        # 代理服务
│   ├── dns/         # DNS 服务
│   ├── subscription/ # 订阅服务
│   └── system/      # 系统服务
├── model/           # 数据模型
├── config/          # 配置文件
├── web/            # 前端应用
│   ├── src/        # 源代码
│   ├── public/     # 静态资源
│   └── dist/       # 构建输出
└── docs/           # 文档
```

## 开发文档

详细的开发文档请参考:
- [开发指南](docs/development.md)
- [API 文档](docs/api.md)

## 贡献代码

欢迎提交 Pull Request 来贡献代码！

## 开源协议

本项目采用 MIT 协议开源 - 详见 LICENSE 文件 