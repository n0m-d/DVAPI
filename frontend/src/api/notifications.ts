import { api } from '@/lib/api'
import type { ApiResponse, Pagination, UUID } from '@/api/types'

interface StatusResponse {
  status: string
}

export type ApiNotificationType = 'grade' | 'assignment' | 'announcement' | 'system' | 'enrollment'

export interface ApiNotification {
  id: UUID
  type: ApiNotificationType
  title: string
  body: string
  read: boolean
  created_at: string
  course_id?: UUID
}

export interface NotificationsListData {
  notifications: ApiNotification[]
  unread_count: number
  pagination: Pagination
}

export type NotificationsListResponse = ApiResponse<NotificationsListData>
export type NotificationResponse = ApiResponse<ApiNotification>

export interface ListNotificationsParams {
  page?: number
  page_size?: number
  unread?: boolean
}

export async function listNotifications(params: ListNotificationsParams = {}) {
  const { data } = await api.get<NotificationsListResponse>('/notifications', {
    params: {
      page: params.page ?? 1,
      page_size: params.page_size ?? 20,
      ...(params.unread ? { unread: true } : {}),
    },
  })
  return data
}

export async function markNotificationRead(id: UUID) {
  const { data } = await api.post<NotificationResponse>(`/notifications/${id}/read`)
  return data
}

export async function markAllNotificationsRead() {
  const { data } = await api.post<StatusResponse>('/notifications/read-all')
  return data
}

export async function deleteNotification(id: UUID) {
  const { data } = await api.delete<StatusResponse>(`/notifications/${id}`)
  return data
}
