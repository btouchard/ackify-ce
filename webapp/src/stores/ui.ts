// SPDX-License-Identifier: AGPL-3.0-or-later
import { defineStore } from 'pinia'
import { ref } from 'vue'

export type NotificationType = 'success' | 'error' | 'warning' | 'info'

export interface Notification {
  id: string
  type: NotificationType
  title: string
  message?: string
  duration?: number
}

export const useUIStore = defineStore('ui', () => {
  const notifications = ref<Notification[]>([])
  const loading = ref(false)
  const locale = ref<'en' | 'fr'>('fr')

  function showNotification(notification: Omit<Notification, 'id'>) {
    const id = `notification-${Date.now()}`
    const newNotification: Notification = {
      ...notification,
      id,
      duration: notification.duration || 5000,
    }

    notifications.value.push(newNotification)

    if (newNotification.duration && newNotification.duration > 0) {
      setTimeout(() => {
        removeNotification(id)
      }, newNotification.duration)
    }

    return id
  }

  function removeNotification(id: string) {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index !== -1) {
      notifications.value.splice(index, 1)
    }
  }

  function clearNotifications() {
    notifications.value = []
  }

  function showSuccess(title: string, message?: string) {
    return showNotification({ type: 'success', title, message })
  }

  function showError(title: string, message?: string) {
    return showNotification({ type: 'error', title, message })
  }

  function showWarning(title: string, message?: string) {
    return showNotification({ type: 'warning', title, message })
  }

  function showInfo(title: string, message?: string) {
    return showNotification({ type: 'info', title, message })
  }

  function setLoading(isLoading: boolean) {
    loading.value = isLoading
  }

  function setLocale(newLocale: 'en' | 'fr') {
    locale.value = newLocale
    document.cookie = `lang=${newLocale};path=/;max-age=31536000`
  }

  return {
    notifications,
    loading,
    locale,
    showNotification,
    removeNotification,
    clearNotifications,
    showSuccess,
    showError,
    showWarning,
    showInfo,
    setLoading,
    setLocale,
  }
})