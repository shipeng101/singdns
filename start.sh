#!/bin/bash

# 启动 sing-box
./bin/sing-box run -c configs/sing-box/config.json &

# 启动 mosdns
./bin/mosdns start -d configs/mosdns &

# 启动主程序
go run main.go 