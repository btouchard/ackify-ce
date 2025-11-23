// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, it, expect } from 'vitest'
import { detectReference, type ReferenceInfo } from '@/services/referenceDetector'

describe('referenceDetector', () => {
  describe('detectReference', () => {
    it('should detect HTTP URL', () => {
      const result = detectReference('http://example.com/document.pdf')

      expect(result.type).toBe('url')
      expect(result.value).toBe('http://example.com/document.pdf')
      expect(result.isDownloadable).toBe(true)
      expect(result.fileExtension).toBe('pdf')
    })

    it('should detect HTTPS URL', () => {
      const result = detectReference('https://example.com/file.docx')

      expect(result.type).toBe('url')
      expect(result.isDownloadable).toBe(true)
      expect(result.fileExtension).toBe('docx')
    })

    it('should detect downloadable PDF', () => {
      const result = detectReference('https://example.com/report.pdf')

      expect(result.isDownloadable).toBe(true)
      expect(result.fileExtension).toBe('pdf')
    })

    it('should detect downloadable HTML', () => {
      const result = detectReference('https://example.com/page.html')

      expect(result.isDownloadable).toBe(true)
      expect(result.fileExtension).toBe('html')
    })

    it('should detect downloadable Markdown', () => {
      const result = detectReference('https://example.com/README.md')

      expect(result.isDownloadable).toBe(true)
      expect(result.fileExtension).toBe('md')
    })

    it('should detect downloadable text file', () => {
      const result = detectReference('https://example.com/notes.txt')

      expect(result.isDownloadable).toBe(true)
      expect(result.fileExtension).toBe('txt')
    })

    it('should detect non-downloadable URL without extension', () => {
      const result = detectReference('https://example.com/api/endpoint')

      expect(result.type).toBe('url')
      expect(result.isDownloadable).toBe(false)
      expect(result.fileExtension).toBeUndefined()
    })

    it('should detect non-downloadable URL with non-document extension', () => {
      const result = detectReference('https://example.com/image.jpg')

      expect(result.type).toBe('url')
      expect(result.isDownloadable).toBe(false)
      expect(result.fileExtension).toBe('jpg')
    })

    it('should detect Unix file path', () => {
      const result = detectReference('/home/user/documents/file.pdf')

      expect(result.type).toBe('path')
      expect(result.value).toBe('/home/user/documents/file.pdf')
      expect(result.isDownloadable).toBe(false)
    })

    it('should detect Windows file path', () => {
      const result = detectReference('C:\\Users\\John\\file.docx')

      expect(result.type).toBe('path')
      expect(result.isDownloadable).toBe(false)
    })

    it('should detect relative path', () => {
      const result = detectReference('./documents/file.pdf')

      expect(result.type).toBe('path')
      expect(result.isDownloadable).toBe(false)
    })

    it('should detect simple reference without path or URL', () => {
      const result = detectReference('CONTRACT-2024-001')

      expect(result.type).toBe('reference')
      expect(result.value).toBe('CONTRACT-2024-001')
      expect(result.isDownloadable).toBe(false)
    })

    it('should detect alphanumeric reference', () => {
      const result = detectReference('DOC123ABC')

      expect(result.type).toBe('reference')
      expect(result.isDownloadable).toBe(false)
    })

    it('should handle case-insensitive file extensions', () => {
      const result = detectReference('https://example.com/file.PDF')

      expect(result.fileExtension).toBe('pdf')
      expect(result.isDownloadable).toBe(true)
    })

    it('should detect ODT files as downloadable', () => {
      const result = detectReference('https://example.com/document.odt')

      expect(result.isDownloadable).toBe(true)
      expect(result.fileExtension).toBe('odt')
    })

    it('should detect RTF files as downloadable', () => {
      const result = detectReference('https://example.com/document.rtf')

      expect(result.isDownloadable).toBe(true)
      expect(result.fileExtension).toBe('rtf')
    })
  })
})
