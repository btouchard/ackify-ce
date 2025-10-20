// SPDX-License-Identifier: AGPL-3.0-or-later
export interface ChecksumResult {
  checksum: string
  algorithm: string
  size: number
}

/**
 * Downloads a file and calculates its SHA-256 checksum
 */
export async function calculateFileChecksum(
  url: string,
  maxSize: number = 50 * 1024 * 1024
): Promise<ChecksumResult> {
  try {
    const response = await fetch(url, {
      mode: 'cors',
      credentials: 'omit'
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`)
    }

    const contentLength = response.headers.get('content-length')
    if (contentLength) {
      const size = parseInt(contentLength, 10)
      if (size > maxSize) {
        throw new Error(`File too large: ${(size / 1024 / 1024).toFixed(2)}MB (max: ${(maxSize / 1024 / 1024).toFixed(0)}MB)`)
      }
    }

    const arrayBuffer = await response.arrayBuffer()

    if (arrayBuffer.byteLength > maxSize) {
      throw new Error(`File too large: ${(arrayBuffer.byteLength / 1024 / 1024).toFixed(2)}MB (max: ${(maxSize / 1024 / 1024).toFixed(0)}MB)`)
    }

    const hashBuffer = await crypto.subtle.digest('SHA-256', arrayBuffer)

    const hashArray = Array.from(new Uint8Array(hashBuffer))
    const checksum = hashArray.map(b => b.toString(16).padStart(2, '0')).join('')

    return {
      checksum,
      algorithm: 'SHA-256',
      size: arrayBuffer.byteLength
    }
  } catch (error) {
    console.error('Checksum calculation failed:', error)

    if (error instanceof Error) {
      throw new Error(`Failed to calculate checksum: ${error.message}`)
    }

    throw new Error('Failed to calculate checksum: Unknown error')
  }
}

export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'

  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
}
