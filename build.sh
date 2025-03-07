#!/bin/bash

# 创建 build 目录
mkdir -p build

# 编译 Linux 版本
GOOS=linux GOARCH=amd64 go build -o build/sol-circular-tool
echo "Linux 版本编译完成"

echo "所有版本编译完成！"
echo "使用方法: ./sol-circular-tool -j <Jupiter API URL> -k <API Key> [选项]"
echo "可选参数:"
echo "  -i, --dex-program-ids <ID1,ID2,...>         只包含指定的DEX程序ID"
echo "  -e, --exclude-dex-program-ids <ID1,ID2,...> 排除指定的DEX程序ID" 