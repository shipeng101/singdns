# SingDNS

SingDNS 是一个基于 sing-box 的代理管理工具，集成了 DNS 分流、规则管理、订阅转换等功能。

## 功能特性

- 支持多种代理协议（Shadowsocks、VMess、Trojan）
- DNS 分流和广告过滤
- 规则管理和自动分流
- 订阅转换和节点管理
- 系统代理和透明代理
- Web 管理界面
- 主题切换
- JWT 认证

## 安装

```bash
# 克隆仓库
git clone https://github.com/shipeng101/singdns.git

# 进入项目目录
cd singdns

# 安装依赖
go mod download

# 编译
go build
```

## 配置

配置文件位于 `config/config.yaml`，包含以下主要配置项：

- 服务器配置（监听地址、端口等）
- DNS 配置（上游服务器、缓存等）
- 代理配置（入站、出站规则等）
- 日志配置

## 使用

```bash
# 启动服务
./singdns

# 指定配置文件
./singdns -config config/config.yaml
```

## 开发

```bash
# 运行后端
go run main.go

# 运行前端
cd web
npm install
npm run dev
```

## 许可证

MIT License