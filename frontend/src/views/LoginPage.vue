<script setup lang="ts">
import AuthPageLayout from '@/layouts/AuthPageLayout.vue'
import UserAuthForm from '@/components/auth/UserAuthForm.vue'
</script>

<template>
  <AuthPageLayout
    title="Sign in to Schole"
    description="Access your courses, grades, and student submissions."
  >
    <UserAuthForm />
  </AuthPageLayout>
</template>
