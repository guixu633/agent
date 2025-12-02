package model

// Workspace 工作区
type Workspace struct {
	Name      string `json:"name"`        // 工作区名称
	IsCurrent bool   `json:"is_current"`  // 是否为当前工作区
	CreatedAt string `json:"created_at"`  // 创建时间（可选）
}

// ListWorkspacesResponse 列出工作区响应
type ListWorkspacesResponse struct {
	Workspaces []Workspace `json:"workspaces"`
}

// CreateWorkspaceRequest 创建工作区请求
type CreateWorkspaceRequest struct {
	Name string `json:"name" binding:"required"` // 工作区名称
}

// CreateWorkspaceResponse 创建工作区响应
type CreateWorkspaceResponse struct {
	Workspace Workspace `json:"workspace"`
}

// DeleteWorkspaceRequest 删除工作区请求
type DeleteWorkspaceRequest struct {
	Name string `json:"name" binding:"required"` // 工作区名称
}

// SetCurrentWorkspaceRequest 设置当前工作区请求
type SetCurrentWorkspaceRequest struct {
	Name string `json:"name" binding:"required"` // 工作区名称
}

// SetCurrentWorkspaceResponse 设置当前工作区响应
type SetCurrentWorkspaceResponse struct {
	Workspace Workspace `json:"workspace"`
}

// GetCurrentWorkspaceResponse 获取当前工作区响应
type GetCurrentWorkspaceResponse struct {
	Workspace *Workspace `json:"workspace"` // 如果为 nil 表示没有当前工作区
}
