import { api } from '@/lib/api'
import type { ApiResponse, UUID } from '@/api/types'

export interface ApiGrade {
  submission_id: UUID
  assignment_id: UUID
  assignment_title: string
  course_id: UUID
  course_title: string
  grade: number | null
  feedback: string
  submitted_at: string
  updated_at: string
}

export interface GradesData {
  submitted: number
  average_grade: number
  grades: ApiGrade[]
}

export type GradesResponse = ApiResponse<GradesData>

export async function getOwnGrades() {
  const { data } = await api.get<GradesResponse>('/grades')
  return data
}
