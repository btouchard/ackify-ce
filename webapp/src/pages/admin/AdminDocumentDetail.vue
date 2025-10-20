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
  type DocumentStatus,
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
const signerToRemove = ref('')
const remindersMessage = ref('')

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
const expectedSigners = computed(() => documentStatus.value?.expectedSigners || [])
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
    success.value = 'Métadonnées enregistrées avec succès'
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
    success.value = `${addedCount} lecteur(s) ajouté(s) avec succès`
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
    success.value = `${email} retiré avec succès`
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
      ? `Envoyer des relances à ${reminderStats.value?.pendingCount || 0} lecteur(s) en attente de confirmation ?`
      : `Envoyer des relances à ${selectedEmails.value.length} lecteur(s) sélectionné(s) ?`
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
        success.value = `${result.successfullySent} relance(s) envoyée(s), ${result.failed} échec(s)`
      } else {
        success.value = `${result.successfullySent} relance(s) envoyée(s) avec succès`
      }
    } else {
      success.value = 'Relances envoyées avec succès'
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
  success.value = 'Copié dans le presse-papiers'
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
          <Button variant="ghost" size="icon" @click="router.push('/admin')" aria-label="Retour">
            <ArrowLeft :size="20" />
          </Button>
          <h1 class="text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
            Document {{ docId }}
          </h1>
        </div>
        <div class="flex items-center gap-3 ml-14">
          <p class="text-sm text-muted-foreground font-mono">{{ shareLink }}</p>
          <Button @click="copyToClipboard(shareLink)" variant="ghost" size="icon" aria-label="Copier le lien">
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
        <p class="mt-4 text-muted-foreground">Chargement...</p>
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
                  <p class="text-sm font-medium text-muted-foreground">Attendus</p>
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
                  <p class="text-sm font-medium text-muted-foreground">Confirmés</p>
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
                  <p class="text-sm font-medium text-muted-foreground">En attente</p>
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
                  <p class="text-sm font-medium text-muted-foreground">Complétion</p>
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
              <CardTitle>📄 Informations sur le document</CardTitle>
              <CardDescription>Métadonnées et checksum du document</CardDescription>
            </div>
          </CardHeader>
          <CardContent>
            <form @submit.prevent="handleSaveMetadata" class="space-y-4">
              <!-- Titre et URL côte à côte -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium mb-2">Titre</label>
                  <Input v-model="metadataForm.title" placeholder="Politique de sécurité 2025" />
                </div>
                <div>
                  <label class="block text-sm font-medium mb-2">URL</label>
                  <Input v-model="metadataForm.url" type="url" placeholder="https://example.com/doc.pdf" />
                </div>
              </div>

              <!-- Checksum et Algorithme côte à côte -->
              <div class="grid grid-cols-1 md:grid-cols-[1fr_auto] gap-4">
                <div>
                  <label class="block text-sm font-medium mb-2">Checksum</label>
                  <Input v-model="metadataForm.checksum" placeholder="e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" class="font-mono text-sm" />
                </div>
                <div class="md:min-w-[140px]">
                  <label class="block text-sm font-medium mb-2">Algorithme</label>
                  <select v-model="metadataForm.checksumAlgorithm" class="flex h-10 w-full rounded-md clay-input px-3 py-2 text-sm">
                    <option value="SHA-256">SHA-256</option>
                    <option value="SHA-512">SHA-512</option>
                    <option value="MD5">MD5</option>
                  </select>
                </div>
              </div>

              <div>
                <label class="block text-sm font-medium mb-2">Description</label>
                <Textarea v-model="metadataForm.description" :rows="4" placeholder="Description du document..." />
              </div>
              <div v-if="documentMetadata" class="text-xs text-muted-foreground pt-2 border-t">
                Créé par {{ documentMetadata.createdBy }} le {{ formatDate(documentMetadata.createdAt) }}
              </div>
              <div class="flex justify-end">
                <Button type="submit" :disabled="savingMetadata">
                  {{ savingMetadata ? 'Enregistrement...' : 'Enregistrer' }}
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
                <CardTitle>✓ Lecteurs attendus</CardTitle>
                <CardDescription v-if="stats">{{ stats.signedCount }} / {{ stats.expectedCount }} confirmés</CardDescription>
              </div>
              <Button @click="showAddSignersModal = true" size="sm">
                <Plus :size="16" class="mr-2" />
                Ajouter
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <!-- Expected Signers Table -->
            <div v-if="expectedSigners.length > 0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>
                      <input type="checkbox" class="rounded"
                             @change="(e: any) => selectedEmails = e.target.checked ? expectedSigners.filter(s => !s.hasSigned).map(s => s.email) : []" />
                    </TableHead>
                    <TableHead>Lecteur</TableHead>
                    <TableHead>Statut</TableHead>
                    <TableHead>Confirmé le</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRow v-for="signer in expectedSigners" :key="signer.email">
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
                        {{ signer.hasSigned ? '✓ Confirmé' : '⏳ En attente' }}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {{ signer.signedAt ? formatDate(signer.signedAt) : '-' }}
                    </TableCell>
                    <TableCell>
                      <Button @click="confirmRemoveSigner(signer.email)" variant="ghost" size="sm">
                        <Trash2 :size="14" class="text-destructive" />
                      </Button>
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>
            <div v-else class="text-center py-8 text-muted-foreground">
              <Users :size="48" class="mx-auto mb-4 opacity-50" />
              <p>Aucun lecteur attendu</p>
            </div>

            <!-- Confirmations complémentaires (toujours visible si présents) -->
            <div v-if="unexpectedSignatures.length > 0" class="mt-8 pt-8 border-t border-border">
              <h3 class="text-lg font-semibold mb-4 flex items-center">
                <span class="mr-2">⚠</span>
                Confirmations de lecture complémentaires
                <Badge variant="secondary" class="ml-2">{{ unexpectedSignatures.length }}</Badge>
              </h3>
              <p class="text-sm text-muted-foreground mb-4">
                Utilisateurs ayant confirmé mais non présents dans la liste des lecteurs attendus
              </p>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Utilisateur</TableHead>
                    <TableHead>Confirmé le</TableHead>
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
        <Card v-if="reminderStats && stats && stats.expectedCount > 0" class="clay-card">
          <CardHeader>
            <CardTitle>📧 Relances par email</CardTitle>
            <CardDescription>Envoyer des rappels aux lecteurs en attente de confirmation</CardDescription>
          </CardHeader>
          <CardContent class="space-y-6">
            <!-- Stats -->
            <div class="grid gap-4 sm:grid-cols-3">
              <div class="bg-muted rounded-lg p-4">
                <p class="text-sm text-muted-foreground">Relances envoyées</p>
                <p class="text-2xl font-bold">{{ reminderStats.totalSent }}</p>
              </div>
              <div class="bg-muted rounded-lg p-4">
                <p class="text-sm text-muted-foreground">À relancer</p>
                <p class="text-2xl font-bold">{{ reminderStats.pendingCount }}</p>
              </div>
              <div v-if="reminderStats.lastSentAt" class="bg-muted rounded-lg p-4">
                <p class="text-sm text-muted-foreground">Dernière relance</p>
                <p class="text-sm font-bold">{{ formatDate(reminderStats.lastSentAt) }}</p>
              </div>
            </div>

            <!-- Send Form -->
            <div v-if="reminderStats.pendingCount > 0" class="space-y-4">
              <div class="space-y-2">
                <label class="flex items-center space-x-2">
                  <input type="radio" v-model="sendMode" value="all" class="rounded-full" />
                  <span>Envoyer à tous les lecteurs en attente ({{ reminderStats.pendingCount }})</span>
                </label>
                <label class="flex items-center space-x-2">
                  <input type="radio" v-model="sendMode" value="selected" class="rounded-full" />
                  <span>Envoyer uniquement aux sélectionnés ({{ selectedEmails.length }})</span>
                </label>
              </div>
              <Button @click="confirmSendReminders" :disabled="sendingReminders || (sendMode === 'selected' && selectedEmails.length === 0)">
                <Mail :size="16" class="mr-2" />
                {{ sendingReminders ? 'Envoi...' : 'Envoyer les relances' }}
              </Button>
            </div>
            <div v-else class="text-center py-4 text-muted-foreground">
              ✓ Tous les lecteurs attendus ont été contactés ou ont confirmé
            </div>
          </CardContent>
        </Card>

        <!-- Danger Zone -->
        <Card class="clay-card border-destructive/50">
          <CardHeader>
            <CardTitle class="text-destructive">⚠️ Zone de danger</CardTitle>
            <CardDescription>Actions irréversibles sur ce document</CardDescription>
          </CardHeader>
          <CardContent>
            <div class="flex items-center justify-between p-4 bg-destructive/5 rounded-lg">
              <div class="flex-1">
                <h3 class="font-semibold text-foreground mb-1">Supprimer ce document</h3>
                <p class="text-sm text-muted-foreground">
                  Cette action supprimera définitivement le document, ses métadonnées, les lecteurs attendus et toutes les confirmations associées.<br>
                  Cette action est irréversible.
                </p>
              </div>
              <Button
                @click="showDeleteConfirmModal = true"
                variant="destructive"
                class="ml-4"
              >
                <Trash2 :size="16" class="mr-2" />
                Supprimer
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
            <CardTitle>Ajouter des lecteurs attendus</CardTitle>
            <Button variant="ghost" size="icon" @click="showAddSignersModal = false">
              <X :size="20" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <form @submit.prevent="addSigners" class="space-y-4">
            <div>
              <label class="block text-sm font-medium mb-2">Emails (un par ligne)</label>
              <Textarea v-model="signersEmails" :rows="8"
                        placeholder="Marie Dupont <marie.dupont@example.com>&#10;jean.martin@example.com&#10;Sophie Bernard <sophie@example.com>" />
              <p class="text-xs text-muted-foreground mt-2">
                Formats acceptés : "Nom Prénom &lt;email@example.com&gt;" ou "email@example.com"
              </p>
            </div>
            <div class="flex justify-end space-x-3">
              <Button type="button" variant="outline" @click="showAddSignersModal = false">Annuler</Button>
              <Button type="submit" :disabled="addingSigners || !signersEmails.trim()">
                {{ addingSigners ? 'Ajout...' : 'Ajouter' }}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteConfirmModal" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4" @click.self="showDeleteConfirmModal = false">
      <Card class="max-w-md w-full border-destructive">
        <CardHeader>
          <div class="flex items-center justify-between">
            <CardTitle class="text-destructive">⚠️ Confirmer la suppression</CardTitle>
            <Button variant="ghost" size="icon" @click="showDeleteConfirmModal = false">
              <X :size="20" />
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <div class="space-y-4">
            <Alert variant="destructive" class="border-destructive">
              <AlertDescription>
                <p class="font-semibold mb-2">Cette action est irréversible !</p>
                <p class="text-sm">
                  La suppression de ce document entraînera la perte définitive de :
                </p>
                <ul class="text-sm list-disc list-inside mt-2 space-y-1">
                  <li>Toutes les métadonnées du document</li>
                  <li>La liste des lecteurs attendus</li>
                  <li>Toutes les confirmations cryptographiques</li>
                  <li>L'historique des relances</li>
                </ul>
              </AlertDescription>
            </Alert>

            <div class="bg-muted p-3 rounded-lg">
              <p class="text-sm font-mono text-muted-foreground">
                Document ID: <span class="text-foreground font-semibold">{{ docId }}</span>
              </p>
            </div>

            <div class="flex justify-end space-x-3 pt-4">
              <Button type="button" variant="outline" @click="showDeleteConfirmModal = false">
                Annuler
              </Button>
              <Button
                @click="handleDeleteDocument"
                variant="destructive"
                :disabled="deletingDocument"
              >
                <Trash2 v-if="!deletingDocument" :size="16" class="mr-2" />
                <Loader2 v-else :size="16" class="mr-2 animate-spin" />
                {{ deletingDocument ? 'Suppression...' : 'Supprimer définitivement' }}
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
      title="⚠️ Retirer le lecteur attendu"
      :message="`Retirer ${signerToRemove} de la liste des lecteurs attendus ?`"
      confirm-text="Retirer"
      cancel-text="Annuler"
      variant="warning"
      @confirm="removeSigner"
      @cancel="cancelRemoveSigner"
    />

    <!-- Send Reminders Confirmation Dialog -->
    <ConfirmDialog
      v-if="showSendRemindersModal"
      title="📧 Envoyer des relances"
      :message="remindersMessage"
      confirm-text="Envoyer"
      cancel-text="Annuler"
      variant="default"
      :loading="sendingReminders"
      @confirm="sendRemindersAction"
      @cancel="cancelSendReminders"
    />
  </div>
</template>
