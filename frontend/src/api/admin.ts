import { api } from '@/lib/api'
import type { ApiUser } from '@/api/auth'
import type { ApiDataMessageResponse, ApiResponse, Pagination, UUID } from '@/api/types'
import type { UserRole } from '@/types/roles'

export interface AdminStats {
  users: number
  students: number
  instructors: number
  courses: number
  enrollments: number
  assignments: number
  submissions: number
}

export type AdminStatsResponse = ApiResponse<AdminStats>

export async function getAdminStats() {
  const { data } = await api.get<AdminStatsResponse>('/admin/stats')
  return data
}

export interface ListAdminUsersParams {
  search?: string
  role?: UserRole
  page?: number
  page_size?: number
}

export interface AdminUsersData {
  users: ApiUser[]
  pagination: Pagination
}

export type AdminUsersResponse = ApiResponse<AdminUsersData>

export async function listAdminUsers(params: ListAdminUsersParams = {}) {
  const { data } = await api.get<AdminUsersResponse>('/admin/users', {
    params: {
      ...(params.search?.trim() ? { search: params.search.trim() } : {}),
      ...(params.role ? { role: params.role } : {}),
      page: params.page ?? 1,
      page_size: params.page_size ?? 10,
    },
  })
  return data
}

export interface CreateAdminUserPayload {
  email: string
  password: string
  full_name: string
  role: UserRole
}

export interface UpdateAdminUserPayload {
  email?: string
  full_name?: string
  role?: UserRole
}

export type AdminUserMutationResponse = ApiDataMessageResponse<ApiUser>

export async function createAdminUser(payload: CreateAdminUserPayload) {
  const { data } = await api.post<AdminUserMutationResponse>('/admin/users', payload)
  return data
}

export async function updateAdminUser(
  userId: UUID,
  payload: UpdateAdminUserPayload,
) {
  const { data } = await api.patch<AdminUserMutationResponse>(
    `/admin/users/${userId}`,
    payload,
  )
  return data
}

export interface AdminLogsResponse {
  status: string
  file: string
  lines: string | number
  fetchedAt: string
  data: string
}

export async function getAdminLogs(lines = 100) {
  const { data } = await api.get<AdminLogsResponse>('/admin/logs', {
    params: { lines },
  })
  return data
}
