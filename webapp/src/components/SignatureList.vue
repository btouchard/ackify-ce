<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<template>
  <div class="w-full">
    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center py-12">
      <svg
        class="animate-spin h-8 w-8 text-blue-600 dark:text-blue-400"
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
        />
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
      </svg>
    </div>

    <!-- Empty State -->
    <div v-else-if="signatures.length === 0" class="text-center py-12 px-4">
      <div class="w-16 h-16 mx-auto bg-slate-100 dark:bg-slate-800 rounded-2xl flex items-center justify-center mb-4">
        <svg
          class="h-8 w-8 text-slate-400"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
      </div>
      <p class="text-slate-500 dark:text-slate-400">{{ emptyMessage || $t('signatureList.empty') }}</p>
    </div>

    <!-- Signatures List -->
    <div v-else class="space-y-4">
      <div
        v-for="signature in signatures"
        :key="signature.id"
        :class="[
          'bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4 sm:p-5 hover:shadow-md transition-shadow',
          isDeleted ? 'opacity-60' : ''
        ]"
      >
        <div class="flex items-start justify-between gap-4">
          <div class="flex-1 min-w-0">
            <!-- Title Row -->
            <div class="flex flex-wrap items-center gap-2 mb-3">
              <h3 class="text-base sm:text-lg font-semibold text-slate-900 dark:text-white truncate">
                {{ signature.docTitle || signature.docId }}
              </h3>
              <!-- Status Badge -->
              <span
                v-if="!isDeleted"
                class="inline-flex items-center gap-1 px-2.5 py-1 bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 text-xs font-medium rounded-full"
              >
                <svg class="h-3.5 w-3.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                </svg>
                {{ $t('signatureList.confirmed') }}
              </span>
              <span
                v-else
                class="inline-flex items-center gap-1 px-2.5 py-1 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-400 text-xs font-medium rounded-full"
              >
                <svg class="h-3.5 w-3.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                {{ $t('signatureList.documentDeleted') }}{{ signature.docDeletedAt ? ` ${formatDate(signature.docDeletedAt)}` : '' }}
              </span>
            </div>

            <!-- Info Grid -->
            <div class="space-y-2 text-sm text-slate-600 dark:text-slate-400">
              <p v-if="signature.docTitle" class="flex items-start gap-2">
                <span class="text-xs font-medium text-slate-500 dark:text-slate-500 uppercase tracking-wide min-w-[60px]">{{ $t('signatureList.fields.id') }}</span>
                <span class="font-mono text-slate-700 dark:text-slate-300 break-all">{{ signature.docId }}</span>
              </p>
              <p v-if="signature.docUrl" class="flex items-start gap-2">
                <span class="text-xs font-medium text-slate-500 dark:text-slate-500 uppercase tracking-wide min-w-[60px]">{{ $t('signatureList.fields.document') }}</span>
                <a :href="signature.docUrl" target="_blank" rel="noopener noreferrer" class="text-blue-600 dark:text-blue-400 hover:underline break-all">
                  {{ signature.docUrl }}
                </a>
              </p>
              <p v-if="showUserInfo" class="flex items-start gap-2">
                <span class="text-xs font-medium text-slate-500 dark:text-slate-500 uppercase tracking-wide min-w-[60px]">{{ $t('signatureList.fields.reader') }}</span>
                <span class="text-slate-700 dark:text-slate-300">{{ signature.userName || signature.userEmail }}</span>
              </p>
              <p class="flex items-start gap-2">
                <span class="text-xs font-medium text-slate-500 dark:text-slate-500 uppercase tracking-wide min-w-[60px]">{{ $t('signatureList.fields.date') }}</span>
                <span class="text-slate-700 dark:text-slate-300">{{ formatDate(signature.signedAt) }}</span>
              </p>
              <p v-if="signature.serviceInfo" class="flex items-start gap-2">
                <span class="text-xs font-medium text-slate-500 dark:text-slate-500 uppercase tracking-wide min-w-[60px]">{{ $t('signatureList.fields.source') }}</span>
                <span class="inline-flex items-center gap-1.5 text-slate-700 dark:text-slate-300">
                  <span v-html="signature.serviceInfo.icon"></span>
                  <span>{{ signature.serviceInfo.name }}</span>
                </span>
              </p>
            </div>

            <!-- Verification Details -->
            <div v-if="showDetails" class="mt-4 pt-4 border-t border-slate-200 dark:border-slate-700">
              <details class="text-xs text-slate-500 dark:text-slate-400 group">
                <summary class="cursor-pointer hover:text-slate-700 dark:hover:text-slate-300 font-medium flex items-center gap-1.5">
                  <svg class="h-4 w-4 transition-transform group-open:rotate-90" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                  </svg>
                  {{ $t('signatureList.verificationDetails') }}
                </summary>
                <div class="mt-3 accent-border bg-slate-50 dark:bg-slate-900/50 rounded-r-lg p-3 space-y-1.5 font-mono text-xs">
                  <p><span class="font-semibold text-slate-600 dark:text-slate-400">{{ $t('signatureList.fields.id') }}</span> {{ signature.id }}</p>
                  <p><span class="font-semibold text-slate-600 dark:text-slate-400">{{ $t('signatureList.fields.nonce') }}</span> {{ signature.nonce }}</p>
                  <p class="break-all">
                    <span class="font-semibold text-slate-600 dark:text-slate-400">{{ $t('signatureList.fields.hash') }}</span> {{ signature.payloadHash }}
                  </p>
                  <p class="break-all">
                    <span class="font-semibold text-slate-600 dark:text-slate-400">{{ $t('signatureList.confirmation') }}</span>
                    {{ signature.signature.substring(0, 64) }}...
                  </p>
                  <p v-if="signature.prevHash" class="break-all">
                    <span class="font-semibold text-slate-600 dark:text-slate-400">{{ $t('signatureList.previousHash') }}</span> {{ signature.prevHash }}
                  </p>
                </div>
              </details>
            </div>
          </div>

          <!-- Actions -->
          <div v-if="showActions" class="flex-shrink-0">
            <button
              @click="$emit('view-details', signature)"
              class="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 text-sm font-medium px-3 py-1.5 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
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
