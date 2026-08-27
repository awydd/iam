import { appConfig } from '@/configs'
import { getItem, setItem } from '@/utils/storage'
import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ThemeMode = 'light' | 'dark'

export const useThemeStore = defineStore('theme', () => {
  const currentTheme = ref<ThemeMode>(
    (getItem(appConfig.themeStorageKey, 'cookie') as ThemeMode) || appConfig.defaultTheme
  )

  function changeTheme(theme: ThemeMode) {
    currentTheme.value = theme
    setItem(appConfig.themeStorageKey, theme, 'cookie')

    const root = document.documentElement
    if (theme === 'dark') {
      root.classList.add('dark')
    } else {
      root.classList.remove('dark')
    }
  }

  function initTheme() {
    changeTheme(currentTheme.value)
  }

  return {
    currentTheme,
    changeTheme,
    initTheme,
  }
})
