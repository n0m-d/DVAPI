<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  BookOpen,
  ClipboardList,
  GraduationCap,
  Send,
  UserCog,
  Users,
  UserRoundCheck,
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
import { getAdminStats, type AdminStats } from '@/api/admin'
import { ApiError } from '@/lib/api'

const stats = ref<AdminStats | null>(null)
const loading = ref(true)
const errorMessage = ref('')

async function fetchStats() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await getAdminStats()
    stats.value = response.data
  }
  catch (error) {
    stats.value = null
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load platform statistics.'
  }
  finally {
    loading.value = false
  }
}

onMounted(fetchStats)
</script>

<template>
  <LearningPageShell
    eyebrow="Admin"
    title="Dashboard"
    description="Platform activity and account totals at a glance."
  >
    <template #breadcrumb>
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem class="hidden md:block">
            <BreadcrumbLink href="/admin/dashboard">
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
      Loading platform statistics...
    </p>

    <div
      v-else-if="stats"
      class="grid gap-4 md:grid-cols-2 lg:grid-cols-4"
    >
      <StatCard title="Users" :value="stats.users" delta="All accounts" :icon="Users" />
      <StatCard title="Students" :value="stats.students" delta="Student accounts" :icon="GraduationCap" />
      <StatCard title="Instructors" :value="stats.instructors" delta="Instructor accounts" :icon="UserRoundCheck" />
      <StatCard title="Courses" :value="stats.courses" delta="Platform courses" :icon="BookOpen" />
      <StatCard title="Enrollments" :value="stats.enrollments" delta="Course enrollments" :icon="UserCog" />
      <StatCard title="Assignments" :value="stats.assignments" delta="Course assignments" :icon="ClipboardList" />
      <StatCard title="Submissions" :value="stats.submissions" delta="Student submissions" :icon="Send" />
    </div>
  </LearningPageShell>
</template>
