import { api, ApiError } from '@/lib/api'
import type { UUID } from '@/api/types'

export interface ApiNote {
  id: UUID
  userId: UUID
  title: string
  body: string
  createdAt: string
  updatedAt: string
}

export interface CreateNoteInput {
  userId: UUID
  title: string
  body: string
}

export interface UpdateNoteInput {
  title?: string
  body?: string
}

interface GqlResponse<T> {
  data?: T
  errors?: { message: string }[]
}

async function graphql<T>(
  query: string,
  variables?: Record<string, unknown>,
): Promise<T> {
  const { data } = await api.post<GqlResponse<T>>('/query', { query, variables })

  if (data.errors?.length) {
    throw new ApiError(data.errors[0]?.message || 'GraphQL request failed', 400, data)
  }
  if (!data.data) {
    throw new ApiError('Empty GraphQL response', 500, data)
  }

  return data.data
}

const NOTE_FIELDS = `
  id
  userId
  title
  body
  createdAt
  updatedAt
`

export async function listNotes(userId?: UUID): Promise<ApiNote[]> {
  const data = await graphql<{ notes: ApiNote[] }>(
    `query Notes($userId: ID) {
      notes(userId: $userId) { ${NOTE_FIELDS} }
    }`,
    { userId: userId ?? null },
  )
  return data.notes
}

export async function createNote(input: CreateNoteInput): Promise<ApiNote> {
  const data = await graphql<{ createNote: ApiNote }>(
    `mutation CreateNote($input: NewNote!) {
      createNote(input: $input) { ${NOTE_FIELDS} }
    }`,
    { input },
  )
  return data.createNote
}

export async function updateNote(
  id: UUID,
  input: UpdateNoteInput,
): Promise<ApiNote> {
  const data = await graphql<{ updateNote: ApiNote }>(
    `mutation UpdateNote($id: ID!, $input: UpdateNote!) {
      updateNote(id: $id, input: $input) { ${NOTE_FIELDS} }
    }`,
    { id, input },
  )
  return data.updateNote
}

export async function deleteNote(id: UUID): Promise<boolean> {
  const data = await graphql<{ deleteNote: boolean }>(
    `mutation DeleteNote($id: ID!) {
      deleteNote(id: $id)
    }`,
    { id },
  )
  return data.deleteNote
}
