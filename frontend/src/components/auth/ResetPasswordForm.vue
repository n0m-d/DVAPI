<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { cn } from '@/lib/utils'
import { ApiError } from '@/lib/api'
import { confirmPasswordReset } from '@/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import AuthCard from '@/components/auth/AuthCard.vue'
import FormActions from '@/components/forms/FormActions.vue'
import FormField from '@/components/forms/FormField.vue'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()

const route = useRoute()
const router = useRouter()

const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const errorMessage = ref('')
const fieldErrors = reactive<Record<string, string>>({})

const email = computed(() => {
  const value = route.query.email
  return typeof value === 'string' ? value.trim() : ''
})

function clearErrors() {
  errorMessage.value = ''
  for (const key of Object.keys(fieldErrors)) {
    delete fieldErrors[key]
  }
}

onMounted(() => {
  if (!email.value) {
    router.replace({ name: 'forgot-password' })
  }
})

async function handleSubmit(event: Event) {
  event.preventDefault()
  clearErrors()

  if (!email.value) {
    await router.replace({ name: 'forgot-password' })
    return
  }

  if (password.value !== confirmPassword.value) {
    fieldErrors.confirm_password = 'Passwords do not match.'
    return
  }

  loading.value = true

  try {
    await confirmPasswordReset({
      email: email.value,
      password: password.value,
      confirm_password: confirmPassword.value,
    })
    toast.success('Password reset successfully. Please sign in.')
    await router.push({ name: 'login' })
  }
  catch (error) {
    if (error instanceof ApiError) {
      const errors = error.fieldErrors
      if (Object.keys(errors).length > 0) {
        Object.assign(fieldErrors, errors)
      }
      else {
        errorMessage.value = error.message
      }
    }
    else {
      errorMessage.value = 'Unable to reset password. Please try again.'
    }
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

        <p
          v-if="email"
          class="text-center text-sm auth-muted"
        >
          Resetting password for
          <span class="auth-accent font-medium">{{ email }}</span>
        </p>

        <FormField
          label="New password"
          html-for="password"
          :error="fieldErrors.password"
        >
          <Input
            id="password"
            v-model="password"
            type="password"
            autocomplete="new-password"
            minlength="8"
            required
            @input="delete fieldErrors.password"
          />
        </FormField>

        <FormField
          label="Confirm password"
          html-for="confirm-password"
          :error="fieldErrors.confirm_password"
        >
          <Input
            id="confirm-password"
            v-model="confirmPassword"
            type="password"
            autocomplete="new-password"
            minlength="8"
            required
            @input="delete fieldErrors.confirm_password"
          />
        </FormField>

        <FormActions>
          <Button
            type="submit"
            class="auth-button w-full"
            :disabled="loading || !email"
          >
            {{ loading ? 'Resetting...' : 'Reset password' }}
          </Button>
          <p class="text-center text-sm auth-muted">
            <RouterLink
              :to="{ name: 'login' }"
              class="auth-link"
            >
              Back to login
            </RouterLink>
          </p>
        </FormActions>
      </form>
    </AuthCard>
  </div>
</template>
