<script setup lang="ts">
import { ref } from 'vue'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'
import {
  Tabs,
  TabsContent,
} from '@/components/ui/tabs'
import CourseBrowseSection from '@/components/learning/CourseBrowseSection.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'

const activeTab = ref('live')
</script>

<template>
  <LearningPageShell eyebrow="Admin" title="Courses"
    description="View and manage all course offerings on the platform.">
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
            <BreadcrumbPage>Courses</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <Tabs v-model="activeTab" class="w-full">
      <TabsContent value="live" class="mt-6">
        <CourseBrowseSection title="Platform catalog"
          description="Search live courses and review details. Instructor email is hidden in the modal." />
      </TabsContent>
    </Tabs>
  </LearningPageShell>
</template>
