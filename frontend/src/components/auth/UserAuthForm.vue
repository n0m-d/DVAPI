<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { cn } from '@/lib/utils'
import { ApiError } from '@/lib/api'
import { canRoleAccessRoute } from '@/lib/routeAccess'
import { useAuth } from '@/composables/useAuth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import AuthCard from '@/components/auth/AuthCard.vue'
import FormActions from '@/components/forms/FormActions.vue'
import FormField from '@/components/forms/FormField.vue'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()

const router = useRouter()
const route = useRoute()
const { login, homeRouteForRole, role } = useAuth()

const email = ref('')
const password = ref('')
const loading = ref(false)
const errorMessage = ref('')

async function handleSubmit(event: Event) {
  event.preventDefault()
  errorMessage.value = ''
  loading.value = true

  try {
    await login({
      email: email.value.trim(),
      password: password.value,
    })
    toast.success('Signed in')

    const requested = typeof route.query.redirect === 'string'
      ? route.query.redirect
      : null

    const home = homeRouteForRole(role.value)
    if (
      requested
      && requested.startsWith('/')
      && !requested.startsWith('//')
      && canRoleAccessRoute(role.value, router.resolve(requested))
    ) {
      await router.push(requested)
    }
    else {
      await router.push(home)
    }
  }
  catch (error) {
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to sign in. Please try again.'
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <div :class="cn('flex flex-col', props.class)">
    <AuthCard>
      <form
        class="app-form relative z-10"
        @submit="handleSubmit"
      >
        <p
          v-if="errorMessage"
          class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
        >
          {{ errorMessage }}
        </p>

        <FormField
          label="Email"
          html-for="email"
        >
          <Input
            id="email"
            v-model="email"
            type="email"
            placeholder="m@example.com"
            autocomplete="email"
            required
          />
        </FormField>

        <FormField
          label="Password"
          html-for="password"
        >
          <template #action>
            <RouterLink
              :to="{ name: 'forgot-password' }"
              class="auth-link text-xs"
            >
              Forgot password?
            </RouterLink>
          </template>
          <Input
            id="password"
            v-model="password"
            type="password"
            autocomplete="current-password"
            required
          />
        </FormField>

        <FormActions>
          <Button
            type="submit"
            class="auth-button w-full"
            :disabled="loading"
          >
            {{ loading ? 'Signing in...' : 'Sign in' }}
          </Button>
          <p class="text-center text-sm auth-muted">
            Don't have an account?
            <RouterLink
              :to="{ name: 'register' }"
              class="auth-link ml-1"
            >
              Create one
            </RouterLink>
          </p>
        </FormActions>
      </form>
    </AuthCard>
  </div>
</template>
