#!/bin/bash

# 创建 build 目录
mkdir -p build

# 编译 Linux 版本
GOOS=linux GOARCH=amd64 go build -o build/circular
echo "Linux 版本编译完成"


echo "所有版本编译完成！" 