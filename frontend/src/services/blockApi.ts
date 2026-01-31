// Block API client for frontend
import { workspaceApi } from './workspaceApi';
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

export interface Block {
  id: string;
  page_id?: string; // For internal use
  pageId?: string; // Backend field name
  parent_block_id?: string;
  type_id?: string; // For internal use (normalized from type)
  type: string; // Backend field name
  content?: string;
  metadata?: Record<string, any>;
  blockConfig?: Record<string, any>; // Backend field name
  position: number; // Backend field name
  created_at: string;
  updated_at: string;
}

// Helper to normalize block from backend response
function normalizeBlock(block: any): Block {
  return {
    ...block,
    page_id: block.pageId || block.page_id,
    type_id: block.type || block.type_id,
    // Ensure position is set
    position: block.position ?? block.rank ?? 0,
    // Ensure metadata is populated from blockConfig (backend uses blockConfig, frontend components often use metadata)
    metadata: block.metadata || block.blockConfig || {},
  };
}

export interface CreateBlockRequest {
  pageId: string;
  type: string;
  content?: string;
  position: number;
  parent_block_id?: string | null;
  blockConfig?: Record<string, any>;
}

export interface UpdateBlockRequest {
  content?: string;
  type?: string;
  position?: number;
  parent_block_id?: string | null;
  blockConfig?: Record<string, any>;
}

export interface ReorderBlocksRequest {
  pageId: string;
  blockIds: string[];
}

export interface BatchSyncRequest {
  creates?: Array<{
    pageId: string;
    type: string;
    content?: string;
    position: number;
    parent_block_id?: string | null;
    blockConfig?: Record<string, any>;
    tempId: string;
  }>;
  updates?: Array<{
    id: string;
    content?: string;
    type?: string;
    position?: number;
    parent_block_id?: string | null;
    blockConfig?: Record<string, any>;
  }>;
  deletes?: string[];
}

export interface BatchSyncResponse {
  creates: Array<{
    tempId: string;
    block: Block;
  }>;
  updates: Array<{
    id: string;
    block: Block;
  }>;
  deletes: string[];
  errors?: Array<{
    operationId: string;
    type: 'create' | 'update' | 'delete';
    error: string;
  }>;
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

class BlockApiClient {
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

  // Block endpoints
  async createBlock(req: CreateBlockRequest): Promise<Block> {
    const requestBody: Record<string, any> = {
      pageId: req.pageId,
      type: req.type,
      position: req.position,
    };
    if (req.content !== undefined) requestBody.content = req.content;
    if (req.parent_block_id !== undefined) requestBody.parent_block_id = req.parent_block_id;
    if (req.blockConfig !== undefined) requestBody.blockConfig = req.blockConfig;

    const response = await this.request<Block>('/notion/blocks', {
      method: 'POST',
      body: JSON.stringify(requestBody),
    });
    return normalizeBlock(response.data!);
  }

  async getBlock(id: string): Promise<Block> {
    const response = await this.request<Block>(`/notion/blocks/${id}`);
    return normalizeBlock(response.data!);
  }

  async listBlocks(pageId: string): Promise<Block[]> {
    const response = await this.request<Block[]>(
      `/notion/blocks?pageId=${encodeURIComponent(pageId)}`
    );
    return (response.data || []).map(normalizeBlock);
  }

  async updateBlock(id: string, req: UpdateBlockRequest): Promise<Block> {
    const requestBody: Record<string, any> = {};
    if (req.content !== undefined) requestBody.content = req.content;
    if (req.type !== undefined) requestBody.type = req.type;
    if (req.position !== undefined) requestBody.position = req.position;
    if (req.parent_block_id !== undefined) requestBody.parent_block_id = req.parent_block_id;
    if (req.blockConfig !== undefined) requestBody.blockConfig = req.blockConfig;

    const response = await this.request<Block>(`/notion/blocks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(requestBody),
    });
    return normalizeBlock(response.data!);
  }

  async deleteBlock(id: string): Promise<void> {
    await this.request(`/notion/blocks/${id}`, {
      method: 'DELETE',
    });
  }

  async reorderBlocks(req: ReorderBlocksRequest): Promise<void> {
    // Backend expects pageId as a query parameter
    const url = `/notion/blocks/reorder?pageId=${encodeURIComponent(req.pageId)}`;
    await this.request(url, {
      method: 'POST',
      body: JSON.stringify({
        blockIds: req.blockIds,
      }),
    });
  }

  async batchSync(req: BatchSyncRequest): Promise<BatchSyncResponse> {
    const response = await this.request<BatchSyncResponse>('/notion/blocks/batch', {
      method: 'POST',
      body: JSON.stringify(req),
    });

    // Normalize blocks in response
    const normalizedResponse: BatchSyncResponse = {
      creates: (response.data?.creates || []).map((create) => ({
        ...create,
        block: normalizeBlock(create.block),
      })),
      updates: (response.data?.updates || []).map((update) => ({
        ...update,
        block: normalizeBlock(update.block),
      })),
      deletes: response.data?.deletes || [],
      errors: response.data?.errors,
    };

    return normalizedResponse;
  }
}

// Export a singleton instance with explicit base URL
// This ensures the base URL is always correct, even if env vars aren't loaded yet
export const blockApi = new BlockApiClient(getApiBaseUrl());
export default blockApi;
