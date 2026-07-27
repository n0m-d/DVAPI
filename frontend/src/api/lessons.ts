import { api } from '@/lib/api'
import type { ApiMessageResponse, ApiResponse, UUID } from '@/api/types'

export interface ApiLesson {
  id: UUID
  course_id: UUID
  title: string
  sort_order: number
  content: string
  created_at: string
  updated_at: string
}

export type LessonsResponse = ApiResponse<ApiLesson[]>
export type LessonResponse = ApiResponse<ApiLesson>

export interface CreateLessonPayload {
  title: string
  sort_order: number
  content: string
}

export type UpdateLessonPayload = Partial<CreateLessonPayload>

export async function listStudentLessons(courseId: UUID) {
  const { data } = await api.get<LessonsResponse>(
    `/courses/${courseId}/lessons`,
  )
  return data
}

export async function listInstructorLessons(courseId: UUID) {
  const { data } = await api.get<LessonsResponse>(
    `/my-courses/${courseId}/lessons`,
  )
  return data
}

export async function createLesson(courseId: UUID, payload: CreateLessonPayload) {
  const { data } = await api.post<LessonResponse>(
    `/courses/${courseId}/lessons`,
    payload,
  )
  return data
}

export async function updateLesson(lessonId: UUID, payload: UpdateLessonPayload) {
  const { data } = await api.patch<LessonResponse>(
    `/lessons/${lessonId}`,
    payload,
  )
  return data
}

export async function deleteLesson(lessonId: UUID): Promise<void> {
  await api.delete(`/lessons/${lessonId}`)
}

export interface UpdateLessonProgressPayload {
  completed: boolean
}

export async function updateLessonProgress(
  lessonId: UUID,
  payload: UpdateLessonProgressPayload,
) {
  const { data } = await api.put<ApiMessageResponse>(
    `/lessons/${lessonId}/progress`,
    payload,
  )
  return data
}

export interface CourseProgress {
  course_id: UUID
  total_lessons: number
  completed_lessons: number
  percentage: number
}

export type CourseProgressResponse = ApiResponse<CourseProgress>

export async function getCourseProgress(courseId: UUID) {
  const { data } = await api.get<CourseProgressResponse>(
    `/courses/${courseId}/progress`,
  )
  return data
}

export async function getContinueLesson(courseId: UUID) {
  const { data } = await api.get<LessonResponse>(
    `/courses/${courseId}/continue`,
  )
  return data
}
