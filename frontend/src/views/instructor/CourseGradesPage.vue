<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
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
import InstructorCourseTabs from '@/components/learning/InstructorCourseTabs.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { getCourseById, type ApiCourse } from '@/api/courses'
import { ApiError } from '@/lib/api'

const route = useRoute()
const courseId = computed(() => route.params.id as string)
const course = ref<ApiCourse | null>(null)
const loading = ref(true)
const errorMessage = ref('')

async function loadCourse() {
  loading.value = true
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
  finally {
    loading.value = false
  }
}

onMounted(loadCourse)
</script>

<template>
  <LearningPageShell
    eyebrow="Grades"
    title="Grade entry"
    description="Enter and update scores for assignments in this course."
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
            <BreadcrumbPage>Grades</BreadcrumbPage>
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

    <Card>
      <CardHeader class="flex flex-row items-center justify-between">
        <div>
          <CardTitle>Gradebook</CardTitle>
          <CardDescription>
            {{ loading ? 'Loading...' : '0 graded entries' }}
          </CardDescription>
        </div>
        <Button size="sm" disabled>
          Save all
        </Button>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Student</TableHead>
              <TableHead>Assignment</TableHead>
              <TableHead>Score</TableHead>
              <TableHead>Letter</TableHead>
              <TableHead class="hidden md:table-cell">Submitted</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell
                colspan="5"
                class="text-center text-muted-foreground"
              >
                {{ loading ? 'Loading grades...' : 'No grades recorded for this course yet.' }}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </LearningPageShell>
</template>
