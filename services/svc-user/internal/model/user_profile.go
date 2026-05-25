package model

import "itms-server/pkg/db"

// UserProfile corresponds to sch_user.t_user_profile (1:1 with sch_auth.t_user).
type UserProfile struct {
	db.BaseModel
	UserID     int64  `gorm:"column:user_id;uniqueIndex:uk_user_profile_uid;not null" json:"user_id"`
	Department string `gorm:"column:department;type:varchar(100)" json:"department"`
	Phone      string `gorm:"column:phone;type:varchar(30)" json:"phone"`
	Email      string `gorm:"column:email;type:varchar(100)" json:"email"`
	Avatar     string `gorm:"column:avatar;type:varchar(500)" json:"avatar"`
}

func (UserProfile) TableName() string {
	return "sch_user.t_user_profile"
}
