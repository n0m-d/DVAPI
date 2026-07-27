<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { cn } from '@/lib/utils'
import { ApiError } from '@/lib/api'
import { requestPasswordReset, verifyPasswordResetOtp } from '@/api/auth'
import { Button } from '@/components/ui/button'
import AuthCard from '@/components/auth/AuthCard.vue'
import FormActions from '@/components/forms/FormActions.vue'
import PinInput from '@/components/auth/PinInput.vue'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()

const route = useRoute()
const router = useRouter()
const pin = ref('')
const loading = ref(false)
const resending = ref(false)
const errorMessage = ref('')

const email = computed(() => {
  const value = route.query.email
  return typeof value === 'string' ? value : ''
})

async function handleSubmit(event: Event) {
  event.preventDefault()
  errorMessage.value = ''

  if (!email.value) {
    await router.replace({ name: 'forgot-password' })
    return
  }

  if (pin.value.length !== 6) {
    errorMessage.value = 'Enter the 6-digit verification code.'
    return
  }

  loading.value = true

  try {
    await verifyPasswordResetOtp({
      email: email.value,
      otp: pin.value,
    })
    toast.success('Code verified')
    await router.push({
      name: 'reset-password',
      query: { email: email.value },
    })
  }
  catch (error) {
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to verify code. Please try again.'
  }
  finally {
    loading.value = false
  }
}

async function handleResend() {
  if (!email.value || resending.value) return

  errorMessage.value = ''
  resending.value = true

  try {
    const response = await requestPasswordReset({ email: email.value })
    if (response.otp) {
      toast.info(`Dev OTP: ${response.otp}`)
    }
    else {
      toast.success(response.message || 'A new code has been sent')
    }
    pin.value = ''
  }
  catch (error) {
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to resend code. Please try again.'
  }
  finally {
    resending.value = false
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
          Code sent to
          <span class="auth-accent font-medium">{{ email }}</span>
        </p>
        <p
          v-else
          class="text-center text-sm text-destructive"
        >
          Missing email. Start again from forgot password.
        </p>

        <div class="flex justify-center py-2">
          <PinInput v-model="pin" />
        </div>

        <FormActions>
          <Button
            type="submit"
            class="auth-button w-full"
            :disabled="loading || pin.length !== 6 || !email"
          >
            {{ loading ? 'Verifying...' : 'Verify code' }}
          </Button>
          <p class="text-center text-sm auth-muted">
            Didn't receive a code?
            <button
              type="button"
              class="auth-link ml-1"
              :disabled="resending || !email"
              @click="handleResend"
            >
              {{ resending ? 'Sending...' : 'Resend' }}
            </button>
          </p>
          <p class="text-center text-sm auth-muted">
            <RouterLink
              :to="{ name: 'forgot-password' }"
              class="auth-link"
            >
              Use a different email
            </RouterLink>
          </p>
        </FormActions>
      </form>
    </AuthCard>
  </div>
</template>
