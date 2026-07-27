<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'
import { Search } from 'lucide-vue-next'
import { getLibraryList, type LibraryBook } from '@/api/library'
import { ApiError } from '@/lib/api'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'

withDefaults(defineProps<{
  title?: string
  description?: string
}>(), {
  title: 'Library Books catalog',
  description: 'Search books on the online library',
})

const search = ref('')
const books = ref<LibraryBook[]>([])
const loading = ref(false)
const errorMessage = ref('')

const selected = ref<LibraryBook | null>(null)
const detailsOpen = ref(false)

async function fetchBooks() {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await getLibraryList(search.value.trim())
    books.value = response.data.books ?? []
  }
  catch (error) {
    books.value = []
    errorMessage.value = error instanceof ApiError
      ? error.message
      : 'Unable to load books.'
  }
  finally {
    loading.value = false
  }
}

const debouncedSearch = useDebounceFn(() => {
  fetchBooks()
}, 300)

watch(search, () => {
  debouncedSearch()
})

function openDetails(book: LibraryBook) {
  selected.value = book
  detailsOpen.value = true
}

function onDetailsOpenChange(open: boolean) {
  detailsOpen.value = open
  if (!open) {
    window.setTimeout(() => {
      if (!detailsOpen.value) selected.value = null
    }, 160)
  }
}

onMounted(() => {
  fetchBooks()
})
</script>

<template>
  <Card class="form-card">
    <CardHeader class="gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <CardTitle>{{ title }}</CardTitle>
        <CardDescription>
          {{ description }}
        </CardDescription>
      </div>
      <div class="relative w-full sm:max-w-xs">
        <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          v-model="search"
          type="search"
          placeholder="Search book by title..."
          class="pl-9"
          aria-label="Search books"
        />
      </div>
    </CardHeader>
    <CardContent class="space-y-4">
      <p
        v-if="errorMessage"
        class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
      >
        {{ errorMessage }}
      </p>

      <p v-else-if="loading" class="text-sm text-muted-foreground">
        Loading books...
      </p>

      <template v-else>
        <p class="text-sm text-muted-foreground">
          {{ books.length }} book{{ books.length === 1 ? '' : 's' }} found
        </p>

        <div v-if="books.length" class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          <button
            v-for="book in books"
            :key="book.id"
            type="button"
            class="rounded-xl border border-border/60 bg-card/40 p-4 text-left transition hover:border-primary/40 hover:bg-primary/5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            @click="openDetails(book)"
          >
            <p class="text-xs font-medium uppercase tracking-wider text-muted-foreground">
              {{ book.genre || 'Book' }}
            </p>
            <h3 class="mt-1 text-base font-semibold text-foreground">
              {{ book.title }}
            </h3>
            <p class="mt-1 text-sm text-muted-foreground">
              {{ book.author }}
            </p>
            <p v-if="book.year" class="mt-2 text-sm text-muted-foreground">
              {{ book.year }}
            </p>
          </button>
        </div>

        <p v-else class="text-sm text-muted-foreground">
          No books match your search.
        </p>
      </template>
    </CardContent>
  </Card>

  <Dialog :open="detailsOpen" @update:open="onDetailsOpenChange">
    <DialogContent class="sm:max-w-lg">
      <DialogHeader>
        <DialogTitle>{{ selected?.title }}</DialogTitle>
        <DialogDescription>
          {{ selected?.author }}
        </DialogDescription>
      </DialogHeader>

      <div v-if="selected" class="space-y-4 text-sm">
        <div>
          <p class="form-label">Author</p>
          <p class="mt-1 text-foreground">
            {{ selected.author }}
          </p>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
          <div v-if="selected.genre">
            <p class="form-label">Genre</p>
            <p class="mt-1 text-foreground">
              {{ selected.genre }}
            </p>
          </div>
          <div v-if="selected.year">
            <p class="form-label">Year</p>
            <p class="mt-1 text-foreground">
              {{ selected.year }}
            </p>
          </div>
        </div>

        <div v-if="selected.bookmark">
          <p class="form-label">Bookmark</p>
          <p class="mt-1 font-mono text-foreground">
            {{ selected.bookmark }}
          </p>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
