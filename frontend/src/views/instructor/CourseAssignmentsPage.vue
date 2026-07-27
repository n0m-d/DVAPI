<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useDebounceFn } from '@vueuse/core'
import { Search } from 'lucide-vue-next'
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import InstructorCourseTabs from '@/components/learning/InstructorCourseTabs.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import StatusBadge from '@/components/learning/StatusBadge.vue'
import {
  getCourseById,
  listMyCourseAssignments,
  type ApiAssignment,
  type ApiCourse,
} from '@/api/courses'
import { deleteAssignment } from '@/api/assignments'
import { ApiError } from '@/lib/api'

const route = useRoute()
const courseId = route.params.id as string

const course = ref<ApiCourse | null>(null)
const assignments = ref<ApiAssignment[]>([])
const loading = ref(true)
const errorMessage = ref('')
const search = ref('')
const page = ref(1)
const pageSize = 10
const total = ref(0)
const totalPages = ref(0)
const deletingId = ref('')

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

async function fetchAssignments() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await listMyCourseAssignments(courseId, {
      page: page.value,
      page_size: pageSize,
      title: search.value,
    })
    assignments.value = response.data.assignments
    total.value = response.data.pagination.total
    totalPages.value = response.data.pagination.total_pages
    page.value = response.data.pagination.page
  }
  catch (error) {
    assignments.value = []
    total.value = 0
    totalPages.value = 0
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load assignments.'
  }
  finally {
    loading.value = false
  }
}

const debouncedSearch = useDebounceFn(() => {
  page.value = 1
  fetchAssignments()
}, 300)

watch(search, () => {
  debouncedSearch()
})

function goToPage(next: number) {
  if (next < 1 || next > totalPages.value || next === page.value) return
  page.value = next
  fetchAssignments()
}

async function handleDelete(assignment: ApiAssignment) {
  if (deletingId.value) return
  deletingId.value = assignment.id

  try {
    await deleteAssignment(assignment.id)
    toast.success('Assignment deleted successfully.')
    if (assignments.value.length === 1 && page.value > 1) page.value -= 1
    await fetchAssignments()
  }
  catch (error) {
    toast.error(error instanceof ApiError ? error.message : 'Unable to delete assignment.')
  }
  finally {
    deletingId.value = ''
  }
}

onMounted(async () => {
  try {
    course.value = await getCourseById(courseId)
  }
  catch {
    course.value = null
  }
  await fetchAssignments()
})
</script>

<template>
  <LearningPageShell
    eyebrow="Assignments"
    title="Course assignments"
    description="Manage assignments and review their submissions."
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
            <BreadcrumbPage>Assignments</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <InstructorCourseTabs />

    <Card>
      <CardHeader class="gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <CardTitle>Assignments</CardTitle>
          <CardDescription>
            {{ loading ? 'Loading assignments...' : `${total} assignment${total === 1 ? '' : 's'}` }}
          </CardDescription>
        </div>
        <div class="flex w-full flex-col gap-2 sm:max-w-md sm:flex-row">
          <div class="relative flex-1">
            <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              v-model="search"
              type="search"
              placeholder="Search by title..."
              class="pl-9"
              aria-label="Search assignments by title"
            />
          </div>
          <Button as-child>
            <RouterLink :to="`/instructor/courses/${courseId}/assignments/new`">
              New assignment
            </RouterLink>
          </Button>
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
              <TableHead>Title</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Due date</TableHead>
              <TableHead class="text-right">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="loading">
              <TableCell
                colspan="4"
                class="text-center text-muted-foreground"
              >
                Loading assignments...
              </TableCell>
            </TableRow>
            <template v-else>
              <TableRow
                v-for="assignment in assignments"
                :key="assignment.id"
              >
                <TableCell>
                  <p class="font-medium">{{ assignment.title }}</p>
                  <p class="max-w-md truncate text-xs text-muted-foreground">
                    {{ assignment.description }}
                  </p>
                </TableCell>
                <TableCell>
                  <StatusBadge :status="assignment.status" />
                </TableCell>
                <TableCell>{{ formatDate(assignment.due_date) }}</TableCell>
                <TableCell class="text-right">
                  <div class="flex justify-end gap-2">
                    <Button as-child variant="outline" size="sm">
                      <RouterLink
                        :to="`/instructor/courses/${courseId}/assignments/${assignment.id}/submissions`"
                      >
                        Submissions
                      </RouterLink>
                    </Button>
                    <Button as-child variant="outline" size="sm">
                      <RouterLink
                        :to="`/instructor/courses/${courseId}/assignments/${assignment.id}/edit`"
                      >
                        Edit
                      </RouterLink>
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
                          <AlertDialogTitle>Delete this assignment?</AlertDialogTitle>
                          <AlertDialogDescription>
                            This permanently deletes “{{ assignment.title }}” and its submissions.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel :disabled="deletingId === assignment.id">
                            Cancel
                          </AlertDialogCancel>
                          <AlertDialogAction
                            class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                            :disabled="deletingId === assignment.id"
                            @click="handleDelete(assignment)"
                          >
                            {{ deletingId === assignment.id ? 'Deleting...' : 'Delete assignment' }}
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </div>
                </TableCell>
              </TableRow>
              <TableRow v-if="assignments.length === 0">
                <TableCell
                  colspan="4"
                  class="text-center text-muted-foreground"
                >
                  No assignments found.
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
