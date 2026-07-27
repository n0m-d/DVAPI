<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { instructorCourseNav } from '@/config/navigation'
import { getCourseById, type ApiCourse } from '@/api/courses'
import { cn } from '@/lib/utils'

const route = useRoute()
const courseId = ref(route.params.id as string)
const course = ref<ApiCourse | null>(null)
const tabs = ref(instructorCourseNav(courseId.value))

const activeTabUrl = computed(() => {
  const exact = tabs.value.find(tab => tab.url === route.path)
  if (exact) return exact.url

  return tabs.value
    .filter(tab => route.path.startsWith(`${tab.url}/`))
    .sort((a, b) => b.url.length - a.url.length)[0]?.url
})

function isActive(url: string) {
  return activeTabUrl.value === url
}

watch(
  () => route.params.id as string,
  async (id) => {
    courseId.value = id
    tabs.value = instructorCourseNav(id)
    try {
      course.value = await getCourseById(id)
    }
    catch {
      course.value = null
    }
  },
  { immediate: true },
)
</script>

<template>
  <div class="space-y-4">
    <div
      v-if="course"
      class="flex flex-wrap items-center justify-between gap-2"
    >
      <div>
        <p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">
          {{ course.slug }}
        </p>
        <h3 class="text-lg font-semibold">
          {{ course.title }}
        </h3>
      </div>
      <RouterLink
        :to="`/instructor/courses/${courseId}/assignments/new`"
        class="dash-link text-sm hover:underline"
      >
        + New assignment
      </RouterLink>
    </div>
    <nav class="course-tabs flex flex-wrap gap-1 rounded-lg border border-border/60 bg-muted/30 p-1">
      <RouterLink
        v-for="tab in tabs"
        :key="tab.url"
        :to="tab.url"
        :class="cn(
          'course-tab rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
          isActive(tab.url)
            ? 'bg-primary !text-primary-foreground shadow-sm'
            : '!text-muted-foreground hover:bg-muted/80 hover:!text-foreground',
        )"
      >
        {{ tab.title }}
      </RouterLink>
    </nav>
  </div>
</template>
