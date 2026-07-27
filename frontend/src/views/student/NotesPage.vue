<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'
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
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'
import FormField from '@/components/forms/FormField.vue'
import FormActions from '@/components/forms/FormActions.vue'
import FormTextarea from '@/components/forms/FormTextarea.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import {
  createNote,
  deleteNote,
  listNotes,
  updateNote,
  type ApiNote,
} from '@/api/notes'
import { useAuth } from '@/composables/useAuth'
import { ApiError } from '@/lib/api'

const { user } = useAuth()

const notes = ref<ApiNote[]>([])
const loading = ref(true)
const pageError = ref('')
const dialogOpen = ref(false)
const editingNote = ref<ApiNote | null>(null)
const title = ref('')
const body = ref('')
const saving = ref(false)
const deletingId = ref('')
const formError = ref('')

function sortNotes(items: ApiNote[]) {
  return [...items].sort(
    (a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime(),
  )
}

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString([], {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function loadPage() {
  loading.value = true
  pageError.value = ''

  try {
    if (!user.value.id) {
      throw new Error('You must be signed in to view notes.')
    }
    notes.value = sortNotes(await listNotes(user.value.id))
  }
  catch (error) {
    notes.value = []
    pageError.value = error instanceof ApiError
      ? error.message
      : error instanceof Error
        ? error.message
        : 'Unable to load notes.'
  }
  finally {
    loading.value = false
  }
}

function openCreateDialog() {
  editingNote.value = null
  title.value = ''
  body.value = ''
  formError.value = ''
  dialogOpen.value = true
}

function openEditDialog(note: ApiNote) {
  editingNote.value = note
  title.value = note.title
  body.value = note.body
  formError.value = ''
  dialogOpen.value = true
}

async function handleSave() {
  if (saving.value) return

  const trimmedTitle = title.value.trim()
  const trimmedBody = body.value.trim()

  if (!trimmedTitle) {
    formError.value = 'Title is required.'
    return
  }
  if (!trimmedBody) {
    formError.value = 'Body is required.'
    return
  }
  if (!user.value.id) {
    formError.value = 'You must be signed in to save notes.'
    return
  }

  saving.value = true
  formError.value = ''

  try {
    if (editingNote.value) {
      const updated = await updateNote(editingNote.value.id, {
        title: trimmedTitle,
        body: trimmedBody,
      })
      notes.value = sortNotes(
        notes.value.map(item => item.id === updated.id ? updated : item),
      )
      toast.success('Note updated.')
    }
    else {
      const created = await createNote({
        userId: user.value.id,
        title: trimmedTitle,
        body: trimmedBody,
      })
      notes.value = sortNotes([created, ...notes.value])
      toast.success('Note created.')
    }
    dialogOpen.value = false
  }
  catch (error) {
    formError.value = error instanceof ApiError
      ? error.message
      : 'Unable to save note.'
    toast.error(formError.value)
  }
  finally {
    saving.value = false
  }
}

async function handleDelete(note: ApiNote) {
  if (deletingId.value) return
  deletingId.value = note.id

  try {
    await deleteNote(note.id)
    notes.value = notes.value.filter(item => item.id !== note.id)
    toast.success('Note deleted.')
  }
  catch (error) {
    toast.error(error instanceof ApiError ? error.message : 'Unable to delete note.')
  }
  finally {
    deletingId.value = ''
  }
}

onMounted(loadPage)
</script>

<template>
  <LearningPageShell
    eyebrow="Personal"
    title="Notes"
    description="Private notes for your courses and study sessions."
  >
    <template #breadcrumb>
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem class="hidden md:block">
            <BreadcrumbLink href="/">
              Schole
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator class="hidden md:block" />
          <BreadcrumbItem>
            <BreadcrumbPage>Notes</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <p
      v-if="pageError"
      class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
    >
      {{ pageError }}
    </p>

    <Card>
      <CardHeader class="gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <CardTitle>Your notes</CardTitle>
          <CardDescription>
            {{ loading ? 'Loading notes...' : `${notes.length} note${notes.length === 1 ? '' : 's'}` }}
          </CardDescription>
        </div>
        <Button :disabled="loading" @click="openCreateDialog">
          New note
        </Button>
      </CardHeader>
      <CardContent class="space-y-4">
        <p v-if="loading" class="text-sm text-muted-foreground">
          Loading notes...
        </p>
        <template v-else>
          <div
            v-for="note in notes"
            :key="note.id"
            class="rounded-lg border border-border/60 bg-muted/20 p-4"
          >
            <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
              <div class="min-w-0 space-y-2">
                <h4 class="font-medium">
                  {{ note.title }}
                </h4>
                <p class="whitespace-pre-wrap text-sm text-muted-foreground" v-html="note.body">
                </p>
                <p class="text-xs text-muted-foreground">
                  Updated {{ formatDate(note.updatedAt) }}
                </p>
              </div>
              <div class="flex shrink-0 flex-wrap gap-2">
                <Button type="button" variant="outline" size="sm" @click="openEditDialog(note)">
                  Edit
                </Button>
                <AlertDialog>
                  <AlertDialogTrigger as-child>
                    <Button
                      type="button"
                      variant="destructive"
                      size="sm"
                      :disabled="Boolean(deletingId)"
                    >
                      Delete
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Delete this note?</AlertDialogTitle>
                      <AlertDialogDescription>
                        This permanently deletes “{{ note.title }}”.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel :disabled="deletingId === note.id">
                        Cancel
                      </AlertDialogCancel>
                      <AlertDialogAction
                        class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        :disabled="deletingId === note.id"
                        @click="handleDelete(note)"
                      >
                        {{ deletingId === note.id ? 'Deleting...' : 'Delete note' }}
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            </div>
          </div>
          <p
            v-if="notes.length === 0"
            class="text-center text-sm text-muted-foreground"
          >
            No notes yet. Create one to capture ideas from class.
          </p>
        </template>
      </CardContent>
    </Card>

    <Dialog v-model:open="dialogOpen">
      <DialogContent class="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {{ editingNote ? 'Edit note' : 'Create note' }}
          </DialogTitle>
          <DialogDescription>
            {{ editingNote ? 'Update the title or body of this note.' : 'Add a personal note for later.' }}
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-6" @submit.prevent="handleSave">
          <p
            v-if="formError"
            class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
          >
            {{ formError }}
          </p>

          <FormField label="Title" html-for="note-title">
            <Input
              id="note-title"
              v-model="title"
              required
              placeholder="Lecture recap"
            />
          </FormField>

          <FormField label="Body" html-for="note-body">
            <FormTextarea
              id="note-body"
              v-model="body"
              :rows="8"
              required
              placeholder="Write your note..."
            />
          </FormField>

          <FormActions>
            <Button type="button" variant="outline" :disabled="saving" @click="dialogOpen = false">
              Cancel
            </Button>
            <Button type="submit" :disabled="saving">
              {{ saving ? 'Saving...' : (editingNote ? 'Save changes' : 'Create note') }}
            </Button>
          </FormActions>
        </form>
      </DialogContent>
    </Dialog>
  </LearningPageShell>
</template>
