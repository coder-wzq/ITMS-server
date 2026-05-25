package logic

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"itms-server/pkg/db"
	"itms-server/services/svc-user/internal/model"
	"itms-server/services/svc-user/internal/types"
)

// OrganizationLogic handles organization tree management.
type OrganizationLogic struct {
	gdb *gorm.DB
}

func NewOrganizationLogic(gdb *gorm.DB) *OrganizationLogic {
	return &OrganizationLogic{gdb: gdb}
}

// Tree returns the full organization tree using recursive CTE.
func (l *OrganizationLogic) Tree(ctx context.Context) ([]*model.OrgTreeNode, error) {
	var orgs []model.Organization
	if err := l.gdb.WithContext(ctx).
		Order("sort_order ASC, created_at ASC").
		Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("query organizations: %w", err)
	}

	// Build tree in-memory
	nodeMap := make(map[int64]*model.OrgTreeNode)
	var roots []*model.OrgTreeNode

	for _, org := range orgs {
		node := &model.OrgTreeNode{
			ID:        org.ID,
			Name:      org.Name,
			Type:      org.Type,
			SortOrder: org.SortOrder,
			ParentID:  org.ParentID,
			Children:  []*model.OrgTreeNode{},
		}
		nodeMap[org.ID] = node
	}

	for _, node := range nodeMap {
		if node.ParentID != nil {
			if parent, ok := nodeMap[*node.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				roots = append(roots, node)
			}
		} else {
			roots = append(roots, node)
		}
	}

	return roots, nil
}

// Create inserts a new organization node.
func (l *OrganizationLogic) Create(ctx context.Context, req *types.CreateOrgReq) (*types.CreateOrgResp, error) {
	// Validate type
	if req.Type != "zone" && req.Type != "base" && req.Type != "brigade" {
		return nil, fmt.Errorf("invalid organization type: %s (must be zone/base/brigade)", req.Type)
	}

	// Validate parent exists if set
	if req.ParentID != nil {
		var parent model.Organization
		if err := l.gdb.WithContext(ctx).Where("id = ?", *req.ParentID).First(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("parent organization not found")
			}
			return nil, fmt.Errorf("find parent: %w", err)
		}
	}

	org := &model.Organization{
		BaseModel: db.BaseModel{ID: db.GetSnowflake().NextID()},
		Name:      req.Name,
		ParentID:  req.ParentID,
		Type:      req.Type,
		SortOrder: req.SortOrder,
	}

	if err := l.gdb.WithContext(ctx).Create(org).Error; err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}

	return &types.CreateOrgResp{ID: org.ID, Name: org.Name}, nil
}

// Delete soft-deletes an organization node and all its descendants.
func (l *OrganizationLogic) Delete(ctx context.Context, id int64) error {
	// Collect all descendant IDs via recursive CTE
	var ids []int64
	err := l.gdb.WithContext(ctx).Raw(`
		WITH RECURSIVE org_tree AS (
			SELECT id FROM sch_user.t_organization WHERE id = ? AND deleted_at IS NULL
			UNION ALL
			SELECT o.id FROM sch_user.t_organization o
			INNER JOIN org_tree ot ON o.parent_id = ot.id
			WHERE o.deleted_at IS NULL
		)
		SELECT id FROM org_tree
	`, id).Scan(&ids).Error
	if err != nil {
		return fmt.Errorf("collect descendants: %w", err)
	}
	if len(ids) == 0 {
		return fmt.Errorf("organization not found")
	}

	if err := l.gdb.WithContext(ctx).Where("id IN ?", ids).Delete(&model.Organization{}).Error; err != nil {
		return fmt.Errorf("delete organizations: %w", err)
	}
	return nil
}
