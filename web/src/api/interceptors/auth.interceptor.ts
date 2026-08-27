import { i18n } from '@/locales'
import { type AxiosError, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import request from '../request'

declare module 'axios' {
  export interface AxiosRequestConfig {
    _retry?: boolean
  }
  export interface InternalAxiosRequestConfig {
    _retry?: boolean
  }
}

let isRefreshing = false
let pendingQueue: Array<{
  resolve: () => void
  reject: (reason: unknown) => void
}> = []

function resolveQueue() {
  pendingQueue.forEach(({ resolve }) => resolve())
  pendingQueue = []
}

function rejectQueue(error: unknown) {
  pendingQueue.forEach(({ reject }) => reject(error))
  pendingQueue = []
}

request.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  config.headers.set('Accept-Language', i18n.global.locale.value)
  return config
})

async function refreshToken(): Promise<void> {
  await request.post('/auth/refresh', null, { _retry: true, withCredentials: true })
}

function redirectToLogin() {
  if (window.location.pathname.includes('/login')) {
    return
  }
  const redirect = window.location.pathname + window.location.search
  window.location.href = `/login?redirect=${encodeURIComponent(redirect)}`
}

request.interceptors.response.use(
  (response: AxiosResponse) => response,

  async (error: AxiosError) => {
    const originalConfig = error.config as InternalAxiosRequestConfig | undefined

    if (!originalConfig) {
      return Promise.reject(error)
    }

    const status = error.response?.status

    if (status !== 401 || originalConfig._retry) {
      return Promise.reject(error)
    }

    if (isRefreshing) {
      return new Promise<void>((resolve, reject) => {
        pendingQueue.push({ resolve, reject })
      }).then(() => {
        originalConfig._retry = true
        return request(originalConfig)
      })
    }

    isRefreshing = true
    originalConfig._retry = true

    try {
      await refreshToken()
      resolveQueue()
      return await request(originalConfig)
    } catch (refreshError) {
      rejectQueue(refreshError)

      originalConfig._retry = false
      redirectToLogin()

      return Promise.reject(refreshError)
    } finally {
      isRefreshing = false
    }
  }
)

export default request
