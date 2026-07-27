import { api } from "@/lib/api";

export interface LibraryBook {
  id: string
  title: string
  author: string
  year?: string
  genre?: string
  bookmark?: string
}

interface ApiResponse<T> {
  data: T
  url: string
}

interface BooksResponseData {
  books: LibraryBook[]
  found: boolean
}

type BooksApiResponse = ApiResponse<BooksResponseData>

export async function getLibraryList(title: string) {
  // Deliberately open SSRF proxy target
  const apiBase = (import.meta.env.VITE_LIBRARY_BASE_URL || 'http://localhost:5000/api').replace(/\/$/, '')
  const target = title
    ? `${apiBase}/books?title=${encodeURIComponent(title)}`
    : `${apiBase}/books`
  const { data } = await api.get<BooksApiResponse>('/library', {
    params: { url: target },
  })
  return data
}
