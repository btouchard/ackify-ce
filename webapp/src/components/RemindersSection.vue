<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ReminderStats } from '@/services/admin'
import { Mail } from 'lucide-vue-next'

interface Props {
  reminderStats: ReminderStats
  smtpEnabled: boolean
  selectedEmailsCount: number
  sending?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  sending: false,
})

const emit = defineEmits<{
  (e: 'send', mode: 'all' | 'selected'): void
}>()

const { t, locale } = useI18n()

// Local state
const sendMode = ref<'all' | 'selected'>('all')

// Computed
const canSend = computed(() => {
  if (props.sending) return false
  if (sendMode.value === 'selected' && props.selectedEmailsCount === 0) return false
  return true
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

function handleSend() {
  emit('send', sendMode.value)
}
</script>

<template>
  <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700">
    <div class="p-6 border-b border-slate-100 dark:border-slate-700">
      <h2 class="font-semibold text-slate-900 dark:text-slate-100 flex items-center gap-2">
        <Mail :size="18" class="text-blue-600 dark:text-blue-400" />
        {{ t('admin.documentDetail.reminders') }}
      </h2>
      <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
        {{ t('admin.documentDetail.remindersDescription') }}
      </p>
    </div>
    <div class="p-6 space-y-6">
      <div class="grid gap-4 grid-cols-1 sm:grid-cols-3">
        <div class="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4">
          <p class="text-sm text-slate-500 dark:text-slate-400">
            {{ t('admin.documentDetail.remindersSent') }}
          </p>
          <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">
            {{ reminderStats.totalSent }}
          </p>
        </div>
        <div class="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4">
          <p class="text-sm text-slate-500 dark:text-slate-400">
            {{ t('admin.documentDetail.toRemind') }}
          </p>
          <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">
            {{ reminderStats.pendingCount }}
          </p>
        </div>
        <div v-if="reminderStats.lastSentAt" class="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-4">
          <p class="text-sm text-slate-500 dark:text-slate-400">
            {{ t('admin.documentDetail.lastReminder') }}
          </p>
          <p class="text-sm font-bold text-slate-900 dark:text-slate-100">
            {{ formatDate(reminderStats.lastSentAt) }}
          </p>
        </div>
      </div>

      <div
        v-if="!smtpEnabled"
        class="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-xl p-4"
      >
        <p class="text-sm text-amber-800 dark:text-amber-200">
          {{ t('admin.documentDetail.emailServiceDisabled') }}
        </p>
      </div>

      <div v-if="smtpEnabled" class="space-y-4">
        <div class="space-y-2">
          <label class="flex items-center space-x-2 cursor-pointer">
            <input
              type="radio"
              v-model="sendMode"
              value="all"
              class="text-blue-600 focus:ring-blue-500"
            />
            <span class="text-sm text-slate-700 dark:text-slate-300">
              {{ t('admin.documentDetail.sendToAll', { count: reminderStats.pendingCount }) }}
            </span>
          </label>
          <label class="flex items-center space-x-2 cursor-pointer">
            <input
              type="radio"
              v-model="sendMode"
              value="selected"
              class="text-blue-600 focus:ring-blue-500"
            />
            <span class="text-sm text-slate-700 dark:text-slate-300">
              {{ t('admin.documentDetail.sendToSelected', { count: selectedEmailsCount }) }}
            </span>
          </label>
        </div>
        <button
          @click="handleSend"
          :disabled="!canSend"
          class="trust-gradient text-white font-medium rounded-lg px-4 py-2.5 text-sm hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ sending ? t('admin.documentDetail.sending') : t('admin.documentDetail.sendReminders') }}
        </button>
      </div>
    </div>
  </div>
</template>
