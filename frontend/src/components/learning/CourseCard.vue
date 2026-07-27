<script setup lang="ts">
import { RouterLink } from 'vue-router'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import StatusBadge from '@/components/learning/StatusBadge.vue'
import { truncateText } from '@/lib/utils'

export type CourseCardStatus = 'active' | 'draft' | 'archived'

export interface CourseCardItem {
  id: string
  code: string
  title: string
  instructor: string
  status: CourseCardStatus
  term: string
  description?: string
  enrolled?: boolean
}

defineProps<{
  course: CourseCardItem
  showEnroll?: boolean
  detailLink?: string
}>()
</script>

<template>
  <Card>
    <CardHeader class="pb-3">
      <div class="flex items-start justify-between gap-2">
        <div>
          <p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">
            {{ course.code }}
          </p>
          <CardTitle class="text-base">{{ course.title }}</CardTitle>
          <CardDescription>{{ course.instructor }} · {{ course.term }}</CardDescription>
        </div>
        <StatusBadge :status="course.enrolled ? 'enrolled' : course.status" />
      </div>
    </CardHeader>
    <CardContent class="space-y-3">
      <p
        v-if="course.description"
        class="text-sm text-muted-foreground"
      >
        {{ truncateText(course.description, 50) }}
      </p>
      <div class="flex gap-2">
        <Button
          v-if="detailLink"
          as-child
          size="sm"
          variant="outline"
          class="flex-1"
        >
          <RouterLink :to="detailLink">
            View course
          </RouterLink>
        </Button>
        <Button
          v-if="showEnroll && !course.enrolled"
          size="sm"
          class="flex-1"
        >
          Enroll
        </Button>
      </div>
    </CardContent>
  </Card>
</template>
