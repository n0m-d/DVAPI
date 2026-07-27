<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import FormActions from '@/components/forms/FormActions.vue'
import FormField from '@/components/forms/FormField.vue'
import FormTextarea from '@/components/forms/FormTextarea.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import {
  createAssignment,
  getAssignmentById,
  updateAssignment,
  type AssignmentStatus,
} from '@/api/assignments'
import { getCourseById, type ApiCourse } from '@/api/courses'
import { ApiError } from '@/lib/api'

const route = useRoute()
const router = useRouter()
const courseId = route.params.id as string
const assignmentId = route.params.assignmentId as string | undefined
const isEditing = computed(() => Boolean(assignmentId))

const course = ref<ApiCourse | null>(null)
const title = ref('')
const description = ref('')
const dueAt = ref('')
const status = ref<AssignmentStatus>('draft')
const loading = ref(true)
const saving = ref(false)
const errorMessage = ref('')
const fieldErrors = ref<Record<string, string>>({})

function toLocalDateTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

async function load() {
  loading.value = true
  errorMessage.value = ''

  try {
    const [courseResult, assignmentResult] = await Promise.all([
      getCourseById(courseId),
      assignmentId ? getAssignmentById(assignmentId) : Promise.resolve(null),
    ])
    course.value = courseResult

    if (assignmentResult) {
      const assignment = assignmentResult.data
      if (assignment.course_id !== courseId) {
        throw new Error('Assignment does not belong to this course.')
      }
      title.value = assignment.title
      description.value = assignment.description
      dueAt.value = toLocalDateTime(assignment.due_date)
      status.value = assignment.status as AssignmentStatus
    }
  }
  catch (error) {
    errorMessage.value = error instanceof ApiError
      ? error.message
      : error instanceof Error
        ? error.message
        : 'Unable to load assignment.'
  }
  finally {
    loading.value = false
  }
}

async function handleSubmit() {
  if (saving.value) return
  errorMessage.value = ''
  fieldErrors.value = {}

  const dueDate = new Date(dueAt.value)
  if (Number.isNaN(dueDate.getTime())) {
    fieldErrors.value.due_date = 'Enter a valid due date and time.'
    return
  }

  saving.value = true
  try {
    const payload = {
      title: title.value.trim(),
      description: description.value.trim(),
      due_date: dueDate.toISOString(),
      status: status.value,
    }

    let successMessage = 'Assignment updated successfully.'
    if (assignmentId) {
      await updateAssignment(assignmentId, payload)
    }
    else {
      const response = await createAssignment({ ...payload, course_id: courseId })
      successMessage = response.message || 'Assignment created successfully.'
    }

    toast.success(successMessage)
    await router.push(`/instructor/courses/${courseId}/assignments`)
  }
  catch (error) {
    if (error instanceof ApiError) {
      errorMessage.value = error.message
      fieldErrors.value = error.fieldErrors
    }
    else {
      errorMessage.value = `Unable to ${isEditing.value ? 'update' : 'create'} assignment.`
    }
    toast.error(errorMessage.value)
  }
  finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<template>
  <LearningPageShell
    eyebrow="Assignment"
    :title="isEditing ? 'Edit assignment' : 'New assignment'"
    :description="course ? `${isEditing ? 'Update' : 'Create'} an assignment for ${course.slug}` : 'Manage assignment details.'"
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
            <BreadcrumbPage>{{ isEditing ? 'Edit assignment' : 'New assignment' }}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <Card class="form-card max-w-2xl">
      <CardHeader>
        <CardTitle>Assignment details</CardTitle>
        <CardDescription>
          Students will see this on the course assignments page.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <p
          v-if="errorMessage"
          class="mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
        >
          {{ errorMessage }}
        </p>
        <p v-if="loading" class="text-sm text-muted-foreground">
          Loading assignment...
        </p>
        <form
          v-else
          class="app-form"
          @submit.prevent="handleSubmit"
        >
          <FormField
            label="Title"
            html-for="assignment-title"
            :error="fieldErrors.title"
          >
            <Input
              id="assignment-title"
              v-model="title"
              placeholder="Lab 7: Recursion"
              required
            />
          </FormField>

          <FormField
            label="Description"
            html-for="assignment-description"
            hint="Include instructions, requirements, and grading criteria."
            :error="fieldErrors.description"
          >
            <FormTextarea
              id="assignment-description"
              v-model="description"
              placeholder="Instructions and requirements..."
            />
          </FormField>

          <FormField
            label="Due date"
            html-for="due-at"
            :error="fieldErrors.due_date"
          >
            <Input
              id="due-at"
              v-model="dueAt"
              type="datetime-local"
              required
            />
          </FormField>

          <FormField label="Status" :error="fieldErrors.status">
            <Select v-model="status">
              <SelectTrigger>
                <SelectValue placeholder="Select status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="draft">
                  Draft
                </SelectItem>
                <SelectItem value="published">
                  Published
                </SelectItem>
                <SelectItem value="closed">
                  Closed
                </SelectItem>
              </SelectContent>
            </Select>
          </FormField>

          <FormActions>
            <Button type="submit" :disabled="saving">
              {{ saving ? 'Saving...' : isEditing ? 'Save changes' : 'Create assignment' }}
            </Button>
            <Button
              type="button"
              variant="outline"
              :disabled="saving"
              @click="router.push(`/instructor/courses/${courseId}/assignments`)"
            >
              Cancel
            </Button>
          </FormActions>
        </form>
      </CardContent>
    </Card>
  </LearningPageShell>
</template>
