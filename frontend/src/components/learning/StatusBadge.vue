<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  status: string
  class?: string
}>()

const styles: Record<string, string> = {
  active: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
  inactive: 'border-muted-foreground/20 bg-muted text-muted-foreground',
  draft: 'border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400',
  published: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
  closed: 'border-muted-foreground/20 bg-muted text-muted-foreground',
  archived: 'border-muted-foreground/20 bg-muted text-muted-foreground',
  pending: 'border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400',
  graded: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
  late: 'border-red-500/30 bg-red-500/10 text-red-600 dark:text-red-400',
  returned: 'border-blue-500/30 bg-blue-500/10 text-blue-600 dark:text-blue-400',
  enrolled: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
}

const label = computed(() => props.status.charAt(0).toUpperCase() + props.status.slice(1))
</script>

<template>
  <span
    :class="cn(
      'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium capitalize',
      styles[status] ?? 'border-border bg-muted text-muted-foreground',
      props.class,
    )"
  >
    {{ label }}
  </span>
</template>
