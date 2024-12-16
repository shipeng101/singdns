# API 文档

## API 概述

所有 API 使用 RESTful 风格，返回 JSON 格式数据。基础 URL 为 `http://localhost:8080/api`。

## 错误响应

错误响应格式:
```json
{
    "code": 400,
    "message": "错误信息"
}
```

## API 列表

### 代理相关 API

#### 1. 获取代理状态
- 请求: `GET /api/proxy/status`
- 响应:
```json
{
    "mode": "rule",
    "running": true,
    "uptime": 3600,
    "up_speed": 1024,
    "down_speed": 2048,
    "up_total": 10240,
    "down_total": 20480,
    "nodes": [
        {
            "id": "node1",
            "connected": true,
            "up_speed": 512,
            "down_speed": 1024,
            "up_total": 5120,
            "down_total": 10240
        }
    ]
}
```

#### 2. 切换代理模式
- 请求: `POST /api/proxy/mode`
- 参数:
```json
{
    "mode": "rule"  // rule/global/direct
}
```
- 响应: `200 OK`

#### 3. 测试节点延迟
- 请求: `POST /api/proxy/test/{id}`
- 响应:
```json
{
    "latency": 100  // 延迟(ms)
}
```

### 订阅相关 API

#### 1. 获取订阅列表
- 请求: `GET /api/subscription/list`
- 响应:
```json
{
    "subscriptions": [
        {
            "id": "sub1",
            "name": "订阅1",
            "url": "https://example.com/sub",
            "auto_update": true,
            "update_hours": 24,
            "nodes": [
                {
                    "id": "node1",
                    "name": "节点1",
                    "type": "ss",
                    "server": "1.2.3.4",
                    "port": 443
                }
            ]
        }
    ]
}
```

#### 2. 添加订阅
- 请求: `POST /api/subscription/add`
- 参数:
```json
{
    "name": "订阅1",
    "url": "https://example.com/sub",
    "auto_update": true,
    "update_hours": 24
}
```
- 响应: `200 OK`

#### 3. 更新订阅
- 请求: `POST /api/subscription/update/{id}`
- 响应: `200 OK`

### DNS 相关 API

#### 1. 获取 DNS 设置
- 请求: `GET /api/dns/settings`
- 响应:
```json
{
    "listen": "0.0.0.0:53",
    "upstream": [
        "tcp://8.8.8.8:53"
    ],
    "china_dns": [
        "tcp://223.5.5.5:53"
    ],
    "rules": []
}
```

#### 2. 更新 DNS 设置
- 请求: `PUT /api/dns/settings`
- 参数: 同上
- 响应: `200 OK`

#### 3. 获取 DNS 规则集
- 请求: `GET /api/dns/rulesets`
- 响应:
```json
{
    "rulesets": [
        {
            "name": "广告过滤",
            "type": "domain",
            "url": "https://example.com/rules.txt",
            "count": 1000,
            "enabled": true
        }
    ]
}
```

### 系统相关 API

#### 1. 获取系统状态
- 请求: `GET /api/system/status`
- 响应:
```json
{
    "version": "1.0.0",
    "uptime": 3600,
    "memory": {
        "total": 8589934592,
        "used": 4294967296
    },
    "cpu": {
        "cores": 4,
        "usage": 0.5
    }
}
```

#### 2. 获取系统设置
- 请求: `GET /api/system/settings`
- 响应:
```json
{
    "log": {
        "level": "info",
        "file": "logs/singdns.log",
        "max_size": 100,
        "max_backups": 3,
        "max_age": 28,
        "compress": true
    },
    "ui": {
        "theme": "light",
        "language": "zh-CN",
        "dashboard": "yacd"
    }
}
```

#### 3. 更新系统设置
- 请求: `PUT /api/system/settings`
- 参数: 同上
- 响应: `200 OK`

#### 4. 获取系统日志
- 请求: `GET /api/system/logs`
- 参数:
```
?level=info&start=0&limit=100
```
- 响应:
```json
{
    "total": 1000,
    "logs": [
        {
            "time": "2024-01-01 12:00:00",
            "level": "info",
            "message": "服务启动"
        }
    ]
}
```

#### 5. 重启服务
- 请求: `POST /api/system/restart`
- 响应: `200 OK`