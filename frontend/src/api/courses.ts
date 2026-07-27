import { api } from '@/lib/api'
import type { ApiMessageResponse, ApiResponse, Pagination, UUID } from '@/api/types'

export interface ApiCourseInstructor {
  id: string
  full_name: string
  email: string
}

export interface ApiCourse {
  id: string
  instructor: ApiCourseInstructor
  title: string
  slug: string
  description: string
  published: boolean
  created_at: string
  updated_at: string
}

export interface CoursePagination extends Pagination {}

export interface CoursesListData {
  courses: ApiCourse[]
  pagination: CoursePagination
}

export interface CoursesListResponse {
  status: string
  data: CoursesListData
}

export interface ListCoursesParams {
  page?: number
  page_size?: number
  title?: string
  published?: boolean
}

export async function listCourses(params: ListCoursesParams = {}) {
  const { data } = await api.get<CoursesListResponse>('/courses', {
    params: {
      page: params.page ?? 1,
      page_size: params.page_size ?? 10,
      published: true,
      ...(params.title?.trim() ? { title: params.title.trim() } : {}),
      ...(params.published !== undefined ? { published: params.published } : {}),
    },
  })
  return data
}

export async function listEnrolledCourses(params: ListCoursesParams = {}) {
  const { data } = await api.get<CoursesListResponse>('/enrolled-courses', {
    params: {
      page: params.page ?? 1,
      page_size: params.page_size ?? 12,
      ...(params.title?.trim() ? { title: params.title.trim() } : {}),
    },
  })
  return data
}

export async function listMyCourses(params: ListCoursesParams = {}) {
  const { data } = await api.get<CoursesListResponse>('/my-courses', {
    params: {
      page: params.page ?? 1,
      page_size: params.page_size ?? 12,
      ...(params.title?.trim() ? { title: params.title.trim() } : {}),
    },
  })
  return data
}

export async function getCourseById(courseId: string) {
  const { data } = await api.get<ApiCourse>(`/courses/${courseId}`)
  return data
}

export interface CoursePayload {
  title: string
  description: string
  published: boolean
}

export type CreateCoursePayload = CoursePayload
export type UpdateCoursePayload = Partial<CoursePayload>
export type CourseResponse = ApiResponse<ApiCourse>

export async function createCourse(payload: CreateCoursePayload) {
  const { data } = await api.post<CourseResponse>('/courses', payload)
  return data
}

export async function updateCourse(courseId: UUID, payload: UpdateCoursePayload) {
  const { data } = await api.patch<CourseResponse>(`/courses/${courseId}`, payload)
  return data
}

export async function deleteCourse(courseId: UUID): Promise<void> {
  await api.delete(`/courses/${courseId}`)
}

export async function enrollInCourse(courseId: UUID) {
  const { data } = await api.post<ApiMessageResponse>(
    `/courses/${courseId}/enrollments/me`,
  )
  return data
}

export async function unenrollFromCourse(courseId: UUID): Promise<void> {
  await api.delete(`/courses/${courseId}/enrollments/me`)
}

export interface ApiAssignment {
  id: string
  course_id: string
  title: string
  description: string
  due_date: string
  status: string
  created_by: string
  created_at: string
  updated_at: string
}

export interface CourseAssignmentsResponse {
  status: string
  data: ApiAssignment[]
}

export async function listCourseAssignments(courseId: string) {
  const { data } = await api.get<CourseAssignmentsResponse>(
    `/courses/${courseId}/assignments`,
  )
  return data
}

export interface MyCourseAssignmentsData {
  assignments: ApiAssignment[]
  pagination: CoursePagination
}

export interface MyCourseAssignmentsResponse {
  status: string
  data: MyCourseAssignmentsData
}

export interface ListMyCourseAssignmentsParams {
  page?: number
  page_size?: number
  title?: string
}

export async function listMyCourseAssignments(
  courseId: string,
  params: ListMyCourseAssignmentsParams = {},
) {
  const { data } = await api.get<MyCourseAssignmentsResponse>(
    `/my-courses/${courseId}/assignments`,
    {
      params: {
        page: params.page ?? 1,
        page_size: params.page_size ?? 10,
        ...(params.title?.trim() ? { title: params.title.trim() } : {}),
      },
    },
  )
  return data
}

export interface ApiEnrolledStudent {
  id: string
  email: string
  full_name: string
  enrolled_at: string
}

export interface CourseStudentsData {
  students: ApiEnrolledStudent[]
  pagination: CoursePagination
}

export interface CourseStudentsResponse {
  status: string
  data: CourseStudentsData
}

export interface ListCourseStudentsParams {
  page?: number
  page_size?: number
  name?: string
}

export async function listCourseStudents(
  courseId: string,
  params: ListCourseStudentsParams = {},
) {
  const { data } = await api.get<CourseStudentsResponse>(
    `/courses/${courseId}/students`,
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
