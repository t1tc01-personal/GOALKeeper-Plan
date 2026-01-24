import { NextRequest } from 'next/server'
import { jwtDecode } from 'jwt-decode'
import { RoleEmployee } from '../constant/permission'

export async function verifyInMiddleware(req: NextRequest): Promise<boolean> {
  const token = req.cookies.get('infoToken')?.value
  return !!token
}

export async function getUserRole(req: NextRequest): Promise<RoleEmployee[]> {
  const token = req.cookies.get('infoToken')?.value
  if (!token) return []
  const decoded = jwtDecode<{ employeeInfo: { roles: RoleEmployee[] } }>(token ?? '')
  return decoded?.employeeInfo?.roles || []
}

