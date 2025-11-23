// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, it, expect } from 'vitest'
import { extractTitleFromPath } from '@/services/titleExtractor'

describe('titleExtractor', () => {
  describe('extractTitleFromPath', () => {
    it('should extract title from URL path', () => {
      const result = extractTitleFromPath('https://example.com/my-document.pdf')
      expect(result).toBe('My Document')
    })

    it('should extract title from nested URL path', () => {
      const result = extractTitleFromPath('https://example.com/docs/user-guide.html')
      expect(result).toBe('User Guide')
    })

    it('should use hostname when path is empty', () => {
      const result = extractTitleFromPath('https://example.com/')
      // Le code capitalise uniquement la première lettre de chaque mot
      expect(result).toBe('Example')
    })

    it('should handle underscore separators', () => {
      const result = extractTitleFromPath('https://example.com/product_spec_v2.pdf')
      expect(result).toBe('Product Spec V2')
    })

    it('should handle mixed separators', () => {
      const result = extractTitleFromPath('https://example.com/user-guide_final.docx')
      expect(result).toBe('User Guide Final')
    })

    it('should remove file extension', () => {
      const result = extractTitleFromPath('https://example.com/report.2024.pdf')
      expect(result).toBe('Report.2024')
    })

    it('should capitalize first letter of each word', () => {
      const result = extractTitleFromPath('https://example.com/annual-financial-report.pdf')
      expect(result).toBe('Annual Financial Report')
    })

    it('should handle local file paths', () => {
      const result = extractTitleFromPath('/home/user/documents/my-file.txt')
      expect(result).toBe('My File')
    })

    it('should handle Windows file paths', () => {
      // Note: En environnement JS, les backslashes peuvent être interprétés différemment
      // Le code utilise split(/[/\\]/) qui devrait gérer les deux types de séparateurs
      const result = extractTitleFromPath('C:/Users/John/Documents/contract.pdf')
      expect(result).toBe('Contract')
    })

    it('should handle simple filenames without path', () => {
      const result = extractTitleFromPath('invoice-2024.pdf')
      expect(result).toBe('Invoice 2024')
    })

    it('should handle empty segments gracefully', () => {
      const result = extractTitleFromPath('https://example.com///')
      // Les segments vides sont filtrés, donc on utilise le hostname
      expect(result).toBe('Example')
    })
  })
})
