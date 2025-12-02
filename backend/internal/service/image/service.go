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
	"github.com/guixu633/agent/backend/pkg/thumbnail"
	"google.golang.org/genai"
)

const (
	DefaultModel = "gemini-3-pro-image-preview"
)

// Service 图片服务
type Service struct {
	genaiClient *genai.Client
	ossClient   *oss.Client
}

// NewService 创建图片服务实例
func NewService(genaiClient *genai.Client, ossClient *oss.Client) *Service {
	return &Service{
		genaiClient: genaiClient,
		ossClient:   ossClient,
	}
}

// UploadImage 上传图片到 OSS
func (s *Service) UploadImage(file io.Reader, filename string, workspace string) (*model.ImageUploadResponse, error) {
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
	err = s.uploadThumbnail(imageData, filename, workspace)
	if err != nil {
		// 缩略图生成失败不影响主流程，只记录错误
		// 可以选择记录日志，这里简化处理
	}

	// 获取访问 URL
	url := s.ossClient.GetImageURL(path)

	return &model.ImageUploadResponse{
		Path: path,
		URL:  url,
	}, nil
}

// uploadThumbnail 生成并上传缩略图
func (s *Service) uploadThumbnail(imageData []byte, filename string, workspace string) error {
	// 检测 MIME 类型
	mimeType := s.detectMimeTypeFromFilename(filename)

	// 生成缩略图
	thumbnailData, err := thumbnail.GenerateThumbnail(imageData, mimeType)
	if err != nil {
		return fmt.Errorf("生成缩略图失败: %w", err)
	}

	// 获取缩略图文件名
	thumbnailFilename := thumbnail.GetThumbnailFilename(filename)

	// 上传缩略图
	_, err = s.ossClient.UploadImage(bytes.NewReader(thumbnailData), thumbnailFilename, workspace)
	if err != nil {
		return fmt.Errorf("上传缩略图失败: %w", err)
	}

	return nil
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

			// 上传原图到 OSS
			path, err := s.ossClient.UploadImage(bytes.NewReader(imageData), filename, workspace)
			if err != nil {
				return nil, fmt.Errorf("上传生成的图片到 OSS 失败: %w", err)
			}

			// 生成并上传缩略图
			err = s.uploadThumbnail(imageData, filename, workspace)
			if err != nil {
				// 缩略图生成失败不影响主流程
			}

			// 获取访问 URL
			url := s.ossClient.GetImageURL(path)

			result.Parts = append(result.Parts, model.GeneratePart{
				Type: "image",
				Image: &model.GeneratedImage{
					MimeType: mimeType,
					Path:     path,
					URL:      url,
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

// ListWorkspaceImages 列出工作区的所有图片（不包括缩略图）
func (s *Service) ListWorkspaceImages(ctx context.Context, workspace string) (*model.ListWorkspaceImagesResponse, error) {
	// 从 OSS 获取图片列表
	ossImages, err := s.ossClient.ListWorkspaceImages(workspace)
	if err != nil {
		return nil, fmt.Errorf("列出工作区图片失败: %w", err)
	}

	// 转换为 model 格式，过滤掉缩略图，并添加缩略图 URL
	images := make([]model.ImageInfo, 0, len(ossImages))
	for _, ossImage := range ossImages {
		// 跳过缩略图
		if thumbnail.IsThumbnail(ossImage.Name) {
			continue
		}

		// 获取缩略图文件名和路径
		thumbnailFilename := thumbnail.GetThumbnailFilename(ossImage.Name)
		// 构建缩略图路径（替换文件名部分）
		parts := strings.Split(ossImage.Path, "/")
		if len(parts) > 0 {
			parts[len(parts)-1] = thumbnailFilename
		}
		thumbnailPath := strings.Join(parts, "/")
		thumbnailURL := s.ossClient.GetImageURL(thumbnailPath)

		images = append(images, model.ImageInfo{
			Path:         ossImage.Path,
			URL:          ossImage.URL,
			ThumbnailURL: thumbnailURL,
			Name:         ossImage.Name,
			Size:         ossImage.Size,
			Updated:      ossImage.Updated,
		})
	}

	return &model.ListWorkspaceImagesResponse{
		Images: images,
	}, nil
}

// DeleteImage 删除图片（同时删除对应的缩略图）
func (s *Service) DeleteImage(ctx context.Context, path string) error {
	// 删除原图
	err := s.ossClient.DeleteImage(path)
	if err != nil {
		return fmt.Errorf("删除图片失败: %w", err)
	}

	// 获取文件名并生成缩略图路径
	// path 格式: image/workspace/filename
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		filename := parts[len(parts)-1]
		// 如果不是缩略图，则删除对应的缩略图
		if !thumbnail.IsThumbnail(filename) {
			thumbnailFilename := thumbnail.GetThumbnailFilename(filename)
			thumbnailPath := strings.Join(append(parts[:len(parts)-1], thumbnailFilename), "/")
			// 删除缩略图（忽略错误，因为缩略图可能不存在）
			s.ossClient.DeleteImage(thumbnailPath)
		}
	}

	return nil
}

// RenameImage 重命名图片（同时重命名对应的缩略图）
func (s *Service) RenameImage(ctx context.Context, req *model.RenameImageRequest) (*model.RenameImageResponse, error) {
	// 验证新文件名
	if req.NewName == "" {
		return nil, fmt.Errorf("新文件名不能为空")
	}

	// 重命名原图
	newPath, err := s.ossClient.RenameImage(req.Path, req.NewName, req.Workspace)
	if err != nil {
		return nil, fmt.Errorf("重命名图片失败: %w", err)
	}

	// 获取原文件名并重命名对应的缩略图
	parts := strings.Split(req.Path, "/")
	if len(parts) > 0 {
		oldFilename := parts[len(parts)-1]
		// 如果不是缩略图，则重命名对应的缩略图
		if !thumbnail.IsThumbnail(oldFilename) {
			oldThumbnailFilename := thumbnail.GetThumbnailFilename(oldFilename)
			newThumbnailFilename := thumbnail.GetThumbnailFilename(req.NewName)
			oldThumbnailPath := strings.Join(append(parts[:len(parts)-1], oldThumbnailFilename), "/")
			// 重命名缩略图（忽略错误，因为缩略图可能不存在）
			s.ossClient.RenameImage(oldThumbnailPath, newThumbnailFilename, req.Workspace)
		}
	}

	// 获取新图片信息
	// 从新路径中提取文件名
	fileName := req.NewName

	// 获取图片 URL
	url := s.ossClient.GetImageURL(newPath)

	// 获取缩略图路径和 URL
	thumbnailFilename := thumbnail.GetThumbnailFilename(fileName)
	newPathParts := strings.Split(newPath, "/")
	if len(newPathParts) > 0 {
		newPathParts[len(newPathParts)-1] = thumbnailFilename
	}
	thumbnailPath := strings.Join(newPathParts, "/")
	thumbnailURL := s.ossClient.GetImageURL(thumbnailPath)

	// 获取文件信息（大小和更新时间）
	// 注意：OSS 的 ListObjects 可以获取文件信息，但这里简化处理
	// 如果需要完整信息，可以调用 ListWorkspaceImages 然后查找对应的图片
	imageInfo := model.ImageInfo{
		Path:         newPath,
		URL:          url,
		ThumbnailURL: thumbnailURL,
		Name:         fileName,
		Size:         0, // 如果需要，可以从 OSS 获取
		Updated:      time.Now().Format(time.RFC3339),
	}

	return &model.RenameImageResponse{
		Image: imageInfo,
	}, nil
}
