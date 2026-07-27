import type { UserRole } from '@/types/roles'
import type { LucideIcon } from 'lucide-vue-next'
import {
  BarChart3,
  Bell,
  BookOpen,
  BookSearch,
  ClipboardList,
  GraduationCap,
  LayoutDashboard,
  Megaphone,
  ScrollText,
  Shield,
  StickyNote,
  UserCog,
  Users,
} from 'lucide-vue-next'

export interface NavItem {
  title: string
  url: string
  icon: LucideIcon
}

export const studentNav: NavItem[] = [
  { title: 'Dashboard', url: '/dashboard', icon: LayoutDashboard },
  { title: 'Courses', url: '/courses', icon: BookOpen },
  { title: 'Submissions', url: '/submissions', icon: ClipboardList },
  { title: 'Grades', url: '/grades', icon: GraduationCap },
  { title: 'Notes', url: '/notes', icon: StickyNote },
  // { title: 'Transcript', url: '/transcript', icon: ScrollText }, //Future feature; May be added later
  { title: 'Notifications', url: '/notifications', icon: Bell },
]

export const instructorNav: NavItem[] = [
  { title: 'Dashboard', url: '/instructor/dashboard', icon: LayoutDashboard },
  { title: 'Analytics', url: '/instructor/analytics', icon: BarChart3 },
  { title: 'Notifications', url: '/notifications', icon: Bell },
]

export const adminNav: NavItem[] = [
  { title: 'Dashboard', url: '/admin/dashboard', icon: LayoutDashboard },
  { title: 'Users', url: '/admin/users', icon: Users },
  { title: 'Courses', url: '/admin/courses', icon: BookOpen },
  { title: 'Library', url: '/admin/library', icon: BookSearch },
  // { title: 'Enrollments', url: '/admin/enrollments', icon: UserCog },
  { title: 'Logs', url: '/admin/logs', icon: ScrollText },
  { title: 'Vault', url: '/admin/vault', icon: Shield },
]

export const sharedNav: NavItem[] = [
  { title: 'Account', url: '/profile', icon: Users },
]

export function navForRole(role: UserRole): NavItem[] {
  if (role === 'instructor') return [...instructorNav, ...sharedNav]
  if (role === 'admin') return [...adminNav, ...sharedNav]
  return [...studentNav, ...sharedNav]
}

export function instructorCourseNav(courseId: string): NavItem[] {
  const base = `/instructor/courses/${courseId}`
  return [
    { title: 'Overview', url: base, icon: BookOpen },
    { title: 'Roster', url: `${base}/roster`, icon: Users },
    { title: 'Lessons', url: `${base}/lessons`, icon: BookOpen },
    { title: 'Assignments', url: `${base}/assignments`, icon: ClipboardList },
    // { title: 'Grades', url: `${base}/grades`, icon: GraduationCap },
    { title: 'Announcements', url: `${base}/announcements`, icon: Megaphone },
    { title: 'Analytics', url: `${base}/analytics`, icon: BarChart3 },
    // { title: 'Import grades', url: `${base}/grades/import`, icon: Upload },
  ]
}

export const roleLabels: Record<UserRole, string> = {
  student: 'Student',
  instructor: 'Instructor',
  admin: 'Admin',
}

export const roleBadgeIcon: Record<UserRole, LucideIcon> = {
  student: GraduationCap,
  instructor: BookOpen,
  admin: Shield,
}
