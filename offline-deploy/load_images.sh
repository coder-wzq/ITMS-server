#!/bin/bash
# 内网环境执行：导入 Docker 镜像
# Run on air-gapped machine to load Docker images

set -e

IMAGES_DIR="$(cd "$(dirname "$0")" && pwd)/images"

if [ ! -d "$IMAGES_DIR" ]; then
  echo "Error: images/ directory not found at $IMAGES_DIR"
  exit 1
fi

echo "=== ITMS Docker Image Import ==="
echo "Source: $IMAGES_DIR"
echo ""

for f in "$IMAGES_DIR"/*.tar; do
  if [ -f "$f" ]; then
    fname=$(basename "$f")
    echo "[LOAD] $fname"
    docker load -i "$f"
  fi
done

echo ""
echo "=== Import Complete ==="
docker images | grep -E "itms-gateway|postgres|redis|nginx|etcd"
echo ""
echo "镜像导入完毕，执行 docker compose 启动服务："
echo "  cd offline-deploy/"
echo "  docker compose up -d"
echo ""
echo "Images loaded. Start services with:"
echo "  cd offline-deploy/"
echo "  docker compose up -d"
