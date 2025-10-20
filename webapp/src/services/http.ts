// SPDX-License-Identifier: AGPL-3.0-or-later
import axios, { type AxiosInstance, AxiosError, type InternalAxiosRequestConfig } from 'axios'

const API_BASE = import.meta.env.VITE_API_URL || '/api/v1'

export interface ApiResponse<T = any> {
  data: T
  meta?: Record<string, any>
}

export interface ApiError {
  error: {
    code: string
    message: string
    details?: Record<string, any>
  }
}

export interface PaginationMeta {
  page: number
  limit: number
  total: number
  totalPages: number
}

const http: AxiosInstance = axios.create({
  baseURL: API_BASE,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
})

let csrfToken: string | null = null

http.interceptors.request.use(
  async (config: InternalAxiosRequestConfig) => {
    if (config.method && ['post', 'put', 'patch', 'delete'].includes(config.method.toLowerCase())) {
      if (!csrfToken) {
        try {
          const response = await axios.get(`${API_BASE}/csrf`, { withCredentials: true })
          csrfToken = response.data.data?.token || response.data.token
          console.log('Fetched CSRF token:', csrfToken ? 'success' : 'failed')
        } catch (error) {
          console.error('Failed to fetch CSRF token:', error)
        }
      }

      if (csrfToken && config.headers) {
        config.headers['X-CSRF-Token'] = csrfToken
      }
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

http.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiError>) => {
    if (error.response?.status === 401) {
      csrfToken = null

      if (!window.location.pathname.startsWith('/')) {
        window.location.href = '/'
      }
    }

    if (error.response?.status === 403 && error.response?.data?.error?.code === 'CSRF_INVALID') {
      csrfToken = null
    }

    return Promise.reject(error)
  }
)

export default http

export const extractError = (error: any): string => {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<ApiError>
    if (axiosError.response?.data?.error?.message) {
      return axiosError.response.data.error.message
    }
    if (axiosError.message) {
      return axiosError.message
    }
  }
  return 'An unexpected error occurred'
}

export const resetCsrfToken = () => {
  csrfToken = null
}