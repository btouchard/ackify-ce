<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<template>
  <div class="container mx-auto px-4 py-8">
    <!-- Back link -->
    <div class="mb-6">
      <router-link to="/admin" class="text-primary hover:text-primary/80 flex items-center focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background rounded px-2 py-1 -ml-2 transition-colors">
        <svg class="w-5 h-5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        Retour au tableau de bord
      </router-link>
    </div>

    <!-- Error message -->
    <div v-if="error" class="bg-destructive/10 dark:bg-destructive/20 border border-destructive/50 text-destructive px-4 py-3 rounded mb-6">
      {{ error }}
    </div>

    <!-- Success message -->
    <div v-if="successMessage" class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-800 dark:text-green-400 px-4 py-3 rounded mb-6">
      {{ successMessage }}
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="text-center py-12">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      <p class="mt-2 text-muted-foreground">Chargement...</p>
    </div>

    <!-- Document details -->
    <div v-else-if="documentData">
      <!-- Document info -->
      <div class="bg-card text-card-foreground shadow-md rounded-lg p-6 mb-6 border border-border" v-if="documentData.document">
        <h1 class="text-2xl font-bold text-foreground mb-2">{{ documentData.document.title }}</h1>
        <p class="text-muted-foreground mb-4">{{ documentData.document.description }}</p>
        <div class="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span class="font-medium text-muted-foreground">ID:</span>
            <span class="text-foreground ml-2">{{ documentData.document.docId }}</span>
          </div>
          <div>
            <span class="font-medium text-muted-foreground">URL:</span>
            <a
              :href="documentData.document.url"
              target="_blank"
              rel="noopener noreferrer"
              class="text-primary hover:text-primary/80 ml-2 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background rounded"
            >
              {{ documentData.document.url }}
            </a>
          </div>
          <div>
            <span class="font-medium text-muted-foreground">Créé le:</span>
            <span class="text-foreground ml-2">{{ formatDate(documentData.document.createdAt) }}</span>
          </div>
          <div>
            <span class="font-medium text-muted-foreground">Par:</span>
            <span class="text-foreground ml-2">{{ documentData.document.createdBy }}</span>
          </div>
        </div>
      </div>

      <!-- Statistics -->
      <div class="grid grid-cols-4 gap-4 mb-6">
        <div class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
          <div class="text-blue-900 dark:text-blue-100 text-2xl font-bold">{{ documentData.stats.expectedCount }}</div>
          <div class="text-blue-700 dark:text-blue-300 text-sm">Attendus</div>
        </div>
        <div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-4">
          <div class="text-green-900 dark:text-green-100 text-2xl font-bold">{{ documentData.stats.signedCount }}</div>
          <div class="text-green-700 dark:text-green-300 text-sm">Signés</div>
        </div>
        <div class="bg-orange-50 dark:bg-orange-900/20 border border-orange-200 dark:border-orange-800 rounded-lg p-4">
          <div class="text-orange-900 dark:text-orange-100 text-2xl font-bold">{{ documentData.stats.pendingCount }}</div>
          <div class="text-orange-700 dark:text-orange-300 text-sm">En attente</div>
        </div>
        <div class="bg-purple-50 dark:bg-purple-900/20 border border-purple-200 dark:border-purple-800 rounded-lg p-4">
          <div class="text-purple-900 dark:text-purple-100 text-2xl font-bold">{{ documentData.stats.completionRate.toFixed(0) }}%</div>
          <div class="text-purple-700 dark:text-purple-300 text-sm">Complétude</div>
        </div>
      </div>

      <!-- Add signer form -->
      <div class="bg-card text-card-foreground shadow-md rounded-lg p-6 mb-6 border border-border">
        <h2 class="text-xl font-bold text-foreground mb-4">Ajouter un signataire attendu</h2>
        <form @submit.prevent="handleAddSigner" class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label for="email" class="block text-sm font-medium text-foreground mb-1">Email *</label>
            <input
              id="email"
              v-model="newSigner.email"
              type="email"
              required
              class="w-full px-3 py-2 bg-background text-foreground border border-input rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
              placeholder="email@example.com"
            />
          </div>
          <div>
            <label for="name" class="block text-sm font-medium text-foreground mb-1">Nom</label>
            <input
              id="name"
              v-model="newSigner.name"
              type="text"
              class="w-full px-3 py-2 bg-background text-foreground border border-input rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
              placeholder="Nom complet"
            />
          </div>
          <div class="flex items-end">
            <button
              type="submit"
              :disabled="adding"
              class="w-full bg-primary text-primary-foreground px-4 py-2 rounded-md hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background transition-colors"
            >
              {{ adding ? 'Ajout...' : 'Ajouter' }}
            </button>
          </div>
        </form>
      </div>

      <!-- Reminders section -->
      <div class="bg-card text-card-foreground shadow-md rounded-lg p-6 mb-6 border border-border" v-if="documentData.stats.pendingCount > 0">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-xl font-bold text-foreground">Relances email</h2>
          <button
            @click="confirmSendReminders"
            :disabled="sendingReminders"
            class="bg-green-600 dark:bg-green-700 text-white px-4 py-2 rounded-md hover:bg-green-700 dark:hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background transition-colors"
          >
            {{ sendingReminders ? 'Envoi...' : `Relancer tous (${documentData.stats.pendingCount})` }}
          </button>
        </div>
        <p class="text-sm text-muted-foreground">
          Envoyez une relance par email aux signataires qui n'ont pas encore signé le document.
        </p>
      </div>

      <!-- Signers list -->
      <div class="bg-card text-card-foreground shadow-md rounded-lg overflow-hidden border border-border">
        <div class="px-6 py-4 border-b border-border">
          <h2 class="text-xl font-bold text-foreground">Signataires attendus</h2>
        </div>
        <div v-if="documentData.expectedSigners.length === 0" class="p-6 text-center text-muted-foreground">
          Aucun signataire attendu pour ce document
        </div>
        <table v-else class="min-w-full divide-y divide-border">
          <thead class="bg-muted">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Signataire
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Statut
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Ajouté
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Relances
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="bg-card divide-y divide-border">
            <tr v-for="signer in documentData.expectedSigners" :key="signer.id" class="hover:bg-accent transition-colors">
              <td class="px-6 py-4">
                <div class="text-sm font-medium text-foreground">{{ signer.name || signer.email }}</div>
                <div class="text-sm text-muted-foreground">{{ signer.email }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  v-if="signer.hasSigned"
                  class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400"
                >
                  Signé {{ formatDate(signer.signedAt!) }}
                </span>
                <span
                  v-else
                  class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-orange-100 dark:bg-orange-900/30 text-orange-800 dark:text-orange-400"
                >
                  En attente ({{ signer.daysSinceAdded }}j)
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                {{ formatDate(signer.addedAt) }}<br />
                <span class="text-xs">par {{ signer.addedBy }}</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                {{ signer.reminderCount }} relance(s)
                <span v-if="signer.lastReminderSent" class="block text-xs">
                  Dernière: il y a {{ signer.daysSinceLastReminder }}j
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button
                  v-if="!signer.hasSigned"
                  @click="confirmRemoveSigner(signer.email)"
                  :disabled="removing === signer.email"
                  class="text-destructive hover:text-destructive/80 disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background rounded px-2 py-1 transition-colors"
                >
                  {{ removing === signer.email ? 'Suppression...' : 'Retirer' }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Remove Signer Confirmation Dialog -->
    <ConfirmDialog
      v-if="showRemoveSignerModal"
      title="⚠️ Retirer le signataire attendu"
      :message="`Êtes-vous sûr de vouloir retirer ${signerToRemove} des signataires attendus ?`"
      confirm-text="Retirer"
      cancel-text="Annuler"
      variant="warning"
      @confirm="handleRemoveSigner"
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
      @confirm="handleSendReminders"
      @cancel="cancelSendReminders"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { usePageTitle } from '@/composables/usePageTitle'
import {
  getDocumentStatus,
  addExpectedSigner,
  removeExpectedSigner,
  sendReminders,
  type DocumentStatus,
} from '@/services/admin'
import { extractError } from '@/services/http'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'

const route = useRoute()
const { locale } = useI18n()
const docId = route.params.docId as string
usePageTitle('admin.documentDetail.title', { docId })

const documentData = ref<DocumentStatus | null>(null)
const loading = ref(true)
const adding = ref(false)
const removing = ref('')
const sendingReminders = ref(false)
const error = ref('')
const successMessage = ref('')

// Modals
const showRemoveSignerModal = ref(false)
const showSendRemindersModal = ref(false)
const signerToRemove = ref('')
const remindersMessage = ref('')

const newSigner = ref<{
  email: string
  name: string
  notes?: string
}>({
  email: '',
  name: '',
})

async function loadDocument() {
  try {
    loading.value = true
    error.value = ''
    const response = await getDocumentStatus(docId)
    documentData.value = response.data
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to load document:', err)
  } finally {
    loading.value = false
  }
}

async function handleAddSigner() {
  try {
    adding.value = true
    error.value = ''
    successMessage.value = ''

    await addExpectedSigner(docId, newSigner.value)

    successMessage.value = `Signataire ${newSigner.value.email} ajouté avec succès`
    newSigner.value = { email: '', name: '' }

    // Reload document to refresh signers list
    await loadDocument()
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to add signer:', err)
  } finally {
    adding.value = false
  }
}

function confirmRemoveSigner(email: string) {
  signerToRemove.value = email
  showRemoveSignerModal.value = true
}

async function handleRemoveSigner() {
  const email = signerToRemove.value
  if (!email) return

  try {
    removing.value = email
    error.value = ''
    successMessage.value = ''

    await removeExpectedSigner(docId, email)

    successMessage.value = `Signataire ${email} retiré avec succès`
    showRemoveSignerModal.value = false
    signerToRemove.value = ''

    // Reload document to refresh signers list
    await loadDocument()
  } catch (err) {
    error.value = extractError(err)
    console.error('Failed to remove signer:', err)
  } finally {
    removing.value = ''
  }
}

function cancelRemoveSigner() {
  showRemoveSignerModal.value = false
  signerToRemove.value = ''
}

function confirmSendReminders() {
  remindersMessage.value = `Envoyer une relance par email aux ${documentData.value?.stats.pendingCount} signataire(s) en attente ?`
  showSendRemindersModal.value = true
}

async function handleSendReminders() {
  try {
    sendingReminders.value = true
    error.value = ''
    successMessage.value = ''

    // Normalize locale to base language code (it-IT -> it, en-US -> en, etc.)
    const normalizedLocale = locale.value.split('-')[0]
    console.log('Sending reminders with locale:', normalizedLocale, '(from', locale.value, ')')

    const response = await sendReminders(docId, {}, normalizedLocale)
    const result = response.data.result

    showSendRemindersModal.value = false

    if (result.failed > 0) {
      successMessage.value = `${result.successfullySent} relance(s) envoyée(s), ${result.failed} échec(s)`
      if (result.errors) {
        error.value = result.errors.join(', ')
      }
    } else {
      successMessage.value = `${result.successfullySent} relance(s) envoyée(s) avec succès`
    }

    // Reload document to refresh reminder stats
    await loadDocument()
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

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString('fr-FR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

onMounted(() => {
  loadDocument()
})
</script>
