import axios from 'axios';
import type {
  Workspace,
  ListWorkspacesResponse,
  CreateWorkspaceRequest,
  CreateWorkspaceResponse,
  DeleteWorkspaceRequest,
  ApiResponse,
} from '@/types/workspace';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

// 创建 axios 实例
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
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
 * 工作区服务类
 */
class WorkspaceService {
  /**
   * 列出所有工作区
   */
  async listWorkspaces(): Promise<ListWorkspacesResponse> {
    const response = await apiClient.get<ApiResponse<ListWorkspacesResponse>>(
      '/workspace'
    );

    if (response.data.code !== 0) {
      throw new Error(response.data.message || '获取工作区列表失败');
    }

    return response.data.data!;
  }

  /**
   * 创建工作区
   */
  async createWorkspace(
    request: CreateWorkspaceRequest
  ): Promise<CreateWorkspaceResponse> {
    const response = await apiClient.post<ApiResponse<CreateWorkspaceResponse>>(
      '/workspace',
      request
    );

    if (response.data.code !== 0) {
      throw new Error(response.data.message || '创建工作区失败');
    }

    return response.data.data!;
  }

  /**
   * 删除工作区
   */
  async deleteWorkspace(request: DeleteWorkspaceRequest): Promise<void> {
    const response = await apiClient.delete<ApiResponse<void>>('/workspace', {
      data: request,
    });

    if (response.data.code !== 0) {
      throw new Error(response.data.message || '删除工作区失败');
    }
  }
}

// 导出单例
export const workspaceService = new WorkspaceService();

