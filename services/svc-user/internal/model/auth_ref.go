package model

import "itms-server/pkg/db"

// AuthUser references sch_auth.t_user (read-only from svc-user).
type AuthUser struct {
	db.BaseModel
	Username     string `gorm:"column:username;type:varchar(50);uniqueIndex:uk_user_username;not null" json:"username"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`
	RealName     string `gorm:"column:real_name;type:varchar(50)" json:"real_name"`
	Status       int    `gorm:"column:status;default:1" json:"status"`
	LastLoginAt  *string `gorm:"column:last_login_at" json:"last_login_at"`
	LastLoginIP  string  `gorm:"column:last_login_ip;type:varchar(45)" json:"last_login_ip"`
}

func (AuthUser) TableName() string {
	return "sch_auth.t_user"
}

// AuthRole references sch_auth.t_role.
type AuthRole struct {
	db.BaseModel
	RoleCode    string `gorm:"column:role_code;type:varchar(50);uniqueIndex:uk_role_code;not null" json:"role_code"`
	RoleName    string `gorm:"column:role_name;type:varchar(50);not null" json:"role_name"`
	Description string `gorm:"column:description;type:varchar(200)" json:"description"`
	IsSystem    bool   `gorm:"column:is_system;default:false" json:"is_system"`
}

func (AuthRole) TableName() string {
	return "sch_auth.t_role"
}

// AuthUserRole references sch_auth.t_user_role.
type AuthUserRole struct {
	db.BaseModel
	UserID int64 `gorm:"column:user_id;uniqueIndex:uk_user_role;not null" json:"user_id"`
	RoleID int64 `gorm:"column:role_id;uniqueIndex:uk_user_role;not null" json:"role_id"`
}

func (AuthUserRole) TableName() string {
	return "sch_auth.t_user_role"
}
