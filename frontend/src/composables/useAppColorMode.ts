import { useColorMode } from '@vueuse/core'

export function useAppColorMode() {
  return useColorMode({
    attribute: 'class',
    modes: {
      light: '',
      dark: 'dark',
    },
    disableTransition: false,
  })
}
