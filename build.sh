#!/bin/bash

# 创建 build 目录
mkdir -p build

# 编译 Linux 版本
GOOS=linux GOARCH=amd64 go build -o build/sol-circular-tool-linux
echo "Linux 版本编译完成"

# 编译 Windows 版本
GOOS=windows GOARCH=amd64 go build -o build/sol-circular-tool-windows.exe
echo "Windows 版本编译完成"

# 编译 MacOS 版本
GOOS=darwin GOARCH=amd64 go build -o build/sol-circular-tool-mac
echo "MacOS 版本编译完成"

echo "所有版本编译完成！" 