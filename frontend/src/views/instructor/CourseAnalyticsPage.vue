<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  BookOpen,
  CheckCircle2,
  ClipboardList,
  GraduationCap,
  Send,
  UserCog,
} from 'lucide-vue-next'
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
import InstructorCourseTabs from '@/components/learning/InstructorCourseTabs.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import StatCard from '@/components/learning/StatCard.vue'
import { getCourseAnalytics, type CourseAnalytics } from '@/api/analytics'
import { getCourseById, type ApiCourse } from '@/api/courses'
import { ApiError } from '@/lib/api'

const route = useRoute()
const courseId = route.params.id as string
const course = ref<ApiCourse | null>(null)
const analytics = ref<CourseAnalytics | null>(null)
const loading = ref(true)
const errorMessage = ref('')

async function loadAnalytics() {
  loading.value = true
  errorMessage.value = ''

  try {
    const [courseResult, analyticsResult] = await Promise.all([
      getCourseById(courseId),
      getCourseAnalytics(courseId),
    ])
    course.value = courseResult
    analytics.value = analyticsResult.data
  }
  catch (error) {
    analytics.value = null
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load course analytics.'
  }
  finally {
    loading.value = false
  }
}

onMounted(loadAnalytics)
</script>

<template>
  <LearningPageShell
    eyebrow="Course analytics"
    title="Performance overview"
    description="Live engagement and assessment totals for this course."
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
            <BreadcrumbLink :href="`/instructor/courses/${courseId}`">
              {{ course?.slug ?? 'Course' }}
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Analytics</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <InstructorCourseTabs />

    <div
      v-if="errorMessage"
      class="flex items-center justify-between gap-3 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-3 text-sm text-destructive"
    >
      <span>{{ errorMessage }}</span>
      <Button variant="outline" size="sm" @click="loadAnalytics">
        Retry
      </Button>
    </div>

    <p v-else-if="loading" class="text-sm text-muted-foreground">
      Loading course analytics...
    </p>

    <template v-else-if="analytics">
      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <StatCard title="Enrollments" :value="analytics.enrollments" delta="Students enrolled" :icon="UserCog" />
        <StatCard title="Lessons" :value="analytics.lessons" delta="Lessons in course" :icon="BookOpen" />
        <StatCard title="Lesson completions" :value="analytics.lesson_completions" delta="Recorded completions" :icon="CheckCircle2" />
        <StatCard title="Assignments" :value="analytics.assignments" delta="Assignments in course" :icon="ClipboardList" />
        <StatCard title="Submissions" :value="analytics.submissions" delta="Student submissions" :icon="Send" />
        <StatCard title="Average grade" :value="analytics.average_grade.toFixed(1)" delta="Across graded submissions" :icon="GraduationCap" />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Analytics source</CardTitle>
          <CardDescription>The course identifier returned with these metrics.</CardDescription>
        </CardHeader>
        <CardContent>
          <code class="break-all text-sm">{{ analytics.course_id }}</code>
        </CardContent>
      </Card>
    </template>
  </LearningPageShell>
</template>
