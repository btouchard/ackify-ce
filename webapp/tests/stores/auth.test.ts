// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore, type User } from '@/stores/auth'
import http from '@/services/http'

vi.mock('@/services/http')

describe('Auth Store', () => {
  let consoleErrorSpy: any
  let consoleLogSpy: any

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    delete (window as any).location
    ;(window as any).location = { href: '' }
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    consoleLogSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
  })

  afterEach(() => {
    consoleErrorSpy.mockRestore()
    consoleLogSpy.mockRestore()
  })

  const mockUser: User = {
    id: 'user-123',
    email: 'test@example.com',
    name: 'Test User',
    isAdmin: false
  }

  const mockAdminUser: User = {
    id: 'admin-123',
    email: 'admin@example.com',
    name: 'Admin User',
    isAdmin: true
  }

  describe('Initial state', () => {
    it('should initialize with null user', () => {
      const store = useAuthStore()

      expect(store.user).toBeNull()
    })

    it('should initialize with loading false', () => {
      const store = useAuthStore()

      expect(store.loading).toBe(false)
    })

    it('should initialize with initialized false', () => {
      const store = useAuthStore()

      expect(store.initialized).toBe(false)
    })

    it('should compute isAuthenticated as false when no user', () => {
      const store = useAuthStore()

      expect(store.isAuthenticated).toBe(false)
    })

    it('should compute isAdmin as false when no user', () => {
      const store = useAuthStore()

      expect(store.isAdmin).toBe(false)
    })
  })

  describe('checkAuth', () => {
    it('should fetch user data successfully', async () => {
      const store = useAuthStore()

      vi.mocked(http.get).mockResolvedValueOnce({
        data: { data: mockUser }
      } as any)

      await store.checkAuth()

      expect(store.user).toEqual(mockUser)
      expect(store.isAuthenticated).toBe(true)
      expect(store.initialized).toBe(true)
      expect(store.loading).toBe(false)
    })

    it('should set user to null on auth failure', async () => {
      const store = useAuthStore()

      vi.mocked(http.get).mockRejectedValueOnce(new Error('Unauthorized'))

      await store.checkAuth()

      expect(store.user).toBeNull()
      expect(store.isAuthenticated).toBe(false)
      expect(store.initialized).toBe(true)
      expect(store.loading).toBe(false)
    })

    it('should skip check if already initialized', async () => {
      const store = useAuthStore()

      vi.mocked(http.get).mockResolvedValueOnce({
        data: { data: mockUser }
      } as any)

      await store.checkAuth()
      expect(http.get).toHaveBeenCalledTimes(1)

      await store.checkAuth()
      expect(http.get).toHaveBeenCalledTimes(1)
    })

    it('should set loading state during check', async () => {
      const store = useAuthStore()

      let loadingDuringCall = false

      vi.mocked(http.get).mockImplementationOnce(async () => {
        loadingDuringCall = store.loading
        return { data: { data: mockUser } } as any
      })

      await store.checkAuth()

      expect(loadingDuringCall).toBe(true)
      expect(store.loading).toBe(false)
    })
  })

  describe('fetchCurrentUser', () => {
    it('should fetch and update user data', async () => {
      const store = useAuthStore()

      vi.mocked(http.get).mockResolvedValueOnce({
        data: { data: mockUser }
      } as any)

      await store.fetchCurrentUser()

      expect(store.user).toEqual(mockUser)
      expect(http.get).toHaveBeenCalledWith('/users/me')
    })

    it('should log error on fetch failure', async () => {
      const store = useAuthStore()

      vi.mocked(http.get).mockRejectedValueOnce(new Error('Network error'))

      await store.fetchCurrentUser()

      expect(consoleErrorSpy).toHaveBeenCalledWith('Failed to fetch user info:', expect.any(Error))
    })
  })

  describe('startOAuthLogin', () => {
    it('should redirect to OAuth URL from data.redirectUrl', async () => {
      const store = useAuthStore()

      vi.mocked(http.post).mockResolvedValueOnce({
        data: {
          data: {
            redirectUrl: 'https://oauth.provider.com/auth'
          }
        }
      } as any)

      await store.startOAuthLogin()

      expect(window.location.href).toBe('https://oauth.provider.com/auth')
    })

    it('should redirect to OAuth URL from redirectUrl (legacy)', async () => {
      const store = useAuthStore()

      vi.mocked(http.post).mockResolvedValueOnce({
        data: {
          redirectUrl: 'https://oauth.provider.com/auth'
        }
      } as any)

      await store.startOAuthLogin()

      expect(window.location.href).toBe('https://oauth.provider.com/auth')
    })

    it('should pass redirectTo parameter', async () => {
      const store = useAuthStore()

      vi.mocked(http.post).mockResolvedValueOnce({
        data: {
          data: { redirectUrl: 'https://oauth.provider.com/auth' }
        }
      } as any)

      await store.startOAuthLogin('/documents/123')

      expect(http.post).toHaveBeenCalledWith('/auth/start', { redirectTo: '/documents/123' })
    })

    it('should throw error when no redirect URL in response', async () => {
      const store = useAuthStore()

      vi.mocked(http.post).mockResolvedValueOnce({
        data: {}
      } as any)

      await store.startOAuthLogin()

      expect(consoleErrorSpy).toHaveBeenCalledWith('No redirect URL in response:', {})
    })

    it('should throw error on OAuth start failure', async () => {
      const store = useAuthStore()

      vi.mocked(http.post).mockRejectedValueOnce(new Error('OAuth configuration error'))

      await expect(store.startOAuthLogin()).rejects.toThrow('OAuth configuration error')
    })
  })

  describe('logout', () => {
    it('should logout and redirect to home', async () => {
      const store = useAuthStore()
      store.setUser(mockUser)

      vi.mocked(http.get).mockResolvedValueOnce({
        data: {}
      } as any)

      await store.logout()

      expect(store.user).toBeNull()
      expect(window.location.href).toBe('/')
    })

    it('should redirect to custom logout URL if provided', async () => {
      const store = useAuthStore()
      store.setUser(mockUser)

      vi.mocked(http.get).mockResolvedValueOnce({
        data: {
          redirectUrl: 'https://oauth.provider.com/logout'
        }
      } as any)

      await store.logout()

      expect(store.user).toBeNull()
      expect(window.location.href).toBe('https://oauth.provider.com/logout')
    })

    it('should clear user and redirect to home on logout error', async () => {
      const store = useAuthStore()
      store.setUser(mockUser)

      vi.mocked(http.get).mockRejectedValueOnce(new Error('Logout failed'))

      await store.logout()

      expect(store.user).toBeNull()
      expect(window.location.href).toBe('/')
      expect(consoleErrorSpy).toHaveBeenCalledWith('Logout failed:', expect.any(Error))
    })
  })

  describe('setUser', () => {
    it('should set user data', () => {
      const store = useAuthStore()

      store.setUser(mockUser)

      expect(store.user).toEqual(mockUser)
      expect(store.isAuthenticated).toBe(true)
    })
  })

  describe('isAdmin computed property', () => {
    it('should return true for admin user', () => {
      const store = useAuthStore()

      store.setUser(mockAdminUser)

      expect(store.isAdmin).toBe(true)
    })

    it('should return false for non-admin user', () => {
      const store = useAuthStore()

      store.setUser(mockUser)

      expect(store.isAdmin).toBe(false)
    })

    it('should return false when user is null', () => {
      const store = useAuthStore()

      expect(store.isAdmin).toBe(false)
    })
  })

  describe('isAuthenticated computed property', () => {
    it('should return true when user is set', () => {
      const store = useAuthStore()

      store.setUser(mockUser)

      expect(store.isAuthenticated).toBe(true)
    })

    it('should return false when user is null', () => {
      const store = useAuthStore()

      expect(store.isAuthenticated).toBe(false)
    })
  })
})
