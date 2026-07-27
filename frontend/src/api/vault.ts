import { api } from '@/lib/api'
import type { ApiResponse } from '@/api/types'

export type FounderNotesResponse = ApiResponse<string>

export async function getFounderNote() {
  const apiBase = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v2'
  const v1Base = apiBase.replace(/\/v2(\/?|$)/, '/v1')

  const { data } = await api.get<FounderNotesResponse>(
    '/admin/founder-notes',
    { baseURL: v1Base }
  )

  return data
}