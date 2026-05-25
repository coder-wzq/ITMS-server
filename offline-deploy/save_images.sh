#!/bin/bash
# 外网环境执行：导出所有 Docker 镜像
# Run on internet-connected machine to export Docker images

set -e

IMAGES_DIR="$(cd "$(dirname "$0")" && pwd)/images"
mkdir -p "$IMAGES_DIR"

echo "=== ITMS Docker Image Export ==="
echo "Target: $IMAGES_DIR"
echo ""

# 基础设施镜像列表
IMAGES=(
  "postgres:16-alpine"
  "redis:7-alpine"
  "nginx:1.27-alpine"
  "quay.io/coreos/etcd:v3.5.17"
  "golang:1.24-alpine"
)

for img in "${IMAGES[@]}"; do
  echo "[PULL] $img"
  docker pull "$img"

  # 文件名：替换 / 和 : 为 -
  fname=$(echo "$img" | sed 's/\//-/g; s/:/-/g')
  echo "[SAVE] $img → $fname.tar"
  docker save -o "$IMAGES_DIR/${fname}.tar" "$img"
done

# 构建并导出网关镜像
echo ""
echo "[BUILD] itms-gateway:latest"
docker build -t itms-gateway:latest -f "$(dirname "$0")/../deploy/docker/Dockerfile" "$(dirname "$0")/.."
echo "[SAVE] itms-gateway:latest → itms-gateway-latest.tar"
docker save -o "$IMAGES_DIR/itms-gateway-latest.tar" itms-gateway:latest

echo ""
echo "=== Export Complete ==="
ls -lh "$IMAGES_DIR/"
echo ""
echo "将 offline-deploy/ 整个目录拷贝到内网服务器，然后执行 load_images.sh"
echo "Copy the entire offline-deploy/ directory to the air-gapped server, then run load_images.sh"
