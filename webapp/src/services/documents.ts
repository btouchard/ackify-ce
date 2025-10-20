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
  createdAt: string
  isNew: boolean // true if created, false if found
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
}

export default documentService
