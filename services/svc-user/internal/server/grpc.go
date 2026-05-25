package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userpb "itms-server/api/proto/user"
	"itms-server/services/svc-user/internal/logic"
	"itms-server/services/svc-user/internal/model"
	"itms-server/services/svc-user/internal/types"
)

// GRPCServer implements the generated UserServiceServer.
type GRPCServer struct {
	userpb.UnimplementedUserServiceServer
	userLogic *logic.UserLogic
	orgLogic  *logic.OrganizationLogic
}

func NewGRPCServer(ul *logic.UserLogic, ol *logic.OrganizationLogic) *GRPCServer {
	return &GRPCServer{userLogic: ul, orgLogic: ol}
}

// ---------------------------------------------------------------------------
// User CRUD
// ---------------------------------------------------------------------------

func (s *GRPCServer) ListUsers(ctx context.Context, req *userpb.ListUsersReq) (*userpb.ListUsersResp, error) {
	list, total, err := s.userLogic.List(ctx, &types.UserListReq{
		Page: int(req.Page), PageSize: int(req.PageSize), Keyword: req.Keyword,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &userpb.ListUsersResp{
		PageInfo: &userpb.PageInfo{Page: req.Page, PageSize: req.PageSize, Total: total},
	}
	for _, u := range list {
		resp.List = append(resp.List, &userpb.UserItem{
			Id: u.ID, Username: u.Username, RealName: u.RealName,
			Department: u.Department, Role: u.Role, Status: int32(u.Status),
			LastLogin: u.LastLogin,
		})
	}
	return resp, nil
}

func (s *GRPCServer) CreateUser(ctx context.Context, req *userpb.CreateUserReq) (*userpb.CreateUserResp, error) {
	r, err := s.userLogic.Create(ctx, &types.CreateUserReq{
		Username: req.Username, Password: req.Password,
		RealName: req.RealName, Department: req.Department, Role: req.Role,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &userpb.CreateUserResp{Id: r.ID, Username: r.Username}, nil
}

func (s *GRPCServer) UpdateUser(ctx context.Context, req *userpb.UpdateUserReq) (*userpb.UpdateUserResp, error) {
	r := &types.UpdateUserReq{
		Password: req.Password, RealName: req.RealName,
		Department: req.Department, Role: req.Role,
	}
	if req.Status != nil {
		v := int(*req.Status)
		r.Status = &v
	}
	if err := s.userLogic.Update(ctx, req.Id, r); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &userpb.UpdateUserResp{}, nil
}

func (s *GRPCServer) DeleteUser(ctx context.Context, req *userpb.DeleteUserReq) (*userpb.DeleteUserResp, error) {
	n, err := s.userLogic.Delete(ctx, req.Ids, req.OperatorId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &userpb.DeleteUserResp{DeletedCount: int32(n)}, nil
}

func (s *GRPCServer) BatchDeleteUsers(ctx context.Context, req *userpb.BatchDeleteUsersReq) (*userpb.BatchDeleteUsersResp, error) {
	n, err := s.userLogic.Delete(ctx, req.Ids, req.OperatorId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &userpb.BatchDeleteUsersResp{DeletedCount: int32(n)}, nil
}

// ---------------------------------------------------------------------------
// Profile
// ---------------------------------------------------------------------------

func (s *GRPCServer) GetProfile(ctx context.Context, req *userpb.GetProfileReq) (*userpb.GetProfileResp, error) {
	p, err := s.userLogic.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &userpb.GetProfileResp{
		Id: p.ID, Username: p.Username, RealName: p.RealName,
		Department: p.Department, Phone: p.Phone, Email: p.Email,
		Avatar: p.Avatar, Role: p.Role,
	}, nil
}

func (s *GRPCServer) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileReq) (*userpb.UpdateProfileResp, error) {
	if err := s.userLogic.UpdateProfile(ctx, req.UserId, &types.UpdateProfileReq{
		Department: req.Department, Phone: req.Phone,
		Email: req.Email, Avatar: req.Avatar,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &userpb.UpdateProfileResp{}, nil
}

// ---------------------------------------------------------------------------
// Organization
// ---------------------------------------------------------------------------

func (s *GRPCServer) GetOrgTree(ctx context.Context, req *userpb.GetOrgTreeReq) (*userpb.GetOrgTreeResp, error) {
	nodes, err := s.orgLogic.Tree(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &userpb.GetOrgTreeResp{}
	for _, n := range nodes {
		resp.Nodes = append(resp.Nodes, toProtoTreeNode(n))
	}
	return resp, nil
}

func toProtoTreeNode(n *model.OrgTreeNode) *userpb.OrgTreeNode {
	if n == nil {
		return nil
	}
	p := &userpb.OrgTreeNode{
		Id: n.ID, Name: n.Name, Type: n.Type, SortOrder: int32(n.SortOrder),
	}
	if n.ParentID != nil {
		p.ParentId = *n.ParentID
	}
	for _, ch := range n.Children {
		p.Children = append(p.Children, toProtoTreeNode(ch))
	}
	return p
}

func (s *GRPCServer) CreateOrganization(ctx context.Context, req *userpb.CreateOrganizationReq) (*userpb.CreateOrganizationResp, error) {
	var parentID *int64
	if req.ParentId != 0 {
		parentID = &req.ParentId
	}
	r, err := s.orgLogic.Create(ctx, &types.CreateOrgReq{
		Name: req.Name, ParentID: parentID, Type: req.Type, SortOrder: int(req.SortOrder),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &userpb.CreateOrganizationResp{Id: r.ID, Name: r.Name}, nil
}

func (s *GRPCServer) DeleteOrganization(ctx context.Context, req *userpb.DeleteOrganizationReq) (*userpb.DeleteOrganizationResp, error) {
	if err := s.orgLogic.Delete(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &userpb.DeleteOrganizationResp{}, nil
}
