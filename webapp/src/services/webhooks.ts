// SPDX-License-Identifier: AGPL-3.0-or-later
import http, { type ApiResponse } from './http'

export interface Webhook {
  id: number
  title: string
  targetUrl: string
  active: boolean
  events: string[]
  description?: string
  createdAt: string
  updatedAt: string
}

export interface WebhookInput {
  title: string
  targetUrl: string
  secret: string
  active: boolean
  events: string[]
  description?: string
}

export interface WebhookDelivery {
  id: number
  webhookId: number
  eventType: string
  eventId: string
  status: string
  retryCount: number
  maxRetries: number
  createdAt: string
  processedAt?: string
  responseStatus?: number
  lastError?: string
}

export async function listWebhooks(): Promise<ApiResponse<Webhook[]>> {
  const res = await http.get('/admin/webhooks')
  return res.data
}

export async function getWebhook(id: number): Promise<ApiResponse<Webhook>> {
  const res = await http.get(`/admin/webhooks/${id}`)
  return res.data
}

export async function createWebhook(payload: WebhookInput): Promise<ApiResponse<Webhook>> {
  const res = await http.post('/admin/webhooks', payload)
  return res.data
}

export async function updateWebhook(id: number, payload: WebhookInput): Promise<ApiResponse<Webhook>> {
  const res = await http.put(`/admin/webhooks/${id}`, payload)
  return res.data
}

export async function toggleWebhook(id: number, enable: boolean): Promise<ApiResponse<{ message: string }>> {
  const res = await http.patch(`/admin/webhooks/${id}/${enable ? 'enable' : 'disable'}`)
  return res.data
}

export async function deleteWebhook(id: number): Promise<ApiResponse<{ message: string }>> {
  const res = await http.delete(`/admin/webhooks/${id}`)
  return res.data
}

export async function listDeliveries(id: number): Promise<ApiResponse<WebhookDelivery[]>> {
  const res = await http.get(`/admin/webhooks/${id}/deliveries`)
  return res.data
}

export const availableWebhookEvents: { key: string; labelKey: string }[] = [
  { key: 'document.created', labelKey: 'admin.webhooks.events.documentCreated' },
  { key: 'signature.created', labelKey: 'admin.webhooks.events.signatureCreated' },
  { key: 'document.completed', labelKey: 'admin.webhooks.events.documentCompleted' },
  { key: 'reminder.sent', labelKey: 'admin.webhooks.events.reminderSent' },
  { key: 'reminder.failed', labelKey: 'admin.webhooks.events.reminderFailed' },
]

