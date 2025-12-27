<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { documentService, type FindOrCreateDocumentResponse } from '@/services/documents'
import { extractError } from '@/services/http'
import {
  ArrowRight,
  ChevronDown,
  ChevronUp,
  Loader2,
  FileText,
  ExternalLink,
  Eye,
  Download,
  ScrollText,
  Hash,
  ShieldCheck
} from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import Textarea from '@/components/ui/Textarea.vue'

export interface DocumentCreateFormProps {
  mode?: 'compact' | 'full'
  redirectOnCreate?: boolean
}

const props = withDefaults(defineProps<DocumentCreateFormProps>(), {
  mode: 'compact',
  redirectOnCreate: true
})

const emit = defineEmits<{
  created: [document: FindOrCreateDocumentResponse]
}>()

const { t } = useI18n()
const router = useRouter()

// Form state
const url = ref('')
const title = ref('')
const readMode = ref<'integrated' | 'external'>('integrated')
const allowDownload = ref(true)
const requireFullRead = ref(false)
const checksum = ref('')
const checksumAlgorithm = ref<'SHA-256' | 'SHA-512'>('SHA-256')
const verifyChecksum = ref(true)
const description = ref('')

// UI state
const optionsExpanded = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref<string | null>(null)

// Computed
const isValid = computed(() => url.value.trim().length > 0)
const showOptions = computed(() => props.mode === 'full' || optionsExpanded.value)

// Toggle options accordion (compact mode only)
function toggleOptions() {
  if (props.mode === 'compact') {
    optionsExpanded.value = !optionsExpanded.value
  }
}

// Reset form
function resetForm() {
  url.value = ''
  title.value = ''
  readMode.value = 'integrated'
  allowDownload.value = true
  requireFullRead.value = false
  checksum.value = ''
  checksumAlgorithm.value = 'SHA-256'
  verifyChecksum.value = true
  description.value = ''
  optionsExpanded.value = false
  errorMessage.value = null
}

// Submit form
async function handleSubmit() {
  if (!isValid.value || isSubmitting.value) return

  errorMessage.value = null
  isSubmitting.value = true

  try {
    // Call the API to find or create document
    // Note: The current API only accepts reference. Extended options would need backend updates.
    const response = await documentService.findOrCreateDocument(url.value.trim())

    if (props.redirectOnCreate) {
      // Redirect to document page
      await router.push(`/?doc=${response.docId}`)
    } else {
      // Emit event and reset form
      emit('created', response)
      resetForm()
    }
  } catch (error) {
    errorMessage.value = extractError(error)
  } finally {
    isSubmitting.value = false
  }
}

// Watch read mode to reset related options
watch(readMode, (newMode) => {
  if (newMode === 'external') {
    allowDownload.value = false
    requireFullRead.value = false
  }
})
</script>

<template>
  <div class="document-create-form">
    <!-- Error message -->
    <div
      v-if="errorMessage"
      class="mb-4 rounded-lg bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-900 p-4 text-sm text-red-800 dark:text-red-200"
    >
      {{ errorMessage }}
    </div>

    <!-- Main form -->
    <div class="space-y-4">
      <!-- URL input + Submit button -->
      <div class="flex gap-3" :class="mode === 'full' ? 'flex-col sm:flex-row' : ''">
        <div class="flex-1">
          <Label v-if="mode === 'full'" for="doc-url" class="mb-1.5">
            {{ t('documentCreateForm.url.label') }}
          </Label>
          <Input
            id="doc-url"
            v-model="url"
            type="text"
            :placeholder="t('documentCreateForm.url.placeholder')"
            class="w-full"
            :class="mode === 'full' ? 'h-11' : 'h-12'"
            :disabled="isSubmitting"
            @keyup.enter="handleSubmit"
          />
          <p v-if="mode === 'full'" class="mt-1.5 text-xs text-slate-400 dark:text-slate-500">
            {{ t('documentCreateForm.url.helper') }}
          </p>
        </div>
        <div :class="mode === 'full' ? 'flex items-end' : ''">
          <Button
            @click="handleSubmit"
            :size="mode === 'full' ? 'default' : 'lg'"
            class="group whitespace-nowrap"
            :class="mode === 'full' ? 'h-11' : 'h-12'"
            :disabled="!isValid || isSubmitting"
          >
            <Loader2 v-if="isSubmitting" class="w-4 h-4 animate-spin" />
            <template v-else>
              <FileText class="w-4 h-4" />
              <span class="ml-2">{{ t('documentCreateForm.submit') }}</span>
              <ArrowRight :size="16" class="ml-2 transition-transform group-hover:translate-x-1" />
            </template>
          </Button>
        </div>
      </div>

      <!-- Options accordion (compact mode) -->
      <div v-if="mode === 'compact'" class="border border-slate-200 dark:border-slate-700 rounded-lg overflow-hidden">
        <button
          type="button"
          @click="toggleOptions"
          class="w-full flex items-center justify-between px-4 py-3 text-sm font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors"
        >
          <span>{{ t('documentCreateForm.options.title') }}</span>
          <ChevronUp v-if="optionsExpanded" class="w-4 h-4" />
          <ChevronDown v-else class="w-4 h-4" />
        </button>
      </div>

      <!-- Options content -->
      <div
        v-if="showOptions"
        class="space-y-4"
        :class="mode === 'compact' ? 'border border-slate-200 dark:border-slate-700 rounded-lg p-4 -mt-2 border-t-0 rounded-t-none' : ''"
      >
        <!-- Title -->
        <div>
          <Label for="doc-title">{{ t('documentCreateForm.title.label') }}</Label>
          <Input
            id="doc-title"
            v-model="title"
            type="text"
            :placeholder="t('documentCreateForm.title.placeholder')"
            class="mt-1.5"
            :disabled="isSubmitting"
          />
        </div>

        <!-- Read mode -->
        <div>
          <Label>{{ t('documentCreateForm.readMode.label') }}</Label>
          <div class="mt-2 flex gap-4">
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                v-model="readMode"
                value="integrated"
                class="w-4 h-4 text-blue-600 border-slate-300 focus:ring-blue-500"
                :disabled="isSubmitting"
              />
              <Eye class="w-4 h-4 text-slate-500" />
              <span class="text-sm text-slate-700 dark:text-slate-300">
                {{ t('documentCreateForm.readMode.integrated') }}
              </span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                v-model="readMode"
                value="external"
                class="w-4 h-4 text-blue-600 border-slate-300 focus:ring-blue-500"
                :disabled="isSubmitting"
              />
              <ExternalLink class="w-4 h-4 text-slate-500" />
              <span class="text-sm text-slate-700 dark:text-slate-300">
                {{ t('documentCreateForm.readMode.external') }}
              </span>
            </label>
          </div>
        </div>

        <!-- Integrated mode options -->
        <div v-if="readMode === 'integrated'" class="pl-4 border-l-2 border-blue-200 dark:border-blue-800 space-y-3">
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              v-model="allowDownload"
              class="w-4 h-4 text-blue-600 border-slate-300 rounded focus:ring-blue-500"
              :disabled="isSubmitting"
            />
            <Download class="w-4 h-4 text-slate-500" />
            <span class="text-sm text-slate-700 dark:text-slate-300">
              {{ t('documentCreateForm.options.allowDownload') }}
            </span>
          </label>
          <label class="flex items-center gap-2 cursor-pointer">
            <input
              type="checkbox"
              v-model="requireFullRead"
              class="w-4 h-4 text-blue-600 border-slate-300 rounded focus:ring-blue-500"
              :disabled="isSubmitting"
            />
            <ScrollText class="w-4 h-4 text-slate-500" />
            <span class="text-sm text-slate-700 dark:text-slate-300">
              {{ t('documentCreateForm.options.requireFullRead') }}
            </span>
          </label>
        </div>

        <!-- Checksum -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <Label for="doc-checksum">{{ t('documentCreateForm.checksum.label') }}</Label>
            <div class="relative mt-1.5">
              <Hash class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
              <Input
                id="doc-checksum"
                v-model="checksum"
                type="text"
                :placeholder="t('documentCreateForm.checksum.placeholder')"
                class="pl-10 font-mono text-sm"
                :disabled="isSubmitting"
              />
            </div>
          </div>
          <div>
            <Label for="doc-algorithm">{{ t('documentCreateForm.algorithm.label') }}</Label>
            <select
              id="doc-algorithm"
              v-model="checksumAlgorithm"
              class="mt-1.5 w-full px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              :disabled="isSubmitting"
            >
              <option value="SHA-256">SHA-256</option>
              <option value="SHA-512">SHA-512</option>
            </select>
          </div>
        </div>

        <!-- Verify checksum -->
        <label class="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            v-model="verifyChecksum"
            class="w-4 h-4 text-blue-600 border-slate-300 rounded focus:ring-blue-500"
            :disabled="isSubmitting"
          />
          <ShieldCheck class="w-4 h-4 text-slate-500" />
          <span class="text-sm text-slate-700 dark:text-slate-300">
            {{ t('documentCreateForm.options.verifyChecksum') }}
          </span>
        </label>

        <!-- Description (full mode only) -->
        <div v-if="mode === 'full'">
          <Label for="doc-description">{{ t('documentCreateForm.description.label') }}</Label>
          <Textarea
            id="doc-description"
            v-model="description"
            :placeholder="t('documentCreateForm.description.placeholder')"
            class="mt-1.5"
            :rows="3"
            :disabled="isSubmitting"
          />
        </div>
      </div>
    </div>
  </div>
</template>
