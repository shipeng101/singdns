#!/bin/bash

# 创建必要的目录
mkdir -p bin tmp
cd tmp

# 下载 mosdns
echo "正在下载 mosdns..."
rm -f ../bin/mosdns
curl -L -o mosdns.zip "https://github.com/IrineSistiana/mosdns/releases/download/v5.3.1/mosdns-darwin-amd64.zip"
unzip mosdns.zip
chmod +x mosdns
mv mosdns ../bin/
rm mosdns.zip

# 下载 sing-box
echo "正在下载 sing-box..."
curl -L -o sing-box.tar.gz "https://github.com/SagerNet/sing-box/releases/download/v1.7.0/sing-box-1.7.0-darwin-amd64.tar.gz"
tar xzf sing-box.tar.gz
mv sing-box-*/sing-box ../bin/
chmod +x ../bin/sing-box
rm -rf sing-box-* sing-box.tar.gz

cd ..
rm -rf tmp

echo "下载完成！" 