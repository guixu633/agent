package model

// ImageGenerateRequest 图片生成请求
type ImageGenerateRequest struct {
	Prompt string   `json:"prompt" binding:"required"`
	Images []string `json:"images"` // Base64 编码的图片数据
}

// ImageGenerateResponse 图片生成响应
type ImageGenerateResponse struct {
	Parts []GeneratePart `json:"parts"` // 有序的内容片段
}

// GeneratePart 生成内容片段
type GeneratePart struct {
	Type  string          `json:"type"`            // 类型: "text" | "image"
	Text  string          `json:"text,omitempty"`  // 文本内容 (Type="text" 时有效)
	Image *GeneratedImage `json:"image,omitempty"` // 图片信息 (Type="image" 时有效)
}

// GeneratedImage 生成的图片信息
type GeneratedImage struct {
	Data     string `json:"data"`     // Base64 编码的图片数据
	MimeType string `json:"mimeType"` // 图片 MIME 类型
}
