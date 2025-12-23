<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { listWebhooks, deleteWebhook, toggleWebhook, type Webhook } from '@/services/webhooks'
import { extractError } from '@/services/http'
import { Loader2, Plus, Pencil, Trash2, ToggleLeft, ToggleRight, BadgeCheck, Webhook as WebhookIcon } from 'lucide-vue-next'

const router = useRouter()
const { t } = useI18n()

const loading = ref(true)
const error = ref('')
const items = ref<Webhook[]>([])
const deleting = ref<number | null>(null)
const toggling = ref<number | null>(null)

async function load() {
  try {
    loading.value = true
    error.value = ''
    const resp = await listWebhooks()
    items.value = resp.data || []
  } catch (err) {
    error.value = extractError(err)
  } finally {
    loading.value = false
  }
}

function gotoNew() { router.push({ name: 'admin-webhook-new' }) }
function gotoEdit(id: number) { router.push({ name: 'admin-webhook-edit', params: { id } }) }

async function onDelete(id: number) {
  if (!confirm(t('admin.webhooks.confirmDelete'))) return
  try {
    deleting.value = id
    await deleteWebhook(id)
    await load()
  } catch (err) {
    error.value = extractError(err)
  } finally {
    deleting.value = null
  }
}

async function onToggle(id: number, enable: boolean) {
  try {
    toggling.value = id
    await toggleWebhook(id, enable)
    await load()
  } catch (err) {
    error.value = extractError(err)
  } finally {
    toggling.value = null
  }
}

function formatEvents(evts: string[] | null | undefined): string[] {
  if (!evts || !Array.isArray(evts)) return []
  return evts.map(e => t(`admin.webhooks.eventsMap.${e}`, e))
}

onMounted(load)
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 sm:px-6 py-6 sm:py-8">
    <!-- Page Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6 sm:mb-8">
      <div class="flex items-start gap-4">
        <div class="w-12 h-12 sm:w-14 sm:h-14 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
          <WebhookIcon class="w-6 h-6 sm:w-7 sm:h-7 text-blue-600 dark:text-blue-400" />
        </div>
        <div>
          <h1 class="text-xl sm:text-2xl font-bold text-slate-900 dark:text-white">{{ t('admin.webhooks.title') }}</h1>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">{{ t('admin.webhooks.subtitle') }}</p>
        </div>
      </div>
      <button
        @click="gotoNew"
        class="w-full sm:w-auto inline-flex items-center justify-center gap-2 trust-gradient text-white font-medium rounded-lg px-4 py-2.5 hover:opacity-90 transition-opacity min-h-[44px]"
      >
        <Plus :size="18" />
        {{ t('admin.webhooks.new') }}
      </button>
    </div>

    <!-- Error Alert -->
    <div v-if="error" class="mb-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl p-4">
      <div class="flex items-start gap-3">
        <div class="w-8 h-8 rounded-lg bg-red-100 dark:bg-red-900/30 flex items-center justify-center flex-shrink-0">
          <svg class="w-4 h-4 text-red-600 dark:text-red-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
        </div>
        <p class="text-red-700 dark:text-red-400 text-sm">{{ error }}</p>
      </div>
    </div>

    <!-- Main Card -->
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700">
      <!-- Card Header -->
      <div class="p-4 sm:p-6 border-b border-slate-200 dark:border-slate-700">
        <h2 class="font-semibold text-slate-900 dark:text-white">{{ t('admin.webhooks.listTitle') }}</h2>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">{{ t('admin.webhooks.listSubtitle') }}</p>
      </div>

      <!-- Content -->
      <div class="p-4 sm:p-6">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center gap-3 py-12">
          <Loader2 :size="24" class="animate-spin text-blue-600 dark:text-blue-400" />
          <span class="text-slate-500 dark:text-slate-400">{{ t('admin.loading') }}</span>
        </div>

        <!-- Content -->
        <div v-else>
          <!-- Desktop Table -->
          <div v-if="items.length > 0" class="hidden md:block overflow-x-auto">
            <table class="w-full">
              <thead>
                <tr class="border-b border-slate-200 dark:border-slate-700">
                  <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('admin.webhooks.columns.title') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('admin.webhooks.columns.url') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('admin.webhooks.columns.events') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('admin.webhooks.columns.status') }}</th>
                  <th class="px-4 py-3 text-right text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">{{ t('admin.webhooks.columns.actions') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
                <tr v-for="wh in items" :key="wh.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors">
                  <td class="px-4 py-4">
                    <div class="font-medium text-slate-900 dark:text-white">{{ wh.title || '-' }}</div>
                    <div v-if="wh.description" class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{{ wh.description }}</div>
                  </td>
                  <td class="px-4 py-4">
                    <a :href="wh.targetUrl" target="_blank" class="text-blue-600 dark:text-blue-400 hover:underline text-sm font-mono break-all">{{ wh.targetUrl }}</a>
                  </td>
                  <td class="px-4 py-4">
                    <div class="flex flex-wrap gap-1">
                      <span v-for="e in formatEvents(wh.events)" :key="e" class="px-2 py-0.5 text-xs rounded-full bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300">{{ e }}</span>
                    </div>
                  </td>
                  <td class="px-4 py-4">
                    <span v-if="wh.active" class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 text-xs font-medium rounded-full">
                      <BadgeCheck :size="14" />
                      {{ t('admin.webhooks.status.enabled') }}
                    </span>
                    <span v-else class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400 text-xs font-medium rounded-full">
                      {{ t('admin.webhooks.status.disabled') }}
                    </span>
                  </td>
                  <td class="px-4 py-4">
                    <div class="flex items-center justify-end gap-2">
                      <button
                        @click="gotoEdit(wh.id)"
                        class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-slate-600 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
                      >
                        <Pencil :size="14" />
                        {{ t('admin.webhooks.edit') }}
                      </button>
                      <button
                        @click="onToggle(wh.id, !wh.active)"
                        :disabled="toggling === wh.id"
                        class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-slate-600 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors disabled:opacity-50"
                      >
                        <Loader2 v-if="toggling === wh.id" :size="14" class="animate-spin" />
                        <ToggleRight v-else-if="!wh.active" :size="14" />
                        <ToggleLeft v-else :size="14" />
                        {{ wh.active ? t('admin.webhooks.disable') : t('admin.webhooks.enable') }}
                      </button>
                      <button
                        @click="onDelete(wh.id)"
                        :disabled="deleting === wh.id"
                        class="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-red-600 dark:text-red-400 bg-white dark:bg-slate-800 border border-red-200 dark:border-red-800 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors disabled:opacity-50"
                      >
                        <Loader2 v-if="deleting === wh.id" :size="14" class="animate-spin" />
                        <Trash2 v-else :size="14" />
                        {{ t('admin.webhooks.delete') }}
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Mobile Cards -->
          <div v-if="items.length > 0" class="md:hidden space-y-4">
            <div
              v-for="wh in items"
              :key="wh.id"
              class="bg-slate-50 dark:bg-slate-700/50 rounded-xl p-4"
            >
              <!-- Header -->
              <div class="flex items-start justify-between gap-3 mb-3">
                <div class="min-w-0">
                  <h3 class="font-medium text-slate-900 dark:text-white truncate">{{ wh.title || '-' }}</h3>
                  <p v-if="wh.description" class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">{{ wh.description }}</p>
                </div>
                <span v-if="wh.active" class="inline-flex items-center gap-1 px-2 py-0.5 bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 text-xs font-medium rounded-full flex-shrink-0">
                  <BadgeCheck :size="12" />
                  {{ t('admin.webhooks.status.enabled') }}
                </span>
                <span v-else class="inline-flex items-center px-2 py-0.5 bg-slate-200 dark:bg-slate-600 text-slate-500 dark:text-slate-400 text-xs font-medium rounded-full flex-shrink-0">
                  {{ t('admin.webhooks.status.disabled') }}
                </span>
              </div>

              <!-- URL -->
              <div class="mb-3">
                <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1">{{ t('admin.webhooks.columns.url') }}</p>
                <a :href="wh.targetUrl" target="_blank" class="text-blue-600 dark:text-blue-400 hover:underline text-sm font-mono break-all">{{ wh.targetUrl }}</a>
              </div>

              <!-- Events -->
              <div class="mb-4">
                <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1.5">{{ t('admin.webhooks.columns.events') }}</p>
                <div class="flex flex-wrap gap-1">
                  <span v-for="e in formatEvents(wh.events)" :key="e" class="px-2 py-0.5 text-xs rounded-full bg-white dark:bg-slate-600 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-500">{{ e }}</span>
                </div>
              </div>

              <!-- Actions -->
              <div class="flex flex-wrap gap-2 pt-3 border-t border-slate-200 dark:border-slate-600">
                <button
                  @click="gotoEdit(wh.id)"
                  class="inline-flex items-center gap-1.5 px-3 py-2 text-sm font-medium text-slate-600 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors min-h-[44px]"
                >
                  <Pencil :size="14" />
                  {{ t('admin.webhooks.edit') }}
                </button>
                <button
                  @click="onToggle(wh.id, !wh.active)"
                  :disabled="toggling === wh.id"
                  class="inline-flex items-center gap-1.5 px-3 py-2 text-sm font-medium text-slate-600 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors disabled:opacity-50 min-h-[44px]"
                >
                  <Loader2 v-if="toggling === wh.id" :size="14" class="animate-spin" />
                  <ToggleRight v-else-if="!wh.active" :size="14" />
                  <ToggleLeft v-else :size="14" />
                  {{ wh.active ? t('admin.webhooks.disable') : t('admin.webhooks.enable') }}
                </button>
                <button
                  @click="onDelete(wh.id)"
                  :disabled="deleting === wh.id"
                  class="inline-flex items-center gap-1.5 px-3 py-2 text-sm font-medium text-red-600 dark:text-red-400 bg-white dark:bg-slate-800 border border-red-200 dark:border-red-800 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors disabled:opacity-50 min-h-[44px]"
                >
                  <Loader2 v-if="deleting === wh.id" :size="14" class="animate-spin" />
                  <Trash2 v-else :size="14" />
                  {{ t('admin.webhooks.delete') }}
                </button>
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-if="items.length === 0" class="py-12 text-center">
            <div class="w-16 h-16 mx-auto bg-slate-100 dark:bg-slate-700 rounded-2xl flex items-center justify-center mb-4">
              <WebhookIcon class="w-8 h-8 text-slate-400" />
            </div>
            <p class="text-slate-500 dark:text-slate-400">{{ t('admin.webhooks.empty') }}</p>
            <button
              @click="gotoNew"
              class="mt-4 inline-flex items-center gap-2 trust-gradient text-white font-medium rounded-lg px-4 py-2.5 hover:opacity-90 transition-opacity min-h-[44px]"
            >
              <Plus :size="18" />
              {{ t('admin.webhooks.new') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
