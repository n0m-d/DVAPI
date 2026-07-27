import { ref } from 'vue'

const isRouteLoading = ref(true)

export function useRouteLoading() {
  return { isRouteLoading }
}

export function startRouteLoading() {
  isRouteLoading.value = true
}

export function stopRouteLoading() {
  isRouteLoading.value = false
}
