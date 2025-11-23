<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSignatureStore } from '@/stores/signatures'
import { FileSignature, FileCheck, Clock, Search, Info } from 'lucide-vue-next'
import Card from '@/components/ui/Card.vue'
import CardHeader from '@/components/ui/CardHeader.vue'
import CardTitle from '@/components/ui/CardTitle.vue'
import CardDescription from '@/components/ui/CardDescription.vue'
import CardContent from '@/components/ui/CardContent.vue'
import Input from '@/components/ui/Input.vue'
import Alert from '@/components/ui/Alert.vue'
import AlertDescription from '@/components/ui/AlertDescription.vue'
import SignatureList from '@/components/SignatureList.vue'
import {usePageTitle} from "@/composables/usePageTitle.ts";

const { t } = useI18n()
usePageTitle('signatures.title')
const signatureStore = useSignatureStore()
const searchQuery = ref('')

const filteredSignatures = computed(() => {
  if (!searchQuery.value.trim()) {
    return signatureStore.userSignatures
  }

  const query = searchQuery.value.toLowerCase()
  return signatureStore.userSignatures.filter((sig: any) =>
    sig.docId.toLowerCase().includes(query) ||
    sig.docTitle?.toLowerCase().includes(query) ||
    sig.docUrl?.toLowerCase().includes(query)
  )
})

const activeSignatures = computed(() => {
  return filteredSignatures.value.filter((sig: any) => !sig.docDeletedAt)
})

const deletedSignatures = computed(() => {
  return filteredSignatures.value.filter((sig: any) => sig.docDeletedAt)
})

const uniqueDocumentsCount = computed(() => {
  const docIds = new Set(signatureStore.userSignatures.map((sig: any) => sig.docId))
  return docIds.size
})

const lastSignatureDate = computed(() => {
  if (signatureStore.userSignatures.length === 0) return null

  const latest = signatureStore.userSignatures[0]
  if (!latest) return null
  const date = new Date(latest.signedAt)
  return date.toLocaleDateString('fr-FR', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
})

async function refreshSignatures() {
  try {
    await signatureStore.fetchUserSignatures()
  } catch (error) {
    console.error('Failed to refresh signatures:', error)
  }
}

onMounted(() => {
  refreshSignatures()
})
</script>

<template>
  <div class="relative min-h-[calc(100vh-4rem)]">
    <!-- Background decoration -->
    <div class="absolute inset-0 -z-10 overflow-hidden">
      <div class="absolute left-1/3 top-0 h-[400px] w-[400px] rounded-full bg-primary/5 blur-3xl"></div>
      <div class="absolute right-1/3 bottom-0 h-[400px] w-[400px] rounded-full bg-primary/5 blur-3xl"></div>
    </div>

    <!-- Main Content -->
    <main class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
      <!-- Page Header -->
      <div class="mb-8">
        <h1 class="mb-2 text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
          {{ t('signatures.title') }}
        </h1>
        <p class="text-lg text-muted-foreground">
          {{ t('signatures.subtitle') }}
        </p>
      </div>

      <!-- Stats Pills Mobile (compact horizontal full-width) -->
      <div class="sm:hidden mb-6 grid grid-cols-3 gap-3">
        <div class="flex flex-col items-center justify-center gap-1 px-3 py-3 rounded-lg bg-primary/10 text-primary">
          <FileSignature :size="18" />
          <span class="text-xl font-bold">{{ signatureStore.getUserSignaturesCount }}</span>
          <span class="text-xs whitespace-nowrap">{{ t('signatures.stats.total') }}</span>
        </div>
        <div class="flex flex-col items-center justify-center gap-1 px-3 py-3 rounded-lg bg-green-500/10 text-green-600 dark:text-green-400">
          <FileCheck :size="18" />
          <span class="text-xl font-bold">{{ uniqueDocumentsCount }}</span>
          <span class="text-xs whitespace-nowrap">{{ t('signatures.stats.unique') }}</span>
        </div>
        <div class="flex flex-col items-center justify-center gap-1 px-3 py-3 rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400">
          <Clock :size="18" />
          <span class="text-sm font-bold">{{ lastSignatureDate || t('signatures.stats.notAvailable') }}</span>
          <span class="text-xs whitespace-nowrap">{{ t('signatures.stats.last') }}</span>
        </div>
      </div>

      <!-- Stats Cards Desktop -->
      <div class="hidden sm:grid mb-8 gap-6 sm:grid-cols-2 lg:grid-cols-3">
        <!-- Total Confirmations -->
        <Card class="clay-card-hover">
          <CardContent class="pt-6">
            <div class="flex items-center space-x-4">
              <div class="rounded-lg bg-primary/10 p-3">
                <FileSignature :size="24" class="text-primary" />
              </div>
              <div class="flex-1">
                <p class="text-sm font-medium text-muted-foreground">{{ t('signatures.stats.totalConfirmations') }}</p>
                <p class="text-2xl font-bold text-foreground">
                  {{ signatureStore.getUserSignaturesCount }}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Unique Documents -->
        <Card class="clay-card-hover">
          <CardContent class="pt-6">
            <div class="flex items-center space-x-4">
              <div class="rounded-lg bg-green-500/10 p-3">
                <FileCheck :size="24" class="text-green-600 dark:text-green-400" />
              </div>
              <div class="flex-1">
                <p class="text-sm font-medium text-muted-foreground">{{ t('signatures.stats.uniqueDocuments') }}</p>
                <p class="text-2xl font-bold text-foreground">{{ uniqueDocumentsCount }}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <!-- Last Confirmation -->
        <Card class="clay-card-hover">
          <CardContent class="pt-6">
            <div class="flex items-center space-x-4">
              <div class="rounded-lg bg-blue-500/10 p-3">
                <Clock :size="24" class="text-blue-600 dark:text-blue-400" />
              </div>
              <div class="flex-1">
                <p class="text-sm font-medium text-muted-foreground">{{ t('signatures.stats.lastConfirmation') }}</p>
                <p class="text-lg font-semibold text-foreground">
                  {{ lastSignatureDate || t('signatures.stats.notAvailable') }}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <!-- Signatures List -->
      <Card class="clay-card">
        <CardHeader>
          <div class="flex flex-col gap-4">
            <div class="flex items-center justify-between">
              <div>
                <CardTitle>{{ t('signatures.allConfirmations') }}</CardTitle>
                <CardDescription class="mt-2">
                  {{ t('signatures.results', { count: filteredSignatures.length }) }}
                </CardDescription>
              </div>
            </div>

            <!-- Search -->
            <div class="relative">
              <Search :size="18" class="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
              <Input
                v-model="searchQuery"
                type="text"
                :placeholder="t('signatures.search')"
                class="pl-10"
              />
            </div>
          </div>
        </CardHeader>

        <CardContent>
          <div v-if="signatureStore.loading" class="flex justify-center py-8">
            <svg
              class="animate-spin h-8 w-8 text-primary"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>

          <div v-else-if="filteredSignatures.length === 0" class="text-center py-8">
            <svg
              class="mx-auto h-12 w-12 text-muted-foreground"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <p class="mt-2 text-muted-foreground">{{ t('signatures.empty.alternative') }}</p>
          </div>

          <div v-else class="space-y-4">
            <!-- Active signatures -->
            <SignatureList
              v-if="activeSignatures.length > 0"
              :signatures="activeSignatures"
              :loading="false"
              :show-user-info="false"
              :show-details="true"
              :show-actions="false"
              :is-deleted="false"
            />

            <!-- Deleted documents section header -->
            <div v-if="deletedSignatures.length > 0" class="py-4">
              <hr v-if="activeSignatures.length > 0" class="border-border" />
              <p class="text-center text-sm text-muted-foreground mt-4 mb-2">
                {{ t('signatures.deletedDocuments') }}
              </p>
            </div>

            <!-- Deleted signatures -->
            <SignatureList
              v-if="deletedSignatures.length > 0"
              :signatures="deletedSignatures"
              :loading="false"
              :show-user-info="false"
              :show-details="true"
              :show-actions="false"
              :is-deleted="true"
            />
          </div>
        </CardContent>
      </Card>

      <!-- Info Card -->
      <Alert variant="info" class="mt-6 clay-card border-l-4">
        <div class="flex items-start">
          <Info :size="20" class="mr-3 mt-0.5" />
          <div class="flex-1">
            <h3 class="mb-2 font-medium">{{ t('signatures.about.title') }}</h3>
            <AlertDescription>
              {{ t('signatures.about.description') }}
            </AlertDescription>
          </div>
        </div>
      </Alert>
    </main>
  </div>
</template>