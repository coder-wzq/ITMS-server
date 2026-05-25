package svcuser

import (
	"gorm.io/gorm"

	rediscli "itms-server/pkg/redis"
	"itms-server/services/svc-user/internal/logic"
	"itms-server/services/svc-user/internal/server"
)

// NewServer creates the svc-user gRPC server implementation bound to the
// given database and cache connections. Callers register it with a
// *grpc.Server, listen, and optionally register with etcd.
//
//	grpcSrv := grpc.NewServer()
//	userpb.RegisterUserServiceServer(grpcSrv, svcuser.NewServer(db, rdb))
func NewServer(db *gorm.DB, rdb *rediscli.Client) *server.GRPCServer {
	userLogic := logic.NewUserLogic(db, rdb)
	orgLogic := logic.NewOrganizationLogic(db)
	return server.NewGRPCServer(userLogic, orgLogic)
}
