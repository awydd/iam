import { getNativeStorage } from './core'
import { withPrefix } from './namespace'
import { safeParse, safeStringify } from './safeJSON'
import type { StorageType } from './types'

const getDefaultStorageType = (): StorageType => {
  const env = import.meta.env?.VITE_MENU_STORAGE
  if (env === 'session' || env === 'cookie') return env
  return 'local'
}

export function setItem<T>(
  key: string,
  value: T,
  type: StorageType = getDefaultStorageType()
): void {
  const storage = getNativeStorage(type)
  storage.setItem(withPrefix(key), safeStringify(value))
}

export function getItem<T>(key: string, type: StorageType = getDefaultStorageType()): T | null {
  const storage = getNativeStorage(type)
  const raw = storage.getItem(withPrefix(key))
  return safeParse<T>(raw)
}

export function getItemWithDefault<T>(key: string, defaultValue: T, type?: StorageType): T {
  const value = getItem<T>(key, type)
  return value ?? defaultValue
}

export function removeItem(key: string, type: StorageType = getDefaultStorageType()): void {
  const storage = getNativeStorage(type)
  storage.removeItem(withPrefix(key))
}

export function clear(type: StorageType = getDefaultStorageType()): void {
  const storage = getNativeStorage(type)
  storage.clear()
}

export type { StorageType }
