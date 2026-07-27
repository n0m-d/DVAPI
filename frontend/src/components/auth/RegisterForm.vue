<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { cn } from '@/lib/utils'
import { ApiError } from '@/lib/api'
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
const { register } = useAuth()

const fullName = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const errorMessage = ref('')
const fieldErrors = reactive<Record<string, string>>({})

function clearErrors() {
  errorMessage.value = ''
  for (const key of Object.keys(fieldErrors)) {
    delete fieldErrors[key]
  }
}

function applyFieldErrors(errors: Record<string, string>) {
  clearErrors()
  Object.assign(fieldErrors, errors)
}

function clearFieldError(field: string) {
  delete fieldErrors[field]
}

async function handleSubmit(event: Event) {
  event.preventDefault()
  clearErrors()

  if (password.value !== confirmPassword.value) {
    fieldErrors.confirm_password = 'Passwords do not match.'
    return
  }

  loading.value = true

  try {
    await register({
      email: email.value.trim(),
      full_name: fullName.value.trim(),
      password: password.value,
      confirm_password: confirmPassword.value,
    })
    toast.success('Account created. Please sign in.')
    await router.push({ name: 'login' })
  }
  catch (error) {
    if (error instanceof ApiError) {
      const errors = error.fieldErrors
      if (Object.keys(errors).length > 0) {
        applyFieldErrors(errors)
      }
      else {
        errorMessage.value = error.message
      }
    }
    else {
      errorMessage.value = 'Unable to create account. Please try again.'
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

        <FormField
          label="Full name"
          html-for="name"
          :error="fieldErrors.full_name"
        >
          <Input
            id="name"
            v-model="fullName"
            type="text"
            placeholder="John Doe"
            autocomplete="name"
            required
            @input="clearFieldError('full_name')"
          />
        </FormField>

        <FormField
          label="Email"
          html-for="email"
          :error="fieldErrors.email"
        >
          <Input
            id="email"
            v-model="email"
            type="email"
            placeholder="m@example.com"
            autocomplete="email"
            required
            @input="clearFieldError('email')"
          />
        </FormField>

        <FormField
          label="Password"
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
            @input="clearFieldError('password')"
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
            @input="clearFieldError('confirm_password')"
          />
        </FormField>

        <FormActions>
          <Button
            type="submit"
            class="auth-button w-full"
            :disabled="loading"
          >
            {{ loading ? 'Creating account...' : 'Create account' }}
          </Button>
          <p class="text-center text-sm auth-muted">
            Already have an account?
            <RouterLink
              :to="{ name: 'login' }"
              class="auth-link ml-1"
            >
              Sign in
            </RouterLink>
          </p>
        </FormActions>
      </form>
    </AuthCard>
  </div>
</template>
