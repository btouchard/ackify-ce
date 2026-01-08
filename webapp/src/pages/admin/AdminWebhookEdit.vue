<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { availableWebhookEvents, createWebhook, getWebhook, updateWebhook, type WebhookInput, type Webhook } from '@/services/webhooks'
import { extractError } from '@/services/http'
import { Loader2, Save, ArrowLeft, Webhook as WebhookIcon, ChevronRight } from 'lucide-vue-next'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const isNew = computed(() => route.name === 'admin-webhook-new')
const id = computed(() => Number(route.params.id))

const loading = ref(false)
const saving = ref(false)
const error = ref('')

const title = ref('')
const targetUrl = ref('')
const secret = ref('')
const active = ref(true)
const events = ref<string[]>([])
const description = ref('')

async function load() {
  if (isNew.value) return
  try {
    loading.value = true
    const resp = await getWebhook(id.value)
    const wh = resp.data as Webhook
    title.value = wh.title || ''
    targetUrl.value = wh.targetUrl
    active.value = wh.active
    events.value = [...(wh.events||[])]
    description.value = wh.description || ''
  } catch (err) {
    error.value = extractError(err)
  } finally {
    loading.value = false
  }
}

function toggleEvent(key: string) {
  if (events.value.includes(key)) {
    events.value = events.value.filter(k => k !== key)
  } else {
    events.value = [...events.value, key]
  }
}

async function save() {
  error.value = ''
  if (!title.value || !targetUrl.value || (!secret.value && isNew.value) || events.value.length === 0) {
    error.value = t('admin.webhooks.form.validation')
    return
  }
  try {
    saving.value = true
    const payload: WebhookInput = {
      title: title.value.trim(),
      targetUrl: targetUrl.value.trim(),
      secret: secret.value.trim(),
      active: active.value,
      events: events.value,
      description: description.value.trim() || undefined,
    }
    if (isNew.value) {
      await createWebhook(payload)
    } else {
      // Keep existing secret when left blank during edit
      if (!payload.secret) delete (payload as any).secret
      await updateWebhook(id.value, payload)
    }
    router.push({ name: 'admin-webhooks' })
  } catch (err) {
    error.value = extractError(err)
  } finally {
    saving.value = false
  }
}

function goBack() { router.push({ name: 'admin-webhooks' }) }

onMounted(load)
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 sm:px-6 py-6 sm:py-8">
    <!-- Breadcrumb -->
    <nav class="flex items-center gap-2 text-sm mb-6">
      <router-link to="/admin" class="text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 transition-colors">
        {{ t('admin.title') }}
      </router-link>
      <ChevronRight :size="16" class="text-slate-300 dark:text-slate-600" />
      <router-link :to="{ name: 'admin-webhooks' }" class="text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 transition-colors">
        {{ t('admin.webhooks.title') }}
      </router-link>
      <ChevronRight :size="16" class="text-slate-300 dark:text-slate-600" />
      <span class="text-slate-900 dark:text-slate-100 font-medium truncate max-w-[200px]">
        {{ isNew ? t('admin.webhooks.new') : (title || t('admin.webhooks.editTitle')) }}
      </span>
    </nav>

    <!-- Page Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6 sm:mb-8">
      <div class="flex items-start gap-4">
        <div class="w-12 h-12 sm:w-14 sm:h-14 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
          <WebhookIcon class="w-6 h-6 sm:w-7 sm:h-7 text-blue-600 dark:text-blue-400" />
        </div>
        <div>
          <h1 class="text-xl sm:text-2xl font-bold text-slate-900 dark:text-white">
            {{ isNew ? t('admin.webhooks.new') : t('admin.webhooks.editTitle') }}
          </h1>
          <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">{{ t('admin.webhooks.form.subtitle') }}</p>
        </div>
      </div>
      <button
        @click="goBack"
        class="w-full sm:w-auto inline-flex items-center justify-center gap-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 font-medium rounded-lg px-4 py-2.5 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors min-h-[44px]"
      >
        <ArrowLeft :size="18" />
        {{ t('common.back') || 'Retour' }}
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
        <h2 class="font-semibold text-slate-900 dark:text-white">{{ t('admin.webhooks.form.title') }}</h2>
      </div>

      <!-- Content -->
      <div class="p-4 sm:p-6">
        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center gap-3 py-12">
          <Loader2 :size="24" class="animate-spin text-blue-600 dark:text-blue-400" />
          <span class="text-slate-500 dark:text-slate-400">{{ t('admin.loading') }}</span>
        </div>

        <!-- Form -->
        <form v-else @submit.prevent="save" class="space-y-6">
          <!-- Title -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
              {{ t('admin.webhooks.form.nameLabel') }}
            </label>
            <input
              v-model="title"
              type="text"
              required
              :placeholder="t('admin.webhooks.form.namePlaceholder')"
              class="w-full px-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-colors"
            />
          </div>

          <!-- URL -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
              {{ t('admin.webhooks.form.urlLabel') }}
            </label>
            <input
              v-model="targetUrl"
              type="url"
              required
              placeholder="https://example.com/webhook"
              class="w-full px-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-colors"
            />
          </div>

          <!-- Secret -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
              {{ t('admin.webhooks.form.secretLabel') }}
            </label>
            <input
              v-model="secret"
              :type="isNew ? 'text' : 'password'"
              :placeholder="isNew ? t('admin.webhooks.form.secretPlaceholder') : t('admin.webhooks.form.secretKeep')"
              :required="isNew"
              class="w-full px-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-colors"
            />
          </div>

          <!-- Events -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-3">
              {{ t('admin.webhooks.form.eventsLabel') }}
            </label>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <label
                v-for="e in availableWebhookEvents"
                :key="e.key"
                :class="[
                  'flex items-center gap-3 p-3 rounded-lg border cursor-pointer transition-colors',
                  events.includes(e.key)
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                    : 'border-slate-200 dark:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-700/50'
                ]"
              >
                <input
                  type="checkbox"
                  :value="e.key"
                  :checked="events.includes(e.key)"
                  @change="toggleEvent(e.key)"
                  class="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500 focus:ring-offset-0"
                />
                <span :class="events.includes(e.key) ? 'text-blue-700 dark:text-blue-400 font-medium' : 'text-slate-700 dark:text-slate-300'">
                  {{ t(e.labelKey) }}
                </span>
              </label>
            </div>
          </div>

          <!-- Description -->
          <div>
            <label class="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-2">
              {{ t('admin.webhooks.form.descriptionLabel') }}
            </label>
            <textarea
              v-model="description"
              rows="3"
              :placeholder="t('admin.webhooks.form.descriptionPlaceholder')"
              class="w-full px-4 py-2.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-colors resize-none"
            />
          </div>

          <!-- Submit -->
          <div class="pt-4 border-t border-slate-200 dark:border-slate-700">
            <button
              type="submit"
              :disabled="saving"
              class="w-full sm:w-auto inline-flex items-center justify-center gap-2 trust-gradient text-white font-medium rounded-lg px-6 py-2.5 hover:opacity-90 transition-opacity disabled:opacity-50 min-h-[44px]"
            >
              <Loader2 v-if="saving" :size="18" class="animate-spin" />
              <Save v-else :size="18" />
              {{ t('common.save') || 'Enregistrer' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
