import axios, { AxiosError, type AxiosInstance, type AxiosRequestConfig } from 'axios'

export class ApiError extends Error {
  status: number
  body: unknown

  constructor(message: string, status: number, body?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.body = body
  }

  get fieldErrors(): Record<string, string> {
    if (!this.body || typeof this.body !== 'object') return {}

    const errors = (this.body as Record<string, unknown>).errors
    if (!errors || typeof errors !== 'object') return {}

    const result: Record<string, string> = {}
    for (const [key, value] of Object.entries(errors as Record<string, unknown>)) {
      if (typeof value === 'string' && value) {
        result[key] = value
      }
    }
    return result
  }
}

const TOKEN_KEY = 'access_token'

export function getAccessToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setAccessToken(token: string | null) {
  if (token) {
    localStorage.setItem(TOKEN_KEY, token)
  }
  else {
    localStorage.removeItem(TOKEN_KEY)
  }
}

function apiBaseUrl() {
  return import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '') || 'http://localhost:8081/api/v2'
}

function messageFromBody(body: unknown, fallback: string): string {
  if (!body || typeof body !== 'object') return fallback

  const record = body as Record<string, unknown>

  if (typeof record.error === 'string') return record.error
  if (typeof record.message === 'string') return record.message

  if (record.errors && typeof record.errors === 'object') {
    const first = Object.values(record.errors as Record<string, string>)[0]
    if (typeof first === 'string') return first
  }

  return fallback
}

export function toApiError(error: unknown): ApiError {
  if (error instanceof ApiError) return error

  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError
    const body = axiosError.response?.data
    const status = axiosError.response?.status ?? 0

    return new ApiError(
      messageFromBody(body, axiosError.message || 'Request failed'),
      status,
      body,
    )
  }

  if (error instanceof Error) {
    return new ApiError(error.message, 0)
  }

  return new ApiError('Request failed', 0)
}

export const api: AxiosInstance = axios.create({
  baseURL: apiBaseUrl(),
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }

  // FormData must use multipart boundary from the runtime, not application/json.
  if (typeof FormData !== 'undefined' && config.data instanceof FormData) {
    const headers = config.headers
    if (headers && typeof headers.delete === 'function') {
      headers.delete('Content-Type')
    }
    else if (headers) {
      delete (headers as Record<string, unknown>)['Content-Type']
    }
  }

  return config
})

api.interceptors.response.use(
  response => response,
  async (error) => {
    const apiError = toApiError(error)

    if (apiError.status === 401) {
      const url = String(error.config?.url ?? '')
      const isAuthEndpoint = url.includes('/auth/login') || url.includes('/auth/register')
      const isInvalidCurrentPassword = url.includes('/auth/update-password')
        && apiError.message === 'Invalid current password'

      if (!isAuthEndpoint && !isInvalidCurrentPassword && typeof window !== 'undefined') {
        const { useAuthStore } = await import('@/stores/auth')
        const auth = useAuthStore()
        auth.clearSession()

        if (!window.location.pathname.startsWith('/login')) {
          const redirect = encodeURIComponent(
            `${window.location.pathname}${window.location.search}`,
          )
          window.location.assign(`/login?redirect=${redirect}`)
        }
      }
    }

    return Promise.reject(apiError)
  },
)

export async function apiRequest<T>(
  path: string,
  config: AxiosRequestConfig = {},
): Promise<T> {
  const response = await api.request<T>({
    url: path,
    ...config,
  })
  return response.data
}
