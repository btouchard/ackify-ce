<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import DOMPurify from 'dompurify'
import { marked } from 'marked'
import {
  useDocumentProxy,
  isPdf,
  isImage,
  isSvg,
  isHtml,
  isMarkdown,
  isText
} from '@/composables/useDocumentProxy'
import {
  ZoomIn,
  ZoomOut,
  Download,
  ExternalLink,
  RefreshCw,
  AlertCircle,
  FileText,
  Loader2,
  ShieldCheck,
  ShieldAlert
} from 'lucide-vue-next'
import PdfViewer from './PdfViewer.vue'

const props = defineProps<{
  documentId: string
  url: string
  allowDownload?: boolean
  requireFullRead?: boolean
  isStored?: boolean
  storedMimeType?: string
  alreadyRead?: boolean
  verifyChecksum?: boolean
  storedChecksum?: string
  checksumAlgorithm?: string
}>()

const emit = defineEmits<{
  readComplete: []
  checksumMismatch: [expected: string, actual: string]
  checksumVerified: []
}>()

const { t } = useI18n()

// Document proxy composable
const documentIdRef = computed(() => props.documentId)
const urlRef = computed(() => props.url)
const isStoredRef = computed(() => props.isStored ?? false)
const storedMimeTypeRef = computed(() => props.storedMimeType)
const { proxyUrl, contentType, isLoading, error, retry } = useDocumentProxy(
  documentIdRef,
  urlRef,
  { isStored: isStoredRef, storedMimeType: storedMimeTypeRef }
)

// Viewer state
const zoom = ref(100)
const currentPage = ref(1)
const totalPages = ref(1)
const readProgress = ref(props.alreadyRead ? 100 : 0)
const hasCompletedRead = ref(props.alreadyRead ?? false)
const contentContainer = ref<HTMLElement | null>(null)
const rawContent = ref<string>('')
const sanitizedContent = ref<string>('')
const contentLoading = ref(false)
const contentError = ref<string | null>(null)

// Checksum verification state
const checksumVerifying = ref(false)
const checksumValid = ref<boolean | null>(null)
const checksumError = ref<string | null>(null)
const calculatedChecksum = ref<string | null>(null)

// Zoom controls
const zoomLevels = [50, 75, 100, 125, 150, 175, 200]
const minZoom = 50
const maxZoom = 200

function zoomIn() {
  const nextLevel = zoomLevels.find(level => level > zoom.value)
  if (nextLevel) {
    zoom.value = nextLevel
  }
}

function zoomOut() {
  const prevLevel = [...zoomLevels].reverse().find(level => level < zoom.value)
  if (prevLevel) {
    zoom.value = prevLevel
  }
}

// Sanitize content with strict DOMPurify configuration for security
function sanitize(content: string): string {
  return DOMPurify.sanitize(content, {
    ALLOWED_TAGS: [
      'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
      'p', 'br', 'hr',
      'ul', 'ol', 'li',
      'table', 'thead', 'tbody', 'tr', 'th', 'td',
      'blockquote', 'pre', 'code',
      'a', 'strong', 'em', 'b', 'i', 'u', 's', 'del', 'ins',
      'span', 'div', 'section', 'article', 'header', 'footer', 'nav', 'aside', 'main',
      'img', 'figure', 'figcaption',
      'dl', 'dt', 'dd',
      'sup', 'sub', 'mark', 'small', 'abbr', 'cite', 'q',
      'svg', 'path', 'circle', 'rect', 'line', 'polyline', 'polygon', 'ellipse', 'g', 'defs', 'use', 'text', 'tspan'
    ],
    ALLOWED_ATTR: [
      'href', 'src', 'alt', 'title', 'class', 'id', 'name',
      'width', 'height', 'style',
      'target', 'rel',
      'colspan', 'rowspan',
      'viewBox', 'fill', 'stroke', 'stroke-width', 'd', 'cx', 'cy', 'r', 'x', 'y', 'rx', 'ry',
      'x1', 'y1', 'x2', 'y2', 'points', 'transform', 'xmlns'
    ],
    FORBID_TAGS: ['script', 'style', 'iframe', 'object', 'embed', 'form', 'input', 'button', 'select', 'textarea'],
    FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover', 'onmouseout', 'onfocus', 'onblur', 'onchange', 'onsubmit', 'onkeydown', 'onkeyup', 'onkeypress'],
    ALLOW_DATA_ATTR: false,
    ADD_ATTR: ['target'],
    RETURN_DOM: false,
    RETURN_DOM_FRAGMENT: false
  })
}

// Load text-based content (HTML, Markdown, Text, SVG)
async function loadTextContent() {
  if (!proxyUrl.value) return

  const type = contentType.value
  if (!isHtml(type) && !isMarkdown(type) && !isText(type) && !isSvg(type)) {
    return
  }

  try {
    contentLoading.value = true
    contentError.value = null

    const response = await fetch(proxyUrl.value, {
      credentials: 'include'
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    rawContent.value = await response.text()

    // Process content based on type
    if (isMarkdown(type)) {
      const htmlFromMarkdown = await marked(rawContent.value)
      sanitizedContent.value = sanitize(htmlFromMarkdown)
    } else if (isHtml(type)) {
      sanitizedContent.value = sanitize(rawContent.value)
    } else if (isSvg(type)) {
      sanitizedContent.value = sanitize(rawContent.value)
    } else {
      // Plain text - no sanitization needed, will be displayed in <pre>
      sanitizedContent.value = rawContent.value
    }
  } catch (err) {
    contentError.value = err instanceof Error ? err.message : t('documentViewer.error.loadFailed')
  } finally {
    contentLoading.value = false
  }
}

// Watch content type changes to load text content
watch(contentType, (newType) => {
  if (isHtml(newType) || isMarkdown(newType) || isText(newType) || isSvg(newType)) {
    loadTextContent()
  }
}, { immediate: true })

// Unified scroll progress tracking for all scrollable content types
function updateReadProgress(progress: number) {
  if (!props.requireFullRead || hasCompletedRead.value) return

  readProgress.value = progress

  if (progress >= 100 && !hasCompletedRead.value) {
    hasCompletedRead.value = true
    emit('readComplete')
  }
}

// Handle scroll for non-PDF content
function handleScroll(event: Event) {
  if (!props.requireFullRead || hasCompletedRead.value) return

  const target = event.target as HTMLElement
  const scrollTop = target.scrollTop
  const scrollHeight = target.scrollHeight
  const clientHeight = target.clientHeight

  const maxScroll = scrollHeight - clientHeight
  if (maxScroll <= 0) {
    updateReadProgress(100)
    return
  }

  const progress = Math.min(100, Math.round((scrollTop / maxScroll) * 100))
  updateReadProgress(progress)
}

// Handle PDF scroll event from PdfViewer
function handlePdfScroll(progress: number, page: number, total: number) {
  currentPage.value = page
  totalPages.value = total
  updateReadProgress(progress)
}

// Handle PDF loaded event
function handlePdfLoaded(pages: number) {
  totalPages.value = pages
}

// Handle PDF error
function handlePdfError(message: string) {
  contentError.value = message
}

// Download document
function downloadDocument() {
  if (!proxyUrl.value) return
  const downloadUrl = proxyUrl.value + (proxyUrl.value.includes('?') ? '&' : '?') + 'download=true'

  const link = document.createElement('a')
  link.href = downloadUrl
  link.download = ''
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

// Open in new tab
function openInNewTab() {
  const urlToOpen = props.isStored ? proxyUrl.value : props.url
  if (!urlToOpen) return
  window.open(urlToOpen, '_blank', 'noopener,noreferrer')
}

// Check if content fits without scrolling
function checkContentFits() {
  if (!props.requireFullRead || hasCompletedRead.value || !contentContainer.value) return
  if (viewerType.value === 'pdf') return // PDF handles its own scroll tracking

  const el = contentContainer.value
  if (el.scrollHeight <= el.clientHeight) {
    updateReadProgress(100)
  }
}

// Verify document checksum
async function verifyDocumentChecksum() {
  if (!props.verifyChecksum || !props.storedChecksum || !proxyUrl.value) {
    return
  }

  checksumVerifying.value = true
  checksumError.value = null
  checksumValid.value = null

  try {
    const response = await fetch(proxyUrl.value, {
      credentials: 'include'
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const arrayBuffer = await response.arrayBuffer()
    const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer)
    const hashArray = Array.from(new Uint8Array(hashBuffer))
    const checksum = hashArray.map(b => b.toString(16).padStart(2, '0')).join('')

    calculatedChecksum.value = checksum

    if (checksum.toLowerCase() === props.storedChecksum.toLowerCase()) {
      checksumValid.value = true
      emit('checksumVerified')
    } else {
      checksumValid.value = false
      emit('checksumMismatch', props.storedChecksum, checksum)
    }
  } catch (err) {
    checksumError.value = err instanceof Error ? err.message : t('documentViewer.error.checksumFailed')
    checksumValid.value = false
  } finally {
    checksumVerifying.value = false
  }
}

// Watch for proxyUrl to trigger checksum verification
watch(proxyUrl, (newUrl) => {
  if (newUrl && props.verifyChecksum && props.storedChecksum) {
    verifyDocumentChecksum()
  }
}, { immediate: true })

onMounted(() => {
  setTimeout(checkContentFits, 500)
})

// Viewer type based on content type
const viewerType = computed((): 'pdf' | 'image' | 'svg' | 'html' | 'markdown' | 'text' | 'unsupported' => {
  const type = contentType.value
  if (isPdf(type)) return 'pdf'
  if (isSvg(type)) return 'svg'
  if (isImage(type)) return 'image'
  if (isHtml(type)) return 'html'
  if (isMarkdown(type)) return 'markdown'
  if (isText(type)) return 'text'
  return 'unsupported'
})

const showPageIndicator = computed(() => viewerType.value === 'pdf' && totalPages.value > 0)
const showZoomControls = computed(() => ['pdf', 'image', 'svg'].includes(viewerType.value))
</script>

<template>
  <div class="document-viewer bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
    <!-- Toolbar -->
    <div class="flex items-center justify-between px-4 py-2 border-b border-slate-100 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/50">
      <div class="flex items-center gap-2">
        <!-- Zoom controls -->
        <template v-if="showZoomControls">
          <button
            @click="zoomOut"
            :disabled="zoom <= minZoom"
            class="p-1.5 rounded hover:bg-slate-200/50 dark:hover:bg-slate-700/50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            :title="t('documentViewer.toolbar.zoomOut')"
          >
            <ZoomOut class="w-4 h-4 text-slate-500 dark:text-slate-400" />
          </button>
          <span class="text-xs text-slate-400 dark:text-slate-500 font-mono px-2 min-w-[40px] text-center">{{ zoom }}%</span>
          <button
            @click="zoomIn"
            :disabled="zoom >= maxZoom"
            class="p-1.5 rounded hover:bg-slate-200/50 dark:hover:bg-slate-700/50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            :title="t('documentViewer.toolbar.zoomIn')"
          >
            <ZoomIn class="w-4 h-4 text-slate-500 dark:text-slate-400" />
          </button>
          <div class="w-px h-4 bg-slate-200 dark:bg-slate-600 mx-2" />
        </template>

        <!-- Download button -->
        <button
          v-if="allowDownload"
          @click="downloadDocument"
          class="p-1.5 rounded hover:bg-slate-200/50 dark:hover:bg-slate-700/50 transition-colors"
          :title="t('documentViewer.toolbar.download')"
        >
          <Download class="w-4 h-4 text-slate-500 dark:text-slate-400" />
        </button>

        <!-- Open in new tab -->
        <button
          @click="openInNewTab"
          class="p-1.5 rounded hover:bg-slate-200/50 dark:hover:bg-slate-700/50 transition-colors"
          :title="t('documentViewer.toolbar.openInNewTab')"
        >
          <ExternalLink class="w-4 h-4 text-slate-500 dark:text-slate-400" />
        </button>
      </div>

      <!-- Right side: Checksum status, Page indicator, progress -->
      <div class="flex items-center gap-3">
        <!-- Checksum verification status -->
        <div v-if="verifyChecksum && storedChecksum" class="flex items-center gap-1.5">
          <Loader2 v-if="checksumVerifying" class="w-4 h-4 text-slate-400 animate-spin" />
          <template v-else-if="checksumValid === true">
            <ShieldCheck class="w-4 h-4 text-emerald-500" />
            <span class="text-xs text-emerald-600 dark:text-emerald-400 font-medium">{{ t('documentViewer.checksum.valid') }}</span>
          </template>
          <template v-else-if="checksumValid === false">
            <ShieldAlert class="w-4 h-4 text-red-500" />
            <span class="text-xs text-red-600 dark:text-red-400 font-medium">{{ t('documentViewer.checksum.invalid') }}</span>
          </template>
          <div v-if="checksumValid !== null" class="w-px h-4 bg-slate-200 dark:bg-slate-600 ml-1" />
        </div>

        <span v-if="showPageIndicator" class="text-xs text-slate-400 dark:text-slate-500 font-mono">
          {{ t('documentViewer.toolbar.page', { current: currentPage, total: totalPages }) }}
        </span>
        <span v-if="requireFullRead" class="text-xs text-slate-500 dark:text-slate-400 font-medium">
          {{ t('documentViewer.progress', { percent: readProgress }) }}
        </span>
      </div>
    </div>

    <!-- Checksum mismatch warning -->
    <div v-if="checksumValid === false && !checksumVerifying" class="bg-red-50 dark:bg-red-900/20 border-b border-red-200 dark:border-red-800 px-4 py-3">
      <div class="flex items-start gap-3">
        <ShieldAlert class="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" />
        <div class="flex-1">
          <h4 class="font-medium text-red-900 dark:text-red-200 text-sm">{{ t('documentViewer.checksum.mismatchTitle') }}</h4>
          <p class="text-xs text-red-700 dark:text-red-300 mt-1">{{ t('documentViewer.checksum.mismatchDescription') }}</p>
          <div class="mt-2 text-xs font-mono">
            <div class="text-red-600 dark:text-red-400">{{ t('documentViewer.checksum.expected') }}: {{ storedChecksum?.substring(0, 16) }}...</div>
            <div class="text-red-600 dark:text-red-400">{{ t('documentViewer.checksum.actual') }}: {{ calculatedChecksum?.substring(0, 16) }}...</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Content area -->
    <div
      ref="contentContainer"
      class="document-content relative"
      :class="viewerType === 'pdf' ? 'overflow-hidden' : 'grid-bg overflow-auto'"
      :style="{ height: '500px' }"
      @scroll="viewerType !== 'pdf' ? handleScroll($event) : undefined"
    >
      <!-- Loading state -->
      <div v-if="isLoading || contentLoading" class="flex items-center justify-center h-full">
        <div class="text-center">
          <Loader2 class="w-8 h-8 text-blue-600 animate-spin mx-auto mb-3" />
          <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('documentViewer.loading') }}</p>
        </div>
      </div>

      <!-- Error state -->
      <div v-else-if="error || contentError" class="flex items-center justify-center h-full p-8">
        <div class="text-center max-w-sm">
          <div class="w-14 h-14 rounded-xl bg-red-50 dark:bg-red-900/20 flex items-center justify-center mx-auto mb-4">
            <AlertCircle class="w-7 h-7 text-red-600 dark:text-red-400" />
          </div>
          <h3 class="font-semibold text-slate-900 dark:text-slate-100 mb-2">{{ t('documentViewer.error.title') }}</h3>
          <p class="text-sm text-slate-500 dark:text-slate-400 mb-4">{{ error || contentError }}</p>
          <div class="flex items-center justify-center gap-3">
            <button
              @click="retry"
              class="px-4 py-2 text-sm font-medium text-slate-600 dark:text-slate-300 bg-slate-100 dark:bg-slate-700 rounded-lg hover:bg-slate-200 dark:hover:bg-slate-600 transition-colors flex items-center gap-2"
            >
              <RefreshCw class="w-4 h-4" />
              {{ t('documentViewer.error.retry') }}
            </button>
            <button
              @click="openInNewTab"
              class="px-4 py-2 text-sm font-medium text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/40 transition-colors flex items-center gap-2"
            >
              <ExternalLink class="w-4 h-4" />
              {{ t('documentViewer.error.openExternal') }}
            </button>
          </div>
        </div>
      </div>

      <!-- PDF Viewer (using PDF.js) -->
      <PdfViewer
        v-else-if="viewerType === 'pdf' && proxyUrl"
        :url="proxyUrl"
        :scale="zoom"
        @scroll="handlePdfScroll"
        @loaded="handlePdfLoaded"
        @error="handlePdfError"
      />

      <!-- Image Viewer -->
      <div v-else-if="viewerType === 'image'" class="flex items-center justify-center h-full p-8">
        <img
          :src="proxyUrl"
          :alt="t('documentViewer.imageAlt')"
          class="max-w-full max-h-full object-contain rounded-lg shadow-sm"
          :style="{ transform: `scale(${zoom / 100})` }"
        />
      </div>

      <!-- SVG Viewer (sanitized) -->
      <div v-else-if="viewerType === 'svg'" class="flex items-center justify-center h-full p-8">
        <div
          class="svg-container max-w-full max-h-full"
          :style="{ transform: `scale(${zoom / 100})` }"
          v-html="sanitizedContent"
        />
      </div>

      <!-- HTML Viewer (sanitized in sandboxed iframe) -->
      <div v-else-if="viewerType === 'html'" class="h-full p-8">
        <iframe
          sandbox="allow-same-origin"
          :srcdoc="sanitizedContent"
          class="w-full h-full border border-slate-100 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900"
          :title="t('documentViewer.htmlViewer')"
        />
      </div>

      <!-- Markdown Viewer (rendered and sanitized) -->
      <div v-else-if="viewerType === 'markdown'" class="p-8">
        <div
          class="prose prose-slate dark:prose-invert max-w-none bg-white dark:bg-slate-900 rounded-lg shadow-sm border border-slate-100 dark:border-slate-700 p-10"
          v-html="sanitizedContent"
        />
      </div>

      <!-- Text Viewer -->
      <div v-else-if="viewerType === 'text'" class="p-8">
        <pre class="bg-white dark:bg-slate-900 rounded-lg shadow-sm border border-slate-100 dark:border-slate-700 p-6 text-sm text-slate-700 dark:text-slate-300 font-mono whitespace-pre-wrap overflow-x-auto">{{ rawContent }}</pre>
      </div>

      <!-- Unsupported format -->
      <div v-else class="flex items-center justify-center h-full p-8">
        <div class="text-center max-w-sm">
          <div class="w-14 h-14 rounded-xl bg-slate-100 dark:bg-slate-700 flex items-center justify-center mx-auto mb-4">
            <FileText class="w-7 h-7 text-slate-400" />
          </div>
          <h3 class="font-semibold text-slate-900 dark:text-slate-100 mb-2">{{ t('documentViewer.unsupported.title') }}</h3>
          <p class="text-sm text-slate-500 dark:text-slate-400 mb-4">{{ t('documentViewer.unsupported.description') }}</p>
          <button
            @click="openInNewTab"
            class="px-4 py-2 text-sm font-medium text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/40 transition-colors flex items-center gap-2 mx-auto"
          >
            <ExternalLink class="w-4 h-4" />
            {{ t('documentViewer.unsupported.openExternal') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Progress bar (if requireFullRead) -->
    <div v-if="requireFullRead" class="border-t border-slate-100 dark:border-slate-700">
      <div class="h-1 bg-slate-100 dark:bg-slate-700">
        <div
          class="h-full transition-all duration-300"
          :class="hasCompletedRead ? 'bg-emerald-500' : 'bg-blue-500'"
          :style="{ width: `${readProgress}%` }"
        />
      </div>
      <div class="px-4 py-2 flex items-center justify-between text-xs">
        <span class="text-slate-500 dark:text-slate-400">
          {{ hasCompletedRead ? t('documentViewer.readComplete') : t('documentViewer.scrollToRead') }}
        </span>
        <span :class="hasCompletedRead ? 'text-emerald-600 dark:text-emerald-400 font-medium' : 'text-slate-400 dark:text-slate-500'">
          {{ readProgress }}%
        </span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.grid-bg {
  background-image:
    linear-gradient(rgba(0,0,0,0.02) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0,0,0,0.02) 1px, transparent 1px);
  background-size: 20px 20px;
}

.dark .grid-bg {
  background-image:
    linear-gradient(rgba(255,255,255,0.02) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255,255,255,0.02) 1px, transparent 1px);
}

.document-viewer :deep(.prose) {
  font-family: 'IBM Plex Sans', system-ui, sans-serif;
}

.document-viewer :deep(.prose pre),
.document-viewer :deep(.prose code) {
  font-family: 'IBM Plex Mono', monospace;
}

.svg-container :deep(svg) {
  max-width: 100%;
  max-height: 100%;
}
</style>
