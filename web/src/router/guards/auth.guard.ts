import { useUserStore } from '@/stores/user'
import type { Router } from 'vue-router'

export function setupAuthGuard(router: Router) {
  router.beforeEach(async to => {
    if (to.meta.public) {
      return true
    }

    const userStore = useUserStore()

    try {
      if (!userStore.userInfo) {
        await userStore.getUserInfo()
      }
    } catch {
      userStore.clearUserInfo()
      return { name: 'login' }
    }

    if (!userStore.userInfo) {
      userStore.clearUserInfo()
      return { name: 'login' }
    }

    const allowedForNonSystem = ['dashboard', 'login', 'NotFound', 'profile']

    if (!userStore.isSystem && to.name && !allowedForNonSystem.includes(to.name as string)) {
      return { name: 'dashboard' }
    }

    return true
  })
}
