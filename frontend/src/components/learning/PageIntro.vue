<script setup lang="ts">
defineProps<{
  eyebrow: string
  title: string
  description?: string
}>()
</script>

<template>
  <div class="space-y-1 py-2">
    <p class="dash-eyebrow">
      {{ eyebrow }}
    </p>
    <h2 class="dash-title text-3xl font-bold tracking-tight">
      {{ title }}
    </h2>
    <p
      v-if="description"
      class="dash-subtitle text-sm"
    >
      {{ description }}
    </p>
  </div>
</template>
