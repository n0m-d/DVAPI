<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useDebounceFn } from '@vueuse/core'
import { Search } from 'lucide-vue-next'
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
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import CourseCard from '@/components/learning/CourseCard.vue'
import { listMyCourses, type ApiCourse } from '@/api/courses'
import { ApiError } from '@/lib/api'
import { useAuth } from '@/composables/useAuth'
import type { CourseCardItem } from '@/components/learning/CourseCard.vue'

const { user } = useAuth()

const search = ref('')
const loading = ref(true)
const errorMessage = ref('')
const apiCourses = ref<ApiCourse[]>([])
const page = ref(1)
const pageSize = 10
const total = ref(0)
const totalPages = ref(0)

function mapCourse(course: ApiCourse): CourseCardItem {
  return {
    id: course.id,
    code: course.slug,
    title: course.title,
    instructor: course.instructor.full_name || user.value.name,
    status: course.published ? 'active' : 'draft',
    term: new Date(course.created_at).getFullYear().toString(),
    description: course.description,
    enrolled: false,
  }
}

const myCourses = computed(() => apiCourses.value.map(mapCourse))

async function fetchCourses() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await listMyCourses({
      page: page.value,
      page_size: pageSize,
      title: search.value,
    })
    apiCourses.value = response.data.courses
    total.value = response.data.pagination.total
    totalPages.value = response.data.pagination.total_pages
    page.value = response.data.pagination.page
  }
  catch (error) {
    apiCourses.value = []
    total.value = 0
    totalPages.value = 0
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load your courses.'
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

function goToPage(next: number) {
  if (next < 1 || next > totalPages.value || next === page.value) return
  page.value = next
  fetchCourses()
}

onMounted(fetchCourses)
</script>

<template>
  <LearningPageShell
    eyebrow="Instructor"
    title="My courses"
    description="Open a course to manage content, grades, and roster."
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
            <BreadcrumbPage>Dashboard</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <p class="text-sm text-muted-foreground">
        {{ loading ? 'Loading courses...' : `${total} course${total === 1 ? '' : 's'} taught by ${user.name}` }}
      </p>
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
        <div class="relative w-full sm:w-64">
          <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="search"
            type="search"
            placeholder="Search by title..."
            class="pl-9"
            aria-label="Search my courses"
          />
        </div>
        <Button as-child>
          <RouterLink to="/instructor/courses/new">
            New course
          </RouterLink>
        </Button>
      </div>
    </div>

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
      <div
        v-if="myCourses.length"
        class="grid gap-4 md:grid-cols-2 xl:grid-cols-3"
      >
        <CourseCard
          v-for="course in myCourses"
          :key="course.id"
          :course="course"
          :detail-link="`/instructor/courses/${course.id}`"
        />
      </div>

      <Card v-else>
        <CardHeader>
          <CardTitle>No courses yet</CardTitle>
          <CardDescription>
            {{ search ? 'No courses match your search.' : 'Create your first course to get started.' }}
          </CardDescription>
        </CardHeader>
        <CardContent v-if="!search">
          <Button as-child>
            <RouterLink to="/instructor/courses/new">
              Create course
            </RouterLink>
          </Button>
        </CardContent>
      </Card>

      <div
        v-if="totalPages > 1"
        class="flex items-center justify-between gap-3"
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
  </LearningPageShell>
</template>
