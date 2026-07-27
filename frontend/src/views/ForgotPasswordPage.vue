<script setup lang="ts">
import AuthPageLayout from '@/layouts/AuthPageLayout.vue'
import ForgotPasswordForm from '@/components/auth/ForgotPasswordForm.vue'
</script>

<template>
  <AuthPageLayout
    title="Forgot your password?"
    description="Enter your email and we'll send you a verification code"
    :show-terms="false"
  >
    <ForgotPasswordForm />
  </AuthPageLayout>
</template>
