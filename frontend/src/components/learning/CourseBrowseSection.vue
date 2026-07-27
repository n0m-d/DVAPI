<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'
import { Search } from 'lucide-vue-next'
import { toast } from 'vue-sonner'
import { enrollInCourse, listCourses, type ApiCourse } from '@/api/courses'
import { ApiError } from '@/lib/api'
import { truncateText } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'

const props = withDefaults(defineProps<{
  title?: string
  description?: string
  pageSize?: number
  canEnroll?: boolean
  enrolledCourseIds?: string[]
}>(), {
  title: 'Browse courses',
  description: 'Search and open a course to view details.',
  pageSize: 20,
  canEnroll: false,
  enrolledCourseIds: () => [],
})

const emit = defineEmits<{
  enrolled: [courseId: string]
}>()

const search = ref('')
const courses = ref<ApiCourse[]>([])
const loading = ref(false)
const errorMessage = ref('')
const page = ref(1)
const totalPages = ref(1)
const total = ref(0)

const selected = ref<ApiCourse | null>(null)
const detailsOpen = ref(false)
const enrollingCourseId = ref('')

async function fetchCourses() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await listCourses({
      page: page.value,
      page_size: props.pageSize,
      title: search.value,
    })
    courses.value = response.data.courses
    total.value = response.data.pagination.total
    totalPages.value = response.data.pagination.total_pages
    page.value = response.data.pagination.page
  }
  catch (error) {
    courses.value = []
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load courses.'
  }
  finally {
    loading.value = false
  }
}

const debouncedSearch = useDebounceFn(() => {
  page.value = 1
  fetchCourses()
}, 300)

watch(search, () => {
  debouncedSearch()
})

function openDetails(course: ApiCourse) {
  selected.value = course
  detailsOpen.value = true
}

function onDetailsOpenChange(open: boolean) {
  detailsOpen.value = open
  if (!open) {
    // Keep content mounted briefly for exit animation, then clear.
    window.setTimeout(() => {
      if (!detailsOpen.value) selected.value = null
    }, 160)
  }
}

function goToPage(next: number) {
  if (next < 1 || next > totalPages.value || next === page.value) return
  page.value = next
  fetchCourses()
}

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleDateString()
}

function isEnrolled(courseId: string) {
  return props.enrolledCourseIds.includes(courseId)
}

async function handleEnroll(course: ApiCourse) {
  if (enrollingCourseId.value || isEnrolled(course.id)) return
  enrollingCourseId.value = course.id

  try {
    const response = await enrollInCourse(course.id)
    emit('enrolled', course.id)
    toast.success(response.message || 'Enrolled successfully.')
  }
  catch (error) {
    if (error instanceof ApiError && error.status === 409) {
      emit('enrolled', course.id)
      toast.info(error.message || 'You are already enrolled in this course.')
    }
    else {
      toast.error(error instanceof ApiError ? error.message : 'Unable to enroll in course.')
    }
  }
  finally {
    enrollingCourseId.value = ''
  }
}

onMounted(fetchCourses)
</script>

<template>
  <Card class="form-card">
    <CardHeader class="gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <CardTitle>{{ title }}</CardTitle>
        <CardDescription>
          {{ description }}
        </CardDescription>
      </div>
      <div class="relative w-full sm:max-w-xs">
        <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          v-model="search"
          type="search"
          placeholder="Search courses by title..."
          class="pl-9"
          aria-label="Search courses"
        />
      </div>
    </CardHeader>
    <CardContent class="space-y-4">
      <p
        v-if="errorMessage"
        class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
      >
        {{ errorMessage }}
      </p>

      <p
        v-else-if="loading"
        class="text-sm text-muted-foreground"
      >
        Loading courses...
      </p>

      <template v-else>
        <p class="text-sm text-muted-foreground">
          {{ total }} course{{ total === 1 ? '' : 's' }} found
        </p>

        <div
          v-if="courses.length"
          class="grid gap-3 md:grid-cols-2 xl:grid-cols-3"
        >
          <button
            v-for="course in courses"
            :key="course.id"
            type="button"
            class="rounded-xl border border-border/60 bg-card/40 p-4 text-left transition hover:border-primary/40 hover:bg-primary/5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            @click="openDetails(course)"
          >
            <p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              {{ course.slug }}
            </p>
            <h3 class="mt-1 text-base font-semibold text-foreground">
              {{ course.title }}
            </h3>
            <p class="mt-1 text-sm text-muted-foreground">
              {{ course.instructor.full_name }}
            </p>
            <p class="mt-2 line-clamp-2 text-sm text-muted-foreground">
              {{ course.description ? truncateText(course.description, 50) : 'No description provided.' }}
            </p>
          </button>
        </div>

        <p
          v-else
          class="text-sm text-muted-foreground"
        >
          No courses match your search.
        </p>

        <div
          v-if="totalPages > 1"
          class="flex items-center justify-between gap-3 pt-2"
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
      </template>
    </CardContent>
  </Card>

  <Dialog :open="detailsOpen" @update:open="onDetailsOpenChange">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>{{ selected?.title }}</DialogTitle>
        <DialogDescription>
          {{ selected?.slug }}
        </DialogDescription>
      </DialogHeader>

      <div
        v-if="selected"
        class="space-y-4 text-sm"
      >
        <div>
          <p class="form-label">Instructor</p>
          <p class="mt-1 text-foreground">
            {{ selected.instructor.full_name }}
          </p>
        </div>

        <div>
          <p class="form-label">Description</p>
          <p class="mt-1 text-muted-foreground">
            {{ selected.description ? truncateText(selected.description, 50) : 'No description provided.' }}
          </p>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
          <div>
            <p class="form-label">Status</p>
            <p class="mt-1 capitalize text-foreground">
              {{ selected.published ? 'Published' : 'Draft' }}
            </p>
          </div>
          <div>
            <p class="form-label">Created</p>
            <p class="mt-1 text-foreground">
              {{ formatDate(selected.created_at) }}
            </p>
          </div>
        </div>

        <div
          v-if="canEnroll"
          class="flex justify-end"
        >
          <Button
            v-if="!isEnrolled(selected.id)"
            type="button"
            :disabled="Boolean(enrollingCourseId)"
            @click="handleEnroll(selected)"
          >
            {{ enrollingCourseId === selected.id ? 'Enrolling...' : 'Enroll' }}
          </Button>
          <p
            v-else
            class="text-sm font-medium text-primary"
          >
            Already enrolled
          </p>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
