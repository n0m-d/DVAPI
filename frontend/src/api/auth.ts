import { api } from '@/lib/api'
import type { UserRole } from '@/types/roles'

export interface ApiUser {
  id: string
  email: string
  full_name: string
  role: UserRole
  created_at: string
  updated_at: string
}

export interface AuthResponse {
  access_token: string
  token_type: string
  expires_in: number
  user: ApiUser
}

export interface RegisterPayload {
  email: string
  full_name: string
  password: string
  confirm_password: string
}

export interface LoginPayload {
  email: string
  password: string
}

export interface MessageResponse {
  message: string
  status: string
}

export interface UpdatePasswordPayload {
  current_password: string
  password: string
  confirm_password: string
}

export async function updatePassword(payload: UpdatePasswordPayload) {
  const { data } = await api.post<MessageResponse>('/auth/update-password', payload)
  return data
}


export async function register(payload: RegisterPayload) {
  const { data } = await api.post<MessageResponse>('/auth/register', payload)
  return data
}

export async function login(payload: LoginPayload) {
  const { data } = await api.post<AuthResponse>('/auth/login', payload)
  return data
}

export async function logout() {
  const { data } = await api.post<MessageResponse>('/auth/logout')
  return data
}

export interface PasswordResetRequestPayload {
  email: string
}

export interface PasswordResetRequestResponse extends MessageResponse {
  otp?: string
}

export interface PasswordResetVerifyPayload {
  email: string
  otp: string
}

export interface PasswordResetConfirmPayload {
  email: string
  password: string
  confirm_password: string
}

export async function requestPasswordReset(payload: PasswordResetRequestPayload) {
  const { data } = await api.post<PasswordResetRequestResponse>(
    '/auth/password-reset/request',
    payload,
  )
  return data
}

export async function verifyPasswordResetOtp(payload: PasswordResetVerifyPayload) {
  const { data } = await api.post<MessageResponse>(
    '/auth/password-reset/verify',
    payload,
  )
  return data
}

export async function confirmPasswordReset(payload: PasswordResetConfirmPayload) {
  const { data } = await api.post<MessageResponse>(
    '/auth/password-reset/confirm',
    payload,
  )
  return data
}
