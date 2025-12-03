// 消息类型
export type MessageType = 'text' | 'image';

// 对话消息
export interface Message {
  role: string; // 角色: "user" | "assistant"
  type: MessageType; // 消息类型: "text" | "image"
  content?: string; // 文本内容 (type="text" 时有效)
  url?: string; // 图片 URL (type="image" 时有效)
}

// 图片生成请求
export interface ImageGenerateRequest {
  prompt: string;
  images?: string[]; // OSS 中的图片路径列表
  workspace?: string; // 工作区名称（可选，用于生成图片存储）
  messages?: Message[]; // 完整的对话历史 (可选，用于记录)
  enable_web_search?: boolean; // 是否启用联网搜索（默认 false）
}

// 图片上传响应
export interface ImageUploadResponse {
  path: string; // OSS 中的图片路径
  url: string; // 图片访问 URL
}

// 生成的内容片段
export interface GeneratePart {
  type: 'text' | 'image';
  text?: string;
  image?: GeneratedImage;
}

// 生成的图片
export interface GeneratedImage {
  data?: string; // Base64 编码（可选，用于兼容）
  mimeType: string;
  path?: string; // OSS 中的图片路径
  url?: string; // 图片访问 URL
}

// 图片生成响应
export interface ImageGenerateResponse {
  parts: GeneratePart[];
}

// 图片信息
export interface ImageInfo {
  id: number; // 图片 ID
  path: string; // OSS 中的路径
  url: string; // 图片访问 URL（原图）
  thumbnail_url: string; // 缩略图访问 URL
  name: string; // 文件名
  size: number; // 文件大小（字节）
  updated: string; // 最后更新时间（ISO 8601 格式）
  source_type: 'upload' | 'generate'; // 来源类型
  prompt?: string; // 生成时的提示词
  ref_images?: string[]; // 生成时的引用图片
  message_list?: Message[]; // 生成时的对话历史
}

// 列出工作区图片响应
export interface ListWorkspaceImagesResponse {
  images: ImageInfo[];
}

// 删除图片请求
export interface DeleteImageRequest {
  path: string; // OSS 中的图片路径
}

// 重命名图片请求
export interface RenameImageRequest {
  path: string; // OSS 中的图片路径
  new_name: string; // 新文件名
  workspace: string; // 工作区名称
}

// 重命名图片响应
export interface RenameImageResponse {
  image: ImageInfo; // 重命名后的图片信息
}

// 获取图片详情响应
export interface GetImageDetailResponse {
  image: ImageInfo; // 图片详细信息（包含 message_list）
}

// API 响应包装
export interface ApiResponse<T> {
  code: number;
  message: string;
  data?: T;
}
