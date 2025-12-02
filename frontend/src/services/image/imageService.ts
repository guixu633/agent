import axios from 'axios';
import type {
  ImageGenerateRequest,
  ImageGenerateResponse,
  ImageUploadResponse,
  ListWorkspaceImagesResponse,
  DeleteImageRequest,
  RenameImageRequest,
  RenameImageResponse,
  ApiResponse,
} from '@/types/image';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

// 创建 axios 实例（用于 JSON 请求）
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 120000, // 图片生成可能需要较长时间
  headers: {
    'Content-Type': 'application/json',
  },
});

// 创建用于文件上传的 axios 实例
const uploadClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 60000, // 上传超时时间
  headers: {
    'Content-Type': 'multipart/form-data',
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

// 上传客户端的响应拦截器
uploadClient.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('上传请求错误:', error);
    return Promise.reject(error);
  }
);

/**
 * 图片服务类
 */
class ImageService {
  /**
   * 上传图片到 OSS
   */
  async uploadImage(
    file: File,
    workspace: string
  ): Promise<ImageUploadResponse> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('workspace', workspace);

    const response = await uploadClient.post<ApiResponse<ImageUploadResponse>>(
      '/image/upload',
      formData
    );

    if (response.data.code !== 0) {
      throw new Error(response.data.message || '上传图片失败');
    }

    return response.data.data!;
  }

  /**
   * 批量上传图片到 OSS
   */
  async uploadImages(
    files: File[],
    workspace: string
  ): Promise<ImageUploadResponse[]> {
    return Promise.all(
      files.map((file) => this.uploadImage(file, workspace))
    );
  }

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
   * 列出工作区的所有图片
   */
  async listWorkspaceImages(
    workspace: string
  ): Promise<ListWorkspaceImagesResponse> {
    const response = await apiClient.get<ApiResponse<ListWorkspaceImagesResponse>>(
      '/image/list',
      {
        params: {
          workspace,
        },
      }
    );

    if (response.data.code !== 0) {
      throw new Error(response.data.message || '获取图片列表失败');
    }

    return response.data.data!;
  }

  /**
   * 删除图片
   */
  async deleteImage(request: DeleteImageRequest): Promise<void> {
    const response = await apiClient.delete<ApiResponse<void>>('/image', {
      data: request,
    });

    if (response.data.code !== 0) {
      throw new Error(response.data.message || '删除图片失败');
    }
  }

  /**
   * 重命名图片
   */
  async renameImage(
    request: RenameImageRequest
  ): Promise<RenameImageResponse> {
    const response = await apiClient.post<ApiResponse<RenameImageResponse>>(
      '/image/rename',
      request
    );

    if (response.data.code !== 0) {
      throw new Error(response.data.message || '重命名图片失败');
    }

    return response.data.data!;
  }
}

// 导出单例
export const imageService = new ImageService();


