export type UUID = string

export interface ApiResponse<T> {
  status: string
  data: T
}

export interface ApiMessageResponse {
  status: string
  message: string
}

export interface ApiDataMessageResponse<T> extends ApiResponse<T> {
  message: string
}

export interface Pagination {
  total: number
  page: number
  page_size: number
  total_pages: number
}
