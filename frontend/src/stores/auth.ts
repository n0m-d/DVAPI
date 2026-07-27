import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import * as authApi from '@/api/auth'
import type { ApiUser, LoginPayload, RegisterPayload, UpdatePasswordPayload } from '@/api/auth'
import { getCurrentUser } from '@/api/users'
import { ApiError, getAccessToken, setAccessToken } from '@/lib/api'
import type { AuthUser, UserRole } from '@/types/roles'

const USER_KEY = 'auth-user'
const LEGACY_ROLE_KEY = 'user-role'

const VALID_ROLES: UserRole[] = ['student', 'instructor', 'admin']

function isValidRole(value: unknown): value is UserRole {
  return typeof value === 'string' && VALID_ROLES.includes(value as UserRole)
}

function normalizeUser(raw: unknown): ApiUser | null {
  if (!raw || typeof raw !== 'object') return null

  const user = raw as Partial<ApiUser>
  if (
    typeof user.id !== 'string'
    || typeof user.email !== 'string'
    || typeof user.full_name !== 'string'
    || !isValidRole(user.role)
  ) {
    return null
  }

  return {
    id: user.id,
    email: user.email,
    full_name: user.full_name,
    role: user.role,
    created_at: typeof user.created_at === 'string' ? user.created_at : '',
    updated_at: typeof user.updated_at === 'string' ? user.updated_at : '',
  }
}

function readStoredUser(): ApiUser | null {
  // Drop legacy bypass key that could disagree with the real user role.
  localStorage.removeItem(LEGACY_ROLE_KEY)

  const raw = localStorage.getItem(USER_KEY)
  if (!raw || raw === 'null') return null

  try {
    return normalizeUser(JSON.parse(raw))
  }
  catch {
    return null
  }
}

function mapApiUser(apiUser: ApiUser): AuthUser {
  return {
    id: apiUser.id,
    name: apiUser.full_name,
    email: apiUser.email,
    role: apiUser.role,
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getAccessToken())
  const apiUser = ref<ApiUser | null>(readStoredUser())
  const sessionReady = ref(false)
  const hydrating = ref(false)

  // Incomplete session (token without user or vice versa) is treated as logged out.
  if (!token.value || !apiUser.value) {
    if (token.value) setAccessToken(null)
    token.value = null
    apiUser.value = null
    localStorage.removeItem(USER_KEY)
  }

  watch(token, (value) => {
    setAccessToken(value)
  })

  watch(apiUser, (value) => {
    if (value) {
      localStorage.setItem(USER_KEY, JSON.stringify(value))
    }
    else {
      localStorage.removeItem(USER_KEY)
    }
  }, { deep: true })

  /** Always derived from the authenticated user — never a separate editable role. */
  const role = computed<UserRole>(() => apiUser.value?.role ?? 'student')

  const user = computed<AuthUser>(() => {
    if (apiUser.value) {
      return mapApiUser(apiUser.value)
    }

    return {
      id: '',
      name: 'Guest',
      email: '',
      role: 'student',
    }
  })

  const isLoggedIn = computed(() => Boolean(token.value && apiUser.value && isValidRole(apiUser.value.role)))

  function syncUser(nextUser: ApiUser) {
    const normalized = normalizeUser(nextUser)
    if (!normalized) {
      clearSession()
      return
    }
    apiUser.value = normalized
  }

  function homeRouteForRole(r: UserRole = role.value) {
    if (r === 'instructor') return '/instructor/dashboard'
    if (r === 'admin') return '/admin/dashboard'
    return '/dashboard'
  }

  function persistSession(accessToken: string, nextUser: ApiUser) {
    const normalized = normalizeUser(nextUser)
    if (!normalized) {
      clearSession()
      throw new ApiError('Invalid user payload from server', 0)
    }
    token.value = accessToken
    apiUser.value = normalized
  }

  function clearSession() {
    token.value = null
    apiUser.value = null
    sessionReady.value = true
    localStorage.removeItem(USER_KEY)
    localStorage.removeItem(LEGACY_ROLE_KEY)
  }

  async function hydrateSession(force = false) {
    if (!token.value) {
      clearSession()
      sessionReady.value = true
      return false
    }

    if (sessionReady.value && !force) {
      return isLoggedIn.value
    }

    if (hydrating.value) {
      return isLoggedIn.value
    }

    hydrating.value = true
    try {
      const response = await getCurrentUser()
      syncUser(response.data)
      return isLoggedIn.value
    }
    catch {
      clearSession()
      return false
    }
    finally {
      hydrating.value = false
      sessionReady.value = true
    }
  }

  function hasRole(allowed: UserRole | UserRole[]) {
    if (!isLoggedIn.value) return false
    const list = Array.isArray(allowed) ? allowed : [allowed]
    return list.includes(role.value)
  }

  async function login(payload: LoginPayload) {
    const response = await authApi.login(payload)
    persistSession(response.access_token, response.user)
    sessionReady.value = true
    return response
  }

  async function register(payload: RegisterPayload) {
    return authApi.register(payload)
  }
  async function updatePassword(payload: UpdatePasswordPayload) {
    return authApi.updatePassword(payload)
  }

  async function logout() {
    try {
      if (token.value) {
        await authApi.logout()
      }
    }
    catch (error) {
      if (!(error instanceof ApiError && (error.status === 401 || error.status === 403))) {
        throw error
      }
    }
    finally {
      clearSession()
      sessionReady.value = true
    }
  }

  return {
    token,
    apiUser,
    role,
    user,
    isLoggedIn,
    sessionReady,
    hydrating,
    syncUser,
    hydrateSession,
    hasRole,
    homeRouteForRole,
    login,
    register,
    logout,
    clearSession,
    updatePassword,
  }
})
