package workspace

import (
	"context"
	"fmt"

	"github.com/guixu633/agent/backend/internal/model"
	"github.com/guixu633/agent/backend/internal/oss"
)

// Service 工作区服务
type Service struct {
	ossClient *oss.Client
}

// NewService 创建工作区服务实例
func NewService(ossClient *oss.Client) *Service {
	return &Service{
		ossClient: ossClient,
	}
}

// ListWorkspaces 列出所有工作区
func (s *Service) ListWorkspaces(ctx context.Context) (*model.ListWorkspacesResponse, error) {
	workspaceNames, err := s.ossClient.ListWorkspaces()
	if err != nil {
		return nil, fmt.Errorf("列出工作区失败: %w", err)
	}

	workspaces := make([]model.Workspace, 0, len(workspaceNames))
	for _, name := range workspaceNames {
		workspaces = append(workspaces, model.Workspace{
			Name: name,
		})
	}

	return &model.ListWorkspacesResponse{
		Workspaces: workspaces,
	}, nil
}

// CreateWorkspace 创建工作区
func (s *Service) CreateWorkspace(ctx context.Context, req *model.CreateWorkspaceRequest) (*model.CreateWorkspaceResponse, error) {
	// 验证工作区名称
	if req.Name == "" {
		return nil, fmt.Errorf("工作区名称不能为空")
	}

	// 创建工作区
	err := s.ossClient.CreateWorkspace(req.Name)
	if err != nil {
		return nil, fmt.Errorf("创建工作区失败: %w", err)
	}

	return &model.CreateWorkspaceResponse{
		Workspace: model.Workspace{
			Name: req.Name,
		},
	}, nil
}

// DeleteWorkspace 删除工作区
func (s *Service) DeleteWorkspace(ctx context.Context, req *model.DeleteWorkspaceRequest) error {
	// 验证工作区名称
	if req.Name == "" {
		return fmt.Errorf("工作区名称不能为空")
	}

	// 删除工作区
	err := s.ossClient.DeleteWorkspace(req.Name)
	if err != nil {
		return fmt.Errorf("删除工作区失败: %w", err)
	}

	return nil
}

