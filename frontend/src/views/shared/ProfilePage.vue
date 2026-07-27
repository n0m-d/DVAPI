<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
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
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import FormActions from '@/components/forms/FormActions.vue'
import FormCheckbox from '@/components/forms/FormCheckbox.vue'
import FormField from '@/components/forms/FormField.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'
import { getCurrentUser, updateOwnProfile } from '@/api/users'
import { ApiError } from '@/lib/api'
import { useAuth } from '@/composables/useAuth'

type AccountTab = 'profile' | 'password'

const route = useRoute()
const router = useRouter()
const { syncUser, updatePassword } = useAuth()

const validTabs: AccountTab[] = ['profile', 'password']

const activeTab = computed({
  get(): AccountTab {
    const tab = route.query.tab
    if (typeof tab === 'string' && validTabs.includes(tab as AccountTab)) {
      return tab as AccountTab
    }
    return 'profile'
  },
  set(tab: AccountTab) {
    router.replace({ query: tab === 'profile' ? {} : { tab } })
  },
})

const name = ref('')
const email = ref('')
const role = ref('')
const profileLoading = ref(true)
const profileSaving = ref(false)
const profileError = ref('')
const profileFieldErrors = ref<Record<string, string>>({})
const profileUserId = ref('')

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

const emailGrades = ref(true)
const emailAssignments = ref(true)
const emailAnnouncements = ref(false)
const emailReminders = ref(true)

const passwordLoading = ref(false)


async function loadProfile() {
  profileLoading.value = true
  profileError.value = ''
  profileFieldErrors.value = {}

  try {
    const response = await getCurrentUser()
    const user = response.data
    name.value = user.full_name
    email.value = user.email
    role.value = user.role
    profileUserId.value = user.id
    syncUser(user)
  }
  catch (error) {
    profileError.value = error instanceof ApiError
      ? error.message
      : 'Unable to load profile.'
  }
  finally {
    profileLoading.value = false
  }
}

async function handleProfileSubmit() {
  if (profileSaving.value || !profileUserId.value) return

  profileSaving.value = true
  profileError.value = ''
  profileFieldErrors.value = {}

  try {
    const response = await updateOwnProfile(profileUserId.value, {
      full_name: name.value.trim(),
      email: email.value.trim(),
    })
    const user = response.data
    name.value = user.full_name
    email.value = user.email
    role.value = user.role
    syncUser(user)
    toast.success(response.message || 'Profile updated successfully.')
  }
  catch (error) {
    if (error instanceof ApiError) {
      profileError.value = error.message
      profileFieldErrors.value = error.fieldErrors
    }
    else {
      profileError.value = 'Unable to update profile.'
    }
  }
  finally {
    profileSaving.value = false
  }
}

async function handlePasswordSubmit(event: Event) {
  event.preventDefault()

  if (passwordLoading.value) return

  if (newPassword.value !== confirmPassword.value) {
    toast.error('New passwords do not match.')
    return
  }

  passwordLoading.value = true

  try {
    const response = await updatePassword({
      current_password: currentPassword.value,
      password: newPassword.value,
      confirm_password: confirmPassword.value,
    })
    toast.success(response.message || 'Password updated successfully.')
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  }
  catch (error) {
    toast.error(error instanceof ApiError ? error.message : 'Unable to update password.')
  }
  finally {
    passwordLoading.value = false
  }
}

function handleSettingsSave(event: Event) {
  event.preventDefault()
  toast.success('Preferences saved.')
}

watch(activeTab, (tab) => {
  if (tab === 'profile' && !name.value && !profileLoading.value) {
    loadProfile()
  }
})

onMounted(loadProfile)
</script>

<template>
  <LearningPageShell
    eyebrow="Account"
    title="Account"
    description="Manage your profile, password, and preferences."
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
            <BreadcrumbPage>Account</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <Tabs
      v-model="activeTab"
      class="max-w-2xl"
    >
      <TabsList class="grid w-full grid-cols-2">
        <TabsTrigger value="profile">
          Profile
        </TabsTrigger>
        <TabsTrigger value="password">
          Password
        </TabsTrigger>
        <!-- <TabsTrigger value="settings">
          Settings
        </TabsTrigger> -->
      </TabsList>

      <TabsContent
        value="profile"
        class="mt-6"
      >
        <Card class="form-card">
          <CardHeader>
            <CardTitle>Personal details</CardTitle>
          </CardHeader>
          <CardContent>
            <p
              v-if="profileError"
              class="mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive"
            >
              {{ profileError }}
            </p>

            <p
              v-else-if="profileLoading"
              class="text-sm text-muted-foreground"
            >
              Loading profile...
            </p>

            <form
              v-else
              class="app-form"
              @submit.prevent="handleProfileSubmit"
            >
              <FormField
                label="Full name"
                html-for="name"
                :error="profileFieldErrors.full_name"
              >
                <Input
                  id="name"
                  v-model="name"
                  autocomplete="name"
                  :disabled="profileSaving"
                  required
                />
              </FormField>

              <FormField
                label="Email"
                html-for="email"
                :error="profileFieldErrors.email"
              >
                <Input
                  id="email"
                  v-model="email"
                  type="email"
                  autocomplete="email"
                  :disabled="profileSaving"
                  required
                />
              </FormField>

              <FormField
                label="Role"
                html-for="role"
              >
                <Input
                  id="role"
                  v-model="role"
                  class="capitalize"
                  readonly
                  disabled
                />
              </FormField>

              <FormActions>
                <Button
                  type="button"
                  variant="outline"
                  :disabled="profileLoading || profileSaving"
                  @click="loadProfile"
                >
                  Refresh
                </Button>
                <Button
                  type="submit"
                  :disabled="profileSaving"
                >
                  {{ profileSaving ? 'Saving...' : 'Save changes' }}
                </Button>
              </FormActions>
            </form>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent
        value="password"
        class="mt-6"
      >
        <Card class="form-card">
          <CardHeader>
            <CardTitle>Change password</CardTitle>
            <CardDescription>
              Choose a strong password with at least 8 characters.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form
              class="app-form"
              @submit="handlePasswordSubmit"
            >
              <FormField
                label="Current password"
                html-for="current-password"
              >
                <Input
                  id="current-password"
                  v-model="currentPassword"
                  type="password"
                  autocomplete="current-password"
                  required
                />
              </FormField>

              <FormField
                label="New password"
                html-for="new-password"
              >
                <Input
                  id="new-password"
                  v-model="newPassword"
                  type="password"
                  autocomplete="new-password"
                  minlength="8"
                  required
                />
              </FormField>

              <FormField
                label="Confirm new password"
                html-for="confirm-password"
              >
                <Input
                  id="confirm-password"
                  v-model="confirmPassword"
                  type="password"
                  autocomplete="new-password"
                  minlength="8"
                  required
                />
              </FormField>

              <FormActions>
                <Button
                  type="submit"
                  :disabled="passwordLoading"
                >
                  {{ passwordLoading ? 'Updating password...' : 'Update password' }}
                </Button>
              </FormActions>
            </form>
          </CardContent>
        </Card>
      </TabsContent>

      <!-- <TabsContent
        value="settings"
        class="mt-6"
      >
        <Card class="form-card">
          <CardHeader>
            <CardTitle>Preferences</CardTitle>
            <CardDescription>
              Choose which emails you receive from Schole.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form
              class="app-form"
              @submit="handleSettingsSave"
            >
              <FormCheckbox
                v-model="emailGrades"
                title="Grade notifications"
                description="When a new grade is posted"
              />
              <FormCheckbox
                v-model="emailAssignments"
                title="Assignment reminders"
                description="Upcoming due dates and new assignments"
              />
              <FormCheckbox
                v-model="emailAnnouncements"
                title="Course announcements"
                description="Updates from your instructors"
              />
              <FormCheckbox
                v-model="emailReminders"
                title="Weekly digest"
                description="Summary of activity across your courses"
              />

              <div class="form-note">
                <p class="form-note-title">Theme</p>
                <p class="form-note-desc">
                  Use the theme toggle in the header to switch between light and dark
                  mode. Your preference is saved automatically.
                </p>
              </div>

              <FormActions>
                <Button type="submit">
                  Save preferences
                </Button>
              </FormActions>
            </form>
          </CardContent>
        </Card>
      </TabsContent> -->
    </Tabs>
  </LearningPageShell>
</template>
