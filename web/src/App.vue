<script setup lang="ts">
import type { SupportedLocale } from '@/locales'
import { useLocaleStore } from '@/stores/locale'
import enUS from 'element-plus/es/locale/lang/en'
import zhCN from 'element-plus/es/locale/lang/zh-cn'

const elementLocaleMap: Record<SupportedLocale, typeof zhCN> = {
  'zh-CN': zhCN,
  'en-US': enUS,
}

const localeStore = useLocaleStore()

const elementLocale = computed(() => {
  return elementLocaleMap[localeStore.currentLocale]
})
</script>

<template>
  <el-config-provider :locale="elementLocale">
    <router-view />
  </el-config-provider>
</template>

<style>
html,
body,
#app {
  margin: 0;
  padding: 0;
  height: 100%;
}
</style>
