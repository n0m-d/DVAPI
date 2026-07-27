<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import FormField from '@/components/forms/FormField.vue'
import FormActions from '@/components/forms/FormActions.vue'
import FormTextarea from '@/components/forms/FormTextarea.vue'
import InstructorCourseTabs from '@/components/learning/InstructorCourseTabs.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import StatusBadge from '@/components/learning/StatusBadge.vue'
import { getCourseById, type ApiCourse } from '@/api/courses'
import {
  createAnnouncement,
  deleteAnnouncement,
  listInstructorAnnouncements,
  updateAnnouncement,
  type AnnouncementStatus,
  type ApiAnnouncement,
} from '@/api/announcements'
import { ApiError } from '@/lib/api'

const route = useRoute()

const course = ref<ApiCourse | null>(null)
const announcements = ref<ApiAnnouncement[]>([])
const loading = ref(true)
const pageError = ref('')
const dialogOpen = ref(false)
const editingAnnouncement = ref<ApiAnnouncement | null>(null)
const title = ref('')
const content = ref('')
const status = ref<AnnouncementStatus>('draft')
const saving = ref(false)
const publishingId = ref('')
const deletingId = ref('')
const formError = ref('')
const fieldErrors = ref<Record<string, string>>({})

function courseId() {
  return route.params.id as string
}

function sortAnnouncements(items: ApiAnnouncement[]) {
  return [...items].sort((a, b) =>
    new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime(),
  )
}

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString([], {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function loadPage() {
  loading.value = true
  pageError.value = ''

  try {
    const [courseData, announcementData] = await Promise.all([
      getCourseById(courseId()),
      listInstructorAnnouncements(courseId()),
    ])
    course.value = courseData
    announcements.value = sortAnnouncements(announcementData.data ?? [])
  }
  catch (error) {
    course.value = null
    announcements.value = []
    pageError.value = error instanceof ApiError
      ? error.message
      : 'Unable to load announcements.'
  }
  finally {
    loading.value = false
  }
}

function openCreateDialog() {
  editingAnnouncement.value = null
  title.value = ''
  content.value = ''
  status.value = 'draft'
  formError.value = ''
  fieldErrors.value = {}
  dialogOpen.value = true
}

function openEditDialog(announcement: ApiAnnouncement) {
  editingAnnouncement.value = announcement
  title.value = announcement.title
  content.value = announcement.content
  status.value = announcement.status
  formError.value = ''
  fieldErrors.value = {}
  dialogOpen.value = true
}

async function handleSave() {
  if (saving.value) return
  saving.value = true
  formError.value = ''
  fieldErrors.value = {}

  const payload = {
    title: title.value.trim(),
    content: content.value.trim(),
    status: status.value,
  }

  try {
    const response = editingAnnouncement.value
      ? await updateAnnouncement(editingAnnouncement.value.id, payload)
      : await createAnnouncement(courseId(), payload)

    if (editingAnnouncement.value) {
      announcements.value = sortAnnouncements(
        announcements.value.map(item => item.id === response.data.id ? response.data : item),
      )
      toast.success('Announcement updated successfully.')
    }
    else {
      announcements.value = sortAnnouncements([response.data, ...announcements.value])
      toast.success(
        response.data.status === 'published'
          ? 'Announcement published successfully.'
          : 'Announcement draft created.',
      )
    }
    dialogOpen.value = false
  }
  catch (error) {
    if (error instanceof ApiError) {
      formError.value = error.message
      fieldErrors.value = error.fieldErrors
    }
    else {
      formError.value = 'Unable to save announcement.'
    }
    toast.error(formError.value)
  }
  finally {
    saving.value = false
  }
}

async function handlePublish(announcement: ApiAnnouncement) {
  if (publishingId.value) return
  publishingId.value = announcement.id

  try {
    const response = await updateAnnouncement(announcement.id, { status: 'published' })
    announcements.value = sortAnnouncements(
      announcements.value.map(item => item.id === response.data.id ? response.data : item),
    )
    toast.success('Announcement published successfully.')
  }
  catch (error) {
    toast.error(error instanceof ApiError ? error.message : 'Unable to publish announcement.')
  }
  finally {
    publishingId.value = ''
  }
}

async function handleDelete(announcement: ApiAnnouncement) {
  if (deletingId.value) return
  deletingId.value = announcement.id

  try {
    await deleteAnnouncement(announcement.id)
    announcements.value = announcements.value.filter(item => item.id !== announcement.id)
    toast.success('Announcement deleted successfully.')
  }
  catch (error) {
    toast.error(error instanceof ApiError ? error.message : 'Unable to delete announcement.')
  }
  finally {
    deletingId.value = ''
  }
}

watch(() => route.params.id, loadPage)
onMounted(loadPage)
</script>

<template>
  <LearningPageShell
    eyebrow="Announcements"
    :title="course?.title ?? (loading ? 'Loading announcements...' : 'Course announcements')"
    description="Post updates and reminders for enrolled students."
  >
    <template #breadcrumb>
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem class="hidden md:block">
            <BreadcrumbLink href="/instructor/dashboard">
              Schole
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator class="hidden md:block" />
          <BreadcrumbItem>
            <BreadcrumbLink :href="`/instructor/courses/${courseId()}`">
              {{ course?.slug ?? 'Course' }}
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Announcements</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <InstructorCourseTabs />

    <p
      v-if="pageError"
      class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
    >
      {{ pageError }}
    </p>

    <Card>
      <CardHeader class="gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <CardTitle>Announcements</CardTitle>
          <CardDescription>
            {{ loading ? 'Loading announcements...' : `${announcements.length} announcement${announcements.length === 1 ? '' : 's'}` }}
          </CardDescription>
        </div>
        <Button :disabled="loading" @click="openCreateDialog">
          New announcement
        </Button>
      </CardHeader>
      <CardContent class="space-y-4">
        <p v-if="loading" class="text-sm text-muted-foreground">
          Loading announcements...
        </p>
        <template v-else>
          <div
            v-for="item in announcements"
            :key="item.id"
            class="rounded-lg border border-border/60 bg-muted/20 p-4"
          >
            <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
              <div class="min-w-0 space-y-2">
                <div class="flex flex-wrap items-center gap-2">
                  <h4 class="font-medium"><p v-html="item.title"></p></h4> <!-- Vuln: Added XSS -->
                  <StatusBadge :status="item.status" />
                </div>
                <p class="whitespace-pre-wrap text-sm text-muted-foreground">
                  {{ item.content || 'No content.' }}
                </p>
                <p class="text-xs text-muted-foreground">
                  Updated {{ formatDate(item.updated_at) }}
                </p>
              </div>
              <div class="flex shrink-0 flex-wrap gap-2">
                <Button
                  v-if="item.status === 'draft'"
                  type="button"
                  size="sm"
                  :disabled="Boolean(publishingId)"
                  @click="handlePublish(item)"
                >
                  {{ publishingId === item.id ? 'Publishing...' : 'Publish' }}
                </Button>
                <Button type="button" variant="outline" size="sm" @click="openEditDialog(item)">
                  Edit
                </Button>
                <AlertDialog>
                  <AlertDialogTrigger as-child>
                    <Button
                      type="button"
                      variant="destructive"
                      size="sm"
                      :disabled="Boolean(deletingId)"
                    >
                      Delete
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Delete this announcement?</AlertDialogTitle>
                      <AlertDialogDescription>
                        This permanently deletes “{{ item.title }}”.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel :disabled="deletingId === item.id">
                        Cancel
                      </AlertDialogCancel>
                      <AlertDialogAction
                        class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        :disabled="deletingId === item.id"
                        @click="handleDelete(item)"
                      >
                        {{ deletingId === item.id ? 'Deleting...' : 'Delete announcement' }}
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            </div>
          </div>
          <p
            v-if="announcements.length === 0"
            class="text-center text-sm text-muted-foreground"
          >
            No announcements yet. Create a draft or publish the first update.
          </p>
        </template>
      </CardContent>
    </Card>

    <Dialog v-model:open="dialogOpen">
      <DialogContent class="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {{ editingAnnouncement ? 'Edit announcement' : 'Create announcement' }}
          </DialogTitle>
          <DialogDescription>
            Save as a draft or publish immediately for enrolled students.
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-6" @submit.prevent="handleSave">
          <p
            v-if="formError"
            class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
          >
            {{ formError }}
          </p>

          <div class="space-y-5">
            <FormField label="Title" html-for="announcement-title" :error="fieldErrors.title">
              <Input
                id="announcement-title"
                v-model="title"
                required
                placeholder="Office hours update"
              />
            </FormField>

            <FormField label="Status" html-for="announcement-status" :error="fieldErrors.status">
              <Select v-model="status">
                <SelectTrigger id="announcement-status">
                  <SelectValue placeholder="Select a status" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="draft">Draft</SelectItem>
                  <SelectItem value="published">Published</SelectItem>
                </SelectContent>
              </Select>
            </FormField>
          </div>

          <FormField label="Content" html-for="announcement-content" :error="fieldErrors.content">
            <FormTextarea
              id="announcement-content"
              v-model="content"
              :rows="8"
              required
              placeholder="Write your announcement..."
            />
          </FormField>

          <FormActions>
            <Button type="button" variant="outline" :disabled="saving" @click="dialogOpen = false">
              Cancel
            </Button>
            <Button type="submit" :disabled="saving">
              {{ saving ? 'Saving...' : (editingAnnouncement ? 'Save changes' : 'Create announcement') }}
            </Button>
          </FormActions>
        </form>
      </DialogContent>
    </Dialog>
  </LearningPageShell>
</template>
