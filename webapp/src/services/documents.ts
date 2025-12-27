// SPDX-License-Identifier: AGPL-3.0-or-later
import http, { type ApiResponse } from './http'

export interface CreateDocumentRequest {
  reference: string
  title?: string
}

export interface CreateDocumentResponse {
  docId: string
  url?: string
  title: string
  createdAt: string
}

export interface FindOrCreateDocumentResponse {
  docId: string
  url?: string
  title: string
  checksum?: string
  checksumAlgorithm?: string
  description?: string
  readMode: 'external' | 'integrated'
  allowDownload: boolean
  requireFullRead: boolean
  verifyChecksum: boolean
  createdAt: string
  isNew: boolean // true if created, false if found
}

// MyDocument represents a document in the user's document list
export interface MyDocument {
  id: string
  title: string
  url?: string
  description: string
  createdAt: string
  updatedAt: string
  signatureCount: number
  expectedSignerCount: number
}

// PaginatedResponse for paginated API responses
export interface PaginatedResponse<T> {
  data: T[]
  meta: {
    page: number
    pageSize: number
    total: number
    totalPages: number
  }
}

/**
 * Document service for managing documents
 */
export const documentService = {
  /**
   * Create a new document with a unique ID
   */
  async createDocument(request: CreateDocumentRequest): Promise<CreateDocumentResponse> {
    const response = await http.post<ApiResponse<CreateDocumentResponse>>('/documents', request)
    return response.data.data
  },

  /**
   * Find an existing document by reference or create it if not found
   * @param reference URL, path, or document ID
   * @returns Document information with isNew flag
   */
  async findOrCreateDocument(reference: string): Promise<FindOrCreateDocumentResponse> {
    const response = await http.get<ApiResponse<FindOrCreateDocumentResponse>>(
      '/documents/find-or-create',
      { params: { ref: reference } }
    )
    return response.data.data
  },

  /**
   * List documents created by the current user
   * @param limit Number of documents per page (default: 20)
   * @param page Page number (1-indexed)
   * @param search Optional search query
   * @returns Paginated list of user's documents
   */
  async listMyDocuments(
    limit = 20,
    page = 1,
    search?: string
  ): Promise<PaginatedResponse<MyDocument>> {
    const params: Record<string, any> = { limit, page }
    if (search && search.trim()) {
      params.search = search.trim()
    }

    const response = await http.get<{
      data: MyDocument[]
      success: boolean
      meta: { page: number; pageSize: number; total: number; totalPages: number }
    }>('/users/me/documents', { params })

    return {
      data: response.data.data,
      meta: response.data.meta
    }
  },

  /**
   * Delete a document by ID
   * @param docId Document ID to delete
   */
  async deleteDocument(docId: string): Promise<void> {
    await http.delete(`/admin/documents/${docId}`)
  },
}

export default documentService
