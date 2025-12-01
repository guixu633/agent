package model

// ImageGenerateRequest 图片生成请求
type ImageGenerateRequest struct {
	Prompt string   `json:"prompt" binding:"required"`
	Images []string `json:"images"` // Base64 编码的图片数据
}

// ImageGenerateResponse 图片生成响应
type ImageGenerateResponse struct {
	Images      []GeneratedImage `json:"images"`
	Description string           `json:"description,omitempty"`
}

// GeneratedImage 生成的图片信息
type GeneratedImage struct {
	Data     string `json:"data"`     // Base64 编码的图片数据
	MimeType string `json:"mimeType"` // 图片 MIME 类型
}


