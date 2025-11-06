<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<template>
  <div class="signature-list">
    <div v-if="loading" class="flex justify-center py-8">
      <svg
        class="animate-spin h-8 w-8 text-primary"
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
      >
        <circle
          class="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          stroke-width="4"
        ></circle>
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        ></path>
      </svg>
    </div>

    <div v-else-if="signatures.length === 0" class="empty-state">
      <svg
        class="mx-auto h-12 w-12 text-muted-foreground"
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        stroke="currentColor"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
        />
      </svg>
      <p class="mt-2 text-muted-foreground">{{ emptyMessage || $t('signatureList.empty') }}</p>
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="signature in signatures"
        :key="signature.id"
        :class="[
          'signature-card shadow rounded-lg p-4 hover:shadow-md transition-shadow bg-card text-card-foreground border border-border',
          isDeleted ? 'opacity-50' : ''
        ]"
      >
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <div class="flex items-center space-x-2">
              <h3 class="text-lg font-medium text-foreground">
                {{ signature.docTitle || signature.docId }}
              </h3>
              <span
                v-if="!isDeleted"
                class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400"
              >
                <svg
                  class="mr-1 h-3 w-3"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M5 13l4 4L19 7"
                  />
                </svg>
                {{ $t('signatureList.confirmed') }}
              </span>
              <span
                v-else
                class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400"
              >
                <svg
                  class="mr-1 h-3 w-3"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
                {{ $t('signatureList.documentDeleted') }}{{ signature.docDeletedAt ? ` ${formatDate(signature.docDeletedAt)}` : '' }}
              </span>
            </div>

            <div class="mt-2 space-y-1 text-sm text-muted-foreground">
              <p v-if="signature.docTitle">
                <span class="font-medium">{{ $t('signatureList.fields.id') }}</span> {{ signature.docId }}
              </p>
              <p v-if="signature.docUrl">
                <span class="font-medium">{{ $t('signatureList.fields.document') }}</span>
                <a :href="signature.docUrl" target="_blank" rel="noopener noreferrer" class="text-primary hover:text-primary/80 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background rounded">
                  {{ signature.docUrl }}
                </a>
              </p>
              <p v-if="showUserInfo">
                <span class="font-medium">{{ $t('signatureList.fields.reader') }}</span> {{ signature.userName || signature.userEmail }}
              </p>
              <p>
                <span class="font-medium">{{ $t('signatureList.fields.date') }}</span> {{ formatDate(signature.signedAt) }}
              </p>
              <p v-if="signature.serviceInfo" class="flex items-center">
                <span class="font-medium mr-2">{{ $t('signatureList.fields.source') }}</span>
                <span class="inline-flex items-center space-x-1">
                  <span v-html="signature.serviceInfo.icon"></span>
                  <span>{{ signature.serviceInfo.name }}</span>
                </span>
              </p>
            </div>

            <div v-if="showDetails" class="mt-3 pt-3 border-t border-border">
              <details class="text-xs text-muted-foreground">
                <summary class="cursor-pointer hover:text-foreground font-medium focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background rounded">
                  {{ $t('signatureList.verificationDetails') }}
                </summary>
                <div class="mt-2 space-y-1 font-mono bg-muted p-2 rounded border border-border">
                  <p><span class="font-semibold">{{ $t('signatureList.fields.id') }}</span> {{ signature.id }}</p>
                  <p><span class="font-semibold">{{ $t('signatureList.fields.nonce') }}</span> {{ signature.nonce }}</p>
                  <p class="break-all">
                    <span class="font-semibold">{{ $t('signatureList.fields.hash') }}</span> {{ signature.payloadHash }}
                  </p>
                  <p class="break-all">
                    <span class="font-semibold">{{ $t('signatureList.confirmation') }}</span>
                    {{ signature.signature.substring(0, 64) }}...
                  </p>
                  <p v-if="signature.prevHash" class="break-all">
                    <span class="font-semibold">{{ $t('signatureList.previousHash') }}</span> {{ signature.prevHash }}
                  </p>
                </div>
              </details>
            </div>
          </div>

          <div v-if="showActions" class="ml-4">
            <button
              @click="$emit('view-details', signature)"
              class="text-primary hover:text-primary/80 text-sm font-medium focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background rounded px-2 py-1"
            >
              {{ $t('signatureList.viewDetails') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Signature } from '@/services/signatures'

interface Props {
  signatures: Signature[]
  loading?: boolean
  showUserInfo?: boolean
  showDetails?: boolean
  showActions?: boolean
  emptyMessage?: string
  isDeleted?: boolean
}

withDefaults(defineProps<Props>(), {
  loading: false,
  showUserInfo: false,
  showDetails: true,
  showActions: false,
  isDeleted: false,
})

defineEmits<{
  'view-details': [signature: Signature]
}>()

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleDateString('fr-FR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<style scoped>
.signature-list {
  width: 100%;
}

.empty-state {
  text-align: center;
  padding: 3rem 1rem;
}
</style>
