<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import {computed, onMounted, ref, watch} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useAuthStore} from '@/stores/auth'
import {useSignatureStore} from '@/stores/signatures'
import {useI18n} from 'vue-i18n'
import {usePageTitle} from '@/composables/usePageTitle'

const {t} = useI18n()
usePageTitle('sign.title')

import {AlertTriangle, CheckCircle2, FileText, Info, Users, Loader2, Shield, Zap, Clock} from 'lucide-vue-next'
import Card from '@/components/ui/Card.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import CardTitle from '@/components/ui/CardTitle.vue'
import CardDescription from '@/components/ui/CardDescription.vue'
import CardContent from '@/components/ui/CardContent.vue'
import Alert from '@/components/ui/Alert.vue'
import AlertTitle from '@/components/ui/AlertTitle.vue'
import AlertDescription from '@/components/ui/AlertDescription.vue'
import Button from '@/components/ui/Button.vue'
import SignButton from '@/components/SignButton.vue'
import SignatureList from '@/components/SignatureList.vue'
import {documentService, type FindOrCreateDocumentResponse} from '@/services/documents'
import {detectReference} from '@/services/referenceDetector'
import {calculateFileChecksum} from '@/services/checksumCalculator'
import {updateDocumentMetadata} from '@/services/admin'
import DocumentForm from "@/components/DocumentForm.vue";

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const signatureStore = useSignatureStore()

const docId = ref<string | undefined>(undefined)
const user = computed(() => authStore.user)
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
  <div class="relative">
    <!-- Background decoration -->
    <div class="absolute inset-0 -z-10 overflow-hidden">
      <div class="absolute left-1/4 top-0 h-[400px] w-[400px] rounded-full bg-primary/5 blur-3xl"></div>
      <div class="absolute right-1/4 bottom-0 h-[400px] w-[400px] rounded-full bg-primary/5 blur-3xl"></div>
    </div>

    <!-- Main Content -->
    <div class="mx-auto max-w-4xl px-4 py-12 sm:px-6 lg:px-8">
      <!-- Page Header -->
      <div class="mb-8 text-center">
        <h1 class="mb-2 text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
          {{ t('sign.title') }}
        </h1>
        <p class="text-lg text-muted-foreground">
          {{ t('sign.subtitle') }}
        </p>
      </div>

      <!-- Error Message (shown independently of docId state) -->
      <transition
          enter-active-class="transition ease-out duration-300"
          enter-from-class="opacity-0 translate-y-2"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition ease-in duration-200"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 translate-y-2"
      >
        <Alert v-if="errorMessage && !loadingDocument" variant="destructive" class="clay-card mb-6">
          <div class="flex items-start">
            <AlertTriangle :size="20" class="mr-3 mt-0.5"/>
            <div class="flex-1">
              <AlertTitle>{{ t('sign.error.title') }}</AlertTitle>
              <AlertDescription>{{ errorMessage }}</AlertDescription>
              <div v-if="needsAuth" class="mt-4">
                <Button @click="handleLoginClick" variant="default">
                  {{ t('sign.error.loginButton') }}
                </Button>
              </div>
            </div>
          </div>
        </Alert>
      </transition>

      <!-- Loading state -->
      <Card v-if="loadingDocument" class="clay-card">
        <CardContent class="py-12 text-center">
          <Loader2 :size="48" class="mx-auto mb-4 animate-spin text-primary"/>
          <h2 class="text-xl font-semibold mb-2">{{ t('sign.loading.title') }}</h2>
          <p class="text-muted-foreground">
            {{ t('sign.loading.description') }}
          </p>
        </CardContent>
      </Card>

      <!-- No Document: Show help message -->
      <Card v-else-if="!docId" class="clay-card">
        <CardContent class="py-12 text-center">
          <FileText :size="48" class="mx-auto mb-4 text-muted-foreground"/>
          <h2 class="text-xl font-semibold mb-2">{{ t('sign.noDocument.title') }}</h2>
          <p class="text-muted-foreground mb-4">
            {{ t('sign.noDocument.description', { code: '?doc=' }) }}
          </p>
          <div class="text-sm text-muted-foreground space-y-2">
            <p><strong>{{ t('sign.noDocument.examples') }}</strong></p>
            <code class="block px-3 py-2 bg-muted rounded text-xs">/?doc=https://example.com/policy.pdf</code>
            <code class="block px-3 py-2 bg-muted rounded text-xs">/?doc=/path/to/document</code>
            <code class="block px-3 py-2 bg-muted rounded text-xs">/?doc=my-unique-ref</code>
            <DocumentForm />
          </div>
        </CardContent>
      </Card>

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
          <Alert v-if="showSuccessMessage" variant="success" class="clay-card">
            <div class="flex items-start">
              <CheckCircle2 :size="20" class="mr-3 mt-0.5 text-green-600 dark:text-green-400"/>
              <div class="flex-1">
                <AlertTitle>{{ t('sign.success.title') }}</AlertTitle>
                <AlertDescription>
                  {{ t('sign.success.description') }}
                </AlertDescription>
              </div>
            </div>
          </Alert>
        </transition>

        <!-- Document Info Card -->
        <Card class="clay-card">
          <CardHeader>
            <div class="flex items-start space-x-4">
              <div class="rounded-lg bg-primary/10 p-3">
                <FileText :size="28" class="text-primary"/>
              </div>
              <div class="flex-1">
                <CardTitle>
                  {{ t('sign.document.title') }}<template v-if="currentDocument?.title"> : {{ currentDocument.title }}</template>
                </CardTitle>
                <CardDescription class="mt-2">
                  <template v-if="currentDocument?.url">
                    <a
                      :href="currentDocument.url"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="text-primary hover:underline font-mono text-xs"
                    >
                      {{ currentDocument.url }}
                    </a>
                  </template>
                  <template v-else>
                    <span class="font-mono text-xs">{{ docId }}</span>
                  </template>
                </CardDescription>
              </div>
            </div>
          </CardHeader>

          <CardContent>
            <div class="space-y-4">
              <!-- Sign Button Component -->
              <div class="pb-4">
                <SignButton
                    :doc-id="docId"
                    :signatures="documentSignatures"
                    @signed="handleSigned"
                    @error="handleError"
                />
              </div>

              <!-- Info Box (only shown if user hasn't signed yet) -->
              <Alert v-if="!userHasSigned" variant="info" class="border-l-4">
                <div class="flex items-start">
                  <Info :size="18" class="mr-3 mt-0.5"/>
                  <div class="flex-1 space-y-2 text-sm">
                    <p>
                      {{ t('sign.info.description') }}
                    </p>
                    <p class="font-medium">
                      {{ t('sign.info.recorded') }}
                    </p>
                    <ul class="list-disc space-y-1 pl-5">
                      <li>{{ t('sign.info.email') }} : <strong class="text-foreground">{{ user?.email }}</strong></li>
                      <li>{{ t('sign.info.timestamp') }}</li>
                      <li>{{ t('sign.info.signature') }}</li>
                      <li>{{ t('sign.info.hash') }}</li>
                    </ul>
                  </div>
                </div>
              </Alert>
            </div>
          </CardContent>
        </Card>

        <!-- Existing Confirmations -->
        <Card v-if="documentSignatures.length > 0" class="clay-card">
          <CardHeader>
            <div class="flex items-center space-x-3">
              <div class="rounded-lg bg-primary/10 p-2">
                <Users :size="20" class="text-primary"/>
              </div>
              <div>
                <CardTitle>{{ t('sign.confirmations.title') }}</CardTitle>
                <CardDescription>
                  {{ t('sign.confirmations.count', { count: documentSignatures.length }, documentSignatures.length) }}
                  {{ t('sign.confirmations.recorded', {}, documentSignatures.length) }}
                </CardDescription>
              </div>
            </div>
          </CardHeader>

          <CardContent>
            <SignatureList
                :signatures="documentSignatures"
                :loading="loadingSignatures"
                :show-user-info="true"
                :show-details="true"
            />
          </CardContent>
        </Card>

        <!-- Empty State -->
        <Card v-else-if="!loadingSignatures" class="clay-card">
          <CardContent class="py-12 text-center">
            <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
              <Users :size="28" class="text-muted-foreground"/>
            </div>
            <h3 class="mb-2 text-lg font-semibold text-foreground">
              {{ t('sign.empty.title') }}
            </h3>
            <p class="text-sm text-muted-foreground">
              {{ t('sign.empty.description') }}
            </p>
          </CardContent>
        </Card>
      </div>

      <!-- How it Works Section (always visible) -->
      <div class="mt-16 pt-12 border-t border-border/40">
        <div class="text-center mb-12">
          <h2 class="mb-3 text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
            {{ t('sign.howItWorks.title') }}
          </h2>
          <p class="text-muted-foreground max-w-2xl mx-auto">
            {{ t('sign.howItWorks.subtitle') }}
          </p>
        </div>

        <!-- Steps Grid -->
        <div class="grid gap-8 md:grid-cols-3 mb-12">
          <!-- Step 1 -->
          <Card class="clay-card-hover text-center">
            <CardContent class="pt-6">
              <div class="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                <FileText :size="24" class="text-primary" />
              </div>
              <h3 class="mb-2 text-lg font-semibold text-foreground">{{ t('sign.howItWorks.step1.title') }}</h3>
              <p class="text-sm text-muted-foreground">
                {{ t('sign.howItWorks.step1.description', { code: '?doc=URL' }) }}
              </p>
            </CardContent>
          </Card>

          <!-- Step 2 -->
          <Card class="clay-card-hover text-center">
            <CardContent class="pt-6">
              <div class="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                <Shield :size="24" class="text-primary" />
              </div>
              <h3 class="mb-2 text-lg font-semibold text-foreground">{{ t('sign.howItWorks.step2.title') }}</h3>
              <p class="text-sm text-muted-foreground">
                {{ t('sign.howItWorks.step2.description') }}
              </p>
            </CardContent>
          </Card>

          <!-- Step 3 -->
          <Card class="clay-card-hover text-center">
            <CardContent class="pt-6">
              <div class="mb-4 inline-flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
                <CheckCircle2 :size="24" class="text-primary" />
              </div>
              <h3 class="mb-2 text-lg font-semibold text-foreground">{{ t('sign.howItWorks.step3.title') }}</h3>
              <p class="text-sm text-muted-foreground">
                {{ t('sign.howItWorks.step3.description') }}
              </p>
            </CardContent>
          </Card>
        </div>

        <!-- Features -->
        <div class="grid gap-6 md:grid-cols-3">
          <div class="flex items-start space-x-3">
            <div class="rounded-lg bg-primary/10 p-2 mt-1">
              <Shield :size="20" class="text-primary" />
            </div>
            <div>
              <h4 class="font-medium text-foreground mb-1">{{ t('sign.howItWorks.features.crypto.title') }}</h4>
              <p class="text-sm text-muted-foreground">
                {{ t('sign.howItWorks.features.crypto.description') }}
              </p>
            </div>
          </div>

          <div class="flex items-start space-x-3">
            <div class="rounded-lg bg-primary/10 p-2 mt-1">
              <Zap :size="20" class="text-primary" />
            </div>
            <div>
              <h4 class="font-medium text-foreground mb-1">{{ t('sign.howItWorks.features.instant.title') }}</h4>
              <p class="text-sm text-muted-foreground">
                {{ t('sign.howItWorks.features.instant.description') }}
              </p>
            </div>
          </div>

          <div class="flex items-start space-x-3">
            <div class="rounded-lg bg-primary/10 p-2 mt-1">
              <Clock :size="20" class="text-primary" />
            </div>
            <div>
              <h4 class="font-medium text-foreground mb-1">{{ t('sign.howItWorks.features.timestamp.title') }}</h4>
              <p class="text-sm text-muted-foreground">
                {{ t('sign.howItWorks.features.timestamp.description') }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
