# SingDNS 开发文档

## 一、项目准备阶段

### 1. 环境搭建 ✅
- [x] 创建项目目录结构 ✅
- [x] 编写项目文档 ✅
- [x] 搭建开发环境 ✅
  - [x] Go 1.19+ ✅
  - [ ] Node.js 16+ (前端未开发) ❌
  - [x] sing-box ✅
  - [x] mosdns ✅
- [x] 初始化项目
  - [x] 创建 go.mod ✅
  - [ ] 初始化前端项目 (未开始) ❌

### 2. 基础框架搭建
- [x] 后端框架搭建 ✅
  - [x] 选择并集成 Web 框架 (gin) ✅
  - [x] 配置日志系统 ✅
  - [x] 实现中间件(认证、日志等) ✅
- [ ] 前端框架搭建 (未开始) ❌
  - [ ] 创建 React 项目
  - [ ] 配置路由
  - [ ] 配置状态管理
  - [ ] 实现主题切换

## 二、核心功能开发

### 1. 配置管理模块
- [x] 配置文件结构设计 ✅
- [x] 配置读写功能 ✅
- [x] 配置验证功能 ✅
- [x] 配置热重载 ✅
- [ ] 配置备份/恢复 (未实现) ❌

### 2. 代理管理模块
- [x] sing-box 配置管理 ✅
  - [x] 配置生成 ✅
  - [x] 服务控制 ✅
  - [x] 状态监控 ✅
- [ ] 全局/规则代理切换 (部分实现)
  - [x] 基本代理切换 ✅
  - [ ] 自定义规则路由 ❌
- [ ] 流量统计 (未实现) ❌

### 3. 订阅管理模块
- [x] 订阅格式解析 ✅
  - [x] SS/SSR ✅
  - [x] Vmess ✅
  - [x] Trojan ✅
- [x] 订阅转换 ✅
  - [x] 转换为 sing-box 配置 ✅
  - [x] 节点过滤 ✅
  - [ ] 节点分组 (未实现) ❌
- [ ] 自动更新 (未实现) ❌
  - [ ] 定时更新
  - [ ] 错误处理
  - [ ] 更新通知

### 4. DNS 管理模块
- [x] mosdns 配置管理 ✅
  - [x] 配置生成 ✅
  - [x] 规则管理 ✅
  - [x] 服务控制 ✅
- [x] DNS 分流规则 ✅
  - [x] 国内外分流 ✅
  - [ ] 自定义分流规则 (未实现) ❌
- [ ] 广告过滤 (未实现) ❌

### 5. 系统管理模块
- [x] 系统监控 ✅
  - [x] 服务状态 ✅
  - [ ] 资源使用统计 (未实现) ❌
  - [ ] 连接统计 (未实现) ❌
- [x] 日志管理 ✅
- [ ] 系统维护 (部分实现)
  - [x] 服务重启 ✅
  - [ ] 规则更新 ❌
  - [ ] 系统更新 ❌

## 三、前端界面开发 (未开始) ❌

### 1. 基础界面
- [ ] 布局设计 ❌
- [ ] 导航菜单 ❌
- [ ] 主题切换 ❌
- [ ] 响应式适配 ❌

### 2. 功能界面
- [ ] 仪表盘 ❌
  - [ ] 系统状态
  - [ ] 流量统计
  - [ ] 快速操作
- [ ] 代理管理 ❌
  - [ ] 节点列表
  - [ ] 规则配置
  - [ ] 模式切换
- [ ] 订阅管理 ❌
  - [ ] 订阅列表
  - [ ] 节点分组
  - [ ] 更新配置
- [ ] 设置界面 ❌
  - [ ] 系统设置
  - [ ] DNS 设置
  - [ ] 规则设置

## 四、API 接口文档

### 1. 认证接口 ✅
- [x] POST `/api/auth/login` ✅
  - 功能：用户登录
  - 请求体：
    ```json
    {
      "username": "admin",
      "password": "password"
    }
    ```
  - 响应：
    ```json
    {
      "code": 0,
      "message": "success",
      "data": {
        "token": "jwt-token"
      }
    }
    ```

- [x] POST `/api/auth/logout` ✅
  - 功能：用户登出
  - 请求头：`Authorization: Bearer <token>`
  - 响应：
    ```json
    {
      "code": 0,
      "message": "success"
    }
    ```

- [x] GET `/api/auth/info` ✅
  - 功能：获取用户信息
  - 请求头：`Authorization: Bearer <token>`
  - 响应：
    ```json
    {
      "code": 0,
      "message": "success",
      "data": {
        "username": "admin",
        "role": "admin"
      }
    }
    ```

### 2. DNS 管理接口 ✅
- [x] GET `/api/dns/config` ✅
  - 功能：获取 DNS 配置
  - 请求头：`Authorization: Bearer <token>`
  - 响应：
    ```json
    {
      "code": 0,
      "message": "success",
      "data": {
        "listen": "127.0.0.1",
        "port": 5354,
        "cache": true,
        "upstream": ["udp://8.8.8.8:53"],
        "china_dns": ["udp://223.5.5.5:53"]
      }
    }
    ```

- [x] PUT `/api/dns/config` ✅
  - 功能：更新 DNS 配置
  - 请求头：`Authorization: Bearer <token>`
  - 请求体：与 GET 响应格式相同
  - 响应：
    ```json
    {
      "code": 0,
      "message": "success"
    }
    ```

- [x] GET `/api/dns/status` ✅
  - 功能：获取 DNS 服务状态
  - 请求头：`Authorization: Bearer <token>`
  - 响应：
    ```json
    {
      "code": 0,
      "message": "success",
      "data": {
        "running": true
      }
    }
    ```

### 3. 代理管理接口 ✅
- [x] GET `/api/proxy/nodes` ✅
  - 功能：获取节点列表
  - 请求头：`Authorization: Bearer <token>`
  - 响应：
    ```json
    {
      "code": 0,
      "message": "success",
      "data": {
        "nodes": [
          {
            "id": "node-id",
            "name": "节点名称",
            "type": "shadowsocks",
            "server": "server.com",
            "port": 443
          }
        ]
      }
    }
    ```

- [x] POST `/api/proxy/nodes` ✅
  - 功能：添加节点
  - 请求头：`Authorization: Bearer <token>`
  - 请求体：节点配置信息
  - 响应：
    ```json
    {
      "code": 0,
      "message": "success",
      "data": {
        "id": "new-node-id"
      }
    }
    ```

- [x] PUT `/api/proxy/nodes/:id` ✅
  - 功能：更新节点
  - 请求头：`Authorization: Bearer <token>`
  - 请求体：节点配置信息
  - 响应：
    ```json
    {
      "code": 0,
      "message": "success"
    }
    ```

- [x] DELETE `/api/proxy/nodes/:id` ✅
  - 功能：删除节点
  - 请求头：`Authorization: Bearer <token>`
  - 响应：
    ```json
    {
      "code": 0,
      "message": "success"
    }
    ```

- [x] POST `/api/proxy/subscription/import` ✅
  - 功能：导入订阅
  - 请求头：`Authorization: Bearer <token>`
  - 请求体：
    ```json
    {
      "url": "订阅地址",
      "type": "订阅类型"
    }
    ```
  - 响应：
    ```json
    {
      "code": 0,
      "message": "success",
      "data": {
        "imported": 10,
        "failed": 0
      }
    }
    ```

### 4. 系统管理接口 ✅
- [x] GET `/api/system/status` ✅
  - 功能：获取系统状态
  - 请求头：`Authorization: Bearer <token>`
  - 响应：
    ```json
    {
      "code": 0,
      "message": "success",
      "data": {
        "dns": {
          "running": true
        },
        "proxy": {
          "running": true
        }
      }
    }
    ```

## 五、待开发功能 ❌

1. 前端界面
   - Web 管理界面的所有功能
   - 响应式设计
   - 主题切换

2. 高级功能
   - 节点分组管理
   - 自定义路由规则
   - 流量统计
   - 广告过滤
   - 资源使用监控

3. 自动化功能
   - 订阅自动更新
   - 规则自动更新
   - 系统自动更新

4. 其他功能
   - 配置备份/恢复
   - 详细的使用统计
   - 多用户支持

## 六、后续开发计划

1. 优先级高
   - 完成前端界面开发
   - 实现配置备份/恢复
   - 添加自定义路由规则

2. 优先级中
   - 实现节点分组
   - 添加流量统计
   - 完善系统监控

3. 优先级低
   - 广告过滤
   - 自动更新功能
   - 多用户支持