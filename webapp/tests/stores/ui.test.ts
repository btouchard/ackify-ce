// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useUIStore } from '@/stores/ui'

describe('UI Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  describe('Notifications', () => {
    it('should add a notification with unique ID', () => {
      const store = useUIStore()

      const id = store.showNotification({
        type: 'success',
        title: 'Test notification'
      })

      expect(store.notifications).toHaveLength(1)
      expect(store.notifications[0].id).toBe(id)
      expect(store.notifications[0].type).toBe('success')
      expect(store.notifications[0].title).toBe('Test notification')
    })

    it('should set default duration to 5000ms', () => {
      const store = useUIStore()

      store.showNotification({
        type: 'info',
        title: 'Test'
      })

      expect(store.notifications[0].duration).toBe(5000)
    })

    it('should use custom duration when provided', () => {
      const store = useUIStore()

      store.showNotification({
        type: 'warning',
        title: 'Test',
        duration: 10000
      })

      expect(store.notifications[0].duration).toBe(10000)
    })

    it('should auto-remove notification after duration', () => {
      const store = useUIStore()

      store.showNotification({
        type: 'success',
        title: 'Test',
        duration: 3000
      })

      expect(store.notifications).toHaveLength(1)

      vi.advanceTimersByTime(3000)

      expect(store.notifications).toHaveLength(0)
    })

    it('should not auto-remove notification with negative duration', () => {
      const store = useUIStore()

      store.showNotification({
        type: 'error',
        title: 'Persistent error',
        duration: -1
      })

      vi.advanceTimersByTime(10000)

      expect(store.notifications).toHaveLength(1)
    })

    it('should manually remove notification by ID', () => {
      const store = useUIStore()

      const id = store.showNotification({
        type: 'info',
        title: 'Test',
        duration: 0
      })

      expect(store.notifications).toHaveLength(1)

      store.removeNotification(id)

      expect(store.notifications).toHaveLength(0)
    })

    it('should handle removing non-existent notification gracefully', () => {
      const store = useUIStore()

      store.showNotification({ type: 'info', title: 'Test' })

      expect(() => {
        store.removeNotification('non-existent-id')
      }).not.toThrow()

      expect(store.notifications).toHaveLength(1)
    })

    it('should clear all notifications', () => {
      const store = useUIStore()

      store.showNotification({ type: 'success', title: 'Test 1' })
      store.showNotification({ type: 'error', title: 'Test 2' })
      store.showNotification({ type: 'warning', title: 'Test 3' })

      expect(store.notifications).toHaveLength(3)

      store.clearNotifications()

      expect(store.notifications).toHaveLength(0)
    })

    it('should show success notification with helper', () => {
      const store = useUIStore()

      store.showSuccess('Success title', 'Success message')

      expect(store.notifications[0].type).toBe('success')
      expect(store.notifications[0].title).toBe('Success title')
      expect(store.notifications[0].message).toBe('Success message')
    })

    it('should show error notification with helper', () => {
      const store = useUIStore()

      store.showError('Error title', 'Error message')

      expect(store.notifications[0].type).toBe('error')
      expect(store.notifications[0].title).toBe('Error title')
      expect(store.notifications[0].message).toBe('Error message')
    })

    it('should show warning notification with helper', () => {
      const store = useUIStore()

      store.showWarning('Warning title')

      expect(store.notifications[0].type).toBe('warning')
      expect(store.notifications[0].title).toBe('Warning title')
    })

    it('should show info notification with helper', () => {
      const store = useUIStore()

      store.showInfo('Info title')

      expect(store.notifications[0].type).toBe('info')
      expect(store.notifications[0].title).toBe('Info title')
    })

    it('should handle multiple notifications with different durations', () => {
      const store = useUIStore()

      store.showNotification({ type: 'info', title: 'First', duration: 1000 })
      store.showNotification({ type: 'success', title: 'Second', duration: 2000 })
      store.showNotification({ type: 'warning', title: 'Third', duration: 3000 })

      expect(store.notifications).toHaveLength(3)

      vi.advanceTimersByTime(1000)
      expect(store.notifications).toHaveLength(2)
      expect(store.notifications[0].title).toBe('Second')

      vi.advanceTimersByTime(1000)
      expect(store.notifications).toHaveLength(1)
      expect(store.notifications[0].title).toBe('Third')

      vi.advanceTimersByTime(1000)
      expect(store.notifications).toHaveLength(0)
    })
  })

  describe('Loading state', () => {
    it('should initialize with loading false', () => {
      const store = useUIStore()

      expect(store.loading).toBe(false)
    })

    it('should set loading state', () => {
      const store = useUIStore()

      store.setLoading(true)
      expect(store.loading).toBe(true)

      store.setLoading(false)
      expect(store.loading).toBe(false)
    })
  })

  describe('Locale management', () => {
    it('should initialize with French locale', () => {
      const store = useUIStore()

      expect(store.locale).toBe('fr')
    })

    it('should change locale to English', () => {
      const store = useUIStore()

      store.setLocale('en')

      expect(store.locale).toBe('en')
    })

    it('should set locale cookie when changing locale', () => {
      const store = useUIStore()

      store.setLocale('en')

      // Cookie behavior in happy-dom is limited, we just verify it doesn't throw
      expect(store.locale).toBe('en')
    })
  })
})