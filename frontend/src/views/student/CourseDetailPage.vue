<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
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
import { Calendar } from 'lucide-vue-next'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import StatusBadge from '@/components/learning/StatusBadge.vue'
import {
  enrollInCourse,
  getCourseById,
  listCourseAssignments,
  listEnrolledCourses,
  unenrollFromCourse,
  type ApiAssignment,
  type ApiCourse,
} from '@/api/courses'
import {
  getContinueLesson,
  getCourseProgress,
  listStudentLessons,
  type ApiLesson,
  type CourseProgress,
} from '@/api/lessons'
import {
  listStudentAnnouncements,
  type ApiAnnouncement,
} from '@/api/announcements'
import { ApiError } from '@/lib/api'
import { truncateText } from '@/lib/utils'

const route = useRoute()

const course = ref<ApiCourse | null>(null)
const assignments = ref<ApiAssignment[]>([])
const lessons = ref<ApiLesson[]>([])
const progress = ref<CourseProgress | null>(null)
const continueLesson = ref<ApiLesson | null>(null)
const announcements = ref<ApiAnnouncement[]>([])
const loading = ref(true)
const learningLoading = ref(false)
const announcementsLoading = ref(false)
const errorMessage = ref('')
const learningError = ref('')
const announcementsError = ref('')
const notFound = ref(false)
const enrolled = ref(false)
const enrollmentLoading = ref(false)

function courseId() {
  return route.params.id as string
}

function formatDate(date: string) {
  const parsed = new Date(date)
  if (Number.isNaN(parsed.getTime())) return date
  return parsed.toLocaleDateString()
}

function instructorName(value: ApiCourse) {
  return value.instructor.full_name?.trim() || 'Instructor'
}

async function loadAnnouncements(id: string) {
  announcementsLoading.value = true
  announcementsError.value = ''
  announcements.value = []

  try {
    const response = await listStudentAnnouncements(id)
    announcements.value = [...(response.data ?? [])]
      .filter(item => item.status === 'published')
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  }
  catch (error) {
    announcementsError.value = error instanceof ApiError
      ? error.message
      : 'Unable to load announcements.'
  }
  finally {
    announcementsLoading.value = false
  }
}

async function loadAssignments(id: string) {
  try {
    const assignmentsData = await listCourseAssignments(id)
    assignments.value = assignmentsData.data ?? []
  }
  catch (error) {
    assignments.value = []
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load assignments.'
  }
}

async function loadLearning(id: string) {
  learningLoading.value = true
  learningError.value = ''
  lessons.value = []
  progress.value = null
  continueLesson.value = null

  const [lessonsResult, progressResult, continueResult] = await Promise.allSettled([
    listStudentLessons(id),
    getCourseProgress(id),
    getContinueLesson(id),
  ])

  if (lessonsResult.status === 'fulfilled') {
    lessons.value = [...(lessonsResult.value.data ?? [])].sort((a, b) =>
      a.sort_order - b.sort_order || a.created_at.localeCompare(b.created_at),
    )
  }
  else {
    learningError.value = lessonsResult.reason instanceof ApiError
      ? lessonsResult.reason.message
      : 'Unable to load lessons.'
  }

  if (progressResult.status === 'fulfilled') {
    progress.value = progressResult.value.data
  }
  else if (!learningError.value) {
    learningError.value = progressResult.reason instanceof ApiError
      ? progressResult.reason.message
      : 'Unable to load course progress.'
  }

  if (continueResult.status === 'fulfilled') {
    continueLesson.value = continueResult.value.data
  }
  else if (
    !(continueResult.reason instanceof ApiError && continueResult.reason.status === 404)
    && !learningError.value
  ) {
    learningError.value = continueResult.reason instanceof ApiError
      ? continueResult.reason.message
      : 'Unable to load the next lesson.'
  }

  learningLoading.value = false
}

async function loadCourse() {
  loading.value = true
  errorMessage.value = ''
  notFound.value = false
  course.value = null
  assignments.value = []
  enrolled.value = false
  lessons.value = []
  progress.value = null
  continueLesson.value = null
  announcements.value = []
  announcementsError.value = ''

  const id = courseId()

  try {
    const courseData = await getCourseById(id)
    course.value = courseData
    try {
      const enrolledData = await listEnrolledCourses({
        page: 1,
        page_size: 100,
        title: courseData.title,
      })
      enrolled.value = enrolledData.data.courses.some(item => item.id === id)
    }
    catch (error) {
      errorMessage.value = error instanceof ApiError
        ? error.message
        : 'Unable to check enrollment status.'
    }

    if (enrolled.value) {
      await Promise.all([loadAssignments(id), loadLearning(id), loadAnnouncements(id)])
    }
  }
  catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound.value = true
    }
    else {
      errorMessage.value = error instanceof ApiError
        ? error.message
        : 'Unable to load course.'
      notFound.value = true
    }
  }
  finally {
    loading.value = false
  }
}

async function handleEnroll() {
  if (enrollmentLoading.value || !course.value) return
  enrollmentLoading.value = true

  try {
    const response = await enrollInCourse(course.value.id)
    enrolled.value = true
    toast.success(response.message || 'Enrolled successfully.')
    await Promise.all([
      loadAssignments(course.value.id),
      loadLearning(course.value.id),
      loadAnnouncements(course.value.id),
    ])
  }
  catch (error) {
    if (error instanceof ApiError && error.status === 409) {
      enrolled.value = true
      toast.info(error.message || 'You are already enrolled in this course.')
    }
    else {
      toast.error(error instanceof ApiError ? error.message : 'Unable to enroll in course.')
    }
  }
  finally {
    enrollmentLoading.value = false
  }
}

async function handleUnenroll() {
  if (enrollmentLoading.value || !course.value) return
  enrollmentLoading.value = true

  try {
    await unenrollFromCourse(course.value.id)
    enrolled.value = false
    assignments.value = []
    lessons.value = []
    progress.value = null
    continueLesson.value = null
    announcements.value = []
    learningError.value = ''
    announcementsError.value = ''
    toast.success('Unenrolled from course.')
  }
  catch (error) {
    toast.error(error instanceof ApiError ? error.message : 'Unable to unenroll from course.')
  }
  finally {
    enrollmentLoading.value = false
  }
}

watch(() => route.params.id, loadCourse)
onMounted(loadCourse)
</script>

<template>
  <LearningPageShell
    v-if="loading"
    eyebrow="Course"
    title="Loading course..."
    description="Fetching course details and assignments."
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
            <BreadcrumbLink href="/courses">
              Courses
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Loading</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <p class="text-sm text-muted-foreground">
      Loading course...
    </p>
  </LearningPageShell>

  <LearningPageShell
    v-else-if="course"
    :eyebrow="course.slug"
    :title="course.title"
    :description="course.description"
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
            <BreadcrumbLink href="/courses">
              Courses
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <div class="flex flex-wrap items-center gap-2">
      <StatusBadge :status="course.published ? 'active' : 'draft'" />
      <span class="text-sm text-muted-foreground">{{ instructorName(course) }}</span>
      <Button
        v-if="!enrolled"
        size="sm"
        class="ml-auto"
        :disabled="enrollmentLoading"
        @click="handleEnroll"
      >
        {{ enrollmentLoading ? 'Enrolling...' : 'Enroll' }}
      </Button>
      <AlertDialog v-else>
        <AlertDialogTrigger as-child>
          <Button
            size="sm"
            variant="outline"
            class="ml-auto"
            :disabled="enrollmentLoading"
          >
            Unenroll
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Unenroll from this course?</AlertDialogTitle>
            <AlertDialogDescription>
              You will lose access to assignments for “{{ course.title }}”.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel :disabled="enrollmentLoading">Cancel</AlertDialogCancel>
            <AlertDialogAction
              class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              :disabled="enrollmentLoading"
              @click="handleUnenroll"
            >
              {{ enrollmentLoading ? 'Unenrolling...' : 'Unenroll' }}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>

    <p
      v-if="errorMessage"
      class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
    >
      {{ errorMessage }}
    </p>

    <div class="grid">
      <Card>
        <CardHeader>
          <CardTitle>About</CardTitle>
          <CardDescription>Course description</CardDescription>
        </CardHeader>
        <CardContent>
          <p class="text-sm leading-relaxed text-muted-foreground">
            {{ course.description || 'No description published yet.' }}
          </p>
        </CardContent>
      </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>Announcements</CardTitle>
        <CardDescription>
          {{ announcementsLoading ? 'Loading announcements...' : 'Latest course updates from your instructor' }}
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <p
          v-if="announcementsError"
          class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
        >
          {{ announcementsError }}
        </p>
        <p v-else-if="announcementsLoading" class="text-sm text-muted-foreground">
          Loading announcements...
        </p>
        <div v-else-if="announcements.length" class="divide-y divide-border">
          <article
            v-for="announcement in announcements"
            :key="announcement.id"
            class="space-y-2 py-4 first:pt-0 last:pb-0"
          >
            <div class="flex flex-col gap-1 sm:flex-row sm:items-start sm:justify-between">
              <h3 class="font-medium">{{ announcement.title }}</h3>
              <time
                :datetime="announcement.created_at"
                class="shrink-0 text-xs text-muted-foreground"
              >
                {{ formatDate(announcement.created_at) }}
              </time>
            </div>
            <p class="whitespace-pre-wrap text-sm leading-relaxed text-muted-foreground">
              {{ announcement.content }}
            </p>
          </article>
        </div>
        <p v-else class="text-sm text-muted-foreground">
          No announcements have been published yet.
        </p>
      </CardContent>
    </Card>

    <template v-if="enrolled">
      <p
        v-if="learningError"
        class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
      >
        {{ learningError }}
      </p>

      <Card>
        <CardHeader class="gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <CardTitle>Course progress</CardTitle>
            <CardDescription v-if="learningLoading">
              Loading your progress...
            </CardDescription>
            <CardDescription v-else-if="progress">
              {{ progress.completed_lessons }} of {{ progress.total_lessons }} lessons complete
            </CardDescription>
            <CardDescription v-else>
              Progress is not available yet.
            </CardDescription>
          </div>
          <Button
            v-if="continueLesson"
            as-child
            class="shrink-0"
          >
            <RouterLink :to="`/courses/${course.id}/lessons/${continueLesson.id}`">
              Continue Learning
            </RouterLink>
          </Button>
        </CardHeader>
        <CardContent v-if="progress" class="space-y-2">
          <div class="h-2 overflow-hidden rounded-full bg-muted">
            <div
              class="h-full rounded-full bg-primary transition-all"
              :style="{ width: `${Math.min(100, Math.max(0, progress.percentage))}%` }"
            />
          </div>
          <div class="flex justify-between text-xs text-muted-foreground">
            <span>{{ Math.round(progress.percentage) }}% complete</span>
            <span v-if="!continueLesson && lessons.length">All lessons complete</span>
            <span v-else-if="!lessons.length">No lessons yet</span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Lessons</CardTitle>
          <CardDescription>
            {{ learningLoading ? 'Loading lessons...' : `${lessons.length} lesson${lessons.length === 1 ? '' : 's'} in this course` }}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ul v-if="lessons.length" class="divide-y divide-border">
            <li
              v-for="(lessonItem, index) in lessons"
              :key="lessonItem.id"
              class="flex items-center justify-between gap-4 py-4 first:pt-0 last:pb-0"
            >
              <div class="min-w-0">
                <p class="font-medium">
                  {{ index + 1 }}. {{ lessonItem.title }}
                </p>
                <p class="text-sm text-muted-foreground">
                  {{ lessonItem.content ? truncateText(lessonItem.content, 50) : 'No content preview available.' }}
                </p>
              </div>
              <Button as-child size="sm" variant="outline" class="shrink-0">
                <RouterLink :to="`/courses/${course.id}/lessons/${lessonItem.id}`">
                  Open lesson
                </RouterLink>
              </Button>
            </li>
          </ul>
          <p v-else-if="learningLoading" class="text-sm text-muted-foreground">
            Loading lessons...
          </p>
          <p v-else class="text-sm text-muted-foreground">
            No lessons have been published yet.
          </p>
        </CardContent>
      </Card>
    </template>

    <Card v-if="enrolled">
      <CardHeader>
        <CardTitle>Assignments</CardTitle>
        <CardDescription>
          {{ assignments.length }} assignment{{ assignments.length === 1 ? '' : 's' }} for this course
        </CardDescription>
      </CardHeader>
      <CardContent>
        <ul
          v-if="assignments.length"
          class="divide-y divide-border"
        >
          <li
            v-for="assignment in assignments"
            :key="assignment.id"
            class="flex flex-col gap-2 py-4 first:pt-0 last:pb-0 sm:flex-row sm:items-center sm:justify-between"
          >
            <div class="space-y-1">
              <p class="font-medium">{{ assignment.title }}</p>
              <p class="text-sm text-muted-foreground">
                {{ assignment.description }}
              </p>
              <div class="flex items-center gap-3 text-xs text-muted-foreground">
                <span class="inline-flex items-center gap-1">
                  <Calendar class="h-3.5 w-3.5" />
                  Due {{ formatDate(assignment.due_date) }}
                </span>
                <span class="capitalize">{{ assignment.status }}</span>
              </div>
            </div>
            <Button
              as-child
              size="sm"
              variant="outline"
              class="shrink-0"
            >
              <RouterLink :to="`/courses/${course.id}/assignments/${assignment.id}`">
                View assignment
              </RouterLink>
            </Button>
          </li>
        </ul>
        <p
          v-else
          class="text-sm text-muted-foreground"
        >
          No assignments posted yet.
        </p>
      </CardContent>
    </Card>

    <Card v-else>
      <CardHeader>
        <CardTitle>Assignments</CardTitle>
        <CardDescription>Enroll to access this course’s assignments.</CardDescription>
      </CardHeader>
    </Card>
  </LearningPageShell>

  <LearningPageShell
    v-else
    eyebrow="Course"
    title="Course not found"
    description="The course you are looking for does not exist or has been removed."
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
            <BreadcrumbPage>Not found</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <Button
      as-child
      variant="outline"
    >
      <RouterLink to="/courses">
        Back to catalog
      </RouterLink>
    </Button>
  </LearningPageShell>
</template>
