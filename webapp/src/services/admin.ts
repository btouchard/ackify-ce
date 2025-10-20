// SPDX-License-Identifier: AGPL-3.0-or-later
import http, { type ApiResponse } from './http'

// ============================================================================
// TYPES
// ============================================================================

export interface Document {
  docId: string
  title: string
  url: string
  checksum?: string
  checksumAlgorithm?: string
  description: string
  createdAt: string
  updatedAt: string
  createdBy: string
}

export interface ExpectedSigner {
  id: number
  docId: string
  email: string
  name: string
  addedAt: string
  addedBy: string
  notes?: string
  hasSigned: boolean
  signedAt?: string
  userName?: string
  lastReminderSent?: string
  reminderCount: number
  daysSinceAdded: number
  daysSinceLastReminder?: number
}

export interface DocumentStats {
  docId: string
  expectedCount: number
  signedCount: number
  pendingCount: number
  completionRate: number
}

export interface ReminderStats {
  totalSent: number
  pendingCount: number
  lastSentAt?: string
}

export interface UnexpectedSignature {
  userEmail: string
  userName?: string
  signedAtUTC: string
}

export interface DocumentStatus {
  docId: string
  document?: Document
  expectedSigners: ExpectedSigner[]
  unexpectedSignatures: UnexpectedSignature[]
  stats: DocumentStats
  reminderStats?: ReminderStats
  shareLink: string
}

// ============================================================================
// DOCUMENTS
// ============================================================================

// List all documents
export async function listDocuments(limit = 100, offset = 0): Promise<ApiResponse<Document[]>> {
  const response = await http.get('/admin/documents', {
    params: { limit, offset },
  })
  return response.data
}

// Get document details
export async function getDocument(docId: string): Promise<ApiResponse<Document>> {
  const response = await http.get(`/admin/documents/${docId}`)
  return response.data
}

// Get complete document status (main endpoint used by AdminDocumentDetail)
export async function getDocumentStatus(docId: string): Promise<ApiResponse<DocumentStatus>> {
  const response = await http.get(`/admin/documents/${docId}/status`)
  return response.data
}

// Update document metadata
export async function updateDocumentMetadata(
  docId: string,
  metadata: Partial<{
    title: string
    url: string
    checksum: string
    checksumAlgorithm: string
    description: string
  }>
): Promise<ApiResponse<{ message: string; document: Document }>> {
  const response = await http.put(`/admin/documents/${docId}/metadata`, metadata)
  return response.data
}

// Delete document
export async function deleteDocument(docId: string): Promise<ApiResponse<{ message: string }>> {
  const response = await http.delete(`/admin/documents/${docId}`)
  return response.data
}

// ============================================================================
// EXPECTED SIGNERS
// ============================================================================

// Add expected signer (single)
export async function addExpectedSigner(
  docId: string,
  request: { email: string; name: string; notes?: string }
): Promise<ApiResponse<{ message: string; email: string }>> {
  const response = await http.post(`/admin/documents/${docId}/signers`, request)
  return response.data
}

// Remove expected signer
export async function removeExpectedSigner(
  docId: string,
  email: string
): Promise<ApiResponse<{ message: string }>> {
  const response = await http.delete(`/admin/documents/${docId}/signers/${encodeURIComponent(email)}`)
  return response.data
}

// ============================================================================
// REMINDERS
// ============================================================================

export interface ReminderSendResult {
  totalAttempted: number
  successfullySent: number
  failed: number
  errors?: string[]
}

export interface ReminderLog {
  id: number
  docId: string
  recipientEmail: string
  sentAt: string
  sentBy: string
  templateUsed: string
  status: string
  errorMessage?: string
}

// Send reminders
export async function sendReminders(
  docId: string,
  request: { emails?: string[] } = {},
  locale?: string
): Promise<ApiResponse<{ message: string; result: ReminderSendResult }>> {
  const headers: Record<string, string> = {}

  // Send Accept-Language header if locale is provided
  if (locale) {
    headers['Accept-Language'] = locale
  }

  const response = await http.post(`/admin/documents/${docId}/reminders`, request, { headers })
  return response.data
}

// Get reminder history
export async function getReminderHistory(docId: string): Promise<ApiResponse<ReminderLog[]>> {
  const response = await http.get(`/admin/documents/${docId}/reminders`)
  return response.data
}

// ============================================================================
// LEGACY - These endpoints are not yet migrated to API v1
// They will return empty/stub responses until backend support is added
// ============================================================================

export interface ChecksumVerificationHistory {
  id: number
  docId: string
  verifiedBy: string
  verifiedAt: string
  isValid: boolean
  algorithm: string
  calculatedChecksum: string
  storedChecksum: string
  errorMessage?: string
}

// Verify checksum (TODO: migrate to API v1)
export async function verifyChecksum(
  _docId: string,
  _request: { calculated_checksum: string }
): Promise<ApiResponse<any>> {
  // For now, return stub - needs backend implementation in API v1
  return Promise.reject(new Error('Checksum verification not yet available in API v1'))
}

// Get checksum verification history (TODO: migrate to API v1)
export async function getChecksumVerificationHistory(
  _docId: string,
  _limit = 10
): Promise<ApiResponse<{ verifications: ChecksumVerificationHistory[] }>> {
  // Return empty history for now
  return Promise.resolve({
    data: { verifications: [] },
    success: true,
    error: null,
    meta: {}
  } as any)
}
