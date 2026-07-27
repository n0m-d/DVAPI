<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  BookOpen,
  CheckCircle2,
  ClipboardList,
  GraduationCap,
  Send,
} from 'lucide-vue-next'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { Button } from '@/components/ui/button'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import StatCard from '@/components/learning/StatCard.vue'
import { getDashboardStats, type StudentStats } from '@/api/stats'
import { ApiError } from '@/lib/api'
import { useAuth } from '@/composables/useAuth'

const { user } = useAuth()

const stats = ref<StudentStats | null>(null)
const loading = ref(true)
const errorMessage = ref('')

function formatAverageGrade(value: number) {
  if (!Number.isFinite(value)) return '—'
  return Number.isInteger(value) ? String(value) : value.toFixed(1)
}

async function fetchStats() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await getDashboardStats()
    if (response.role !== 'student') {
      throw new Error('Unexpected stats response for this account.')
    }
    stats.value = response.data
  }
  catch (error) {
    stats.value = null
    errorMessage.value = error instanceof ApiError
      ? error.message
      : error instanceof Error
        ? error.message
        : 'Unable to load dashboard statistics.'
  }
  finally {
    loading.value = false
  }
}

onMounted(fetchStats)
</script>

<template>
  <LearningPageShell
    eyebrow="Student"
    title="Dashboard"
    :description="`Welcome back${user.name ? `, ${user.name}` : ''}.`"
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
            <BreadcrumbPage>Dashboard</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <div
      v-if="errorMessage"
      class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-3 text-sm text-destructive"
    >
      <div class="flex items-center justify-between gap-3">
        <span>{{ errorMessage }}</span>
        <Button variant="outline" size="sm" @click="fetchStats">
          Retry
        </Button>
      </div>
    </div>

    <p
      v-else-if="loading"
      class="text-sm text-muted-foreground"
    >
      Loading dashboard statistics...
    </p>

    <div
      v-else-if="stats"
      class="grid gap-4 md:grid-cols-2 lg:grid-cols-3"
    >
      <StatCard
        title="Enrolled courses"
        :value="stats.enrolled_courses"
        delta="Active enrollments"
        :icon="BookOpen"
      />
      <StatCard
        title="Pending assignments"
        :value="stats.pending_assignments"
        delta="Still due"
        :icon="ClipboardList"
      />
      <StatCard
        title="Average grade"
        :value="formatAverageGrade(stats.average_grade)"
        delta="Across graded work"
        :icon="GraduationCap"
      />
      <StatCard
        title="Submissions"
        :value="stats.submissions"
        delta="All submitted work"
        :icon="Send"
      />
      <StatCard
        title="Graded submissions"
        :value="stats.graded_submissions"
        delta="Returned with a grade"
        :icon="CheckCircle2"
      />
      <StatCard
        title="Completed lessons"
        :value="stats.completed_lessons"
        delta="Lessons finished"
        :icon="BookOpen"
      />
    </div>
  </LearningPageShell>
</template>
