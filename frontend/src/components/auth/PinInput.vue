<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { computed, nextTick, ref, watch } from 'vue'
import { cn } from '@/lib/utils'

const props = withDefaults(defineProps<{
  length?: number
  modelValue?: string
  class?: HTMLAttributes['class']
}>(), {
  length: 6,
  modelValue: '',
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'complete', value: string): void
}>()

const inputs = ref<HTMLInputElement[]>([])

const digits = computed({
  get: () => {
    const chars = props.modelValue.split('')
    return Array.from({ length: props.length }, (_, index) => chars[index] ?? '')
  },
  set: (value: string[]) => {
    const nextValue = value.join('').slice(0, props.length)
    emit('update:modelValue', nextValue)

    if (nextValue.length === props.length) {
      emit('complete', nextValue)
    }
  },
})

function focusInput(index: number) {
  nextTick(() => {
    inputs.value[index]?.focus()
    inputs.value[index]?.select()
  })
}

function handleInput(index: number, event: Event) {
  const target = event.target as HTMLInputElement
  const value = target.value.replace(/\D/g, '').slice(-1)
  const nextDigits = [...digits.value]
  nextDigits[index] = value
  digits.value = nextDigits
  target.value = value

  if (value && index < props.length - 1) {
    focusInput(index + 1)
  }
}

function handleKeydown(index: number, event: KeyboardEvent) {
  if (event.key === 'Backspace' && !digits.value[index] && index > 0) {
    const nextDigits = [...digits.value]
    nextDigits[index - 1] = ''
    digits.value = nextDigits
    focusInput(index - 1)
  }

  if (event.key === 'ArrowLeft' && index > 0) {
    event.preventDefault()
    focusInput(index - 1)
  }

  if (event.key === 'ArrowRight' && index < props.length - 1) {
    event.preventDefault()
    focusInput(index + 1)
  }
}

function handlePaste(event: ClipboardEvent) {
  event.preventDefault()
  const pasted = event.clipboardData?.getData('text').replace(/\D/g, '').slice(0, props.length) ?? ''
  digits.value = pasted.split('').concat(Array(props.length).fill('')).slice(0, props.length)

  if (pasted.length === props.length) {
    focusInput(props.length - 1)
  }
  else {
    focusInput(Math.min(pasted.length, props.length - 1))
  }
}

watch(
  () => props.modelValue,
  (value) => {
    if (!value) {
      focusInput(0)
    }
  },
)
</script>

<template>
  <div :class="cn('flex items-center justify-center gap-2', props.class)">
    <input
      v-for="(_, index) in length"
      :key="index"
      :ref="(el) => { if (el) inputs[index] = el as HTMLInputElement }"
      :value="digits[index]"
      type="text"
      inputmode="numeric"
      autocomplete="one-time-code"
      maxlength="1"
      class="auth-pin-input"
      @input="handleInput(index, $event)"
      @keydown="handleKeydown(index, $event)"
      @paste="handlePaste"
    >
  </div>
</template>
