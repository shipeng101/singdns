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

## 技术栈

### 后端
- Go
- Gin Web Framework
- SQLite/MySQL
- JWT Authentication

### 前端
- React
- Material-UI (MUI)
- Axios
- React Router
- Redux

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
│   │   ├── services/   # API 服务
│   │   └── styles/     # 样式文件
│   └── public/
└── docs/           # 项目文档
```

## API 文档

### 已实现的 API

#### 认证相关
- POST `/api/auth/login` - 用户登录
- POST `/api/auth/register` - 用户注册
- GET `/api/user` - 获取用户信息
- PUT `/api/user/password` - 修改密码

#### 系统相关
- GET `/api/system/status` - 获取系统状态
- GET `/api/system/services` - 获取服务列表
- POST `/api/system/services/:name/start` - 启动服务
- POST `/api/system/services/:name/stop` - 停止服务
- POST `/api/system/services/:name/restart` - 重启服务

#### 节点相关
- GET `/api/nodes` - 获取节点列表
- POST `/api/nodes` - 创建节点
- PUT `/api/nodes/:id` - 更新节点
- DELETE `/api/nodes/:id` - 删除节点

#### 规则相关
- GET `/api/rules` - 获取规则列表
- POST `/api/rules` - 创建规则
- PUT `/api/rules/:id` - 更新规则
- DELETE `/api/rules/:id` - 删除规则

#### 订阅相关
- GET `/api/subscriptions` - 获取订阅列表
- POST `/api/subscriptions` - 创建订阅
- PUT `/api/subscriptions/:id` - 更新订阅
- DELETE `/api/subscriptions/:id` - 删除订阅
- POST `/api/subscriptions/:id/update` - 更新订阅节点

#### 设置相关
- GET `/api/settings` - 获取设置
- PUT `/api/settings` - 更新设置

### 待实现的 API

#### 系统相关
- GET `/api/system/info` - 获取系统详细信息（CPU、内存、运行时间等）

#### 节点相关
- GET `/api/nodes/{id}/status` - 获取节点状态
- POST `/api/nodes/import` - 导入节点
- POST `/api/nodes/{id}/test` - 测试节点

#### 订阅相关
- POST `/api/subscriptions/{id}/refresh` - 刷新订阅

#### 流量统计
- GET `/api/traffic/stats` - 获取流量统计
- GET `/api/traffic/realtime` - 获取实时流量

#### 节点组管理
- GET `/api/node-groups` - 获取节点组列表
- POST `/api/node-groups` - 创建节点组
- PUT `/api/node-groups/{id}` - 更新节点组
- DELETE `/api/node-groups/{id}` - 删除节点组

## 开发进度

### 后端进度
- [x] 基础框架搭建
- [x] 用户认证系统
- [x] 节点管理
- [x] 规则管理
- [x] 订阅管理
- [x] 设置管理
- [ ] 系统监控
- [ ] 流量统计
- [ ] 节点组管理
- [ ] 性能优化

### 前端进度
- [x] 项目初始化
- [x] 登录注册页面
- [x] 仪表盘页面
- [x] 节点管理页面
- [x] 规则管理页面
- [x] 订阅管理页面
- [x] 设置页面
- [x] 深色模式
- [x] 响应式布局
- [ ] 实时数据更新
- [ ] 性能优化

## 安装部署

### 环境要求
- Go 1.16+
- Node.js 14+
- npm 6+ 或 yarn 1.22+

### 后端部署
```bash
# 克隆项目
git clone https://github.com/shipeng101/singdns.git

# 进入项目目录
cd singdns

# 编译
go build -o singdns cmd/main.go

# 运行
./singdns
```

### 前端部署
```bash
# 进入前端目录
cd web

# 安装依赖
npm install
# 或
yarn install

# 开发模式运行
npm start
# 或
yarn start

# 构建生产版本
npm run build
# 或
yarn build
```

## 配置说明

### 配置文件
配置文件位于 `configs/config.yaml`，包含以下主要配置项：

```yaml
server:
  host: 0.0.0.0
  port: 8080
  jwt_secret: your-secret-key

database:
  type: sqlite
  path: data/singdns.db

dns:
  listen: 0.0.0.0:53
  cache_size: 4096
  cache_ttl: 60

log:
  level: info
  file: logs/singdns.log
```

## 贡献指南

1. Fork 本仓库
2. 创建新的功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。
