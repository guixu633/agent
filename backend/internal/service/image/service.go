package image

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/guixu633/agent/backend/internal/model"
	"github.com/guixu633/agent/backend/internal/oss"
	"github.com/guixu633/agent/backend/internal/repository"
	"github.com/guixu633/agent/backend/pkg/thumbnail"
	"google.golang.org/genai"
)

const (
	DefaultModel = "gemini-3-pro-image-preview"
)

// Service 图片服务
type Service struct {
	genaiClient  *genai.Client
	ossClient    *oss.Client
	imageRepo    repository.ImageRepository
	workspaceRepo repository.WorkspaceRepository
}

// NewService 创建图片服务实例
func NewService(genaiClient *genai.Client, ossClient *oss.Client, imageRepo repository.ImageRepository, workspaceRepo repository.WorkspaceRepository) *Service {
	return &Service{
		genaiClient:  genaiClient,
		ossClient:    ossClient,
		imageRepo:    imageRepo,
		workspaceRepo: workspaceRepo,
	}
}

// UploadImage 上传图片到 OSS 并保存到数据库
func (s *Service) UploadImage(ctx context.Context, file io.Reader, filename string, workspace string) (*model.ImageUploadResponse, error) {
	// 获取或创建工作区
	ws, err := s.workspaceRepo.GetByName(ctx, workspace)
	if err != nil {
		return nil, fmt.Errorf("获取工作区失败: %w", err)
	}
	if ws == nil {
		return nil, fmt.Errorf("工作区 %s 不存在", workspace)
	}

	// 读取文件数据（需要读取两次：一次上传原图，一次生成缩略图）
	imageData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取图片数据失败: %w", err)
	}

	// 上传原图到 OSS
	path, err := s.ossClient.UploadImage(bytes.NewReader(imageData), filename, workspace)
	if err != nil {
		return nil, fmt.Errorf("上传图片失败: %w", err)
	}

	// 生成并上传缩略图
	thumbnailPath, thumbnailURL, err := s.uploadThumbnail(ctx, imageData, filename, workspace)
	if err != nil {
		// 缩略图生成失败不影响主流程，只记录错误
		thumbnailPath = ""
		thumbnailURL = ""
	}

	// 获取访问 URL
	url := s.ossClient.GetImageURL(path)

	// 检测 MIME 类型
	mimeType := s.detectMimeTypeFromFilename(filename)

	// 保存到数据库
	dbImage, err := s.imageRepo.Create(ctx, &repository.Image{
		WorkspaceID:   ws.ID,
		Name:          filename,
		OSSPath:       path,
		OSSUrl:        url,
		ThumbnailPath: thumbnailPath,
		ThumbnailUrl:  thumbnailURL,
		Size:          int64(len(imageData)),
		MimeType:      mimeType,
	})
	if err != nil {
		// 如果数据库保存失败，删除 OSS 文件（回滚）
		s.ossClient.DeleteImage(path)
		if thumbnailPath != "" {
			s.ossClient.DeleteImage(thumbnailPath)
		}
		return nil, fmt.Errorf("保存图片记录到数据库失败: %w", err)
	}

	return &model.ImageUploadResponse{
		Path: dbImage.OSSPath,
		URL:  dbImage.OSSUrl,
	}, nil
}

// uploadThumbnail 生成并上传缩略图
// 返回缩略图路径和 URL
func (s *Service) uploadThumbnail(ctx context.Context, imageData []byte, filename string, workspace string) (string, string, error) {
	// 检测 MIME 类型
	mimeType := s.detectMimeTypeFromFilename(filename)

	// 生成缩略图
	thumbnailData, err := thumbnail.GenerateThumbnail(imageData, mimeType)
	if err != nil {
		return "", "", fmt.Errorf("生成缩略图失败: %w", err)
	}

	// 获取缩略图文件名
	thumbnailFilename := thumbnail.GetThumbnailFilename(filename)

	// 上传缩略图
	thumbnailPath, err := s.ossClient.UploadImage(bytes.NewReader(thumbnailData), thumbnailFilename, workspace)
	if err != nil {
		return "", "", fmt.Errorf("上传缩略图失败: %w", err)
	}

	thumbnailURL := s.ossClient.GetImageURL(thumbnailPath)
	return thumbnailPath, thumbnailURL, nil
}

// detectMimeTypeFromFilename 根据文件名检测 MIME 类型
func (s *Service) detectMimeTypeFromFilename(filename string) string {
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		// 默认使用 image/jpeg
		mimeType = "image/jpeg"
	}
	return mimeType
}

// GenerateImage 生成图片
func (s *Service) GenerateImage(ctx context.Context, req *model.ImageGenerateRequest) (*model.ImageGenerateResponse, error) {
	// 构建请求内容
	contents := []*genai.Content{
		genai.NewContentFromText(req.Prompt, genai.RoleUser),
	}

	// 添加输入图片（从 OSS 获取）
	for _, imagePath := range req.Images {
		// 从 OSS 下载图片数据（Gemini API 需要二进制数据）
		// 注意：如果未来 Gemini API 支持 URL，可以将 useURL 参数改为 true
		_, imageData, err := s.ossClient.GetImageURLOrDownload(imagePath, false)
		if err != nil {
			return nil, fmt.Errorf("从 OSS 获取图片失败 (path: %s): %w", imagePath, err)
		}

		// 使用二进制数据
		mimeType := s.detectMimeType(imagePath)
		contents = append(contents, genai.NewContentFromBytes(imageData, mimeType, genai.RoleUser))
	}

	// 调用 Gemini API
	resp, err := s.genaiClient.Models.GenerateContent(ctx, DefaultModel, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("调用 Gemini API 失败: %w", err)
	}

	// 解析响应
	result := &model.ImageGenerateResponse{
		Parts: make([]model.GeneratePart, 0),
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("模型未返回任何内容")
	}

	// 按顺序提取内容
	for _, part := range resp.Candidates[0].Content.Parts {
		// 1. 处理文本
		if text := strings.TrimSpace(part.Text); text != "" {
			result.Parts = append(result.Parts, model.GeneratePart{
				Type: "text",
				Text: text,
			})
		}

		// 2. 处理图片
		if part.InlineData != nil {
			// 将生成的图片上传到 OSS
			imageData := part.InlineData.Data
			mimeType := part.InlineData.MIMEType

			// 根据 MIME 类型确定文件扩展名
			ext := s.getExtensionFromMimeType(mimeType)
			filename := fmt.Sprintf("generated-%d%s", time.Now().UnixNano(), ext)

			// 使用请求中的 workspace，如果没有则使用 "default"
			workspace := req.Workspace
			if workspace == "" {
				workspace = "default"
			}

			// 获取或创建工作区
			ws, err := s.workspaceRepo.GetByName(ctx, workspace)
			if err != nil {
				return nil, fmt.Errorf("获取工作区失败: %w", err)
			}
			if ws == nil {
				return nil, fmt.Errorf("工作区 %s 不存在", workspace)
			}

			// 上传原图到 OSS
			path, err := s.ossClient.UploadImage(bytes.NewReader(imageData), filename, workspace)
			if err != nil {
				return nil, fmt.Errorf("上传生成的图片到 OSS 失败: %w", err)
			}

			// 生成并上传缩略图
			thumbnailPath, thumbnailURL, err := s.uploadThumbnail(ctx, imageData, filename, workspace)
			if err != nil {
				// 缩略图生成失败不影响主流程
				thumbnailPath = ""
				thumbnailURL = ""
			}

			// 获取访问 URL
			url := s.ossClient.GetImageURL(path)

			// 保存到数据库
			dbImage, err := s.imageRepo.Create(ctx, &repository.Image{
				WorkspaceID:   ws.ID,
				Name:          filename,
				OSSPath:       path,
				OSSUrl:        url,
				ThumbnailPath: thumbnailPath,
				ThumbnailUrl:  thumbnailURL,
				Size:          int64(len(imageData)),
				MimeType:      mimeType,
			})
			if err != nil {
				// 如果数据库保存失败，删除 OSS 文件（回滚）
				s.ossClient.DeleteImage(path)
				if thumbnailPath != "" {
					s.ossClient.DeleteImage(thumbnailPath)
				}
				return nil, fmt.Errorf("保存生成的图片记录到数据库失败: %w", err)
			}

			result.Parts = append(result.Parts, model.GeneratePart{
				Type: "image",
				Image: &model.GeneratedImage{
					MimeType: mimeType,
					Path:     dbImage.OSSPath,
					URL:      dbImage.OSSUrl,
				},
			})
		}
	}

	return result, nil
}

// parseBase64Image 解析 base64 图片数据
func (s *Service) parseBase64Image(base64Str string) ([]byte, string, error) {
	// 默认 MIME 类型
	mimeType := "image/png"

	// 检查是否包含 data URL 前缀
	if strings.HasPrefix(base64Str, "data:") {
		// 格式: data:image/png;base64,iVBORw0KG...
		parts := strings.SplitN(base64Str, ",", 2)
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("无效的 data URL 格式")
		}

		// 提取 MIME 类型
		header := parts[0]
		if strings.Contains(header, ";") {
			mimeTypePart := strings.Split(header, ";")[0]
			mimeType = strings.TrimPrefix(mimeTypePart, "data:")
		}

		base64Str = parts[1]
	}

	// 解码 base64
	imageData, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, "", fmt.Errorf("base64 解码失败: %w", err)
	}

	return imageData, mimeType, nil
}

// detectMimeType 根据文件路径检测 MIME 类型
func (s *Service) detectMimeType(path string) string {
	ext := filepath.Ext(path)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		// 默认使用 image/jpeg
		mimeType = "image/jpeg"
	}
	return mimeType
}

// getExtensionFromMimeType 根据 MIME 类型获取文件扩展名
func (s *Service) getExtensionFromMimeType(mimeType string) string {
	exts, err := mime.ExtensionsByType(mimeType)
	if err != nil || len(exts) == 0 {
		// 默认使用 .png
		return ".png"
	}
	// 返回第一个扩展名（去掉点号）
	return exts[0]
}

// ListWorkspaceImages 列出工作区的所有图片（从数据库读取）
func (s *Service) ListWorkspaceImages(ctx context.Context, workspace string) (*model.ListWorkspaceImagesResponse, error) {
	// 从数据库获取图片列表
	dbImages, err := s.imageRepo.ListByWorkspaceName(ctx, workspace)
	if err != nil {
		return nil, fmt.Errorf("列出工作区图片失败: %w", err)
	}

	// 转换为 model 格式
	images := make([]model.ImageInfo, 0, len(dbImages))
	for _, dbImage := range dbImages {
		images = append(images, model.ImageInfo{
			Path:         dbImage.OSSPath,
			URL:          dbImage.OSSUrl,
			ThumbnailURL: dbImage.ThumbnailUrl,
			Name:         dbImage.Name,
			Size:         dbImage.Size,
			Updated:      dbImage.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &model.ListWorkspaceImagesResponse{
		Images: images,
	}, nil
}

// DeleteImage 删除图片（同时删除 OSS 文件和数据库记录）
func (s *Service) DeleteImage(ctx context.Context, path string) error {
	// 从数据库获取图片信息
	dbImage, err := s.imageRepo.GetByOSSPath(ctx, path)
	if err != nil {
		return fmt.Errorf("获取图片记录失败: %w", err)
	}
	if dbImage == nil {
		return fmt.Errorf("图片记录不存在")
	}

	// 删除 OSS 中的原图
	err = s.ossClient.DeleteImage(path)
	if err != nil {
		return fmt.Errorf("删除 OSS 图片失败: %w", err)
	}

	// 删除缩略图（如果存在）
	if dbImage.ThumbnailPath != "" {
		s.ossClient.DeleteImage(dbImage.ThumbnailPath)
	}

	// 删除数据库记录
	err = s.imageRepo.Delete(ctx, dbImage.ID)
	if err != nil {
		return fmt.Errorf("删除图片记录失败: %w", err)
	}

	return nil
}

// RenameImage 重命名图片（同时更新 OSS 和数据库）
func (s *Service) RenameImage(ctx context.Context, req *model.RenameImageRequest) (*model.RenameImageResponse, error) {
	// 验证新文件名
	if req.NewName == "" {
		return nil, fmt.Errorf("新文件名不能为空")
	}

	// 从数据库获取图片信息
	dbImage, err := s.imageRepo.GetByOSSPath(ctx, req.Path)
	if err != nil {
		return nil, fmt.Errorf("获取图片记录失败: %w", err)
	}
	if dbImage == nil {
		return nil, fmt.Errorf("图片记录不存在")
	}

	// 重命名 OSS 中的原图
	newPath, err := s.ossClient.RenameImage(req.Path, req.NewName, req.Workspace)
	if err != nil {
		return nil, fmt.Errorf("重命名 OSS 图片失败: %w", err)
	}

	// 重命名缩略图（如果存在）
	var newThumbnailPath string
	if dbImage.ThumbnailPath != "" {
		oldFilename := dbImage.Name
		newThumbnailFilename := thumbnail.GetThumbnailFilename(req.NewName)
		parts := strings.Split(dbImage.ThumbnailPath, "/")
		if len(parts) > 0 {
			parts[len(parts)-1] = newThumbnailFilename
		}
		oldThumbnailPath := dbImage.ThumbnailPath
		newThumbnailPath = strings.Join(parts, "/")
		
		// 重命名缩略图
		_, err = s.ossClient.RenameImage(oldThumbnailPath, newThumbnailFilename, req.Workspace)
		if err != nil {
			// 如果缩略图重命名失败，回滚原图重命名
			s.ossClient.RenameImage(newPath, oldFilename, req.Workspace)
			return nil, fmt.Errorf("重命名缩略图失败: %w", err)
		}
	}

	// 更新数据库记录
	updatedImage, err := s.imageRepo.Update(ctx, dbImage.ID, map[string]interface{}{
		"name": req.NewName,
	})
	if err != nil {
		// 如果数据库更新失败，回滚 OSS 重命名
		s.ossClient.RenameImage(newPath, dbImage.Name, req.Workspace)
		if newThumbnailPath != "" {
			oldThumbnailFilename := thumbnail.GetThumbnailFilename(dbImage.Name)
			s.ossClient.RenameImage(newThumbnailPath, oldThumbnailFilename, req.Workspace)
		}
		return nil, fmt.Errorf("更新图片记录失败: %w", err)
	}

	// 如果缩略图路径也改变了，需要更新数据库中的缩略图路径
	// 注意：这里简化处理，如果需要可以扩展 Update 方法支持更多字段

	return &model.RenameImageResponse{
		Image: model.ImageInfo{
			Path:         updatedImage.OSSPath,
			URL:          updatedImage.OSSUrl,
			ThumbnailURL: updatedImage.ThumbnailUrl,
			Name:         updatedImage.Name,
			Size:         updatedImage.Size,
			Updated:      updatedImage.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

