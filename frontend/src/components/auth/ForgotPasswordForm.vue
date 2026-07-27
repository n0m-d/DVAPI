<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { cn } from '@/lib/utils'
import { ApiError } from '@/lib/api'
import { requestPasswordReset } from '@/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import AuthCard from '@/components/auth/AuthCard.vue'
import FormActions from '@/components/forms/FormActions.vue'
import FormField from '@/components/forms/FormField.vue'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()

const router = useRouter()
const email = ref('')
const loading = ref(false)
const errorMessage = ref('')
const fieldErrors = reactive<Record<string, string>>({})

function clearErrors() {
  errorMessage.value = ''
  for (const key of Object.keys(fieldErrors)) {
    delete fieldErrors[key]
  }
}

async function handleSubmit(event: Event) {
  event.preventDefault()
  clearErrors()
  loading.value = true

  try {
    const response = await requestPasswordReset({
      email: email.value.trim(),
    })

    if (response.otp) {
      toast.info(`Dev OTP: ${response.otp}`)
    }
    else {
      toast.success(response.message || 'Verification code sent')
    }

    await router.push({
      name: 'forgot-password-verify',
      query: { email: email.value.trim() },
    })
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
      errorMessage.value = 'Unable to send verification code. Please try again.'
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
          label="Email address"
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
            @input="delete fieldErrors.email"
          />
        </FormField>

        <FormActions>
          <Button
            type="submit"
            class="auth-button w-full"
            :disabled="loading"
          >
            {{ loading ? 'Sending...' : 'Send verification code' }}
          </Button>
          <p class="text-center text-sm auth-muted">
            Remember your password?
            <RouterLink
              :to="{ name: 'login' }"
              class="auth-link ml-1"
            >
              Back to login
            </RouterLink>
          </p>
        </FormActions>
      </form>
    </AuthCard>
  </div>
</template>
