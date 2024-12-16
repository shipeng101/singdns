# SMBox 开发文档

## 项目介绍

SMBox 是一个基于 Singbox 的代理管理工具，提供了易用的 Web 界面来管理代理配置、订阅、规则等功能。

## 功能特性

1. 代理管理
   - 自动配置 iptables/nftables 代理转发
   - 支持全局代理/规则代理模式切换
   - 可视化节点管理
   - 规则编辑与自动重启

2. 订阅功能
   - 支持多种订阅格式转换为 singbox 格式
   - 自动更新订阅
   - 节点分组管理

3. DNS 管理
   - 基于 mosdns 的 DNS 解析
   - 自动更新规则
   - 广告过滤

4. 系统功能
   - 支持夜间主题
   - 支持 yacd/metacubexd 面板切换
   - 系统监控和日志查看
   - 配置导入导出

5. 远程访问
   - DDNS 配置
   - 端口转发
   - 远程访问家里服务

## 技术栈

- 前端: React + TypeScript
- 后端: Go
- 代理: Singbox
- DNS: mosdns

## 目录结构

```
smbox/
├── api/                # API接口
│   ├── proxy/         # 代理相关API
│   ├── subscription/  # 订阅相关API
│   ├── config/        # 配置相关API
│   └── system/        # 系统相关API
├── service/           # 业务逻辑
│   ├── proxy/         # 代理服务
│   ├── dns/          # DNS服务
│   ├── subscription/ # 订阅服务
│   └── system/       # 系统服务
├── model/            # 数据模型
├── config/           # 配置文件
└── utils/            # 工具函数
```

## 开发环境搭建

1. 依赖安装
   - Go 1.19+
   - Node.js 16+
   - singbox
   - mosdns

2. 编译运行
   ```bash
   # 编译前端
   cd web
   npm install
   npm run build

   # 编译后端
   go mod tidy
   go build
   ```

3. 配置文件
   - config.yaml: 主配置文件
   - mosdns/config.yaml: DNS配置
   - rules/: 规则文件目录 