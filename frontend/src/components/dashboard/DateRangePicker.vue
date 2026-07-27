<script setup lang="ts">
import { CalendarDate, DateFormatter, getLocalTimeZone } from '@internationalized/date'

import { Calendar as CalendarIcon } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import { ref } from 'vue'

const date = ref({
  start: new CalendarDate(2023, 1, 20),
  end: new CalendarDate(2023, 1, 20).add({ days: 20 }),
})

const df = new DateFormatter('en-US', {
  dateStyle: 'medium',
})
</script>

<template>
  <div :class="cn('grid gap-2', $attrs.class ?? '')">
    <Popover>
      <PopoverTrigger as-child>
        <Button
          id="date"
          variant="outline"
          :class="cn(
            'w-[260px] justify-start text-left font-normal',
            !date && 'text-muted-foreground',
          )"
        >
          <CalendarIcon class="mr-2 h-4 w-4" />
          <span>
            {{ date.start ? (date.end ? `${df.format(date.start.toDate(getLocalTimeZone()))} - ${df.format(date.end.toDate(getLocalTimeZone()))}` : df.format(date.start.toDate(getLocalTimeZone()))) : "Pick a date" }}
          </span>
        </Button>
      </PopoverTrigger>
      <PopoverContent class="w-auto p-0" align="end">
          <div class="p-4 text-center text-sm text-muted-foreground">
             Calendar Component not installed (requires `v-calendar` or similar shadcn component)
          </div>
      </PopoverContent>
    </Popover>
  </div>
</template>
