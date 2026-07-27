export type UserRole = 'student' | 'instructor' | 'admin'

export interface AuthUser {
  id: string
  name: string
  email: string
  role: UserRole
  avatar?: string
}
