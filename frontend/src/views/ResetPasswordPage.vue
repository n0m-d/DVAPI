<script setup lang="ts">
import AuthPageLayout from '@/layouts/AuthPageLayout.vue'
import ResetPasswordForm from '@/components/auth/ResetPasswordForm.vue'
</script>

<template>
  <AuthPageLayout
    title="Reset your password"
    description="Choose a new password for your account"
    :show-terms="false"
  >
    <ResetPasswordForm />
  </AuthPageLayout>
</template>
