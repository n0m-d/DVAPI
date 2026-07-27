<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { RefreshCw } from 'lucide-vue-next'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { getAdminLogs } from '@/api/admin'
import { ApiError } from '@/lib/api'
import { cn } from '@/lib/utils'

type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'UNKNOWN'

interface ParsedLogLine {
  id: number
  raw: string
  level: LogLevel
  time?: string
  method?: string
  path?: string
  status?: number
  msg?: string
  latencyMs?: number
  clientIp?: string
  requestId?: string
  extras: string[]
}

const LINE_OPTIONS = ['50', '100', '200', '500'] as const

const lines = ref<(typeof LINE_OPTIONS)[number]>('100')
const loading = ref(true)
const errorMessage = ref('')
const rawLog = ref('')
const sourceFile = ref('')
const fetchedAt = ref('')
const logEl = ref<HTMLElement | null>(null)

const levelStyles: Record<LogLevel, string> = {
  DEBUG: 'bg-slate-500/15 text-slate-300',
  INFO: 'bg-emerald-500/15 text-emerald-300',
  WARN: 'bg-amber-500/15 text-amber-300',
  ERROR: 'bg-red-500/15 text-red-300',
  UNKNOWN: 'bg-zinc-500/15 text-zinc-300',
}

const methodStyles: Record<string, string> = {
  GET: 'text-sky-300',
  POST: 'text-emerald-300',
  PUT: 'text-amber-300',
  PATCH: 'text-violet-300',
  DELETE: 'text-red-300',
  OPTIONS: 'text-zinc-400',
  HEAD: 'text-zinc-400',
}

function normalizeLevel(value: unknown): LogLevel {
  const level = String(value ?? '').toUpperCase()
  if (level === 'DEBUG' || level === 'INFO' || level === 'WARN' || level === 'WARNING' || level === 'ERROR') {
    return level === 'WARNING' ? 'WARN' : level
  }
  if (/\bERROR\b/i.test(String(value ?? ''))) return 'ERROR'
  if (/\bWARN(ING)?\b/i.test(String(value ?? ''))) return 'WARN'
  if (/\bDEBUG\b/i.test(String(value ?? ''))) return 'DEBUG'
  if (/\bINFO\b/i.test(String(value ?? ''))) return 'INFO'
  return 'UNKNOWN'
}

function statusTone(status?: number) {
  if (!status) return 'text-zinc-400'
  if (status >= 500) return 'text-red-300'
  if (status >= 400) return 'text-amber-300'
  if (status >= 300) return 'text-sky-300'
  return 'text-emerald-300'
}

function formatTime(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}

function formatLatency(ns?: number) {
  if (ns == null || Number.isNaN(ns)) return ''
  if (ns >= 1_000_000_000) return `${(ns / 1_000_000_000).toFixed(2)}s`
  if (ns >= 1_000_000) return `${(ns / 1_000_000).toFixed(1)}ms`
  if (ns >= 1_000) return `${(ns / 1_000).toFixed(1)}µs`
  return `${ns}ns`
}

function redactSensitive(text: string) {
  return text
    .replace(/(token=)[^&\s"]+/gi, '$1[redacted]')
    .replace(/(authorization["']?\s*[:=]\s*["']?)bearer\s+[^"'\s]+/gi, '$1Bearer [redacted]')
    .replace(/(password["']?\s*[:=]\s*["']?)[^"'\s]+/gi, '$1[redacted]')
}

function parseLogLine(raw: string, id: number): ParsedLogLine {
  const sanitized = redactSensitive(raw)
  try {
    const parsed = JSON.parse(sanitized) as Record<string, unknown>
    const reserved = new Set([
      'time', 'level', 'msg', 'method', 'path', 'status', 'latency',
      'client_ip', 'request_id', 'query',
    ])
    const extras = Object.entries(parsed)
      .filter(([key, value]) => !reserved.has(key) && value != null && value !== '')
      .map(([key, value]) => `${key}=${typeof value === 'string' ? value : JSON.stringify(value)}`)

    if (typeof parsed.query === 'string' && parsed.query) {
      extras.unshift(`query=${redactSensitive(parsed.query)}`)
    }

    return {
      id,
      raw: sanitized,
      level: normalizeLevel(parsed.level ?? parsed.msg),
      time: typeof parsed.time === 'string' ? parsed.time : undefined,
      method: typeof parsed.method === 'string' ? parsed.method : undefined,
      path: typeof parsed.path === 'string' ? parsed.path : undefined,
      status: typeof parsed.status === 'number' ? parsed.status : undefined,
      msg: typeof parsed.msg === 'string' ? parsed.msg : undefined,
      latencyMs: typeof parsed.latency === 'number' ? parsed.latency : undefined,
      clientIp: typeof parsed.client_ip === 'string' ? parsed.client_ip : undefined,
      requestId: typeof parsed.request_id === 'string' ? parsed.request_id : undefined,
      extras,
    }
  }
  catch {
    return {
      id,
      raw: sanitized,
      level: normalizeLevel(sanitized),
      msg: sanitized,
      extras: [],
    }
  }
}

const entries = computed(() => {
  if (!rawLog.value.trim()) return []
  return rawLog.value
    .replace(/\r\n/g, '\n')
    .split('\n')
    .filter(line => line.trim().length > 0)
    .map((line, index) => parseLogLine(line, index))
})

async function scrollToBottom() {
  await nextTick()
  if (logEl.value) {
    logEl.value.scrollTop = logEl.value.scrollHeight
  }
}

async function fetchLogs() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await getAdminLogs(Number(lines.value))
    rawLog.value = response.data ?? ''
    sourceFile.value = response.file ?? ''
    fetchedAt.value = response.fetchedAt ?? ''
    await scrollToBottom()
  }
  catch (error) {
    rawLog.value = ''
    sourceFile.value = ''
    fetchedAt.value = ''
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load application logs.'
  }
  finally {
    loading.value = false
  }
}

watch(lines, () => {
  void fetchLogs()
})

onMounted(fetchLogs)
</script>

<template>
  <LearningPageShell eyebrow="Admin" title="Logs"
    description="Tail recent application access logs for troubleshooting.">
    <template #breadcrumb>
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem class="hidden md:block">
            <BreadcrumbLink href="/admin/dashboard">
              Schole
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator class="hidden md:block" />
          <BreadcrumbItem>
            <BreadcrumbPage>Logs</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <Card class="overflow-hidden border-border/70">
      <CardHeader class="gap-4 border-b border-border/60 pb-4 sm:flex-row sm:items-center sm:justify-between">
        <div class="space-y-1">
          <CardTitle>Access log tail</CardTitle>
          <CardDescription>
            <span v-if="sourceFile">{{ sourceFile }}</span>
            <span v-else>Server log output</span>
            <span v-if="fetchedAt"> · fetched {{ formatTime(fetchedAt) }}</span>
            <span v-if="entries.length"> · {{ entries.length }} lines</span>
          </CardDescription>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <Select v-model="lines">
            <SelectTrigger class="w-[120px]">
              <SelectValue placeholder="Lines" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="option in LINE_OPTIONS" :key="option" :value="option">
                {{ option }} lines
              </SelectItem>
            </SelectContent>
          </Select>

          <Button variant="outline" size="sm" :disabled="loading" @click="fetchLogs">
            <RefreshCw :class="cn('mr-2 size-4', loading && 'animate-spin')" />
            Refresh
          </Button>
        </div>
      </CardHeader>

      <CardContent class="p-0">
        <div v-if="errorMessage" class="flex items-center justify-between gap-3 px-4 py-3 text-sm text-destructive">
          <span>{{ errorMessage }}</span>
          <Button variant="outline" size="sm" @click="fetchLogs">
            Retry
          </Button>
        </div>

        <div v-else ref="logEl"
          class="log-console max-h-[min(70vh,720px)] overflow-auto bg-[#0d1117] text-[13px] leading-5 text-zinc-200">
          <p v-if="loading && !entries.length" class="px-4 py-6 font-mono text-zinc-400">
            Loading logs…
          </p>

          <p v-else-if="!entries.length" class="px-4 py-6 font-mono text-zinc-400">
            No log lines returned.
          </p>

          <ul v-else class="divide-y divide-white/5 font-mono">
            <li v-for="entry in entries" :key="entry.id"
              class="grid grid-cols-[auto_minmax(0,1fr)] gap-3 px-3 py-2 hover:bg-white/[0.03]">
              <span class="select-none pt-0.5 text-right text-[11px] tabular-nums text-zinc-600">
                {{ entry.id + 1 }}
              </span>

              <div class="min-w-0 space-y-1">
                <div class="flex flex-wrap items-center gap-2">
                  <span :class="cn(
                    'inline-flex rounded px-1.5 py-0.5 text-[10px] font-semibold tracking-wide',
                    levelStyles[entry.level],
                  )">
                    {{ entry.level }}
                  </span>

                  <span v-if="entry.time" class="text-[11px] text-zinc-500">
                    {{ formatTime(entry.time) }}
                  </span>

                  <span v-if="entry.method" :class="cn('font-semibold', methodStyles[entry.method] ?? 'text-zinc-300')">
                    {{ entry.method }}
                  </span>

                  <span v-if="entry.path" class="truncate text-zinc-100" :title="entry.path">
                    {{ entry.path }}
                  </span>

                  <span v-if="entry.status != null" :class="cn('font-semibold tabular-nums', statusTone(entry.status))">
                    {{ entry.status }}
                  </span>

                  <span v-if="entry.latencyMs != null" class="text-zinc-500">
                    {{ formatLatency(entry.latencyMs) }}
                  </span>
                </div>

                <p v-if="entry.msg && !(entry.method && entry.path)" class="break-words text-zinc-300">
                  {{ entry.msg }}
                </p>

                <div v-if="entry.clientIp || entry.requestId || entry.extras.length"
                  class="flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-zinc-500">
                  <span v-if="entry.clientIp">ip={{ entry.clientIp }}</span>
                  <span v-if="entry.requestId" class="truncate" :title="entry.requestId">
                    req={{ entry.requestId }}
                  </span>
                  <span v-for="extra in entry.extras" :key="extra" class="truncate" :title="extra">
                    {{ extra }}
                  </span>
                </div>
              </div>
            </li>
          </ul>
        </div>
      </CardContent>
    </Card>
  </LearningPageShell>
</template>
