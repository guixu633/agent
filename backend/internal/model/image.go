package model


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
	Data     string `json:"data,omitempty"`     // Base64 编码的图片数据（可选，用于兼容）
	MimeType string `json:"mimeType"`           // 图片 MIME 类型
	Path     string `json:"path,omitempty"`     // OSS 中的图片路径
	URL      string `json:"url,omitempty"`      // 图片访问 URL
}

// ImageUploadRequest 图片上传请求
type ImageUploadRequest struct {
	Workspace string `form:"workspace" binding:"required"` // 工作区名称
}

// ImageUploadResponse 图片上传响应
type ImageUploadResponse struct {
	Path string `json:"path"` // OSS 中的图片路径
	URL  string `json:"url"`  // 图片访问 URL
}

// ImageGenerateRequest 图片生成请求
type ImageGenerateRequest struct {
	Prompt    string   `json:"prompt" binding:"required"`
	Images    []string `json:"images"`              // OSS 中的图片路径列表
	Workspace string   `json:"workspace"`           // 工作区名称（可选，用于生成图片存储）
}

// ListWorkspaceImagesRequest 列出工作区图片请求
type ListWorkspaceImagesRequest struct {
	Workspace string `form:"workspace" binding:"required"` // 工作区名称
}

// ImageInfo 图片信息
type ImageInfo struct {
	Path         string `json:"path"`          // OSS 中的路径
	URL          string `json:"url"`           // 图片访问 URL（原图）
	ThumbnailURL string `json:"thumbnail_url"` // 缩略图访问 URL
	Name         string `json:"name"`          // 文件名
	Size         int64  `json:"size"`         // 文件大小（字节）
	Updated      string `json:"updated"`      // 最后更新时间（ISO 8601 格式）
}

// ListWorkspaceImagesResponse 列出工作区图片响应
type ListWorkspaceImagesResponse struct {
	Images []ImageInfo `json:"images"` // 图片列表
}

// DeleteImageRequest 删除图片请求
type DeleteImageRequest struct {
	Path string `json:"path" binding:"required"` // OSS 中的图片路径
}

// RenameImageRequest 重命名图片请求
type RenameImageRequest struct {
	Path      string `json:"path" binding:"required"`      // OSS 中的图片路径
	NewName   string `json:"new_name" binding:"required"`  // 新文件名
	Workspace string `json:"workspace" binding:"required"`  // 工作区名称
}

// RenameImageResponse 重命名图片响应
type RenameImageResponse struct {
	Image ImageInfo `json:"image"` // 重命名后的图片信息
}
