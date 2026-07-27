import { api } from '@/lib/api'
import type { ApiResponse, UUID } from '@/api/types'

export interface CourseAnalytics {
  course_id: UUID
  enrollments: number
  assignments: number
  submissions: number
  average_grade: number
  lessons: number
  lesson_completions: number
}

export type CourseAnalyticsResponse = ApiResponse<CourseAnalytics>

export async function getCourseAnalytics(courseId: UUID) {
  const { data } = await api.get<CourseAnalyticsResponse>(
    `/courses/${courseId}/analytics`,
  )
  return data
}
