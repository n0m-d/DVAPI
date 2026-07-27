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
import LibrarySection from '@/components/learning/LibrarySection.vue'
import LearningPageShell from '@/components/learning/LearningPageShell.vue'

const activeTab = ref('live')
</script>

<template>
  <LearningPageShell eyebrow="Admin" title="Library"
    description="View and manage all books offerings on the e-library.">
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
            <BreadcrumbPage>Library</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
    </template>

    <Tabs v-model="activeTab" class="w-full">
      <TabsContent value="live" class="mt-6">
        <LibrarySection title="Library Books catalog"
          description="Search books on the online library" />
      </TabsContent>
    </Tabs>
  </LearningPageShell>
</template>
