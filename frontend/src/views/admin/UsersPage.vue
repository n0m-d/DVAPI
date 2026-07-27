<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'
import { Search } from 'lucide-vue-next'
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import FormField from '@/components/forms/FormField.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { roleLabels } from '@/config/navigation'
import {
  createAdminUser,
  listAdminUsers,
  updateAdminUser,
  type CreateAdminUserPayload,
} from '@/api/admin'
import type { ApiUser } from '@/api/auth'
import { useAuth } from '@/composables/useAuth'
import { ApiError } from '@/lib/api'
import { cn } from '@/lib/utils'
import type { UserRole } from '@/types/roles'

const roles: UserRole[] = ['student', 'instructor', 'admin']
const pageSize = 10
const { user: currentUser, syncUser } = useAuth()

const roleStyles: Record<UserRole, string> = {
  student: 'border-blue-500/30 bg-blue-500/10 text-blue-600 dark:text-blue-400',
  instructor: 'border-violet-500/30 bg-violet-500/10 text-violet-600 dark:text-violet-400',
  admin: 'border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400',
}

const users = ref<ApiUser[]>([])
const search = ref('')
const roleFilter = ref<'all' | UserRole>('all')
const page = ref(1)
const total = ref(0)
const totalPages = ref(0)
const loading = ref(true)
const loadError = ref('')
let requestId = 0

const createOpen = ref(false)
const editOpen = ref(false)
const saving = ref(false)
const formError = ref('')
const fieldErrors = ref<Record<string, string>>({})
const editingUser = ref<ApiUser | null>(null)
const createForm = ref<CreateAdminUserPayload>({
  email: '',
  full_name: '',
  role: 'student',
  password: '',
})
const editForm = ref({
  email: '',
  full_name: '',
  role: 'student' as UserRole,
})

const editingSelf = computed(() => editingUser.value?.id === currentUser.value.id)

async function fetchUsers() {
  const activeRequest = ++requestId
  loading.value = true
  loadError.value = ''

  try {
    const response = await listAdminUsers({
      search: search.value,
      role: roleFilter.value === 'all' ? undefined : roleFilter.value,
      page: page.value,
      page_size: pageSize,
    })
    if (activeRequest !== requestId) return
    users.value = response.data.users
    total.value = response.data.pagination.total
    totalPages.value = response.data.pagination.total_pages
    page.value = response.data.pagination.page
  }
  catch (error) {
    if (activeRequest !== requestId) return
    users.value = []
    total.value = 0
    totalPages.value = 0
    loadError.value = error instanceof ApiError ? error.message : 'Unable to load users.'
  }
  finally {
    if (activeRequest === requestId) loading.value = false
  }
}

const debouncedSearch = useDebounceFn(() => {
  page.value = 1
  fetchUsers()
}, 300)

watch(search, debouncedSearch)
watch(roleFilter, () => {
  page.value = 1
  fetchUsers()
})

function resetFormErrors() {
  formError.value = ''
  fieldErrors.value = {}
}

function openCreateDialog() {
  createForm.value = {
    email: '',
    full_name: '',
    role: 'student',
    password: '',
  }
  resetFormErrors()
  createOpen.value = true
}

function openEditDialog(user: ApiUser) {
  editingUser.value = user
  editForm.value = {
    email: user.email,
    full_name: user.full_name,
    role: user.role,
  }
  resetFormErrors()
  editOpen.value = true
}

function setMutationError(error: unknown, fallback: string) {
  if (error instanceof ApiError) {
    formError.value = error.message
    fieldErrors.value = error.fieldErrors
    toast.error(error.message)
  }
  else {
    formError.value = fallback
    toast.error(fallback)
  }
}

async function handleCreate() {
  if (saving.value) return
  saving.value = true
  resetFormErrors()

  try {
    const response = await createAdminUser({
      email: createForm.value.email.trim(),
      full_name: createForm.value.full_name.trim(),
      role: createForm.value.role,
      password: createForm.value.password,
    })
    toast.success(response.message || 'User created successfully.')
    createOpen.value = false
    page.value = 1
    await fetchUsers()
  }
  catch (error) {
    setMutationError(error, 'Unable to create user.')
  }
  finally {
    saving.value = false
  }
}

async function handleEdit() {
  if (saving.value || !editingUser.value) return
  saving.value = true
  resetFormErrors()

  try {
    const response = await updateAdminUser(editingUser.value.id, {
      email: editForm.value.email.trim(),
      full_name: editForm.value.full_name.trim(),
      role: editingSelf.value ? editingUser.value.role : editForm.value.role,
    })
    if (editingSelf.value) syncUser(response.data)
    toast.success(response.message || 'User updated successfully.')
    editOpen.value = false
    await fetchUsers()
  }
  catch (error) {
    setMutationError(error, 'Unable to update user.')
  }
  finally {
    saving.value = false
  }
}

function goToPage(nextPage: number) {
  if (loading.value || nextPage < 1 || nextPage > totalPages.value || nextPage === page.value) return
  page.value = nextPage
  fetchUsers()
}

onMounted(fetchUsers)
</script>

<template>
  <LearningPageShell
    eyebrow="Admin"
    title="Users"
    description="Manage platform accounts, roles, and access."
  >
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
            <BreadcrumbPage>Users</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
        <div class="relative w-full sm:w-64">
          <Search class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            v-model="search"
            type="search"
            placeholder="Search users..."
            class="pl-9"
            aria-label="Search users"
          />
        </div>
        <Select v-model="roleFilter">
          <SelectTrigger class="w-full sm:w-44" aria-label="Filter by role">
            <SelectValue placeholder="All roles" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All roles</SelectItem>
            <SelectItem
              v-for="roleOption in roles"
              :key="roleOption"
              :value="roleOption"
            >
              {{ roleLabels[roleOption] }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
      <Button @click="openCreateDialog">
        Create user
      </Button>
    </div>

    <p
      v-if="loadError"
      class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
    >
      {{ loadError }}
    </p>

    <Card>
      <CardHeader>
        <CardTitle>Platform users</CardTitle>
        <CardDescription>
          {{ loading ? 'Loading users...' : `${total} registered account${total === 1 ? '' : 's'}` }}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <p
          v-if="loading"
          class="py-6 text-center text-sm text-muted-foreground"
        >
          Loading users...
        </p>
        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead class="hidden md:table-cell">Joined</TableHead>
              <TableHead class="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="user in users"
              :key="user.id"
            >
              <TableCell class="font-medium">{{ user.full_name }}</TableCell>
              <TableCell>{{ user.email }}</TableCell>
              <TableCell>
                <span
                  :class="cn(
                    'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium',
                    roleStyles[user.role],
                  )"
                >
                  {{ roleLabels[user.role] }}
                </span>
              </TableCell>
              <TableCell class="hidden md:table-cell">
                {{ new Date(user.created_at).toLocaleDateString() }}
              </TableCell>
              <TableCell class="text-right">
                <Button
                  variant="outline"
                  size="sm"
                  @click="openEditDialog(user)"
                >
                  Edit
                </Button>
              </TableCell>
            </TableRow>
            <TableRow v-if="!users.length">
              <TableCell
                :colspan="5"
                class="py-8 text-center text-muted-foreground"
              >
                No users found.
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>

        <div
          v-if="!loading && totalPages > 1"
          class="mt-4 flex items-center justify-between gap-3"
        >
          <Button
            variant="outline"
            size="sm"
            :disabled="page <= 1"
            @click="goToPage(page - 1)"
          >
            Previous
          </Button>
          <p class="text-xs text-muted-foreground">
            Page {{ page }} of {{ totalPages }}
          </p>
          <Button
            variant="outline"
            size="sm"
            :disabled="page >= totalPages"
            @click="goToPage(page + 1)"
          >
            Next
          </Button>
        </div>
      </CardContent>
    </Card>

    <Dialog v-model:open="createOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create user</DialogTitle>
          <DialogDescription>
            Add a platform account and assign its role.
          </DialogDescription>
        </DialogHeader>
        <form class="app-form" @submit.prevent="handleCreate">
          <p
            v-if="formError"
            class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
          >
            {{ formError }}
          </p>
          <FormField label="Full name" html-for="create-name" :error="fieldErrors.full_name">
            <Input
              id="create-name"
              v-model="createForm.full_name"
              autocomplete="name"
              :disabled="saving"
              required
            />
          </FormField>
          <FormField label="Email" html-for="create-email" :error="fieldErrors.email">
            <Input
              id="create-email"
              v-model="createForm.email"
              type="email"
              autocomplete="email"
              :disabled="saving"
              required
            />
          </FormField>
          <FormField label="Role" html-for="create-role" :error="fieldErrors.role">
            <Select v-model="createForm.role" :disabled="saving">
              <SelectTrigger id="create-role">
                <SelectValue placeholder="Select a role" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="roleOption in roles"
                  :key="roleOption"
                  :value="roleOption"
                >
                  {{ roleLabels[roleOption] }}
                </SelectItem>
              </SelectContent>
            </Select>
          </FormField>
          <FormField label="Password" html-for="create-password" :error="fieldErrors.password">
            <Input
              id="create-password"
              v-model="createForm.password"
              type="password"
              autocomplete="new-password"
              minlength="8"
              :disabled="saving"
              required
            />
          </FormField>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="createOpen = false">
              Cancel
            </Button>
            <Button type="submit" :disabled="saving">
              {{ saving ? 'Creating...' : 'Create user' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="editOpen">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit user</DialogTitle>
          <DialogDescription>
            Update account details and role.
          </DialogDescription>
        </DialogHeader>
        <form class="app-form" @submit.prevent="handleEdit">
          <p
            v-if="formError"
            class="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
          >
            {{ formError }}
          </p>
          <FormField label="Full name" html-for="edit-name" :error="fieldErrors.full_name">
            <Input
              id="edit-name"
              v-model="editForm.full_name"
              autocomplete="name"
              :disabled="saving"
              required
            />
          </FormField>
          <FormField label="Email" html-for="edit-email" :error="fieldErrors.email">
            <Input
              id="edit-email"
              v-model="editForm.email"
              type="email"
              autocomplete="email"
              :disabled="saving"
              required
            />
          </FormField>
          <FormField
            label="Role"
            html-for="edit-role"
            :error="fieldErrors.role"
            :hint="editingSelf ? 'You cannot change your own admin role here.' : undefined"
          >
            <Select v-model="editForm.role" :disabled="saving || editingSelf">
              <SelectTrigger id="edit-role">
                <SelectValue placeholder="Select a role" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="roleOption in roles"
                  :key="roleOption"
                  :value="roleOption"
                >
                  {{ roleLabels[roleOption] }}
                </SelectItem>
              </SelectContent>
            </Select>
          </FormField>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="editOpen = false">
              Cancel
            </Button>
            <Button type="submit" :disabled="saving">
              {{ saving ? 'Saving...' : 'Save changes' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </LearningPageShell>
</template>
