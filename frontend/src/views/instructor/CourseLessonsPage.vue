<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
import { Input } from '@/components/ui/input'
import FormField from '@/components/forms/FormField.vue'
import FormActions from '@/components/forms/FormActions.vue'
import FormTextarea from '@/components/forms/FormTextarea.vue'
import InstructorCourseTabs from '@/components/learning/InstructorCourseTabs.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { getCourseById, type ApiCourse } from '@/api/courses'
import {
  createLesson,
  deleteLesson,
  listInstructorLessons,
  updateLesson,
  type ApiLesson,
} from '@/api/lessons'
import { ApiError } from '@/lib/api'

const route = useRoute()

const course = ref<ApiCourse | null>(null)
const lessons = ref<ApiLesson[]>([])
const loading = ref(true)
const errorMessage = ref('')
const dialogOpen = ref(false)
const editingLesson = ref<ApiLesson | null>(null)
const title = ref('')
const sortOrder = ref(0)
const content = ref('')
const saving = ref(false)
const deletingId = ref('')
const fieldErrors = ref<Record<string, string>>({})

function courseId() {
  return route.params.id as string
}

function sortLessons(items: ApiLesson[]) {
  return [...items].sort((a, b) =>
    a.sort_order - b.sort_order || a.created_at.localeCompare(b.created_at),
  )
}

async function loadPage() {
  loading.value = true
  errorMessage.value = ''

  try {
    const [courseData, lessonData] = await Promise.all([
      getCourseById(courseId()),
      listInstructorLessons(courseId()),
    ])
    course.value = courseData
    lessons.value = sortLessons(lessonData.data ?? [])
  }
  catch (error) {
    course.value = null
    lessons.value = []
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load lessons.'
  }
  finally {
    loading.value = false
  }
}

function openCreateDialog() {
  editingLesson.value = null
  title.value = ''
  sortOrder.value = lessons.value.length
    ? Math.max(...lessons.value.map(lesson => lesson.sort_order)) + 1
    : 0
  content.value = ''
  fieldErrors.value = {}
  errorMessage.value = ''
  dialogOpen.value = true
}

function openEditDialog(lesson: ApiLesson) {
  editingLesson.value = lesson
  title.value = lesson.title
  sortOrder.value = lesson.sort_order
  content.value = lesson.content
  fieldErrors.value = {}
  errorMessage.value = ''
  dialogOpen.value = true
}

async function handleSave() {
  if (saving.value) return
  saving.value = true
  fieldErrors.value = {}
  errorMessage.value = ''

  const payload = {
    title: title.value.trim(),
    sort_order: sortOrder.value,
    content: content.value.trim(),
  }

  try {
    const response = editingLesson.value
      ? await updateLesson(editingLesson.value.id, payload)
      : await createLesson(courseId(), payload)

    if (editingLesson.value) {
      lessons.value = sortLessons(
        lessons.value.map(lesson => lesson.id === response.data.id ? response.data : lesson),
      )
      toast.success('Lesson updated successfully.')
    }
    else {
      lessons.value = sortLessons([...lessons.value, response.data])
      toast.success('Lesson created successfully.')
    }
    dialogOpen.value = false
  }
  catch (error) {
    if (error instanceof ApiError) {
      errorMessage.value = error.message
      fieldErrors.value = error.fieldErrors
    }
    else {
      errorMessage.value = 'Unable to save lesson.'
    }
    toast.error(errorMessage.value)
  }
  finally {
    saving.value = false
  }
}

async function handleDelete(lesson: ApiLesson) {
  if (deletingId.value) return
  deletingId.value = lesson.id

  try {
    await deleteLesson(lesson.id)
    lessons.value = lessons.value.filter(item => item.id !== lesson.id)
    toast.success('Lesson deleted successfully.')
  }
  catch (error) {
    toast.error(error instanceof ApiError ? error.message : 'Unable to delete lesson.')
  }
  finally {
    deletingId.value = ''
  }
}

watch(() => route.params.id, loadPage)
onMounted(loadPage)
</script>

<template>
  <LearningPageShell
    eyebrow="Lessons"
    :title="course?.title ?? (loading ? 'Loading lessons...' : 'Course lessons')"
    description="Create and arrange the learning content for this course."
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
            <BreadcrumbLink :href="`/instructor/courses/${courseId()}`">
              {{ course?.slug ?? 'Course' }}
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Lessons</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <InstructorCourseTabs />

    <p
      v-if="errorMessage && !dialogOpen"
      class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
    >
      {{ errorMessage }}
    </p>

    <Card>
      <CardHeader class="gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <CardTitle>Lessons</CardTitle>
          <CardDescription>
            {{ loading ? 'Loading lessons...' : `${lessons.length} lesson${lessons.length === 1 ? '' : 's'} in learning order` }}
          </CardDescription>
        </div>
        <Button :disabled="loading" @click="openCreateDialog">
          New lesson
        </Button>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-24">Order</TableHead>
              <TableHead>Title</TableHead>
              <TableHead>Content</TableHead>
              <TableHead class="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="loading">
              <TableCell colspan="4" class="text-center text-muted-foreground">
                Loading lessons...
              </TableCell>
            </TableRow>
            <template v-else>
              <TableRow v-for="lesson in lessons" :key="lesson.id">
                <TableCell>{{ lesson.sort_order }}</TableCell>
                <TableCell class="font-medium">{{ lesson.title }}</TableCell>
                <TableCell>
                  <p class="max-w-lg truncate text-sm text-muted-foreground">
                    {{ lesson.content || 'No content yet.' }}
                  </p>
                </TableCell>
                <TableCell>
                  <div class="flex justify-end gap-2">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      @click="openEditDialog(lesson)"
                    >
                      Edit
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
                          <AlertDialogTitle>Delete this lesson?</AlertDialogTitle>
                          <AlertDialogDescription>
                            This permanently deletes “{{ lesson.title }}” and its progress records.
                          </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                          <AlertDialogCancel :disabled="deletingId === lesson.id">
                            Cancel
                          </AlertDialogCancel>
                          <AlertDialogAction
                            class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                            :disabled="deletingId === lesson.id"
                            @click="handleDelete(lesson)"
                          >
                            {{ deletingId === lesson.id ? 'Deleting...' : 'Delete lesson' }}
                          </AlertDialogAction>
                        </AlertDialogFooter>
                      </AlertDialogContent>
                    </AlertDialog>
                  </div>
                </TableCell>
              </TableRow>
              <TableRow v-if="lessons.length === 0">
                <TableCell colspan="4" class="text-center text-muted-foreground">
                  No lessons yet. Create the first lesson to get started.
                </TableCell>
              </TableRow>
            </template>
          </TableBody>
        </Table>
      </CardContent>
    </Card>

    <Dialog v-model:open="dialogOpen">
      <DialogContent class="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{{ editingLesson ? 'Edit lesson' : 'Create lesson' }}</DialogTitle>
          <DialogDescription>
            Set the title, learning order, and content students will read.
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-6" @submit.prevent="handleSave">
          <p
            v-if="errorMessage"
            class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
          >
            {{ errorMessage }}
          </p>

          <div class="space-y-5">
            <FormField label="Title" html-for="lesson-title" :error="fieldErrors.title">
              <Input
                id="lesson-title"
                v-model="title"
                required
                placeholder="Introduction"
              />
            </FormField>

            <FormField
              label="Sort order"
              html-for="lesson-sort-order"
              :error="fieldErrors.sort_order"
              hint="Lower numbers appear first."
            >
              <Input
                id="lesson-sort-order"
                v-model.number="sortOrder"
                type="number"
                min="0"
                required
              />
            </FormField>
          </div>

          <FormField label="Content" html-for="lesson-content" :error="fieldErrors.content">
            <FormTextarea
              id="lesson-content"
              v-model="content"
              :rows="10"
              placeholder="Write the lesson content..."
            />
          </FormField>

          <FormActions>
            <Button
              type="button"
              variant="outline"
              :disabled="saving"
              @click="dialogOpen = false"
            >
              Cancel
            </Button>
            <Button type="submit" :disabled="saving">
              {{ saving ? 'Saving...' : (editingLesson ? 'Save changes' : 'Create lesson') }}
            </Button>
          </FormActions>
        </form>
      </DialogContent>
    </Dialog>
  </LearningPageShell>
</template>
