// SPDX-License-Identifier: AGPL-3.0-or-later
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import http, { resetCsrfToken } from '@/services/http'

export interface User {
  id: string
  email: string
  name: string
  isAdmin: boolean
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const loading = ref(false)
  const initialized = ref(false)

  const isAuthenticated = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.isAdmin ?? false)

  // Check if user can create documents: admin OR only_admin_can_create is false
  const onlyAdminCanCreate = computed(() => (window as any).ACKIFY_ONLY_ADMIN_CAN_CREATE || false)
  const canCreateDocuments = computed(() => isAdmin.value || !onlyAdminCanCreate.value)

  async function checkAuth() {
    if (initialized.value) return

    loading.value = true
    try {
      const response = await http.get('/users/me')
      user.value = response.data.data
    } catch (error) {
      user.value = null
    } finally {
      loading.value = false
      initialized.value = true
    }
  }

  async function fetchCurrentUser() {
    try {
      const response = await http.get('/users/me')
      user.value = response.data.data
    } catch (error) {
      console.error('Failed to fetch user info:', error)
    }
  }

  async function startOAuthLogin(redirectTo?: string) {
    try {
      console.log('Starting OAuth login...', { redirectTo })
      const response = await http.post('/auth/start', { redirectTo })
      console.log('OAuth response:', response.data)

      if (response.data.data?.redirectUrl) {
        console.log('Redirecting to:', response.data.data.redirectUrl)
        window.location.href = response.data.data.redirectUrl
      } else if (response.data.redirectUrl) {
        console.log('Redirecting to:', response.data.redirectUrl)
        window.location.href = response.data.redirectUrl
      } else {
        console.error('No redirect URL in response:', response.data)
      }
    } catch (error) {
      console.error('OAuth login error:', error)
      throw error
    }
  }

  async function logout() {
    try {
      const response = await http.get('/auth/logout')
      user.value = null
      resetCsrfToken()

      if (response.data.redirectUrl) {
        window.location.href = response.data.redirectUrl
      } else {
        window.location.href = '/'
      }
    } catch (error) {
      console.error('Logout failed:', error)
      user.value = null
      window.location.href = '/'
    }
  }

  function setUser(userData: User) {
    user.value = userData
  }

  return {
    user,
    loading,
    initialized,
    isAuthenticated,
    isAdmin,
    canCreateDocuments,
    checkAuth,
    fetchCurrentUser,
    startOAuthLogin,
    logout,
    setUser,
  }
})