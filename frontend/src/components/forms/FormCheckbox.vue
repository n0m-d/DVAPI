<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  modelValue?: boolean
}>()

defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()
</script>

<template>
  <label class="form-check">
    <input
      type="checkbox"
      :checked="modelValue"
      @change="$emit('update:modelValue', ($event.target as HTMLInputElement).checked)"
    >
    <div>
      <p class="form-check-title">{{ title }}</p>
      <p
        v-if="description"
        class="form-check-desc"
      >
        {{ description }}
      </p>
    </div>
  </label>
</template>
