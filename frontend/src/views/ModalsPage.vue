<script setup lang="ts">
import { ref } from 'vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
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
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import DashboardPageHeader from '@/components/dashboard/DashboardPageHeader.vue'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { toast } from 'vue-sonner'

const dialogOpen = ref(false)
const formData = ref({ name: '', email: '' })

function showSuccessToast() {
  toast.success('Success!', {
    description: 'Your action was completed successfully.',
  })
}

function showErrorToast() {
  toast.error('Error', {
    description: 'Something went wrong. Please try again.',
  })
}

function showInfoToast() {
  toast.info('Information', {
    description: 'This is an informational message.',
  })
}

function handleFormSubmit() {
  toast.success('Form Submitted', {
    description: `Name: ${formData.value.name}, Email: ${formData.value.email}`,
  })
  dialogOpen.value = false
  formData.value = { name: '', email: '' }
}
</script>

<template>
  <DashboardPageHeader>
    <template #breadcrumb>
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem class="hidden md:block">
            <BreadcrumbLink href="/">
              Dashboard
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator class="hidden md:block" />
          <BreadcrumbItem>
            <BreadcrumbPage>Modals & Alerts</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>
  </DashboardPageHeader>
  <div class="dash-content container mx-auto space-y-8">
    <div class="space-y-1">
      <p class="dash-eyebrow">
        Interface kit
      </p>
      <h1 class="dash-title text-3xl font-bold tracking-tight">
        Modals & Alerts
      </h1>
      <p class="dash-subtitle">
        Examples of different modal dialogs and alert notifications
      </p>
    </div>

    <!-- Dialogs Section -->
    <Card>
      <CardHeader>
        <CardTitle>Dialog Modals</CardTitle>
        <CardDescription>
          Interactive dialog boxes for user input and confirmations
        </CardDescription>
      </CardHeader>
      <CardContent class="flex flex-wrap gap-4">
        <!-- Simple Dialog -->
        <Dialog>
          <DialogTrigger as-child>
            <Button>Open Simple Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Simple Dialog</DialogTitle>
              <DialogDescription>
                This is a basic dialog with some information.
              </DialogDescription>
            </DialogHeader>
            <div class="py-4">
              <p class="text-sm text-muted-foreground">
                Dialog content goes here. You can add any content you want.
              </p>
            </div>
            <DialogFooter>
              <Button type="button">Close</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <!-- Form Dialog -->
        <Dialog v-model:open="dialogOpen">
          <DialogTrigger as-child>
            <Button variant="outline">Open Form Dialog</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Edit Profile</DialogTitle>
              <DialogDescription>
                Make changes to your profile here. Click save when you're done.
              </DialogDescription>
            </DialogHeader>
            <div class="grid gap-4 py-4">
              <div class="grid gap-2">
                <Label for="name">Name</Label>
                <Input
                  id="name"
                  v-model="formData.name"
                  placeholder="Enter your name"
                />
              </div>
              <div class="grid gap-2">
                <Label for="email">Email</Label>
                <Input
                  id="email"
                  v-model="formData.email"
                  type="email"
                  placeholder="Enter your email"
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" @click="handleFormSubmit">
                Save changes
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>

    <!-- Alert Dialogs Section -->
    <Card>
      <CardHeader>
        <CardTitle>Alert Dialogs</CardTitle>
        <CardDescription>
          Confirmation dialogs for important actions
        </CardDescription>
      </CardHeader>
      <CardContent class="flex flex-wrap gap-4">
        <!-- Delete Confirmation -->
        <AlertDialog>
          <AlertDialogTrigger as-child>
            <Button variant="destructive">Delete Item</Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
              <AlertDialogDescription>
                This action cannot be undone. This will permanently delete your
                account and remove your data from our servers.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction @click="showSuccessToast">
                Continue
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>

        <!-- Confirmation Dialog -->
        <AlertDialog>
          <AlertDialogTrigger as-child>
            <Button variant="outline">Confirm Action</Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Confirm your action</AlertDialogTitle>
              <AlertDialogDescription>
                Are you sure you want to proceed with this action?
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction>Confirm</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </CardContent>
    </Card>

    <!-- Toast Notifications Section -->
    <Card>
      <CardHeader>
        <CardTitle>Toast Notifications</CardTitle>
        <CardDescription>
          Temporary notification messages
        </CardDescription>
      </CardHeader>
      <CardContent class="flex flex-wrap gap-4">
        <Button @click="showSuccessToast">
          Show Success Toast
        </Button>
        <Button variant="destructive" @click="showErrorToast">
          Show Error Toast
        </Button>
        <Button variant="outline" @click="showInfoToast">
          Show Info Toast
        </Button>
      </CardContent>
    </Card>
  </div>
</template>
