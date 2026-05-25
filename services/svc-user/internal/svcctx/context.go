package svcctx

import (
	"gorm.io/gorm"

	jwtpkg "itms-server/pkg/jwt"
	"itms-server/pkg/redis"
)

// ServiceContext holds shared resources for svc-user.
type ServiceContext struct {
	DB       *gorm.DB
	Redis    *redis.Client
	JwtCfg   *jwtpkg.Config
}

// New creates a ServiceContext with the given dependencies.
func New(db *gorm.DB, rdb *redis.Client, jwtCfg *jwtpkg.Config) *ServiceContext {
	return &ServiceContext{
		DB:     db,
		Redis:  rdb,
		JwtCfg: jwtCfg,
	}
}
