-- ITMS Seed Data: Default roles, configuration, and forces
-- Uses Snowflake-like IDs based on a fixed epoch base (2024-01-01)

-- ============================================
-- CSCI-01: Default Roles
-- ============================================
INSERT INTO sch_auth.t_role (id, role_code, role_name, description, is_system, created_at, updated_at)
VALUES
  (1000000000000001, 'admin', '系统管理员', '拥有系统全部权限', true, NOW(), NOW()),
  (1000000000000002, 'director', '导调员', '负责任务导调与控制', true, NOW(), NOW()),
  (1000000000000003, 'operator', '操作员', '负责日常操作', true, NOW(), NOW()),
  (1000000000000004, 'viewer', '观察员', '仅查看权限', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Admin role: all permissions (*:*)
INSERT INTO sch_auth.t_role_permission (id, role_id, resource, action, created_at, updated_at)
VALUES
  (1000000000000101, 1000000000000001, '*', '*', NOW(), NOW()),
  (1000000000000102, 1000000000000002, 'sessions', 'manage', NOW(), NOW()),
  (1000000000000103, 1000000000000002, 'events', 'inject', NOW(), NOW()),
  (1000000000000104, 1000000000000003, 'tasks', 'read', NOW(), NOW()),
  (1000000000000105, 1000000000000003, 'tasks', 'write', NOW(), NOW()),
  (1000000000000106, 1000000000000004, '*', 'read', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- CSCI-01: Default Admin User (password: admin123)
-- bcrypt cost=12 hash for 'admin123'
-- ============================================
INSERT INTO sch_auth.t_user (id, username, password_hash, real_name, status, created_at, updated_at)
VALUES (
  10000000000000001,
  'admin',
  '$2a$12$LJ3m4ys3LkBCVxJGqOjqOeFpXtB6PcHJY5VYvGqER2KXOjxJqDRKq', -- admin123
  '系统管理员',
  1,
  NOW(),
  NOW()
) ON CONFLICT (id) DO NOTHING;

-- Assign admin role to admin user
INSERT INTO sch_auth.t_user_role (id, user_id, role_id, created_at, updated_at)
VALUES (10000000000000001, 10000000000000001, 1000000000000001, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- CSCI-03: Default System Configuration
-- ============================================
INSERT INTO sch_config.t_system_config (id, config_key, config_value, description, created_at, updated_at)
VALUES
  (2000000000000001, 'system.name', '"ITMS综合训练管理系统"', '系统名称', NOW(), NOW()),
  (2000000000000002, 'safety.min_altitude', '100', '飞行安全最低高度(m)', NOW(), NOW()),
  (2000000000000003, 'safety.max_altitude', '12000', '飞行安全最高高度(m)', NOW(), NOW()),
  (2000000000000004, 'safety.max_speed', '2500', '飞行安全最大速度(km/h)', NOW(), NOW()),
  (2000000000000005, 'safety.min_fuel', '500', '飞行安全最低燃油(kg)', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- CSCI-11: Default Forces (阵营)
-- ============================================
INSERT INTO sch_dict.t_force (id, name, code, color, description, created_at, updated_at)
VALUES
  (3000000000000001, '红方', 'red', '#EF4444', '红方阵营', NOW(), NOW()),
  (3000000000000002, '蓝方', 'blue', '#3B82F6', '蓝方阵营', NOW(), NOW()),
  (3000000000000003, '白方', 'white', '#FFFFFF', '白方阵营（导演方）', NOW(), NOW()),
  (3000000000000004, '绿方', 'green', '#22C55E', '绿方阵营', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
