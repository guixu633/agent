import axios from 'axios';
import type {
  ImageGenerateRequest,
  ImageGenerateResponse,
  ApiResponse,
} from '@/types/image';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

// 创建 axios 实例
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 120000, // 图片生成可能需要较长时间
  headers: {
    'Content-Type': 'application/json',
  },
});

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API 请求错误:', error);
    return Promise.reject(error);
  }
);

/**
 * 图片服务类
 */
class ImageService {
  /**
   * 生成图片
   */
  async generateImage(
    request: ImageGenerateRequest
  ): Promise<ImageGenerateResponse> {
    const response = await apiClient.post<ApiResponse<ImageGenerateResponse>>(
      '/image/generate',
      request
    );

    if (response.data.code !== 0) {
      throw new Error(response.data.message || '生成图片失败');
    }

    return response.data.data!;
  }

  /**
   * 将文件转换为 Base64
   */
  async fileToBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => {
        const result = reader.result as string;
        resolve(result);
      };
      reader.onerror = reject;
      reader.readAsDataURL(file);
    });
  }

  /**
   * 批量转换文件为 Base64
   */
  async filesToBase64(files: File[]): Promise<string[]> {
    return Promise.all(files.map((file) => this.fileToBase64(file)));
  }
}

// 导出单例
export const imageService = new ImageService();

