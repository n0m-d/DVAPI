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
import StatusBadge from '@/components/learning/StatusBadge.vue'
import { getOwnGrades, type ApiGrade } from '@/api/grades'
import { ApiError } from '@/lib/api'

const submissions = ref<ApiGrade[]>([])
const submitted = ref(0)
const loading = ref(true)
const errorMessage = ref('')

const pendingCount = computed(() => submissions.value.filter(item => item.grade === null).length)

function formatDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString()
}

async function loadSubmissions() {
  loading.value = true
  errorMessage.value = ''
  try {
    const response = await getOwnGrades()
    submissions.value = response.data.grades
    submitted.value = response.data.submitted
  }
  catch (error) {
    submissions.value = []
    submitted.value = 0
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load submissions.'
  }
  finally {
    loading.value = false
  }
}

onMounted(loadSubmissions)
</script>

<template>
  <LearningPageShell
    eyebrow="My work"
    title="Submissions"
    description="Track the status of work you have submitted across all courses."
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
            <BreadcrumbPage>Submissions</BreadcrumbPage>
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

    <Card>
      <CardHeader>
        <CardTitle>My submissions</CardTitle>
        <CardDescription>
          {{ loading ? 'Loading submissions...' : `${submitted} total · ${pendingCount} pending review` }}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Course</TableHead>
              <TableHead>Assignment</TableHead>
              <TableHead>Submitted</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Score</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="loading">
              <TableCell colspan="5" class="text-center text-muted-foreground">
                Loading submissions...
              </TableCell>
            </TableRow>
            <template v-else>
              <TableRow
                v-for="submission in submissions"
                :key="submission.submission_id"
              >
                <TableCell class="font-medium">{{ submission.course_title }}</TableCell>
                <TableCell class="font-medium">{{ submission.assignment_title }}</TableCell>
                <TableCell>{{ formatDate(submission.submitted_at) }}</TableCell>
                <TableCell>
                  <StatusBadge :status="submission.grade === null ? 'pending' : 'graded'" />
                </TableCell>
                <TableCell>
                  <span v-if="submission.grade !== null">{{ submission.grade }}</span>
                  <span v-else class="text-muted-foreground">—</span>
                </TableCell>
              </TableRow>
              <TableRow v-if="!submissions.length">
                <TableCell
                  colspan="5"
                  class="text-center text-muted-foreground"
                >
                  No submissions yet.
                </TableCell>
              </TableRow>
            </template>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </LearningPageShell>
</template>
