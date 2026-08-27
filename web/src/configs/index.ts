import type { ThemeMode } from '@/stores/theme'
import type { StorageType } from '@/utils/storage/types'

interface AppConfig {
  title: string
  baseURL: string
  storageType: StorageType
  localeStorageKey: string
  themeStorageKey: string
  defaultTheme: ThemeMode
}

function resolveStorageType(raw: string | undefined): StorageType {
  return raw === 'local' ? 'local' : 'cookie'
}

export const appConfig: AppConfig = {
  title: import.meta.env.VITE_APP_TITLE ?? 'IAM',
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '/api/v1',
  storageType: resolveStorageType(import.meta.env.VITE_STORAGE_TYPE),
  localeStorageKey: import.meta.env.VITE_LOCALE_STORAGE_KEY ?? 'iam_locale',
  themeStorageKey: import.meta.env.VITE_THEME_STORAGE_KEY ?? 'iam_theme',
  defaultTheme: 'light',
}
