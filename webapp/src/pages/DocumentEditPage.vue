<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePageTitle } from '@/composables/usePageTitle'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'
import {
  getDocumentStatus,
  updateDocumentMetadata,
  addExpectedSigner,
  removeExpectedSigner,
  sendReminders,
  deleteDocument,
  type DocumentStatus,
} from '@/services/admin'
import { extractError } from '@/services/http'
import { useConfigStore } from '@/stores/config'
import {
  ArrowLeft,
  Users,
  CheckCircle,
  Mail,
  Plus,
  Loader2,
  Copy,
  Clock,
  X,
  Trash2,
  Search,
  AlertCircle,
  ChevronRight,
  ExternalLink,
  Check,
  FileText,
  Eye,
  Download,
  ScrollText,
  ShieldCheck,
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const configStore = useConfigStore()
const { t, locale } = useI18n()

// Data
const docId = computed(() => route.params.id as string)
usePageTitle('documentEdit.title', { docId: docId.value })
const documentStatus = ref<DocumentStatus | null>(null)
const loading = ref(true)
const error = ref('')
const success = ref('')
const unauthorized = ref(false)

// Modals
const showAddSignersModal = ref(false)
const showDeleteConfirmModal = ref(false)
const showRemoveSignerModal = ref(false)
const showSendRemindersModal = ref(false)
const signerToRemove = ref('')
const remindersMessage = ref('')

// Metadata form
const metadataForm = ref<Partial<{
  title: string
  url: string
  checksum: string
  checksumAlgorithm: string
  description: string
  readMode: string
  allowDownload: boolean
  requireFullRead: boolean
  verifyChecksum: boolean
}>>({
  title: '',
  url: '',
  checksum: '',
  checksumAlgorithm: 'SHA-256',
  description: '',
  readMode: 'integrated',
  allowDownload: true,
  requireFullRead: false,
  verifyChecksum: true,
})
const savingMetadata = ref(false)

// Expected signers form
const signersEmails = ref('')
const addingSigners = ref(false)
const signerFilter = ref('')

// Reminders
const sendMode = ref<'all' | 'selected'>('all')
const selectedEmails = ref<string[]>([])
const sendingReminders = ref(false)

// Delete
const deletingDocument = ref(false)

// Copy feedback
const copied = ref(false)

// Computed
const shareLink = computed(() => {
  if (!documentStatus.value) return ''
  return documentStatus.value.shareLink
})

const stats = computed(() => documentStatus.value?.stats)
const reminderStats = computed(() => documentStatus.value?.reminderStats)
const smtpEnabled = computed(() => configStore.smtpEnabled)
const expectedSigners = computed(() => documentStatus.value?.expectedSigners || [])
const filteredSigners = computed(() => {
  const filter = signerFilter.value.toLowerCase().trim()
  if (!filter) return expectedSigners.value
  return expectedSigners.value.filter(signer =>
    signer.email.toLowerCase().includes(filter) ||
    (signer.name && signer.name.toLowerCase().includes(filter)) ||
    (signer.userName && signer.userName.toLowerCase().includes(filter))
  )
})
const documentMetadata = computed(() => documentStatus.value?.document)
const documentTitle = computed(() => documentMetadata.value?.title || docId.value)
const isStoredDocument = computed(() => !!documentMetadata.value?.storageKey)

// Methods
async function loadDocumentStatus() {
  try {
    loading.value = true
    error.value = ''
    unauthorized.value = false
    const response = await getDocumentStatus(docId.value)
    documentStatus.value = response.data

    // Check authorization
    if (!authStore.isAdmin && documentStatus.value.document?.createdBy !== authStore.user?.email) {
      unauthorized.value = true
      return
    }

    // Pre-fill metadata form if document exists
    if (documentStatus.value.document) {
      const doc = documentStatus.value.document
      metadataForm.value = {
        title: doc.title || '',
        url: doc.url || '',
        checksum: doc.checksum || '',
        checksumAlgorithm: doc.checksumAlgorithm || 'SHA-256',
        description: doc.description || '',
        readMode: doc.readMode || 'integrated',
        allowDownload: doc.allowDownload ?? true,
        requireFullRead: doc.requireFullRead ?? false,
        verifyChecksum: doc.verifyChecksum ?? true,
      }
    }
  } catch (err: any) {
    if (err?.response?.status === 403 || err?.response?.status === 404) {
      unauthorized.value = true
    } else {
      error.value = extractError(err)
    }
    console.error('Failed to load document status:', err)
  } finally {
    loading.value = false
  }
}

async function saveMetadata() {
  try {
    savingMetadata.value = true
    error.value = ''
    success.value = ''
    await updateDocumentMetadata(docId.value, metadataForm.value)
    success.value = t('documentEdit.metadataSaved')
    await loadDocumentStatus()
    setTimeout(() => (success.value = ''), 3000)
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to save metadata:', err)
  } finally {
    savingMetadata.value = false
  }
}

async function addSigners() {
  if (!signersEmails.value.trim()) return

  try {
    addingSigners.value = true
    error.value = ''
    success.value = ''

    const lines = signersEmails.value.split('\n').filter(l => l.trim())
    let addedCount = 0

    for (const line of lines) {
      const trimmed = line.trim()
      const match = trimmed.match(/^(.+?)\s*<(.+?)>$/)
      const email = match && match[2] ? match[2].trim() : trimmed
      const name = match && match[1] ? match[1].trim() : ''

      try {
        await addExpectedSigner(docId.value, { email, name })
        addedCount++
      } catch (err) {
        console.error(`Failed to add ${email}:`, err)
      }
    }

    showAddSignersModal.value = false
    signersEmails.value = ''
    success.value = t('documentEdit.signersAdded', { count: addedCount })
    await loadDocumentStatus()
    setTimeout(() => (success.value = ''), 3000)
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to add signers:', err)
  } finally {
    addingSigners.value = false
  }
}

function confirmRemoveSigner(email: string) {
  signerToRemove.value = email
  showRemoveSignerModal.value = true
}

async function removeSigner() {
  const email = signerToRemove.value
  if (!email) return

  try {
    error.value = ''
    success.value = ''
    await removeExpectedSigner(docId.value, email)
    success.value = t('documentEdit.signerRemoved', { email })
    showRemoveSignerModal.value = false
    signerToRemove.value = ''
    await loadDocumentStatus()
    setTimeout(() => (success.value = ''), 3000)
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to remove signer:', err)
  }
}

function confirmSendReminders() {
  remindersMessage.value =
    sendMode.value === 'all'
      ? t('documentEdit.confirmSendReminders', { count: reminderStats.value?.pendingCount || 0 })
      : t('documentEdit.confirmSendRemindersSelected', { count: selectedEmails.value.length })
  showSendRemindersModal.value = true
}

async function sendRemindersAction() {
  try {
    sendingReminders.value = true
    error.value = ''
    success.value = ''

    const normalizedLocale = locale.value.split('-')[0]
    const response = await sendReminders(
      docId.value,
      {
        emails: sendMode.value === 'selected' ? selectedEmails.value : undefined,
      },
      normalizedLocale
    )

    selectedEmails.value = []
    showSendRemindersModal.value = false

    if (response.data.result) {
      const result = response.data.result
      if (result.failed > 0) {
        success.value = t('documentEdit.remindersSentPartial', { sent: result.successfullySent, failed: result.failed })
      } else {
        success.value = t('documentEdit.remindersSentSuccess', { count: result.successfullySent })
      }
    } else {
      success.value = t('documentEdit.remindersSentGeneric')
    }

    await loadDocumentStatus()
    setTimeout(() => (success.value = ''), 3000)
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to send reminders:', err)
  } finally {
    sendingReminders.value = false
  }
}

async function copyToClipboard() {
  try {
    await navigator.clipboard.writeText(shareLink.value)
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  } catch (err) {
    console.error('Failed to copy:', err)
  }
}

function formatDate(dateString: string | undefined): string {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleDateString(locale.value, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function toggleEmailSelection(email: string) {
  const index = selectedEmails.value.indexOf(email)
  if (index > -1) {
    selectedEmails.value.splice(index, 1)
  } else {
    selectedEmails.value.push(email)
  }
}

async function handleDeleteDocument() {
  try {
    deletingDocument.value = true
    error.value = ''
    await deleteDocument(docId.value)
    showDeleteConfirmModal.value = false
    router.push('/documents')
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to delete document:', err)
    showDeleteConfirmModal.value = false
  } finally {
    deletingDocument.value = false
  }
}

onMounted(async () => {
  if (!authStore.initialized) {
    await authStore.checkAuth()
  }
  loadDocumentStatus()
})
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)]">
    <main class="mx-auto max-w-6xl px-4 sm:px-6 py-6 sm:py-8">
      <!-- Breadcrumb -->
      <nav class="flex items-center gap-2 text-sm mb-6">
        <router-link to="/documents" class="text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 transition-colors">
          {{ t('documentEdit.breadcrumb.myDocuments') }}
        </router-link>
        <ChevronRight :size="16" class="text-slate-300 dark:text-slate-600" />
        <span class="text-slate-900 dark:text-slate-100 font-medium truncate max-w-[200px]">{{ documentTitle }}</span>
      </nav>

      <!-- Page Header -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6 sm:mb-8">
        <div class="flex items-start gap-4">
          <div class="w-12 h-12 sm:w-14 sm:h-14 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
            <FileText class="w-6 h-6 sm:w-7 sm:h-7 text-blue-600 dark:text-blue-400" />
          </div>
          <div>
            <h1 class="text-xl sm:text-2xl font-bold text-slate-900 dark:text-white">{{ documentTitle }}</h1>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">{{ t('documentEdit.subtitle') }}</p>
          </div>
        </div>
        <button
          @click="router.push('/documents')"
          class="w-full sm:w-auto inline-flex items-center justify-center gap-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 font-medium rounded-lg px-4 py-2.5 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors min-h-[44px]"
        >
          <ArrowLeft :size="18" />
          {{ t('common.back') }}
        </button>
      </div>

      <!-- Alerts -->
      <div v-if="error" class="mb-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4">
        <div class="flex items-start">
          <AlertCircle :size="20" class="mr-3 mt-0.5 text-red-600 dark:text-red-400 flex-shrink-0" />
          <p class="text-sm text-red-700 dark:text-red-300">{{ error }}</p>
        </div>
      </div>

      <div v-if="success" class="mb-6 bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 rounded-xl p-4">
        <div class="flex items-start">
          <CheckCircle :size="20" class="mr-3 mt-0.5 text-emerald-600 dark:text-emerald-400 flex-shrink-0" />
          <p class="text-sm text-emerald-700 dark:text-emerald-300">{{ success }}</p>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex flex-col items-center justify-center py-24">
        <Loader2 :size="48" class="animate-spin text-blue-600" />
        <p class="mt-4 text-slate-500 dark:text-slate-400">{{ t('common.loading') }}</p>
      </div>

      <!-- Unauthorized -->
      <div v-else-if="unauthorized" class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-8 text-center">
        <div class="w-16 h-16 mx-auto bg-red-100 dark:bg-red-900/30 rounded-2xl flex items-center justify-center mb-4">
          <AlertCircle :size="32" class="text-red-600 dark:text-red-400" />
        </div>
        <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-100 mb-2">{{ t('documentEdit.unauthorized.title') }}</h2>
        <p class="text-slate-500 dark:text-slate-400 mb-6">{{ t('documentEdit.unauthorized.description') }}</p>
        <router-link
          to="/documents"
          class="inline-flex items-center gap-2 trust-gradient text-white font-medium rounded-lg px-6 py-2.5 hover:opacity-90 transition-opacity"
        >
          <ArrowLeft :size="16" />
          {{ t('documentEdit.unauthorized.backToDocuments') }}
        </router-link>
      </div>

      <!-- Document Content -->
      <div v-else-if="documentStatus" class="space-y-6">
        <!-- Share Link Card -->
        <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6">
          <h2 class="font-semibold text-slate-900 dark:text-slate-100 mb-4">{{ t('documentEdit.shareLink.title') }}</h2>
          <div class="flex flex-col sm:flex-row gap-3">
            <div class="flex-1 relative">
              <input
                type="text"
                :value="shareLink"
                readonly
                class="w-full px-4 py-2.5 pr-10 rounded-lg border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-700 text-slate-900 dark:text-slate-100 text-sm font-mono"
              />
              <a
                :href="shareLink"
                target="_blank"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300"
              >
                <ExternalLink :size="16" />
              </a>
            </div>
            <button
              @click="copyToClipboard"
              class="inline-flex items-center justify-center gap-2 bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg px-4 py-2.5 text-sm hover:bg-slate-50 dark:hover:bg-slate-600 transition-colors"
            >
              <Check v-if="copied" :size="16" class="text-emerald-500" />
              <Copy v-else :size="16" />
              {{ copied ? t('documentEdit.shareLink.copied') : t('documentEdit.shareLink.copy') }}
            </button>
          </div>
        </div>

        <!-- Stats Cards -->
        <div v-if="stats && stats.expectedCount > 0" class="grid gap-4 grid-cols-2 lg:grid-cols-3">
          <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
                <Users :size="20" class="text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <p class="text-xs text-slate-500 dark:text-slate-400">{{ t('documentEdit.stats.expected') }}</p>
                <p class="text-xl font-bold text-slate-900 dark:text-slate-100">{{ stats.expectedCount }}</p>
              </div>
            </div>
          </div>

          <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-emerald-50 dark:bg-emerald-900/30 flex items-center justify-center flex-shrink-0">
                <CheckCircle :size="20" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <div>
                <p class="text-xs text-slate-500 dark:text-slate-400">{{ t('documentEdit.stats.confirmed') }}</p>
                <p class="text-xl font-bold text-slate-900 dark:text-slate-100">{{ stats.signedCount }}</p>
              </div>
            </div>
          </div>

          <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-amber-50 dark:bg-amber-900/30 flex items-center justify-center flex-shrink-0">
                <Clock :size="20" class="text-amber-600 dark:text-amber-400" />
              </div>
              <div>
                <p class="text-xs text-slate-500 dark:text-slate-400">{{ t('documentEdit.stats.pending') }}</p>
                <p class="text-xl font-bold text-slate-900 dark:text-slate-100">{{ stats.pendingCount }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Document Metadata -->
        <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700">
          <div class="p-6 border-b border-slate-100 dark:border-slate-700">
            <h2 class="font-semibold text-slate-900 dark:text-slate-100">{{ t('documentEdit.metadata.title') }}</h2>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('documentEdit.metadata.description') }}</p>
          </div>
          <div class="p-6">
            <form @submit.prevent="saveMetadata" class="space-y-4">
              <div :class="['grid gap-4', isStoredDocument ? 'grid-cols-1' : 'grid-cols-1 md:grid-cols-2']">
                <div>
                  <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">{{ t('documentEdit.metadata.titleLabel') }}</label>
                  <input v-model="metadataForm.title" :placeholder="t('documentEdit.metadata.titlePlaceholder')" class="w-full px-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
                </div>
                <div v-if="!isStoredDocument">
                  <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">{{ t('documentEdit.metadata.urlLabel') }}</label>
                  <input v-model="metadataForm.url" type="url" :placeholder="t('documentEdit.metadata.urlPlaceholder')" class="w-full px-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
                </div>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-[1fr_auto] gap-4">
                <div>
                  <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">{{ t('documentEdit.metadata.checksumLabel') }}</label>
                  <input v-model="metadataForm.checksum" :placeholder="t('documentEdit.metadata.checksumPlaceholder')" class="w-full px-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
                </div>
                <div class="md:min-w-[140px]">
                  <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">{{ t('documentEdit.metadata.algorithmLabel') }}</label>
                  <select v-model="metadataForm.checksumAlgorithm" class="w-full px-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent">
                    <option value="SHA-256">SHA-256</option>
                    <option value="SHA-512">SHA-512</option>
                    <option value="MD5">MD5</option>
                  </select>
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">{{ t('documentEdit.metadata.descriptionLabel') }}</label>
                <textarea v-model="metadataForm.description" rows="4" :placeholder="t('documentEdit.metadata.descriptionPlaceholder')" class="w-full px-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"></textarea>
              </div>

              <!-- Reader Options -->
              <div class="pt-4 border-t border-slate-100 dark:border-slate-700">
                <h3 class="text-sm font-medium text-slate-700 dark:text-slate-300 mb-3">{{ t('documentCreateForm.readMode.label') }}</h3>

                <!-- Read mode -->
                <div class="flex gap-4 mb-4">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input type="radio" v-model="metadataForm.readMode" value="integrated" class="w-4 h-4 text-blue-600 border-slate-300 focus:ring-blue-500" />
                    <Eye class="w-4 h-4 text-slate-500" />
                    <span class="text-sm text-slate-700 dark:text-slate-300">{{ t('documentCreateForm.readMode.integrated') }}</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input type="radio" v-model="metadataForm.readMode" value="external" class="w-4 h-4 text-blue-600 border-slate-300 focus:ring-blue-500" />
                    <ExternalLink class="w-4 h-4 text-slate-500" />
                    <span class="text-sm text-slate-700 dark:text-slate-300">{{ t('documentCreateForm.readMode.external') }}</span>
                  </label>
                </div>

                <!-- Integrated mode options -->
                <div v-if="metadataForm.readMode === 'integrated'" class="pl-4 border-l-2 border-blue-200 dark:border-blue-800 space-y-3 mb-4">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" v-model="metadataForm.allowDownload" class="w-4 h-4 text-blue-600 border-slate-300 rounded focus:ring-blue-500" />
                    <Download class="w-4 h-4 text-slate-500" />
                    <span class="text-sm text-slate-700 dark:text-slate-300">{{ t('documentCreateForm.options.allowDownload') }}</span>
                  </label>
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" v-model="metadataForm.requireFullRead" class="w-4 h-4 text-blue-600 border-slate-300 rounded focus:ring-blue-500" />
                    <ScrollText class="w-4 h-4 text-slate-500" />
                    <span class="text-sm text-slate-700 dark:text-slate-300">{{ t('documentCreateForm.options.requireFullRead') }}</span>
                  </label>
                </div>

                <!-- Verify checksum -->
                <label class="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" v-model="metadataForm.verifyChecksum" class="w-4 h-4 text-blue-600 border-slate-300 rounded focus:ring-blue-500" />
                  <ShieldCheck class="w-4 h-4 text-slate-500" />
                  <span class="text-sm text-slate-700 dark:text-slate-300">{{ t('documentCreateForm.options.verifyChecksum') }}</span>
                </label>
              </div>

              <div v-if="documentMetadata" class="text-xs text-slate-500 dark:text-slate-400 pt-2 border-t border-slate-100 dark:border-slate-700">
                {{ t('documentEdit.metadata.createdBy', { by: documentMetadata.createdBy, date: formatDate(documentMetadata.createdAt) }) }}
              </div>

              <div class="flex justify-end">
                <button type="submit" :disabled="savingMetadata" class="trust-gradient text-white font-medium rounded-lg px-6 py-2.5 text-sm hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2">
                  <Loader2 v-if="savingMetadata" :size="16" class="animate-spin" />
                  {{ savingMetadata ? t('documentEdit.metadata.saving') : t('common.save') }}
                </button>
              </div>
            </form>
          </div>
        </div>

        <!-- Expected Readers -->
        <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700">
          <div class="p-6 border-b border-slate-100 dark:border-slate-700">
            <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
              <div>
                <h2 class="font-semibold text-slate-900 dark:text-slate-100">{{ t('documentEdit.readers.title') }}</h2>
                <p v-if="stats" class="text-sm text-slate-500 dark:text-slate-400">{{ stats.signedCount }} / {{ stats.expectedCount }} {{ t('documentEdit.readers.confirmed') }}</p>
              </div>
              <button @click="showAddSignersModal = true" class="trust-gradient text-white font-medium rounded-lg px-3 py-2 text-sm hover:opacity-90 transition-opacity inline-flex items-center gap-2">
                <Plus :size="16" />
                {{ t('documentEdit.readers.add') }}
              </button>
            </div>
          </div>
          <div class="p-6">
            <div v-if="expectedSigners.length > 0">
              <div class="relative mb-4">
                <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
                <input v-model="signerFilter" :placeholder="t('documentEdit.readers.filterPlaceholder')" class="w-full pl-9 pr-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
              </div>

              <!-- Table Desktop -->
              <div class="hidden md:block overflow-x-auto">
                <table class="w-full">
                  <thead>
                    <tr class="border-b border-slate-100 dark:border-slate-700">
                      <th class="px-4 py-3 w-10">
                        <input type="checkbox" class="rounded border-slate-300 dark:border-slate-600" @change="(e: any) => selectedEmails = e.target.checked ? expectedSigners.filter(s => !s.hasSigned).map(s => s.email) : []" />
                      </th>
                      <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('documentEdit.readers.reader') }}</th>
                      <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('documentEdit.readers.status') }}</th>
                      <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('documentEdit.readers.confirmedOn') }}</th>
                      <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('common.actions') }}</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
                    <tr v-for="signer in filteredSigners" :key="signer.email" class="hover:bg-slate-50 dark:hover:bg-slate-700/50">
                      <td class="px-4 py-3">
                        <input v-if="!signer.hasSigned" type="checkbox" class="rounded border-slate-300 dark:border-slate-600" :checked="selectedEmails.includes(signer.email)" @change="toggleEmailSelection(signer.email)" />
                      </td>
                      <td class="px-4 py-3">
                        <div>
                          <p class="font-medium text-slate-900 dark:text-slate-100">{{ signer.userName || signer.name || signer.email }}</p>
                          <p class="text-xs text-slate-500 dark:text-slate-400">{{ signer.email }}</p>
                        </div>
                      </td>
                      <td class="px-4 py-3">
                        <span :class="['inline-flex items-center gap-1 px-2.5 py-1 text-xs font-medium rounded-full', signer.hasSigned ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400']">
                          {{ signer.hasSigned ? t('documentEdit.readers.statusConfirmed') : t('documentEdit.readers.statusPending') }}
                        </span>
                      </td>
                      <td class="px-4 py-3 text-sm text-slate-500 dark:text-slate-400">
                        {{ signer.signedAt ? formatDate(signer.signedAt) : '-' }}
                      </td>
                      <td class="px-4 py-3">
                        <button v-if="!signer.hasSigned" @click="confirmRemoveSigner(signer.email)" class="p-1.5 rounded-md hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors">
                          <Trash2 :size="16" class="text-red-600 dark:text-red-400" />
                        </button>
                        <span v-else class="text-xs text-slate-400">-</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <!-- Cards Mobile -->
              <div class="md:hidden space-y-3">
                <div v-for="signer in filteredSigners" :key="signer.email" class="bg-slate-50 dark:bg-slate-700/50 rounded-xl p-4">
                  <div class="flex items-start justify-between mb-2">
                    <div class="flex items-start gap-3">
                      <input v-if="!signer.hasSigned" type="checkbox" class="mt-1 rounded border-slate-300 dark:border-slate-600" :checked="selectedEmails.includes(signer.email)" @change="toggleEmailSelection(signer.email)" />
                      <div>
                        <p class="font-medium text-slate-900 dark:text-slate-100">{{ signer.userName || signer.name || signer.email }}</p>
                        <p class="text-xs text-slate-500 dark:text-slate-400">{{ signer.email }}</p>
                      </div>
                    </div>
                    <span :class="['inline-flex items-center px-2 py-0.5 text-xs font-medium rounded-full', signer.hasSigned ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400']">
                      {{ signer.hasSigned ? t('documentEdit.readers.statusConfirmed') : t('documentEdit.readers.statusPending') }}
                    </span>
                  </div>
                  <div class="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
                    <span>{{ signer.signedAt ? formatDate(signer.signedAt) : '-' }}</span>
                    <button v-if="!signer.hasSigned" @click="confirmRemoveSigner(signer.email)" class="p-1 text-red-600 dark:text-red-400">
                      <Trash2 :size="14" />
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <div v-else class="text-center py-8">
              <Users :size="48" class="mx-auto mb-4 text-slate-300 dark:text-slate-600" />
              <p class="text-slate-500 dark:text-slate-400">{{ t('documentEdit.readers.noReaders') }}</p>
            </div>
          </div>
        </div>

        <!-- Email Reminders -->
        <div v-if="reminderStats && stats && stats.expectedCount > 0 && (smtpEnabled || reminderStats.totalSent > 0)" class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700">
          <div class="p-6 border-b border-slate-100 dark:border-slate-700">
            <h2 class="font-semibold text-slate-900 dark:text-slate-100">{{ t('documentEdit.reminders.title') }}</h2>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('documentEdit.reminders.description') }}</p>
          </div>
          <div class="p-6 space-y-6">
            <div class="grid gap-4 grid-cols-1 sm:grid-cols-3">
              <div class="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4">
                <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('documentEdit.reminders.sent') }}</p>
                <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ reminderStats.totalSent }}</p>
              </div>
              <div class="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4">
                <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('documentEdit.reminders.toRemind') }}</p>
                <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ reminderStats.pendingCount }}</p>
              </div>
              <div v-if="reminderStats.lastSentAt" class="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4">
                <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('documentEdit.reminders.lastSent') }}</p>
                <p class="text-sm font-bold text-slate-900 dark:text-slate-100">{{ formatDate(reminderStats.lastSentAt) }}</p>
              </div>
            </div>

            <div v-if="!smtpEnabled" class="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-xl p-4">
              <p class="text-sm text-amber-800 dark:text-amber-200">{{ t('documentEdit.reminders.emailDisabled') }}</p>
            </div>

            <div v-if="smtpEnabled" class="space-y-4">
              <div class="space-y-2">
                <label class="flex items-center space-x-2 cursor-pointer">
                  <input type="radio" v-model="sendMode" value="all" class="text-blue-600 focus:ring-blue-500" />
                  <span class="text-sm text-slate-700 dark:text-slate-300">{{ t('documentEdit.reminders.sendToAll', { count: reminderStats.pendingCount }) }}</span>
                </label>
                <label class="flex items-center space-x-2 cursor-pointer">
                  <input type="radio" v-model="sendMode" value="selected" class="text-blue-600 focus:ring-blue-500" />
                  <span class="text-sm text-slate-700 dark:text-slate-300">{{ t('documentEdit.reminders.sendToSelected', { count: selectedEmails.length }) }}</span>
                </label>
              </div>
              <button @click="confirmSendReminders" :disabled="sendingReminders || (sendMode === 'selected' && selectedEmails.length === 0)" class="trust-gradient text-white font-medium rounded-lg px-4 py-2.5 text-sm hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2">
                <Mail :size="16" />
                {{ sendingReminders ? t('documentEdit.reminders.sending') : t('documentEdit.reminders.send') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Danger Zone -->
        <div class="bg-white dark:bg-slate-800 rounded-xl border border-red-200 dark:border-red-800/50">
          <div class="p-6 border-b border-red-100 dark:border-red-800/30">
            <h2 class="font-semibold text-red-600 dark:text-red-400">{{ t('documentEdit.danger.title') }}</h2>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">{{ t('documentEdit.danger.description') }}</p>
          </div>
          <div class="p-6">
            <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 p-4 bg-red-50 dark:bg-red-900/20 rounded-xl">
              <div>
                <h3 class="font-semibold text-slate-900 dark:text-slate-100 mb-1">{{ t('documentEdit.danger.deleteDocument') }}</h3>
                <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('documentEdit.danger.deleteDescription') }}</p>
              </div>
              <button @click="showDeleteConfirmModal = true" class="inline-flex items-center justify-center gap-2 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg px-4 py-2.5 text-sm transition-colors flex-shrink-0">
                <Trash2 :size="16" />
                {{ t('common.delete') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Add Signers Modal -->
    <div v-if="showAddSignersModal" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="showAddSignersModal = false">
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 max-w-2xl w-full max-h-[90vh] overflow-auto">
        <div class="p-6 border-b border-slate-100 dark:border-slate-700 flex items-center justify-between">
          <h2 class="font-semibold text-slate-900 dark:text-slate-100">{{ t('documentEdit.addSigners.title') }}</h2>
          <button @click="showAddSignersModal = false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
            <X :size="20" class="text-slate-400" />
          </button>
        </div>
        <div class="p-6">
          <form @submit.prevent="addSigners" class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1.5">{{ t('documentEdit.addSigners.emailsLabel') }}</label>
              <textarea v-model="signersEmails" rows="8" :placeholder="t('documentEdit.addSigners.emailsPlaceholder')" class="w-full px-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"></textarea>
              <p class="text-xs text-slate-500 dark:text-slate-400 mt-2">{{ t('documentEdit.addSigners.emailsHelper') }}</p>
            </div>
            <div class="flex justify-end space-x-3">
              <button type="button" @click="showAddSignersModal = false" class="bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg px-4 py-2.5 text-sm hover:bg-slate-50 dark:hover:bg-slate-600 transition-colors">{{ t('common.cancel') }}</button>
              <button type="submit" :disabled="addingSigners || !signersEmails.trim()" class="trust-gradient text-white font-medium rounded-lg px-4 py-2.5 text-sm hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2">
                <Loader2 v-if="addingSigners" :size="16" class="animate-spin" />
                {{ addingSigners ? t('documentEdit.addSigners.adding') : t('documentEdit.addSigners.add') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteConfirmModal" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="showDeleteConfirmModal = false">
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-red-200 dark:border-red-800 max-w-md w-full">
        <div class="p-6 border-b border-red-100 dark:border-red-800/30 flex items-center justify-between">
          <h2 class="font-semibold text-red-600 dark:text-red-400">{{ t('documentEdit.deleteConfirm.title') }}</h2>
          <button @click="showDeleteConfirmModal = false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
            <X :size="20" class="text-slate-400" />
          </button>
        </div>
        <div class="p-6 space-y-4">
          <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4">
            <p class="font-semibold text-red-900 dark:text-red-200 mb-2">{{ t('documentEdit.deleteConfirm.warning') }}</p>
            <p class="text-sm text-red-700 dark:text-red-300">{{ t('documentEdit.deleteConfirm.message') }}</p>
          </div>

          <div class="flex justify-end space-x-3 pt-4">
            <button type="button" @click="showDeleteConfirmModal = false" class="bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg px-4 py-2.5 text-sm hover:bg-slate-50 dark:hover:bg-slate-600 transition-colors">{{ t('common.cancel') }}</button>
            <button @click="handleDeleteDocument" :disabled="deletingDocument" class="bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg px-4 py-2.5 text-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2">
              <Trash2 v-if="!deletingDocument" :size="16" />
              <Loader2 v-else :size="16" class="animate-spin" />
              {{ deletingDocument ? t('documentEdit.deleteConfirm.deleting') : t('documentEdit.deleteConfirm.confirm') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Remove Signer Modal -->
    <div v-if="showRemoveSignerModal" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="showRemoveSignerModal = false">
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 max-w-md w-full">
        <div class="p-6 border-b border-slate-100 dark:border-slate-700 flex items-center justify-between">
          <h2 class="font-semibold text-slate-900 dark:text-slate-100">{{ t('documentEdit.removeSigner.title') }}</h2>
          <button @click="showRemoveSignerModal = false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
            <X :size="20" class="text-slate-400" />
          </button>
        </div>
        <div class="p-6 space-y-4">
          <p class="text-sm text-slate-600 dark:text-slate-400">{{ t('documentEdit.removeSigner.message', { email: signerToRemove }) }}</p>
          <div class="flex justify-end space-x-3 pt-4">
            <button type="button" @click="showRemoveSignerModal = false" class="bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg px-4 py-2.5 text-sm hover:bg-slate-50 dark:hover:bg-slate-600 transition-colors">{{ t('common.cancel') }}</button>
            <button @click="removeSigner" class="bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg px-4 py-2.5 text-sm transition-colors">{{ t('common.delete') }}</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Send Reminders Modal -->
    <div v-if="showSendRemindersModal" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="showSendRemindersModal = false">
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 max-w-md w-full">
        <div class="p-6 border-b border-slate-100 dark:border-slate-700 flex items-center justify-between">
          <h2 class="font-semibold text-slate-900 dark:text-slate-100">{{ t('documentEdit.sendReminders.title') }}</h2>
          <button @click="showSendRemindersModal = false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
            <X :size="20" class="text-slate-400" />
          </button>
        </div>
        <div class="p-6 space-y-4">
          <p class="text-sm text-slate-600 dark:text-slate-400">{{ remindersMessage }}</p>
          <div class="flex justify-end space-x-3 pt-4">
            <button type="button" @click="showSendRemindersModal = false" class="bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg px-4 py-2.5 text-sm hover:bg-slate-50 dark:hover:bg-slate-600 transition-colors">{{ t('common.cancel') }}</button>
            <button @click="sendRemindersAction" :disabled="sendingReminders" class="trust-gradient text-white font-medium rounded-lg px-4 py-2.5 text-sm hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2">
              <Loader2 v-if="sendingReminders" :size="16" class="animate-spin" />
              {{ t('common.confirm') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
