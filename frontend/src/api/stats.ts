import { api } from '@/lib/api'
import type { ApiResponse } from '@/api/types'

export interface StudentStats {
  enrolled_courses: number
  completed_lessons: number
  submissions: number
  graded_submissions: number
  average_grade: number
  pending_assignments: number
}

export interface InstructorStats {
  courses: number
  published_courses: number
  enrollments: number
  students: number
  lessons: number
  assignments: number
  submissions: number
  ungraded_submissions: number
  announcements: number
}

export type StudentStatsResponse = ApiResponse<StudentStats> & {
  role: 'student'
}

export type InstructorStatsResponse = ApiResponse<InstructorStats> & {
  role: 'instructor'
}

export type DashboardStatsResponse = StudentStatsResponse | InstructorStatsResponse

export async function getDashboardStats() {
  const { data } = await api.get<DashboardStatsResponse>('/stats')
  return data
}
