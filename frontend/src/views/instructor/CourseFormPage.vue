<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
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
import FormActions from '@/components/forms/FormActions.vue'
import FormCheckbox from '@/components/forms/FormCheckbox.vue'
import FormField from '@/components/forms/FormField.vue'
import FormTextarea from '@/components/forms/FormTextarea.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { createCourse } from '@/api/courses'
import { ApiError } from '@/lib/api'

const router = useRouter()

const title = ref('')
const description = ref('')
const published = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const fieldErrors = ref<Record<string, string>>({})

async function handleSubmit() {
  if (saving.value) return

  saving.value = true
  errorMessage.value = ''
  fieldErrors.value = {}

  try {
    const response = await createCourse({
      title: title.value.trim(),
      description: description.value.trim(),
      published: published.value,
    })
    toast.success('Course created successfully.')
    await router.push(`/instructor/courses/${response.data.id}`)
  }
  catch (error) {
    if (error instanceof ApiError) {
      errorMessage.value = error.message
      fieldErrors.value = error.fieldErrors
    }
    else {
      errorMessage.value = 'Unable to create course.'
    }
    toast.error(errorMessage.value)
  }
  finally {
    saving.value = false
  }
}
</script>

<template>
  <LearningPageShell
    eyebrow="Instructor"
    title="Create course"
    description="Add a new course offering for an upcoming term."
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
            <BreadcrumbPage>New course</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <Card class="form-card max-w-2xl">
      <CardHeader>
        <CardTitle>Course details</CardTitle>
        <CardDescription>
          Basic information visible to enrolled students.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form
          class="app-form"
          @submit.prevent="handleSubmit"
        >
          <p
            v-if="errorMessage"
            class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
          >
            {{ errorMessage }}
          </p>

          <FormField
            label="Title"
            html-for="title"
            hint="Full course name as shown on the catalog."
            :error="fieldErrors.title"
          >
            <Input
              id="title"
              v-model="title"
              placeholder="Introduction to Programming"
              required
            />
          </FormField>

          <FormField
            label="Description"
            html-for="description"
            :error="fieldErrors.description"
          >
            <FormTextarea
              id="description"
              v-model="description"
              placeholder="Brief overview of the course..."
            />
          </FormField>

          <FormCheckbox
            v-model="published"
            title="Publish course"
            description="Published courses are visible in the student catalog."
          />

          <FormActions>
            <Button
              type="submit"
              :disabled="saving"
            >
              {{ saving ? 'Creating...' : 'Create course' }}
            </Button>
            <Button
              type="button"
              variant="outline"
              :disabled="saving"
              @click="router.push('/instructor/dashboard')"
            >
              Cancel
            </Button>
          </FormActions>
        </form>
      </CardContent>
    </Card>
  </LearningPageShell>
</template>
