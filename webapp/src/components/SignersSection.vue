<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ExpectedSigner, UnexpectedSignature, DocumentStats } from '@/services/admin'
import {
  Users,
  CheckCircle,
  Plus,
  Trash2,
  Upload,
  AlertTriangle,
  Search,
} from 'lucide-vue-next'

interface Props {
  expectedSigners: ExpectedSigner[]
  unexpectedSignatures: UnexpectedSignature[]
  stats: DocumentStats | null
  showImportCSV?: boolean
  selectedEmails?: string[]
}

const props = withDefaults(defineProps<Props>(), {
  showImportCSV: false,
  selectedEmails: () => [],
})

const emit = defineEmits<{
  (e: 'add-signer'): void
  (e: 'remove-signer', email: string): void
  (e: 'import-csv'): void
  (e: 'selection-change', emails: string[]): void
}>()

const { t, locale } = useI18n()

// Local state
const signerFilter = ref('')
const localSelectedEmails = ref<string[]>([...props.selectedEmails])

// Computed
const filteredSigners = computed(() => {
  const filter = signerFilter.value.toLowerCase().trim()
  if (!filter) return props.expectedSigners
  return props.expectedSigners.filter(signer =>
    signer.email.toLowerCase().includes(filter) ||
    (signer.name && signer.name.toLowerCase().includes(filter)) ||
    (signer.userName && signer.userName.toLowerCase().includes(filter))
  )
})

// Methods
function formatDate(dateString: string | undefined): string {
  if (!dateString) return '-'
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
  const index = localSelectedEmails.value.indexOf(email)
  if (index > -1) {
    localSelectedEmails.value.splice(index, 1)
  } else {
    localSelectedEmails.value.push(email)
  }
  emit('selection-change', [...localSelectedEmails.value])
}

function selectAllPending(checked: boolean) {
  if (checked) {
    localSelectedEmails.value = props.expectedSigners
      .filter(s => !s.hasSigned)
      .map(s => s.email)
  } else {
    localSelectedEmails.value = []
  }
  emit('selection-change', [...localSelectedEmails.value])
}

function isSelected(email: string): boolean {
  return localSelectedEmails.value.includes(email)
}
</script>

<template>
  <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700">
    <div class="p-6 border-b border-slate-100 dark:border-slate-700">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h2 class="font-semibold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <CheckCircle :size="18" class="text-emerald-600 dark:text-emerald-400" />
            {{ t('admin.documentDetail.readers') }}
          </h2>
          <p v-if="stats" class="text-sm text-slate-500 dark:text-slate-400">
            {{ stats.signedCount }} / {{ stats.expectedCount }} {{ t('admin.dashboard.stats.signed').toLowerCase() }}
          </p>
        </div>
        <div class="flex gap-2">
          <button
            v-if="showImportCSV"
            @click="emit('import-csv')"
            class="inline-flex items-center gap-2 bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-200 font-medium rounded-lg px-3 py-2 text-sm hover:bg-slate-50 dark:hover:bg-slate-600 transition-colors"
          >
            <Upload :size="16" />
            {{ t('admin.documentDetail.importCSV') }}
          </button>
          <button
            @click="emit('add-signer')"
            data-testid="open-add-signers-btn"
            class="trust-gradient text-white font-medium rounded-lg px-3 py-2 text-sm hover:opacity-90 transition-opacity inline-flex items-center gap-2"
          >
            <Plus :size="16" />
            {{ t('admin.documentDetail.addButton') }}
          </button>
        </div>
      </div>
    </div>
    <div class="p-6">
      <div v-if="expectedSigners.length > 0">
        <div class="relative mb-4">
          <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
          <input
            v-model="signerFilter"
            :placeholder="t('admin.documentDetail.filterPlaceholder')"
            class="w-full pl-9 pr-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            name="ackify-signer-filter"
            autocomplete="off"
            data-1p-ignore
            data-lpignore="true"
          />
        </div>

        <!-- Table Desktop -->
        <div class="hidden md:block overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-slate-100 dark:border-slate-700">
                <th class="px-4 py-3 w-10">
                  <input
                    type="checkbox"
                    class="rounded border-slate-300 dark:border-slate-600"
                    @change="(e: any) => selectAllPending(e.target.checked)"
                  />
                </th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                  {{ t('admin.documentDetail.reader') }}
                </th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                  {{ t('admin.documentDetail.status') }}
                </th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                  {{ t('admin.documentDetail.confirmedOn') }}
                </th>
                <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                  {{ t('common.actions') }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr v-for="signer in filteredSigners" :key="signer.email" class="hover:bg-slate-50 dark:hover:bg-slate-700/50">
                <td class="px-4 py-3">
                  <input
                    v-if="!signer.hasSigned"
                    type="checkbox"
                    class="rounded border-slate-300 dark:border-slate-600"
                    :checked="isSelected(signer.email)"
                    @change="toggleEmailSelection(signer.email)"
                  />
                </td>
                <td class="px-4 py-3">
                  <div>
                    <p class="font-medium text-slate-900 dark:text-slate-100">
                      {{ signer.userName || signer.name || signer.email }}
                    </p>
                    <p class="text-xs text-slate-500 dark:text-slate-400">{{ signer.email }}</p>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <span
                    :class="[
                      'inline-flex items-center px-2.5 py-1 text-xs font-medium rounded-full',
                      signer.hasSigned
                        ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                        : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400'
                    ]"
                  >
                    {{ signer.hasSigned ? t('admin.documentDetail.statusConfirmed') : t('admin.documentDetail.statusPending') }}
                  </span>
                </td>
                <td class="px-4 py-3 text-sm text-slate-500 dark:text-slate-400">
                  {{ formatDate(signer.signedAt) }}
                </td>
                <td class="px-4 py-3">
                  <button
                    v-if="!signer.hasSigned"
                    @click="emit('remove-signer', signer.email)"
                    class="p-1.5 rounded-md hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                  >
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
                <input
                  v-if="!signer.hasSigned"
                  type="checkbox"
                  class="mt-1 rounded border-slate-300 dark:border-slate-600"
                  :checked="isSelected(signer.email)"
                  @change="toggleEmailSelection(signer.email)"
                />
                <div>
                  <p class="font-medium text-slate-900 dark:text-slate-100">
                    {{ signer.userName || signer.name || signer.email }}
                  </p>
                  <p class="text-xs text-slate-500 dark:text-slate-400">{{ signer.email }}</p>
                </div>
              </div>
              <span
                :class="[
                  'inline-flex items-center px-2 py-0.5 text-xs font-medium rounded-full',
                  signer.hasSigned
                    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                    : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400'
                ]"
              >
                {{ signer.hasSigned ? t('admin.documentDetail.statusConfirmed') : t('admin.documentDetail.statusPending') }}
              </span>
            </div>
            <div class="flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
              <span>{{ formatDate(signer.signedAt) }}</span>
              <button
                v-if="!signer.hasSigned"
                @click="emit('remove-signer', signer.email)"
                class="p-1 text-red-600 dark:text-red-400"
              >
                <Trash2 :size="14" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="text-center py-8">
        <Users :size="48" class="mx-auto mb-4 text-slate-300 dark:text-slate-600" />
        <p class="text-slate-500 dark:text-slate-400">{{ t('admin.documentDetail.noExpectedSigners') }}</p>
      </div>

      <!-- Unexpected signatures -->
      <div v-if="unexpectedSignatures.length > 0" class="mt-8 pt-6 border-t border-slate-200 dark:border-slate-700">
        <h3 class="text-base font-semibold mb-4 flex items-center text-slate-900 dark:text-slate-100">
          <AlertTriangle :size="18" class="mr-2 text-amber-500" />
          {{ t('admin.documentDetail.unexpectedSignatures') }}
          <span class="ml-2 inline-flex items-center px-2 py-0.5 text-xs font-medium rounded-full bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400">
            {{ unexpectedSignatures.length }}
          </span>
        </h3>
        <p class="text-sm text-slate-500 dark:text-slate-400 mb-4">
          {{ t('admin.documentDetail.unexpectedDescription') }}
        </p>
        <div class="space-y-2">
          <div
            v-for="(sig, idx) in unexpectedSignatures"
            :key="idx"
            class="flex items-center justify-between p-3 bg-slate-50 dark:bg-slate-700/50 rounded-lg"
          >
            <div>
              <p class="font-medium text-slate-900 dark:text-slate-100">
                {{ sig.userName || sig.userEmail }}
              </p>
              <p class="text-xs text-slate-500 dark:text-slate-400">{{ sig.userEmail }}</p>
            </div>
            <span class="text-sm text-slate-500 dark:text-slate-400">
              {{ formatDate(sig.signedAtUTC) }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
