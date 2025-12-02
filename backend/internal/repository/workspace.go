package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/guixu633/agent/backend/internal/database"
)

// Workspace 工作区数据库模型
type Workspace struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	IsCurrent bool      `json:"is_current"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WorkspaceRepository 工作区仓库接口
type WorkspaceRepository interface {
	Create(ctx context.Context, name string) (*Workspace, error)
	GetByName(ctx context.Context, name string) (*Workspace, error)
	GetByID(ctx context.Context, id int64) (*Workspace, error)
	GetCurrent(ctx context.Context) (*Workspace, error)
	List(ctx context.Context) ([]*Workspace, error)
	SetCurrent(ctx context.Context, id int64) error
	SetCurrentByName(ctx context.Context, name string) error
	Delete(ctx context.Context, id int64) error
	DeleteByName(ctx context.Context, name string) error
}

type workspaceRepository struct {
	db *sql.DB
}

// NewWorkspaceRepository 创建工作区仓库实例
func NewWorkspaceRepository() WorkspaceRepository {
	return &workspaceRepository{
		db: database.DB,
	}
}

// Create 创建工作区
func (r *workspaceRepository) Create(ctx context.Context, name string) (*Workspace, error) {
	query := `
		INSERT INTO workspaces (name, is_current, created_at, updated_at)
		VALUES ($1, FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, name, is_current, created_at, updated_at
	`

	var ws Workspace
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&ws.ID,
		&ws.Name,
		&ws.IsCurrent,
		&ws.CreatedAt,
		&ws.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("创建工作区失败: %w", err)
	}

	return &ws, nil
}

// GetByName 根据名称获取工作区
func (r *workspaceRepository) GetByName(ctx context.Context, name string) (*Workspace, error) {
	query := `
		SELECT id, name, is_current, created_at, updated_at
		FROM workspaces
		WHERE name = $1
	`

	var ws Workspace
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&ws.ID,
		&ws.Name,
		&ws.IsCurrent,
		&ws.CreatedAt,
		&ws.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取工作区失败: %w", err)
	}

	return &ws, nil
}

// GetByID 根据 ID 获取工作区
func (r *workspaceRepository) GetByID(ctx context.Context, id int64) (*Workspace, error) {
	query := `
		SELECT id, name, is_current, created_at, updated_at
		FROM workspaces
		WHERE id = $1
	`

	var ws Workspace
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ws.ID,
		&ws.Name,
		&ws.IsCurrent,
		&ws.CreatedAt,
		&ws.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取工作区失败: %w", err)
	}

	return &ws, nil
}

// GetCurrent 获取当前工作区
func (r *workspaceRepository) GetCurrent(ctx context.Context) (*Workspace, error) {
	query := `
		SELECT id, name, is_current, created_at, updated_at
		FROM workspaces
		WHERE is_current = TRUE
		LIMIT 1
	`

	var ws Workspace
	err := r.db.QueryRowContext(ctx, query).Scan(
		&ws.ID,
		&ws.Name,
		&ws.IsCurrent,
		&ws.CreatedAt,
		&ws.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取当前工作区失败: %w", err)
	}

	return &ws, nil
}

// List 列出所有工作区
func (r *workspaceRepository) List(ctx context.Context) ([]*Workspace, error) {
	query := `
		SELECT id, name, is_current, created_at, updated_at
		FROM workspaces
		ORDER BY is_current DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("列出工作区失败: %w", err)
	}
	defer rows.Close()

	workspaces := make([]*Workspace, 0)
	for rows.Next() {
		var ws Workspace
		if err := rows.Scan(
			&ws.ID,
			&ws.Name,
			&ws.IsCurrent,
			&ws.CreatedAt,
			&ws.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描工作区数据失败: %w", err)
		}
		workspaces = append(workspaces, &ws)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历工作区数据失败: %w", err)
	}

	return workspaces, nil
}

// SetCurrent 设置当前工作区（通过 ID）
func (r *workspaceRepository) SetCurrent(ctx context.Context, id int64) error {
	// 开始事务
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 先清除所有工作区的 is_current 标志
	_, err = tx.ExecContext(ctx, `UPDATE workspaces SET is_current = FALSE`)
	if err != nil {
		return fmt.Errorf("清除当前工作区标志失败: %w", err)
	}

	// 设置指定工作区为当前工作区
	result, err := tx.ExecContext(ctx, `UPDATE workspaces SET is_current = TRUE WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("设置当前工作区失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("工作区不存在")
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// SetCurrentByName 设置当前工作区（通过名称）
func (r *workspaceRepository) SetCurrentByName(ctx context.Context, name string) error {
	// 先获取工作区 ID
	ws, err := r.GetByName(ctx, name)
	if err != nil {
		return fmt.Errorf("获取工作区失败: %w", err)
	}
	if ws == nil {
		return fmt.Errorf("工作区 %s 不存在", name)
	}

	return r.SetCurrent(ctx, ws.ID)
}

// Delete 根据 ID 删除工作区
func (r *workspaceRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM workspaces WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除工作区失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("工作区不存在")
	}

	return nil
}

// DeleteByName 根据名称删除工作区
func (r *workspaceRepository) DeleteByName(ctx context.Context, name string) error {
	query := `DELETE FROM workspaces WHERE name = $1`
	result, err := r.db.ExecContext(ctx, query, name)
	if err != nil {
		return fmt.Errorf("删除工作区失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("工作区不存在")
	}

	return nil
}
