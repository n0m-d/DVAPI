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
    eyebrow="Submissions"
    title="Course submissions"
    description="Review and grade student work for this course."
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
            <BreadcrumbPage>Submissions</BreadcrumbPage>
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
      <CardHeader>
        <CardTitle>Submissions</CardTitle>
        <CardDescription>
          {{ loading ? 'Loading...' : '0 pending review' }}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Student</TableHead>
              <TableHead>Assignment</TableHead>
              <TableHead>Submitted</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Score</TableHead>
              <TableHead class="text-right">Action</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell
                colspan="6"
                class="text-center text-muted-foreground"
              >
                {{ loading ? 'Loading submissions...' : 'No submissions for this course yet.' }}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </LearningPageShell>
</template>
