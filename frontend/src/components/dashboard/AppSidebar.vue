<script setup lang="ts">
import type { SidebarProps } from '@/components/ui/sidebar'
import { School } from 'lucide-vue-next'
import NavMain from '@/components/NavMain.vue'
import NavUser from '@/components/NavUser.vue'
import TeamSwitcher from '@/components/TeamSwitcher.vue'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
import { useAuth } from '@/composables/useAuth'
import { navForRole, roleLabels } from '@/config/navigation'

const props = withDefaults(defineProps<SidebarProps>(), {
  collapsible: 'icon',
})

const { role, user } = useAuth()

const teams = [
  {
    name: 'Schole',
    logo: School,
    plan: roleLabels[role.value],
  },
]
</script>

<template>
  <Sidebar
    v-bind="props"
    class="dash-sidebar transition-all duration-150 ease-in-out"
  >
    <SidebarHeader class="transition-all duration-100 ease-in-out">
      <TeamSwitcher :teams="teams" />
    </SidebarHeader>
    <SidebarContent class="transition-opacity duration-100 ease-in-out">
      <NavMain :items="navForRole(role)" />
    </SidebarContent>
    <SidebarFooter class="transition-all duration-100 ease-in-out">
      <NavUser :user="user" />
    </SidebarFooter>
    <SidebarRail class="transition-all duration-75 ease-in-out" />
  </Sidebar>
</template>
