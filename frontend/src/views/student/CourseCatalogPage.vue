<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
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
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import CourseBrowseSection from '@/components/learning/CourseBrowseSection.vue'
import CourseCard from '@/components/learning/CourseCard.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { listEnrolledCourses, type ApiCourse } from '@/api/courses'
import { ApiError } from '@/lib/api'
import type { CourseCardItem } from '@/components/learning/CourseCard.vue'

const activeTab = ref('overview')
const search = ref('')
const loading = ref(false)
const errorMessage = ref('')
const apiCourses = ref<ApiCourse[]>([])
const enrolledCourseIds = ref<string[]>([])
const total = ref(0)

function mapEnrolledCourse(course: ApiCourse): CourseCardItem {
  return {
    id: course.id,
    code: course.slug,
    title: course.title,
    instructor: course.instructor.full_name,
    status: course.published ? 'active' : 'draft',
    term: new Date(course.created_at).getFullYear().toString(),
    description: course.description,
    enrolled: true,
  }
}

const overviewCourses = computed(() => apiCourses.value.map(mapEnrolledCourse))

async function fetchEnrolledCourses() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await listEnrolledCourses({
      page: 1,
      page_size: 24,
      title: search.value,
    })
    apiCourses.value = response.data.courses
    const fetchedIds = response.data.courses.map(course => course.id)
    enrolledCourseIds.value = search.value.trim()
      ? [...new Set([...enrolledCourseIds.value, ...fetchedIds])]
      : fetchedIds
    total.value = response.data.pagination.total
  }
  catch (error) {
    apiCourses.value = []
    total.value = 0
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load enrolled courses.'
  }
  finally {
    loading.value = false
  }
}

async function handleEnrolled(courseId: string) {
  if (!enrolledCourseIds.value.includes(courseId)) {
    enrolledCourseIds.value = [...enrolledCourseIds.value, courseId]
  }
  await fetchEnrolledCourses()
}

const debouncedSearch = useDebounceFn(fetchEnrolledCourses, 300)

watch(search, () => {
  debouncedSearch()
})

onMounted(fetchEnrolledCourses)
</script>

<template>
  <LearningPageShell
    eyebrow="Catalog"
    title="Course catalog"
    description="Browse available courses and enroll for the current academic term."
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
            <BreadcrumbPage>Courses</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <Tabs
      v-model="activeTab"
      class="w-full"
    >
      <TabsList class="grid w-full max-w-md grid-cols-2">
        <TabsTrigger value="overview">
          Overview
        </TabsTrigger>
        <TabsTrigger value="live">
          Live catalog
        </TabsTrigger>
      </TabsList>

      <TabsContent
        value="overview"
        class="mt-6"
      >
        <Card>
          <CardHeader class="gap-4 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <CardTitle>All courses</CardTitle>
              <CardDescription>
                {{ loading ? 'Loading enrolled courses...' : `${total} enrolled course${total === 1 ? '' : 's'}` }}
              </CardDescription>
            </div>
            <div class="relative w-full sm:max-w-xs">
              <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                v-model="search"
                type="search"
                placeholder="Search by title..."
                class="pl-9"
                aria-label="Search enrolled courses"
              />
            </div>
          </CardHeader>
          <CardContent>
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

            <div
              v-else-if="overviewCourses.length"
              class="grid gap-4 md:grid-cols-2 xl:grid-cols-3"
            >
              <CourseCard
                v-for="course in overviewCourses"
                :key="course.id"
                :course="course"
                :detail-link="`/courses/${course.id}`"
              />
            </div>

            <p
              v-else
              class="text-sm text-muted-foreground"
            >
              No enrolled courses found.
            </p>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent
        value="live"
        class="mt-6"
      >
        <CourseBrowseSection
          title="Live catalog"
          description="Search courses from the platform and open details in a modal."
          can-enroll
          :enrolled-course-ids="enrolledCourseIds"
          @enrolled="handleEnrolled"
        />
      </TabsContent>
    </Tabs>
  </LearningPageShell>
</template>
