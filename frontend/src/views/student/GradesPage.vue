<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
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
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { getOwnGrades, type ApiGrade } from '@/api/grades'
import { ApiError } from '@/lib/api'

const grades = ref<ApiGrade[]>([])
const submitted = ref(0)
const averageGrade = ref(0)
const loading = ref(true)
const errorMessage = ref('')

const gradedCount = computed(() => grades.value.filter(item => item.grade !== null).length)
const pendingCount = computed(() => grades.value.filter(item => item.grade === null).length)

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString()
}

async function loadGrades() {
  loading.value = true
  errorMessage.value = ''
  try {
    const response = await getOwnGrades()
    grades.value = response.data.grades
    submitted.value = response.data.submitted
    averageGrade.value = response.data.average_grade
  }
  catch (error) {
    grades.value = []
    submitted.value = 0
    averageGrade.value = 0
    errorMessage.value = error instanceof ApiError ? error.message : 'Unable to load grades.'
  }
  finally {
    loading.value = false
  }
}

onMounted(loadGrades)
</script>

<template>
  <LearningPageShell
    eyebrow="Academic records"
    title="My grades"
    description="View scored assignments and letter grades for your enrolled courses."
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
            <BreadcrumbPage>Grades</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <p
      v-if="errorMessage"
      class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
    >
      {{ errorMessage }}
    </p>

    <div class="grid gap-4 sm:grid-cols-3">
      <Card>
        <CardHeader>
          <CardDescription>Submitted</CardDescription>
          <CardTitle>{{ submitted }}</CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader>
          <CardDescription>Graded</CardDescription>
          <CardTitle>{{ gradedCount }}</CardTitle>
        </CardHeader>
      </Card>
      <Card>
        <CardHeader>
          <CardDescription>Average grade</CardDescription>
          <CardTitle>{{ gradedCount ? `${averageGrade.toFixed(1)}%` : '—' }}</CardTitle>
        </CardHeader>
      </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>Gradebook</CardTitle>
        <CardDescription>
          {{ loading ? 'Loading grades...' : `${gradedCount} graded · ${pendingCount} pending` }}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Course</TableHead>
              <TableHead>Assignment</TableHead>
              <TableHead>Score</TableHead>
              <TableHead>Feedback</TableHead>
              <TableHead class="hidden md:table-cell">Submitted</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="loading">
              <TableCell colspan="5" class="text-center text-muted-foreground">
                Loading grades...
              </TableCell>
            </TableRow>
            <template v-else>
              <TableRow
                v-for="grade in grades"
                :key="grade.submission_id"
              >
                <TableCell class="font-medium">{{ grade.course_title }}</TableCell>
                <TableCell class="font-medium">{{ grade.assignment_title }}</TableCell>
                <TableCell>
                  <span v-if="grade.grade !== null">{{ grade.grade }}</span>
                  <span v-else class="text-muted-foreground">Pending</span>
                </TableCell>
                <TableCell>{{ grade.grade === null ? 'Awaiting review' : (grade.feedback || '—') }}</TableCell>
                <TableCell class="hidden md:table-cell">
                  {{ formatDate(grade.submitted_at) }}
                </TableCell>
              </TableRow>
              <TableRow v-if="!grades.length">
                <TableCell
                  colspan="5"
                  class="text-center text-muted-foreground"
                >
                  No submissions found.
                </TableCell>
              </TableRow>
            </template>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </LearningPageShell>
</template>
