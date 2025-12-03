package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/guixu633/agent/backend/internal/database"
)

// Message 对话消息（与 model.Message 保持一致）
type Message struct {
	Role    string `json:"role"`              // 角色: "user" | "assistant"
	Type    string `json:"type"`              // 消息类型: "text" | "image"
	Content string `json:"content,omitempty"` // 文本内容 (type="text" 时有效)
	URL     string `json:"url,omitempty"`     // 图片 URL (type="image" 时有效)
}

// Image 图片数据库模型
type Image struct {
	ID            int64     `json:"id"`
	WorkspaceID   int64     `json:"workspace_id"`
	Name          string    `json:"name"`
	OSSPath       string    `json:"oss_path"`
	OSSUrl        string    `json:"oss_url"`
	ThumbnailPath string    `json:"thumbnail_path"`
	ThumbnailUrl  string    `json:"thumbnail_url"`
	Size          int64     `json:"size"`
	MimeType      string    `json:"mime_type"`
	SourceType    string    `json:"source_type"`
	Prompt        string    `json:"prompt"`
	RefImages     []string  `json:"ref_images"`
	MessageList   []Message `json:"message_list"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ImageRepository 图片仓库接口
type ImageRepository interface {
	Create(ctx context.Context, img *Image) (*Image, error)
	GetByID(ctx context.Context, id int64) (*Image, error)
	GetByOSSPath(ctx context.Context, ossPath string) (*Image, error)
	ListByWorkspace(ctx context.Context, workspaceID int64) ([]*Image, error)
	ListByWorkspaceName(ctx context.Context, workspaceName string) ([]*Image, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) (*Image, error)
	Delete(ctx context.Context, id int64) error
	DeleteByOSSPath(ctx context.Context, ossPath string) error
}

type imageRepository struct {
	db *sql.DB
}

// NewImageRepository 创建图片仓库实例
func NewImageRepository() ImageRepository {
	return &imageRepository{
		db: database.DB,
	}
}

// Create 创建图片记录
func (r *imageRepository) Create(ctx context.Context, img *Image) (*Image, error) {
	query := `
		INSERT INTO images (
			workspace_id, name, oss_path, oss_url, 
			thumbnail_path, thumbnail_url, size, mime_type,
			source_type, prompt, ref_images, message_list,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, workspace_id, name, oss_path, oss_url, 
		          thumbnail_path, thumbnail_url, size, mime_type,
		          source_type, prompt, ref_images, message_list,
		          created_at, updated_at
	`

	refImagesJSON, err := json.Marshal(img.RefImages)
	if err != nil {
		return nil, fmt.Errorf("序列化引用图片失败: %w", err)
	}
	if img.RefImages == nil {
		refImagesJSON = []byte("[]")
	}

	messageListJSON, err := json.Marshal(img.MessageList)
	if err != nil {
		return nil, fmt.Errorf("序列化消息列表失败: %w", err)
	}
	if img.MessageList == nil {
		messageListJSON = []byte("[]")
	}

	// 默认值处理
	if img.SourceType == "" {
		img.SourceType = "upload"
	}

	var result Image
	var refImagesBytes, messageListBytes []byte

	err = r.db.QueryRowContext(ctx, query,
		img.WorkspaceID,
		img.Name,
		img.OSSPath,
		img.OSSUrl,
		img.ThumbnailPath,
		img.ThumbnailUrl,
		img.Size,
		img.MimeType,
		img.SourceType,
		img.Prompt,
		refImagesJSON,
		messageListJSON,
	).Scan(
		&result.ID,
		&result.WorkspaceID,
		&result.Name,
		&result.OSSPath,
		&result.OSSUrl,
		&result.ThumbnailPath,
		&result.ThumbnailUrl,
		&result.Size,
		&result.MimeType,
		&result.SourceType,
		&result.Prompt,
		&refImagesBytes,
		&messageListBytes,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("创建图片记录失败: %w", err)
	}

	if len(refImagesBytes) > 0 {
		if err := json.Unmarshal(refImagesBytes, &result.RefImages); err != nil {
			return nil, fmt.Errorf("反序列化引用图片失败: %w", err)
		}
	}
	if len(messageListBytes) > 0 {
		if err := json.Unmarshal(messageListBytes, &result.MessageList); err != nil {
			return nil, fmt.Errorf("反序列化消息列表失败: %w", err)
		}
	}

	return &result, nil
}

// GetByID 根据 ID 获取图片
func (r *imageRepository) GetByID(ctx context.Context, id int64) (*Image, error) {
	query := `
		SELECT id, workspace_id, name, oss_path, oss_url,
		       thumbnail_path, thumbnail_url, size, mime_type,
		       source_type, prompt, ref_images, message_list,
		       created_at, updated_at
		FROM images
		WHERE id = $1
	`

	var img Image
	var refImagesBytes, messageListBytes []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&img.ID,
		&img.WorkspaceID,
		&img.Name,
		&img.OSSPath,
		&img.OSSUrl,
		&img.ThumbnailPath,
		&img.ThumbnailUrl,
		&img.Size,
		&img.MimeType,
		&img.SourceType,
		&img.Prompt,
		&refImagesBytes,
		&messageListBytes,
		&img.CreatedAt,
		&img.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取图片失败: %w", err)
	}

	if len(refImagesBytes) > 0 {
		json.Unmarshal(refImagesBytes, &img.RefImages)
	}
	if len(messageListBytes) > 0 {
		json.Unmarshal(messageListBytes, &img.MessageList)
	}

	return &img, nil
}

// GetByOSSPath 根据 OSS 路径获取图片
func (r *imageRepository) GetByOSSPath(ctx context.Context, ossPath string) (*Image, error) {
	query := `
		SELECT id, workspace_id, name, oss_path, oss_url,
		       thumbnail_path, thumbnail_url, size, mime_type,
		       source_type, prompt, ref_images, message_list,
		       created_at, updated_at
		FROM images
		WHERE oss_path = $1
	`

	var img Image
	var refImagesBytes, messageListBytes []byte

	err := r.db.QueryRowContext(ctx, query, ossPath).Scan(
		&img.ID,
		&img.WorkspaceID,
		&img.Name,
		&img.OSSPath,
		&img.OSSUrl,
		&img.ThumbnailPath,
		&img.ThumbnailUrl,
		&img.Size,
		&img.MimeType,
		&img.SourceType,
		&img.Prompt,
		&refImagesBytes,
		&messageListBytes,
		&img.CreatedAt,
		&img.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("获取图片失败: %w", err)
	}

	if len(refImagesBytes) > 0 {
		json.Unmarshal(refImagesBytes, &img.RefImages)
	}
	if len(messageListBytes) > 0 {
		json.Unmarshal(messageListBytes, &img.MessageList)
	}

	return &img, nil
}

// ListByWorkspace 根据工作区 ID 列出图片
// 注意：不查询 prompt, ref_images, message_list 字段以减少数据传输量，需要完整信息请使用 GetByID
func (r *imageRepository) ListByWorkspace(ctx context.Context, workspaceID int64) ([]*Image, error) {
	query := `
		SELECT id, workspace_id, name, oss_path, oss_url,
		       thumbnail_path, thumbnail_url, size, mime_type,
		       source_type, created_at, updated_at
		FROM images
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("列出图片失败: %w", err)
	}
	defer rows.Close()

	images := make([]*Image, 0)
	for rows.Next() {
		var img Image
		if err := rows.Scan(
			&img.ID,
			&img.WorkspaceID,
			&img.Name,
			&img.OSSPath,
			&img.OSSUrl,
			&img.ThumbnailPath,
			&img.ThumbnailUrl,
			&img.Size,
			&img.MimeType,
			&img.SourceType,
			&img.CreatedAt,
			&img.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描图片数据失败: %w", err)
		}

		images = append(images, &img)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历图片数据失败: %w", err)
	}

	return images, nil
}

// ListByWorkspaceName 根据工作区名称列出图片
// 注意：不查询 prompt, ref_images, message_list 字段以减少数据传输量，需要完整信息请使用 GetByID
func (r *imageRepository) ListByWorkspaceName(ctx context.Context, workspaceName string) ([]*Image, error) {
	query := `
		SELECT i.id, i.workspace_id, i.name, i.oss_path, i.oss_url,
		       i.thumbnail_path, i.thumbnail_url, i.size, i.mime_type,
		       i.source_type, i.created_at, i.updated_at
		FROM images i
		INNER JOIN workspaces w ON i.workspace_id = w.id
		WHERE w.name = $1
		ORDER BY i.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, workspaceName)
	if err != nil {
		return nil, fmt.Errorf("列出图片失败: %w", err)
	}
	defer rows.Close()

	images := make([]*Image, 0)
	for rows.Next() {
		var img Image
		if err := rows.Scan(
			&img.ID,
			&img.WorkspaceID,
			&img.Name,
			&img.OSSPath,
			&img.OSSUrl,
			&img.ThumbnailPath,
			&img.ThumbnailUrl,
			&img.Size,
			&img.MimeType,
			&img.SourceType,
			&img.CreatedAt,
			&img.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描图片数据失败: %w", err)
		}

		images = append(images, &img)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历图片数据失败: %w", err)
	}

	return images, nil
}

// Update 更新图片记录
// 支持更新 name, oss_path, oss_url, thumbnail_path, thumbnail_url 字段
func (r *imageRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) (*Image, error) {
	if len(updates) == 0 {
		return r.GetByID(ctx, id)
	}

	// 构建动态 SQL
	query := "UPDATE images SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}
	argId := 1

	// 处理支持的字段
	if name, ok := updates["name"].(string); ok {
		query += fmt.Sprintf(", name = $%d", argId)
		args = append(args, name)
		argId++
	}

	if ossPath, ok := updates["oss_path"].(string); ok {
		query += fmt.Sprintf(", oss_path = $%d", argId)
		args = append(args, ossPath)
		argId++
	}

	if ossUrl, ok := updates["oss_url"].(string); ok {
		query += fmt.Sprintf(", oss_url = $%d", argId)
		args = append(args, ossUrl)
		argId++
	}

	if thumbnailPath, ok := updates["thumbnail_path"].(string); ok {
		query += fmt.Sprintf(", thumbnail_path = $%d", argId)
		args = append(args, thumbnailPath)
		argId++
	}

	if thumbnailUrl, ok := updates["thumbnail_url"].(string); ok {
		query += fmt.Sprintf(", thumbnail_url = $%d", argId)
		args = append(args, thumbnailUrl)
		argId++
	}

	// source_type, prompt, ref_images, message_list 也可以支持更新，但目前需求主要是重命名

	// 添加 WHERE 子句
	query += fmt.Sprintf(" WHERE id = $%d", argId)
	args = append(args, id)

	// 添加 RETURNING 子句
	query += `
		RETURNING id, workspace_id, name, oss_path, oss_url,
		          thumbnail_path, thumbnail_url, size, mime_type,
		          source_type, prompt, ref_images, message_list,
		          created_at, updated_at
	`

	var img Image
	var refImagesBytes, messageListBytes []byte

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&img.ID,
		&img.WorkspaceID,
		&img.Name,
		&img.OSSPath,
		&img.OSSUrl,
		&img.ThumbnailPath,
		&img.ThumbnailUrl,
		&img.Size,
		&img.MimeType,
		&img.SourceType,
		&img.Prompt,
		&refImagesBytes,
		&messageListBytes,
		&img.CreatedAt,
		&img.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("更新图片失败: %w", err)
	}

	if len(refImagesBytes) > 0 {
		json.Unmarshal(refImagesBytes, &img.RefImages)
	}
	if len(messageListBytes) > 0 {
		json.Unmarshal(messageListBytes, &img.MessageList)
	}

	return &img, nil
}

// Delete 根据 ID 删除图片
func (r *imageRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM images WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除图片失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("图片不存在")
	}

	return nil
}

// DeleteByOSSPath 根据 OSS 路径删除图片
func (r *imageRepository) DeleteByOSSPath(ctx context.Context, ossPath string) error {
	query := `DELETE FROM images WHERE oss_path = $1`
	result, err := r.db.ExecContext(ctx, query, ossPath)
	if err != nil {
		return fmt.Errorf("删除图片失败: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除行数失败: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("图片不存在")
	}

	return nil
}
