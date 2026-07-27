import { api } from '@/lib/api'
import type { ApiAssignment, CoursePagination } from '@/api/courses'
import type { MessageResponse } from '@/api/auth'
import type {
  ApiDataMessageResponse,
  ApiMessageResponse,
  ApiResponse,
  UUID,
} from '@/api/types'

export interface AssignmentResponse {
  status: string
  data: ApiAssignment
}

export interface ApiSubmission {
  id: string
  assignment_id: string
  student_id: string
  student_name?: string
  student_email?: string
  submission_text: string
  file_path: string
  file_name: string
  submitted_at: string
  grade: number | null
  feedback: string
  created_at: string
  updated_at: string
}

export interface SubmissionResponse {
  status: string
  data: ApiSubmission
}

export type AssignmentStatus = 'draft' | 'published' | 'closed'

export interface CreateAssignmentPayload {
  course_id: UUID
  title: string
  description: string
  due_date: string
  status: AssignmentStatus
}

export type UpdateAssignmentPayload = Partial<
  Omit<CreateAssignmentPayload, 'course_id'>
>

export async function createAssignment(payload: CreateAssignmentPayload) {
  const { data } = await api.post<ApiMessageResponse>('/assignments', payload)
  return data
}

export async function updateAssignment(
  assignmentId: UUID,
  payload: UpdateAssignmentPayload,
) {
  const { data } = await api.patch<AssignmentResponse>(
    `/assignments/${assignmentId}`,
    payload,
  )
  return data
}

export async function deleteAssignment(assignmentId: UUID): Promise<void> {
  await api.delete(`/assignments/${assignmentId}`)
}

export async function getAssignmentById(assignmentId: string) {
  const { data } = await api.get<AssignmentResponse>(`/assignments/${assignmentId}`)
  return data
}

export async function getMySubmission(assignmentId: string) {
  const { data } = await api.get<SubmissionResponse>(
    `/assignments/${assignmentId}/submissions/me`,
  )
  return data
}

export interface AssignmentSubmissionsData {
  submissions: ApiSubmission[]
  pagination: CoursePagination
}

export interface AssignmentSubmissionsResponse {
  status: string
  data: AssignmentSubmissionsData
}

export interface ListAssignmentSubmissionsParams {
  page?: number
  page_size?: number
  name?: string
}

export async function listAssignmentSubmissions(
  assignmentId: string,
  params: ListAssignmentSubmissionsParams = {},
) {
  const { data } = await api.get<AssignmentSubmissionsResponse>(
    `/assignments/${assignmentId}/submissions`,
    {
      params: {
        page: params.page ?? 1,
        page_size: params.page_size ?? 10,
        ...(params.name?.trim() ? { name: params.name.trim() } : {}),
      },
    },
  )
  return data
}

export async function createSubmission(
  assignmentId: string,
  payload: { submission_text: string; file: File },
) {
  const form = new FormData()
  form.append('submission_text', payload.submission_text)
  form.append('file', payload.file)

  const { data } = await api.post<MessageResponse>(
    `/assignments/${assignmentId}/submissions`,
    form,
  )
  return data
}

export type SubmissionDetail = Omit<ApiSubmission, 'grade'> & {
  grade: number
}

export type SubmissionDetailResponse = ApiResponse<SubmissionDetail>

export async function getSubmission(submissionId: UUID) {
  const { data } = await api.get<SubmissionDetailResponse>(
    `/submissions/${submissionId}`,
  )
  return data
}

export interface GradeSubmissionPayload {
  grade: number
  feedback: string
}

export type GradeSubmissionResponse = ApiDataMessageResponse<SubmissionDetail>

export async function gradeSubmission(
  submissionId: UUID,
  payload: GradeSubmissionPayload,
) {
  const { data } = await api.patch<GradeSubmissionResponse>(
    `/submissions/${submissionId}/grade`,
    payload,
  )
  return data
}

export interface ResubmitAssignmentPayload {
  submission_text: string
  file: File
}

export async function resubmitAssignment(
  submissionId: UUID,
  payload: ResubmitAssignmentPayload,
) {
  const form = new FormData()
  form.append('submission_text', payload.submission_text)
  form.append('file', payload.file)

  const { data } = await api.put<SubmissionDetailResponse>(
    `/submissions/${submissionId}`,
    form,
  )
  return data
}

export interface ApiSubmissionVersion {
  id: UUID
  submission_id: UUID
  version: number
  submission_text: string
  file_path: string
  file_name: string
  submitted_at: string
  created_at: string
}

export type SubmissionVersionsResponse = ApiResponse<ApiSubmissionVersion[]>

export async function listSubmissionVersions(submissionId: UUID) {
  const { data } = await api.get<SubmissionVersionsResponse>(
    `/submissions/${submissionId}/versions`,
  )
  return data
}
