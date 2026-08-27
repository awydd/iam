<script setup lang="ts">
import { login } from '@/api/modules/auth'
import { SUPPORTED_LOCALES, type SupportedLocale } from '@/locales'
import { useLocaleStore } from '@/stores/locale'
import { useThemeStore } from '@/stores/theme'
import { reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const localeStore = useLocaleStore()
const themeStore = useThemeStore()

themeStore.initTheme()

const form = reactive({
  username: '',
  password: '',
})

const loading = ref(false)
const errorMsg = ref('')

async function handleSubmit() {
  if (!form.username || !form.password) {
    errorMsg.value = t('login.usernameRequired')
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    const { data } = await login({ username: form.username, password: form.password })
    if (data.status) {
      const redirect = (route.query.redirect as string) || '/'
      router.replace(redirect)
    } else {
      errorMsg.value = data?.message || t('login.invalidCredentials')
    }
  } catch {
    errorMsg.value = t('login.invalidCredentials')
  } finally {
    loading.value = false
  }
}

function handleLocaleChange(lang: SupportedLocale) {
  localeStore.changeLocale(lang)
}

function toggleTheme() {
  const targetTheme = themeStore.currentTheme === 'light' ? 'dark' : 'light'
  themeStore.changeTheme(targetTheme)
}
</script>

<template>
  <div class="login-page">
    <form class="login-card" @submit.prevent="handleSubmit">
      <div class="locale-switch">
        <span class="login__theme-btn" @click="toggleTheme">
          <el-icon :size="20">
            <EpMoon v-if="themeStore.currentTheme === 'light'" />
            <EpSunny v-else />
          </el-icon>
        </span>
        <el-dropdown trigger="click" @command="handleLocaleChange">
          <span class="login__lang-btn">
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
      </div>

      <h1 class="title">{{ t('login.title') }}</h1>

      <div class="field">
        <label for="username">{{ t('login.username') }}</label>
        <input id="username" v-model="form.username" type="text" autocomplete="username" />
      </div>

      <div class="field">
        <label for="password">{{ t('login.password') }}</label>
        <input
          id="password"
          v-model="form.password"
          type="password"
          autocomplete="current-password"
        />
      </div>

      <p v-if="errorMsg" class="error">{{ errorMsg }}</p>

      <button type="submit" :disabled="loading">
        {{ loading ? t('login.submitting') : t('login.submit') }}
      </button>
    </form>
  </div>
</template>

<style lang="scss" scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background-color: var(--bg-color);
  transition: background-color 0.3s;
}

.login-card {
  width: 360px;
  padding: 32px;
  position: relative;
  border: 1px solid var(--border-color);
  border-radius: 0.368rem;
}

.locale-switch {
  position: absolute;
  top: 16px;
  right: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.login__lang-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  padding: 6px;
  border-radius: 50%;
  transition: all 0.2s;
}

.title {
  margin: 12px 0 24px;
  font-size: 20px;
  text-align: center;
  font-weight: 600;
}

.field {
  margin-bottom: 16px;
}

.field label {
  display: block;
  margin-bottom: 6px;
  font-size: 14px;
  font-weight: 500;
}

.field input {
  width: 100%;
  height: 38px;
  padding: 0 12px;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  box-sizing: border-box;
  transition: border-color 0.2s;

  &:focus {
    outline: none;
  }
}

.error {
  margin: 0 0 16px;
  font-size: 13px;
  color: var(--error-color);
}

button {
  width: 100%;
  height: 40px;
  background: var(--success-color);
  border: 1px solid var(--success-border-color);
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  cursor: pointer;

  &:hover {
    background-color: var(--success-hover-color);
  }
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
