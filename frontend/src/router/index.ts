import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import LoginPage from '@/views/LoginPage.vue'
import { useAuthStore } from '@/stores/auth'
import { startRouteLoading, stopRouteLoading } from '@/composables/useRouteLoading'
import {
  canRoleAccessRoute,
  requiredRolesForRoute,
  routeRequiresAuth,
} from '@/lib/routeAccess'
import type { UserRole } from '@/types/roles'

const allRoles: UserRole[] = ['student', 'instructor', 'admin']
const studentOnly: UserRole[] = ['student']
const instructorOnly: UserRole[] = ['instructor']
const adminOnly: UserRole[] = ['admin']

const appChildren: RouteRecordRaw[] = [
  // ── Student ──────────────────────────────────────────────
  {
    path: 'dashboard',
    name: 'student-dashboard',
    component: () => import('@/views/student/DashboardPage.vue'),
    meta: { roles: studentOnly },
  },
  {
    path: 'courses',
    name: 'student-courses',
    component: () => import('@/views/student/CourseCatalogPage.vue'),
    meta: { roles: studentOnly },
  },
  {
    path: 'courses/:id',
    name: 'student-course-detail',
    component: () => import('@/views/student/CourseDetailPage.vue'),
    meta: { roles: studentOnly },
  },
  {
    path: 'courses/:id/lessons/:lessonId',
    name: 'student-lesson-detail',
    component: () => import('@/views/student/LessonDetailPage.vue'),
    meta: { roles: studentOnly },
  },
  {
    path: 'courses/:id/assignments/:assignmentId',
    name: 'student-assignment',
    component: () => import('@/views/student/AssignmentDetailPage.vue'),
    meta: { roles: studentOnly },
  },
  {
    path: 'submissions',
    name: 'student-submissions',
    component: () => import('@/views/student/SubmissionsPage.vue'),
    meta: { roles: studentOnly },
  },
  {
    path: 'grades',
    name: 'student-grades',
    component: () => import('@/views/student/GradesPage.vue'),
    meta: { roles: studentOnly },
  },
  {
    path: 'notes',
    name: 'student-notes',
    component: () => import('@/views/student/NotesPage.vue'),
    meta: { roles: studentOnly },
  },
  {
    path: 'notifications',
    name: 'student-notifications',
    component: () => import('@/views/student/NotificationsPage.vue'),
    meta: { roles: [...studentOnly, ...instructorOnly] }, // student and instructor can access this route
  },

  // ── Instructor ───────────────────────────────────────────
  {
    path: 'instructor/dashboard',
    name: 'instructor-dashboard',
    component: () => import('@/views/instructor/DashboardPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/new',
    name: 'instructor-course-new',
    component: () => import('@/views/instructor/CourseFormPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/:id',
    name: 'instructor-course-manage',
    component: () => import('@/views/instructor/CourseManagePage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/:id/assignments/new',
    name: 'instructor-assignment-new',
    component: () => import('@/views/instructor/AssignmentFormPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/:id/assignments/:assignmentId/edit',
    name: 'instructor-assignment-edit',
    component: () => import('@/views/instructor/AssignmentFormPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/:id/assignments',
    name: 'instructor-course-assignments',
    component: () => import('@/views/instructor/CourseAssignmentsPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/:id/lessons',
    name: 'instructor-course-lessons',
    component: () => import('@/views/instructor/CourseLessonsPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/:id/assignments/:assignmentId/submissions',
    name: 'instructor-assignment-submissions',
    component: () => import('@/views/instructor/AssignmentSubmissionsPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/:id/submissions',
    name: 'instructor-course-submissions',
    component: () => import('@/views/instructor/CourseSubmissionsPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/:id/grades/import',
    name: 'instructor-grade-import',
    component: () => import('@/views/instructor/GradeImportPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/:id/grades',
    name: 'instructor-course-grades',
    component: () => import('@/views/instructor/CourseGradesPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/:id/roster',
    name: 'instructor-course-roster',
    component: () => import('@/views/instructor/CourseRosterPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/:id/announcements',
    name: 'instructor-course-announcements',
    component: () => import('@/views/instructor/CourseAnnouncementsPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/courses/:id/analytics',
    name: 'instructor-course-analytics',
    component: () => import('@/views/instructor/CourseAnalyticsPage.vue'),
    meta: { roles: instructorOnly },
  },
  {
    path: 'instructor/analytics',
    name: 'instructor-analytics',
    component: () => import('@/views/instructor/AnalyticsPage.vue'),
    meta: { roles: instructorOnly },
  },

  // ── Admin ────────────────────────────────────────────────
  {
    path: 'admin/dashboard',
    name: 'admin-dashboard',
    component: () => import('@/views/admin/DashboardPage.vue'),
    meta: { roles: adminOnly },
  },
  {
    path: 'admin/users',
    name: 'admin-users',
    component: () => import('@/views/admin/UsersPage.vue'),
    meta: { roles: adminOnly },
  },
  {
    path: 'admin/courses',
    name: 'admin-courses',
    component: () => import('@/views/admin/CoursesPage.vue'),
    meta: { roles: adminOnly },
  },
   {
    path: 'admin/library',
    name: 'admin-library',
    component: () => import('@/views/admin/Library.vue'),
    meta: { roles: adminOnly },
  },
  // {
  //   path: 'admin/enrollments',
  //   name: 'admin-enrollments',
  //   component: () => import('@/views/admin/EnrollmentsPage.vue'),
  //   meta: { roles: adminOnly },
  // },
  {
    path: 'admin/logs',
    name: 'admin-logs',
    component: () => import('@/views/admin/LogsPage.vue'),
    meta: { roles: adminOnly },
  },
  {
    path: 'admin/vault',
    name: 'admin-vault',
    component: () => import('@/views/admin/VaultPage.vue'),
    meta: { roles: adminOnly },
  },

  // ── Shared ───────────────────────────────────────────────
  {
    path: 'profile',
    name: 'profile',
    component: () => import('@/views/shared/ProfilePage.vue'),
    meta: { roles: allRoles },
  },
  {
    path: 'change-password',
    redirect: { path: '/profile', query: { tab: 'password' } },
  },

  // ── Errors (in-app) ────────────────────────────────────
  {
    path: '401',
    name: 'unauthorized',
    component: () => import('@/views/errors/ErrorPage.vue'),
    meta: { roles: allRoles, code: '401' },
  },
  {
    path: '403',
    name: 'forbidden',
    component: () => import('@/views/errors/ErrorPage.vue'),
    meta: { roles: allRoles, code: '403' },
  },
  {
    path: '404',
    name: 'not-found',
    component: () => import('@/views/errors/ErrorPage.vue'),
    meta: { roles: allRoles, code: '404' },
  },
  {
    path: '500',
    name: 'server-error',
    component: () => import('@/views/errors/ErrorPage.vue'),
    meta: { roles: allRoles, code: '500' },
  },
]

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/DashboardLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: () => {
          const auth = useAuthStore()
          if (!auth.isLoggedIn) return '/login'
          if (auth.role === 'instructor') return '/instructor/dashboard'
          if (auth.role === 'admin') return '/admin/dashboard'
          return '/dashboard'
        },
      },
      ...appChildren,
    ],
  },

  // ── Auth (public) ────────────────────────────────────────
  { path: '/login', name: 'login', component: LoginPage },
  { path: '/register', name: 'register', component: () => import('@/views/RegisterPage.vue') },
  { path: '/forgot-password', name: 'forgot-password', component: () => import('@/views/ForgotPasswordPage.vue') },
  { path: '/forgot-password/verify', name: 'forgot-password-verify', component: () => import('@/views/ForgotPasswordVerifyPage.vue') },
  { path: '/verify-email', name: 'verify-email', component: () => import('@/views/auth/VerifyEmailPage.vue') },
  { path: '/reset-password', name: 'reset-password', component: () => import('@/views/ResetPasswordPage.vue') },

  // Legacy redirects
  {
    path: '/reset-password/:token',
    redirect: to => ({
      name: 'reset-password',
      query: to.query,
    }),
  },
  {
    path: '/forgot-password/reset',
    redirect: to => ({
      name: 'reset-password',
      query: to.query,
    }),
  },

  // Catch-all
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

const authRouteNames = new Set([
  'login',
  'register',
  'forgot-password',
  'forgot-password-verify',
  'verify-email',
  'reset-password',
])

router.beforeEach(async (to, _, next) => {
  startRouteLoading()

  const auth = useAuthStore()
  const isAuthRoute = authRouteNames.has(String(to.name ?? ''))
  const requiredRoles = requiredRolesForRoute(to)
  const requiresAuth = routeRequiresAuth(to)

  // Shared routes (all roles) only hydrate once; role-restricted routes revalidate
  // against /users/me so localStorage edits cannot elevate privileges.
  const isElevatedRoute = Boolean(
    requiredRoles?.length
    && !(
      requiredRoles.includes('student')
      && requiredRoles.includes('instructor')
      && requiredRoles.includes('admin')
    ),
  )

  if (auth.token && !isAuthRoute) {
    await auth.hydrateSession(!auth.sessionReady || isElevatedRoute)
  }
  else if (!auth.token) {
    auth.clearSession()
  }

  const loggedIn = auth.isLoggedIn

  if (!loggedIn && (requiresAuth || requiredRoles?.length)) {
    next({ name: 'login', query: { redirect: to.fullPath } })
    return
  }

  if (loggedIn && isAuthRoute) {
    next(auth.homeRouteForRole(auth.role))
    return
  }

  if (loggedIn && requiredRoles?.length && !requiredRoles.includes(auth.role)) {
    next({ name: 'forbidden' })
    return
  }

  if (loggedIn && requiresAuth && !canRoleAccessRoute(auth.role, to)) {
    next({ name: 'forbidden' })
    return
  }

  next()
})

router.afterEach(() => {
  stopRouteLoading()
})

router.onError(() => {
  stopRouteLoading()
})

export default router

declare module 'vue-router' {
  interface RouteMeta {
    roles?: UserRole[]
    requiresAuth?: boolean
    code?: '401' | '403' | '404' | '500'
  }
}
