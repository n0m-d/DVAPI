<script setup lang="ts">
import { type LucideIcon } from 'lucide-vue-next'
import { RouterLink, useRoute } from 'vue-router'
import {
  SidebarGroup,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'
import { useAuth } from '@/composables/useAuth'
import { roleLabels } from '@/config/navigation'

const props = defineProps<{
  items: {
    title: string
    url: string
    icon: LucideIcon
  }[]
}>()

const route = useRoute()
const { role } = useAuth()

const isRouterLink = (url: string) => url?.startsWith('/') && !url.startsWith('//')

const isActive = (url: string) => {
  if (!isRouterLink(url)) return false
  const currentPath = route.path
  const itemPath = url
  
  if (itemPath === '/' && currentPath === '/') return true
  
  if (itemPath !== '/' && currentPath.startsWith(itemPath)) return true
  
  return currentPath === itemPath
}
</script>

<template>
  <SidebarGroup class="transition-all duration-100 ease-in-out">
    <SidebarGroupLabel class="transition-colors duration-75 ease-in-out">{{ roleLabels[role] }} portal</SidebarGroupLabel>
    <SidebarMenu class="transition-all duration-100 ease-in-out">
      <SidebarMenuItem 
        v-for="item in items" 
        :key="item.title"
        class="transition-all duration-75 ease-in-out"
      >
        <SidebarMenuButton 
          :tooltip="item.title" 
          :isActive="isActive(item.url)" 
          as-child
          class="transition-all duration-75 ease-in-out hover:scale-105 active:scale-95 cursor-pointer"
        >
          <component
            :is="isRouterLink(item.url) ? RouterLink : 'a'"
            :to="isRouterLink(item.url) ? item.url : undefined"
            :href="!isRouterLink(item.url) ? item.url : undefined"
            class="transition-all duration-75 ease-in-out"
          >
            <component 
              :is="item.icon" 
              class="transition-transform duration-75 ease-in-out group-hover:scale-110"
            />
            <span class="transition-colors duration-75 ease-in-out">{{ item.title }}</span>
          </component>
        </SidebarMenuButton>
      </SidebarMenuItem>
    </SidebarMenu>
  </SidebarGroup>
</template>
