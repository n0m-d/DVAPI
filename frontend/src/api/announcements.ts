import { api } from '@/lib/api'
import type { ApiResponse, UUID } from '@/api/types'

export type AnnouncementStatus = 'draft' | 'published'

export interface ApiAnnouncement {
  id: UUID
  course_id: UUID
  title: string
  content: string
  status: AnnouncementStatus
  created_by: UUID
  created_at: string
  updated_at: string
}

export interface CreateAnnouncementPayload {
  title: string
  content: string
  status: AnnouncementStatus
}

export type UpdateAnnouncementPayload = Partial<CreateAnnouncementPayload>
export type AnnouncementResponse = ApiResponse<ApiAnnouncement>
export type AnnouncementsResponse = ApiResponse<ApiAnnouncement[]>

export async function createAnnouncement(
  courseId: UUID,
  payload: CreateAnnouncementPayload,
) {
  const { data } = await api.post<AnnouncementResponse>(
    `/courses/${courseId}/announcements`,
    payload,
  )
  return data
}

export async function listInstructorAnnouncements(courseId: UUID) {
  const { data } = await api.get<AnnouncementsResponse>(
    `/my-courses/${courseId}/announcements`,
  )
  return data
}

export async function listStudentAnnouncements(courseId: UUID) {
  const { data } = await api.get<AnnouncementsResponse>(
    `/courses/${courseId}/announcements`,
  )
  return data
}

export async function updateAnnouncement(
  announcementId: UUID,
  payload: UpdateAnnouncementPayload,
) {
  const { data } = await api.patch<AnnouncementResponse>(
    `/announcements/${announcementId}`,
    payload,
  )
  return data
}

export async function deleteAnnouncement(announcementId: UUID): Promise<void> {
  await api.delete(`/announcements/${announcementId}`)
}
