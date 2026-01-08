// SPDX-License-Identifier: AGPL-3.0-or-later
import { ref, computed, watch, type Ref } from 'vue'

export type ContentType =
  | 'application/pdf'
  | 'image/png'
  | 'image/jpeg'
  | 'image/gif'
  | 'image/webp'
  | 'image/svg+xml'
  | 'text/html'
  | 'text/markdown'
  | 'text/plain'
  | 'unknown'

export interface DocumentProxyState {
  proxyUrl: Ref<string>
  contentType: Ref<ContentType>
  isLoading: Ref<boolean>
  error: Ref<string | null>
  retry: () => Promise<void>
}

export interface UseDocumentProxyOptions {
  // If true, use the storage content endpoint instead of proxy
  isStored?: Ref<boolean> | boolean
  // MIME type for stored documents (to avoid HEAD request)
  storedMimeType?: Ref<string | undefined> | string
}

/**
 * Composable for managing document proxy access
 * @param documentId - Document ID
 * @param url - URL of the source document (not used for stored documents)
 * @param options - Additional options for stored documents
 */
export function useDocumentProxy(
  documentId: Ref<string> | string,
  url: Ref<string> | string,
  options?: UseDocumentProxyOptions
): DocumentProxyState {
  const docId = typeof documentId === 'string' ? ref(documentId) : documentId
  const docUrl = typeof url === 'string' ? ref(url) : url
  const isStored = options?.isStored
    ? (typeof options.isStored === 'boolean' ? ref(options.isStored) : options.isStored)
    : ref(false)
  const storedMimeType = options?.storedMimeType
    ? (typeof options.storedMimeType === 'string' ? ref(options.storedMimeType) : options.storedMimeType)
    : ref<string | undefined>(undefined)

  const isLoading = ref(false)
  const error = ref<string | null>(null)
  const contentType = ref<ContentType>('unknown')

  // Build proxy URL (or content URL for stored documents)
  const proxyUrl = computed(() => {
    const baseUrl = (window as any).ACKIFY_BASE_URL || ''

    // For stored documents, use the content endpoint
    if (isStored.value && docId.value) {
      return `${baseUrl}/api/v1/documents/${docId.value}/content`
    }

    // For external documents, use the proxy endpoint
    if (!docId.value || !docUrl.value) {
      return ''
    }
    const params = new URLSearchParams({
      doc: docId.value,
      url: docUrl.value
    })
    return `${baseUrl}/api/v1/proxy?${params.toString()}`
  })

  /**
   * Detect content type from URL extension or HEAD request
   */
  async function detectContentType(): Promise<void> {
    // For stored documents with known MIME type, use it directly
    if (isStored.value && storedMimeType.value) {
      const mimeMap: Record<string, ContentType> = {
        'application/pdf': 'application/pdf',
        'image/png': 'image/png',
        'image/jpeg': 'image/jpeg',
        'image/gif': 'image/gif',
        'image/webp': 'image/webp',
        'image/svg+xml': 'image/svg+xml',
        'text/html': 'text/html',
        'text/markdown': 'text/markdown',
        'text/x-markdown': 'text/markdown',
        'text/plain': 'text/plain'
      }
      contentType.value = mimeMap[storedMimeType.value] || 'unknown'
      return
    }

    // For stored documents without MIME type, make HEAD request to content endpoint
    if (isStored.value && proxyUrl.value) {
      try {
        isLoading.value = true
        error.value = null

        const response = await fetch(proxyUrl.value, {
          method: 'HEAD',
          credentials: 'include'
        })

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`)
        }

        const mimeType = response.headers.get('Content-Type')?.split(';')[0]?.trim()
        const mimeMap: Record<string, ContentType> = {
          'application/pdf': 'application/pdf',
          'image/png': 'image/png',
          'image/jpeg': 'image/jpeg',
          'image/gif': 'image/gif',
          'image/webp': 'image/webp',
          'image/svg+xml': 'image/svg+xml',
          'text/html': 'text/html',
          'text/markdown': 'text/markdown',
          'text/x-markdown': 'text/markdown',
          'text/plain': 'text/plain'
        }
        contentType.value = mimeType ? (mimeMap[mimeType] || 'unknown') : 'unknown'
      } catch (err) {
        error.value = err instanceof Error ? err.message : 'Failed to detect content type'
        contentType.value = 'unknown'
      } finally {
        isLoading.value = false
      }
      return
    }

    if (!docUrl.value) {
      contentType.value = 'unknown'
      return
    }

    // First try to detect from URL extension
    const urlLower = docUrl.value.toLowerCase()
    const extensionMap: Record<string, ContentType> = {
      '.pdf': 'application/pdf',
      '.png': 'image/png',
      '.jpg': 'image/jpeg',
      '.jpeg': 'image/jpeg',
      '.gif': 'image/gif',
      '.webp': 'image/webp',
      '.svg': 'image/svg+xml',
      '.html': 'text/html',
      '.htm': 'text/html',
      '.md': 'text/markdown',
      '.markdown': 'text/markdown',
      '.txt': 'text/plain',
      '.text': 'text/plain'
    }

    // Check for extension match (before query params)
    const urlWithoutQuery = urlLower.split('?')[0] ?? urlLower
    for (const [ext, type] of Object.entries(extensionMap)) {
      if (urlWithoutQuery.endsWith(ext)) {
        contentType.value = type
        return
      }
    }

    // If no extension match, try HEAD request via proxy
    if (proxyUrl.value) {
      try {
        isLoading.value = true
        error.value = null

        const response = await fetch(proxyUrl.value, {
          method: 'HEAD',
          credentials: 'include'
        })

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`)
        }

        const mimeType = response.headers.get('Content-Type')?.split(';')[0]?.trim()

        if (mimeType) {
          const mimeMap: Record<string, ContentType> = {
            'application/pdf': 'application/pdf',
            'image/png': 'image/png',
            'image/jpeg': 'image/jpeg',
            'image/gif': 'image/gif',
            'image/webp': 'image/webp',
            'image/svg+xml': 'image/svg+xml',
            'text/html': 'text/html',
            'text/markdown': 'text/markdown',
            'text/x-markdown': 'text/markdown',
            'text/plain': 'text/plain'
          }
          contentType.value = mimeMap[mimeType] || 'unknown'
        } else {
          contentType.value = 'unknown'
        }
      } catch (err) {
        error.value = err instanceof Error ? err.message : 'Failed to detect content type'
        contentType.value = 'unknown'
      } finally {
        isLoading.value = false
      }
    } else {
      contentType.value = 'unknown'
    }
  }

  /**
   * Retry loading the document
   */
  async function retry(): Promise<void> {
    error.value = null
    await detectContentType()
  }

  // Watch for URL/storage changes and re-detect content type
  watch([docId, docUrl, isStored, storedMimeType], () => {
    detectContentType()
  }, { immediate: true })

  return {
    proxyUrl,
    contentType,
    isLoading,
    error,
    retry
  }
}

/**
 * Check if content type is an image
 */
export function isImage(type: ContentType): boolean {
  return type.startsWith('image/')
}

/**
 * Check if content type is SVG (requires sanitization)
 */
export function isSvg(type: ContentType): boolean {
  return type === 'image/svg+xml'
}

/**
 * Check if content type is HTML (requires sanitization)
 */
export function isHtml(type: ContentType): boolean {
  return type === 'text/html'
}

/**
 * Check if content type is Markdown
 */
export function isMarkdown(type: ContentType): boolean {
  return type === 'text/markdown'
}

/**
 * Check if content type is PDF
 */
export function isPdf(type: ContentType): boolean {
  return type === 'application/pdf'
}

/**
 * Check if content type is plain text
 */
export function isText(type: ContentType): boolean {
  return type === 'text/plain'
}
