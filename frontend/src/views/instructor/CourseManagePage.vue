<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
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
import FormActions from '@/components/forms/FormActions.vue'
import FormCheckbox from '@/components/forms/FormCheckbox.vue'
import FormField from '@/components/forms/FormField.vue'
import FormTextarea from '@/components/forms/FormTextarea.vue'
import InstructorCourseTabs from '@/components/learning/InstructorCourseTabs.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { Input } from '@/components/ui/input'
import {
  deleteCourse,
  getCourseById,
  updateCourse,
  type ApiCourse,
} from '@/api/courses'
import { ApiError } from '@/lib/api'

const route = useRoute()
const router = useRouter()
const course = ref<ApiCourse | null>(null)
const title = ref('')
const description = ref('')
const published = ref(false)
const loading = ref(true)
const saving = ref(false)
const deleting = ref(false)
const errorMessage = ref('')
const fieldErrors = ref<Record<string, string>>({})

function courseId() {
  return route.params.id as string
}

async function loadCourse() {
  loading.value = true
  errorMessage.value = ''

  try {
    const data = await getCourseById(courseId())
    course.value = data
    title.value = data.title
    description.value = data.description
    published.value = data.published
  }
  catch (error) {
    course.value = null
    errorMessage.value = error instanceof ApiError ? error.message : 'Unable to load course.'
  }
  finally {
    loading.value = false
  }
}

async function handleSave() {
  if (saving.value || !course.value) return
  saving.value = true
  errorMessage.value = ''
  fieldErrors.value = {}

  try {
    const response = await updateCourse(courseId(), {
      title: title.value.trim(),
      description: description.value.trim(),
      published: published.value,
    })
    course.value = response.data
    title.value = response.data.title
    description.value = response.data.description
    published.value = response.data.published
    toast.success('Course updated successfully.')
  }
  catch (error) {
    if (error instanceof ApiError) {
      errorMessage.value = error.message
      fieldErrors.value = error.fieldErrors
    }
    else {
      errorMessage.value = 'Unable to update course.'
    }
    toast.error(errorMessage.value)
  }
  finally {
    saving.value = false
  }
}

async function handleDelete() {
  if (deleting.value || !course.value) return
  deleting.value = true

  try {
    await deleteCourse(courseId())
    toast.success('Course deleted successfully.')
    await router.push('/instructor/dashboard')
  }
  catch (error) {
    toast.error(error instanceof ApiError ? error.message : 'Unable to delete course.')
    deleting.value = false
  }
}

watch(() => route.params.id, loadCourse)
onMounted(loadCourse)
</script>

<template>
  <LearningPageShell
    eyebrow="Course"
    :title="course?.title ?? (loading ? 'Loading course...' : 'Course')"
    description="Edit course details and publishing status."
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
            <BreadcrumbLink href="/instructor/dashboard">
              Courses
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{{ course?.slug ?? 'Course' }}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <InstructorCourseTabs />

    <p
      v-if="errorMessage"
      class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
    >
      {{ errorMessage }}
    </p>

    <p
      v-if="loading"
      class="text-sm text-muted-foreground"
    >
      Loading course...
    </p>

    <div
      v-else-if="course"
      class="grid gap-6"
    >
      <Card class="form-card">
        <CardHeader>
          <CardTitle>Course details</CardTitle>
          <CardDescription>
            Information shown to students in the course catalog.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form
            class="app-form"
            @submit.prevent="handleSave"
          >
            <FormField label="Title" html-for="title" :error="fieldErrors.title">
              <Input
                id="title"
                v-model="title"
                required
              />
            </FormField>

            <FormField label="Description" html-for="description" :error="fieldErrors.description">
              <FormTextarea
                id="description"
                v-model="description"
                :rows="6"
                placeholder="Brief overview of the course..."
              />
            </FormField>

            <FormCheckbox
              v-model="published"
              title="Published"
              description="Published courses are visible in the student catalog."
            />
          </form>
        </CardContent>
      </Card>

      <FormActions class="!border-0 !pt-0 justify-between">
        <AlertDialog>
          <AlertDialogTrigger as-child>
            <Button variant="destructive" :disabled="saving || deleting">
              Delete course
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Delete this course?</AlertDialogTitle>
              <AlertDialogDescription>
                This permanently deletes “{{ course.title }}” and cannot be undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel :disabled="deleting">Cancel</AlertDialogCancel>
              <AlertDialogAction
                class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                :disabled="deleting"
                @click="handleDelete"
              >
                {{ deleting ? 'Deleting...' : 'Delete course' }}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>

        <Button :disabled="saving || deleting" @click="handleSave">
          {{ saving ? 'Saving...' : 'Save changes' }}
        </Button>
      </FormActions>
    </div>
  </LearningPageShell>
</template>
