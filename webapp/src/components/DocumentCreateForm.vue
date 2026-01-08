<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { documentService, type FindOrCreateDocumentResponse, type UploadProgress } from '@/services/documents'
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
  ShieldCheck,
  Upload
} from 'lucide-vue-next'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import Textarea from '@/components/ui/Textarea.vue'

export interface DocumentCreateFormProps {
  mode?: 'compact' | 'full' | 'hero'
  redirectOnCreate?: boolean
  showUploadButton?: boolean
  redirectRoute?: 'home' | 'document-edit'
}

const props = withDefaults(defineProps<DocumentCreateFormProps>(), {
  mode: 'compact',
  redirectOnCreate: true,
  showUploadButton: true,
  redirectRoute: 'home'
})

const emit = defineEmits<{
  created: [document: FindOrCreateDocumentResponse]
}>()

const { t } = useI18n()
const router = useRouter()

const storageEnabled = (window as any).ACKIFY_STORAGE_ENABLED || false

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

// Upload state
const fileInputRef = ref<HTMLInputElement | null>(null)
const selectedFile = ref<File | null>(null)
const uploadProgress = ref<UploadProgress | null>(null)
const isUploading = ref(false)

// UI state
const optionsExpanded = ref(false)
const isSubmitting = ref(false)
const errorMessage = ref<string | null>(null)

// Computed
const isValid = computed(() => url.value.trim().length > 0 || selectedFile.value !== null)
const showOptions = computed(() => props.mode === 'full' || (props.mode !== 'hero' && optionsExpanded.value))
const isHeroMode = computed(() => props.mode === 'hero')
const canShowUpload = computed(() => storageEnabled && props.showUploadButton)

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
  // Reset upload state
  selectedFile.value = null
  uploadProgress.value = null
  isUploading.value = false
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

// Handle file selection
function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) {
    selectedFile.value = file
    // Clear URL when file is selected
    url.value = ''
    // Auto-set title from filename if not set
    if (!title.value) {
      title.value = file.name
    }
  }
}

// Trigger file input click
function triggerFileInput() {
  fileInputRef.value?.click()
}

// Clear selected file
function clearSelectedFile() {
  selectedFile.value = null
  uploadProgress.value = null
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

// Format file size
function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

// Submit form
async function handleSubmit() {
  if (!isValid.value || isSubmitting.value) return

  errorMessage.value = null
  isSubmitting.value = true

  try {
    let docId: string

    if (selectedFile.value) {
      // Upload file
      isUploading.value = true
      const uploadResponse = await documentService.uploadDocument(
        selectedFile.value,
        title.value || undefined,
        (progress) => {
          uploadProgress.value = progress
        }
      )
      docId = uploadResponse.doc_id
      isUploading.value = false
    } else {
      // Find or create document by reference
      const response = await documentService.findOrCreateDocument(url.value.trim())
      docId = response.docId
    }

    if (props.redirectOnCreate) {
      // Redirect based on redirectRoute prop
      if (props.redirectRoute === 'document-edit') {
        await router.push({ name: 'document-edit', params: { id: docId } })
      } else {
        await router.push(`/?doc=${docId}`)
      }
    } else {
      // Emit event and reset form
      const response = await documentService.findOrCreateDocument(docId)
      emit('created', response)
      resetForm()
    }
  } catch (error) {
    errorMessage.value = extractError(error)
    isUploading.value = false
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
      data-testid="error-message"
      class="mb-4 rounded-lg bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-900 p-4 text-sm text-red-800 dark:text-red-200"
    >
      {{ errorMessage }}
    </div>

    <!-- Hidden file input -->
    <input
      ref="fileInputRef"
      type="file"
      class="hidden"
      accept=".pdf,.doc,.docx,.txt,.html,.htm,.png,.jpg,.jpeg,.gif,.webp"
      @change="handleFileSelect"
    />

    <!-- Main form -->
    <div class="space-y-4">
      <!-- URL input + Upload button + Submit button -->
      <div class="flex gap-3" :class="mode === 'full' ? 'flex-col sm:flex-row' : ''">
        <div class="flex-1">
          <Label v-if="mode === 'full'" for="doc-url" class="mb-1.5">
            {{ t('documentCreateForm.url.label') }}
          </Label>

          <!-- Show selected file or URL input -->
          <div v-if="selectedFile" class="flex items-center gap-2 px-4 py-2.5 rounded-lg border border-blue-300 dark:border-blue-700 bg-blue-50 dark:bg-blue-950/30" :class="mode === 'full' ? 'h-11' : 'h-12'">
            <FileText class="w-4 h-4 text-blue-600 dark:text-blue-400 flex-shrink-0" />
            <span data-testid="selected-file-name" class="flex-1 text-sm text-blue-800 dark:text-blue-200 truncate">{{ selectedFile.name }}</span>
            <span class="text-xs text-blue-600 dark:text-blue-400">{{ formatFileSize(selectedFile.size) }}</span>
            <button
              type="button"
              data-testid="clear-file-button"
              @click="clearSelectedFile"
              class="p-1 rounded hover:bg-blue-200 dark:hover:bg-blue-800 text-blue-600 dark:text-blue-400"
              :disabled="isSubmitting"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div v-else class="flex gap-2">
            <Input
              id="doc-url"
              data-testid="doc-url-input"
              v-model="url"
              type="text"
              :placeholder="canShowUpload ? t('documentCreateForm.url.placeholderWithUpload') : t('documentCreateForm.url.placeholder')"
              class="flex-1"
              :class="[
                mode === 'full' ? 'h-11' : 'h-12',
                isHeroMode ? 'text-sm' : ''
              ]"
              :disabled="isSubmitting"
              @keyup.enter="handleSubmit"
            />
            <!-- Upload button (only when storage is enabled AND showUploadButton is true) -->
            <Button
              v-if="canShowUpload"
              type="button"
              data-testid="upload-button"
              variant="outline"
              @click="triggerFileInput"
              :class="mode === 'full' ? 'h-11' : 'h-12'"
              :disabled="isSubmitting"
              :title="t('documentCreateForm.upload.button')"
            >
              <Upload class="w-4 h-4" />
            </Button>
          </div>

          <p v-if="mode === 'full'" class="mt-1.5 text-xs text-slate-400 dark:text-slate-500">
            {{ storageEnabled ? t('documentCreateForm.url.helperWithUpload') : t('documentCreateForm.url.helper') }}
          </p>
        </div>
        <div :class="mode === 'full' ? 'flex items-end' : ''">
          <Button
            data-testid="submit-button"
            @click="handleSubmit"
            :size="mode === 'full' ? 'default' : 'lg'"
            class="group whitespace-nowrap"
            :class="mode === 'full' ? 'h-11' : 'h-12'"
            :disabled="!isValid || isSubmitting"
          >
            <Loader2 v-if="isSubmitting" class="w-4 h-4 animate-spin" />
            <template v-else>
              <Upload v-if="selectedFile" class="w-4 h-4" />
              <FileText v-else class="w-4 h-4" />
              <span class="ml-2">{{ selectedFile ? t('documentCreateForm.upload.submit') : t('documentCreateForm.submit') }}</span>
              <ArrowRight :size="16" class="ml-2 transition-transform group-hover:translate-x-1" />
            </template>
          </Button>
        </div>
      </div>

      <!-- Upload progress bar -->
      <div v-if="isUploading && uploadProgress" class="w-full">
        <div class="flex justify-between text-xs text-slate-500 dark:text-slate-400 mb-1">
          <span>{{ t('documentCreateForm.upload.uploading') }}</span>
          <span>{{ uploadProgress.percent }}%</span>
        </div>
        <div class="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden">
          <div
            class="h-full bg-blue-600 dark:bg-blue-500 transition-all duration-300"
            :style="{ width: `${uploadProgress.percent}%` }"
          ></div>
        </div>
      </div>

      <!-- Options accordion (compact mode) -->
      <div v-if="mode === 'compact'" class="border border-slate-200 dark:border-slate-700 rounded-lg overflow-hidden">
        <button
          type="button"
          data-testid="options-toggle"
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
