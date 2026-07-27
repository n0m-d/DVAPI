import type { ClassValue } from 'clsx'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function truncateText(value: string, maxLength = 50, ellipsis = '...'): string {
  const text = value.trim()
  if (text.length <= maxLength) return text
  return `${text.slice(0, maxLength)}${ellipsis}`
}
