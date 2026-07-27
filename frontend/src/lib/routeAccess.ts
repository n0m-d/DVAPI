import type { RouteLocationNormalized, RouteRecordNormalized } from 'vue-router'
import type { UserRole } from '@/types/roles'

export function requiredRolesForRoute(
  route: Pick<RouteLocationNormalized, 'matched' | 'meta'>,
): UserRole[] | undefined {
  for (let i = route.matched.length - 1; i >= 0; i--) {
    const roles = route.matched[i]?.meta.roles
    if (roles?.length) return roles
  }

  const top = route.meta.roles
  return top?.length ? top : undefined
}

export function routeRequiresAuth(
  route: Pick<RouteLocationNormalized, 'matched' | 'meta'>,
): boolean {
  return route.matched.some(
    (record: RouteRecordNormalized) => Boolean(record.meta.roles?.length || record.meta.requiresAuth),
  )
}

export function canRoleAccessRoute(
  role: UserRole,
  route: Pick<RouteLocationNormalized, 'matched' | 'meta'>,
): boolean {
  const required = requiredRolesForRoute(route)
  // No role list means any authenticated user may access (auth checked separately).
  if (!required?.length) return true
  return required.includes(role)
}
