<script setup lang="ts">
import { Label } from '@/components/ui/label'

defineProps<{
  label: string
  htmlFor?: string
  hint?: string
  error?: string
}>()
</script>

<template>
  <div
    class="form-field"
    :data-invalid="error ? '' : undefined"
  >
    <div class="flex items-center justify-between gap-2">
      <Label
        :for="htmlFor"
        class="form-label"
      >
        {{ label }}
      </Label>
      <slot name="action" />
    </div>
    <slot />
    <p
      v-if="error"
      class="form-error"
    >
      {{ error }}
    </p>
    <p
      v-else-if="hint"
      class="form-hint"
    >
      {{ hint }}
    </p>
  </div>
</template>
