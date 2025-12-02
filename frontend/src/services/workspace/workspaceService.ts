import axios from 'axios';
import type {
  ListWorkspacesResponse,
  CreateWorkspaceRequest,
  CreateWorkspaceResponse,
  DeleteWorkspaceRequest,
  SetCurrentWorkspaceRequest,
  SetCurrentWorkspaceResponse,
  GetCurrentWorkspaceResponse,
} from '@/types/workspace';
import type { ApiResponse } from '@/types/image';

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

  /**
   * 设置当前工作区（切换工作区）
   */
  async setCurrentWorkspace(
    request: SetCurrentWorkspaceRequest
  ): Promise<SetCurrentWorkspaceResponse> {
    const response = await apiClient.put<ApiResponse<SetCurrentWorkspaceResponse>>(
      '/workspace/current',
      request
    );

    if (response.data.code !== 0) {
      throw new Error(response.data.message || '切换工作区失败');
    }

    return response.data.data!;
  }

  /**
   * 获取当前工作区
   */
  async getCurrentWorkspace(): Promise<GetCurrentWorkspaceResponse> {
    const response = await apiClient.get<ApiResponse<GetCurrentWorkspaceResponse>>(
      '/workspace/current'
    );

    if (response.data.code !== 0) {
      throw new Error(response.data.message || '获取当前工作区失败');
    }

    return response.data.data!;
  }
}

// 导出单例
export const workspaceService = new WorkspaceService();

