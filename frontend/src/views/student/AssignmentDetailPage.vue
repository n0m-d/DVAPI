<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
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
import { Input } from '@/components/ui/input'
import { Calendar, Upload } from 'lucide-vue-next'
import FormActions from '@/components/forms/FormActions.vue'
import FormField from '@/components/forms/FormField.vue'
import FormTextarea from '@/components/forms/FormTextarea.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import {
  createSubmission,
  getAssignmentById,
  getMySubmission,
  listSubmissionVersions,
  resubmitAssignment,
  type ApiSubmission,
  type ApiSubmissionVersion,
} from '@/api/assignments'
import { getCourseById, type ApiAssignment, type ApiCourse } from '@/api/courses'
import { ApiError } from '@/lib/api'

const route = useRoute()

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v2'

const course = ref<ApiCourse | null>(null)
const assignment = ref<ApiAssignment | null>(null)
const existingSubmission = ref<ApiSubmission | null>(null)
const loading = ref(true)
const submitting = ref(false)
const errorMessage = ref('')
const submitError = ref('')
const notFound = ref(false)

const responseText = ref('')
const selectedFile = ref<File | null>(null)
const fileInputKey = ref(0)
const versions = ref<ApiSubmissionVersion[]>([])
const versionsLoading = ref(false)
const versionsError = ref('')

const submissionClosedReason = computed(() => {
  if (!assignment.value) return ''
  if (assignment.value.status !== 'published') {
    return assignment.value.status === 'closed'
      ? 'This assignment is closed.'
      : 'This assignment is not published yet.'
  }
  const dueDate = new Date(assignment.value.due_date)
  if (!Number.isNaN(dueDate.getTime()) && dueDate.getTime() < Date.now()) {
    return 'The due date has passed.'
  }
  return ''
})
const canSubmit = computed(() => Boolean(assignment.value) && !submissionClosedReason.value)

function courseId() {
  return (route.params.id as string) || assignment.value?.course_id || ''
}

function assignmentId() {
  return route.params.assignmentId as string
}

function formatDate(date: string) {
  const parsed = new Date(date)
  if (Number.isNaN(parsed.getTime())) return date
  return parsed.toLocaleDateString()
}

function attachmentUrl(path: string) {
  return `${API_BASE_URL}/get-uploads?path=${encodeURIComponent(path)}`
}

function onFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  selectedFile.value = input.files?.[0] ?? null
}

async function handleSubmit() {
  submitError.value = ''

  if (!canSubmit.value) {
    submitError.value = submissionClosedReason.value
    return
  }

  if (!responseText.value.trim()) {
    submitError.value = 'Submission text is required.'
    return
  }

  if (!selectedFile.value) {
    submitError.value = 'Please attach a file.'
    return
  }

  submitting.value = true

  try {
    if (existingSubmission.value) {
      await resubmitAssignment(existingSubmission.value.id, {
        submission_text: responseText.value.trim(),
        file: selectedFile.value,
      })
      toast.success('Assignment resubmitted successfully.')
    }
    else {
      const response = await createSubmission(assignmentId(), {
        submission_text: responseText.value.trim(),
        file: selectedFile.value,
      })
      toast.success(response.message || 'Assignment submitted successfully.')
    }

    await refreshSubmission()

    responseText.value = ''
    selectedFile.value = null
    fileInputKey.value += 1
  }
  catch (error) {
    submitError.value = error instanceof ApiError
      ? error.message
      : 'Unable to submit assignment. Please try again.'
  }
  finally {
    submitting.value = false
  }
}

async function loadVersions(submissionId: string) {
  versionsLoading.value = true
  versionsError.value = ''
  try {
    const response = await listSubmissionVersions(submissionId)
    versions.value = response.data ?? []
  }
  catch (error) {
    versions.value = []
    versionsError.value = error instanceof ApiError
      ? error.message
      : 'Unable to load submission history.'
  }
  finally {
    versionsLoading.value = false
  }
}

async function refreshSubmission() {
  const mine = await getMySubmission(assignmentId())
  existingSubmission.value = mine.data
  await loadVersions(mine.data.id)
}

async function load() {
  loading.value = true
  notFound.value = false
  errorMessage.value = ''
  submitError.value = ''
  course.value = null
  assignment.value = null
  existingSubmission.value = null
  responseText.value = ''
  selectedFile.value = null
  versions.value = []
  versionsError.value = ''
  fileInputKey.value += 1

  try {
    const assignmentResponse = await getAssignmentById(assignmentId())
    assignment.value = assignmentResponse.data

    const relatedCourseId = (route.params.id as string) || assignment.value.course_id
    try {
      course.value = await getCourseById(relatedCourseId)
    }
    catch {
      course.value = null
    }

    try {
      await refreshSubmission()
    }
    catch {
      existingSubmission.value = null
    }
  }
  catch (error) {
    notFound.value = true
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load assignment.'
  }
  finally {
    loading.value = false
  }
}

watch(
  () => [route.params.id, route.params.assignmentId],
  load,
)
onMounted(load)
</script>

<template>
  <LearningPageShell
    v-if="loading"
    eyebrow="Assignment"
    title="Loading assignment..."
    description="Fetching assignment details."
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
            <BreadcrumbPage>Loading</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <p class="text-sm text-muted-foreground">
      Loading assignment...
    </p>
  </LearningPageShell>

  <LearningPageShell
    v-else-if="assignment"
    :eyebrow="course?.slug || 'Assignment'"
    :title="assignment.title"
    :description="assignment.description"
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
            <BreadcrumbLink :href="`/courses/${courseId()}`">
              {{ course?.slug || 'Course' }}
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{{ assignment.title }}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <div class="flex flex-wrap items-center gap-3 text-sm text-muted-foreground">
      <span class="inline-flex items-center gap-1">
        <Calendar class="h-4 w-4" />
        Due {{ formatDate(assignment.due_date) }}
      </span>
      <span class="rounded-full border border-border bg-muted px-2.5 py-0.5 text-xs font-medium capitalize">
        {{ assignment.status }}
      </span>
    </div>

    <div class="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Assignment details</CardTitle>
          <CardDescription>Instructions and requirements</CardDescription>
        </CardHeader>
        <CardContent class="space-y-3 text-sm text-muted-foreground">
          <p>{{ assignment.description }}</p>
          <p>
            Submit your work before the due date. Late submissions may receive a penalty.
          </p>
        </CardContent>
      </Card>

      <Card class="form-card">
        <CardHeader>
          <CardTitle>Submit your work</CardTitle>
          <CardDescription>
            Upload your response and a required file attachment
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div
            v-if="existingSubmission"
            class="mb-4 space-y-2 rounded-lg border border-border/60 bg-muted/20 p-4 text-sm"
          >
            <p class="font-medium text-foreground">
              Already submitted
            </p>
            <p class="text-muted-foreground">
              {{ existingSubmission.submission_text }}
            </p>
            <p class="text-xs text-muted-foreground">
              Submitted {{ formatDate(existingSubmission.submitted_at) }}
              <span v-if="existingSubmission.file_name">
                · <a
                  :href="attachmentUrl(existingSubmission.file_path)"
                  target="_blank"
                  class="text-primary hover:underline"
                >
                  {{ existingSubmission.file_name }}
                </a>
              </span>
            </p>
          </div>

          <p
            v-if="submissionClosedReason"
            class="rounded-md border border-border bg-muted/40 px-3 py-2 text-sm text-muted-foreground"
          >
            {{ submissionClosedReason }}
          </p>

          <form
            v-if="canSubmit"
            class="app-form"
            @submit.prevent="handleSubmit"
          >
            <p
              v-if="submitError"
              class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
            >
              {{ submitError }}
            </p>

            <FormField
              label="Written response"
              html-for="response"
            >
              <FormTextarea
                id="response"
                v-model="responseText"
                :rows="5"
                placeholder="Type your answer or notes here..."
                required
              />
            </FormField>

            <FormField
              label="Attachment"
              html-for="file"
              hint="A file is required."
            >
              <div class="flex items-center gap-2">
                <Input
                  :key="fileInputKey"
                  id="file"
                  type="file"
                  class="cursor-pointer"
                  required
                  @change="onFileChange"
                />
                <Upload class="h-4 w-4 shrink-0 text-muted-foreground" />
              </div>
              <p
                v-if="selectedFile"
                class="form-hint"
              >
                Selected: {{ selectedFile.name }}
              </p>
            </FormField>

            <FormActions>
              <Button
                type="submit"
                :disabled="submitting"
              >
                {{ submitting ? 'Submitting...' : existingSubmission ? 'Submit again' : 'Submit assignment' }}
              </Button>
            </FormActions>
          </form>
        </CardContent>
      </Card>
    </div>

    <Card v-if="existingSubmission">
      <CardHeader>
        <CardTitle>Submission history</CardTitle>
        <CardDescription>Previous versions saved when you resubmit.</CardDescription>
      </CardHeader>
      <CardContent class="space-y-3">
        <p v-if="versionsLoading" class="text-sm text-muted-foreground">
          Loading submission history...
        </p>
        <p
          v-else-if="versionsError"
          class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
        >
          {{ versionsError }}
        </p>
        <template v-else>
          <div
            v-for="version in versions"
            :key="version.id"
            class="rounded-lg border border-border/60 p-4 text-sm"
          >
            <div class="flex flex-wrap items-center justify-between gap-2">
              <p class="font-medium">Version {{ version.version }}</p>
              <p class="text-xs text-muted-foreground">
                {{ formatDate(version.submitted_at || version.created_at) }}
              </p>
            </div>
            <p class="mt-2 whitespace-pre-wrap text-muted-foreground">
              {{ version.submission_text }}
            </p>
            <a
              v-if="version.file_path"
              :href="attachmentUrl(version.file_path)"
              target="_blank"
              rel="noopener noreferrer"
              class="mt-2 inline-block text-primary hover:underline"
            >
              {{ version.file_name || 'Download attachment' }}
            </a>
          </div>
          <p
            v-if="versions.length === 0"
            class="text-sm text-muted-foreground"
          >
            No previous versions yet.
          </p>
        </template>
      </CardContent>
    </Card>
  </LearningPageShell>

  <LearningPageShell
    v-else
    eyebrow="Assignment"
    title="Assignment not found"
    :description="errorMessage || 'This assignment does not exist or is no longer available.'"
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
