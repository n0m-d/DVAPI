<script setup lang="ts">
import {
  SidebarInset,
  SidebarProvider,
} from '@/components/ui/sidebar'
import AppSidebar from '@/components/dashboard/AppSidebar.vue'
</script>

<template>
  <SidebarProvider data-dashboard class="dash-shell min-h-svh">
    <div class="dash-bg" aria-hidden="true">
      <div class="dash-orb dash-orb-1" />
      <div class="dash-orb dash-orb-2" />
      <div class="dash-grid" />
    </div>

    <AppSidebar />
    <SidebarInset class="dash-main">
      <router-view />
    </SidebarInset>
  </SidebarProvider>
</template>
