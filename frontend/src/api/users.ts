import { api } from '@/lib/api'
import type { ApiUser } from '@/api/auth'
import type { ApiDataMessageResponse, UUID } from '@/api/types'

export interface CurrentUserResponse {
  message: string
  status: string
  data: ApiUser
}

export async function getCurrentUser() {
  const { data } = await api.get<CurrentUserResponse>('/users/me')
  return data
}

export interface UpdateOwnProfilePayload {
  email?: string
  full_name?: string
}

export type UpdateOwnProfileResponse = ApiDataMessageResponse<ApiUser>

export async function updateOwnProfile(
  userId: UUID,
  payload: UpdateOwnProfilePayload,
) {
  const { data } = await api.patch<UpdateOwnProfileResponse>(
    `/users/${userId}`,
    payload,
  )
  return data
}
