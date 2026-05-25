# CSCI-02 用户管理 — HTTP 接口文档

**标识号:** ITMS-API-HTTP-V1.0  
**Base URL:** `http://{host}:8080`  
**认证方式:** `Authorization: Bearer {accessToken}`（除 `/health` 外所有接口必需）  
**Content-Type:** `application/json; charset=utf-8`

---

## 统一响应格式

### 成功

```json
{ "code": 0, "message": "success", "data": {...}, "requestId": "" }
```

### 分页

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```

### 错误

```json
{ "code": 2101, "message": "username already exists", "data": null }
```

---

## 1. 用户管理

### 1.1 用户列表（分页+搜索）

**USER-001**

```
GET /api/users?page=1&pageSize=20&keyword=admin
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| pageSize | int | 否 | 每页条数，默认 20，最大 100 |
| keyword | string | 否 | 按 username / realName 模糊搜索 |

**响应 data.list:**

```json
[{
  "id": 7234567890123456,
  "username": "admin",
  "realName": "管理员",
  "department": "指挥中心",
  "role": "系统管理员",
  "status": 1,
  "lastLogin": "2026-05-25 10:30:00"
}]
```

---

### 1.2 创建用户

**USER-002** | 角色：admin

```
POST /api/users
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，3-50 字符，唯一 |
| password | string | 是 | 密码，6-100 字符 |
| realName | string | 是 | 真实姓名 |
| department | string | 否 | 部门 |
| role | string | 是 | 角色编码，如 admin / director |

**请求:**
```json
{
  "username": "zhangsan",
  "password": "123456",
  "realName": "张三",
  "department": "作训科",
  "role": "director"
}
```

**响应:**
```json
{ "code": 0, "data": { "id": 7234567890123457, "username": "zhangsan" } }
```

**错误码:** 2101（用户名重复）、1001（参数错误）、1005（权限不足）

---

### 1.3 更新用户

**USER-003** | admin 可改任意用户，普通用户仅可改自身

```
PUT /api/users/{id}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| password | string | 否 | 新密码，非空时更新 |
| realName | string | 否 | 真实姓名 |
| department | string | 否 | 部门 |
| role | string | 否 | 角色编码，非空时替换 |
| status | int | 否 | 0=禁用 1=启用 |

**请求:**
```json
{
  "realName": "张三丰",
  "department": "训练部",
  "status": 1
}
```

**响应:** `{ "code": 0, "data": null }`

---

### 1.4 删除用户

**USER-004** | 角色：admin

```
DELETE /api/users/{id}
```

约束：不可删除自身、不可删除 is_system 管理员。

**响应:** `{ "code": 0, "data": null }`

**错误码:** 2103（不可删除自身）、2104（不可删除系统管理员）

---

### 1.5 批量删除用户

**角色：admin**

```
POST /api/users/batch-delete
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ids | int64[] | 是 | 用户 ID 数组，上限 100 |

**请求:**
```json
{ "ids": [7234567890123457, 7234567890123458] }
```

**响应:**
```json
{ "code": 0, "data": { "deletedCount": 2 } }
```

---

## 2. 用户档案

### 2.1 查询当前用户档案

**USR-001**

```
GET /api/users/profile
```

**响应 data:**
```json
{
  "id": 7234567890123456,
  "username": "admin",
  "realName": "管理员",
  "department": "指挥中心",
  "phone": "13800138000",
  "email": "admin@example.com",
  "avatar": "/avatars/admin.png",
  "role": "系统管理员"
}
```

---

### 2.2 更新当前用户档案

**USR-002**

```
PUT /api/users/profile
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| department | string | 否 | 部门 |
| phone | string | 否 | 电话 |
| email | string | 否 | 邮箱 |
| avatar | string | 否 | 头像 URL |

**请求:**
```json
{
  "department": "作训科",
  "phone": "13900139000",
  "email": "zhangsan@example.com",
  "avatar": "/avatars/zhangsan.png"
}
```

**响应:** `{ "code": 0, "data": null }`

---

## 3. 组织架构

### 3.1 组织树

**USR-006**

```
GET /api/organizations/tree
```

**响应 data:**
```json
[{
  "id": 1,
  "name": "东部战区",
  "type": "zone",
  "sortOrder": 1,
  "parentId": 0,
  "children": [{
    "id": 2,
    "name": "某基地",
    "type": "base",
    "sortOrder": 1,
    "parentId": 1,
    "children": [{
      "id": 3,
      "name": "某旅",
      "type": "brigade",
      "sortOrder": 1,
      "parentId": 2,
      "children": []
    }]
  }]
}]
```

---

### 3.2 创建组织节点

**USR-007** | 角色：admin

```
POST /api/organizations
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 名称 |
| type | string | 是 | zone / base / brigade |
| parentId | int64 | 否 | 父节点 ID，顶级节点不填 |
| sortOrder | int | 否 | 排序 |

**请求:**
```json
{ "name": "某旅", "type": "brigade", "parentId": 2, "sortOrder": 1 }
```

**响应:**
```json
{ "code": 0, "data": { "id": 4, "name": "某旅" } }
```

---

### 3.3 删除组织节点

**USR-008** | 角色：admin

```
DELETE /api/organizations/{id}
```

级联软删除所有子孙节点。

**响应:** `{ "code": 0, "data": null }`

---

## 错误码速查

| code | 说明 |
|------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | Token 无效或过期 |
| 1003 | 缺少 Authorization 头 |
| 1005 | 权限不足 |
| 2001 | 用户不存在 |
| 2101 | 用户名已存在 |
| 2102 | 批量操作超限（上限 100） |
| 2103 | 不可删除自身 |
| 2104 | 不可删除系统管理员 |
| 1500 | 服务器内部错误 |

---

## 接口汇总

| 编号 | 方法 | 路径 | 权限 |
|------|------|------|------|
| USER-001 | GET | `/api/users` | 登录用户 |
| USER-002 | POST | `/api/users` | admin |
| USER-003 | PUT | `/api/users/{id}` | admin / 自身 |
| USER-004 | DELETE | `/api/users/{id}` | admin |
| — | POST | `/api/users/batch-delete` | admin |
| USR-001 | GET | `/api/users/profile` | 登录用户 |
| USR-002 | PUT | `/api/users/profile` | 登录用户 |
| USR-006 | GET | `/api/organizations/tree` | 登录用户 |
| USR-007 | POST | `/api/organizations` | admin |
| USR-008 | DELETE | `/api/organizations/{id}` | admin |
