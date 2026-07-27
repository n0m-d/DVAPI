<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Avatar,
  AvatarFallback,
} from '@/components/ui/avatar'
import { Input } from '@/components/ui/input'
import InstructorCourseTabs from '@/components/learning/InstructorCourseTabs.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import {
  getCourseById,
  listCourseStudents,
  type ApiCourse,
  type ApiEnrolledStudent,
} from '@/api/courses'
import { ApiError } from '@/lib/api'

const route = useRoute()
const courseId = route.params.id as string

const course = ref<ApiCourse | null>(null)
const students = ref<ApiEnrolledStudent[]>([])
const loading = ref(true)
const errorMessage = ref('')
const search = ref('')
const page = ref(1)
const pageSize = 10
const total = ref(0)
const totalPages = ref(0)

function initials(name: string) {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .map(part => part[0] ?? '')
    .join('')
    .slice(0, 2)
    .toUpperCase()
}

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleDateString()
}

async function fetchRoster() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await listCourseStudents(courseId, {
      page: page.value,
      page_size: pageSize,
      name: search.value,
    })
    students.value = response.data.students
    total.value = response.data.pagination.total
    totalPages.value = response.data.pagination.total_pages
    page.value = response.data.pagination.page
  }
  catch (error) {
    students.value = []
    total.value = 0
    totalPages.value = 0
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load roster.'
  }
  finally {
    loading.value = false
  }
}

const debouncedSearch = useDebounceFn(() => {
  page.value = 1
  fetchRoster()
}, 300)

watch(search, () => {
  debouncedSearch()
})

function goToPage(next: number) {
  if (next < 1 || next > totalPages.value || next === page.value) return
  page.value = next
  fetchRoster()
}

onMounted(async () => {
  try {
    course.value = await getCourseById(courseId)
  }
  catch {
    course.value = null
  }
  await fetchRoster()
})
</script>

<template>
  <LearningPageShell
    eyebrow="Roster"
    title="Student roster"
    description="Enrolled students for this course."
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
            <BreadcrumbPage>Roster</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <InstructorCourseTabs />

    <Card>
      <CardHeader class="gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <CardTitle>Enrolled students</CardTitle>
          <CardDescription>
            {{ loading ? 'Loading roster...' : `${total} student${total === 1 ? '' : 's'} on the roster` }}
          </CardDescription>
        </div>
        <div class="relative w-full sm:max-w-xs">
          <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="search"
            type="search"
            placeholder="Search by name..."
            class="pl-9"
            aria-label="Search students by name"
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

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Student</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Enrolled</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="loading">
              <TableCell
                colspan="3"
                class="text-center text-muted-foreground"
              >
                Loading students...
              </TableCell>
            </TableRow>
            <template v-else>
              <TableRow
                v-for="student in students"
                :key="student.id"
              >
                <TableCell>
                  <div class="flex items-center gap-3">
                    <Avatar class="h-8 w-8">
                      <AvatarFallback>{{ initials(student.full_name) }}</AvatarFallback>
                    </Avatar>
                    <p class="font-medium">{{ student.full_name }}</p>
                  </div>
                </TableCell>
                <TableCell class="text-muted-foreground">
                  {{ student.email }}
                </TableCell>
                <TableCell>
                  {{ formatDate(student.enrolled_at) }}
                </TableCell>
              </TableRow>
              <TableRow v-if="students.length === 0">
                <TableCell
                  colspan="3"
                  class="text-center text-muted-foreground"
                >
                  No students enrolled in this course yet.
                </TableCell>
              </TableRow>
            </template>
          </TableBody>
        </Table>

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
      </CardContent>
    </Card>
  </LearningPageShell>
</template>
