<script setup lang="ts">
import AuthPageLayout from '@/layouts/AuthPageLayout.vue'
import { Button } from '@/components/ui/button'
import AuthCard from '@/components/auth/AuthCard.vue'
import { Mail } from 'lucide-vue-next'

function handleResend() {
  // Mock resend
}
</script>

<template>
  <AuthPageLayout
    title="Verify your email"
    description="We sent a confirmation link to your inbox"
    :show-terms="false"
  >
    <AuthCard>
      <div class="relative z-10 space-y-6 text-center">
        <div class="mx-auto flex size-14 items-center justify-center rounded-full bg-primary/10">
          <Mail class="size-7 text-primary" />
        </div>
        <div class="space-y-2">
          <p class="text-sm auth-muted">
            Check your email and click the verification link to activate your
            account. The link expires in 24 hours.
          </p>
          <p class="text-xs auth-muted">
            Did not receive it? Check your spam folder or request a new link.
          </p>
        </div>
        <Button
          class="auth-button w-full"
          @click="handleResend"
        >
          Resend verification email
        </Button>
        <p class="text-center text-sm auth-muted">
          <RouterLink
            :to="{ name: 'login' }"
            class="auth-link"
          >
            Back to login
          </RouterLink>
        </p>
      </div>
    </AuthCard>
  </AuthPageLayout>
</template>
