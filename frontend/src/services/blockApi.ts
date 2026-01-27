// Block API client for frontend
import { workspaceApi } from './workspaceApi';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

export interface Block {
  id: string;
  page_id: string;
  parent_block_id?: string;
  type_id: string;
  content?: string;
  metadata?: Record<string, any>;
  rank: number;
  created_at: string;
  updated_at: string;
}

export interface CreateBlockRequest {
  pageId: string;
  type: string;
  content?: string;
  position: number;
}

export interface UpdateBlockRequest {
  content?: string;
}

export interface ReorderBlocksRequest {
  pageId: string;
  blockIds: string[];
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

  // Block endpoints
  async createBlock(req: CreateBlockRequest): Promise<Block> {
    const response = await this.request<Block>('/notion/blocks', {
      method: 'POST',
      body: JSON.stringify(req),
    });
    return response.data!;
  }

  async getBlock(id: string): Promise<Block> {
    const response = await this.request<Block>(`/notion/blocks/${id}`);
    return response.data!;
  }

  async listBlocks(pageId: string): Promise<Block[]> {
    const response = await this.request<Block[]>(
      `/notion/blocks?pageId=${encodeURIComponent(pageId)}`
    );
    return response.data || [];
  }

  async updateBlock(id: string, req: UpdateBlockRequest): Promise<Block> {
    const response = await this.request<Block>(`/notion/blocks/${id}`, {
      method: 'PUT',
      body: JSON.stringify(req),
    });
    return response.data!;
  }

  async deleteBlock(id: string): Promise<void> {
    await this.request(`/notion/blocks/${id}`, {
      method: 'DELETE',
    });
  }

  async reorderBlocks(req: ReorderBlocksRequest): Promise<void> {
    await this.request('/notion/blocks/reorder', {
      method: 'POST',
      body: JSON.stringify(req),
    });
  }
}

// Export a singleton instance
export const blockApi = new BlockApiClient();
export default blockApi;
