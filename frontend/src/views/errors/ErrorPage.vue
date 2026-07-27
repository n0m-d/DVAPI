<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { AlertCircle, Home, Lock, Search, ServerCrash } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { useAuth } from '@/composables/useAuth'

type ErrorCode = '401' | '403' | '404' | '500'

const props = defineProps<{
  code?: ErrorCode
  title?: string
  message?: string
}>()

const route = useRoute()
const { homeRouteForRole } = useAuth()

const errorCode = computed<ErrorCode>(() => {
  const metaCode = route.meta.code as ErrorCode | undefined
  return props.code ?? metaCode ?? '404'
})

const defaults: Record<ErrorCode, { title: string, message: string, icon: typeof AlertCircle }> = {
  '401': {
    title: 'Unauthorized',
    message: 'You need to sign in to access this page.',
    icon: Lock,
  },
  '403': {
    title: 'Forbidden',
    message: 'You do not have permission to view this resource.',
    icon: AlertCircle,
  },
  '404': {
    title: 'Page not found',
    message: 'The page you are looking for does not exist or has been moved.',
    icon: Search,
  },
  '500': {
    title: 'Server error',
    message: 'Something went wrong on our end. Please try again later.',
    icon: ServerCrash,
  },
}

const displayTitle = computed(() => props.title ?? defaults[errorCode.value].title)
const displayMessage = computed(() => props.message ?? defaults[errorCode.value].message)
const Icon = computed(() => defaults[errorCode.value].icon)
const homeLink = computed(() => homeRouteForRole())
</script>

<template>
  <div class="flex flex-1 items-center justify-center p-6">
    <Card class="w-full max-w-md text-center">
      <CardHeader class="items-center">
        <div class="mb-2 flex size-14 items-center justify-center rounded-full bg-muted">
          <component
            :is="Icon"
            class="size-7 text-muted-foreground"
          />
        </div>
        <p class="text-sm font-mono text-muted-foreground">
          Error {{ errorCode }}
        </p>
        <CardTitle class="text-2xl">
          {{ displayTitle }}
        </CardTitle>
        <CardDescription class="text-base">
          {{ displayMessage }}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Button as-child>
          <RouterLink :to="homeLink">
            <Home class="mr-2 size-4" />
            Go home
          </RouterLink>
        </Button>
      </CardContent>
    </Card>
  </div>
</template>
