export const EMPLOYEE_ROLE = {
    ADMIN: 'admin',
    EDITOR: 'editor',
    EMPLOYEE: 'employee',
    NULL: null,
  } as const
  
export type RoleEmployee = (typeof EMPLOYEE_ROLE)[keyof typeof EMPLOYEE_ROLE]

