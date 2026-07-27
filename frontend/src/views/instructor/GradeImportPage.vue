<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Upload } from 'lucide-vue-next'
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
import FormActions from '@/components/forms/FormActions.vue'
import InstructorCourseTabs from '@/components/learning/InstructorCourseTabs.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { getCourseById, type ApiCourse } from '@/api/courses'
import { ApiError } from '@/lib/api'

const route = useRoute()
const courseId = computed(() => route.params.id as string)
const course = ref<ApiCourse | null>(null)
const fileName = ref<string | null>(null)
const errorMessage = ref('')

async function loadCourse() {
  errorMessage.value = ''
  try {
    course.value = await getCourseById(courseId.value)
  }
  catch (error) {
    course.value = null
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load course.'
  }
}

function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  fileName.value = input.files?.[0]?.name ?? null
}

function handleImport() {
  // Import wiring pending API support
}

onMounted(loadCourse)
</script>

<template>
  <LearningPageShell
    eyebrow="Grades"
    title="Import grades"
    description="Upload a CSV file to bulk-import grades for this course."
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
            <BreadcrumbLink :href="`/instructor/courses/${courseId}/grades`">
              Grades
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Import</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <InstructorCourseTabs />

    <div
      v-if="errorMessage"
      class="mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-3 text-sm text-destructive"
    >
      {{ errorMessage }}
    </div>

    <Card class="form-card max-w-2xl">
      <CardHeader>
        <CardTitle>CSV upload</CardTitle>
        <CardDescription>
          Expected columns: student_email, assignment, score, max_score
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form
          class="app-form"
          @submit.prevent="handleImport"
        >
          <div class="form-upload-zone">
            <Upload class="h-10 w-10 text-muted-foreground" />
            <div class="space-y-1">
              <p class="text-sm font-medium">
                Drag and drop your CSV file here
              </p>
              <p class="text-xs text-muted-foreground">
                or choose a file from your computer
              </p>
            </div>
            <label
              for="csv-upload"
              class="cursor-pointer"
            >
              <span class="sr-only">Choose CSV file</span>
              <Input
                id="csv-upload"
                type="file"
                accept=".csv"
                @change="handleFileChange"
              />
            </label>
            <p
              v-if="fileName"
              class="text-sm text-muted-foreground"
            >
              Selected: {{ fileName }}
            </p>
          </div>

          <FormActions>
            <Button
              type="submit"
              :disabled="!fileName"
            >
              Import grades
            </Button>
          </FormActions>
        </form>
      </CardContent>
    </Card>
  </LearningPageShell>
</template>
