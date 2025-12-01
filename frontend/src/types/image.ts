// 图片生成请求
export interface ImageGenerateRequest {
  prompt: string;
  images?: string[]; // Base64 编码的图片
}

// 生成的内容片段
export interface GeneratePart {
  type: 'text' | 'image';
  text?: string;
  image?: GeneratedImage;
}

// 生成的图片
export interface GeneratedImage {
  data: string; // Base64 编码
  mimeType: string;
}

// 图片生成响应
export interface ImageGenerateResponse {
  parts: GeneratePart[];
}

// API 响应包装
export interface ApiResponse<T> {
  code: number;
  message: string;
  data?: T;
}
