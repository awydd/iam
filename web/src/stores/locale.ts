import { appConfig } from '@/configs'
import { i18n, type SupportedLocale } from '@/locales'
import { setItem } from '@/utils/storage'
import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useLocaleStore = defineStore('locale', () => {
  const currentLocale = ref<SupportedLocale>(i18n.global.locale.value)

  function changeLocale(locale: SupportedLocale) {
    currentLocale.value = locale
    i18n.global.locale.value = locale
    setItem(appConfig.localeStorageKey, locale, 'cookie')
    document.documentElement.lang = locale
  }

  return {
    currentLocale,
    changeLocale,
  }
})
