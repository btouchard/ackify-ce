<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { usePageTitle } from '@/composables/usePageTitle'
import { listDocuments, type Document } from '@/services/admin'
import { documentService } from '@/services/documents'
import { extractError } from '@/services/http'
import { FileText, Users, CheckCircle, ExternalLink, Settings, Loader2, Plus, Search, Webhook } from 'lucide-vue-next'
import Card from '@/components/ui/Card.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import CardTitle from '@/components/ui/CardTitle.vue'
import CardDescription from '@/components/ui/CardDescription.vue'
import CardContent from '@/components/ui/CardContent.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Alert from '@/components/ui/Alert.vue'
import AlertDescription from '@/components/ui/AlertDescription.vue'
import Table from '@/components/ui/table/Table.vue'
import TableHeader from '@/components/ui/table/TableHeader.vue'
import TableBody from '@/components/ui/table/TableBody.vue'
import TableRow from '@/components/ui/table/TableRow.vue'
import TableHead from '@/components/ui/table/TableHead.vue'
import TableCell from '@/components/ui/table/TableCell.vue'

const router = useRouter()
const { t } = useI18n()
usePageTitle('admin.title')
const documents = ref<Document[]>([])
const loading = ref(true)
const error = ref('')
const newDocId = ref('')
const creating = ref(false)

// Pagination & Filter
const searchQuery = ref('')
const currentPage = ref(1)
const perPage = ref(20)
const totalDocsCount = ref(0)

// Computed
const filteredDocuments = computed(() => {
  if (!searchQuery.value.trim()) return documents.value

  const query = searchQuery.value.toLowerCase()
  return documents.value.filter(doc =>
    doc.docId.toLowerCase().includes(query) ||
    doc.title?.toLowerCase().includes(query) ||
    doc.url?.toLowerCase().includes(query)
  )
})

const totalPages = computed(() => Math.ceil(totalDocsCount.value / perPage.value) || 1)

// Computed KPIs
const totalDocuments = computed(() => totalDocsCount.value)
const totalSigners = computed(() => {
  // For now, return 0 as expectedSigners might not be in the Document type yet
  return 0
})
const activeDocuments = computed(() => documents.value.length)

async function loadDocuments() {
  try {
    loading.value = true
    error.value = ''
    const offset = (currentPage.value - 1) * perPage.value
    const response = await listDocuments(perPage.value, offset)
    documents.value = response.data

    // Extract pagination metadata if available
    if (response.meta) {
      totalDocsCount.value = response.meta.total || documents.value.length
    } else {
      totalDocsCount.value = documents.value.length
    }
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to load documents:', err)
  } finally {
    loading.value = false
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
    loadDocuments()
  }
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--
    loadDocuments()
  }
}

async function createDocument() {
  if (!newDocId.value.trim()) return

  try {
    creating.value = true
    error.value = ''

    // Use findOrCreateDocument to handle URL, path, or ID
    const response = await documentService.findOrCreateDocument(newDocId.value.trim())

    // Navigate to document detail page with the returned docId
    await router.push({ name: 'admin-document', params: { docId: response.docId } })
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to create document:', err)
  } finally {
    creating.value = false
  }
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString('fr-FR', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

onMounted(() => {
  loadDocuments()
})
</script>

<template>
  <div class="relative min-h-[calc(100vh-4rem)]">
    <!-- Background decoration -->
    <div class="absolute inset-0 -z-10 overflow-hidden">
      <div class="absolute left-1/4 top-0 h-[400px] w-[400px] rounded-full bg-primary/5 blur-3xl"></div>
      <div class="absolute right-1/4 bottom-0 h-[400px] w-[400px] rounded-full bg-primary/5 blur-3xl"></div>
    </div>

    <!-- Main Content -->
    <main class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <!-- Page Header -->
      <div class="mb-8 flex items-start justify-between">
        <div>
          <h1 class="mb-2 text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
            {{ t('admin.title') }}
          </h1>
          <p class="text-lg text-muted-foreground">
            {{ t('admin.subtitle') }}
          </p>
        </div>
        <router-link :to="{ name: 'admin-webhooks' }">
          <Button variant="outline">
            <Webhook :size="16" class="mr-2" />
            {{ t('admin.webhooks.manage') }}
          </Button>
        </router-link>
      </div>

      <!-- Create Document Section -->
      <Card class="clay-card mb-8">
        <CardHeader>
          <CardTitle>{{ t('admin.documents.new') }}</CardTitle>
          <CardDescription>
            {{ t('admin.documents.newDescription') }}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form @submit.prevent="createDocument">
            <!-- Desktop layout -->
            <div class="hidden md:flex flex-row gap-4">
              <div class="flex-1">
                <label for="newDocId" class="block text-sm font-medium text-foreground mb-2">
                  {{ t('admin.documents.idLabel') }}
                </label>
                <Input
                  v-model="newDocId"
                  id="newDocId"
                  type="text"
                  required
                  :placeholder="t('admin.documents.idPlaceholder')"
                  class="w-full"
                />
                <p class="mt-1 text-xs text-muted-foreground">
                  {{ t('admin.documents.idHelper') }}
                </p>
              </div>
              <div class="pt-7">
                <Button type="submit" :disabled="!newDocId || creating">
                  <FileText :size="16" class="mr-2" v-if="!creating" />
                  <Loader2 :size="16" class="mr-2 animate-spin" v-else />
                  {{ creating ? t('admin.documentForm.creating') : t('common.confirm') }}
                </Button>
              </div>
            </div>

            <!-- Mobile layout with icon button -->
            <div class="md:hidden space-y-2">
              <label for="newDocIdMobile" class="block text-sm font-medium text-foreground">
                {{ t('admin.documents.idLabel') }}
              </label>
              <div class="flex gap-2">
                <Input
                  v-model="newDocId"
                  id="newDocIdMobile"
                  type="text"
                  required
                  :placeholder="t('admin.documents.idPlaceholder')"
                  class="flex-1"
                />
                <Button type="submit" size="icon" :disabled="!newDocId || creating" class="shrink-0">
                  <Loader2 :size="20" class="animate-spin" v-if="creating" />
                  <Plus :size="20" v-else />
                </Button>
              </div>
              <p class="text-xs text-muted-foreground">
                {{ t('admin.documents.idHelperShort') }}
              </p>
            </div>
          </form>
        </CardContent>
      </Card>

      <!-- Error Alert -->
      <Alert v-if="error" variant="destructive" class="mb-6 clay-card">
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>

      <!-- Loading State -->
      <div v-if="loading" class="flex flex-col items-center justify-center py-24">
        <Loader2 :size="48" class="animate-spin text-primary" />
        <p class="mt-4 text-muted-foreground">{{ t('admin.loading') }}</p>
      </div>

      <!-- Dashboard Content -->
      <div v-else>
        <!-- KPI Pills Mobile (compact horizontal full-width) -->
        <div class="md:hidden mb-6 grid grid-cols-3 gap-3">
          <div class="flex flex-col items-center justify-center gap-1 px-3 py-3 rounded-lg bg-primary/10 text-primary">
            <FileText :size="18" />
            <span class="text-xl font-bold">{{ totalDocuments }}</span>
            <span class="text-xs whitespace-nowrap">{{ t('admin.dashboard.stats.documents') }}</span>
          </div>
          <div class="flex flex-col items-center justify-center gap-1 px-3 py-3 rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400">
            <Users :size="18" />
            <span class="text-xl font-bold">{{ totalSigners }}</span>
            <span class="text-xs whitespace-nowrap">{{ t('admin.dashboard.stats.readers') }}</span>
          </div>
          <div class="flex flex-col items-center justify-center gap-1 px-3 py-3 rounded-lg bg-green-500/10 text-green-600 dark:text-green-400">
            <CheckCircle :size="18" />
            <span class="text-xl font-bold">{{ activeDocuments }}</span>
            <span class="text-xs whitespace-nowrap">{{ t('admin.dashboard.stats.active') }}</span>
          </div>
        </div>

        <!-- KPI Cards Desktop (unchanged) -->
        <div class="hidden md:grid mb-8 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          <!-- Total Documents -->
          <Card class="clay-card-hover">
            <CardContent class="pt-6">
              <div class="flex items-center space-x-4">
                <div class="rounded-lg bg-primary/10 p-3">
                  <FileText :size="24" class="text-primary" />
                </div>
                <div class="flex-1">
                  <p class="text-sm font-medium text-muted-foreground">{{ t('admin.dashboard.totalDocuments') }}</p>
                  <p class="text-2xl font-bold text-foreground">{{ totalDocuments }}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <!-- Total Expected Readers -->
          <Card class="clay-card-hover">
            <CardContent class="pt-6">
              <div class="flex items-center space-x-4">
                <div class="rounded-lg bg-blue-500/10 p-3">
                  <Users :size="24" class="text-blue-600 dark:text-blue-400" />
                </div>
                <div class="flex-1">
                  <p class="text-sm font-medium text-muted-foreground">{{ t('admin.dashboard.stats.expected') }}</p>
                  <p class="text-2xl font-bold text-foreground">{{ totalSigners }}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <!-- Active Documents -->
          <Card class="clay-card-hover">
            <CardContent class="pt-6">
              <div class="flex items-center space-x-4">
                <div class="rounded-lg bg-green-500/10 p-3">
                  <CheckCircle :size="24" class="text-green-600 dark:text-green-400" />
                </div>
                <div class="flex-1">
                  <p class="text-sm font-medium text-muted-foreground">{{ t('admin.documents.actions') }}</p>
                  <p class="text-2xl font-bold text-foreground">{{ activeDocuments }}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <!-- Documents Table -->
        <Card class="clay-card">
          <CardHeader>
            <CardTitle>{{ t('admin.documents.title') }}</CardTitle>
            <CardDescription class="mt-2">
              {{ t('admin.subtitle') }}
            </CardDescription>
          </CardHeader>

          <CardContent>
            <!-- Search Filter -->
            <div class="mb-6 relative">
              <Search :size="18" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <Input
                v-model="searchQuery"
                type="text"
                :placeholder="t('admin.documents.search')"
                class="pl-10"
              />
            </div>
            <!-- Desktop Table (hidden on mobile) -->
            <div v-if="filteredDocuments.length > 0" class="hidden md:block rounded-md border border-border/40">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{{ t('admin.documents.document') }}</TableHead>
                    <TableHead>{{ t('admin.documents.url') }}</TableHead>
                    <TableHead>{{ t('admin.documents.createdOn') }}</TableHead>
                    <TableHead>{{ t('admin.documents.by') }}</TableHead>
                    <TableHead class="text-right">{{ t('admin.documents.actions') }}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-for="doc in filteredDocuments" :key="doc.docId">
                    <TableCell>
                      <div class="space-y-1">
                        <div class="font-medium text-foreground">{{ doc.title }}</div>
                        <div class="text-xs font-mono text-muted-foreground">{{ doc.docId }}</div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <a
                        v-if="doc.url"
                        :href="doc.url"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="inline-flex items-center space-x-1 text-sm text-primary hover:underline"
                      >
                        <span class="max-w-[200px] truncate">{{ doc.url }}</span>
                        <ExternalLink :size="14" />
                      </a>
                      <span v-else class="text-xs text-muted-foreground">—</span>
                    </TableCell>
                    <TableCell class="text-muted-foreground">
                      {{ formatDate(doc.createdAt) }}
                    </TableCell>
                    <TableCell class="text-muted-foreground">
                      <span class="text-xs">{{ doc.createdBy }}</span>
                    </TableCell>
                    <TableCell class="text-right">
                      <router-link
                        :to="{ name: 'admin-document', params: { docId: doc.docId } }"
                      >
                        <Button variant="ghost" size="sm">
                          <Settings :size="16" class="mr-1" />
                          {{ t('admin.documents.manage') }}
                        </Button>
                      </router-link>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>

            <!-- Mobile Cards (hidden on desktop) -->
            <div v-if="filteredDocuments.length > 0" class="md:hidden space-y-4">
              <Card v-for="doc in filteredDocuments" :key="doc.docId" class="clay-card-hover">
                <CardContent class="p-4">
                  <!-- Document Title & ID -->
                  <div class="mb-3">
                    <h3 class="font-medium text-foreground text-base">{{ doc.title }}</h3>
                    <p class="text-xs font-mono text-muted-foreground mt-1">{{ doc.docId }}</p>
                  </div>

                  <!-- URL -->
                  <div v-if="doc.url" class="mb-3">
                    <a
                      :href="doc.url"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="inline-flex items-center space-x-1 text-sm text-primary hover:underline"
                    >
                      <ExternalLink :size="14" />
                      <span class="truncate max-w-[250px]">{{ doc.url }}</span>
                    </a>
                  </div>

                  <!-- Meta Info -->
                  <div class="flex flex-wrap items-center gap-3 text-sm text-muted-foreground mb-3">
                    <div class="flex items-center space-x-1">
                      <FileText :size="14" />
                      <span>{{ formatDate(doc.createdAt) }}</span>
                    </div>
                    <div class="flex items-center space-x-1">
                      <Users :size="14" />
                      <span class="text-xs">{{ doc.createdBy }}</span>
                    </div>
                  </div>

                  <!-- Actions -->
                  <div class="flex gap-2 pt-2 border-t border-border/40">
                    <router-link
                      :to="{ name: 'admin-document', params: { docId: doc.docId } }"
                      class="flex-1"
                    >
                      <Button variant="outline" size="sm" class="w-full">
                        <Settings :size="16" class="mr-2" />
                        {{ t('admin.documents.manage') }}
                      </Button>
                    </router-link>
                  </div>
                </CardContent>
              </Card>
            </div>

            <!-- Empty State -->
            <div v-else class="flex flex-col items-center justify-center py-12">
              <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
                <FileText :size="28" class="text-muted-foreground" />
              </div>
              <h3 class="mb-2 text-lg font-semibold text-foreground">
                {{ searchQuery ? t('admin.documents.noResults') : t('admin.documents.noDocuments') }}
              </h3>
              <p class="text-sm text-muted-foreground">
                {{ searchQuery ? t('admin.documents.tryAnotherSearch') : t('admin.documents.willAppear') }}
              </p>
            </div>

            <!-- Pagination -->
            <div v-if="filteredDocuments.length > 0 && !searchQuery && totalPages > 1" class="flex items-center justify-between mt-6 pt-4 border-t border-border/40">
              <!-- Mobile Pagination -->
              <div class="md:hidden flex items-center justify-between w-full">
                <Button
                  variant="outline"
                  size="sm"
                  :disabled="currentPage === 1"
                  @click="prevPage"
                >
                  {{ t('common.previous') }}
                </Button>
                <span class="text-sm text-muted-foreground">
                  {{ t('admin.documents.pagination.page', { current: currentPage, total: totalPages }) }}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  :disabled="currentPage >= totalPages"
                  @click="nextPage"
                >
                  {{ t('common.next') }}
                </Button>
              </div>

              <!-- Desktop Pagination -->
              <div class="hidden md:flex items-center justify-between w-full">
                <div class="text-sm text-muted-foreground">
                  {{ t('admin.documents.totalCount', totalDocuments) }}
                </div>
                <div class="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    :disabled="currentPage === 1"
                    @click="prevPage"
                  >
                    {{ t('common.previous') }}
                  </Button>
                  <span class="text-sm text-muted-foreground">
                    {{ t('admin.documents.pagination.pageOf', { current: currentPage, total: totalPages }) }}
                  </span>
                  <Button
                    variant="outline"
                    size="sm"
                    :disabled="currentPage >= totalPages"
                    @click="nextPage"
                  >
                    {{ t('common.next') }}
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  </div>
</template>
