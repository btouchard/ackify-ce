// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useSignatureStore } from '@/stores/signatures'
import signatureService, { type Signature, type SignatureStatus } from '@/services/signatures'

vi.mock('@/services/signatures')

describe('Signature Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  const mockSignature: Signature = {
    id: 1,
    docId: 'doc-123',
    userSub: 'user-123',
    userEmail: 'test@example.com',
    userName: 'Test User',
    signedAt: '2024-01-15T10:00:00Z',
    payloadHash: 'abc123',
    signature: 'sig123',
    nonce: 'nonce123',
    createdAt: '2024-01-15T10:00:00Z',
    referer: 'https://example.com'
  }

  const mockSignatureStatus: SignatureStatus = {
    docId: 'doc-123',
    userEmail: 'test@example.com',
    isSigned: true,
    signedAt: '2024-01-15T10:00:00Z'
  }

  describe('Initial state', () => {
    it('should initialize with empty user signatures', () => {
      const store = useSignatureStore()

      expect(store.userSignatures).toEqual([])
    })

    it('should initialize with empty document signatures map', () => {
      const store = useSignatureStore()

      expect(store.documentSignatures.size).toBe(0)
    })

    it('should initialize with empty signature statuses map', () => {
      const store = useSignatureStore()

      expect(store.signatureStatuses.size).toBe(0)
    })

    it('should initialize with loading false', () => {
      const store = useSignatureStore()

      expect(store.loading).toBe(false)
    })

    it('should initialize with null error', () => {
      const store = useSignatureStore()

      expect(store.error).toBeNull()
    })
  })

  describe('Computed getters', () => {
    it('should compute user signatures count', () => {
      const store = useSignatureStore()

      expect(store.getUserSignaturesCount).toBe(0)

      store.userSignatures.push(mockSignature)

      expect(store.getUserSignaturesCount).toBe(1)
    })

    it('should get document signatures by doc ID', () => {
      const store = useSignatureStore()

      store.documentSignatures.set('doc-123', [mockSignature])

      expect(store.getDocumentSignatures('doc-123')).toEqual([mockSignature])
      expect(store.getDocumentSignatures('doc-456')).toEqual([])
    })

    it('should get signature status by doc ID', () => {
      const store = useSignatureStore()

      store.signatureStatuses.set('doc-123', mockSignatureStatus)

      expect(store.getSignatureStatus('doc-123')).toEqual(mockSignatureStatus)
      expect(store.getSignatureStatus('doc-456')).toBeUndefined()
    })

    it('should check if document is signed', () => {
      const store = useSignatureStore()

      expect(store.isDocumentSigned('doc-123')).toBe(false)

      store.signatureStatuses.set('doc-123', mockSignatureStatus)

      expect(store.isDocumentSigned('doc-123')).toBe(true)
    })

    it('should return false for unsigned document', () => {
      const store = useSignatureStore()

      store.signatureStatuses.set('doc-123', {
        ...mockSignatureStatus,
        isSigned: false
      })

      expect(store.isDocumentSigned('doc-123')).toBe(false)
    })
  })

  describe('createSignature', () => {
    it('should create signature successfully', async () => {
      const store = useSignatureStore()

      vi.mocked(signatureService.createSignature).mockResolvedValueOnce(mockSignature)

      const result = await store.createSignature({
        docId: 'doc-123',
        referer: 'https://example.com'
      })

      expect(result).toEqual(mockSignature)
      expect(store.userSignatures).toHaveLength(1)
      expect(store.userSignatures[0]).toEqual(mockSignature)
      expect(store.loading).toBe(false)
      expect(store.error).toBeNull()
    })

    it('should add signature to document signatures', async () => {
      const store = useSignatureStore()

      vi.mocked(signatureService.createSignature).mockResolvedValueOnce(mockSignature)

      await store.createSignature({
        docId: 'doc-123'
      })

      const docSigs = store.documentSignatures.get('doc-123')
      expect(docSigs).toHaveLength(1)
      expect(docSigs?.[0]).toEqual(mockSignature)
    })

    it('should update signature status after creation', async () => {
      const store = useSignatureStore()

      vi.mocked(signatureService.createSignature).mockResolvedValueOnce(mockSignature)

      await store.createSignature({
        docId: 'doc-123'
      })

      const status = store.signatureStatuses.get('doc-123')

      expect(status?.isSigned).toBe(true)
      expect(status?.docId).toBe('doc-123')
      expect(status?.userEmail).toBe('test@example.com')
    })

    it('should not add duplicate signatures to user list', async () => {
      const store = useSignatureStore()

      store.userSignatures.push(mockSignature)

      vi.mocked(signatureService.createSignature).mockResolvedValueOnce(mockSignature)

      await store.createSignature({ docId: 'doc-123' })

      expect(store.userSignatures.filter(s => s.id === mockSignature.id)).toHaveLength(1)
    })

    it('should not add duplicate signatures to document list', async () => {
      const store = useSignatureStore()

      store.documentSignatures.set('doc-123', [mockSignature])

      vi.mocked(signatureService.createSignature).mockResolvedValueOnce(mockSignature)

      await store.createSignature({ docId: 'doc-123' })

      expect(store.documentSignatures.get('doc-123')?.filter(s => s.id === mockSignature.id)).toHaveLength(1)
    })

    it('should add new signature at beginning of arrays', async () => {
      const store = useSignatureStore()

      const olderSignature = { ...mockSignature, id: 2, signedAt: '2024-01-14T10:00:00Z' }
      store.userSignatures.push(olderSignature)
      store.documentSignatures.set('doc-123', [olderSignature])

      vi.mocked(signatureService.createSignature).mockResolvedValueOnce(mockSignature)

      await store.createSignature({ docId: 'doc-123' })

      expect(store.userSignatures[0]).toEqual(mockSignature)
      expect(store.documentSignatures.get('doc-123')?.[0]).toEqual(mockSignature)
    })

    it('should set error on failure', async () => {
      const store = useSignatureStore()

      vi.mocked(signatureService.createSignature).mockRejectedValueOnce({
        response: {
          data: {
            error: {
              message: 'You have already signed this document'
            }
          }
        }
      })

      await expect(store.createSignature({ docId: 'doc-123' })).rejects.toThrow()

      expect(store.error).toBe('You have already signed this document')
      expect(store.loading).toBe(false)
    })

    it('should use fallback error message when API error is not formatted', async () => {
      const store = useSignatureStore()

      vi.mocked(signatureService.createSignature).mockRejectedValueOnce(new Error('Network error'))

      await expect(store.createSignature({ docId: 'doc-123' })).rejects.toThrow()

      expect(store.error).toBe('Failed to create signature')
    })
  })

  describe('fetchUserSignatures', () => {
    it('should fetch user signatures successfully', async () => {
      const store = useSignatureStore()

      const signatures = [mockSignature]
      vi.mocked(signatureService.getUserSignatures).mockResolvedValueOnce(signatures)

      await store.fetchUserSignatures()

      expect(store.userSignatures).toEqual(signatures)
      expect(store.loading).toBe(false)
      expect(store.error).toBeNull()
    })

    it('should set error on fetch failure', async () => {
      const store = useSignatureStore()

      vi.mocked(signatureService.getUserSignatures).mockRejectedValueOnce({
        response: {
          data: {
            error: {
              message: 'Unauthorized'
            }
          }
        }
      })

      await expect(store.fetchUserSignatures()).rejects.toThrow()

      expect(store.error).toBe('Unauthorized')
    })
  })

  describe('fetchDocumentSignatures', () => {
    it('should fetch document signatures successfully', async () => {
      const store = useSignatureStore()

      const signatures = [mockSignature]
      vi.mocked(signatureService.getDocumentSignatures).mockResolvedValueOnce(signatures)

      const result = await store.fetchDocumentSignatures('doc-123')

      expect(result).toEqual(signatures)
      expect(store.documentSignatures.get('doc-123')).toEqual(signatures)
      expect(store.loading).toBe(false)
      expect(store.error).toBeNull()
    })

    it('should set error on fetch failure', async () => {
      const store = useSignatureStore()

      vi.mocked(signatureService.getDocumentSignatures).mockRejectedValueOnce({
        response: {
          data: {
            error: {
              message: 'Document not found'
            }
          }
        }
      })

      await expect(store.fetchDocumentSignatures('doc-123')).rejects.toThrow()

      expect(store.error).toBe('Document not found')
    })
  })

  describe('fetchSignatureStatus', () => {
    it('should fetch signature status successfully', async () => {
      const store = useSignatureStore()

      vi.mocked(signatureService.getSignatureStatus).mockResolvedValueOnce(mockSignatureStatus)

      const result = await store.fetchSignatureStatus('doc-123')

      expect(result).toEqual(mockSignatureStatus)
      expect(store.signatureStatuses.get('doc-123')).toEqual(mockSignatureStatus)
      expect(store.loading).toBe(false)
      expect(store.error).toBeNull()
    })

    it('should set error on fetch failure', async () => {
      const store = useSignatureStore()

      vi.mocked(signatureService.getSignatureStatus).mockRejectedValueOnce({
        response: {
          data: {
            error: {
              message: 'Not authenticated'
            }
          }
        }
      })

      await expect(store.fetchSignatureStatus('doc-123')).rejects.toThrow()

      expect(store.error).toBe('Not authenticated')
    })
  })

  describe('checkUserSigned', () => {
    it('should return true when user has signed', async () => {
      const store = useSignatureStore()

      vi.mocked(signatureService.getSignatureStatus).mockResolvedValueOnce(mockSignatureStatus)

      const result = await store.checkUserSigned('doc-123')

      expect(result).toBe(true)
    })

    it('should return false when user has not signed', async () => {
      const store = useSignatureStore()

      vi.mocked(signatureService.getSignatureStatus).mockResolvedValueOnce({
        ...mockSignatureStatus,
        isSigned: false
      })

      const result = await store.checkUserSigned('doc-123')

      expect(result).toBe(false)
    })

    it('should return false on error', async () => {
      const store = useSignatureStore()

      vi.mocked(signatureService.getSignatureStatus).mockRejectedValueOnce(new Error('Network error'))

      const result = await store.checkUserSigned('doc-123')

      expect(result).toBe(false)
    })
  })

  describe('clearError', () => {
    it('should clear error state', () => {
      const store = useSignatureStore()

      store.error = 'Some error'

      store.clearError()

      expect(store.error).toBeNull()
    })
  })

  describe('clearCache', () => {
    it('should clear all cached data', () => {
      const store = useSignatureStore()

      store.userSignatures.push(mockSignature)
      store.documentSignatures.set('doc-123', [mockSignature])
      store.signatureStatuses.set('doc-123', mockSignatureStatus)
      store.error = 'Some error'

      store.clearCache()

      expect(store.userSignatures).toEqual([])
      expect(store.documentSignatures.size).toBe(0)
      expect(store.signatureStatuses.size).toBe(0)
      expect(store.error).toBeNull()
    })
  })
})
