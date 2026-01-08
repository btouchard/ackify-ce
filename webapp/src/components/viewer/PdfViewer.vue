<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import * as pdfjsLib from 'pdfjs-dist'
import { Loader2, AlertCircle, ChevronLeft, ChevronRight } from 'lucide-vue-next'

// Configure PDF.js worker
pdfjsLib.GlobalWorkerOptions.workerSrc = new URL(
  'pdfjs-dist/build/pdf.worker.min.mjs',
  import.meta.url
).toString()

const props = defineProps<{
  url: string
  scale?: number
}>()

const emit = defineEmits<{
  scroll: [progress: number, currentPage: number, totalPages: number]
  loaded: [totalPages: number]
  error: [message: string]
}>()

const { t } = useI18n()

// State
const containerRef = ref<HTMLElement | null>(null)
const pagesContainerRef = ref<HTMLElement | null>(null)
const isLoading = ref(true)
const errorMessage = ref<string | null>(null)
const totalPages = ref(0)
const currentPage = ref(1)

// PDF document reference
let pdfDoc: pdfjsLib.PDFDocumentProxy | null = null
let renderTasks: Map<number, pdfjsLib.RenderTask> = new Map()

// Flag to track if we need to render after refs are ready
let pendingRender = false

// Load PDF document
async function loadPdf() {
  if (!props.url) return

  isLoading.value = true
  errorMessage.value = null
  pendingRender = false

  try {
    // Cancel any existing render tasks
    for (const task of renderTasks.values()) {
      task.cancel()
    }
    renderTasks.clear()

    // Destroy previous document
    if (pdfDoc) {
      pdfDoc.destroy()
      pdfDoc = null
    }

    // Load the PDF
    const loadingTask = pdfjsLib.getDocument({
      url: props.url,
      withCredentials: true,
    })

    pdfDoc = await loadingTask.promise
    totalPages.value = pdfDoc.numPages
    currentPage.value = 1

    emit('loaded', pdfDoc.numPages)

    isLoading.value = false

    // Wait for DOM update then render
    await nextTick()

    // Check if refs are ready, if not mark as pending
    if (pagesContainerRef.value) {
      await renderAllPages()
      await nextTick()
      handleScroll()
    } else {
      pendingRender = true
    }
  } catch (err) {
    console.error('PDF load error:', err)
    const message = err instanceof Error ? err.message : t('documentViewer.error.loadFailed')
    errorMessage.value = message
    emit('error', message)
    isLoading.value = false
  }
}

// Render all pages
async function renderAllPages() {
  if (!pdfDoc || !pagesContainerRef.value) {
    console.warn('Cannot render: pdfDoc or container not ready')
    return
  }

  const container = pagesContainerRef.value
  const scale = (props.scale || 100) / 100 * window.devicePixelRatio

  // Clear existing canvases
  container.innerHTML = ''

  // Render each page
  for (let pageNum = 1; pageNum <= pdfDoc.numPages; pageNum++) {
    try {
      const page = await pdfDoc.getPage(pageNum)
      const viewport = page.getViewport({ scale })

      // Create page wrapper
      const pageWrapper = document.createElement('div')
      pageWrapper.className = 'pdf-page-wrapper'
      pageWrapper.dataset.page = String(pageNum)
      pageWrapper.style.cssText = `
        display: flex;
        justify-content: center;
        margin-bottom: 16px;
      `

      // Create canvas
      const canvas = document.createElement('canvas')
      const context = canvas.getContext('2d')!

      canvas.width = viewport.width
      canvas.height = viewport.height
      canvas.style.cssText = `
        display: block;
        width: ${viewport.width / window.devicePixelRatio}px;
        height: ${viewport.height / window.devicePixelRatio}px;
        box-shadow: 0 2px 8px rgba(0,0,0,0.15);
        background: white;
      `

      pageWrapper.appendChild(canvas)
      container.appendChild(pageWrapper)

      // Render page to canvas
      const renderTask = page.render({
        canvasContext: context,
        viewport: viewport,
      })

      renderTasks.set(pageNum, renderTask)

      await renderTask.promise
      renderTasks.delete(pageNum)
    } catch (err) {
      // Ignore cancelled render tasks
      if (err instanceof Error && err.message.includes('cancelled')) continue
      console.error(`Error rendering page ${pageNum}:`, err)
    }
  }
}

// Handle scroll to track progress and current page
function handleScroll() {
  if (!containerRef.value || totalPages.value === 0) return

  const container = containerRef.value
  const scrollTop = container.scrollTop
  const scrollHeight = container.scrollHeight
  const clientHeight = container.clientHeight

  // Calculate progress
  const maxScroll = scrollHeight - clientHeight
  let progress = 0
  if (maxScroll <= 0) {
    progress = 100
  } else {
    progress = Math.min(100, Math.round((scrollTop / maxScroll) * 100))
  }

  // Determine current page based on scroll position
  const pageWrappers = container.querySelectorAll('.pdf-page-wrapper')
  let visiblePage = 1
  const containerTop = container.getBoundingClientRect().top

  for (let i = 0; i < pageWrappers.length; i++) {
    const pageEl = pageWrappers[i] as HTMLElement
    const rect = pageEl.getBoundingClientRect()

    // Page is current if its center is in the viewport
    const pageCenter = rect.top + rect.height / 2
    if (pageCenter > containerTop) {
      visiblePage = i + 1
      break
    }
  }

  currentPage.value = visiblePage
  emit('scroll', progress, visiblePage, totalPages.value)
}

// Navigate to specific page
function goToPage(pageNum: number) {
  if (!containerRef.value || pageNum < 1 || pageNum > totalPages.value) return

  const pageElement = containerRef.value.querySelector(`[data-page="${pageNum}"]`)
  if (pageElement) {
    pageElement.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

// Watch for container ref becoming available (for pending renders)
watch(pagesContainerRef, async (newRef) => {
  if (newRef && pendingRender && pdfDoc) {
    pendingRender = false
    await renderAllPages()
    await nextTick()
    handleScroll()
  }
})

// Watch for URL changes
watch(() => props.url, (newUrl, oldUrl) => {
  if (newUrl && newUrl !== oldUrl) {
    loadPdf()
  }
})

// Watch for scale changes - re-render pages
watch(() => props.scale, async (newScale, oldScale) => {
  if (pdfDoc && newScale !== oldScale) {
    await renderAllPages()
    await nextTick()
    handleScroll()
  }
})

onMounted(() => {
  if (props.url) {
    loadPdf()
  }
})

onUnmounted(() => {
  // Cancel all render tasks
  for (const task of renderTasks.values()) {
    task.cancel()
  }
  renderTasks.clear()

  // Destroy PDF document
  if (pdfDoc) {
    pdfDoc.destroy()
    pdfDoc = null
  }
})

// Expose methods for parent component
defineExpose({
  goToPage,
  currentPage,
  totalPages,
})
</script>

<template>
  <div class="pdf-viewer-wrapper h-full flex flex-col">
    <!-- Loading state -->
    <div v-if="isLoading" class="flex-1 flex items-center justify-center bg-slate-100 dark:bg-slate-800">
      <div class="text-center">
        <Loader2 class="w-8 h-8 text-blue-600 animate-spin mx-auto mb-3" />
        <p class="text-sm text-slate-500 dark:text-slate-400">{{ t('documentViewer.loading') }}</p>
      </div>
    </div>

    <!-- Error state -->
    <div v-else-if="errorMessage" class="flex-1 flex items-center justify-center p-8 bg-slate-100 dark:bg-slate-800">
      <div class="text-center max-w-sm">
        <div class="w-14 h-14 rounded-xl bg-red-50 dark:bg-red-900/20 flex items-center justify-center mx-auto mb-4">
          <AlertCircle class="w-7 h-7 text-red-600 dark:text-red-400" />
        </div>
        <h3 class="font-semibold text-slate-900 dark:text-slate-100 mb-2">{{ t('documentViewer.error.title') }}</h3>
        <p class="text-sm text-slate-500 dark:text-slate-400">{{ errorMessage }}</p>
      </div>
    </div>

    <!-- PDF pages container -->
    <div
      v-else
      ref="containerRef"
      class="flex-1 overflow-auto bg-slate-200 dark:bg-slate-700"
      @scroll="handleScroll"
    >
      <div ref="pagesContainerRef" class="py-4">
        <!-- Pages will be rendered here dynamically -->
      </div>
    </div>

    <!-- Page navigation -->
    <div v-if="!isLoading && !errorMessage && totalPages > 1" class="flex items-center justify-center gap-2 py-2 border-t border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800">
      <button
        @click="goToPage(currentPage - 1)"
        :disabled="currentPage <= 1"
        class="p-1.5 rounded hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <ChevronLeft class="w-4 h-4 text-slate-600 dark:text-slate-400" />
      </button>
      <span class="text-sm text-slate-600 dark:text-slate-400 font-medium min-w-[80px] text-center">
        {{ currentPage }} / {{ totalPages }}
      </span>
      <button
        @click="goToPage(currentPage + 1)"
        :disabled="currentPage >= totalPages"
        class="p-1.5 rounded hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <ChevronRight class="w-4 h-4 text-slate-600 dark:text-slate-400" />
      </button>
    </div>
  </div>
</template>
