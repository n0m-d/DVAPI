import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import type { UserRole } from '@/types/roles'

export { useAuthStore } from '@/stores/auth'

export function getStoredRole(): UserRole {
  return useAuthStore().role
}

export function isAuthenticated(): boolean {
  return useAuthStore().isLoggedIn
}

/** Composition helper over the Pinia auth store (keeps refs reactive when destructured). */
export function useAuth() {
  const store = useAuthStore()
  const { role, user, isLoggedIn, token, apiUser, sessionReady } = storeToRefs(store)

  return {
    role,
    user,
    isLoggedIn,
    token,
    apiUser,
    sessionReady,
    syncUser: store.syncUser,
    hydrateSession: store.hydrateSession,
    hasRole: store.hasRole,
    homeRouteForRole: store.homeRouteForRole,
    login: store.login,
    register: store.register,
    logout: store.logout,
    updatePassword: store.updatePassword,
  }
}
