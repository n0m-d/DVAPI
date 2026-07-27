<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { useVModel } from '@vueuse/core'
import { cn } from '@/lib/utils'

const props = defineProps<{
  defaultValue?: string
  modelValue?: string
  class?: HTMLAttributes['class']
  rows?: number
  placeholder?: string
  id?: string
  required?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const modelValue = useVModel(props, 'modelValue', emit, {
  passive: true,
  defaultValue: props.defaultValue,
})
</script>

<template>
  <textarea
    :id="id"
    v-model="modelValue"
    data-slot="textarea"
    :rows="rows ?? 4"
    :placeholder="placeholder"
    :required="required"
    :class="cn('form-control w-full px-3 py-2 text-sm outline-none', props.class)"
  />
</template>
