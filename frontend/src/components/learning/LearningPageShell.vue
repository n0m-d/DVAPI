<script setup lang="ts">
import DashboardPageHeader from '@/components/dashboard/DashboardPageHeader.vue'
import PageIntro from '@/components/learning/PageIntro.vue'

defineProps<{
  eyebrow: string
  title: string
  description?: string
}>()
</script>

<template>
  <DashboardPageHeader>
    <template #breadcrumb>
      <slot name="breadcrumb" />
    </template>
  </DashboardPageHeader>
  <div class="dash-content flex flex-1 flex-col gap-6">
    <PageIntro
      :eyebrow="eyebrow"
      :title="title"
      :description="description"
    />
    <slot />
  </div>
</template>
