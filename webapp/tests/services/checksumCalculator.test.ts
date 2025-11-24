// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { calculateFileChecksum, formatFileSize } from '@/services/checksumCalculator'

describe('checksumCalculator service', () => {
  let consoleErrorSpy: any

  beforeEach(() => {
    vi.clearAllMocks()
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    consoleErrorSpy.mockRestore()
  })

  describe('calculateFileChecksum', () => {
    it('should calculate SHA-256 checksum successfully', async () => {
      const mockArrayBuffer = new ArrayBuffer(100)
      const mockHashBuffer = new Uint8Array([
        0x6a, 0x09, 0xe6, 0x67, 0xf3, 0xbc, 0xc9, 0x08,
        0xb2, 0xfb, 0x13, 0x66, 0xea, 0x95, 0x7d, 0x3e,
        0x3a, 0xde, 0xc1, 0x75, 0x12, 0x97, 0x4c, 0x47,
        0xf5, 0x60, 0x53, 0x6b, 0x4b, 0x8f, 0x92, 0x00
      ]).buffer

      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        headers: {
          get: vi.fn().mockReturnValue('100')
        },
        arrayBuffer: vi.fn().mockResolvedValue(mockArrayBuffer)
      })

      global.crypto.subtle.digest = vi.fn().mockResolvedValue(mockHashBuffer)

      const result = await calculateFileChecksum('https://example.com/file.pdf')

      expect(result.checksum).toBe('6a09e667f3bcc908b2fb1366ea957d3e3adec17512974c47f560536b4b8f9200')
      expect(result.algorithm).toBe('SHA-256')
      expect(result.size).toBe(100)
    })

    it('should reject files exceeding max size from content-length header', async () => {
      const maxSize = 50 * 1024 * 1024 // 50MB
      const fileSize = 60 * 1024 * 1024 // 60MB

      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        headers: {
          get: vi.fn().mockReturnValue(fileSize.toString())
        }
      })

      await expect(
        calculateFileChecksum('https://example.com/large-file.pdf', maxSize)
      ).rejects.toThrow('Failed to calculate checksum: File too large')
    })

    it('should reject files exceeding max size after download', async () => {
      const maxSize = 50 * 1024 * 1024 // 50MB
      const fileSize = 60 * 1024 * 1024 // 60MB
      const mockArrayBuffer = new ArrayBuffer(fileSize)

      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        headers: {
          get: vi.fn().mockReturnValue(null) // No content-length header
        },
        arrayBuffer: vi.fn().mockResolvedValue(mockArrayBuffer)
      })

      await expect(
        calculateFileChecksum('https://example.com/large-file.pdf', maxSize)
      ).rejects.toThrow('Failed to calculate checksum: File too large')
    })

    it('should handle HTTP errors', async () => {
      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found'
      })

      await expect(
        calculateFileChecksum('https://example.com/missing.pdf')
      ).rejects.toThrow('Failed to calculate checksum: HTTP 404: Not Found')
    })

    it('should handle network errors', async () => {
      global.fetch = vi.fn().mockRejectedValueOnce(new Error('Network error'))

      await expect(
        calculateFileChecksum('https://example.com/file.pdf')
      ).rejects.toThrow('Failed to calculate checksum: Network error')
    })

    it('should use CORS mode and omit credentials', async () => {
      const mockArrayBuffer = new ArrayBuffer(100)
      const mockHashBuffer = new Uint8Array(32).buffer

      const fetchSpy = vi.fn().mockResolvedValueOnce({
        ok: true,
        headers: {
          get: vi.fn().mockReturnValue('100')
        },
        arrayBuffer: vi.fn().mockResolvedValue(mockArrayBuffer)
      })

      global.fetch = fetchSpy
      global.crypto.subtle.digest = vi.fn().mockResolvedValue(mockHashBuffer)

      await calculateFileChecksum('https://example.com/file.pdf')

      expect(fetchSpy).toHaveBeenCalledWith('https://example.com/file.pdf', {
        mode: 'cors',
        credentials: 'omit'
      })
    })

    it('should use default max size of 50MB', async () => {
      const fileSize = 51 * 1024 * 1024 // 51MB

      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        headers: {
          get: vi.fn().mockReturnValue(fileSize.toString())
        }
      })

      await expect(
        calculateFileChecksum('https://example.com/file.pdf')
      ).rejects.toThrow('File too large')
    })

    it('should handle files without content-length header', async () => {
      const mockArrayBuffer = new ArrayBuffer(1000)
      const mockHashBuffer = new Uint8Array(32).buffer

      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        headers: {
          get: vi.fn().mockReturnValue(null)
        },
        arrayBuffer: vi.fn().mockResolvedValue(mockArrayBuffer)
      })

      global.crypto.subtle.digest = vi.fn().mockResolvedValue(mockHashBuffer)

      const result = await calculateFileChecksum('https://example.com/file.pdf')

      expect(result.size).toBe(1000)
    })

    it('should handle crypto API errors', async () => {
      const mockArrayBuffer = new ArrayBuffer(100)

      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        headers: {
          get: vi.fn().mockReturnValue('100')
        },
        arrayBuffer: vi.fn().mockResolvedValue(mockArrayBuffer)
      })

      global.crypto.subtle.digest = vi.fn().mockRejectedValue(new Error('Crypto API not available'))

      await expect(
        calculateFileChecksum('https://example.com/file.pdf')
      ).rejects.toThrow('Failed to calculate checksum: Crypto API not available')
    })

    it('should convert hash to hexadecimal string correctly', async () => {
      const mockArrayBuffer = new ArrayBuffer(10)
      // Hash: 0x00 0x0F 0xFF 0xAB 0xCD ... (test edge cases)
      const mockHashBuffer = new Uint8Array([
        0x00, 0x0f, 0xff, 0xab, 0xcd, 0xef,
        0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc,
        0xde, 0xf0, 0x00, 0xff, 0x00, 0xff,
        0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff,
        0x11, 0x22, 0x33, 0x44, 0x55, 0x66,
        0x77, 0x88
      ]).buffer

      global.fetch = vi.fn().mockResolvedValueOnce({
        ok: true,
        headers: {
          get: vi.fn().mockReturnValue('10')
        },
        arrayBuffer: vi.fn().mockResolvedValue(mockArrayBuffer)
      })

      global.crypto.subtle.digest = vi.fn().mockResolvedValue(mockHashBuffer)

      const result = await calculateFileChecksum('https://example.com/file.pdf')

      // Verify hex conversion with leading zeros
      expect(result.checksum).toMatch(/^[0-9a-f]{64}$/)
      expect(result.checksum).toBe('000fffabcdef123456789abcdef000ff00ffaabbccddeeff1122334455667788')
    })
  })

  describe('formatFileSize', () => {
    it('should format 0 bytes', () => {
      expect(formatFileSize(0)).toBe('0 B')
    })

    it('should format bytes (< 1KB)', () => {
      expect(formatFileSize(500)).toBe('500 B')
      expect(formatFileSize(1023)).toBe('1023 B')
    })

    it('should format kilobytes', () => {
      expect(formatFileSize(1024)).toBe('1 KB')
      expect(formatFileSize(1536)).toBe('1.5 KB')
      expect(formatFileSize(10240)).toBe('10 KB')
      expect(formatFileSize(1024 * 1023)).toBe('1023 KB')
    })

    it('should format megabytes', () => {
      expect(formatFileSize(1024 * 1024)).toBe('1 MB')
      expect(formatFileSize(1024 * 1024 * 1.5)).toBe('1.5 MB')
      expect(formatFileSize(1024 * 1024 * 50)).toBe('50 MB')
    })

    it('should format gigabytes', () => {
      expect(formatFileSize(1024 * 1024 * 1024)).toBe('1 GB')
      expect(formatFileSize(1024 * 1024 * 1024 * 2.5)).toBe('2.5 GB')
    })

    it('should round to 2 decimal places', () => {
      expect(formatFileSize(1536)).toBe('1.5 KB')
      expect(formatFileSize(1024 * 1.234)).toBe('1.23 KB')
      expect(formatFileSize(1024 * 1024 * 1.999)).toBe('2 MB')
    })

    it('should handle edge cases with very small sizes', () => {
      expect(formatFileSize(1)).toBe('1 B')
      expect(formatFileSize(10)).toBe('10 B')
    })

    it('should handle large files correctly', () => {
      expect(formatFileSize(1024 * 1024 * 1024 * 100)).toBe('100 GB')
    })
  })
})
