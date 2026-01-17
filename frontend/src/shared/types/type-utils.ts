/**
 * Type utilities to replace `any` with safer alternatives
 */

// Safe alternative to `any` when you truly don't know the type
export type Unknown = unknown

// For objects with dynamic keys
export type AnyObject = Record<string, unknown>

// For objects with dynamic keys and specific value type
export type RecordOf<T = unknown> = Record<string, T>

// For JSON-like data structures
export type JsonValue = string | number | boolean | null | JsonObject | JsonArray
export type JsonObject = { [key: string]: JsonValue }
export type JsonArray = JsonValue[]

// For function parameters when you don't know the exact type
export type AnyFunction = (...args: unknown[]) => unknown

// For arrays with unknown items
export type AnyArray = unknown[]

// For React component props
export type AnyProps = Record<string, unknown>

// For API responses or external data
export type SafeAny = unknown

// For query params in API calls
export type QueryParams = Record<string, string | number | boolean | undefined | null>

// For table/list filters
export type FilterParams = Record<string, string | number | boolean | string[] | number[] | undefined | null>

// For form data
export type FormValues = Record<string, string | number | boolean | File | FileList | unknown[] | undefined | null>

// Helper: Make properties partial recursively
export type DeepPartial<T> = T extends object
  ? {
      [P in keyof T]?: DeepPartial<T[P]>
    }
  : T

// Helper: Make properties required recursively
export type DeepRequired<T> = T extends object
  ? {
      [P in keyof T]-?: DeepRequired<T[P]>
    }
  : T

// Helper: Make all properties nullable
export type Nullable<T> = {
  [P in keyof T]: T[P] | null
}

// Helper: Make all properties optional and nullable
export type Maybe<T> = {
  [P in keyof T]?: T[P] | null | undefined
}

// For error handling
export type ErrorLike = Error | { message: string; [key: string]: unknown }

// For async operations
export type AsyncResult<T> = Promise<T>
export type AsyncData<T> = T | Promise<T>

// For constructors
export type Constructor<T = unknown> = new (...args: unknown[]) => T

// For class instance
export type Instance<T extends Constructor> = T extends Constructor<infer R> ? R : never

