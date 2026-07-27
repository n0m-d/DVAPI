<script setup lang="ts">
import AuthPageLayout from '@/layouts/AuthPageLayout.vue'
import OtpVerifyForm from '@/components/auth/OtpVerifyForm.vue'
</script>

<template>
  <AuthPageLayout
    title="Verify your identity"
    description="Enter the 6-digit code we sent to your email"
    :show-terms="false"
  >
    <OtpVerifyForm />
  </AuthPageLayout>
</template>
