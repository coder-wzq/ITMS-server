package types

// --- User management types ---

// UserListReq is the query for GET /api/users.
type UserListReq struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Keyword  string `json:"keyword" form:"keyword"`
}

// UserListItem is a single user row in the list response.
type UserListItem struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	RealName  string `json:"realName"`
	Department string `json:"department"`
	Role      string `json:"role"`
	Status    int    `json:"status"`
	LastLogin string `json:"lastLogin"`
}

// CreateUserReq is the body for POST /api/users.
type CreateUserReq struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	RealName   string `json:"realName"`
	Department string `json:"department"`
	Role       string `json:"role"`
}

// CreateUserResp is the response for creating a user.
type CreateUserResp struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// UpdateUserReq is the body for PUT /api/users/{id}.
type UpdateUserReq struct {
	Password   string `json:"password,omitempty"`
	RealName   string `json:"realName"`
	Department string `json:"department"`
	Role       string `json:"role"`
	Status     *int   `json:"status"`
}

// BatchDeleteReq is the body for POST /api/users/batch-delete.
type BatchDeleteReq struct {
	IDs []int64 `json:"ids"`
}

// BatchDeleteResp is the response for batch delete.
type BatchDeleteResp struct {
	DeletedCount int `json:"deletedCount"`
}

// --- Profile types ---

// UserProfileResp is the response for GET /api/users/profile.
type UserProfileResp struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	RealName   string `json:"realName"`
	Department string `json:"department"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
	Role       string `json:"role"`
}

// UpdateProfileReq is the body for PUT /api/users/profile.
type UpdateProfileReq struct {
	Department string `json:"department"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
}

// --- Organization types ---

// CreateOrgReq is the body for POST /api/organizations.
type CreateOrgReq struct {
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id"`
	Type     string `json:"type"`
	SortOrder int   `json:"sort_order"`
}

// CreateOrgResp is the response for creating an organization.
type CreateOrgResp struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

