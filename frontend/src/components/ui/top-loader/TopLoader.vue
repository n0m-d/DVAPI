<script setup lang="ts">
import { onUnmounted, ref, watch } from 'vue'
import { useRouteLoading } from '@/composables/useRouteLoading'

const { isRouteLoading } = useRouteLoading()
const progress = ref(0)
const visible = ref(false)
let tick: number | null = null
let hideTimeout: number | null = null

const clearTimers = () => {
  if (tick) {
    cancelAnimationFrame(tick)
    tick = null
  }
  if (hideTimeout) {
    clearTimeout(hideTimeout)
    hideTimeout = null
  }
}

const animateForward = () => {
  clearTimers()
  visible.value = true
  progress.value = 15

  const step = () => {
    progress.value += Math.max(1, (100 - progress.value) * 0.08)
    progress.value = Math.min(progress.value, 90)
    if (progress.value < 90) {
      tick = requestAnimationFrame(step)
    }
  }

  tick = requestAnimationFrame(step)
}

const completeProgress = () => {
  clearTimers()
  visible.value = true
  progress.value = 100
  hideTimeout = window.setTimeout(() => {
    visible.value = false
    progress.value = 0
  }, 300)
}

watch(
  isRouteLoading,
  (loading) => {
    if (loading) {
      animateForward()
    }
    else {
      completeProgress()
    }
  },
  { immediate: true },
)

onUnmounted(() => {
  clearTimers()
})
</script>

<template>
  <Teleport to="body">
    <transition name="top-loader-fade">
      <div
        v-if="visible"
        class="top-loader"
      >
        <div
          class="top-loader__bar"
          :style="{ width: `${progress}%` }"
        />
      </div>
    </transition>
  </Teleport>
</template>

<style scoped>
.top-loader {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 4px;
  z-index: 99999;
  pointer-events: none;
}

.top-loader__bar {
  height: 100%;
  width: 0;
  border-radius: 999px;
  background-color: var(--top-loader-color, var(--primary));
  background: linear-gradient(
    90deg,
    var(--top-loader-color, var(--primary)) 0%,
    color-mix(in oklch, var(--top-loader-color, var(--primary)) 70%, transparent) 100%
  );
  box-shadow: 0 0 12px color-mix(in oklch, var(--top-loader-color, var(--primary)) 65%, transparent);
  transition: width 0.2s ease;
}

.top-loader-fade-enter-active,
.top-loader-fade-leave-active {
  transition: opacity 0.3s ease;
}

.top-loader-fade-enter-from,
.top-loader-fade-leave-to {
  opacity: 0;
}
</style>
