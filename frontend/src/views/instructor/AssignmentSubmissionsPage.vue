<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import FormActions from '@/components/forms/FormActions.vue'
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
import FormField from '@/components/forms/FormField.vue'
import FormTextarea from '@/components/forms/FormTextarea.vue'
import {
  getSubmission,
  gradeSubmission,
  getAssignmentById,
  listAssignmentSubmissions,
  type ApiSubmission,
  type SubmissionDetail,
} from '@/api/assignments'
import { getCourseById, type ApiAssignment, type ApiCourse } from '@/api/courses'
import { ApiError } from '@/lib/api'

const route = useRoute()
const courseId = route.params.id as string
const assignmentId = route.params.assignmentId as string

const course = ref<ApiCourse | null>(null)
const assignment = ref<ApiAssignment | null>(null)
const submissions = ref<ApiSubmission[]>([])
const loading = ref(true)
const errorMessage = ref('')
const search = ref('')
const page = ref(1)
const pageSize = 10
const total = ref(0)
const totalPages = ref(0)
const dialogOpen = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const selectedSummary = ref<ApiSubmission | null>(null)
const submissionDetail = ref<SubmissionDetail | null>(null)
const grade = ref<number | undefined>()
const feedback = ref('')
const grading = ref(false)

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v2'

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

function attachmentUrl(path: string) {
  return `${API_BASE_URL}/get-uploads?path=${encodeURIComponent(path)}`
}

async function fetchSubmissions() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await listAssignmentSubmissions(assignmentId, {
      page: page.value,
      page_size: pageSize,
      name: search.value,
    })
    submissions.value = response.data.submissions
    total.value = response.data.pagination.total
    totalPages.value = response.data.pagination.total_pages
    page.value = response.data.pagination.page
  }
  catch (error) {
    submissions.value = []
    total.value = 0
    totalPages.value = 0
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load submissions.'
  }
  finally {
    loading.value = false
  }
}

const debouncedSearch = useDebounceFn(() => {
  page.value = 1
  fetchSubmissions()
}, 300)

watch(search, () => {
  debouncedSearch()
})

function goToPage(next: number) {
  if (next < 1 || next > totalPages.value || next === page.value) return
  page.value = next
  fetchSubmissions()
}

async function openSubmission(submission: ApiSubmission) {
  selectedSummary.value = submission
  submissionDetail.value = null
  grade.value = submission.grade === null ? undefined : submission.grade
  feedback.value = submission.feedback || ''
  detailError.value = ''
  detailLoading.value = true
  dialogOpen.value = true

  try {
    const response = await getSubmission(submission.id)
    submissionDetail.value = response.data
    feedback.value = response.data.feedback || ''
    // A detail grade of 0 is ambiguous. Preserve the list endpoint's nullable
    // state until a grade is explicitly saved in this dialog.
    grade.value = submission.grade === null ? undefined : response.data.grade
  }
  catch (error) {
    detailError.value = error instanceof ApiError
      ? error.message
      : 'Unable to load submission details.'
  }
  finally {
    detailLoading.value = false
  }
}

async function handleGrade() {
  if (!selectedSummary.value || grading.value) return
  if (grade.value === undefined || !Number.isFinite(Number(grade.value))) {
    detailError.value = 'Enter a grade from 0 to 100.'
    return
  }

  const numericGrade = Number(grade.value)
  if (numericGrade < 0 || numericGrade > 100) {
    detailError.value = 'Enter a grade from 0 to 100.'
    return
  }

  grading.value = true
  detailError.value = ''
  try {
    const response = await gradeSubmission(selectedSummary.value.id, {
      grade: numericGrade,
      feedback: feedback.value.trim(),
    })
    submissions.value = submissions.value.map(submission =>
      submission.id === selectedSummary.value?.id
        ? {
            ...submission,
            grade: response.data.grade,
            feedback: response.data.feedback,
            updated_at: response.data.updated_at,
          }
        : submission,
    )
    toast.success(response.message || 'Submission graded successfully.')
    dialogOpen.value = false
  }
  catch (error) {
    detailError.value = error instanceof ApiError
      ? error.message
      : 'Unable to grade submission.'
    toast.error(detailError.value)
  }
  finally {
    grading.value = false
  }
}

onMounted(async () => {
  const [courseResult, assignmentResult] = await Promise.allSettled([
    getCourseById(courseId),
    getAssignmentById(assignmentId),
  ])

  course.value = courseResult.status === 'fulfilled' ? courseResult.value : null
  assignment.value = assignmentResult.status === 'fulfilled'
    ? assignmentResult.value.data
    : null

  await fetchSubmissions()
})
</script>

<template>
  <LearningPageShell
    eyebrow="Submissions"
    :title="assignment?.title ?? 'Assignment submissions'"
    description="Review student work submitted for this assignment."
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
            <BreadcrumbLink :href="`/instructor/courses/${courseId}/assignments`">
              Assignments
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Submissions</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <InstructorCourseTabs />

    <Card>
      <CardHeader class="gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <CardTitle>Student submissions</CardTitle>
          <CardDescription>
            {{ loading ? 'Loading submissions...' : `${total} submission${total === 1 ? '' : 's'}` }}
          </CardDescription>
        </div>
        <div class="relative w-full sm:max-w-xs">
          <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="search"
            type="search"
            placeholder="Search by student name..."
            class="pl-9"
            aria-label="Search submissions by student name"
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
              <TableHead>Submitted</TableHead>
              <TableHead>Attachment</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Grade</TableHead>
              <TableHead class="text-right">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="loading">
              <TableCell
                colspan="6"
                class="text-center text-muted-foreground"
              >
                Loading submissions...
              </TableCell>
            </TableRow>
            <template v-else>
              <TableRow
                v-for="submission in submissions"
                :key="submission.id"
              >
                <TableCell>
                  <p class="font-medium">
                    {{ submission.student_name || 'Unknown student' }}
                  </p>
                  <p class="text-xs text-muted-foreground">
                    {{ submission.student_email || submission.student_id }}
                  </p>
                </TableCell>
                <TableCell>{{ formatDate(submission.submitted_at) }}</TableCell>
                <TableCell>
                  <a
                    v-if="submission.file_path"
                    :href="attachmentUrl(submission.file_path)"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-sm text-primary hover:underline"
                  >
                    {{ submission.file_name || 'Download file' }}
                  </a>
                  <span v-else class="text-muted-foreground">—</span>
                </TableCell>
                <TableCell>
                  <StatusBadge :status="submission.grade === null ? 'pending' : 'graded'" />
                </TableCell>
                <TableCell>
                  <span v-if="submission.grade !== null">{{ submission.grade }}</span>
                  <span v-else class="text-muted-foreground">—</span>
                </TableCell>
                <TableCell class="text-right">
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    @click="openSubmission(submission)"
                  >
                    Review / grade
                  </Button>
                </TableCell>
              </TableRow>
              <TableRow v-if="submissions.length === 0">
                <TableCell
                  colspan="6"
                  class="text-center text-muted-foreground"
                >
                  No submissions found for this assignment.
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

    <Dialog v-model:open="dialogOpen">
      <DialogContent class="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Review submission</DialogTitle>
          <DialogDescription>
            {{ selectedSummary?.student_name || 'Student submission' }}
          </DialogDescription>
        </DialogHeader>

        <form v-if="detailLoading" class="text-sm text-muted-foreground">
          Loading submission...
        </form>

        <form v-else-if="submissionDetail" class="space-y-6" @submit.prevent="handleGrade">
          <p
            v-if="detailError"
            class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
          >
            {{ detailError }}
          </p>

          <div class="space-y-3 rounded-lg border border-border/60 bg-muted/20 p-4 text-sm">
            <div>
              <p class="font-semibold">
                {{ submissionDetail.student_name || selectedSummary?.student_name || 'Unknown student' }}
              </p>
              <p class="text-xs text-muted-foreground">
                {{ submissionDetail.student_email || selectedSummary?.student_email || submissionDetail.student_id }}
              </p>
            </div>
            <div class="space-y-2">
              <p class="whitespace-pre-wrap text-muted-foreground">
                {{ submissionDetail.submission_text || 'No written response.' }}
              </p>
              <p class="text-xs text-muted-foreground">
                Submitted {{ formatDate(submissionDetail.submitted_at) }}
              </p>
            </div>
            <a
              v-if="submissionDetail.file_path"
              :href="attachmentUrl(submissionDetail.file_path)"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-block text-sm text-primary hover:underline"
            >
              {{ submissionDetail.file_name || 'Download attachment' }}
            </a>
          </div>

          <div class="space-y-5">
            <FormField label="Grade" html-for="submission-grade">
              <Input
                id="submission-grade"
                v-model.number="grade"
                type="number"
                min="0"
                max="100"
                step="1"
                required
              />
            </FormField>

            <FormField label="Feedback" html-for="submission-feedback">
              <FormTextarea
                id="submission-feedback"
                v-model="feedback"
                :rows="5"
                placeholder="Feedback for the student..."
              />
            </FormField>
          </div>

          <FormActions>
            <Button
              type="button"
              variant="outline"
              :disabled="grading"
              @click="dialogOpen = false"
            >
              Cancel
            </Button>
            <Button type="submit" :disabled="grading">
              {{ grading ? 'Saving...' : 'Save grade' }}
            </Button>
          </FormActions>
        </form>
      </DialogContent>
    </Dialog>
  </LearningPageShell>
</template>
