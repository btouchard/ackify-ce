<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSignatureStore } from '@/stores/signatures'
import { useI18n } from 'vue-i18n'
import { usePageTitle } from '@/composables/usePageTitle'

const { t } = useI18n()
usePageTitle('sign.title')

import { AlertTriangle, CheckCircle2, FileText, Info, Users, Loader2, Shield, Zap, Clock } from 'lucide-vue-next'
import SignButton from '@/components/SignButton.vue'
import SignatureList from '@/components/SignatureList.vue'
import { documentService, type FindOrCreateDocumentResponse } from '@/services/documents'
import { detectReference } from '@/services/referenceDetector'
import { calculateFileChecksum } from '@/services/checksumCalculator'
import { updateDocumentMetadata } from '@/services/admin'
import DocumentForm from "@/components/DocumentForm.vue"

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const signatureStore = useSignatureStore()

const docId = ref<string | undefined>(undefined)
const user = computed(() => authStore.user)
const isAdmin = computed(() => authStore.isAdmin)

// Check if document creation is restricted to admins
const onlyAdminCanCreate = (window as any).ACKIFY_ONLY_ADMIN_CAN_CREATE || false
const canCreateDocument = computed(() => !onlyAdminCanCreate || isAdmin.value)
const currentDocument = ref<FindOrCreateDocumentResponse | null>(null)

const documentSignatures = ref<any[]>([])
const loadingSignatures = ref(false)
const loadingDocument = ref(false)
const showSuccessMessage = ref(false)
const errorMessage = ref<string | null>(null)
const needsAuth = ref(false)
const calculatingChecksum = ref(false)

// Check if current user has signed this document
const userHasSigned = computed(() => {
  if (!user.value?.email || documentSignatures.value.length === 0) {
    return false
  }
  return documentSignatures.value.some(sig => sig.userEmail === user.value?.email)
})

async function loadDocumentSignatures() {
  if (!docId.value) return

  loadingSignatures.value = true
  try {
    documentSignatures.value = await signatureStore.fetchDocumentSignatures(docId.value)
  } catch (error) {
    console.error('Failed to load document signatures:', error)
  } finally {
    loadingSignatures.value = false
  }
}

async function handleDocumentReference(ref: string) {
  try {
    loadingDocument.value = true
    errorMessage.value = null
    needsAuth.value = false

    console.log('Loading document for reference:', ref)

    // Detect reference type
    const refInfo = detectReference(ref)
    console.log('Reference detected as:', refInfo)

    // Call find-or-create API
    const doc = await documentService.findOrCreateDocument(ref)
    console.log('Document loaded:', doc)

    docId.value = doc.docId
    currentDocument.value = doc

    // If the ref is not the same as the docID, redirect to clean URL
    if (ref !== doc.docId) {
      await router.replace({
        name: route.name as string,
        query: { doc: doc.docId }
      })
      // Continue loading even after redirect
    }

    // If new document AND downloadable URL → calculate checksum
    if (doc.isNew && refInfo.isDownloadable && refInfo.type === 'url' && !doc.checksum) {
      await calculateAndUpdateChecksum(doc.docId, refInfo.value)
    }

    // Load signatures
    await loadDocumentSignatures()
  } catch (error: any) {
    console.error('Failed to load/create document:', error)

    // Handle 401 Unauthorized - user needs to authenticate
    if (error.response?.status === 401) {
      errorMessage.value = t('sign.error.authRequired')
      needsAuth.value = true
    } else {
      errorMessage.value = error.message || t('sign.error.loadFailed')
      needsAuth.value = false
    }
  } finally {
    loadingDocument.value = false
  }
}

function handleLoginClick() {
  authStore.startOAuthLogin(route.fullPath)
}

async function calculateAndUpdateChecksum(docId: string, url: string) {
  try {
    calculatingChecksum.value = true
    console.log('Calculating checksum for:', url)

    const checksumData = await calculateFileChecksum(url)
    console.log('Checksum calculated:', checksumData.checksum)

    // Update document metadata with checksum (if user is admin)
    if (authStore.isAdmin) {
      await updateDocumentMetadata(docId, {
        checksum: checksumData.checksum,
        checksumAlgorithm: checksumData.algorithm
      })

      // Update local document reference
      if (currentDocument.value) {
        currentDocument.value.checksum = checksumData.checksum
        currentDocument.value.checksumAlgorithm = checksumData.algorithm
      }

      console.log('Checksum updated in database')
    } else {
      console.log('Checksum calculated but not saved (user not admin)')
    }
  } catch (error) {
    console.warn('Checksum calculation failed:', error)
    // Don't fail the whole operation if checksum fails
  } finally {
    calculatingChecksum.value = false
  }
}

async function handleSigned() {
  showSuccessMessage.value = true
  errorMessage.value = null

  // Reload signatures to show the new one
  await loadDocumentSignatures()

  // Hide success message after 5 seconds
  setTimeout(() => {
    showSuccessMessage.value = false
  }, 5000)
}

function handleError(error: string) {
  errorMessage.value = error
  showSuccessMessage.value = false
}

// Helper to wait for auth to be initialized by App.vue
async function waitForAuth() {
  // If already initialized, return immediately
  if (authStore.initialized) return

  // Otherwise wait for initialized to become true
  return new Promise<void>((resolve) => {
    const stopWatch = watch(
      () => authStore.initialized,
      (isInit) => {
        if (isInit) {
          stopWatch()
          resolve()
        }
      },
      { immediate: true }
    )
  })
}

// Watch for route query changes (only for changes, not initial mount)
watch(() => route.query.doc, async (newRef, oldRef) => {
  // Only process if the doc query parameter actually changed
  if (newRef === oldRef) return

  // Reset state
  showSuccessMessage.value = false
  errorMessage.value = null
  needsAuth.value = false
  docId.value = undefined
  currentDocument.value = null
  documentSignatures.value = []

  // If we have a reference, load/create the document
  if (newRef && typeof newRef === 'string') {
    // Wait for App.vue to finish checking auth
    await waitForAuth()
    await handleDocumentReference(newRef)
  }
})

onMounted(async () => {
  // CRITICAL: Wait for App.vue to finish auth check before doing anything
  // App.vue calls checkAuth() which will set initialized=true when done
  await waitForAuth()

  // Now handle the document reference if present in URL
  const ref = route.query.doc as string | undefined
  if (ref) {
    await handleDocumentReference(ref)
  }
})
</script>

<template>
  <div class="min-h-[calc(100vh-8rem)]">
    <!-- Main Content -->
    <div class="mx-auto max-w-6xl px-4 sm:px-6 py-6 sm:py-8">
      <!-- Page Header -->
      <div class="mb-8 text-center">
        <h1 class="mb-2 text-2xl sm:text-3xl font-bold tracking-tight text-slate-900 dark:text-slate-50">
          {{ t('sign.title') }}
        </h1>
        <p class="text-base sm:text-lg text-slate-500 dark:text-slate-400">
          {{ t('sign.subtitle') }}
        </p>
      </div>

      <!-- Error Message -->
      <transition
        enter-active-class="transition ease-out duration-300"
        enter-from-class="opacity-0 translate-y-2"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition ease-in duration-200"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 translate-y-2"
      >
        <div v-if="errorMessage && !loadingDocument" class="mb-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4">
          <div class="flex items-start">
            <AlertTriangle :size="20" class="mr-3 mt-0.5 text-red-600 dark:text-red-400 flex-shrink-0" />
            <div class="flex-1">
              <h3 class="font-medium text-red-900 dark:text-red-200">{{ t('sign.error.title') }}</h3>
              <p class="mt-1 text-sm text-red-700 dark:text-red-300">{{ errorMessage }}</p>
              <div v-if="needsAuth" class="mt-4">
                <button
                  @click="handleLoginClick"
                  class="trust-gradient text-white font-medium rounded-lg px-4 py-2.5 text-sm hover:opacity-90 transition-opacity"
                >
                  {{ t('sign.error.loginButton') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- Loading state -->
      <div v-if="loadingDocument" class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-12 text-center">
        <Loader2 :size="48" class="mx-auto mb-4 animate-spin text-blue-600" />
        <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-100 mb-2">{{ t('sign.loading.title') }}</h2>
        <p class="text-slate-500 dark:text-slate-400">
          {{ t('sign.loading.description') }}
        </p>
      </div>

      <!-- No Document: Show help message -->
      <div v-else-if="!docId" class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-12 text-center">
        <div class="w-14 h-14 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center mx-auto mb-4">
          <FileText :size="28" class="text-blue-600 dark:text-blue-400" />
        </div>
        <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-100 mb-2">{{ t('sign.noDocument.title') }}</h2>
        <p class="text-slate-500 dark:text-slate-400 mb-4 max-w-md mx-auto">
          {{ t('sign.noDocument.description', { code: '?doc=' }) }}
        </p>
        <div class="text-sm text-slate-500 dark:text-slate-400 max-w-md mx-auto">
          <p class="font-medium text-slate-700 dark:text-slate-300">{{ t('sign.noDocument.examples') }}</p>
          <div class="space-y-1 mt-2">
            <code class="block px-3 bg-slate-50 dark:bg-slate-900 rounded-lg text-xs font-mono text-slate-600 dark:text-slate-400">/?doc=https://example.com/policy.pdf</code>
            <code class="block px-3 bg-slate-50 dark:bg-slate-900 rounded-lg text-xs font-mono text-slate-600 dark:text-slate-400">/?doc=/path/to/document</code>
            <code class="block px-3 bg-slate-50 dark:bg-slate-900 rounded-lg text-xs font-mono text-slate-600 dark:text-slate-400">/?doc=my-unique-ref</code>
          </div>

          <!-- Document creation form -->
          <p v-if="canCreateDocument" class="mt-6 mb-3 font-medium text-slate-700 dark:text-slate-300">{{ t('sign.noDocument.orEnterReference') }}</p>
          <DocumentForm v-if="canCreateDocument" />

          <!-- Restricted message -->
          <div v-else class="mt-4 accent-border bg-amber-50 dark:bg-amber-900/20 rounded-r-lg p-4 text-left">
            <div class="flex items-start">
              <AlertTriangle :size="18" class="mr-3 mt-0.5 text-amber-600 dark:text-amber-400 flex-shrink-0" />
              <p class="text-sm text-amber-700 dark:text-amber-300">{{ t('sign.documentCreation.restrictedToAdmins') }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Main Content when doc ID is present -->
      <div v-else-if="docId" class="space-y-6">
        <!-- Success Message -->
        <transition
          enter-active-class="transition ease-out duration-300"
          enter-from-class="opacity-0 translate-y-2"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition ease-in duration-200"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 translate-y-2"
        >
          <div v-if="showSuccessMessage" class="bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 rounded-xl p-4">
            <div class="flex items-start">
              <CheckCircle2 :size="20" class="mr-3 mt-0.5 text-emerald-600 dark:text-emerald-400 flex-shrink-0" />
              <div class="flex-1">
                <h3 class="font-medium text-emerald-900 dark:text-emerald-200">{{ t('sign.success.title') }}</h3>
                <p class="mt-1 text-sm text-emerald-700 dark:text-emerald-300">{{ t('sign.success.description') }}</p>
              </div>
            </div>
          </div>
        </transition>

        <!-- Document Info Card -->
        <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6">
          <!-- Header with icon -->
          <div class="flex items-start gap-4 mb-6">
            <div class="w-14 h-14 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
              <FileText :size="28" class="text-blue-600 dark:text-blue-400" />
            </div>
            <div class="flex-1 min-w-0">
              <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-100">
                {{ t('sign.document.title') }}<template v-if="currentDocument?.title"> : {{ currentDocument.title }}</template>
              </h2>
              <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
                <template v-if="currentDocument?.url">
                  <a
                    :href="currentDocument.url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="text-blue-600 dark:text-blue-400 hover:underline font-mono text-xs break-all"
                  >
                    {{ currentDocument.url }}
                  </a>
                </template>
                <template v-else>
                  <span class="font-mono text-xs">{{ docId }}</span>
                </template>
              </p>
            </div>
          </div>

          <!-- Sign Button -->
          <div class="pb-4">
            <SignButton
              :doc-id="docId"
              :signatures="documentSignatures"
              @signed="handleSigned"
              @error="handleError"
            />
          </div>

          <!-- Info Box (only shown if user hasn't signed yet) -->
          <div v-if="!userHasSigned" class="accent-border bg-blue-50 dark:bg-blue-900/20 rounded-r-lg p-4">
            <div class="flex items-start">
              <Info :size="18" class="mr-3 mt-0.5 text-blue-600 dark:text-blue-400 flex-shrink-0" />
              <div class="flex-1 space-y-2 text-sm text-blue-800 dark:text-blue-200">
                <p>{{ t('sign.info.description') }}</p>
                <p class="font-medium">{{ t('sign.info.recorded') }}</p>
                <ul class="list-disc space-y-1 pl-5 text-blue-700 dark:text-blue-300">
                  <li>{{ t('sign.info.email') }} : <strong class="text-blue-900 dark:text-blue-100">{{ user?.email }}</strong></li>
                  <li>{{ t('sign.info.timestamp') }}</li>
                  <li>{{ t('sign.info.signature') }}</li>
                  <li>{{ t('sign.info.hash') }}</li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <!-- Existing Confirmations -->
        <div v-if="documentSignatures.length > 0" class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700">
          <!-- Header -->
          <div class="p-6 border-b border-slate-100 dark:border-slate-700">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center">
                <Users :size="20" class="text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <h3 class="font-semibold text-slate-900 dark:text-slate-100">{{ t('sign.confirmations.title') }}</h3>
                <p class="text-sm text-slate-500 dark:text-slate-400">
                  {{ t('sign.confirmations.count', { count: documentSignatures.length }, documentSignatures.length) }}
                  {{ t('sign.confirmations.recorded', {}, documentSignatures.length) }}
                </p>
              </div>
            </div>
          </div>

          <!-- List -->
          <div class="p-6">
            <SignatureList
              :signatures="documentSignatures"
              :loading="loadingSignatures"
              :show-user-info="true"
              :show-details="true"
            />
          </div>
        </div>

        <!-- Empty State -->
        <div v-else-if="!loadingSignatures" class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-12 text-center">
          <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-slate-100 dark:bg-slate-700">
            <Users :size="28" class="text-slate-400" />
          </div>
          <h3 class="mb-2 text-lg font-semibold text-slate-900 dark:text-slate-100">
            {{ t('sign.empty.title') }}
          </h3>
          <p class="text-sm text-slate-500 dark:text-slate-400">
            {{ t('sign.empty.description') }}
          </p>
        </div>
      </div>

      <!-- How it Works Section -->
      <div class="mt-16 pt-12 border-t border-slate-200 dark:border-slate-700">
        <div class="text-center mb-12">
          <h2 class="mb-3 text-xl sm:text-2xl font-bold tracking-tight text-slate-900 dark:text-slate-100">
            {{ t('sign.howItWorks.title') }}
          </h2>
          <p class="text-slate-500 dark:text-slate-400 max-w-2xl mx-auto">
            {{ t('sign.howItWorks.subtitle') }}
          </p>
        </div>

        <!-- Steps Grid -->
        <div class="grid gap-6 sm:gap-8 grid-cols-1 md:grid-cols-3 mb-12">
          <!-- Step 1 -->
          <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 text-center hover:shadow-md transition-shadow">
            <div class="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-xl bg-blue-50 dark:bg-blue-900/30">
              <FileText :size="24" class="text-blue-600 dark:text-blue-400" />
            </div>
            <h3 class="mb-2 text-lg font-semibold text-slate-900 dark:text-slate-100">{{ t('sign.howItWorks.step1.title') }}</h3>
            <p class="text-sm text-slate-500 dark:text-slate-400">
              {{ t('sign.howItWorks.step1.description', { code: '?doc=URL' }) }}
            </p>
          </div>

          <!-- Step 2 -->
          <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 text-center hover:shadow-md transition-shadow">
            <div class="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-xl bg-blue-50 dark:bg-blue-900/30">
              <Shield :size="24" class="text-blue-600 dark:text-blue-400" />
            </div>
            <h3 class="mb-2 text-lg font-semibold text-slate-900 dark:text-slate-100">{{ t('sign.howItWorks.step2.title') }}</h3>
            <p class="text-sm text-slate-500 dark:text-slate-400">
              {{ t('sign.howItWorks.step2.description') }}
            </p>
          </div>

          <!-- Step 3 -->
          <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-6 text-center hover:shadow-md transition-shadow">
            <div class="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-xl bg-emerald-50 dark:bg-emerald-900/30">
              <CheckCircle2 :size="24" class="text-emerald-600 dark:text-emerald-400" />
            </div>
            <h3 class="mb-2 text-lg font-semibold text-slate-900 dark:text-slate-100">{{ t('sign.howItWorks.step3.title') }}</h3>
            <p class="text-sm text-slate-500 dark:text-slate-400">
              {{ t('sign.howItWorks.step3.description') }}
            </p>
          </div>
        </div>

        <!-- Features -->
        <div class="grid gap-6 grid-cols-1 md:grid-cols-3">
          <div class="flex items-start space-x-3">
            <div class="rounded-lg bg-blue-50 dark:bg-blue-900/30 p-2 mt-1 flex-shrink-0">
              <Shield :size="20" class="text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <h4 class="font-medium text-slate-900 dark:text-slate-100 mb-1">{{ t('sign.howItWorks.features.crypto.title') }}</h4>
              <p class="text-sm text-slate-500 dark:text-slate-400">
                {{ t('sign.howItWorks.features.crypto.description') }}
              </p>
            </div>
          </div>

          <div class="flex items-start space-x-3">
            <div class="rounded-lg bg-blue-50 dark:bg-blue-900/30 p-2 mt-1 flex-shrink-0">
              <Zap :size="20" class="text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <h4 class="font-medium text-slate-900 dark:text-slate-100 mb-1">{{ t('sign.howItWorks.features.instant.title') }}</h4>
              <p class="text-sm text-slate-500 dark:text-slate-400">
                {{ t('sign.howItWorks.features.instant.description') }}
              </p>
            </div>
          </div>

          <div class="flex items-start space-x-3">
            <div class="rounded-lg bg-blue-50 dark:bg-blue-900/30 p-2 mt-1 flex-shrink-0">
              <Clock :size="20" class="text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <h4 class="font-medium text-slate-900 dark:text-slate-100 mb-1">{{ t('sign.howItWorks.features.timestamp.title') }}</h4>
              <p class="text-sm text-slate-500 dark:text-slate-400">
                {{ t('sign.howItWorks.features.timestamp.description') }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
