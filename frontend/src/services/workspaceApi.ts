// Workspace API client for frontend
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export interface Workspace {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface Page {
  id: string;
  workspace_id: string;
  parent_page_id?: string;
  title: string;
  is_planner: boolean;
  framework_id?: string;
  view_config?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface CreateWorkspaceRequest {
  name: string;
  description?: string;
}

export interface UpdateWorkspaceRequest {
  name: string;
  description?: string;
}

export interface CreatePageRequest {
  workspaceId: string;
  title: string;
  parentPageId?: string;
}

export interface UpdatePageRequest {
  title: string;
  viewConfig?: Record<string, any>;
}

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
  error?: {
    type: string;
    code: string;
    message: string;
    details?: string;
  };
  timestamp: number;
  request_id?: string;
}

class WorkspaceApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseUrl}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    const data = await response.json() as ApiResponse<T>;
    
    if (!response.ok) {
      throw new Error(data.error?.message || 'API request failed');
    }

    return data;
  }

  // Workspace endpoints
  async createWorkspace(req: CreateWorkspaceRequest): Promise<Workspace> {
    const response = await this.request<Workspace>('/notion/workspaces', {
      method: 'POST',
      body: JSON.stringify(req),
    });
    return response.data!;
  }

  async getWorkspace(id: string): Promise<Workspace> {
    const response = await this.request<Workspace>(`/notion/workspaces/${id}`);
    return response.data!;
  }

  async listWorkspaces(): Promise<Workspace[]> {
    const response = await this.request<Workspace[]>('/notion/workspaces');
    return response.data || [];
  }

  async updateWorkspace(id: string, req: UpdateWorkspaceRequest): Promise<Workspace> {
    const response = await this.request<Workspace>(`/notion/workspaces/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    });
    return response.data!;
  }

  async deleteWorkspace(id: string): Promise<void> {
    await this.request('/notion/workspaces/' + id, {
      method: 'DELETE',
    });
  }

  // Page endpoints
  async createPage(req: CreatePageRequest): Promise<Page> {
    const response = await this.request<Page>('/notion/pages', {
      method: 'POST',
      body: JSON.stringify(req),
    });
    return response.data!;
  }

  async getPage(id: string): Promise<Page> {
    const response = await this.request<Page>(`/notion/pages/${id}`);
    return response.data!;
  }

  async listPages(workspaceId: string): Promise<Page[]> {
    const response = await this.request<Page[]>(
      `/notion/pages?workspaceId=${encodeURIComponent(workspaceId)}`
    );
    return response.data || [];
  }

  async updatePage(id: string, req: UpdatePageRequest): Promise<Page> {
    const response = await this.request<Page>(`/notion/pages/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    });
    return response.data!;
  }

  async deletePage(id: string): Promise<void> {
    await this.request(`/notion/pages/${id}`, {
      method: 'DELETE',
    });
  }

  async getPageHierarchy(workspaceId: string): Promise<Page[]> {
    const response = await this.request<Page[]>(
      `/notion/pages/hierarchy?workspaceId=${encodeURIComponent(workspaceId)}`
    );
    return response.data || [];
  }

  // Sharing endpoints
  async grantAccess(pageId: string, userId: string, role: 'viewer' | 'editor' | 'owner'): Promise<ApiResponse<{ message: string; page_id: string; user_id: string; role: string }>> {
    return this.request(
      `/notion/pages/${pageId}/share`,
      {
        method: 'POST',
        body: JSON.stringify({
          user_id: userId,
          role: role,
        }),
      }
    );
  }

  async revokeAccess(pageId: string, userId: string): Promise<ApiResponse<void>> {
    return this.request(
      `/notion/pages/${pageId}/share/${userId}`,
      {
        method: 'DELETE',
      }
    );
  }

  async listCollaborators(pageId: string): Promise<ApiResponse<{ collaborators: any[]; count: number }>> {
    return this.request(
      `/notion/pages/${pageId}/collaborators`
    );
  }
}

// Export a singleton instance
export const workspaceApi = new WorkspaceApiClient();
export default workspaceApi;
