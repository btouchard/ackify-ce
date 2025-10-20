// SPDX-License-Identifier: AGPL-3.0-or-later
export type ReferenceType = 'url' | 'path' | 'reference'

export interface ReferenceInfo {
  type: ReferenceType
  value: string
  isDownloadable: boolean
  fileExtension?: string
}

/**
 * Detects the type of document reference
 */
export function detectReference(ref: string): ReferenceInfo {
  if (ref.startsWith('http://') || ref.startsWith('https://')) {
    const ext = getFileExtension(ref)
    const downloadableExts = ['pdf', 'doc', 'docx', 'txt', 'html', 'xml', 'md', 'odt', 'rtf']

    return {
      type: 'url',
      value: ref,
      isDownloadable: ext ? downloadableExts.includes(ext.toLowerCase()) : false,
      fileExtension: ext
    }
  }

  if (ref.includes('/') || ref.includes('\\')) {
    return {
      type: 'path',
      value: ref,
      isDownloadable: false
    }
  }

  return {
    type: 'reference',
    value: ref,
    isDownloadable: false
  }
}

function getFileExtension(url: string): string | undefined {
  try {
    const pathname = new URL(url).pathname
    const match = pathname.match(/\.([a-z0-9]+)$/i)
    return match?.[1]?.toLowerCase()
  } catch {
    const match = url.match(/\.([a-z0-9]+)$/i)
    return match?.[1]?.toLowerCase()
  }
}
