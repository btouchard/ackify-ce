<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { usePageTitle } from '@/composables/usePageTitle'
import { listDocuments, type Document } from '@/services/admin'
import { documentService } from '@/services/documents'
import { extractError } from '@/services/http'
import {
  FileText,
  Users,
  CheckCircle,
  ExternalLink,
  Settings,
  Loader2,
  Plus,
  Search,
  Webhook,
  ChevronLeft,
  ChevronRight,
  AlertCircle,
  RefreshCw,
} from 'lucide-vue-next'

const router = useRouter()
const { t } = useI18n()
usePageTitle('admin.title')
const documents = ref<Document[]>([])
const loading = ref(true)        // Initial page load
const searching = ref(false)     // Search/pagination in progress
const error = ref('')
const newDocId = ref('')
const creating = ref(false)

// Pagination & Search (server-side)
const searchQuery = ref('')
const currentPage = ref(1)
const perPage = ref(20)
const totalDocsCount = ref(0)
let searchTimeout: ReturnType<typeof setTimeout> | null = null

// Computed
const totalPages = computed(() => Math.ceil(totalDocsCount.value / perPage.value) || 1)

// Computed KPIs
const totalDocuments = computed(() => totalDocsCount.value)
const totalSigners = computed(() => {
  // For now, return 0 as expectedSigners might not be in the Document type yet
  return 0
})
const activeDocuments = computed(() => documents.value.length)

async function loadDocuments(isInitialLoad = false) {
  try {
    // Use different loading state depending on context
    if (isInitialLoad) {
      loading.value = true
    } else {
      searching.value = true
    }

    error.value = ''
    const offset = (currentPage.value - 1) * perPage.value

    // Pass search query to API
    const response = await listDocuments(
      perPage.value,
      offset,
      searchQuery.value || undefined
    )

    documents.value = response.data

    // Extract pagination metadata
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
    searching.value = false
  }
}

// Debounced search: wait 300ms after user stops typing
function handleSearchInput() {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }

  searchTimeout = setTimeout(() => {
    // Reset to page 1 when searching
    currentPage.value = 1
    loadDocuments()
  }, 300)
}

// Watch searchQuery for changes
watch(searchQuery, () => {
  handleSearchInput()
})

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
  loadDocuments(true)  // Initial load with full loading screen
})
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)]">
    <!-- Main Content -->
    <main class="mx-auto max-w-6xl px-4 sm:px-6 py-6 sm:py-8">
      <!-- Page Header -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
        <div>
          <h1 class="text-2xl sm:text-3xl font-bold tracking-tight text-slate-900 dark:text-slate-50">
            {{ t('admin.title') }}
          </h1>
          <p class="mt-1 text-base text-slate-500 dark:text-slate-400">
            {{ t('admin.subtitle') }}
          </p>
        </div>
        <div class="flex items-center gap-3">
          <router-link :to="{ name: 'admin-webhooks' }">
            <button class="inline-flex items-center gap-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg px-4 py-2.5 text-sm hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
              <Webhook :size="16" />
              <span class="hidden sm:inline">{{ t('admin.webhooks.manage') }}</span>
            </button>
          </router-link>
          <button
            @click="loadDocuments()"
            :disabled="loading || searching"
            class="inline-flex items-center justify-center gap-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg px-4 py-2.5 text-sm hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors disabled:opacity-50"
          >
            <RefreshCw :size="16" :class="(loading || searching) ? 'animate-spin' : ''" />
            <span class="hidden sm:inline">{{ t('common.refresh') }}</span>
          </button>
        </div>
      </div>

      <!-- Error Alert -->
      <div v-if="error" class="mb-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4">
        <div class="flex items-start">
          <AlertCircle :size="20" class="mr-3 mt-0.5 text-red-600 dark:text-red-400 flex-shrink-0" />
          <div class="flex-1">
            <h3 class="font-medium text-red-900 dark:text-red-200">{{ t('common.error') }}</h3>
            <p class="mt-1 text-sm text-red-700 dark:text-red-300">{{ error }}</p>
          </div>
        </div>
      </div>

      <!-- Create Document Section -->
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 mb-8">
        <div class="flex items-start gap-4 mb-4">
          <div class="w-10 h-10 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
            <Plus :size="20" class="text-blue-600 dark:text-blue-400" />
          </div>
          <div>
            <h2 class="font-semibold text-slate-900 dark:text-slate-100">{{ t('admin.documents.new') }}</h2>
            <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('admin.documents.newDescription') }}</p>
          </div>
        </div>
        <form @submit.prevent="createDocument" class="flex flex-col sm:flex-row gap-3">
          <div class="flex-1">
            <input
              v-model="newDocId"
              type="text"
              required
              data-testid="admin-new-doc-input"
              :placeholder="t('admin.documents.idPlaceholder')"
              class="w-full px-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 dark:placeholder:text-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <p class="mt-1 text-xs text-slate-500 dark:text-slate-400 hidden sm:block">
              {{ t('admin.documents.idHelper') }}
            </p>
          </div>
          <button
            type="submit"
            :disabled="!newDocId || creating"
            data-testid="admin-create-doc-btn"
            class="trust-gradient text-white font-medium rounded-lg px-6 py-2.5 text-sm hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2 sm:self-start"
          >
            <FileText v-if="!creating" :size="16" />
            <Loader2 v-else :size="16" class="animate-spin" />
            {{ creating ? t('admin.documentForm.creating') : t('common.confirm') }}
          </button>
        </form>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex flex-col items-center justify-center py-24">
        <Loader2 :size="48" class="animate-spin text-blue-600" />
        <p class="mt-4 text-slate-500 dark:text-slate-400">{{ t('admin.loading') }}</p>
      </div>

      <!-- Dashboard Content -->
      <div v-else>
        <!-- KPI Pills Mobile -->
        <div class="md:hidden mb-6 grid grid-cols-3 gap-3">
          <div class="flex flex-col items-center justify-center gap-1 px-3 py-3 rounded-xl bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400">
            <FileText :size="18" />
            <span class="text-xl font-bold">{{ totalDocuments }}</span>
            <span class="text-xs whitespace-nowrap">{{ t('admin.dashboard.stats.documents') }}</span>
          </div>
          <div class="flex flex-col items-center justify-center gap-1 px-3 py-3 rounded-xl bg-blue-50 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400">
            <Users :size="18" />
            <span class="text-xl font-bold">{{ totalSigners }}</span>
            <span class="text-xs whitespace-nowrap">{{ t('admin.dashboard.stats.readers') }}</span>
          </div>
          <div class="flex flex-col items-center justify-center gap-1 px-3 py-3 rounded-xl bg-emerald-50 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400">
            <CheckCircle :size="18" />
            <span class="text-xl font-bold">{{ activeDocuments }}</span>
            <span class="text-xs whitespace-nowrap">{{ t('admin.dashboard.stats.active') }}</span>
          </div>
        </div>

        <!-- KPI Cards Desktop -->
        <div class="hidden md:grid mb-8 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          <!-- Total Documents -->
          <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-5 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-4">
              <div class="w-12 h-12 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center">
                <FileText :size="24" class="text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('admin.dashboard.totalDocuments') }}</p>
                <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ totalDocuments }}</p>
              </div>
            </div>
          </div>

          <!-- Total Expected Readers -->
          <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-5 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-4">
              <div class="w-12 h-12 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center">
                <Users :size="24" class="text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('admin.dashboard.stats.expected') }}</p>
                <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ totalSigners }}</p>
              </div>
            </div>
          </div>

          <!-- Active Documents -->
          <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-5 hover:shadow-md transition-shadow">
            <div class="flex items-center gap-4">
              <div class="w-12 h-12 rounded-xl bg-emerald-50 dark:bg-emerald-900/30 flex items-center justify-center">
                <CheckCircle :size="24" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <div>
                <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('admin.documents.actions') }}</p>
                <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ activeDocuments }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Documents List Card -->
        <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700">
          <!-- Header -->
          <div class="p-6 border-b border-slate-100 dark:border-slate-700">
            <div class="flex flex-col gap-4">
              <div>
                <h2 class="font-semibold text-slate-900 dark:text-slate-100">{{ t('admin.documents.title') }}</h2>
                <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('admin.subtitle') }}</p>
              </div>

              <!-- Search -->
              <div class="relative">
                <Search v-if="!searching" :size="18" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
                <Loader2 v-else :size="18" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 animate-spin" />
                <input
                  v-model="searchQuery"
                  type="text"
                  :placeholder="t('admin.documents.search')"
                  class="w-full pl-10 pr-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 dark:placeholder:text-slate-500 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>
          </div>

          <!-- Content -->
          <div class="p-6">
            <!-- Desktop Table -->
            <div v-if="documents.length > 0" class="hidden md:block overflow-x-auto">
              <table class="w-full">
                <thead>
                  <tr class="border-b border-slate-100 dark:border-slate-700">
                    <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                      {{ t('admin.documents.document') }}
                    </th>
                    <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                      {{ t('admin.documents.url') }}
                    </th>
                    <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                      {{ t('admin.documents.createdOn') }}
                    </th>
                    <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                      {{ t('admin.documents.by') }}
                    </th>
                    <th class="px-4 py-3 text-right text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                      {{ t('admin.documents.actions') }}
                    </th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
                  <tr
                    v-for="doc in documents"
                    :key="doc.docId"
                    class="hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors"
                  >
                    <td class="px-4 py-4">
                      <div class="space-y-1">
                        <div class="font-medium text-slate-900 dark:text-slate-100">{{ doc.title }}</div>
                        <div class="text-xs font-mono text-slate-500 dark:text-slate-400">{{ doc.docId }}</div>
                      </div>
                    </td>
                    <td class="px-4 py-4">
                      <a
                        v-if="doc.url"
                        :href="doc.url"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="inline-flex items-center gap-1 text-sm text-blue-600 dark:text-blue-400 hover:underline"
                      >
                        <span class="max-w-[200px] truncate">{{ doc.url }}</span>
                        <ExternalLink :size="14" />
                      </a>
                      <span v-else class="text-xs text-slate-400">—</span>
                    </td>
                    <td class="px-4 py-4 text-sm text-slate-500 dark:text-slate-400">
                      {{ formatDate(doc.createdAt) }}
                    </td>
                    <td class="px-4 py-4">
                      <span class="text-xs text-slate-500 dark:text-slate-400">{{ doc.createdBy }}</span>
                    </td>
                    <td class="px-4 py-4 text-right">
                      <router-link :to="{ name: 'admin-document', params: { docId: doc.docId } }">
                        <button class="inline-flex items-center gap-1 text-sm text-slate-600 dark:text-slate-300 hover:text-blue-600 dark:hover:text-blue-400 transition-colors">
                          <Settings :size="16" />
                          {{ t('admin.documents.manage') }}
                        </button>
                      </router-link>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Mobile Cards -->
            <div v-if="documents.length > 0" class="md:hidden space-y-4">
              <div
                v-for="doc in documents"
                :key="doc.docId"
                class="bg-slate-50 dark:bg-slate-700/50 rounded-xl p-4"
              >
                <!-- Document Title & ID -->
                <div class="mb-3">
                  <h3 class="font-medium text-slate-900 dark:text-slate-100">{{ doc.title }}</h3>
                  <p class="text-xs font-mono text-slate-500 dark:text-slate-400 mt-1">{{ doc.docId }}</p>
                </div>

                <!-- URL -->
                <div v-if="doc.url" class="mb-3">
                  <a
                    :href="doc.url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="inline-flex items-center gap-1 text-sm text-blue-600 dark:text-blue-400 hover:underline"
                  >
                    <ExternalLink :size="14" />
                    <span class="truncate max-w-[250px]">{{ doc.url }}</span>
                  </a>
                </div>

                <!-- Meta Info -->
                <div class="flex flex-wrap items-center gap-3 text-sm text-slate-500 dark:text-slate-400 mb-3">
                  <div class="flex items-center gap-1">
                    <FileText :size="14" />
                    <span>{{ formatDate(doc.createdAt) }}</span>
                  </div>
                  <div class="flex items-center gap-1">
                    <Users :size="14" />
                    <span class="text-xs">{{ doc.createdBy }}</span>
                  </div>
                </div>

                <!-- Actions -->
                <div class="flex gap-2 pt-3 border-t border-slate-200 dark:border-slate-600">
                  <router-link
                    :to="{ name: 'admin-document', params: { docId: doc.docId } }"
                    class="flex-1"
                  >
                    <button class="w-full inline-flex items-center justify-center gap-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg px-4 py-2 text-sm hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
                      <Settings :size="16" />
                      {{ t('admin.documents.manage') }}
                    </button>
                  </router-link>
                </div>
              </div>
            </div>

            <!-- Empty State -->
            <div v-if="documents.length === 0" class="text-center py-12">
              <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-slate-100 dark:bg-slate-700">
                <FileText :size="28" class="text-slate-400" />
              </div>
              <h3 class="mb-2 text-lg font-semibold text-slate-900 dark:text-slate-100">
                {{ searchQuery ? t('admin.documents.noResults') : t('admin.documents.noDocuments') }}
              </h3>
              <p class="text-sm text-slate-500 dark:text-slate-400">
                {{ searchQuery ? t('admin.documents.tryAnotherSearch') : t('admin.documents.willAppear') }}
              </p>
            </div>

            <!-- Pagination -->
            <div v-if="documents.length > 0 && totalPages > 1" class="flex items-center justify-between mt-6 pt-4 border-t border-slate-200 dark:border-slate-700">
              <div class="text-sm text-slate-500 dark:text-slate-400 hidden md:block">
                {{ t('admin.documents.totalCount', totalDocuments) }}
              </div>
              <div class="flex items-center gap-2 w-full md:w-auto justify-between md:justify-end">
                <button
                  :disabled="currentPage === 1"
                  @click="prevPage"
                  class="inline-flex items-center gap-1 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg px-3 py-2 text-sm hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <ChevronLeft :size="16" />
                  {{ t('common.previous') }}
                </button>
                <span class="text-sm text-slate-500 dark:text-slate-400">
                  {{ t('admin.documents.pagination.page', { current: currentPage, total: totalPages }) }}
                </span>
                <button
                  :disabled="currentPage >= totalPages"
                  @click="nextPage"
                  class="inline-flex items-center gap-1 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg px-3 py-2 text-sm hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {{ t('common.next') }}
                  <ChevronRight :size="16" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>
