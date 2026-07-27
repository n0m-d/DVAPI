<script setup lang="ts">
import { onMounted, ref } from 'vue'
import {
  BookOpen,
  ClipboardList,
  GraduationCap,
  Megaphone,
  Send,
  Users,
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
import { getDashboardStats, type InstructorStats } from '@/api/stats'
import { ApiError } from '@/lib/api'

const stats = ref<InstructorStats | null>(null)
const loading = ref(true)
const errorMessage = ref('')

async function fetchStats() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await getDashboardStats()
    if (response.role !== 'instructor') {
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
        : 'Unable to load analytics.'
  }
  finally {
    loading.value = false
  }
}

onMounted(fetchStats)
</script>

<template>
  <LearningPageShell
    eyebrow="Analytics"
    title="Course analytics"
    description="Engagement and performance metrics across your courses."
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
            <BreadcrumbPage>Analytics</BreadcrumbPage>
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
      Loading analytics...
    </p>

    <div
      v-else-if="stats"
      class="grid gap-4 md:grid-cols-2 lg:grid-cols-4"
    >
      <StatCard
        title="Courses"
        :value="stats.courses"
        :delta="`${stats.published_courses} published`"
        :icon="BookOpen"
      />
      <StatCard
        title="Students"
        :value="stats.students"
        :delta="`${stats.enrollments} enrollments`"
        :icon="Users"
      />
      <StatCard
        title="Assignments"
        :value="stats.assignments"
        :delta="`${stats.lessons} lessons`"
        :icon="ClipboardList"
      />
      <StatCard
        title="Ungraded"
        :value="stats.ungraded_submissions"
        :delta="`${stats.submissions} total submissions`"
        :icon="Send"
      />
      <StatCard
        title="Announcements"
        :value="stats.announcements"
        delta="Course updates posted"
        :icon="Megaphone"
      />
      <StatCard
        title="Published courses"
        :value="stats.published_courses"
        delta="Visible to students"
        :icon="GraduationCap"
      />
    </div>
  </LearningPageShell>
</template>
