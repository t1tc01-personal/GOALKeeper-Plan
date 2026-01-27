/**
 * Shared interfaces for API responses and error handling
 */

export interface SimpleErrorResponse {
  error: string;
  message?: string;
  code?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  hasMore: boolean;
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
  status?: string;
}

export interface User {
  id: string;
  email: string;
  name?: string;
  roles?: string[];
  created_at: string;
  updated_at: string;
}

export interface Permission {
  id: string;
  user_id: string;
  resource_id: string;
  resource_type: string;
  role: 'viewer' | 'editor' | 'owner';
  created_at: string;
}

export interface ErrorDetail {
  field?: string;
  message: string;
  code?: string;
}

export interface ValidationError {
  errors: ErrorDetail[];
  message?: string;
}
