import { fetchMe } from '@/api/modules/user'
import type { UserMeResp } from '@/api/types/user'
import { defineStore } from 'pinia'
import { ref as vueRef } from 'vue'

export const useUserStore = defineStore('user', () => {
  const userInfo = vueRef<UserMeResp | null>(null)
  const loading = vueRef(false)

  const isSystem = computed(() => userInfo.value?.is_system ?? false)

  async function getUserInfo() {
    if (userInfo.value) return userInfo.value

    try {
      loading.value = true
      const { data } = await fetchMe()
      if (data.status) {
        userInfo.value = data.data
      }
    } catch (error) {
      console.error('Failed to fetch user info', error)
    } finally {
      loading.value = false
    }
  }

  function clearUserInfo() {
    userInfo.value = null
  }

  return {
    userInfo,
    loading,
    isSystem,
    getUserInfo,
    clearUserInfo,
  }
})
