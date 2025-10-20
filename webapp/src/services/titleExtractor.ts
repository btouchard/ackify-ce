// SPDX-License-Identifier: AGPL-3.0-or-later
/**
 * Extracts a title from a URL by fetching the page (if CORS allows)
 */
export async function extractTitleFromUrl(url: string): Promise<string> {
  try {
    const response = await fetch(url, {
      mode: 'cors',
      credentials: 'omit'
    })

    if (!response.ok) {
      console.warn('Failed to fetch URL for title extraction:', response.status)
      return extractTitleFromPath(url)
    }

    const contentType = response.headers.get('content-type') || ''

    if (!contentType.includes('text/html')) {
      return extractTitleFromPath(url)
    }

    const html = await response.text()

    const parser = new DOMParser()
    const doc = parser.parseFromString(html, 'text/html')

    const title = doc.querySelector('title')?.textContent?.trim()

    if (title && title.length > 0) {
      return title
    }
  } catch (error) {
    console.warn('Failed to extract title from URL:', error)
  }

  return extractTitleFromPath(url)
}

export function extractTitleFromPath(pathOrUrl: string): string {
  try {
    const url = new URL(pathOrUrl)
    const pathname = url.pathname

    const segments = pathname.split('/').filter(s => s.trim())
    const lastSegment = segments[segments.length - 1] || url.hostname

    const withoutExt = lastSegment.replace(/\.[^.]+$/, '')

    return withoutExt
      .replace(/[-_]/g, ' ')
      .replace(/\b\w/g, c => c.toUpperCase())
  } catch (error) {
    const segments = pathOrUrl.split(/[/\\]/).filter(s => s.trim())
    const lastSegment = segments[segments.length - 1] || pathOrUrl

    const withoutExt = lastSegment.replace(/\.[^.]+$/, '')

    return withoutExt
      .replace(/[-_]/g, ' ')
      .replace(/\b\w/g, c => c.toUpperCase())
  }
}
