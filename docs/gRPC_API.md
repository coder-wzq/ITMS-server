# CSCI-02 用户管理 — gRPC 接口文档

**标识号:** ITMS-API-gRPC-V1.0  
**Proto 文件:** `api/proto/user/user.proto`  
**Go Package:** `itms-server/api/proto/user`  
**服务发现:** etcd `/itms/svc-user/`  
**序列化:** Protocol Buffers 3

---

## 服务定义

```protobuf
service UserService {
  // 用户 CRUD
  rpc ListUsers(ListUsersReq) returns (ListUsersResp);
  rpc CreateUser(CreateUserReq) returns (CreateUserResp);
  rpc UpdateUser(UpdateUserReq) returns (UpdateUserResp);
  rpc DeleteUser(DeleteUserReq) returns (DeleteUserResp);
  rpc BatchDeleteUsers(BatchDeleteUsersReq) returns (BatchDeleteUsersResp);

  // 用户档案
  rpc GetProfile(GetProfileReq) returns (GetProfileResp);
  rpc UpdateProfile(UpdateProfileReq) returns (UpdateProfileResp);

  // 组织架构
  rpc GetOrgTree(GetOrgTreeReq) returns (GetOrgTreeResp);
}
```

---

## 调用方式

### Go 客户端

```go
import userpb "itms-server/api/proto/user"

// 通过 etcd 发现
etcdCli, _ := etcd.NewClient(&etcd.Config{Endpoints: []string{"localhost:2379"}})
conn, _ := etcd.DialService(etcdCli, "svc-user")
client := userpb.NewUserServiceClient(conn)

// 调用
resp, _ := client.ListUsers(ctx, &userpb.ListUsersReq{Page: 1, PageSize: 20})
```

### 直接连接（开发环境）

```go
conn, _ := grpc.NewClient("localhost:9090",
    grpc.WithTransportCredentials(insecure.NewCredentials()))
client := userpb.NewUserServiceClient(conn)
```

---

## 1. 用户管理

### 1.1 ListUsers — 用户列表

```protobuf
rpc ListUsers(ListUsersReq) returns (ListUsersResp);
```

**请求:**
```protobuf
message ListUsersReq {
  int32 page = 1;       // 页码，默认 1
  int32 page_size = 2;  // 每页条数，默认 20
  string keyword = 3;   // username / realName 模糊搜索
}
```

**响应:**
```protobuf
message ListUsersResp {
  repeated UserItem list = 1;
  PageInfo page_info = 2;
}

message UserItem {
  int64 id = 1;
  string username = 2;
  string real_name = 3;
  string department = 4;
  string role = 5;
  int32 status = 6;        // 0=禁用 1=启用
  string last_login = 7;
}

message PageInfo {
  int32 page = 1;
  int32 page_size = 2;
  int64 total = 3;
}
```

**gRPC 错误码:** `Internal` — 数据库错误

---

### 1.2 CreateUser — 创建用户

```protobuf
rpc CreateUser(CreateUserReq) returns (CreateUserResp);
```

**调用方需校验:** operator 具有 admin 角色。

**请求:**
```protobuf
message CreateUserReq {
  string username = 1;     // 必填 3-50 字符，唯一
  string password = 2;     // 必填 6-100 字符，bcrypt 哈希
  string real_name = 3;    // 必填
  string department = 4;
  string role = 5;         // 角色编码，如 admin / director
  int64 operator_id = 6;   // 操作人 ID
}
```

**响应:**
```protobuf
message CreateUserResp {
  int64 id = 1;
  string username = 2;
}
```

**gRPC 错误码:** `Internal` — 用户名重复 / 角色不存在 / 数据库错误

---

### 1.3 UpdateUser — 更新用户

```protobuf
rpc UpdateUser(UpdateUserReq) returns (UpdateUserResp);
```

**调用方需校验:** admin 或 user.id == operator.id。

**请求:**
```protobuf
message UpdateUserReq {
  int64 id = 1;
  string password = 2;              // 可选，非空时更新
  string real_name = 3;
  string department = 4;
  string role = 5;                  // 可选，非空时替换角色
  optional int32 status = 6;        // 0=禁用 1=启用
  int64 operator_id = 7;
  repeated string operator_roles = 8;
}
```

**响应:**
```protobuf
message UpdateUserResp {}
```

**gRPC 错误码:** `Internal` — 用户不存在 / 角色不存在 / 数据库错误

---

### 1.4 DeleteUser — 删除用户

```protobuf
rpc DeleteUser(DeleteUserReq) returns (DeleteUserResp);
```

**调用方需校验:** operator 具有 admin 角色。  
**服务端约束:** 不删除 operator 自身、不删除 is_system 管理员。

**请求:**
```protobuf
message DeleteUserReq {
  repeated int64 ids = 1;  // 可单可多
  int64 operator_id = 2;
}
```

**响应:**
```protobuf
message DeleteUserResp {
  int32 deleted_count = 1;  // 实际删除数（跳过的不计）
}
```

---

### 1.5 BatchDeleteUsers — 批量删除

```protobuf
rpc BatchDeleteUsers(BatchDeleteUsersReq) returns (BatchDeleteUsersResp);
```

**调用方需校验:** ids 数量 ≤ 100，operator 具有 admin 角色。

**请求:**
```protobuf
message BatchDeleteUsersReq {
  repeated int64 ids = 1;  // 上限 100
  int64 operator_id = 2;
}
```

**响应:**
```protobuf
message BatchDeleteUsersResp {
  int32 deleted_count = 1;
}
```

---

## 2. 用户档案

### 2.1 GetProfile — 查询档案

```protobuf
rpc GetProfile(GetProfileReq) returns (GetProfileResp);
```

**请求:**
```protobuf
message GetProfileReq {
  int64 user_id = 1;
}
```

**响应:**
```protobuf
message GetProfileResp {
  int64 id = 1;
  string username = 2;
  string real_name = 3;
  string department = 4;
  string phone = 5;
  string email = 6;
  string avatar = 7;
  string role = 8;
}
```

---

### 2.2 UpdateProfile — 更新档案

```protobuf
rpc UpdateProfile(UpdateProfileReq) returns (UpdateProfileResp);
```

**请求:**
```protobuf
message UpdateProfileReq {
  int64 user_id = 1;
  string department = 2;
  string phone = 3;
  string email = 4;
  string avatar = 5;
}
```

**响应:**
```protobuf
message UpdateProfileResp {}
```

档案不存在时自动创建。

---

## 3. 组织架构

### 3.1 GetOrgTree — 组织树

```protobuf
rpc GetOrgTree(GetOrgTreeReq) returns (GetOrgTreeResp);
```

**请求:**
```protobuf
message GetOrgTreeReq {}
```

**响应:**
```protobuf
message GetOrgTreeResp {
  repeated OrgTreeNode nodes = 1;
}

message OrgTreeNode {
  int64 id = 1;
  string name = 2;
  string type = 3;          // zone / base / brigade
  int32 sort_order = 4;
  int64 parent_id = 5;
  repeated OrgTreeNode children = 6;
}
```

---

### 3.2 CreateOrganization — 创建节点

```protobuf
rpc CreateOrganization(CreateOrganizationReq) returns (CreateOrganizationResp);
```

**调用方需校验:** operator 具有 admin 角色。

**请求:**
```protobuf
message CreateOrganizationReq {
  string name = 1;          // 必填
  int64 parent_id = 2;      // 顶级节点为 0
  string type = 3;          // 必填 zone / base / brigade
  int32 sort_order = 4;
  int64 operator_id = 5;
}
```

**响应:**
```protobuf
message CreateOrganizationResp {
  int64 id = 1;
  string name = 2;
}
```

---

### 3.3 DeleteOrganization — 删除节点

```protobuf
rpc DeleteOrganization(DeleteOrganizationReq) returns (DeleteOrganizationResp);
```

级联软删除所有子孙节点（WITH RECURSIVE CTE）。

**请求:**
```protobuf
message DeleteOrganizationReq {
  int64 id = 1;
  int64 operator_id = 2;
}
```

**响应:**
```protobuf
message DeleteOrganizationResp {}
```

---

## 错误处理

gRPC 错误通过 `status.Status` 返回：

| gRPC Code | 场景 |
|-----------|------|
| `Internal` | 数据库错误 / 业务校验失败（消息包含中文描述） |
| `Unavailable` | 服务不可用（etcd 健康检查失败） |

> 调用方应根据 `status.Code(err)` 和 `status.Message(err)` 区分具体错误。

---

## RPC 汇总

| RPC | 请求 | 响应 | 说明 |
|-----|------|------|------|
| ListUsers | page,pageSize,keyword | list + PageInfo | 分页+模糊搜索 |
| CreateUser | username,password,realName,role | id,username | 唯一校验, bcrypt |
| UpdateUser | id + 可选字段 | — | 角色替换, 权限缓存清理 |
| DeleteUser | ids[], operatorId | deletedCount | 防自删, 防删系统管理员 |
| BatchDeleteUsers | ids[] (≤100) | deletedCount | 同 DeleteUser |
| GetProfile | userId | 完整档案 | 联查 t_user+profile+role |
| UpdateProfile | userId + 可选字段 | — | 不存在时自动创建 |
| GetOrgTree | — | 树形节点组 | 内存构建树 |
| CreateOrganization | name,type,parentId | id,name | 校验类型+父节点 |
