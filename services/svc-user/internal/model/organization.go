package model

import "itms-server/pkg/db"

// Organization corresponds to sch_user.t_organization (tree structure: zone→base→brigade).
type Organization struct {
	db.BaseModel
	Name     string `gorm:"column:name;type:varchar(100);not null" json:"name"`
	ParentID *int64 `gorm:"column:parent_id;index:idx_org_parent" json:"parent_id"`
	Type     string `gorm:"column:type;type:varchar(20);not null;index:idx_org_type" json:"type"`
	SortOrder int   `gorm:"column:sort_order;default:0" json:"sort_order"`
}

func (Organization) TableName() string {
	return "sch_user.t_organization"
}

// OrgTreeNode represents a node in the organization tree.
type OrgTreeNode struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	SortOrder int            `json:"sort_order"`
	ParentID  *int64         `json:"parent_id"`
	Children  []*OrgTreeNode `json:"children,omitempty"`
}
