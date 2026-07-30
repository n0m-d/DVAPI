import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  deleteNotification,
  listNotifications,
  markAllNotificationsRead,
  markNotificationRead,
  type ApiNotification,
  type ApiNotificationType,
  type ListNotificationsParams,
} from '@/api/notifications'
import type { Pagination } from '@/api/types'
import { getAccessToken } from '@/lib/api'

export type NotificationType = ApiNotificationType

export interface AppNotification {
  id: string
  title: string
  body: string
  read: boolean
  createdAt: string
  type: NotificationType
}

const STORAGE_KEY = 'Schole-notifications'
const DEFAULT_PAGE_SIZE = 20

function apiBaseUrl() {
  return import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '') || 'http://localhost:8081/api/v2'
}

function loadStored(): AppNotification[] {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw) as AppNotification[]
    return Array.isArray(parsed) ? parsed : []
  }
  catch {
    return []
  }
}

function persist(items: AppNotification[]) {
  sessionStorage.setItem(STORAGE_KEY, JSON.stringify(items.slice(0, 100)))
}

const NOTIFICATION_TYPES: NotificationType[] = ['grade', 'assignment', 'announcement', 'system', 'enrollment']

function mapApiNotification(raw: ApiNotification): AppNotification | null {
  if (!NOTIFICATION_TYPES.includes(raw.type)) {
    return null
  }

  return {
    id: raw.id,
    title: raw.title,
    body: raw.body,
    type: raw.type,
    read: Boolean(raw.read),
    createdAt: raw.created_at,
  }
}

function mapEvent(raw: Record<string, unknown>): AppNotification | null {
  if (
    typeof raw.id !== 'string'
    || typeof raw.title !== 'string'
    || typeof raw.body !== 'string'
    || typeof raw.type !== 'string'
  ) {
    return null
  }

  const type = raw.type as NotificationType
  if (!NOTIFICATION_TYPES.includes(type)) {
    return null
  }

  const createdAt = typeof raw.created_at === 'string'
    ? raw.created_at
    : new Date().toISOString()

  return {
    id: raw.id,
    title: raw.title,
    body: raw.body,
    type,
    read: typeof raw.read === 'boolean' ? raw.read : false,
    createdAt,
  }
}

export const useNotificationStore = defineStore('notifications', () => {
  const items = ref<AppNotification[]>(loadStored())
  const serverUnreadCount = ref(items.value.filter(n => !n.read).length)
  const pagination = ref<Pagination | null>(null)
  const loading = ref(false)
  const connected = ref(false)
  let source: EventSource | null = null
  let inboxRequestId = 0

  const unreadCount = computed(() => serverUnreadCount.value)

  function upsert(notification: AppNotification) {
    if (items.value.some(n => n.id === notification.id)) return
    items.value = [notification, ...items.value].slice(0, 100)
    if (!notification.read) {
      serverUnreadCount.value += 1
    }
    persist(items.value)
    toast(notification.title, { description: notification.body })
  }

  async function fetchInbox(params: ListNotificationsParams = {}) {
    const requestId = ++inboxRequestId
    loading.value = true
    try {
      const response = await listNotifications({
        page: params.page ?? 1,
        page_size: params.page_size ?? DEFAULT_PAGE_SIZE,
        unread: params.unread,
      })
      if (requestId !== inboxRequestId) return

      const mapped = response.data.notifications
        .map(mapApiNotification)
        .filter((n): n is AppNotification => n !== null)

      items.value = mapped
      serverUnreadCount.value = response.data.unread_count
      pagination.value = response.data.pagination
      persist(items.value)
    }
    finally {
      if (requestId === inboxRequestId) {
        loading.value = false
      }
    }
  }

  async function markAllRead() {
    const previous = items.value
    const previousUnread = serverUnreadCount.value

    items.value = items.value.map(n => ({ ...n, read: true }))
    serverUnreadCount.value = 0
    persist(items.value)

    try {
      await markAllNotificationsRead()
    }
    catch {
      items.value = previous
      serverUnreadCount.value = previousUnread
      persist(items.value)
      throw new Error('Failed to mark all notifications as read')
    }
  }

  async function markRead(id: string) {
    const current = items.value.find(n => n.id === id)
    if (!current || current.read) return

    const previous = items.value
    const previousUnread = serverUnreadCount.value

    items.value = items.value.map(n => n.id === id ? { ...n, read: true } : n)
    serverUnreadCount.value = Math.max(0, serverUnreadCount.value - 1)
    persist(items.value)

    try {
      await markNotificationRead(id)
    }
    catch {
      items.value = previous
      serverUnreadCount.value = previousUnread
      persist(items.value)
      throw new Error('Failed to mark notification as read')
    }
  }

  async function remove(id: string) {
    const current = items.value.find(n => n.id === id)
    if (!current) return

    const previous = items.value
    const previousUnread = serverUnreadCount.value
    const previousPagination = pagination.value

    items.value = items.value.filter(n => n.id !== id)
    if (!current.read) {
      serverUnreadCount.value = Math.max(0, serverUnreadCount.value - 1)
    }
    if (pagination.value) {
      const total = Math.max(0, pagination.value.total - 1)
      const pageSize = pagination.value.page_size
      pagination.value = {
        ...pagination.value,
        total,
        total_pages: pageSize > 0 ? Math.max(1, Math.ceil(total / pageSize)) : 0,
      }
    }
    persist(items.value)

    try {
      await deleteNotification(id)
    }
    catch {
      items.value = previous
      serverUnreadCount.value = previousUnread
      pagination.value = previousPagination
      persist(items.value)
      throw new Error('Failed to delete notification')
    }
  }

  function clear() {
    items.value = []
    serverUnreadCount.value = 0
    pagination.value = null
    persist(items.value)
  }

  function disconnect() {
    source?.close()
    source = null
    connected.value = false
  }

  function connect() {
    const token = getAccessToken()
    if (!token) {
      disconnect()
      return
    }
    if (source && source.readyState !== EventSource.CLOSED) {
      return
    }

    disconnect()
    void fetchInbox({ page: 1, page_size: DEFAULT_PAGE_SIZE })

    const url = `${apiBaseUrl()}/notifications/stream?token=${encodeURIComponent(token)}`
    source = new EventSource(url)

    source.addEventListener('connected', () => {
      connected.value = true
    })

    source.addEventListener('notification', (event) => {
      try {
        const payload = JSON.parse((event as MessageEvent).data) as Record<string, unknown>
        const notification = mapEvent(payload)
        if (notification) upsert(notification)
      }
      catch {
        // ignore malformed events
      }
    })

    source.onerror = () => {
      connected.value = false
      // Browser will retry EventSource automatically when the connection drops.
    }
  }

  return {
    items,
    connected,
    loading,
    pagination,
    unreadCount,
    connect,
    disconnect,
    fetchInbox,
    markAllRead,
    markRead,
    remove,
    clear,
  }
})
