# ITMS 内网离线部署指南

## 目录结构

```
offline-deploy/
├── images/                          # Docker 镜像包
│   ├── etcd-v3.5.17.tar              # Etcd v3.5.17 (21MB)
│   ├── postgres-16-alpine.tar        # PostgreSQL 16 Alpine (106MB)
│   ├── redis-7-alpine.tar            # Redis 7 Alpine (17MB)
│   ├── nginx-1.27-alpine.tar         # Nginx 1.27 Alpine (21MB)
│   ├── itms-gateway-latest.tar       # ITMS 网关 (9.2MB)
│   └── golang-1.25-alpine.tar        # Go 1.25 Alpine 构建镜像 (62MB)
├── init-scripts/
│   └── 01-init-db.sh                 # PostgreSQL 初始化（schema + 种子数据）
├── docker-compose.yml                # 内网 docker-compose
├── nginx.conf                        # Nginx 反向代理配置
├── .env.example                      # 环境变量模板
├── load_images.sh                    # 镜像导入脚本
├── save_images.sh                    # 镜像导出脚本（外网使用）
└── README.md                         # 本文件
```

## 内网服务器部署步骤

### 1. 环境要求

| 组件 | 版本要求 |
|------|---------|
| Linux | CentOS 7+ / Ubuntu 20.04+ / 麒麟 / 统信 |
| Docker Engine | 24.0+ |
| Docker Compose | v2+ |
| CPU | x86_64 / ARM64 (鲲鹏/飞腾需替换镜像) |

### 2. 将 `offline-deploy/` 目录拷贝到内网服务器

```bash
# 方式1: U盘 / 移动硬盘
cp -r offline-deploy/ /mnt/usb/

# 方式2: 局域网 scp（如果内外网可通）
scp -r offline-deploy/ user@内网IP:/opt/itms/

# 方式3: 压缩后拷贝
tar czf itms-offline.tar.gz offline-deploy/
```

### 3. 导入 Docker 镜像

```bash
cd /opt/itms/offline-deploy/

# 方式1: 使用脚本（推荐）
chmod +x load_images.sh
./load_images.sh

# 方式2: 手动逐条导入
docker load -i images/etcd-v3.5.17.tar
docker load -i images/postgres-16-alpine.tar
docker load -i images/redis-7-alpine.tar
docker load -i images/nginx-1.27-alpine.tar
docker load -i images/itms-gateway-latest.tar
```

验证镜像已导入：

```bash
docker images | grep -E "itms|postgres|redis|nginx|etcd"
```

### 4. 配置环境变量

```bash
cp .env.example .env

# 编辑 .env，修改密码和密钥
vi .env
```

**必须修改的配置：**
```
PG_PASSWORD=<强密码>
REDIS_PASSWORD=<强密码>
JWT_SECRET=<至少32个字符的随机字符串>
```

### 5. 启动服务

```bash
# 启动（按依赖顺序自动编排：etcd → postgres/redis → gateway → nginx）
docker compose up -d

# 查看状态
docker compose ps

# 查看日志
docker compose logs -f

# 等待 PostgreSQL 健康检查通过
docker compose logs postgres | grep "ready to accept"
```

### 6. 验证服务

```bash
# 健康检查
curl http://localhost:8080/health
# → {"code":"00000","message":"ok","data":{"status":"healthy"}}

# 测试登录
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 通过 Nginx 访问
curl http://localhost/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 7. 首次登录

- 地址：`http://<内网IP>`
- 用户名：`admin`
- 密码：`admin123`
- **⚠️ 首次登录后立即修改密码**

---

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Nginx | 80/443 | 反向代理入口 |
| Gateway HTTP | 8080 | API 网关 |
| Gateway gRPC | 9090 | 微服务间通信 |
| PostgreSQL | 5432 | 数据库 |
| Redis | 6379 | 缓存 |
| Etcd | 2379/2380 | 服务注册 |

---

## 常用运维命令

```bash
# 停止所有服务
docker compose down

# 停止并删除数据卷（慎用！）
docker compose down -v

# 重启单个服务
docker compose restart gateway

# 查看 gateway 日志
docker compose logs -f gateway

# 进入 PostgreSQL
docker exec -it itms-postgres psql -U itms -d itms

# 进入 Redis
docker exec -it itms-redis redis-cli -a <密码>

# 更新镜像后重新部署
docker compose up -d --force-recreate gateway
```

---

## 数据库备份

```bash
# 备份
docker exec itms-postgres pg_dump -U itms itms > backup_$(date +%Y%m%d).sql

# 恢复
docker exec -i itms-postgres psql -U itms itms < backup_20260520.sql
```

---

## 国产 CPU 适配说明

当前镜像为 `x86_64` 架构。如需在鲲鹏/飞腾（ARM64）或海光（x86兼容）上运行，需对应替换：

| 镜像 | ARM64 替代 |
|------|-----------|
| postgres:16-alpine | `arm64v8/postgres:16-alpine` |
| redis:7-alpine | `arm64v8/redis:7-alpine` |
| nginx:1.27-alpine | `arm64v8/nginx:1.27-alpine` |
| golang:1.25-alpine | `arm64v8/golang:1.25-alpine` |
| etcd v3.5.17 | `quay.io/coreos/etcd:v3.5.17`（已多架构） |

---

## 故障排除

### PostgreSQL 连接失败
```bash
# 检查是否健康
docker compose ps postgres
# 查看日志
docker compose logs postgres
```

### 端口被占用
```bash
# 修改 docker-compose.yml 中的端口映射
# 例如 PostgreSQL: "5433:5432"
```

### 防火墙配置（CentOS/麒麟）
```bash
firewall-cmd --add-port=80/tcp --permanent
firewall-cmd --add-port=8080/tcp --permanent
firewall-cmd --reload
```

### 重置数据库
```bash
docker compose down -v postgres
docker compose up -d postgres
# 等待 postgres 健康检查通过，初始化脚本会自动执行
```

---

## 磁盘空间预估

| 数据 | 初始 | 运行1个月 |
|------|------|----------|
| 镜像 | ~240MB | ~240MB |
| PostgreSQL | ~50MB | 1-5GB |
| Redis | ~10MB | 200-500MB |
| 日志 | - | 500MB-2GB |
| **合计** | ~300MB | ~3-10GB |
