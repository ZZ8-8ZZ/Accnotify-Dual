#!/bin/bash

# 设置镜像名称
IMAGE_NAME="trah01/lsptify-server"
VERSION="v0.0.1-beta"

echo "🚀 开始构建 x64 (amd64) 架构的 Docker 镜像..."

# 使用 buildx 构建 amd64 镜像并推送到 Docker Hub
# --platform linux/amd64 指定目标架构
# --push 直接推送
docker buildx build --platform linux/amd64 \
  -t ${IMAGE_NAME}:${VERSION} \
  -t ${IMAGE_NAME}:latest \
  --push .

if [ $? -eq 0 ]; then
    echo "✅ 构建并推送成功: ${IMAGE_NAME}:${VERSION}"
else
    echo "❌ 构建失败，请检查 Docker 配置和网络"
    exit 1
fi
