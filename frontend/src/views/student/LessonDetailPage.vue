<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
import { CheckCircle2, Circle } from 'lucide-vue-next'
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
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { getCourseById, type ApiCourse } from '@/api/courses'
import {
  getContinueLesson,
  getCourseProgress,
  listStudentLessons,
  updateLessonProgress,
  type ApiLesson,
  type CourseProgress,
} from '@/api/lessons'
import { ApiError } from '@/lib/api'

const route = useRoute()

const course = ref<ApiCourse | null>(null)
const lessons = ref<ApiLesson[]>([])
const lesson = ref<ApiLesson | null>(null)
const progress = ref<CourseProgress | null>(null)
const completed = ref(false)
const loading = ref(true)
const savingProgress = ref(false)
const errorMessage = ref('')
const notFound = ref(false)

function courseId() {
  return route.params.id as string
}

function lessonId() {
  return route.params.lessonId as string
}

const lessonIndex = computed(() =>
  lessons.value.findIndex(item => item.id === lesson.value?.id),
)

const previousLesson = computed(() =>
  lessonIndex.value > 0 ? lessons.value[lessonIndex.value - 1] : null,
)

const nextLesson = computed(() =>
  lessonIndex.value >= 0 && lessonIndex.value < lessons.value.length - 1
    ? lessons.value[lessonIndex.value + 1]
    : null,
)

function orderedLessons(items: ApiLesson[]) {
  return [...items].sort((a, b) =>
    a.sort_order - b.sort_order || a.created_at.localeCompare(b.created_at),
  )
}

async function inferCompletion(items: ApiLesson[], current: ApiLesson) {
  try {
    const response = await getContinueLesson(courseId())
    const continueIndex = items.findIndex(item => item.id === response.data.id)
    const currentIndex = items.findIndex(item => item.id === current.id)
    completed.value = continueIndex >= 0 && currentIndex >= 0 && currentIndex < continueIndex
  }
  catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      completed.value = items.length > 0
      return
    }
    throw error
  }
}

async function loadLesson() {
  loading.value = true
  errorMessage.value = ''
  notFound.value = false
  lesson.value = null

  try {
    const [courseData, lessonsData, progressData] = await Promise.all([
      getCourseById(courseId()),
      listStudentLessons(courseId()),
      getCourseProgress(courseId()),
    ])
    course.value = courseData
    lessons.value = orderedLessons(lessonsData.data ?? [])
    progress.value = progressData.data

    const selected = lessons.value.find(item => item.id === lessonId())
    if (!selected) {
      notFound.value = true
      return
    }

    lesson.value = selected
    await inferCompletion(lessons.value, selected)
  }
  catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound.value = true
    }
    else {
      errorMessage.value = error instanceof ApiError
        ? error.message
        : 'Unable to load lesson.'
    }
  }
  finally {
    loading.value = false
  }
}

async function setCompleted(value: boolean) {
  if (!lesson.value || savingProgress.value || completed.value === value) return
  savingProgress.value = true

  try {
    const response = await updateLessonProgress(lesson.value.id, { completed: value })
    completed.value = value
    toast.success(response.message || (value ? 'Lesson marked complete.' : 'Lesson marked incomplete.'))

    const progressData = await getCourseProgress(courseId())
    progress.value = progressData.data
  }
  catch (error) {
    toast.error(error instanceof ApiError ? error.message : 'Unable to update lesson progress.')
  }
  finally {
    savingProgress.value = false
  }
}

watch(
  () => [route.params.id, route.params.lessonId],
  loadLesson,
)
onMounted(loadLesson)
</script>

<template>
  <LearningPageShell
    v-if="loading"
    eyebrow="Lesson"
    title="Loading lesson..."
    description="Fetching lesson content and progress."
  >
    <p class="text-sm text-muted-foreground">Loading lesson...</p>
  </LearningPageShell>

  <LearningPageShell
    v-else-if="lesson && course"
    :eyebrow="`Lesson ${lessonIndex + 1} of ${lessons.length}`"
    :title="lesson.title"
    :description="course.title"
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
            <BreadcrumbLink :href="`/courses/${course.id}`">
              {{ course.title }}
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{{ lesson.title }}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <p
      v-if="errorMessage"
      class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
    >
      {{ errorMessage }}
    </p>

    <div class="flex flex-wrap items-center gap-3">
      <div class="text-sm text-muted-foreground">
        <span v-if="progress">
          {{ progress.completed_lessons }} of {{ progress.total_lessons }} complete
          ({{ Math.round(progress.percentage) }}%)
        </span>
      </div>
      <Button
        class="ml-auto"
        :variant="completed ? 'outline' : 'default'"
        :disabled="savingProgress"
        @click="setCompleted(!completed)"
      >
        <CheckCircle2 v-if="completed" class="mr-2 h-4 w-4" />
        <Circle v-else class="mr-2 h-4 w-4" />
        {{ savingProgress ? 'Updating...' : (completed ? 'Mark incomplete' : 'Mark complete') }}
      </Button>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>{{ lesson.title }}</CardTitle>
        <CardDescription>Lesson {{ lessonIndex + 1 }} in this course</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="whitespace-pre-wrap text-sm leading-7">
          {{ lesson.content || 'No lesson content has been published yet.' }}
        </div>
      </CardContent>
    </Card>

    <div class="flex items-center justify-between gap-3">
      <Button
        v-if="previousLesson"
        as-child
        variant="outline"
      >
        <RouterLink :to="`/courses/${course.id}/lessons/${previousLesson.id}`">
          ← {{ previousLesson.title }}
        </RouterLink>
      </Button>
      <span v-else />

      <Button
        v-if="nextLesson"
        as-child
      >
        <RouterLink :to="`/courses/${course.id}/lessons/${nextLesson.id}`">
          {{ nextLesson.title }} →
        </RouterLink>
      </Button>
      <Button v-else as-child variant="outline">
        <RouterLink :to="`/courses/${course.id}`">
          Back to course
        </RouterLink>
      </Button>
    </div>
  </LearningPageShell>

  <LearningPageShell
    v-else-if="notFound"
    eyebrow="Lesson"
    title="Lesson not found"
    description="This lesson does not exist or is not available in this course."
  >
    <Button as-child variant="outline">
      <RouterLink :to="`/courses/${courseId()}`">
        Back to course
      </RouterLink>
    </Button>
  </LearningPageShell>

  <LearningPageShell
    v-else
    eyebrow="Lesson"
    title="Unable to load lesson"
    :description="errorMessage || 'Please try again.'"
  >
    <Button type="button" variant="outline" @click="loadLesson">
      Try again
    </Button>
  </LearningPageShell>
</template>
