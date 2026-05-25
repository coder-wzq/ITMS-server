package logic

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"itms-server/pkg/db"
	"itms-server/pkg/redis"
	"itms-server/services/svc-user/internal/model"
	"itms-server/services/svc-user/internal/types"
)

// UserLogic handles user management business logic.
type UserLogic struct {
	gdb  *gorm.DB
	rdb  *redis.Client
}

func NewUserLogic(gdb *gorm.DB, rdb *redis.Client) *UserLogic {
	return &UserLogic{gdb: gdb, rdb: rdb}
}

// List returns a paginated user list with keyword search.
func (l *UserLogic) List(ctx context.Context, req *types.UserListReq) ([]types.UserListItem, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	type row struct {
		ID         int64  `gorm:"column:id"`
		Username   string `gorm:"column:username"`
		RealName   string `gorm:"column:real_name"`
		Department string `gorm:"column:department"`
		RoleName   string `gorm:"column:role_name"`
		Status     int    `gorm:"column:status"`
		LastLogin  *string `gorm:"column:last_login_at"`
	}

	baseQuery := l.gdb.WithContext(ctx).Table("sch_auth.t_user AS u").
		Select(`
			u.id, u.username, u.real_name, u.status, u.last_login_at,
			COALESCE(p.department, '') AS department,
			COALESCE(r.role_name, '') AS role_name
		`).
		Joins("LEFT JOIN sch_user.t_user_profile AS p ON u.id = p.user_id AND p.deleted_at IS NULL").
		Joins("LEFT JOIN sch_auth.t_user_role AS ur ON u.id = ur.user_id AND ur.deleted_at IS NULL").
		Joins("LEFT JOIN sch_auth.t_role AS r ON ur.role_id = r.id AND r.deleted_at IS NULL").
		Where("u.deleted_at IS NULL")

	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		baseQuery = baseQuery.Where("u.username ILIKE ? OR u.real_name ILIKE ?", kw, kw)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	var rows []row
	if err := baseQuery.
		Order("u.created_at DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("query users: %w", err)
	}

	list := make([]types.UserListItem, len(rows))
	for i, r := range rows {
		lastLogin := ""
		if r.LastLogin != nil {
			lastLogin = *r.LastLogin
		}
		list[i] = types.UserListItem{
			ID:         r.ID,
			Username:   r.Username,
			RealName:   r.RealName,
			Department: r.Department,
			Role:       r.RoleName,
			Status:     r.Status,
			LastLogin:  lastLogin,
		}
	}

	return list, total, nil
}

// Create creates a new user with profile and role assignment.
func (l *UserLogic) Create(ctx context.Context, req *types.CreateUserReq) (*types.CreateUserResp, error) {
	// Check username uniqueness
	var existing model.AuthUser
	err := l.gdb.WithContext(ctx).Where("username = ?", req.Username).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("username already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("check username: %w", err)
	}

	// Find target role
	var role model.AuthRole
	if err := l.gdb.WithContext(ctx).Where("role_code = ?", req.Role).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("role %s not found", req.Role)
		}
		return nil, fmt.Errorf("find role: %w", err)
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	userID := db.GetSnowflake().NextID()

	user := &model.AuthUser{
		BaseModel:    db.BaseModel{ID: userID},
		Username:     req.Username,
		PasswordHash: string(hash),
		RealName:     req.RealName,
		Status:       1,
	}

	profile := &model.UserProfile{
		BaseModel:  db.BaseModel{ID: db.GetSnowflake().NextID()},
		UserID:     userID,
		Department: req.Department,
	}

	userRole := &model.AuthUserRole{
		BaseModel: db.BaseModel{ID: db.GetSnowflake().NextID()},
		UserID:    userID,
		RoleID:    role.ID,
	}

	if err := l.gdb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("insert user: %w", err)
		}
		if err := tx.Create(profile).Error; err != nil {
			return fmt.Errorf("insert profile: %w", err)
		}
		if err := tx.Create(userRole).Error; err != nil {
			return fmt.Errorf("insert user_role: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Clear permission cache
	l.rdb.Del(ctx, fmt.Sprintf(redis.KeyPermission, userID))

	return &types.CreateUserResp{
		ID:       userID,
		Username: req.Username,
	}, nil
}

// Update updates a user's info, password, role, and status.
func (l *UserLogic) Update(ctx context.Context, id int64, req *types.UpdateUserReq) error {
	var user model.AuthUser
	if err := l.gdb.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("find user: %w", err)
	}

	if err := l.gdb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{}
		if req.RealName != "" {
			updates["real_name"] = req.RealName
		}
		if req.Status != nil {
			updates["status"] = *req.Status
		}
		if req.Password != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("hash password: %w", err)
			}
			updates["password_hash"] = string(hash)
		}
		if len(updates) > 0 {
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				return fmt.Errorf("update user: %w", err)
			}
		}

		// Upsert profile
		profileUpdates := map[string]interface{}{
			"department": req.Department,
		}
		if err := tx.Model(&model.UserProfile{}).
			Where("user_id = ?", id).
			Updates(profileUpdates).Error; err != nil {
			return fmt.Errorf("update profile: %w", err)
		}

		// Update role if provided
		if req.Role != "" {
			var role model.AuthRole
			if err := tx.Where("role_code = ?", req.Role).First(&role).Error; err != nil {
				return fmt.Errorf("role %s not found", req.Role)
			}
			// Delete old roles
			if err := tx.Where("user_id = ?", id).Delete(&model.AuthUserRole{}).Error; err != nil {
				return fmt.Errorf("delete old roles: %w", err)
			}
			// Insert new role
			newUR := &model.AuthUserRole{
				BaseModel: db.BaseModel{ID: db.GetSnowflake().NextID()},
				UserID:    id,
				RoleID:    role.ID,
			}
			if err := tx.Create(newUR).Error; err != nil {
				return fmt.Errorf("insert new role: %w", err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Clear permission cache
	l.rdb.Del(ctx, fmt.Sprintf(redis.KeyPermission, id))
	return nil
}

// Delete soft-deletes a user and all related records.
func (l *UserLogic) Delete(ctx context.Context, ids []int64, operatorID int64) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	deleted := 0
	for _, id := range ids {
		// Cannot delete self
		if id == operatorID {
			continue
		}

		// Check if system admin
		var count int64
		l.gdb.WithContext(ctx).Table("sch_auth.t_user_role AS ur").
			Joins("JOIN sch_auth.t_role AS r ON ur.role_id = r.id AND r.deleted_at IS NULL").
			Where("ur.user_id = ? AND r.is_system = true AND ur.deleted_at IS NULL", id).
			Count(&count)
		if count > 0 {
			continue
		}

		if err := l.gdb.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("id = ?", id).Delete(&model.AuthUser{}).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ?", id).Delete(&model.UserProfile{}).Error; err != nil {
				return err
			}
			if err := tx.Where("user_id = ?", id).Delete(&model.AuthUserRole{}).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			return deleted, err
		}

		// Clear Redis caches
		l.rdb.Del(ctx, fmt.Sprintf(redis.KeyAuthToken, id))
		l.rdb.Del(ctx, fmt.Sprintf(redis.KeyPermission, id))
		deleted++
	}

	return deleted, nil
}

// GetProfile returns the user's profile.
func (l *UserLogic) GetProfile(ctx context.Context, userID int64) (*types.UserProfileResp, error) {
	type row struct {
		ID         int64  `gorm:"column:id"`
		Username   string `gorm:"column:username"`
		RealName   string `gorm:"column:real_name"`
		Department string `gorm:"column:department"`
		Phone      string `gorm:"column:phone"`
		Email      string `gorm:"column:email"`
		Avatar     string `gorm:"column:avatar"`
		RoleName   string `gorm:"column:role_name"`
	}

	var r row
	err := l.gdb.WithContext(ctx).Table("sch_auth.t_user AS u").
		Select(`
			u.id, u.username, u.real_name,
			COALESCE(p.department, '') AS department,
			COALESCE(p.phone, '') AS phone,
			COALESCE(p.email, '') AS email,
			COALESCE(p.avatar, '') AS avatar,
			COALESCE(r.role_name, '') AS role_name
		`).
		Joins("LEFT JOIN sch_user.t_user_profile AS p ON u.id = p.user_id AND p.deleted_at IS NULL").
		Joins("LEFT JOIN sch_auth.t_user_role AS ur ON u.id = ur.user_id AND ur.deleted_at IS NULL").
		Joins("LEFT JOIN sch_auth.t_role AS r ON ur.role_id = r.id AND r.deleted_at IS NULL").
		Where("u.id = ? AND u.deleted_at IS NULL", userID).
		Scan(&r).Error
	if err != nil {
		return nil, fmt.Errorf("query profile: %w", err)
	}

	return &types.UserProfileResp{
		ID:         r.ID,
		Username:   r.Username,
		RealName:   r.RealName,
		Department: r.Department,
		Phone:      r.Phone,
		Email:      r.Email,
		Avatar:     r.Avatar,
		Role:       r.RoleName,
	}, nil
}

// UpdateProfile updates the current user's profile fields.
func (l *UserLogic) UpdateProfile(ctx context.Context, userID int64, req *types.UpdateProfileReq) error {
	updates := map[string]interface{}{
		"department": req.Department,
		"phone":      req.Phone,
		"email":      req.Email,
		"avatar":     req.Avatar,
	}

	result := l.gdb.WithContext(ctx).Model(&model.UserProfile{}).
		Where("user_id = ?", userID).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update profile: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		// Profile doesn't exist yet, create it
		profile := &model.UserProfile{
			BaseModel:  db.BaseModel{ID: db.GetSnowflake().NextID()},
			UserID:     userID,
			Department: req.Department,
			Phone:      req.Phone,
			Email:      req.Email,
			Avatar:     req.Avatar,
		}
		if err := l.gdb.WithContext(ctx).Create(profile).Error; err != nil {
			return fmt.Errorf("create profile: %w", err)
		}
	}
	return nil
}
