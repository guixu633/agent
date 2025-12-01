package image

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/guixu633/agent/backend/internal/model"
	"google.golang.org/genai"
)

const (
	DefaultModel = "gemini-3-pro-image-preview"
)

// Service 图片服务
type Service struct {
	genaiClient *genai.Client
}

// NewService 创建图片服务实例
func NewService(genaiClient *genai.Client) *Service {
	return &Service{
		genaiClient: genaiClient,
	}
}

// GenerateImage 生成图片
func (s *Service) GenerateImage(ctx context.Context, req *model.ImageGenerateRequest) (*model.ImageGenerateResponse, error) {
	// 构建请求内容
	contents := []*genai.Content{
		genai.NewContentFromText(req.Prompt, genai.RoleUser),
	}

	// 添加输入图片
	for _, imageBase64 := range req.Images {
		// 解析 base64 数据（可能包含 data:image/png;base64, 前缀）
		imageData, mimeType, err := s.parseBase64Image(imageBase64)
		if err != nil {
			return nil, fmt.Errorf("解析图片数据失败: %w", err)
		}

		contents = append(contents, genai.NewContentFromBytes(imageData, mimeType, genai.RoleUser))
	}

	// 调用 Gemini API
	resp, err := s.genaiClient.Models.GenerateContent(ctx, DefaultModel, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("调用 Gemini API 失败: %w", err)
	}

	// 解析响应
	result := &model.ImageGenerateResponse{
		Images: make([]model.GeneratedImage, 0),
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("模型未返回任何内容")
	}

	// 提取文本和图片
	for _, part := range resp.Candidates[0].Content.Parts {
		// 文本描述
		if part.Text != "" {
			result.Description = part.Text
		}

		// 生成的图片
		if part.InlineData != nil {
			imageBase64 := base64.StdEncoding.EncodeToString(part.InlineData.Data)
			result.Images = append(result.Images, model.GeneratedImage{
				Data:     imageBase64,
				MimeType: part.InlineData.MIMEType,
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
