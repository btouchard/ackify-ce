// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, it, expect } from 'vitest'
import { extractError } from '@/services/http'
import { AxiosError } from 'axios'

describe('http service', () => {
  describe('extractError', () => {
    it('should extract error message from API error response', () => {
      const axiosError = new AxiosError('Request failed')
      axiosError.response = {
        data: {
          error: {
            code: 'VALIDATION_ERROR',
            message: 'Invalid input data'
          }
        },
        status: 400,
        statusText: 'Bad Request',
        headers: {},
        config: {} as any
      }

      const result = extractError(axiosError)
      expect(result).toBe('Invalid input data')
    })

    it('should fallback to axios error message when no API error message', () => {
      const axiosError = new AxiosError('Network Error')

      const result = extractError(axiosError)
      expect(result).toBe('Network Error')
    })

    it('should return generic message for non-axios errors', () => {
      const genericError = new Error('Something went wrong')

      const result = extractError(genericError)
      expect(result).toBe('An unexpected error occurred')
    })

    it('should return generic message for null/undefined errors', () => {
      expect(extractError(null)).toBe('An unexpected error occurred')
      expect(extractError(undefined)).toBe('An unexpected error occurred')
    })

    it('should handle axios error without response', () => {
      const axiosError = new AxiosError('Request timeout')
      axiosError.code = 'ECONNABORTED'

      const result = extractError(axiosError)
      expect(result).toBe('Request timeout')
    })

    it('should handle API error with nested details', () => {
      const axiosError = new AxiosError('Request failed')
      axiosError.response = {
        data: {
          error: {
            code: 'SIGNATURE_EXISTS',
            message: 'You have already signed this document',
            details: {
              signedAt: '2024-01-01T10:00:00Z'
            }
          }
        },
        status: 409,
        statusText: 'Conflict',
        headers: {},
        config: {} as any
      }

      const result = extractError(axiosError)
      expect(result).toBe('You have already signed this document')
    })

    it('should handle malformed API error response', () => {
      const axiosError = new AxiosError('Request failed')
      axiosError.response = {
        data: {
          // Missing error object
        },
        status: 500,
        statusText: 'Internal Server Error',
        headers: {},
        config: {} as any
      }

      const result = extractError(axiosError)
      expect(result).toBe('Request failed')
    })
  })
})
