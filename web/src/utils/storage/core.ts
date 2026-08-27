import type { StorageType } from './types'

interface StorageLike {
  getItem: (key: string) => string | null
  setItem: (key: string, value: string) => void
  removeItem: (key: string) => void
  clear: () => void
}

function getCookieValue(key: string): string | null {
  const match = document.cookie.match(
    new RegExp('(?:^|; )' + key.replace(/([.$?*|{}()[\]\\/+^:])/g, '\\$1') + '=([^;]*)')
  )
  return match ? decodeURIComponent(match[1] ?? '') : null
}

function setCookieValue(key: string, value: string, days = 365): void {
  const expires = new Date(Date.now() + days * 24 * 60 * 60 * 1000).toUTCString()
  document.cookie = `${key}=${encodeURIComponent(value)}; expires=${expires}; path=/; SameSite=Lax`
}

function removeCookieValue(key: string): void {
  document.cookie = `${key}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/`
}

function getAllCookieKeys(): string[] {
  return document.cookie
    .split('; ')
    .map(item => item.split('=')[0])
    .filter((key): key is string => Boolean(key))
}

const cookieStorage: StorageLike = {
  getItem: getCookieValue,
  setItem: setCookieValue,
  removeItem: removeCookieValue,
  clear: () => {
    getAllCookieKeys().forEach(removeCookieValue)
  },
}

export const getNativeStorage = (type: StorageType): StorageLike => {
  if (type === 'cookie') return cookieStorage
  return type === 'session' ? sessionStorage : localStorage
}
