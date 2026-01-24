import { z } from 'zod'

export function createZodSchema<T extends z.ZodTypeAny>(schema: T): T {
  return schema
}

export const paginationSchema = z.object({
  page: z.number().int().positive().optional(),
  limit: z.number().int().positive().max(100).optional(),
})

export const idSchema = z.object({
  id: z.string().min(1),
})

