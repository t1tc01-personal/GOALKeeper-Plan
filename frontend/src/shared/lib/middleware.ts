import { NextRequest } from 'next/server'
import { jwtDecode } from 'jwt-decode'
import { RoleEmployee } from '../constant/permission'

export async function verifyInMiddleware(req: NextRequest): Promise<boolean> {
  // Check for accessToken cookie (used in this app)
  const token = req.cookies.get('accessToken')?.value
  return !!token
}

export async function getUserRole(req: NextRequest): Promise<RoleEmployee[]> {
  const token = req.cookies.get('accessToken')?.value
  if (!token) return []
  try {
    const decoded = jwtDecode<{ roles?: RoleEmployee[] }>(token)
    return decoded?.roles || []
  } catch {
    return []
  }
}

