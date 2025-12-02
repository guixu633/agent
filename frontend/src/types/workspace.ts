// 工作区
export interface Workspace {
  name: string;
  is_current?: boolean;
  created_at?: string;
}

// 列出工作区响应
export interface ListWorkspacesResponse {
  workspaces: Workspace[];
}

// 创建工作区请求
export interface CreateWorkspaceRequest {
  name: string;
}

// 创建工作区响应
export interface CreateWorkspaceResponse {
  workspace: Workspace;
}

// 删除工作区请求
export interface DeleteWorkspaceRequest {
  name: string;
}

// 设置当前工作区请求
export interface SetCurrentWorkspaceRequest {
  name: string;
}

// 设置当前工作区响应
export interface SetCurrentWorkspaceResponse {
  workspace: Workspace;
}

// 获取当前工作区响应
export interface GetCurrentWorkspaceResponse {
  workspace: Workspace | null;
}

