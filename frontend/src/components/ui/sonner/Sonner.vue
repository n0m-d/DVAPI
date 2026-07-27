<script lang="ts" setup>
import { computed } from "vue"
import type { ToasterProps } from "vue-sonner"
import { CircleCheckIcon, InfoIcon, Loader2Icon, OctagonXIcon, TriangleAlertIcon, XIcon } from "lucide-vue-next"
import { Toaster as Sonner } from "vue-sonner"
import { cn } from "@/lib/utils"

const props = withDefaults(defineProps<ToasterProps>(), {
  position: 'top-right',
  richColors: true,
  closeButton: false,
  expand: true,
  offset: '16px',
})

type ToastOptions = NonNullable<ToasterProps["toastOptions"]>

const toasterClass = computed(() =>
  cn(
    "sonner-toaster pointer-events-none fixed inset-x-4 top-4 z-[9999] flex flex-col gap-3 sm:inset-x-auto sm:right-4 sm:top-4 sm:items-end",
    props.class,
  ),
)

const toastOptions = computed(() => {
  const base = props.toastOptions ?? {}

  return {
    ...base,
    class: cn(
      [
        "sonner-toast group pointer-events-auto relative flex w-full max-w-full items-start gap-3 overflow-hidden rounded-xl border border-border bg-card/95 px-4 py-3 text-sm text-card-foreground shadow-lg ring-1 ring-black/5 backdrop-blur supports-[backdrop-filter]:bg-card/80 sm:w-[360px]",
        "data-[type=success]:border-[#16a34a] data-[type=success]:bg-[#16a34a]/10",
        "data-[type=error]:border-[#ef4444] data-[type=error]:bg-[#ef4444]/10",
        "data-[type=info]:border-[#3b82f6] data-[type=info]:bg-[#3b82f6]/10",
        "data-[type=warning]:border-[color:oklch(0.828_0.189_84.429)] data-[type=warning]:bg-[color:oklch(0.828_0.189_84.429)/0.12]",
      ],
      base.class as string | undefined,
    ),
    style: {
      borderRadius: "var(--radius)",
      borderWidth: "1px",
      ...base.style,
    },
    unstyled: true,
  } as ToastOptions
})
</script>

<template>
  <Sonner
    :position="props.position"
    :rich-colors="true"
    :expand="props.expand"
    :offset="props.offset"
    :unstyled="true"
    :class="toasterClass"
    :toast-options="toastOptions"
  >
    <template #success-icon>
      <CircleCheckIcon class="size-4" />
    </template>
    <template #info-icon>
      <InfoIcon class="size-4" />
    </template>
    <template #warning-icon>
      <TriangleAlertIcon class="size-4" />
    </template>
    <template #error-icon>
      <OctagonXIcon class="size-4" />
    </template>
  </Sonner>
</template>
