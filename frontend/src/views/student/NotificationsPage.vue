<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Bell, BookOpen, Megaphone, ClipboardList, GraduationCap, Settings, Trash2 } from 'lucide-vue-next'
import type { Component } from 'vue'
import { storeToRefs } from 'pinia'
import { toast } from 'vue-sonner'
import { cn } from '@/lib/utils'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { useNotificationStore, type AppNotification } from '@/stores/notifications'

const PAGE_SIZE = 20

const store = useNotificationStore()
const { items: notifications, unreadCount, connected, loading, pagination } = storeToRefs(store)

const page = ref(1)
const empty = computed(() => !loading.value && notifications.value.length === 0)
const totalPages = computed(() => pagination.value?.total_pages ?? 1)

const typeIcons: Record<AppNotification['type'], Component> = {
  grade: GraduationCap,
  assignment: ClipboardList,
  announcement: Megaphone,
  system: Settings,
  enrollment: BookOpen,
}

function formatDate(date: string) {
  return new Date(date).toLocaleString()
}

async function loadInbox() {
  try {
    await store.fetchInbox({ page: page.value, page_size: PAGE_SIZE })
  }
  catch {
    toast.error('Failed to load notifications')
  }
}

async function goToPage(next: number) {
  page.value = next
  await loadInbox()
}

async function onMarkAllRead() {
  try {
    await store.markAllRead()
  }
  catch {
    toast.error('Failed to mark all as read')
  }
}

async function onMarkRead(id: string) {
  try {
    await store.markRead(id)
  }
  catch {
    toast.error('Failed to mark notification as read')
  }
}

async function onDelete(id: string) {
  try {
    await store.remove(id)
    if (notifications.value.length === 0 && page.value > 1) {
      page.value -= 1
      await loadInbox()
    }
  }
  catch {
    toast.error('Failed to delete notification')
  }
}

onMounted(() => {
  void loadInbox()
})

watch(pagination, (value) => {
  if (!value) return
  if (page.value > value.total_pages && value.total_pages > 0) {
    page.value = value.total_pages
  }
})
</script>

<template>
  <LearningPageShell
    eyebrow="Inbox"
    title="Notifications"
    description="Live grades, assignments, and announcements from your courses."
  >
    <template #breadcrumb>
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem class="hidden md:block">
            <BreadcrumbLink href="/">
              Schole
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator class="hidden md:block" />
          <BreadcrumbItem>
            <BreadcrumbPage>Notifications</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <Card>
      <CardHeader class="flex flex-row items-start justify-between gap-4 space-y-0">
        <div class="space-y-1.5">
          <CardTitle>All notifications</CardTitle>
          <CardDescription>
            {{ pagination?.total ?? notifications.length }} total · {{ unreadCount }} unread
            <span class="text-muted-foreground">
              · {{ connected ? 'Live' : 'Reconnecting…' }}
            </span>
          </CardDescription>
        </div>
        <Button
          v-if="unreadCount > 0"
          variant="outline"
          size="sm"
          :disabled="loading"
          @click="onMarkAllRead"
        >
          Mark all read
        </Button>
      </CardHeader>
      <CardContent>
        <p
          v-if="loading && notifications.length === 0"
          class="py-8 text-center text-sm text-muted-foreground"
        >
          Loading notifications…
        </p>
        <p
          v-else-if="empty"
          class="py-8 text-center text-sm text-muted-foreground"
        >
          No notifications yet. New grades, submissions, and announcements will appear here live.
        </p>
        <ul
          v-else
          class="divide-y divide-border"
        >
          <li
            v-for="notification in notifications"
            :key="notification.id"
            :class="cn(
              'flex cursor-pointer gap-4 py-4 first:pt-0 last:pb-0',
              !notification.read && 'rounded-lg bg-muted/40 px-3 -mx-3',
            )"
            @click="onMarkRead(notification.id)"
          >
            <div
              :class="cn(
                'flex size-9 shrink-0 items-center justify-center rounded-full',
                notification.read ? 'bg-muted text-muted-foreground' : 'bg-primary/10 text-primary',
              )"
            >
              <component
                :is="typeIcons[notification.type]"
                class="h-4 w-4"
              />
            </div>
            <div class="min-w-0 flex-1 space-y-1">
              <div class="flex items-start justify-between gap-2">
                <p
                  :class="cn(
                    'text-sm',
                    notification.read ? 'font-medium text-muted-foreground' : 'font-semibold',
                  )"
                >
                  {{ notification.title }}
                </p>
                <div class="flex shrink-0 items-center gap-1">
                  <span class="text-xs text-muted-foreground">
                    {{ formatDate(notification.createdAt) }}
                  </span>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    class="size-7 text-muted-foreground hover:text-destructive"
                    aria-label="Delete notification"
                    @click.stop="onDelete(notification.id)"
                  >
                    <Trash2 class="h-3.5 w-3.5" />
                  </Button>
                </div>
              </div>
              <p
                :class="cn(
                  'text-sm',
                  notification.read ? 'text-muted-foreground' : 'text-foreground',
                )"
              >
                {{ notification.body }}
              </p>
              <span
                v-if="!notification.read"
                class="inline-flex items-center gap-1 text-xs font-medium text-primary"
              >
                <Bell class="h-3 w-3" />
                Mark as read
              </span>
            </div>
          </li>
        </ul>

        <div
          v-if="totalPages > 1"
          class="mt-6 flex items-center justify-between gap-3"
        >
          <Button
            type="button"
            variant="outline"
            size="sm"
            :disabled="page <= 1 || loading"
            @click="goToPage(page - 1)"
          >
            Previous
          </Button>
          <p class="text-xs text-muted-foreground">
            Page {{ page }} of {{ totalPages }}
          </p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            :disabled="page >= totalPages || loading"
            @click="goToPage(page + 1)"
          >
            Next
          </Button>
        </div>
      </CardContent>
    </Card>
  </LearningPageShell>
</template>
