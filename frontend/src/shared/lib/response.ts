import { AnyObject } from "../types";

  export interface ApiError {
    code: string;
    message: string;
    details?: unknown;
  }
  
  export interface ApiResponse<T> {
    data?: T;
    error?: ApiError;
    meta?: Record<string, unknown>;
  }
  
  export function jsonOk<T>(
    data: T,
    init?: ResponseInit & { meta?: ApiResponse<T>["meta"] }
  ): Response {
    const body: ApiResponse<T> = { data, meta: init?.meta };
    return new Response(JSON.stringify(body), {
      status: init?.status ?? 200,
      headers: { "Content-Type": "application/json; charset=utf-8" },
    });
  }
  
  export function jsonError(code: number, err: ApiError): Response {
    const body: ApiResponse<never> = { error: err };
    return new Response(JSON.stringify(body), {
      status: code,
      headers: { "Content-Type": "application/json; charset=utf-8" },
    });
  }
  
  export function parseIntParam(
    value: string | string[] | null | undefined,
    fallback: number
  ): number {
    if (typeof value !== "string") return fallback;
    const n = Number.parseInt(value, 10);
    return Number.isFinite(n) && n > 0 ? n : fallback;
  }
  
  export function parseString(
    value: string | string[] | null | undefined
  ): string | undefined {
    if (typeof value !== "string") return undefined;
    const v = value.trim();
    return v.length > 0 ? v : undefined;
  }
  
  export function getError(response: AnyObject) {
    if (response && !response.ok && response.error && (response.error as ApiError).message) {
      return (response.error as ApiError).message;
    }
    return null;
  }

