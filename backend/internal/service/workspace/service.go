package workspace

import (
	"context"
	"fmt"
	"time"

	"github.com/guixu633/agent/backend/internal/model"
	"github.com/guixu633/agent/backend/internal/oss"
	"github.com/guixu633/agent/backend/internal/repository"
)

// Service 工作区服务
type Service struct {
	ossClient        *oss.Client
	workspaceRepo    repository.WorkspaceRepository
}

// NewService 创建工作区服务实例
func NewService(ossClient *oss.Client, workspaceRepo repository.WorkspaceRepository) *Service {
	return &Service{
		ossClient:     ossClient,
		workspaceRepo: workspaceRepo,
	}
}

// ListWorkspaces 列出所有工作区
func (s *Service) ListWorkspaces(ctx context.Context) (*model.ListWorkspacesResponse, error) {
	// 从数据库获取工作区列表
	dbWorkspaces, err := s.workspaceRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("列出工作区失败: %w", err)
	}

	workspaces := make([]model.Workspace, 0, len(dbWorkspaces))
	for _, ws := range dbWorkspaces {
		workspaces = append(workspaces, model.Workspace{
			Name:      ws.Name,
			IsCurrent: ws.IsCurrent,
			CreatedAt: ws.CreatedAt.Format(time.RFC3339),
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

	// 检查工作区是否已存在
	existing, err := s.workspaceRepo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("检查工作区是否存在失败: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("工作区 %s 已存在", req.Name)
	}

	// 在数据库中创建工作区
	dbWorkspace, err := s.workspaceRepo.Create(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("创建工作区失败: %w", err)
	}

	// 在 OSS 中创建工作区目录（用于存储文件）
	err = s.ossClient.CreateWorkspace(req.Name)
	if err != nil {
		// 如果 OSS 创建失败，删除数据库记录（回滚）
		s.workspaceRepo.Delete(ctx, dbWorkspace.ID)
		return nil, fmt.Errorf("在 OSS 创建工作区目录失败: %w", err)
	}

	return &model.CreateWorkspaceResponse{
		Workspace: model.Workspace{
			Name:      dbWorkspace.Name,
			IsCurrent: dbWorkspace.IsCurrent,
			CreatedAt: dbWorkspace.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// SetCurrentWorkspace 设置当前工作区
func (s *Service) SetCurrentWorkspace(ctx context.Context, req *model.SetCurrentWorkspaceRequest) (*model.SetCurrentWorkspaceResponse, error) {
	// 验证工作区名称
	if req.Name == "" {
		return nil, fmt.Errorf("工作区名称不能为空")
	}

	// 设置当前工作区
	err := s.workspaceRepo.SetCurrentByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("设置当前工作区失败: %w", err)
	}

	// 获取更新后的工作区信息
	ws, err := s.workspaceRepo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("获取工作区信息失败: %w", err)
	}
	if ws == nil {
		return nil, fmt.Errorf("工作区不存在")
	}

	return &model.SetCurrentWorkspaceResponse{
		Workspace: model.Workspace{
			Name:      ws.Name,
			IsCurrent: ws.IsCurrent,
			CreatedAt: ws.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetCurrentWorkspace 获取当前工作区
func (s *Service) GetCurrentWorkspace(ctx context.Context) (*model.GetCurrentWorkspaceResponse, error) {
	ws, err := s.workspaceRepo.GetCurrent(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取当前工作区失败: %w", err)
	}

	if ws == nil {
		return &model.GetCurrentWorkspaceResponse{
			Workspace: nil,
		}, nil
	}

	return &model.GetCurrentWorkspaceResponse{
		Workspace: &model.Workspace{
			Name:      ws.Name,
			IsCurrent: ws.IsCurrent,
			CreatedAt: ws.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// DeleteWorkspace 删除工作区
func (s *Service) DeleteWorkspace(ctx context.Context, req *model.DeleteWorkspaceRequest) error {
	// 验证工作区名称
	if req.Name == "" {
		return fmt.Errorf("工作区名称不能为空")
	}

	// 获取工作区信息
	ws, err := s.workspaceRepo.GetByName(ctx, req.Name)
	if err != nil {
		return fmt.Errorf("获取工作区失败: %w", err)
	}
	if ws == nil {
		return fmt.Errorf("工作区 %s 不存在", req.Name)
	}

	// 先删除 OSS 中的文件（级联删除会删除数据库中的图片记录）
	err = s.ossClient.DeleteWorkspace(req.Name)
	if err != nil {
		return fmt.Errorf("删除 OSS 工作区文件失败: %w", err)
	}

	// 删除数据库记录（级联删除会删除关联的图片记录）
	err = s.workspaceRepo.Delete(ctx, ws.ID)
	if err != nil {
		return fmt.Errorf("删除工作区记录失败: %w", err)
	}

	return nil
}

