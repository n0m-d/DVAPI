<script setup lang="ts">
import { watch } from 'vue'
import { Toaster } from '@/components/ui/sonner'
import TopLoader from '@/components/ui/top-loader/TopLoader.vue'
import { useAppColorMode } from '@/composables/useAppColorMode'
import { useThemeColor } from '@/composables/useThemeColor'
import { useAuthStore } from '@/stores/auth'
import { useNotificationStore } from '@/stores/notifications'

useAppColorMode()
useThemeColor()

const auth = useAuthStore()
const notifications = useNotificationStore()

watch(
  () => auth.token, // watch the token in the auth store
  (token) => {
    if (token) {
      notifications.connect() // connect to the notifications service
    }
    else {
      notifications.disconnect()
      notifications.clear()
    }
  },
  { immediate: true },
)
</script>

<template>
  <TopLoader />
  <router-view />
  <Toaster position="top-right" />
</template>
