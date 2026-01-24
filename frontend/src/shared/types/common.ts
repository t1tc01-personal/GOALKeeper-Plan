import { Either } from '../utils/either'
import ErrorException, { DataResponse } from '../lib/errorException'
import { FilterParams } from './type-utils'

export type BaseRecord = Record<string, unknown>

export type PEResult<T> = Promise<Either<ErrorException, T>>
export type ResponseResult<T> = Promise<Either<ErrorException, DataResponse<T>>>
export type ResponseBufferResult = Promise<ResponseResult<Buffer>>

export type StatusInfo = {
  text: string
  bg: string
  textColor: string
}

export type INote = {
  comments: string
}

export type IDataTable<T> = {
  data: T[]
  total: number
}

export type IPagination = {
  page?: number
  limit?: number
}

