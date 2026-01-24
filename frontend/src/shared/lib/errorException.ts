import isArray from 'lodash/isArray'
import { AnyObject } from '../types'

export enum ErrorCodes {
  warning = 'warning',
  error = 'error',
}

class ErrorException {
  message: string
  code: string
  type?: ErrorCodes
  data?: unknown

  constructor({ message, code, data, type }: ErrorException) {
    this.message = message
    this.code = code
    this.data = data
    this.type = type
  }

  static hasError(json: Record<string, unknown>): boolean {
    const success = json?.['success'] ?? false
    if (!success) return true
    return false
  }

  static fromError(error: Record<string, unknown>): ErrorException {
    const message = error?.['message'] 
    console.log('message', message)
    return {
      code: error?.['status'] as string ?? 500,
      message: isArray(message) ? message.join(', ') : message as string,
      data: error?.['data'],
    }
  }
}

export class DataResponse<T> {
  success: boolean
  message: string
  data: T

  constructor({ message, success, data }: DataResponse<T>) {
    this.message = message
    this.success = success
    this.data = data
  }

  static fromResponse<T>(response: AnyObject): DataResponse<T> {
    return {
      success: response?.['success'] as boolean ?? true,
      message: response?.['message'] as string ?? '',
      data: (response?.['data'] as T) ?? (response as T),
    }
  }
}

export default ErrorException

