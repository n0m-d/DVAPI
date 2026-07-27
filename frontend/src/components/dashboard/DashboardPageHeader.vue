<script setup lang="ts">
import { Separator } from '@/components/ui/separator'
import { SidebarTrigger } from '@/components/ui/sidebar'
import ModeToggle from '@/components/dashboard/ModeToggle.vue'
import ThemeCustomizer from '@/components/dashboard/ThemeCustomizer.vue'
import UserNav from '@/components/dashboard/UserNav.vue'
</script>

<template>
  <header class="dash-header">
    <SidebarTrigger class="-ml-1 dash-icon-btn" />
    <Separator orientation="vertical" class="mr-2 h-4" />
    <slot name="breadcrumb" />
    <div class="ml-auto flex items-center gap-3 dash-header-actions">
      <ThemeCustomizer />
      <ModeToggle />
      <UserNav />
    </div>
  </header>
</template>
