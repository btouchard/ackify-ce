<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<template>
  <div class="min-h-screen bg-background text-foreground p-4">
    <!-- Loading state -->
    <div v-if="loading" class="flex items-center justify-center py-8">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="bg-destructive/10 dark:bg-destructive/20 border border-destructive/50 rounded-lg p-4">
      <p class="text-destructive text-sm">{{ error }}</p>
    </div>

    <!-- Document info and signatures -->
    <div v-else-if="documentData" class="max-w-2xl mx-auto">
      <!-- Document header with signatures -->
      <div v-if="documentData.signatures.length > 0">
        <div class="mb-6">
          <h2 class="text-xl font-bold text-foreground mb-2">
            Document: {{ documentData.title }}
          </h2>
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center space-x-4 text-sm text-muted-foreground">
              <span class="flex items-center">
                <svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
                {{ documentData.signatures.length }} confirmation(s)
              </span>
              <span v-if="documentData.metadata?.title">{{ documentData.metadata.title }}</span>
            </div>
            <!-- Sign button -->
            <a
              :href="signUrl"
              target="_blank"
              class="inline-flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background text-sm font-medium whitespace-nowrap"
            >
              <svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"/>
              </svg>
              Signer
            </a>
          </div>
        </div>

        <!-- Signatures list (compact) -->
        <div class="space-y-2">
          <div
            v-for="signature in documentData.signatures"
            :key="signature.id"
            class="bg-card text-card-foreground rounded-md px-3 py-2 border border-border flex items-center justify-between"
          >
            <div class="flex items-center space-x-2 min-w-0 flex-1">
              <svg class="w-4 h-4 text-green-600 dark:text-green-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
              <span class="text-sm font-medium text-foreground truncate">{{ signature.userEmail }}</span>
            </div>
            <span class="text-xs text-muted-foreground whitespace-nowrap ml-2">{{ formatDateCompact(signature.signedAt) }}</span>
          </div>
        </div>
      </div>

      <!-- Empty state - No signatures yet -->
      <div v-else class="text-center py-8">
        <svg class="w-16 h-16 mx-auto mb-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
        </svg>
        <p class="text-sm text-muted-foreground mb-4">Aucune signature pour ce document</p>
        <a
          :href="signUrl"
          target="_blank"
          class="inline-flex items-center px-6 py-3 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background text-base font-medium"
        >
          <svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"/>
          </svg>
          Signer ce document
        </a>
      </div>

      <!-- Footer branding -->
      <div class="mt-8 pt-4 border-t border-border text-center">
        <a
          href="https://github.com/btouchard/ackify-ce"
          target="_blank"
          class="text-xs text-muted-foreground hover:text-foreground transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background rounded"
        >
          Powered by Ackify
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePageTitle } from '@/composables/usePageTitle'
import { documentService } from '@/services/documents'
import http, { extractError } from '@/services/http'

const route = useRoute()
const router = useRouter()
usePageTitle('embed.title')

// State
const loading = ref(false)
const error = ref<string | null>(null)
const documentData = ref<any>(null)
const resolvedDocId = ref<string | null>(null)

// Computed
const docRef = computed(() => route.query.doc as string)

const signUrl = computed(() => {
  const baseUrl = (window as any).ACKIFY_BASE_URL || window.location.origin
  return `${baseUrl}/?doc=${encodeURIComponent(docRef.value)}`
})

// Methods
function formatDateCompact(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString('fr-FR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric'
  })
}

async function loadDocument() {
  if (!docRef.value) {
    error.value = 'ID de document manquant'
    loading.value = false
    return
  }

  try {
    loading.value = true
    error.value = null

    // First, find or create the document to get the docID
    const doc = await documentService.findOrCreateDocument(docRef.value)
    resolvedDocId.value = doc.docId

    // If the docRef is not the same as the docID, redirect to clean URL
    if (docRef.value !== doc.docId) {
      await router.replace({
        name: route.name as string,
        query: { doc: doc.docId }
      })
      return // Router will trigger watch and reload
    }

    // Then fetch signatures using the resolved docID
    const response = await http.get(`/documents/${doc.docId}/signatures`)

    // Build document data from signatures response
    const signatures = response.data.data || []
    documentData.value = {
      id: doc.docId,
      title: doc.title || `Document ${doc.docId}`,
      signatures: signatures,
      metadata: {}
    }

  } catch (err: any) {
    error.value = extractError(err)
  } finally {
    loading.value = false
  }
}

// Watch for changes in doc query param (for navigation)
watch(() => route.query.doc, () => {
  if (route.query.doc) {
    loadDocument()
  }
})

onMounted(() => {
  loadDocument()
})
</script>
