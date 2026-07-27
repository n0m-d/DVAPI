<script setup lang="ts">
import AuthPageLayout from '@/layouts/AuthPageLayout.vue'
import RegisterForm from '@/components/auth/RegisterForm.vue'
</script>

<template>
  <AuthPageLayout
    title="Join Schole"
    description="Create your account to start learning or teaching."
  >
    <RegisterForm />
  </AuthPageLayout>
</template>
