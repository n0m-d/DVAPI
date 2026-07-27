<script setup lang="ts">
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useThemeColor } from '@/composables/useThemeColor'
import { Check, Paintbrush } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

const { currentTheme, setTheme, themes } = useThemeColor()
</script>

<template>
  <DropdownMenu>
    <DropdownMenuTrigger as-child>
      <Button variant="outline" size="icon">
        <Paintbrush class="h-[1.2rem] w-[1.2rem]" />
        <span class="sr-only">Toggle theme</span>
      </Button>
    </DropdownMenuTrigger>
    <DropdownMenuContent align="end">
      <DropdownMenuLabel>
        Theme
      </DropdownMenuLabel>
      <DropdownMenuSeparator />
      <DropdownMenuItem
        v-for="theme in themes"
        :key="theme.name"
        class="gap-2"
        @click="setTheme(theme.class)"
      >
        <div :class="cn('h-4 w-4 rounded-full', theme.color)" />
        <span class="flex-1">{{ theme.name }}</span>
        <Check
          v-if="currentTheme === theme.class"
          class="ml-auto h-4 w-4"
        />
      </DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>
</template>
