// SPDX-License-Identifier: AGPL-3.0-or-later
import { config } from '@vue/test-utils'
import { vi } from 'vitest'

// Mock window globals injectés par le backend
Object.defineProperty(window, 'ACKIFY_BASE_URL', {
  value: 'http://localhost:8080',
  writable: true,
  configurable: true
})

Object.defineProperty(window, 'ACKIFY_VERSION', {
  value: 'v0.0.0-test',
  writable: true,
  configurable: true
})

Object.defineProperty(window, 'ACKIFY_SMTP_ENABLED', {
  value: true,
  writable: true,
  configurable: true
})

// Mock navigator.clipboard pour les tests
Object.defineProperty(navigator, 'clipboard', {
  value: {
    writeText: vi.fn(() => Promise.resolve())
  },
  writable: true,
  configurable: true
})

// Configuration globale de @vue/test-utils
config.global.mocks = {
  $t: (key: string) => key // Mock simple pour vue-i18n
}
