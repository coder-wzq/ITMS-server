#!/bin/bash
# PostgreSQL 初始化脚本 — Docker entrypoint 自动执行
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" << 'SQL'

-- 创建14个Schema
CREATE SCHEMA IF NOT EXISTS sch_auth;
CREATE SCHEMA IF NOT EXISTS sch_user;
CREATE SCHEMA IF NOT EXISTS sch_config;
CREATE SCHEMA IF NOT EXISTS sch_director;
CREATE SCHEMA IF NOT EXISTS sch_task;
CREATE SCHEMA IF NOT EXISTS sch_planning;
CREATE SCHEMA IF NOT EXISTS sch_mpt;
CREATE SCHEMA IF NOT EXISTS sch_situation;
CREATE SCHEMA IF NOT EXISTS sch_voice;
CREATE SCHEMA IF NOT EXISTS sch_dict;
CREATE SCHEMA IF NOT EXISTS sch_agent;
CREATE SCHEMA IF NOT EXISTS sch_report;
CREATE SCHEMA IF NOT EXISTS sch_record;
CREATE SCHEMA IF NOT EXISTS sch_admin;

-- 基础表
CREATE TABLE IF NOT EXISTS sch_auth.t_user (
    id BIGINT PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    real_name VARCHAR(100) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    last_login_at TIMESTAMPTZ,
    last_login_ip VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_user_username ON sch_auth.t_user(username) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS sch_auth.t_role (
    id BIGINT PRIMARY KEY,
    role_code VARCHAR(50) NOT NULL,
    role_name VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_role_code ON sch_auth.t_role(role_code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS sch_auth.t_user_role (
    id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS sch_auth.t_role_permission (
    id BIGINT PRIMARY KEY,
    role_id BIGINT NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS sch_config.t_system_config (
    id BIGINT PRIMARY KEY,
    config_key VARCHAR(100) NOT NULL,
    config_value TEXT NOT NULL DEFAULT '',
    description TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS sch_dict.t_force (
    id BIGINT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(50) NOT NULL,
    color VARCHAR(10) NOT NULL DEFAULT '#000000',
    description TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS sch_admin.t_operation_log (
    id BIGINT PRIMARY KEY,
    operator_id BIGINT,
    operator_name VARCHAR(100) DEFAULT '',
    operation_type VARCHAR(50) NOT NULL,
    resource VARCHAR(200) NOT NULL DEFAULT '',
    request_params TEXT DEFAULT '',
    response_code INT DEFAULT 0,
    ip VARCHAR(50) DEFAULT '',
    duration INT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 默认角色
INSERT INTO sch_auth.t_role (id, role_code, role_name, description, is_system, created_at, updated_at) VALUES
  (1000000000000001, 'admin', '系统管理员', '拥有系统全部权限', true, NOW(), NOW()),
  (1000000000000002, 'director', '导调员', '负责任务导调与控制', true, NOW(), NOW()),
  (1000000000000003, 'operator', '操作员', '负责日常操作', true, NOW(), NOW()),
  (1000000000000004, 'viewer', '观察员', '仅查看权限', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 权限
INSERT INTO sch_auth.t_role_permission (id, role_id, resource, action, created_at, updated_at) VALUES
  (1000000000000101, 1000000000000001, '*', '*', NOW(), NOW()),
  (1000000000000102, 1000000000000002, 'sessions', 'manage', NOW(), NOW()),
  (1000000000000103, 1000000000000002, 'events', 'inject', NOW(), NOW()),
  (1000000000000104, 1000000000000003, 'tasks', 'read', NOW(), NOW()),
  (1000000000000105, 1000000000000003, 'tasks', 'write', NOW(), NOW()),
  (1000000000000106, 1000000000000004, '*', 'read', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 默认管理员 (admin / admin123)
INSERT INTO sch_auth.t_user (id, username, password_hash, real_name, status, created_at, updated_at) VALUES (
  10000000000000001,
  'admin',
  '$2a$12$LJ3m4ys3LkBCVxJGqOjqOeFpXtB6PcHJY5VYvGqER2KXOjxJqDRKq',
  '系统管理员',
  1,
  NOW(), NOW()
) ON CONFLICT (id) DO NOTHING;

INSERT INTO sch_auth.t_user_role (id, user_id, role_id, created_at, updated_at) VALUES
  (10000000000000001, 10000000000000001, 1000000000000001, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 默认阵营
INSERT INTO sch_dict.t_force (id, name, code, color, description, created_at, updated_at) VALUES
  (3000000000000001, '红方', 'red', '#EF4444', '红方阵营', NOW(), NOW()),
  (3000000000000002, '蓝方', 'blue', '#3B82F6', '蓝方阵营', NOW(), NOW()),
  (3000000000000003, '白方', 'white', '#FFFFFF', '白方阵营', NOW(), NOW()),
  (3000000000000004, '绿方', 'green', '#22C55E', '绿方阵营', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

SQL
echo "Database initialization complete"
