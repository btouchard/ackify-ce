<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePageTitle } from '@/composables/usePageTitle'
import { useI18n } from 'vue-i18n'
import {
  getDocumentStatus,
  updateDocumentMetadata,
  addExpectedSigner,
  removeExpectedSigner,
  sendReminders,
  deleteDocument,
  previewCSVSigners,
  importSigners,
  type DocumentStatus,
  type CSVPreviewResult,
  type CSVSignerEntry,
} from '@/services/admin'
import { extractError } from '@/services/http'
import {
  ArrowLeft,
  Users,
  CheckCircle,
  Mail,
  Shield,
  Plus,
  Loader2,
  Copy,
  Clock,
  X,
  Trash2,
  Upload,
  AlertTriangle,
  FileCheck,
  FileX,
  Search,
} from 'lucide-vue-next'
import Card from '@/components/ui/Card.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import CardTitle from '@/components/ui/CardTitle.vue'
import CardDescription from '@/components/ui/CardDescription.vue'
import CardContent from '@/components/ui/CardContent.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Textarea from '@/components/ui/Textarea.vue'
import Alert from '@/components/ui/Alert.vue'
import AlertDescription from '@/components/ui/AlertDescription.vue'
import Badge from '@/components/ui/Badge.vue'
import Table from '@/components/ui/table/Table.vue'
import TableHeader from '@/components/ui/table/TableHeader.vue'
import TableBody from '@/components/ui/table/TableBody.vue'
import TableRow from '@/components/ui/table/TableRow.vue'
import TableHead from '@/components/ui/table/TableHead.vue'
import TableCell from '@/components/ui/table/TableCell.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'

const route = useRoute()
const router = useRouter()
const { t, locale } = useI18n()

// Data
const docId = computed(() => route.params.docId as string)
usePageTitle('admin.documentDetail.title', { docId: docId.value })
const documentStatus = ref<DocumentStatus | null>(null)
const loading = ref(true)
const error = ref('')
const success = ref('')

// Modals
const showAddSignersModal = ref(false)
const showDeleteConfirmModal = ref(false)
const showMetadataWarningModal = ref(false)
const showRemoveSignerModal = ref(false)
const showSendRemindersModal = ref(false)
const showImportCSVModal = ref(false)
const signerToRemove = ref('')
const remindersMessage = ref('')

// CSV Import
const csvFile = ref<File | null>(null)
const csvPreview = ref<CSVPreviewResult | null>(null)
const analyzingCSV = ref(false)
const importingCSV = ref(false)
const csvError = ref('')

// Metadata form
const metadataForm = ref<Partial<{
  title: string
  url: string
  checksum: string
  checksumAlgorithm: string
  description: string
}>>({
  title: '',
  url: '',
  checksum: '',
  checksumAlgorithm: 'SHA-256',
  description: '',
})
const originalMetadata = ref<Partial<{
  title: string
  url: string
  checksum: string
  checksumAlgorithm: string
  description: string
}>>({})
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

// Computed
const shareLink = computed(() => {
  if (!documentStatus.value) return ''
  return documentStatus.value.shareLink
})

const stats = computed(() => documentStatus.value?.stats)
const reminderStats = computed(() => documentStatus.value?.reminderStats)
const smtpEnabled = computed(() => (window as any).ACKIFY_SMTP_ENABLED || false)
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
const unexpectedSignatures = computed(() => documentStatus.value?.unexpectedSignatures || [])
const documentMetadata = computed(() => documentStatus.value?.document)

// Methods
async function loadDocumentStatus() {
  try {
    loading.value = true
    error.value = ''
    const response = await getDocumentStatus(docId.value)
    documentStatus.value = response.data

    // Pre-fill metadata form if document exists
    if (documentStatus.value.document) {
      const doc = documentStatus.value.document
      const metadata = {
        title: doc.title || '',
        url: doc.url || '',
        checksum: doc.checksum || '',
        checksumAlgorithm: doc.checksumAlgorithm || 'SHA-256',
        description: doc.description || '',
      }
      metadataForm.value = { ...metadata }
      originalMetadata.value = { ...metadata }
    }
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to load document status:', err)
  } finally {
    loading.value = false
  }
}

function hasCriticalFieldsChanged(): boolean {
  // Check if critical fields (url, checksum, checksumAlgorithm, description) have changed
  return (
    metadataForm.value.url !== originalMetadata.value.url ||
    metadataForm.value.checksum !== originalMetadata.value.checksum ||
    metadataForm.value.checksumAlgorithm !== originalMetadata.value.checksumAlgorithm ||
    metadataForm.value.description !== originalMetadata.value.description
  )
}

function handleSaveMetadata() {
  // Check if document has signatures (both expected and unexpected) and critical fields changed
  const expectedSignaturesCount = stats.value?.signedCount || 0
  const unexpectedSignaturesCount = unexpectedSignatures.value?.length || 0
  const totalSignatures = expectedSignaturesCount + unexpectedSignaturesCount
  const hasSignatures = totalSignatures > 0
  const criticalFieldsChanged = hasCriticalFieldsChanged()

  if (hasSignatures && criticalFieldsChanged) {
    // Show warning modal
    showMetadataWarningModal.value = true
  } else {
    // Save directly
    saveMetadata()
  }
}

async function saveMetadata() {
  try {
    savingMetadata.value = true
    error.value = ''
    success.value = ''
    showMetadataWarningModal.value = false
    await updateDocumentMetadata(docId.value, metadataForm.value)
    success.value = t('admin.documentDetail.metadataSaved')
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

    // Parse emails - support "Name <email>" or "email" format
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
    success.value = t('admin.documentDetail.signersAdded', { count: addedCount })
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
    success.value = t('admin.documentDetail.signerRemoved', { email })
    showRemoveSignerModal.value = false
    signerToRemove.value = ''
    await loadDocumentStatus()
    setTimeout(() => (success.value = ''), 3000)
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to remove signer:', err)
  }
}

function cancelRemoveSigner() {
  showRemoveSignerModal.value = false
  signerToRemove.value = ''
}

function confirmSendReminders() {
  remindersMessage.value =
    sendMode.value === 'all'
      ? t('admin.documentDetail.confirmSendReminders', { count: reminderStats.value?.pendingCount || 0 })
      : t('admin.documentDetail.confirmSendRemindersSelected', { count: selectedEmails.value.length })
  showSendRemindersModal.value = true
}

async function sendRemindersAction() {
  try {
    sendingReminders.value = true
    error.value = ''
    success.value = ''

    // Normalize locale to base language code (it-IT -> it, en-US -> en, etc.)
    const normalizedLocale = locale.value.split('-')[0]
    console.log('Sending reminders with locale:', normalizedLocale, '(from', locale.value, ')')

    const response = await sendReminders(
      docId.value,
      {
        emails: sendMode.value === 'selected' ? selectedEmails.value : undefined,
      },
      normalizedLocale // Pass normalized locale
    )

    selectedEmails.value = []
    showSendRemindersModal.value = false

    if (response.data.result) {
      const result = response.data.result
      if (result.failed > 0) {
        success.value = t('admin.documentDetail.remindersSentPartial', { sent: result.successfullySent, failed: result.failed })
      } else {
        success.value = t('admin.documentDetail.remindersSentSuccess', { count: result.successfullySent })
      }
    } else {
      success.value = t('admin.documentDetail.remindersSentGeneric')
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

function cancelSendReminders() {
  showSendRemindersModal.value = false
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  success.value = t('admin.documentDetail.copiedToClipboard')
  setTimeout(() => (success.value = ''), 2000)
}

function formatDate(dateString: string | undefined): string {
  if (!dateString) return 'N/A'
  const date = new Date(dateString)
  return date.toLocaleDateString('fr-FR', {
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
    // Redirect to admin dashboard
    router.push('/admin')
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to delete document:', err)
    showDeleteConfirmModal.value = false
  } finally {
    deletingDocument.value = false
  }
}

// CSV Import functions
function openImportCSVModal() {
  csvFile.value = null
  csvPreview.value = null
  csvError.value = ''
  showImportCSVModal.value = true
}

function handleCSVFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    csvFile.value = target.files[0]
    csvPreview.value = null
    csvError.value = ''
  }
}

async function analyzeCSV() {
  if (!csvFile.value) return

  try {
    analyzingCSV.value = true
    csvError.value = ''
    const response = await previewCSVSigners(docId.value, csvFile.value)
    csvPreview.value = response.data
  } catch (err) {
    csvError.value = extractError(err)
    console.error('Failed to analyze CSV:', err)
  } finally {
    analyzingCSV.value = false
  }
}

function getSignerStatus(signer: CSVSignerEntry): 'valid' | 'exists' {
  if (!csvPreview.value) return 'valid'
  return csvPreview.value.existingEmails.includes(signer.email) ? 'exists' : 'valid'
}

const signersToImport = computed(() => {
  if (!csvPreview.value) return []
  return csvPreview.value.signers.filter(
    s => !csvPreview.value!.existingEmails.includes(s.email)
  )
})

async function confirmImportCSV() {
  if (!csvPreview.value || signersToImport.value.length === 0) return

  try {
    importingCSV.value = true
    csvError.value = ''

    const signersData = signersToImport.value.map(s => ({
      email: s.email,
      name: s.name
    }))

    const response = await importSigners(docId.value, signersData)

    showImportCSVModal.value = false
    csvFile.value = null
    csvPreview.value = null

    success.value = t('admin.documentDetail.csvImportSuccess', {
      imported: response.data.imported,
      skipped: response.data.skipped
    })
    await loadDocumentStatus()
    setTimeout(() => (success.value = ''), 3000)
  } catch (err) {
    csvError.value = extractError(err)
    console.error('Failed to import signers:', err)
  } finally {
    importingCSV.value = false
  }
}

function closeImportCSVModal() {
  showImportCSVModal.value = false
  csvFile.value = null
  csvPreview.value = null
  csvError.value = ''
}

onMounted(() => {
  loadDocumentStatus()
})
</script>

<template>
  <div class="relative min-h-[calc(100vh-4rem)]">
    <div class="absolute inset-0 -z-10 overflow-hidden">
      <div class="absolute left-1/4 top-0 h-[400px] w-[400px] rounded-full bg-primary/5 blur-3xl"></div>
      <div class="absolute right-1/4 bottom-0 h-[400px] w-[400px] rounded-full bg-primary/5 blur-3xl"></div>
    </div>

    <main class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <!-- Header -->
      <div class="mb-8">
        <div class="flex items-center space-x-3 mb-2">
          <Button variant="ghost" size="icon" @click="router.push('/admin')" :aria-label="t('admin.documentDetail.back')">
            <ArrowLeft :size="20" />
          </Button>
          <h1 class="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
            {{ t('admin.documentDetail.title') }} {{ docId }}
          </h1>
        </div>
        <div class="flex items-center gap-3 ml-14">
          <p class="text-sm text-muted-foreground font-mono">{{ shareLink }}</p>
          <Button @click="copyToClipboard(shareLink)" variant="ghost" size="icon" :aria-label="t('signatureList.copy')">
            <Copy :size="16" />
          </Button>
        </div>
      </div>

      <!-- Alerts -->
      <Alert v-if="error" variant="destructive" class="mb-6 clay-card">
        <AlertDescription>{{ error }}</AlertDescription>
      </Alert>

      <Alert v-if="success" class="mb-6 clay-card bg-green-50 border-green-200 dark:bg-green-900/20">
        <AlertDescription class="text-green-800 dark:text-green-200">{{ success }}</AlertDescription>
      </Alert>

      <!-- Loading -->
      <div v-if="loading" class="flex flex-col items-center justify-center py-24">
        <Loader2 :size="48" class="animate-spin text-primary" />
        <p class="mt-4 text-muted-foreground">{{ t('common.loading') }}</p>
      </div>

      <!-- Content -->
      <div v-else-if="documentStatus" class="space-y-8">
        <!-- Stats Cards -->
        <div v-if="stats && stats.expectedCount > 0" class="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
          <Card class="clay-card-hover">
            <CardContent class="pt-6">
              <div class="flex items-center space-x-4">
                <div class="rounded-lg bg-blue-500/10 p-3">
                  <Users :size="24" class="text-blue-600" />
                </div>
                <div>
                  <p class="text-sm font-medium text-muted-foreground">{{ t('admin.dashboard.stats.expected') }}</p>
                  <p class="text-2xl font-bold text-foreground">{{ stats.expectedCount }}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card class="clay-card-hover">
            <CardContent class="pt-6">
              <div class="flex items-center space-x-4">
                <div class="rounded-lg bg-green-500/10 p-3">
                  <CheckCircle :size="24" class="text-green-600" />
                </div>
                <div>
                  <p class="text-sm font-medium text-muted-foreground">{{ t('admin.dashboard.stats.signed') }}</p>
                  <p class="text-2xl font-bold text-foreground">{{ stats.signedCount }}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card class="clay-card-hover">
            <CardContent class="pt-6">
              <div class="flex items-center space-x-4">
                <div class="rounded-lg bg-orange-500/10 p-3">
                  <Clock :size="24" class="text-orange-600" />
                </div>
                <div>
                  <p class="text-sm font-medium text-muted-foreground">{{ t('admin.dashboard.stats.pending') }}</p>
                  <p class="text-2xl font-bold text-foreground">{{ stats.pendingCount }}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card class="clay-card-hover">
            <CardContent class="pt-6">
              <div class="flex items-center space-x-4">
                <div class="rounded-lg bg-purple-500/10 p-3">
                  <Shield :size="24" class="text-purple-600" />
                </div>
                <div>
                  <p class="text-sm font-medium text-muted-foreground">{{ t('admin.dashboard.stats.completion') }}</p>
                  <p class="text-2xl font-bold text-foreground">{{ Math.round(stats.completionRate) }}%</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>

        <!-- Document Metadata -->
        <Card class="clay-card">
          <CardHeader>
            <div>
              <CardTitle>{{ t('admin.documentDetail.metadata') }}</CardTitle>
              <CardDescription>{{ t('admin.documentDetail.metadataDescription') }}</CardDescription>
            </div>
          </CardHeader>
          <CardContent>
            <form @submit.prevent="handleSaveMetadata" class="space-y-4">
              <!-- Titre et URL côte à côte -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium mb-2">{{ t('admin.documentDetail.titleLabel') }}</label>
                  <Input v-model="metadataForm.title" :placeholder="t('admin.documentDetail.titlePlaceholder')" />
                </div>
                <div>
                  <label class="block text-sm font-medium mb-2">{{ t('admin.documentDetail.urlLabel') }}</label>
                  <Input v-model="metadataForm.url" type="url" :placeholder="t('admin.documentDetail.urlPlaceholder')" />
                </div>
              </div>

              <!-- Checksum et Algorithme côte à côte -->
              <div class="grid grid-cols-1 md:grid-cols-[1fr_auto] gap-4">
                <div>
                  <label class="block text-sm font-medium mb-2">{{ t('admin.documentDetail.checksumLabel') }}</label>
                  <Input v-model="metadataForm.checksum" :placeholder="t('admin.documentDetail.checksumPlaceholder')" class="font-mono text-sm" />
                </div>
                <div class="md:min-w-[140px]">
                  <label class="block text-sm font-medium mb-2">{{ t('admin.documentDetail.algorithmLabel') }}</label>
                  <select v-model="metadataForm.checksumAlgorithm" class="flex h-10 w-full rounded-md clay-input px-3 py-2 text-sm">
                    <option value="SHA-256">SHA-256</option>
                    <option value="SHA-512">SHA-512</option>
                    <option value="MD5">MD5</option>
                  </select>
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium mb-2">{{ t('admin.documentDetail.descriptionLabel') }}</label>
                <Textarea v-model="metadataForm.description" :rows="4" :placeholder="t('admin.documentDetail.descriptionPlaceholder')" />
              </div>
              <div v-if="documentMetadata" class="text-xs text-muted-foreground pt-2 border-t">
                {{ t('admin.documentDetail.createdBy', { by: documentMetadata.createdBy, date: formatDate(documentMetadata.createdAt) }) }}
              </div>
              <div class="flex justify-end">
                <Button type="submit" :disabled="savingMetadata">
                  {{ savingMetadata ? t('admin.documentDetail.saving') : t('common.save') }}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        <!-- Expected Readers -->
        <Card class="clay-card">
          <CardHeader>
            <div class="flex items-center justify-between">
              <div>
                <CardTitle>{{ t('admin.documentDetail.readers') }}</CardTitle>
                <CardDescription v-if="stats">{{ stats.signedCount }} / {{ stats.expectedCount }} {{ t('admin.dashboard.stats.signed').toLowerCase() }}</CardDescription>
              </div>
              <div class="flex gap-2">
                <Button @click="openImportCSVModal" size="sm" variant="outline">
                  <Upload :size="16" class="mr-2" />
                  {{ t('admin.documentDetail.importCSV') }}
                </Button>
                <Button @click="showAddSignersModal = true" size="sm">
                  <Plus :size="16" class="mr-2" />
                  {{ t('admin.documentDetail.addButton') }}
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <!-- Filter + Expected Signers Table -->
            <div v-if="expectedSigners.length > 0">
              <!-- Filter -->
              <div class="relative mb-4">
                <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground z-10 pointer-events-none" />
                <Input
                  v-model="signerFilter"
                  :placeholder="t('admin.documentDetail.filterPlaceholder')"
                  class="pl-9"
                  name="ackify-signer-filter"
                  autocomplete="off"
                  data-1p-ignore
                  data-lpignore="true"
                />
              </div>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>
                      <input type="checkbox" class="rounded"
                             @change="(e: any) => selectedEmails = e.target.checked ? expectedSigners.filter(s => !s.hasSigned).map(s => s.email) : []" />
                    </TableHead>
                    <TableHead>{{ t('admin.documentDetail.reader') }}</TableHead>
                    <TableHead>{{ t('admin.documentDetail.status') }}</TableHead>
                    <TableHead>{{ t('admin.documentDetail.confirmedOn') }}</TableHead>
                    <TableHead>{{ t('common.actions') }}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-for="signer in filteredSigners" :key="signer.email">
                    <TableCell>
                      <input v-if="!signer.hasSigned" type="checkbox" class="rounded"
                             :checked="selectedEmails.includes(signer.email)"
                             @change="toggleEmailSelection(signer.email)" />
                    </TableCell>
                    <TableCell>
                      <div class="space-y-1">
                        <p class="font-medium">{{ signer.userName || signer.name || signer.email }}</p>
                        <p class="text-xs text-muted-foreground">{{ signer.email }}</p>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge :variant="signer.hasSigned ? 'default' : 'secondary'">
                        {{ signer.hasSigned ? t('admin.documentDetail.statusConfirmed') : t('admin.documentDetail.statusPending') }}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {{ signer.signedAt ? formatDate(signer.signedAt) : '-' }}
                    </TableCell>
                    <TableCell>
                      <Button v-if="!signer.hasSigned" @click="confirmRemoveSigner(signer.email)" variant="ghost" size="sm">
                        <Trash2 :size="14" class="text-destructive" />
                      </Button>
                      <span v-else class="text-xs text-muted-foreground">-</span>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>
            <div v-else class="text-center py-8 text-muted-foreground">
              <Users :size="48" class="mx-auto mb-4 opacity-50" />
              <p>{{ t('admin.documentDetail.noExpectedSigners') }}</p>
            </div>

            <!-- Confirmations complémentaires (toujours visible si présents) -->
            <div v-if="unexpectedSignatures.length > 0" class="mt-8 pt-8 border-t border-border">
              <h3 class="text-lg font-semibold mb-4 flex items-center">
                <span class="mr-2">⚠</span>
                {{ t('admin.documentDetail.unexpectedSignatures') }}
                <Badge variant="secondary" class="ml-2">{{ unexpectedSignatures.length }}</Badge>
              </h3>
              <p class="text-sm text-muted-foreground mb-4">
                {{ t('admin.documentDetail.unexpectedDescription') }}
              </p>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{{ t('admin.documentDetail.user') }}</TableHead>
                    <TableHead>{{ t('admin.documentDetail.confirmedOn') }}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-for="(sig, idx) in unexpectedSignatures" :key="idx">
                    <TableCell>
                      <div class="space-y-1">
                        <p class="font-medium">{{ sig.userName || sig.userEmail }}</p>
                        <p class="text-xs text-muted-foreground">{{ sig.userEmail }}</p>
                      </div>
                    </TableCell>
                    <TableCell>{{ formatDate(sig.signedAtUTC) }}</TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>

        <!-- Email Reminders -->
        <Card v-if="reminderStats && stats && stats.expectedCount > 0 && (smtpEnabled || reminderStats.totalSent > 0)" class="clay-card">
          <CardHeader>
            <CardTitle>{{ t('admin.documentDetail.reminders') }}</CardTitle>
            <CardDescription>{{ t('admin.documentDetail.remindersDescription') }}</CardDescription>
          </CardHeader>
          <CardContent class="space-y-6">
            <!-- Stats -->
            <div class="grid gap-4 sm:grid-cols-3">
              <div class="bg-muted rounded-lg p-4">
                <p class="text-sm text-muted-foreground">{{ t('admin.documentDetail.remindersSent') }}</p>
                <p class="text-2xl font-bold">{{ reminderStats.totalSent }}</p>
              </div>
              <div class="bg-muted rounded-lg p-4">
                <p class="text-sm text-muted-foreground">{{ t('admin.documentDetail.toRemind') }}</p>
                <p class="text-2xl font-bold">{{ reminderStats.pendingCount }}</p>
              </div>
              <div v-if="reminderStats.lastSentAt" class="bg-muted rounded-lg p-4">
                <p class="text-sm text-muted-foreground">{{ t('admin.documentDetail.lastReminder') }}</p>
                <p class="text-sm font-bold">{{ formatDate(reminderStats.lastSentAt) }}</p>
              </div>
            </div>

            <!-- Alert if SMTP disabled but reminders exist -->
            <Alert v-if="!smtpEnabled" class="border-orange-500 bg-orange-50 dark:bg-orange-900/20">
              <AlertDescription class="text-orange-800 dark:text-orange-200">
                {{ t('admin.documentDetail.emailServiceDisabled') }}
              </AlertDescription>
            </Alert>

            <!-- Send Form - Only shown if SMTP is enabled -->
            <div v-if="smtpEnabled" class="space-y-4">
              <div class="space-y-2">
                <label class="flex items-center space-x-2">
                  <input type="radio" v-model="sendMode" value="all" class="rounded-full" />
                  <span>{{ t('admin.documentDetail.sendToAll', { count: reminderStats.pendingCount }) }}</span>
                </label>
                <label class="flex items-center space-x-2">
                  <input type="radio" v-model="sendMode" value="selected" class="rounded-full" />
                  <span>{{ t('admin.documentDetail.sendToSelected', { count: selectedEmails.length }) }}</span>
                </label>
              </div>
              <Button @click="confirmSendReminders" :disabled="sendingReminders || (sendMode === 'selected' && selectedEmails.length === 0)">
                <Mail :size="16" class="mr-2" />
                {{ sendingReminders ? t('admin.documentDetail.sending') : t('admin.documentDetail.sendReminders') }}
              </Button>
            </div>
            <div v-else-if="smtpEnabled && reminderStats.pendingCount === 0" class="text-center py-4 text-muted-foreground">
              {{ t('admin.documentDetail.allContacted') }}
            </div>
          </CardContent>
        </Card>

        <!-- Danger Zone -->
        <Card class="clay-card border-destructive/50">
          <CardHeader>
            <CardTitle class="text-destructive">{{ t('admin.documentDetail.dangerZone') }}</CardTitle>
            <CardDescription>{{ t('admin.documentDetail.dangerZoneDescription') }}</CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex items-center justify-between p-4 bg-destructive/5 rounded-lg">
              <div class="flex-1">
                <h3 class="font-semibold text-foreground mb-1">{{ t('admin.documentDetail.deleteDocument') }}</h3>
                <p class="text-sm text-muted-foreground">
                  {{ t('admin.documentDetail.deleteDocumentDescription') }}
                </p>
              </div>
              <Button
                @click="showDeleteConfirmModal = true"
                variant="destructive"
                class="ml-4"
              >
                <Trash2 :size="16" class="mr-2" />
                {{ t('common.delete') }}
              </Button>
            </div>
          </CardContent>
        </Card>

        <!-- Chain Integrity - Feature not yet available in API v1 -->
        <!-- TODO: Add chain integrity verification endpoint to API v1 -->
      </div>
    </main>

    <!-- Add Signers Modal -->
    <div v-if="showAddSignersModal" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="showAddSignersModal = false">
      <Card class="max-w-2xl w-full">
        <CardHeader>
          <div class="flex items-center justify-between">
            <CardTitle>{{ t('admin.documentDetail.addSigners') }}</CardTitle>
            <Button variant="ghost" size="icon" @click="showAddSignersModal = false">
              <X :size="20" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <form @submit.prevent="addSigners" class="space-y-4">
            <div>
              <label class="block text-sm font-medium mb-2">{{ t('admin.documentDetail.emailsLabel') }}</label>
              <Textarea v-model="signersEmails" :rows="8"
                        :placeholder="t('admin.documentDetail.emailsPlaceholder')" />
              <p class="text-xs text-muted-foreground mt-2">
                {{ t('admin.documentDetail.emailsHelper') }}
              </p>
            </div>
            <div class="flex justify-end space-x-3">
              <Button type="button" variant="outline" @click="showAddSignersModal = false">{{ t('common.cancel') }}</Button>
              <Button type="submit" :disabled="addingSigners || !signersEmails.trim()">
                {{ addingSigners ? t('admin.documentDetail.adding') : t('admin.documentDetail.addButton') }}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>

    <!-- Import CSV Modal -->
    <div v-if="showImportCSVModal" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="closeImportCSVModal">
      <Card class="max-w-3xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <CardHeader>
          <div class="flex items-center justify-between">
            <CardTitle>{{ t('admin.documentDetail.importCSVTitle') }}</CardTitle>
            <Button variant="ghost" size="icon" @click="closeImportCSVModal">
              <X :size="20" />
            </Button>
          </div>
        </CardHeader>
        <CardContent class="flex-1 overflow-auto">
          <!-- Error Alert -->
          <Alert v-if="csvError" variant="destructive" class="mb-4">
            <AlertDescription>{{ csvError }}</AlertDescription>
          </Alert>

          <!-- Step 1: File Selection -->
          <div v-if="!csvPreview" class="space-y-4">
            <div>
              <label class="block text-sm font-medium mb-2">{{ t('admin.documentDetail.selectFile') }}</label>
              <input
                type="file"
                accept=".csv"
                @change="handleCSVFileChange"
                class="block w-full text-sm text-muted-foreground
                       file:mr-4 file:py-2 file:px-4
                       file:rounded-md file:border-0
                       file:text-sm file:font-medium
                       file:bg-primary file:text-primary-foreground
                       hover:file:bg-primary/90
                       cursor-pointer"
              />
              <p class="text-xs text-muted-foreground mt-2">
                {{ t('admin.documentDetail.csvFormatHelp') }}
              </p>
            </div>
            <div class="flex justify-end space-x-3">
              <Button type="button" variant="outline" @click="closeImportCSVModal">
                {{ t('common.cancel') }}
              </Button>
              <Button @click="analyzeCSV" :disabled="!csvFile || analyzingCSV">
                <Loader2 v-if="analyzingCSV" :size="16" class="mr-2 animate-spin" />
                {{ analyzingCSV ? t('admin.documentDetail.analyzing') : t('admin.documentDetail.analyze') }}
              </Button>
            </div>
          </div>

          <!-- Step 2: Preview -->
          <div v-else class="space-y-4">
            <!-- Summary -->
            <div class="grid gap-3 sm:grid-cols-3">
              <div class="bg-green-50 dark:bg-green-900/20 rounded-lg p-3 flex items-center gap-3">
                <FileCheck :size="24" class="text-green-600" />
                <div>
                  <p class="text-sm text-muted-foreground">{{ t('admin.documentDetail.validEntries') }}</p>
                  <p class="text-xl font-bold text-green-600">{{ signersToImport.length }}</p>
                </div>
              </div>
              <div v-if="csvPreview.existingEmails.length > 0" class="bg-orange-50 dark:bg-orange-900/20 rounded-lg p-3 flex items-center gap-3">
                <AlertTriangle :size="24" class="text-orange-600" />
                <div>
                  <p class="text-sm text-muted-foreground">{{ t('admin.documentDetail.existingEntries') }}</p>
                  <p class="text-xl font-bold text-orange-600">{{ csvPreview.existingEmails.length }}</p>
                </div>
              </div>
              <div v-if="csvPreview.invalidCount > 0" class="bg-red-50 dark:bg-red-900/20 rounded-lg p-3 flex items-center gap-3">
                <FileX :size="24" class="text-red-600" />
                <div>
                  <p class="text-sm text-muted-foreground">{{ t('admin.documentDetail.invalidEntries') }}</p>
                  <p class="text-xl font-bold text-red-600">{{ csvPreview.invalidCount }}</p>
                </div>
              </div>
            </div>

            <!-- Preview Table -->
            <div class="border rounded-lg overflow-hidden">
              <div class="max-h-64 overflow-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead class="w-16">{{ t('admin.documentDetail.lineNumber') }}</TableHead>
                      <TableHead>{{ t('admin.documentDetail.email') }}</TableHead>
                      <TableHead>{{ t('admin.documentDetail.name') }}</TableHead>
                      <TableHead class="w-32">{{ t('admin.documentDetail.status') }}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    <TableRow v-for="signer in csvPreview.signers" :key="signer.lineNumber" :class="getSignerStatus(signer) === 'exists' ? 'bg-orange-50/50 dark:bg-orange-900/10' : ''">
                      <TableCell class="text-muted-foreground">{{ signer.lineNumber }}</TableCell>
                      <TableCell>{{ signer.email }}</TableCell>
                      <TableCell>{{ signer.name || '-' }}</TableCell>
                      <TableCell>
                        <Badge :variant="getSignerStatus(signer) === 'exists' ? 'secondary' : 'default'">
                          {{ getSignerStatus(signer) === 'exists' ? t('admin.documentDetail.statusExists') : t('admin.documentDetail.statusValid') }}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  </TableBody>
                </Table>
              </div>
            </div>

            <!-- Errors Table -->
            <div v-if="csvPreview.errors.length > 0" class="border border-destructive rounded-lg overflow-hidden">
              <div class="bg-destructive/10 px-4 py-2 font-medium text-destructive">
                {{ t('admin.documentDetail.parseErrors') }}
              </div>
              <div class="max-h-32 overflow-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead class="w-16">{{ t('admin.documentDetail.lineNumber') }}</TableHead>
                      <TableHead>{{ t('admin.documentDetail.content') }}</TableHead>
                      <TableHead>{{ t('admin.documentDetail.errorReason') }}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    <TableRow v-for="err in csvPreview.errors" :key="err.lineNumber" class="bg-red-50/50 dark:bg-red-900/10">
                      <TableCell class="text-muted-foreground">{{ err.lineNumber }}</TableCell>
                      <TableCell class="font-mono text-xs truncate max-w-48">{{ err.content }}</TableCell>
                      <TableCell class="text-destructive text-sm">{{ t('admin.documentDetail.csvError.' + err.error, err.error) }}</TableCell>
                    </TableRow>
                  </TableBody>
                </Table>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex justify-between items-center pt-4">
              <Button type="button" variant="ghost" @click="csvPreview = null; csvFile = null">
                {{ t('admin.documentDetail.backToFileSelection') }}
              </Button>
              <div class="flex gap-3">
                <Button type="button" variant="outline" @click="closeImportCSVModal">
                  {{ t('common.cancel') }}
                </Button>
                <Button @click="confirmImportCSV" :disabled="importingCSV || signersToImport.length === 0">
                  <Loader2 v-if="importingCSV" :size="16" class="mr-2 animate-spin" />
                  {{ importingCSV ? t('admin.documentDetail.importing') : t('admin.documentDetail.importButton', { count: signersToImport.length }) }}
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteConfirmModal" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="showDeleteConfirmModal = false">
      <Card class="max-w-md w-full border-destructive">
        <CardHeader>
          <div class="flex items-center justify-between">
            <CardTitle class="text-destructive">{{ t('admin.documentDetail.deleteConfirmTitle') }}</CardTitle>
            <Button variant="ghost" size="icon" @click="showDeleteConfirmModal = false">
              <X :size="20" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div class="space-y-4">
            <Alert variant="destructive" class="border-destructive">
              <AlertDescription>
                <p class="font-semibold mb-2">{{ t('admin.documentDetail.deleteWarning') }}</p>
                <p class="text-sm">
                  {{ t('admin.documentDetail.deleteWillRemove') }}
                </p>
                <ul class="text-sm list-disc list-inside mt-2 space-y-1">
                  <li>{{ t('admin.documentDetail.deleteItem1') }}</li>
                  <li>{{ t('admin.documentDetail.deleteItem2') }}</li>
                  <li>{{ t('admin.documentDetail.deleteItem3') }}</li>
                  <li>{{ t('admin.documentDetail.deleteItem4') }}</li>
                </ul>
              </AlertDescription>
            </Alert>

            <div class="bg-muted p-3 rounded-lg">
              <p class="text-sm font-mono text-muted-foreground" v-html="t('admin.documentDetail.documentId') + ' ' + docId"></p>
            </div>

            <div class="flex justify-end space-x-3 pt-4">
              <Button type="button" variant="outline" @click="showDeleteConfirmModal = false">
                {{ t('common.cancel') }}
              </Button>
              <Button
                @click="handleDeleteDocument"
                variant="destructive"
                :disabled="deletingDocument"
              >
                <Trash2 v-if="!deletingDocument" :size="16" class="mr-2" />
                <Loader2 v-else :size="16" class="mr-2 animate-spin" />
                {{ deletingDocument ? t('admin.documentDetail.deleting') : t('admin.documentDetail.deleteConfirmButton') }}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Metadata Warning Modal -->
    <div v-if="showMetadataWarningModal" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="showMetadataWarningModal = false">
      <Card class="max-w-lg w-full border-orange-500">
        <CardHeader>
          <div class="flex items-center justify-between">
            <CardTitle class="text-orange-600">{{ t('admin.documentDetail.metadataWarning.title') }}</CardTitle>
            <Button variant="ghost" size="icon" @click="showMetadataWarningModal = false">
              <X :size="20" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div class="space-y-4">
            <Alert class="border-orange-500 bg-orange-50 dark:bg-orange-900/20">
              <AlertDescription>
                <p class="text-sm text-orange-800 dark:text-orange-200 mb-3">
                  {{ t('admin.documentDetail.metadataWarning.description') }}
                </p>
                <p class="text-sm font-semibold text-orange-900 dark:text-orange-100">
                  {{ t('admin.documentDetail.metadataWarning.warning') }}
                </p>
              </AlertDescription>
            </Alert>

            <div class="bg-muted p-4 rounded-lg">
              <p class="text-sm font-medium mb-2">
                {{ t('admin.documentDetail.metadataWarning.currentSignatures') }}
              </p>
              <div class="flex flex-col gap-3 mt-2">
                <div class="flex items-center gap-2">
                  <CheckCircle :size="20" class="text-green-600" />
                  <span class="text-sm">
                    <span class="font-bold text-lg">{{ (stats?.signedCount || 0) + (unexpectedSignatures?.length || 0) }}</span>
                    <span class="text-muted-foreground ml-1">
                      {{ ((stats?.signedCount || 0) + (unexpectedSignatures?.length || 0)) > 1 ? 'signatures' : 'signature' }}
                    </span>
                  </span>
                </div>
              </div>
            </div>

            <div class="flex justify-end space-x-3 pt-4">
              <Button type="button" variant="outline" @click="showMetadataWarningModal = false">
                {{ t('admin.documentDetail.metadataWarning.cancel') }}
              </Button>
              <Button
                @click="saveMetadata"
                variant="destructive"
                :disabled="savingMetadata"
              >
                <Loader2 v-if="savingMetadata" :size="16" class="mr-2 animate-spin" />
                {{ savingMetadata ? 'Enregistrement...' : t('admin.documentDetail.metadataWarning.confirm') }}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- Remove Signer Confirmation Dialog -->
    <ConfirmDialog
      v-if="showRemoveSignerModal"
      :title="t('admin.documentDetail.removeSignerTitle')"
      :message="t('admin.documentDetail.removeSignerMessage', { email: signerToRemove })"
      :confirm-text="t('common.delete')"
      :cancel-text="t('common.cancel')"
      variant="warning"
      @confirm="removeSigner"
      @cancel="cancelRemoveSigner"
    />

    <!-- Send Reminders Confirmation Dialog -->
    <ConfirmDialog
      v-if="showSendRemindersModal"
      :title="t('admin.documentDetail.confirmSendRemindersTitle')"
      :message="remindersMessage"
      :confirm-text="t('common.confirm')"
      :cancel-text="t('common.cancel')"
      variant="default"
      :loading="sendingReminders"
      @confirm="sendRemindersAction"
      @cancel="cancelSendReminders"
    />
  </div>
</template>
