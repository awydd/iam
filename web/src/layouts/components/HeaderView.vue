<script setup lang="ts">
import { logout } from '@/api/modules/auth'
import { SUPPORTED_LOCALES, type SupportedLocale } from '@/locales'
import { useLocaleStore } from '@/stores/locale'
import { useThemeStore } from '@/stores/theme'
import { useUserStore } from '@/stores/user'
import { reportError } from '@/utils/message'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

defineProps<{
  collapsed: boolean
}>()

const emit = defineEmits<{
  'toggle-collapse': []
}>()

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const localeStore = useLocaleStore()
const themeStore = useThemeStore()
const userStore = useUserStore()

const isFullscreen = ref(false)

const currentTitle = computed(() => {
  const metaName = route.meta?.name as string
  if (metaName) {
    return t(`layout.menus.${metaName}`)
  }
  return (route.name as string) ?? ''
})

function handleLocaleChange(lang: SupportedLocale) {
  localeStore.changeLocale(lang)
}

function toggleTheme() {
  const targetTheme = themeStore.currentTheme === 'light' ? 'dark' : 'light'
  themeStore.changeTheme(targetTheme)
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement
      .requestFullscreen()
      .then(() => {
        isFullscreen.value = true
      })
      .catch(err => {
        ElMessage.error(`Error entering fullscreen: ${err.message}`)
      })
  } else {
    document.exitFullscreen()
    isFullscreen.value = false
  }
}

async function handleLogout() {
  try {
    await ElMessageBox.confirm(t('layout.logoutConfirm'), t('common.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    })

    await logout()

    userStore.clearUserInfo()

    ElMessage.success(t('layout.logoutSuccess'))
    router.push('/login')
  } catch (error) {
    if (error !== 'cancel') {
      reportError(error)
    }
  }
}

function goProfile() {
  router.push('/profile')
}

themeStore.initTheme()
</script>

<template>
  <el-header>
    <div class="header__left">
      <el-icon class="header__collapse-btn" @click="emit('toggle-collapse')">
        <EpFold v-if="!collapsed" />
        <EpExpand v-else />
      </el-icon>

      <el-breadcrumb separator="/">
        <el-breadcrumb-item>{{ currentTitle }}</el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <div class="header__right">
      <span class="header__btn" @click="toggleFullscreen">
        <el-icon :size="20">
          <EpFullScreen v-if="!isFullscreen" />
          <EpAim v-else />
        </el-icon>
      </span>

      <span class="header__theme-btn" @click="toggleTheme">
        <el-icon :size="20">
          <EpMoon v-if="themeStore.currentTheme === 'light'" />
          <EpSunny v-else />
        </el-icon>
      </span>

      <el-dropdown trigger="click" @command="handleLocaleChange">
        <span class="header__lang">
          <el-icon :size="20"><EpConnection /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item
              v-for="l in SUPPORTED_LOCALES"
              :key="l"
              :command="l"
              :disabled="localeStore.currentLocale === l"
            >
              {{ l === 'zh-CN' ? '简体中文' : 'English' }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>

      <el-dropdown trigger="click">
        <span class="header__user">
          <el-avatar :size="32">
            <el-icon><EpUserFilled /></el-icon>
          </el-avatar>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="goProfile">{{ t('layout.profile') }}</el-dropdown-item>
            <el-dropdown-item divided @click="handleLogout">{{
              t('layout.logout')
            }}</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </el-header>
</template>

<style lang="scss" scoped>
.el-header {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  background-color: var(--header-bg-color);
  color: var(--text-color);

  .header__left {
    display: flex;
    align-items: center;

    .el-icon {
      margin-right: 16px;
      cursor: pointer;
    }
  }

  .el-breadcrumb {
    flex-grow: 1;
  }

  .header__right {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 16px;
  }

  .header__btn {
    padding: 0 12px;
    cursor: pointer;
    display: flex;
    align-items: center;
    height: 100%;
    &:hover {
      background-color: var(--el-fill-color-light);
    }
  }

  .header__lang {
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    padding: 6px;
    border-radius: 50%;
  }

  .header__user {
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
  }

  .el-dropdown-menu {
    min-width: 120px;
  }

  .el-avatar {
    display: inline-flex;
    align-items: center;
    justify-content: center;

    .el-icon {
      margin: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      width: 100%;
      height: 100%;
      font-size: 18px;
    }
  }
}
</style>
