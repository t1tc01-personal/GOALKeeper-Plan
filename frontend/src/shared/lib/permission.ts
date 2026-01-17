import { RoleEmployee, EMPLOYEE_ROLE } from '../constant/permission'

export function hasPermission(userRoles: RoleEmployee[], requiredRoles: RoleEmployee[]): boolean {
  if (requiredRoles.length === 0) return true
  return requiredRoles.some(role => userRoles.includes(role))
}

export function isAdmin(role: RoleEmployee): boolean {
  return role === EMPLOYEE_ROLE.ADMIN
}

export function isEditor(role: RoleEmployee): boolean {
  return role === EMPLOYEE_ROLE.EDITOR
}

export function isEmployee(role: RoleEmployee): boolean {
  return role === EMPLOYEE_ROLE.EMPLOYEE
}

