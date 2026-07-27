import { useStorage } from '@vueuse/core'
import { watch } from 'vue'

export const THEME_COLOR_OPTIONS = [
  { name: 'Red', class: 'theme-red', color: 'bg-red-500' },
  { name: 'Orange', class: 'theme-orange', color: 'bg-orange-500' },
  { name: 'Green', class: 'theme-green', color: 'bg-green-500' },
  { name: 'Blue', class: 'theme-blue', color: 'bg-blue-500' },
  { name: 'Yellow', class: 'theme-yellow', color: 'bg-yellow-500' },
  { name: 'Violet', class: 'theme-violet', color: 'bg-violet-500' },
] as const

export const THEME_COLOR_CLASSES = THEME_COLOR_OPTIONS.map(theme => theme.class)

export type ThemeColorClass = typeof THEME_COLOR_CLASSES[number]

const DEFAULT_THEME: ThemeColorClass = 'theme-green'

const currentTheme = useStorage<ThemeColorClass>('theme-color', DEFAULT_THEME)

function normalizeTheme(theme: string): ThemeColorClass {
  return THEME_COLOR_CLASSES.includes(theme as ThemeColorClass)
    ? theme as ThemeColorClass
    : DEFAULT_THEME
}

export function applyThemeColor(theme: ThemeColorClass) {
  document.body.classList.remove(...THEME_COLOR_CLASSES)
  document.body.classList.add(theme)
}

export function useThemeColor() {
  if (!THEME_COLOR_CLASSES.includes(currentTheme.value)) {
    currentTheme.value = DEFAULT_THEME
  }

  watch(currentTheme, theme => applyThemeColor(normalizeTheme(theme)), { immediate: true })

  function setTheme(themeClass: ThemeColorClass) {
    currentTheme.value = themeClass
  }

  return {
    currentTheme,
    setTheme,
    themes: THEME_COLOR_OPTIONS,
  }
}
