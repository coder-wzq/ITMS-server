# ITMS - 综合训练管理系统 (Integrated Training Management System)

**标识号:** ITMS-SDD-V8.5
**版本:** V1.0
**技术栈:** Go 1.24 + PostgreSQL 16 + Redis 7 + Etcd v3.5.17

---

## 项目简介

ITMS 是一个面向军事训练管理的微服务系统，覆盖训练任务全生命周期（策划→规划→执行→评估→复盘）。所有微服务共享同一 PostgreSQL 数据库，通过 Schema 逻辑隔离，对外 HTTP/WebSocket，对内 gRPC 通信。

### 技术架构

```
接入层(Nginx) → 网关层(Gateway) → 服务层(15微服务) → 数据层(PostgreSQL+Redis) → 外部系统(仿真引擎/视景/语音/LLM)
```

### 微服务 (15个)

| 服务 | 模块 | 职责 |
|------|------|------|
| svc-auth | CSCI-01 | 认证授权（JWT + RBAC） |
| svc-user | CSCI-02 | 用户管理、组织架构、设备管理 |
| svc-config | CSCI-03 | 系统配置、界面风格 |
| svc-director | CSCI-04 | 导调控制（会话管理/事件注入/气象/特情） |
| svc-task | CSCI-05 | 任务管理（任务/想定/房间/节点绑定） |
| svc-planning | CSCI-06 | 任务筹划（兵力/装备/版本管理） |
| svc-mpt | CSCI-07 | 任务规划工具（航路/空域/武器/推演） |
| svc-sim-proxy | CSCI-08 | 仿真代理网关（UDP→gRPC→WebSocket） |
| svc-situation | CSCI-09 | 二三维态势（实体位置/回放/统计） |
| svc-voice | CSCI-10 | 话音管理（聊天室/电台/通话） |
| svc-dict | CSCI-11 | 数据字典（阵营/分类/指挥单位/装备） |
| svc-agent | CSCI-12 | 智能体（行为树编辑/AI生成） |
| svc-report | CSCI-13 | 分析报告（BIN解析/评估/导出） |
| svc-record | CSCI-14 | 训练记录（回放/通信/分析） |
| svc-admin | 后台管理 | 节点审批/资源/操作日志 |

---

## 快速开始

### 环境要求

| 组件 | 版本 |
|------|------|
| Go | 1.24.x |
| PostgreSQL | 16 |
| Redis | 7 |
| Etcd | 3.5.x |
| Docker | 24.0+ (可选) |

### 1. 启动基础设施

```bash
docker compose -f deploy/docker/docker-compose.yml up -d postgres redis etcd
```

### 2. 数据库迁移与种子数据

```bash
make migrate-up
```

### 3. 配置环境变量

```bash
cp deploy/docker/.env.example .env
# 编辑 .env 修改 JWT_SECRET 等敏感配置
```

### 4. 编译

```bash
make build
# 生成 build/gateway + build/svc-user
```

### 5. 启动服务

```bash
# 先启动微服务
./build/svc-user -config services/svc-user/etc/svc-user.yaml &

# 再启动网关
./build/gateway -config deploy/configs/config.dev.yaml
```

### 6. 验证

```bash
# 网关健康检查
curl http://localhost:8080/health
# → {"code":0,"message":"success","data":{"status":"healthy"}}

# svc-user 健康检查（通过网关代理）
curl http://localhost:8080/api/user/health
# → {"code":0,"message":"success","data":{"status":"SERVING"}}
```

### 默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 系统管理员 |

---

## 项目结构

```
ITMS-server/
├── build/                      # 编译输出（gitignore）
├── cmd/                        # 入口程序
│   ├── gateway/main.go         # API 网关入口
│   ├── svc-user/main.go        # 用户管理服务入口
│   └── integration_test/       # 集成测试
├── pkg/                        # 共享库
│   ├── config/                 # YAML 配置加载 + 环境变量覆盖 + 热加载
│   ├── db/                     # GORM + PostgreSQL + Snowflake + 软删除
│   ├── etcd/                   # Etcd 服务注册/发现/心跳续约
│   ├── jwt/                    # JWT 令牌生成/解析 (HS256)
│   ├── middleware/              # 中间件链
│   │   ├── jwt.go              #   JWT 认证 (排除 login/refresh)
│   │   ├── cors.go             #   CORS 跨域
│   │   ├── ratelimit.go        #   令牌桶限流
│   │   ├── logger.go           #   请求日志
│   │   ├── recovery.go         #   Panic Recovery
│   │   └── operation_log.go    #   操作审计 (脱敏+异步写入)
│   ├── redis/                  # Redis 客户端封装 + Key 规范
│   └── response/               # 统一响应 {code,message,data,requestId} + 错误码
├── services/                   # 15 个微服务骨架
│   ├── svc-auth/ svc-user/ svc-config/ svc-director/
│   ├── svc-task/ svc-planning/ svc-mpt/ svc-sim-proxy/
│   ├── svc-situation/ svc-voice/ svc-dict/ svc-agent/
│   ├── svc-report/ svc-record/ svc-admin/
│   └── ... 每个含 etc/ internal/{handler,logic,types,svcctx} *.api *.proto
├── api/                        # 共享 API 定义
├── migrations/                 # 数据库迁移 (Schema + 种子数据)
├── vendor/                     # Go 依赖 vendor
├── deploy/                     # 部署配置
│   ├── docker/                 #   Dockerfile (多阶段构建) + docker-compose
│   ├── nginx/nginx.conf        #   Nginx 反向代理 (/api/*, /ws/*)
│   └── configs/config.dev.yaml #   开发环境 YAML 配置
├── offline-deploy/             # 内网离线部署包
│   ├── images/                 #   6 个 Docker 镜像 (~250MB)
│   ├── docker-compose.yml      #   离线版编排 (无需联网)
│   ├── init-scripts/           #   PostgreSQL 自动初始化
│   ├── load_images.sh          #   镜像导入脚本
│   ├── save_images.sh          #   镜像导出脚本
│   └── README.md               #   离线部署说明
├── task_list/                  # 交付清单
├── go.mod / go.sum
├── Makefile
└── README.md
```

---

## 开发指南

```bash
make build          # 编译 (go build -mod=vendor ./...)
make run            # 启动网关
make test           # 运行测试
make vet            # 静态分析
make fmt            # 格式化
make deps           # 更新依赖 + vendor
make migrate-up     # 执行数据库迁移
make migrate-down   # 回滚迁移
make docker-build   # 构建 Docker 镜像
make docker-up      # 启动全部服务 (docker compose)
```

---

## 开发约束与规范

### 架构分层

```
┌─────────────────────────────────────────────────┐
│  浏览器 / 外部系统                                │
└─────────────────┬───────────────────────────────┘
                  │ HTTP
┌─────────────────▼───────────────────────────────┐
│  Gateway (cmd/gateway)                          │
│  - YAML 声明路由映射 (HTTP → gRPC)               │
│  - JWT 认证 / 限流 / CORS / 操作审计             │
│  - 统一响应格式包装                               │
│  - go-zero gateway 包 + 自定义中间件              │
└─────────────────┬───────────────────────────────┘
                  │ gRPC (etcd 服务发现)
     ┌────────────┼────────────┬────────────┐
     ▼            ▼            ▼            ▼
┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
│svc-user │ │svc-auth │ │svc-task │ │  ...    │  ← 独立进程
│ :9091   │ │ :9093   │ │ :9095   │ │         │
└────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘
     │           │           │           │
     └───────────┴───────────┴───────────┘
                  │
┌─────────────────▼───────────────────────────────┐
│  PostgreSQL / Redis / etcd                      │
└─────────────────────────────────────────────────┘
```

### 通信规则

1. **对内通信**：微服务之间一律使用 gRPC + Protobuf，不允许 HTTP 直调
2. **对外暴露**：只有 Gateway 暴露 HTTP 端口（8080），其他服务只监听 gRPC 端口
3. **接口定义**：Proto 文件统一放在 `api/proto/{service}/`，所有服务引用同一份
4. **服务发现**：微服务启动时注册到 etcd（`/itms/{service-name}/{instance-id}`），Gateway 通过 etcd 发现并负载均衡
5. **Gateway 路由**：HTTP → gRPC 映射在 YAML 配置文件中声明，不手写转换代码

### 以 svc-user 为例的开发流程

#### 1. 定义 Proto（`api/proto/user/user.proto`）

```protobuf
service UserService {
  rpc ListUsers(ListUsersReq) returns (ListUsersResp);
  rpc CreateUser(CreateUserReq) returns (CreateUserResp);
  // ...
}
```

#### 2. 生成 Go 代码

```bash
protoc --go_out=. --go_opt=module=itms-server \
       --go-grpc_out=. --go-grpc_opt=module=itms-server \
       api/proto/user/user.proto
```

生成文件：`api/proto/user/user.pb.go` + `user_grpc.pb.go`

#### 3. 实现 gRPC 服务端（`services/svc-user/internal/server/grpc.go`）

```go
type GRPCServer struct {
    userpb.UnimplementedUserServiceServer
    userLogic *logic.UserLogic
}

func (s *GRPCServer) ListUsers(ctx context.Context, req *userpb.ListUsersReq) (*userpb.ListUsersResp, error) {
    // 调用 logic 层，返回 proto 响应
}
```

#### 4. 启动 gRPC 服务 + etcd 注册（`cmd/svc-user/main.go`）

```go
grpcImpl := svcuser.NewServer(db, rdb)
srv := grpc.NewServer()
userpb.RegisterUserServiceServer(srv, grpcImpl)

// 注册到 etcd
reg := etcd.NewRegistrar(etcdCli, "svc-user", 10)
reg.Register(ctx, "10.0.4.12:9091")

srv.Serve(lis)
```

#### 5. Gateway YAML 配置路由映射（`deploy/configs/config.dev.yaml`）

```yaml
Upstreams:
  - Name: svc-user
    Grpc:
      Etcd:
        Hosts:
          - localhost:2379
        Key: svc-user                    # etcd 发现键
    ProtoSets:
      - api/proto/user/user.pb          # proto 描述符（用于 JSON↔Proto 转换）
    Mappings:                            # HTTP → gRPC 路由表
      - Method: GET
        Path: /api/users
        RpcPath: user.UserService/ListUsers
      - Method: POST
        Path: /api/users
        RpcPath: user.UserService/CreateUser
```

#### 6. Gateway 启动（`cmd/gateway/main.go`）

```go
var c GatewayConfig
conf.MustLoad(*configPath, &c)
gw := gateway.MustNewServer(c.GatewayConf,
    gateway.WithMiddleware(middleware.WrapResponse),  // 统一响应格式
    // ...
)
gw.Start()
```

### 请求生命周期

```
HTTP GET /api/users?page=1
  → Gateway 接收
  → Recovery → Logger → RateLimit → CORS → JWT → OpLogger
  → 匹配 YAML Mapping: user.UserService/ListUsers
  → etcd 发现 svc-user → 10.0.4.12:9091
  → gRPC 调用 ListUsers(ListUsersReq{Page:1})
  → svc-user gRPC Server → Logic → PostgreSQL
  → Proto 响应 → JSON 序列化
  → WrapResponse 包装: {code:0, message:"success", data:{...}}
  → HTTP 200 返回
```

### 代码规范

- 目录结构遵循 Go 工程标准布局
- 变量/函数小驼峰，导出接口大驼峰，数据库字段蛇形命名
- 所有业务表使用 Snowflake BIGINT 主键
- 所有业务表含 `deleted_at` 软删除（t_operation_log 和 t_approval 除外）
- 数据库不使用外键约束，引用完整性由应用层保证
- API 统一响应：`{code, message, data, requestId}`
- 分页响应：`{list, total, page, pageSize}`

### 服务目录结构

```
services/svc-{name}/
├── server.go              # 公开 API：NewServer() 构造函数
├── etc/svc-{name}.yaml    # 服务配置
└── internal/              # 私有实现（Go internal 包约束）
    ├── server/grpc.go     # gRPC 服务端实现
    ├── logic/             # 业务逻辑（gRPC server 调用）
    ├── model/             # GORM 模型
    └── types/             # 内部类型
```

### 新增微服务 Checklist

1. `api/proto/{service}/` — 定义 proto + 生成代码
2. `services/svc-{name}/internal/` — 实现 model → logic → server
3. `services/svc-{name}/server.go` — 导出 NewServer()
4. `cmd/svc-{name}/main.go` — 启动入口（DB/Redis/etcd/gRPC）
5. `deploy/configs/config.dev.yaml` — Gateway Upstreams 添加路由映射
6. `docs/` — 更新 HTTP_API.md + gRPC_API.md

### 数据库

- 单一 PostgreSQL 数据库 `itms`，14 个 Schema 逻辑隔离
- Snowflake 算法生成全局唯一主键
- 软删除全局 Scope 自动过滤 `deleted_at IS NULL`
- 连接池：MaxOpen=100, MaxIdle=25, Lifetime=5min

---

## 配置说明

配置文件 `deploy/configs/config.dev.yaml`，支持环境变量覆盖：

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| PG_HOST | PostgreSQL 主机 | localhost |
| PG_PORT | PostgreSQL 端口 | 5432 |
| PG_USER | PostgreSQL 用户 | itms |
| PG_PASSWORD | PostgreSQL 密码 | itms@2026 |
| PG_DBNAME | PostgreSQL 数据库名 | itms |
| REDIS_ADDR | Redis 地址 | localhost:6379 |
| REDIS_PASSWORD | Redis 密码 | itms@2026 |
| ETCD_ENDPOINTS | Etcd 端点（逗号分隔） | localhost:2379 |
| JWT_SECRET | JWT 签名密钥（≥32字符） | - |

---

## 错误码体系

| 范围 | 模块 | 示例 |
|------|------|------|
| 00000 | 成功 | - |
| 01xxx | 认证授权 | 01001 参数错误 / 01101 登录失败 / 01005 权限不足 |
| 02xxx | 用户管理 | 02101 用户名重复 / 02102 批量超限 |
| 04xxx | 导调控制 | 04101 状态非法 |
| 06xxx | 任务管理 | 06101 编号重复 / 06102 导入失败 |
| 07xxx | MPT | 07101 实体超限 |
| 08xxx | 仿真代理 | 08101 连接异常 |
| 10xxx | 话音管理 | 10101 频道占用 |
| 11xxx | 数据字典 | 11101 名称重复 |
| 12xxx | 智能体 | 12101 AI 生成失败 |
| 13xxx | 分析报告 | 13101 BIN 解析失败 |
| 14xxx | 训练记录 | 14101 记录不存在 |
| 15xxx | 后台管理 | 15001 节点不存在 |

---

## 内网离线部署

参见 [`offline-deploy/README.md`](offline-deploy/README.md)

```bash
# 外网：导出镜像 + 编译产物
./offline-deploy/save_images.sh

# 将 offline-deploy/ 整个目录拷贝到内网服务器

# 内网：导入镜像 + 启动
cd offline-deploy/
./load_images.sh
docker compose up -d
```

---


---

## 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| V1.0 | 2026-05-20 | 1.1 基础框架搭建完成，Go 1.24 + vendor + 离线部署包 |
