import { appConfig } from '@/configs'
import { getItem } from '@/utils/storage'
import { createI18n } from 'vue-i18n'
import enUS from './lang/en-US'
import zhCN from './lang/zh-CN'

export const SUPPORTED_LOCALES = ['zh-CN', 'en-US'] as const
export type SupportedLocale = (typeof SUPPORTED_LOCALES)[number]

const messages = {
  'zh-CN': zhCN,
  'en-US': enUS,
}

function isSupportedLocale(value: unknown): value is SupportedLocale {
  return SUPPORTED_LOCALES.includes(value as SupportedLocale)
}

function resolveInitialLocale(): SupportedLocale {
  const storedLocale = getItem<string>(appConfig.localeStorageKey, 'cookie')
  if (isSupportedLocale(storedLocale)) {
    return storedLocale
  }

  const browserLang = navigator.language
  const matched = SUPPORTED_LOCALES.find(l => browserLang.startsWith(l.split('-')[0] ?? ''))
  return matched ?? 'zh-CN'
}

export const i18n = createI18n({
  legacy: false,
  locale: resolveInitialLocale(),
  fallbackLocale: 'zh-CN',
  messages,
})
