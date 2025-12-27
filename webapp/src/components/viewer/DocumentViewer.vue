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
  Loader2
} from 'lucide-vue-next'

const props = defineProps<{
  documentId: string
  url: string
  allowDownload?: boolean
  requireFullRead?: boolean
}>()

const emit = defineEmits<{
  readComplete: []
}>()

const { t } = useI18n()

// Document proxy composable
const documentIdRef = computed(() => props.documentId)
const urlRef = computed(() => props.url)
const { proxyUrl, contentType, isLoading, error, retry } = useDocumentProxy(documentIdRef, urlRef)

// Viewer state
const zoom = ref(100)
const currentPage = ref(1)
const totalPages = ref(1)
const readProgress = ref(0)
const hasCompletedRead = ref(false)
const contentContainer = ref<HTMLElement | null>(null)
const rawContent = ref<string>('')
const sanitizedContent = ref<string>('')
const contentLoading = ref(false)
const contentError = ref<string | null>(null)

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

// Scroll progress tracking
function handleScroll(event: Event) {
  if (!props.requireFullRead || hasCompletedRead.value) return

  const target = event.target as HTMLElement
  const scrollTop = target.scrollTop
  const scrollHeight = target.scrollHeight
  const clientHeight = target.clientHeight

  // Calculate progress (0-100)
  const maxScroll = scrollHeight - clientHeight
  if (maxScroll <= 0) {
    // Content fits without scrolling
    readProgress.value = 100
    if (!hasCompletedRead.value) {
      hasCompletedRead.value = true
      emit('readComplete')
    }
    return
  }

  const progress = Math.min(100, Math.round((scrollTop / maxScroll) * 100))
  readProgress.value = progress

  // Emit readComplete when reaching the bottom
  if (progress >= 100 && !hasCompletedRead.value) {
    hasCompletedRead.value = true
    emit('readComplete')
  }
}

// Download document
function downloadDocument() {
  if (!proxyUrl.value) return
  window.open(proxyUrl.value, '_blank')
}

// Open in new tab
function openInNewTab() {
  if (!props.url) return
  window.open(props.url, '_blank', 'noopener,noreferrer')
}

// Check if content fits without scrolling (for auto-completing read)
function checkContentFits() {
  if (!props.requireFullRead || hasCompletedRead.value || !contentContainer.value) return

  const el = contentContainer.value
  if (el.scrollHeight <= el.clientHeight) {
    readProgress.value = 100
    hasCompletedRead.value = true
    emit('readComplete')
  }
}

onMounted(() => {
  // Small delay to ensure content is rendered
  setTimeout(checkContentFits, 500)
})

// Viewer component based on content type
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

const showPageIndicator = computed(() => viewerType.value === 'pdf')
const showZoomControls = computed(() => ['pdf', 'image', 'svg'].includes(viewerType.value))
</script>

<template>
  <div class="document-viewer bg-white rounded-xl border border-slate-200 overflow-hidden">
    <!-- Toolbar -->
    <div class="flex items-center justify-between px-4 py-2 border-b border-slate-100 bg-slate-50/50">
      <div class="flex items-center gap-2">
        <!-- Zoom controls -->
        <template v-if="showZoomControls">
          <button
            @click="zoomOut"
            :disabled="zoom <= minZoom"
            class="p-1.5 rounded hover:bg-slate-200/50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            :title="t('documentViewer.toolbar.zoomOut')"
          >
            <ZoomOut class="w-4 h-4 text-slate-500" />
          </button>
          <span class="text-xs text-slate-400 font-mono px-2 min-w-[40px] text-center">{{ zoom }}%</span>
          <button
            @click="zoomIn"
            :disabled="zoom >= maxZoom"
            class="p-1.5 rounded hover:bg-slate-200/50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            :title="t('documentViewer.toolbar.zoomIn')"
          >
            <ZoomIn class="w-4 h-4 text-slate-500" />
          </button>
          <div class="w-px h-4 bg-slate-200 mx-2" />
        </template>

        <!-- Download button -->
        <button
          v-if="allowDownload"
          @click="downloadDocument"
          class="p-1.5 rounded hover:bg-slate-200/50 transition-colors"
          :title="t('documentViewer.toolbar.download')"
        >
          <Download class="w-4 h-4 text-slate-500" />
        </button>

        <!-- Open in new tab -->
        <button
          @click="openInNewTab"
          class="p-1.5 rounded hover:bg-slate-200/50 transition-colors"
          :title="t('documentViewer.toolbar.openInNewTab')"
        >
          <ExternalLink class="w-4 h-4 text-slate-500" />
        </button>
      </div>

      <!-- Right side: Page indicator or progress -->
      <div class="flex items-center gap-3">
        <span v-if="showPageIndicator" class="text-xs text-slate-400 font-mono">
          {{ t('documentViewer.toolbar.page', { current: currentPage, total: totalPages }) }}
        </span>
        <span v-if="requireFullRead" class="text-xs text-slate-500 font-medium">
          {{ t('documentViewer.progress', { percent: readProgress }) }}
        </span>
      </div>
    </div>

    <!-- Content area -->
    <div
      ref="contentContainer"
      class="document-content relative grid-bg overflow-auto"
      :style="{ height: '500px' }"
      @scroll="handleScroll"
    >
      <!-- Loading state -->
      <div v-if="isLoading || contentLoading" class="flex items-center justify-center h-full">
        <div class="text-center">
          <Loader2 class="w-8 h-8 text-blue-600 animate-spin mx-auto mb-3" />
          <p class="text-sm text-slate-500">{{ t('documentViewer.loading') }}</p>
        </div>
      </div>

      <!-- Error state -->
      <div v-else-if="error || contentError" class="flex items-center justify-center h-full p-8">
        <div class="text-center max-w-sm">
          <div class="w-14 h-14 rounded-xl bg-red-50 flex items-center justify-center mx-auto mb-4">
            <AlertCircle class="w-7 h-7 text-red-600" />
          </div>
          <h3 class="font-semibold text-slate-900 mb-2">{{ t('documentViewer.error.title') }}</h3>
          <p class="text-sm text-slate-500 mb-4">{{ error || contentError }}</p>
          <div class="flex items-center justify-center gap-3">
            <button
              @click="retry"
              class="px-4 py-2 text-sm font-medium text-slate-600 bg-slate-100 rounded-lg hover:bg-slate-200 transition-colors flex items-center gap-2"
            >
              <RefreshCw class="w-4 h-4" />
              {{ t('documentViewer.error.retry') }}
            </button>
            <button
              @click="openInNewTab"
              class="px-4 py-2 text-sm font-medium text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors flex items-center gap-2"
            >
              <ExternalLink class="w-4 h-4" />
              {{ t('documentViewer.error.openExternal') }}
            </button>
          </div>
        </div>
      </div>

      <!-- PDF Viewer -->
      <div v-else-if="viewerType === 'pdf'" class="h-full">
        <iframe
          :src="proxyUrl"
          class="w-full h-full border-0"
          :style="{ transform: `scale(${zoom / 100})`, transformOrigin: 'top left', width: `${10000 / zoom}%`, height: `${10000 / zoom}%` }"
          :title="t('documentViewer.pdfViewer')"
        />
      </div>

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
          class="w-full h-full border border-slate-100 rounded-lg bg-white"
          :title="t('documentViewer.htmlViewer')"
        />
      </div>

      <!-- Markdown Viewer (rendered and sanitized) -->
      <div v-else-if="viewerType === 'markdown'" class="p-8">
        <div
          class="prose prose-slate max-w-none bg-white rounded-lg shadow-sm border border-slate-100 p-10"
          v-html="sanitizedContent"
        />
      </div>

      <!-- Text Viewer -->
      <div v-else-if="viewerType === 'text'" class="p-8">
        <pre class="bg-white rounded-lg shadow-sm border border-slate-100 p-6 text-sm text-slate-700 font-mono whitespace-pre-wrap overflow-x-auto">{{ rawContent }}</pre>
      </div>

      <!-- Unsupported format -->
      <div v-else class="flex items-center justify-center h-full p-8">
        <div class="text-center max-w-sm">
          <div class="w-14 h-14 rounded-xl bg-slate-100 flex items-center justify-center mx-auto mb-4">
            <FileText class="w-7 h-7 text-slate-400" />
          </div>
          <h3 class="font-semibold text-slate-900 mb-2">{{ t('documentViewer.unsupported.title') }}</h3>
          <p class="text-sm text-slate-500 mb-4">{{ t('documentViewer.unsupported.description') }}</p>
          <button
            @click="openInNewTab"
            class="px-4 py-2 text-sm font-medium text-blue-600 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors flex items-center gap-2 mx-auto"
          >
            <ExternalLink class="w-4 h-4" />
            {{ t('documentViewer.unsupported.openExternal') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Progress bar (if requireFullRead) -->
    <div v-if="requireFullRead" class="border-t border-slate-100">
      <div class="h-1 bg-slate-100">
        <div
          class="h-full transition-all duration-300"
          :class="hasCompletedRead ? 'bg-emerald-500' : 'bg-blue-500'"
          :style="{ width: `${readProgress}%` }"
        />
      </div>
      <div class="px-4 py-2 flex items-center justify-between text-xs">
        <span class="text-slate-500">
          {{ hasCompletedRead ? t('documentViewer.readComplete') : t('documentViewer.scrollToRead') }}
        </span>
        <span :class="hasCompletedRead ? 'text-emerald-600 font-medium' : 'text-slate-400'">
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
