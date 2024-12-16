#!/bin/bash

# 创建规则目录
mkdir -p config/mosdns/rules

# 下载广告过滤规则
echo "正在下载广告过滤规则..."
curl -L -o config/mosdns/rules/adblock.txt "https://raw.githubusercontent.com/privacy-protection-tools/anti-AD/master/anti-ad-domains.txt"

# 下载中国域名列表
echo "正在下载中国域名列表..."
curl -L -o config/mosdns/rules/china-list.txt "https://raw.githubusercontent.com/Loyalsoldier/v2ray-rules-dat/release/direct-list.txt"

# 下载代理域名列表
echo "正在下载代理域名列表..."
curl -L -o config/mosdns/rules/proxy-list.txt "https://raw.githubusercontent.com/Loyalsoldier/v2ray-rules-dat/release/proxy-list.txt"

echo "规则下载完成！" 