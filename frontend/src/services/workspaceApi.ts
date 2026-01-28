// Workspace API client for frontend
import { getAuthTokenClient } from '@/shared/lib/authCookieClient';
import { useAuthStore } from '@/features/auth/store/authStore';

// Get API URL at runtime to ensure environment variables are available
// Automatically appends /api/v1 if not already present
function getApiBaseUrl(): string {
  const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000';
  // Remove trailing slash if present
  const cleanBaseUrl = baseUrl.replace(/\/$/, '');
  // Append /api/v1 if not already present
  if (!cleanBaseUrl.endsWith('/api/v1')) {
    return `${cleanBaseUrl}/api/v1`;
  }
  return cleanBaseUrl;
}

const API_BASE_URL = getApiBaseUrl();

// Helper to get authentication token
function getAuthToken(): string | null {
  if (typeof window === 'undefined') return null;
  
  // Try to get from cookie first (preferred method)
  const cookieToken = getAuthTokenClient();
  if (cookieToken) return cookieToken;
  
  // Fallback to auth store
  try {
    const authStore = useAuthStore.getState();
    return authStore.accessToken;
  } catch (e) {
    return null;
  }
}

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
    // Ensure endpoint starts with / if baseUrl doesn't end with /
    const cleanEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
    // Get baseUrl at runtime to ensure it's correct
    const baseUrl = this.baseUrl || getApiBaseUrl();
    const url = `${baseUrl}${cleanEndpoint}`;
    
    // Debug logging (remove in production)
    if (process.env.NODE_ENV === 'development') {
      console.log('[API Request]', options.method || 'GET', url);
      console.log('[API Base URL]', baseUrl);
      console.log('[API Endpoint]', cleanEndpoint);
    }
    
    // Build headers
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    
    // Merge existing headers if provided
    if (options.headers) {
      if (options.headers instanceof Headers) {
        options.headers.forEach((value, key) => {
          headers[key] = value;
        });
      } else if (Array.isArray(options.headers)) {
        options.headers.forEach(([key, value]) => {
          if (key && value) {
            headers[String(key)] = String(value);
          }
        });
      } else {
        // Handle Record<string, string> or HeadersInit object
        const headerObj = options.headers as Record<string, string>;
        Object.keys(headerObj).forEach(key => {
          if (headerObj[key]) {
            headers[key] = String(headerObj[key]);
          }
        });
      }
    }
    
    // Add Authorization header with Bearer token
    const token = getAuthToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
      if (process.env.NODE_ENV === 'development') {
        console.log('[API Auth] Token included in request');
      }
    } else {
      if (process.env.NODE_ENV === 'development') {
        console.warn('[API Auth] No token found - request may fail with 401');
      }
    }
    
    // Try to get user ID from auth store for X-User-ID header (if needed by backend)
    if (typeof window !== 'undefined') {
      try {
        const authStore = useAuthStore.getState();
        const userId = authStore?.user?.id;
        if (userId) {
          headers['X-User-ID'] = String(userId);
        }
      } catch (e) {
        // Ignore errors
      }
    }
    
    const response = await fetch(url, {
      ...options,
      headers: headers as HeadersInit,
    });

    const data = await response.json() as ApiResponse<T>;
    
    if (!response.ok || !data.success) {
      throw new Error(data.error?.message || data.message || 'API request failed');
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
      body: JSON.stringify({
        workspaceId: req.workspaceId,  // Backend expects camelCase
        title: req.title,
        parentPageId: req.parentPageId,  // Backend expects camelCase
      }),
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

// Export a singleton instance with explicit base URL
// This ensures the base URL is always correct, even if env vars aren't loaded yet
export const workspaceApi = new WorkspaceApiClient(getApiBaseUrl());
export default workspaceApi;
